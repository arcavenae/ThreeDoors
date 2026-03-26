# ThreeDoors Doctor Command — Research Report

**Date:** 2026-03-09
**Type:** Research spike — no implementation code
**Participants:** PM, Architect, UX Designer, Dev (party mode deliberation)

---

## Table of Contents

1. [Prior Art Survey](#1-prior-art-survey)
2. [ThreeDoors-Specific Checks](#2-threedoors-specific-checks)
3. [Channel-Aware Version Checking](#3-channel-aware-version-checking)
4. [Repair vs Report-Only](#4-repair-vs-report-only)
5. [UX Design](#5-ux-design)
6. [Relationship to Existing `health` Command](#6-relationship-to-existing-health-command)
7. [Suggested Epic/Story Breakdown](#7-suggested-epicstory-breakdown)
8. [Party Mode Deliberation Summary](#8-party-mode-deliberation-summary)

---

## 1. Prior Art Survey

### brew doctor
- **Checks:** Directory access, dev tools, symlink integrity, formula/keg health, tap health, PATH, environment
- **Output:** Plain text paragraphs with `Warning:` / `Error:` prefixes. Success: "Your system is ready to brew."
- **Auto-fix:** No `--fix` flag. Suggests exact commands for user to run manually.
- **Takeaway:** Mature model, but output can be overwhelming. No structured format.

### flutter doctor
- **Checks:** SDK version/channel, Android toolchain, iOS toolchain, Chrome, IDEs, connected devices, network
- **Output:** Category-based with Unicode icons: `[✓]` pass, `[!]` warning, `[✗]` failure. Summary line at bottom.
- **Verbose:** `-v` shows sub-checks per category, `-vv` for deeper detail
- **Auto-fix:** No `--fix` flag (open feature request since 2018, never implemented)
- **Takeaway:** Best UX pattern — scannable icons, hierarchical detail. Gold standard for doctor commands.

### npm doctor
- **Checks:** Registry ping, npm version, Node version, config validity, git availability, permissions, cache integrity
- **Output:** Table: CHECK | VALUE | RECOMMENDATION
- **Auto-fix:** None. Cache issues handled by separate `npm cache verify` command.
- **Takeaway:** Table format good for structured results but less scannable than flutter's icons.

### rustup check
- **Checks:** Each installed toolchain channel independently (stable, beta, nightly), plus rustup itself
- **Output:** One line per toolchain: `stable-x86_64-apple-darwin - Up to date : 1.75.0`
- **Channel awareness:** Inherently channel-aware — each channel checked independently. Pre-release is first-class.
- **Takeaway:** Best model for channel-aware version checking.

### docker system df / docker info
- **Checks:** `df` = disk usage by type (images, containers, volumes). `info` = key-value system state dump.
- **Output:** Table format, supports `--format` for Go templates and JSON
- **Auto-fix:** None. `docker system prune` is a separate command.
- **Takeaway:** Good separation of diagnostic (info) vs cleanup (prune) as separate commands.

### gh (GitHub CLI)
- **Checks:** `gh auth status` checks auth state, token scopes, protocol. No `gh doctor`.
- **Version checking:** Background check once per 24 hours. Cached in state file. Opt-out via `GH_NO_UPDATE_NOTIFIER` env var.
- **Takeaway:** Gold standard for version check caching and opt-out patterns in Go CLIs.

### Pattern Summary

| Pattern | Best Example | Recommendation |
|---|---|---|
| Category-based icons | flutter doctor | **Adopt** — best scanability |
| Table format | npm doctor | Use for detailed/verbose mode |
| Cached version check | gh CLI | **Adopt** — 24h cache, env var opt-out |
| Channel-aware checks | rustup check | **Adopt** — check within user's channel |
| Separate cleanup command | docker system prune | **Adopt** — `doctor --fix` not `doctor --prune` |
| No auto-fix | All surveyed tools | Report + suggest; auto-fix only for safe operations |

---

## 2. ThreeDoors-Specific Checks

Based on thorough codebase analysis, here are the check categories organized by priority.

### Category 1: Environment (always runs)

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Config directory | `~/.threedoors/` exists, readable, writable | Missing or bad permissions | FAIL |
| Config file | `config.yaml` exists, valid YAML, schema_version ≤ 2 | Corrupt, missing, outdated schema | FAIL/WARN |
| Go version | Runtime version adequate | Shouldn't happen with compiled binary | INFO |
| Terminal size | Width × height detected | Very narrow terminal → theme fallback | WARN |
| Color support | Ascii/256/TrueColor profile | Ascii-only → degraded experience | INFO |

### Category 2: Task Data Integrity

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Task file | `tasks.yaml` exists, valid YAML | Missing, corrupt, parse error | FAIL |
| Task validation | All tasks pass `Validate()` — valid IDs, text 1-500 chars, valid status/type/effort/location, timestamps consistent | Bad data | WARN per task |
| Duplicate IDs | No two tasks share an ID | UUID collision (extremely rare) | FAIL |
| Orphaned deps | `depends_on` references exist in task pool | Dangling dependency reference | WARN |
| Blocker consistency | `blocker` field only on `blocked` status tasks | Stale blocker data | WARN |
| CompletedAt consistency | Only set on complete/archived tasks | Data inconsistency | WARN |
| Legacy migration | `tasks.txt` exists but `tasks.yaml` doesn't | Needs migration | WARN |
| Legacy fields | `source_provider` still present (should be `source_refs`) | Needs migration | INFO |

### Category 3: Session & Analytics Files

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| sessions.jsonl | Each line is valid JSON | Truncated/corrupt lines | WARN |
| Session fields | Required fields present (session_id, timestamps, duration) | Incomplete entries | WARN |
| patterns.json | Valid JSON, expected structure | Corrupt cache | WARN |
| Session count | At least 1 session recorded | New install or data loss | INFO |

### Category 4: Provider Health

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Provider configured | At least one provider in config | Unconfigured | FAIL |
| Provider load | `LoadTasks()` succeeds | Provider unavailable | FAIL |
| Provider-specific | Obsidian: vault path exists. Jira: URL reachable, auth valid. GitHub: token valid. Apple Notes: accessible. Todoist: API token valid. Reminders: macOS only. | Various | FAIL/WARN |
| Multi-provider | All configured providers healthy | Partial failure | WARN |

### Category 5: Sync & Offline Queue

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Sync state | `sync_state.yaml` valid, not stale (>24h) | Corrupt or old | WARN |
| WAL queue | `sync-queue.jsonl` valid, no stuck entries (retries ≥ 10) | Stuck operations | WARN |
| WAL size | Queue < 10,000 entries | Excessive backlog | WARN |
| Sync log | `sync.log` valid JSONL | Corrupt entries | INFO |
| Temp files | No orphaned `.tmp` files from failed atomic writes | Interrupted writes | WARN |

### Category 6: Enrichment Database

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| DB exists | `enrichment.db` can be opened | Missing or locked | WARN |
| Schema version | Version is 1 (current) | Outdated schema | WARN |
| Integrity | `PRAGMA integrity_check` passes | Corruption | FAIL |
| WAL files | `enrichment.db-wal` and `-shm` exist if expected | WAL mode issues | INFO |

### Category 7: Version (optional, requires network)

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Current version | Binary has version embedded (not "dev") | Dev build | INFO |
| Update available | Newer version in user's channel | Outdated | INFO |
| Channel info | Display current channel (stable/alpha/beta) | N/A | INFO |

### Category 8: Onboarding State

| Check | What | Failure Mode | Severity |
|---|---|---|---|
| Onboarding | `onboarding_complete` is true | Incomplete setup | INFO |
| Default tasks | Tasks present (not empty pool) | Empty task list | INFO |

---

## 3. Channel-Aware Version Checking

### Current State

ThreeDoors already embeds version and channel via ldflags:
```go
// cmd/threedoors/main.go
var version = "dev"
var channel = ""  // empty = stable, "alpha", "beta"
```

Version format: `FormatVersionWithChannel(version, channel)` in `internal/dist/version.go`.

### Proposed Design

#### Channel Hierarchy

```
stable < beta < alpha (development progression)
1.0.0 < 1.0.1-beta.1 < 1.0.1-alpha.20260309.abc1234
```

#### Version Check Logic

```
User's Channel    What to Show
─────────────────────────────────────────────────────
stable            Latest stable release only
beta              Latest beta OR stable (whichever is newer)
alpha             Latest alpha OR beta OR stable (whichever is newer)
```

**Key design decision:** An alpha user who sees a stable release *newer than their alpha version's base semver* should be told about it. Example:

- User has `1.2.0-alpha.20260309.abc1234` (base: 1.2.0)
- Stable `1.3.0` is released
- Doctor should say: "Stable 1.3.0 is available (you're on alpha 1.2.0)"

But if stable is `1.1.0` and user is on alpha `1.2.0-alpha...`, don't suggest "downgrading" to stable.

**Rationale:** This matches rustup's approach — each channel is independent, but the user should know if a higher-numbered stable release exists. The alpha channel is for testing the *next* release, not for permanently avoiding stable.

#### Implementation Approach

```go
// Lightweight — no go-github dependency needed
// GET https://api.github.com/repos/arcavenae/ThreeDoors/releases
// Filter by channel:
//   stable: tag has no pre-release suffix
//   beta:   tag contains "-beta"
//   alpha:  tag contains "-alpha"
// Use Masterminds/semver for comparison (already common in Go ecosystem)
```

#### Caching Strategy (gh CLI pattern)

```
Cache file: ~/.threedoors/version-check.json
{
  "checked_at": "2026-03-09T12:00:00Z",
  "latest_stable": "1.3.0",
  "latest_beta": "1.3.1-beta.2",
  "latest_alpha": "1.3.1-alpha.20260309.abc1234",
  "current_version": "1.2.0-alpha.20260309.abc1234",
  "current_channel": "alpha"
}
```

- Check at most **once per 24 hours** (configurable)
- Background goroutine on CLI startup — never blocks command execution
- Display notification on **next** invocation if update found
- Cache TTL resets on successful check

#### Privacy & Opt-Out

- **Opt-out env var:** `THREEDOORS_NO_UPDATE_CHECK=1`
- **Config option:** `update_check: false` in config.yaml
- **CI auto-disable:** Skip when `CI=true`, `GITHUB_ACTIONS=true`, etc.
- **What's sent:** Only a GET request to GitHub's public API. No user data, no current version, no telemetry.
- **Rate limits:** 60 req/hour unauthenticated — 1 req/24h is well within limits.

#### Edge Cases

| Scenario | Behavior |
|---|---|
| No network | Skip check, use cached result if fresh, show nothing if stale |
| Rate limited | Skip, use cache |
| GitHub API down | Skip, use cache |
| `version = "dev"` | Skip version check entirely, show INFO: "Running dev build" |
| Brand new install | Don't check on first run (onboarding flow takes priority) |

---

## 4. Repair vs Report-Only

### Party Mode Consensus: Conservative Auto-Fix

The deliberation reached consensus that `doctor` should **report by default** and **auto-fix only safe, reversible operations** behind a `--fix` flag.

#### Auto-Fixable (with `--fix`)

| Issue | Fix | Reversibility |
|---|---|---|
| Orphaned `.tmp` files | Delete stale temp files (>1 hour old) | Fully reversible (files are transient) |
| Corrupt `patterns.json` | Delete (will regenerate from sessions.jsonl) | Fully reversible (derived data) |
| Missing `config.yaml` | Generate sample config | Safe (only if truly missing) |
| `tasks.txt` needs migration | Run migration to `tasks.yaml` + create `.bak` | Reversible via `.bak` file |
| Stale version cache | Delete `version-check.json` | Fully reversible |
| Legacy config fields | Run schema migration v1→v2 | Reversible (no data loss) |
| Directory permissions | `chmod 700 ~/.threedoors/` | Reversible |

#### Report-Only (never auto-fix)

| Issue | Why Not Auto-Fix |
|---|---|
| Corrupt `tasks.yaml` | User data — could lose tasks. Suggest backup + manual edit. |
| Corrupt `sessions.jsonl` | Historical data. Suggest keeping corrupt file, starting fresh. |
| Corrupt `enrichment.db` | Complex data. Suggest `VACUUM` or delete + rebuild. |
| Stuck WAL entries | May represent real pending changes. User must decide. |
| Provider auth issues | Credentials are user-managed. |
| Duplicate task IDs | Ambiguous which to keep. Show both, let user decide. |
| Dangling dependency refs | User intent unclear — remove ref or recreate task? |

#### `--fix` UX

```
$ threedoors doctor --fix

[✓] Config directory
[✓] Config file
[!] Orphaned temp files — FIXED: removed 2 stale .tmp files
[!] Pattern cache corrupt — FIXED: deleted patterns.json (will regenerate)
[✓] Task file
[✗] Enrichment database corrupt
    → Run: sqlite3 ~/.threedoors/enrichment.db "PRAGMA integrity_check"
    → If failing, delete enrichment.db (will rebuild on next launch)

Fixed 2 issues. 1 issue requires manual intervention.
```

---

## 5. UX Design

### Recommended Style: flutter doctor hybrid

Combines flutter's icon-based scanability with npm's structured data and brew's actionable suggestions.

### ASCII UX Mockups

#### Default Output (happy path)

```
$ threedoors doctor

ThreeDoors Doctor (alpha) v1.2.0-alpha.20260309.abc1234

[✓] Environment
    • Config directory: ~/.threedoors/ (readable, writable)
    • Config file: valid (schema v2)
    • Terminal: 120×40, TrueColor

[✓] Task Data
    • 12 tasks loaded (3 active, 2 blocked, 7 todo)
    • No duplicate IDs
    • All dependencies valid

[✓] Providers
    • textfile: healthy (12 tasks)

[✓] Session Data
    • 47 sessions recorded
    • Pattern cache: fresh (2 hours ago)

[✓] Sync
    • Last sync: 3 hours ago
    • WAL queue: empty
    • No orphaned temp files

[✓] Database
    • Enrichment DB: healthy
    • Schema version: 1

[i] Version
    • Current: alpha v1.2.0-alpha.20260309.abc1234
    • Latest alpha: v1.2.0-alpha.20260310.def5678
    • Update: brew upgrade threedoors-alpha

No issues found. Your system is ready to use.
```

#### Output with Issues

```
$ threedoors doctor

ThreeDoors Doctor v1.1.0

[✓] Environment
    • Config directory: ~/.threedoors/ (readable, writable)
    • Config file: valid (schema v2)
    • Terminal: 80×24, 256 colors

[!] Task Data — 2 warnings
    • 8 tasks loaded (2 active, 6 todo)
    • WARNING: Task "abc123" has blocker set but status is "todo"
      → Update task status to "blocked" or remove the blocker field
    • WARNING: Task "def456" depends on non-existent task "xyz789"
      → Remove the dependency or recreate the missing task

[✓] Providers
    • textfile: healthy (8 tasks)

[✗] Session Data — 1 error
    • ERROR: sessions.jsonl has 3 corrupt lines (lines 42, 87, 103)
      → These sessions will be skipped in analytics
      → To repair: threedoors doctor --fix (removes corrupt lines)

[!] Sync — 1 warning
    • Last sync: 3 days ago
      → Press S in doors view to trigger sync
    • WARNING: 2 orphaned .tmp files found
      → Run: threedoors doctor --fix

[✓] Database
    • Enrichment DB: healthy

[i] Version
    • Current: v1.1.0 (stable)
    • Latest: v1.3.0
    • Update: brew upgrade threedoors

Doctor found issues in 2 categories.
Run 'threedoors doctor --fix' to auto-repair fixable issues.
```

#### Verbose Output (`-v`)

```
$ threedoors doctor -v

ThreeDoors Doctor v1.1.0

[✓] Environment
    • Config directory: ~/.threedoors/
      Path: /Users/alice/.threedoors
      Permissions: drwx------ (700)
      Disk free: 42.3 GB
    • Config file: config.yaml
      Schema version: 2
      Provider: textfile
      Theme: modern
      Key hints: enabled
      Onboarding: complete
    • Terminal capabilities
      Size: 120×40
      Color profile: TrueColor
      Unicode: supported
      TERM: xterm-256color

[✓] Task Data (detailed)
    • Task file: /Users/alice/.threedoors/tasks.yaml
      Size: 4.2 KB
      Last modified: 2 hours ago
    • 12 tasks:
      - active: 3
      - todo: 5
      - blocked: 2
      - complete: 1
      - archived: 1
    • ID uniqueness: all 12 IDs unique
    • Dependencies: 4 dependencies, all valid
    • Blocker consistency: OK
    • Timestamp consistency: OK

... (continues with same detail level for all categories)
```

#### JSON Output (`--json`)

```json
{
  "schema_version": 1,
  "command": "doctor",
  "data": {
    "overall": "WARN",
    "duration_ms": 142,
    "version": "1.1.0",
    "channel": "stable",
    "categories": [
      {
        "name": "Environment",
        "status": "OK",
        "checks": [
          {"name": "Config Directory", "status": "OK", "message": "~/.threedoors/ readable and writable"},
          {"name": "Config File", "status": "OK", "message": "valid (schema v2)"},
          {"name": "Terminal", "status": "OK", "message": "120×40, TrueColor"}
        ]
      },
      {
        "name": "Task Data",
        "status": "WARN",
        "checks": [
          {"name": "Task File", "status": "OK", "message": "8 tasks loaded"},
          {"name": "Blocker Consistency", "status": "WARN", "message": "Task abc123 has blocker but status is todo", "suggestion": "Update status to blocked or remove blocker"}
        ]
      }
    ]
  }
}
```

### Icon Legend

| Icon | Meaning | Color |
|---|---|---|
| `[✓]` | All checks passed | Green |
| `[!]` | Warning — works but has issues | Yellow |
| `[✗]` | Error — something is broken | Red |
| `[i]` | Informational — no action needed | Blue/Cyan |
| `[ ]` | Skipped — not applicable | Gray |

### CLI Flags

| Flag | Description |
|---|---|
| (none) | Standard check with icons |
| `-v` / `--verbose` | Detailed sub-check information |
| `--json` | JSON output (matches existing CLI pattern) |
| `--fix` | Auto-repair safe issues |
| `--skip-version` | Skip network version check |
| `--category <name>` | Run only specific category (env, tasks, providers, sessions, sync, db, version) |

---

## 6. Relationship to Existing `health` Command

### Current `health` Command (internal/cli/health.go + internal/core/health_checker.go)

The existing `health` command checks 4 things:
1. Task file exists and is writable
2. Database (provider) can load tasks
3. Sync status (last sync time, warns at 24h+)
4. Apple Notes access

Output: tabwriter table (CHECK | STATUS | MESSAGE) + JSON mode.

### Proposed Relationship

**Option A: `doctor` supersedes `health`** (RECOMMENDED)
- `doctor` becomes the comprehensive diagnostic tool
- `health` becomes an alias for `doctor` (or deprecated with a message)
- The existing `HealthChecker` struct and `HealthCheckItem`/`HealthStatus` types are reused and extended
- New check categories added as methods on an expanded `DoctorChecker`

**Option B: `doctor` and `health` coexist**
- `health` stays as a quick, focused check (provider connectivity)
- `doctor` is the deep diagnostic
- Risk: user confusion about which to run

**Recommendation:** Option A. The existing health command has only 4 checks and no users are scripting against its output yet (it was added recently). Absorb it into doctor, keep `health` as an alias.

### Migration Path

1. Create `internal/cli/doctor.go` with the new command
2. Extend `internal/core/health_checker.go` → `internal/core/doctor.go` (or expand in-place)
3. Register `doctor` as primary, `health` as alias in cobra
4. Existing JSON schema for health is a subset of doctor's JSON schema

---

## 7. Suggested Epic/Story Breakdown

### Epic: "ThreeDoors Doctor — Self-Diagnosis Command"

**Total estimate: ~8-10 stories**

#### Story 1: Doctor Command Skeleton
- Create `internal/cli/doctor.go` with cobra command
- Create `internal/core/doctor.go` with `DoctorChecker` struct
- Implement category-based output with icons (`[✓]`, `[!]`, `[✗]`, `[i]`)
- Support `--json` output matching existing CLI envelope pattern
- Register `doctor` command, make `health` an alias
- **AC:** `threedoors doctor` runs and shows Environment category only

#### Story 2: Environment Checks
- Config directory existence and permissions
- Config file validation (YAML parse, schema version)
- Terminal capability detection (size, color profile)
- **AC:** Environment category shows config dir, config file, terminal info

#### Story 3: Task Data Integrity Checks
- Task file existence and YAML validity
- Per-task validation (IDs, text length, status, timestamps)
- Duplicate ID detection
- Dependency validation (dangling refs)
- Blocker/CompletedAt consistency
- Legacy migration detection (tasks.txt, source_provider)
- **AC:** Task Data category reports all validation results

#### Story 4: Provider Health Checks
- Integrate existing `HealthChecker.CheckDatabaseReadWrite()` and provider-specific checks
- Multi-provider support (check each configured provider)
- Provider-specific connectivity tests
- **AC:** Providers category shows health of all configured providers

#### Story 5: Session & Analytics Checks
- sessions.jsonl line-by-line validation
- patterns.json validity check
- Session count and freshness
- **AC:** Session Data category reports file health and stats

#### Story 6: Sync & Offline Queue Checks
- Sync state file validation and staleness
- WAL queue validation (stuck entries, size)
- Sync log validation
- Orphaned temp file detection
- **AC:** Sync category reports queue health and temp file status

#### Story 7: Enrichment Database Checks
- SQLite open/close test
- Schema version check
- `PRAGMA integrity_check`
- **AC:** Database category reports DB health

#### Story 8: Auto-Repair (`--fix` flag)
- Orphaned temp file cleanup
- Corrupt patterns.json deletion
- Legacy config migration
- Stale version cache cleanup
- Directory permission repair
- Corrupt JSONL line removal (sessions.jsonl)
- **AC:** `--fix` repairs safe issues, reports what was fixed

#### Story 9: Channel-Aware Version Checking
- GitHub Releases API integration (lightweight, no go-github)
- Channel-aware filtering (stable/beta/alpha)
- 24-hour cached background check
- Opt-out via env var (`THREEDOORS_NO_UPDATE_CHECK`) and config
- CI auto-disable
- **AC:** Version category shows current vs latest, respects channel

#### Story 10: Verbose Mode & Polish
- `--verbose` flag for detailed sub-check output
- `--category` flag for selective checking
- `--skip-version` flag
- Summary line with issue count
- Color output (green/yellow/red/cyan icons)
- **AC:** All flags work, output matches UX mockups

### Dependencies

```
Story 1 (skeleton) → Stories 2-7 (check categories, parallelizable)
Stories 2-7 → Story 8 (--fix needs checks to exist)
Story 1 → Story 9 (version check is independent category)
Stories 2-9 → Story 10 (polish after all checks exist)
```

### Estimated Complexity

- Stories 1-7: Small-Medium (each is a focused check category)
- Story 8: Medium (repair logic needs careful safety)
- Story 9: Medium-Large (network, caching, channel logic)
- Story 10: Small (flags and formatting)

---

## 8. Party Mode Deliberation Summary

### Adopted Decisions

1. **Command name: `doctor`** (not `check`, `diagnose`, or `inspect`)
   - Rationale: Most widely recognized pattern (brew, flutter, npm all use "doctor")
   - Rejected: `check` (too generic, conflicts with linter terminology), `diagnose` (too medical/verbose)

2. **flutter-style icon output** as default
   - Rationale: Best scanability, well-established UX pattern
   - Rejected: npm-style table (harder to scan), brew-style paragraphs (too verbose)

3. **`doctor` supersedes `health`** (Option A above)
   - Rationale: Avoids user confusion, health command is new enough to absorb
   - Rejected: Keeping both (confusing "which do I run?")

4. **Conservative auto-fix with `--fix`**
   - Rationale: Only fix safe, reversible operations. User data (tasks, sessions) never auto-modified.
   - Rejected: Aggressive auto-fix (too risky), no auto-fix at all (misses easy wins like temp file cleanup)

5. **24-hour cached version check** (gh CLI pattern)
   - Rationale: Proven pattern, respects rate limits, non-blocking
   - Rejected: Check every run (too chatty, rate limit risk), weekly check (too stale)

6. **Channel-aware: show upgrades within channel + cross-channel if higher**
   - Rationale: Alpha users should know about newer stable releases
   - Rejected: Strict channel isolation (user misses important stable releases), always show all channels (noisy)

### Rejected Approaches

1. **Interactive repair wizard** — Too complex for v1. Doctor should be non-interactive. If repair needs user input, print instructions.

2. **`doctor` as part of every command startup** — Too slow. Version check is background-only, doctor is explicit.

3. **Telemetry/crash reporting integration** — Out of scope. Doctor is local-only diagnostics. Telemetry is a separate epic if ever pursued.

4. **Plugin health checks** — No plugin system exists. When/if one is added, doctor can be extended.

5. **Network diagnostics beyond version check** — Don't ping registries or test provider APIs unless the provider is configured. Avoid unnecessary network calls.

### Open Questions for Future Consideration

1. Should `doctor` run automatically after `brew upgrade threedoors`? (Post-install hook)
2. Should the TUI show a subtle indicator when doctor found issues? (Status bar icon)
3. Should `doctor --fix` create backups before modifying files? (Probably yes, but adds complexity)
4. Should there be a `doctor --watch` mode for continuous monitoring? (Probably not — overkill)

---

## Appendix: Key Source Files Referenced

| File | Relevance |
|---|---|
| `cmd/threedoors/main.go` | Entry point, version/channel vars, command registration |
| `internal/cli/health.go` | Existing health command (to be absorbed) |
| `internal/core/health_checker.go` | Existing health checks (to be extended) |
| `internal/core/provider_config.go` | Config loading, validation, migration |
| `internal/core/task.go` | Task struct, Validate() method |
| `internal/core/session_tracker.go` | Session metrics structure |
| `internal/core/metrics_writer.go` | sessions.jsonl append logic |
| `internal/core/wal_provider.go` | WAL queue, replay, corruption handling |
| `internal/core/sync_state.go` | Sync state persistence |
| `internal/core/sync_log.go` | Sync log rotation |
| `internal/core/pattern_analyzer.go` | Pattern analysis types |
| `internal/enrichment/enrichment.go` | SQLite enrichment DB |
| `internal/dist/version.go` | Version formatting with channel |
| `internal/adapters/textfile/file_manager.go` | YAML task persistence, atomic writes |
| `internal/core/config_paths.go` | Config directory paths |
| `internal/core/onboarding.go` | First-run detection |
