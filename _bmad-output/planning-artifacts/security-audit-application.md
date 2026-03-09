# ThreeDoors Application Security Audit

**Date:** 2026-03-09
**Auditor:** bold-rabbit (automated security review)
**Scope:** Application-level security of ThreeDoors Go TUI application
**Codebase:** Go 1.25.4, Bubbletea TUI, YAML task files, JSONL session logs

---

## Executive Summary

**Overall Risk Rating: LOW-MEDIUM**

ThreeDoors is a well-engineered local TUI application with no critical remote-exploitable vulnerabilities. The codebase demonstrates strong security practices in most areas: atomic file writes, proper error handling, input validation, and no shell injection vectors. However, the audit identified **filesystem permission inconsistencies** that could expose user task data to other local users, and a **lack of symlink validation** that could be exploited in multi-user environments.

| Severity | Count | Summary |
|----------|-------|---------|
| Critical | 0 | No remote code execution or data corruption vectors |
| High | 2 | Filesystem permissions too permissive; no symlink protection |
| Medium | 4 | Scanner buffer limits, YAML bomb potential, inconsistent permission model, config may store secrets |
| Low | 3 | Error message path leakage, bufio default limits, .tmp file brief exposure |
| Info | 5 | Positive observations and hardening opportunities |

---

## Detailed Findings

### HIGH-1: Data Directory and Files World-Readable (0o755/0o644)

**Severity:** HIGH
**Category:** File I/O & Permissions
**Impact:** Any local user can read all task data, session logs, and configuration

**Finding:** The `~/.threedoors/` directory is created with `0o755` (world-readable/executable), and all data files inside are created with `0o644` (world-readable). Task data may contain sensitive information (personal notes, passwords mentioned in tasks, work-related confidential items).

**Affected locations:**

| File | Line | Permission | Content at Risk |
|------|------|------------|-----------------|
| `internal/core/config_paths.go` | 41 | `0o755` (dir) | Main config directory |
| `internal/adapters/textfile/file_manager.go` | 120 | `0o666→0o644` (os.Create default) | `tasks.yaml` — all user tasks |
| `internal/adapters/textfile/file_manager.go` | 153 | `0o644` | `completed.txt` — completion history |
| `internal/core/sync_log.go` | 51 | `0o644` | `sync.log` — sync history with task text |
| `internal/core/planning_metrics.go` | 34 | `0o644` | `sessions.jsonl` — usage analytics |
| `internal/core/improvement_writer.go` | 16 | `0o644` | Improvement suggestions log |
| `internal/core/provider_config.go` | 200 | `0o666→0o644` | `config.yaml` — provider settings |
| `internal/core/pattern_analyzer.go` | 544 | `0o644` | Pattern analysis cache |
| `internal/dispatch/audit.go` | 58 | `0o644` | Dispatch audit log |

**Contrast with correct usage:** The MCP subsystem correctly uses `0o700` for directories and `0o600` for files:
- `internal/mcp/middleware.go:198` — `0o700` directory
- `internal/mcp/middleware.go:202` — `0o600` files
- `internal/mcp/proposal_store.go:36` — `0o700` directory
- `internal/mcp/proposal_store.go:229` — `0o600` files

Similarly, API provider cache directories correctly use `0o700`:
- `internal/adapters/github/github_provider.go:441`
- `internal/adapters/jira/jira_provider.go:361`
- `internal/adapters/todoist/todoist_provider.go:303`

**Remediation:**
1. Change `~/.threedoors/` creation to `0o700` in `config_paths.go:41`
2. Change all `os.Create()` calls for data files to `os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)`
3. Change all `os.OpenFile(..., 0o644)` calls for logs to `0o600`
4. Consider a migration step to `chmod` existing directories on startup

---

### HIGH-2: No Symlink Validation on File Operations

**Severity:** HIGH
**Category:** File I/O & Path Traversal
**Impact:** Symlink attacks could redirect file writes to arbitrary locations

**Finding:** No code in the codebase checks whether target files or directories are symlinks before reading or writing. The codebase has zero calls to `os.Lstat()`, `filepath.EvalSymlinks()`, or any symlink detection.

**Attack scenario:**
1. Attacker creates `~/.threedoors/` before the victim first runs ThreeDoors
2. Attacker creates `~/.threedoors/tasks.yaml → /target/file`
3. Victim runs ThreeDoors, which writes task data through the symlink
4. Or: attacker creates `~/.threedoors/config.yaml → /attacker/readable/location` to exfiltrate config

**Affected code paths:**
- All atomic write implementations (`path + ".tmp"` pattern)
- `internal/adapters/textfile/file_manager.go:120` — SaveTasks()
- `internal/core/provider_config.go:200` — SaveProviderConfig()
- `internal/core/wal_provider.go:293` — persistWALLocked()
- `internal/core/sync_log.go:51` — Append()

**Note:** The Obsidian adapter has excellent path traversal protection in `sanitizeDailyNotePath()` (`internal/adapters/obsidian/obsidian_daily.go:33-56`), but this protection is specific to Obsidian paths and not applied to the main data directory.

**Remediation:**
1. On startup, verify `~/.threedoors/` is not a symlink using `os.Lstat()`
2. Before atomic writes, verify the target path is not a symlink
3. Use `O_NOFOLLOW` flag where available (platform-dependent)
4. Consider checking directory ownership matches the current user

---

### MEDIUM-1: No Scanner Buffer Size Limits on Most JSONL Readers

**Severity:** MEDIUM
**Category:** Input Validation / DoS
**Impact:** Malformed JSONL files with extremely long lines could cause excessive memory allocation

**Finding:** Most `bufio.NewScanner()` usages rely on the default 64KB buffer limit. While Go's scanner will return an error (not crash) when a line exceeds the buffer, a crafted JSONL file with many long-but-under-limit lines could still consume significant memory.

The MCP transport correctly sets an explicit buffer limit:
- `internal/mcp/transport.go:37` — `scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)` (1MB max)

**Affected scanners without explicit limits:**
- `internal/core/pattern_analyzer.go:152` — sessions JSONL
- `internal/core/pattern_analyzer.go:185` — sessions JSONL reload
- `internal/core/sync_log.go:127` — sync log JSONL
- `internal/core/wal_provider.go:345` — WAL JSONL
- `internal/core/metrics/reader.go:76` — metrics JSONL
- `internal/core/completion_counter.go:55` — completed tasks

**Remediation:** Add explicit buffer limits to all JSONL scanners. The default 64KB is reasonable for this application but should be explicit.

---

### MEDIUM-2: YAML Deserialization Without Size Limits

**Severity:** MEDIUM
**Category:** Input Validation / DoS
**Impact:** A crafted YAML file with deeply nested structures or YAML bombs could cause excessive CPU/memory usage

**Finding:** All YAML unmarshaling uses `yaml.Unmarshal()` without size validation on the input. While `gopkg.in/yaml.v3` has some built-in protections against billion-laughs-style attacks (alias expansion limits), there's no explicit file size check before reading.

**Affected locations:**
- `internal/adapters/textfile/file_manager.go:80` — task YAML (user-editable file)
- `internal/core/provider_config.go:82` — config YAML
- `internal/core/write_queue.go:122` — pending writes YAML
- `internal/dispatch/queue.go:58` — dev queue YAML
- `internal/core/values_config.go:39` — values config
- `internal/core/onboarding.go:28,46` — onboarding config
- `internal/core/dedup_store.go:47` — dedup decisions
- `internal/core/sync_state.go:90` — sync state
- `internal/core/inline_hints_config.go:61,97` — hints config

**Mitigating factors:**
- `gopkg.in/yaml.v3` has built-in YAML bomb protection
- Files are local-only (no network-sourced YAML)
- The task file is the only user-regularly-edited YAML

**Remediation:** Add a size check before `os.ReadFile()` for user-editable files. A 10MB limit would be generous for task data.

---

### MEDIUM-3: Inconsistent Permission Model Across Subsystems

**Severity:** MEDIUM
**Category:** Design / Architecture
**Impact:** Confusing security posture — some subsystems are secure while core data is not

**Finding:** The codebase has two distinct permission models:

| Subsystem | Dir Perms | File Perms | Status |
|-----------|-----------|------------|--------|
| Core data (tasks, config, logs) | 0o755 | 0o644 | Too permissive |
| MCP (audit, proposals) | 0o700 | 0o600 | Correct |
| API provider caches (GitHub, Jira, Todoist) | 0o700 | varies | Correct dirs |
| Enrichment DB | 0o700 | sqlite default | Correct dir |

This inconsistency suggests the MCP and API subsystems were written with security awareness, but the core data layer predates that practice.

**Remediation:** Standardize on `0o700` for directories and `0o600` for files containing user data across all subsystems.

---

### MEDIUM-4: Config File May Store API Credentials in World-Readable File

**Severity:** MEDIUM
**Category:** Credential Management
**Impact:** API tokens stored in config.yaml would be readable by other local users

**Finding:** The sample config generator in `provider_config.go:225-309` includes commented-out fields for API tokens:
```yaml
#       api_token: ""  # Or set JIRA_API_TOKEN env var
#       api_token: ""  # Or set TODOIST_API_TOKEN env var
```

While environment variables are recommended, the config structure accepts tokens directly. Combined with HIGH-1 (0o644 file permissions), any user who stores tokens in config.yaml exposes them.

**Mitigating factors:**
- Documentation recommends environment variables
- Token fields use `yaml:"-"` tag in some backend configs (`internal/intelligence/llm/backend.go:38`)

**Remediation:**
1. Use `0o600` for config.yaml (covered by HIGH-1 fix)
2. Consider adding a startup warning if config.yaml is world-readable and contains non-empty token fields
3. Ensure all token/credential fields use `yaml:"-"` to prevent accidental serialization

---

### LOW-1: Error Messages May Leak Filesystem Paths

**Severity:** LOW
**Category:** Information Disclosure
**Impact:** Error messages displayed in the TUI or written to stderr reveal full filesystem paths

**Finding:** Error messages throughout the codebase include file paths:
- `internal/adapters/textfile/file_manager.go:77` — `fmt.Errorf("failed to read %s: %w", path, err)`
- `internal/tui/search_view.go:360` — `FlashMsg{Text: fmt.Sprintf("Error reading sync log: %v", err)}`
- `internal/tui/conflict_view.go:67` — error messages in flash UI

**Mitigating factors:**
- ThreeDoors is a local TUI application (not a web service)
- The user running the app already knows their own paths
- stderr output is not visible to other users

**Remediation:** No action required for a local TUI. Would need attention if the app ever exposes a network interface.

---

### LOW-2: Atomic Write .tmp Files Briefly World-Readable

**Severity:** LOW
**Category:** File I/O & Permissions
**Impact:** Brief window where .tmp files are world-readable before atomic rename

**Finding:** All atomic write operations create temporary files using `os.Create()` (default 0o666 before umask), write data, then rename. During the write window (typically milliseconds), the .tmp file is world-readable.

**Mitigating factors:**
- Window is extremely brief (write + fsync + rename)
- Requires precise timing to exploit
- Depends on umask setting

**Remediation:** Use `os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)` instead of `os.Create()` for temporary files. This is a low-effort fix that eliminates the window entirely.

---

### LOW-3: bufio.Scanner Default Token Size May Truncate Long Lines

**Severity:** LOW
**Category:** Data Handling
**Impact:** JSONL entries exceeding 64KB would be silently skipped

**Finding:** The default `bufio.Scanner` token size is 64KB. If a JSONL entry somehow exceeds this (e.g., task with very long text after sync), the scanner returns `bufio.ErrTooLong` and stops scanning. Most JSONL parsers in the codebase skip malformed lines gracefully, so this would result in silent data loss rather than a crash.

**Remediation:** Set explicit scanner buffer sizes on all JSONL readers. The MCP transport already does this correctly.

---

## Dependency Audit

### Direct Dependencies Assessment

| Dependency | Version | Maintainer | Status |
|------------|---------|------------|--------|
| `charmbracelet/bubbletea` | v1.3.10 | Charm (well-funded OSS) | Active, trusted |
| `charmbracelet/bubbles` | v1.0.0 | Charm | Active, trusted |
| `charmbracelet/lipgloss` | v1.1.0 | Charm | Active, trusted |
| `fsnotify/fsnotify` | v1.9.0 | Community maintained | Active, widely used |
| `google/go-github/v68` | v68.0.0 | Google | Active, trusted |
| `google/uuid` | v1.6.0 | Google | Active, trusted |
| `spf13/cobra` | v1.10.2 | Community (spf13) | Active, industry standard |
| `golang.org/x/oauth2` | v0.35.0 | Go team | Active, trusted |
| `golang.org/x/term` | v0.40.0 | Go team | Active, trusted |
| `golang.org/x/time` | v0.14.0 | Go team | Active, trusted |
| `gopkg.in/yaml.v3` | v3.0.1 | Canonical (go-yaml) | Stable, widely used |
| `modernc.org/sqlite` | v1.46.1 | Jan Mercl | Active, pure Go (no CGO) |
| `mattn/go-isatty` | v0.0.20 | mattn | Active, widely used |
| `muesli/termenv` | v0.16.0 | muesli (Charm contributor) | Active |

**Assessment:** All dependencies are from well-known, actively maintained sources. No abandoned or suspicious packages. The use of `modernc.org/sqlite` (pure Go) instead of `mattn/go-sqlite3` (CGO) is a security positive — eliminates C memory safety concerns.

**govulncheck:** Not installed in the environment. Recommend running `govulncheck ./...` in CI to catch known CVEs in dependencies.

### Supply Chain Observations

- No vendored dependencies (`vendor/` directory not present) — relies on Go module proxy
- `go.sum` provides integrity verification
- No replace directives in `go.mod` that could redirect to malicious modules

---

## GitHub Actions / CI Supply Chain Assessment

### Actions Used (ci.yml, release.yml, release-verify.yml)

| Action | Version | Pinning | Risk |
|--------|---------|---------|------|
| `actions/checkout` | `@v6` | Tag, not SHA | LOW — first-party GitHub action |
| `actions/setup-go` | `@v6` | Tag, not SHA | LOW — first-party |
| `actions/upload-artifact` | `@v7` | Tag, not SHA | LOW — first-party |
| `actions/download-artifact` | `@v8` | Tag, not SHA | LOW — first-party |
| `actions/cache` | `@v5` | Tag, not SHA | LOW — first-party |
| `actions/github-script` | `@v8` | Tag, not SHA | LOW — first-party |
| `golangci/golangci-lint-action` | `@v9` | Tag, not SHA | MEDIUM — third-party |
| `dorny/paths-filter` | `@v3` | Tag, not SHA | MEDIUM — third-party |
| `docker/setup-buildx-action` | `@v4` | Tag, not SHA | LOW — Docker official |
| `docker/build-push-action` | `@v7` | Tag, not SHA | LOW — Docker official |
| `goreleaser/goreleaser-action` | `@v7` | Tag, not SHA | MEDIUM — third-party |

**Finding:** All actions are pinned to major version tags (e.g., `@v6`), not SHA hashes. This is standard practice but allows tag mutation. For maximum supply chain security, pin to SHA hashes for third-party actions.

**CI Permissions:**
- `ci.yml`: `contents: write`, `pull-requests: write` — needed for coverage comments and artifact uploads
- `release.yml`: `contents: write`, `id-token: write` — appropriate for release creation
- `release-verify.yml`: `contents: read`, `issues: write` — minimal, appropriate

**Secrets handling in CI:** Certificates and signing credentials properly handled via GitHub Secrets, temporary keychain, and cleanup steps. Certificate files are deleted immediately after import (`rm cert.p12`). Keychain is deleted in `always()` cleanup step.

---

## Runtime Attack Scenarios (DAST-style Analysis)

### Scenario 1: Crafted YAML Task File

**Vector:** User manually edits `~/.threedoors/tasks.yaml` or a malicious program modifies it.

**Tested patterns:**
- **YAML bomb (billion laughs):** `gopkg.in/yaml.v3` has built-in alias expansion limits. The library limits alias dereferencing depth, preventing exponential expansion. **Mitigated by library.**
- **Type confusion:** All YAML unmarshaling targets typed Go structs (e.g., `TasksFile`, `ProviderConfig`). Unknown fields are silently ignored by `yaml.v3`. **No injection vector.**
- **Oversized file:** No explicit size check before `os.ReadFile()`. A multi-GB task file could cause OOM. **Partially vulnerable** (see MEDIUM-2).
- **Malformed UTF-8:** Go handles this gracefully — no buffer overflows possible. **Not vulnerable.**

### Scenario 2: Corrupted JSONL Session Logs

**Vector:** Malicious or corrupted `sessions.jsonl`, `sync.log`, WAL files.

**Behavior:** All JSONL parsers gracefully skip malformed lines:
- `pattern_analyzer.go:160` — `continue` on unmarshal error
- `sync_log.go:134` — `continue` on unmarshal error
- `wal_provider.go:355` — `continue` with stderr warning
- `metrics/reader.go:84` — `continue` on unmarshal error

**Assessment:** Excellent resilience. Corrupted logs cause data loss (skipped entries) but never crashes. **Not exploitable.**

### Scenario 3: Malicious TUI Input

**Vector:** User types crafted input into text fields (task text, search, commands).

**Protections:**
- All text inputs have character limits (100-500 chars depending on field)
- Command mode (`:` prefix) only dispatches to internal command handlers — no shell execution
- Task text is stored as-is in YAML — properly escaped by `yaml.v3` marshaling
- No `fmt.Sprintf` calls use user input as format strings — all user data in `%s` argument positions

**Assessment:** **Not exploitable.** Input validation is thorough.

### Scenario 4: AppleScript Injection

**Vector:** Task text containing AppleScript metacharacters passed to `osascript`.

**Protections:** `escapeAppleScript()` function in `internal/calendar/applescript_reader.go:154-157` and `internal/adapters/applenotes/apple_notes_provider.go:238-241` properly escapes backslashes and double quotes before embedding in AppleScript string literals. All `exec.Command` calls use argument arrays, not shell concatenation.

**Assessment:** **Not exploitable.** Proper escaping in place.

### Scenario 5: Local Privilege Escalation via Symlink

**Vector:** Attacker pre-creates `~/.threedoors/` with symlinks to sensitive files.

**Assessment:** **Partially vulnerable** (see HIGH-2). The application does not verify symlink status before file operations. However, exploitation requires:
1. Attacker has write access to the victim's home directory
2. Victim has not yet created `~/.threedoors/`
3. Race condition between directory check and first use

### Scenario 6: MCP Server Input

**Vector:** Malicious JSON-RPC messages sent to the MCP stdio transport.

**Protections:**
- `internal/mcp/transport.go:37` — 1MB max message buffer
- JSON unmarshaling into typed structs
- `internal/mcp/middleware.go` — audit logging with 0o600 permissions
- `internal/mcp/proposal_store.go` — proposals stored with 0o600 permissions

**Assessment:** **Well protected.** The MCP subsystem shows the strongest security practices in the codebase.

---

## Concurrency Safety Assessment

### Mutex Usage — Correct

| Component | Mutex | Protected Resource | Pattern |
|-----------|-------|--------------------|---------|
| WALProvider | `sync.Mutex` | pending entries, nextSeq | Lock/unlock with explicit unlock on early return |
| ObsidianAdapter | `sync.Mutex` | file read-modify-write operations | `defer a.mu.Unlock()` |
| ProposalStore | `sync.Mutex` | proposals map | Standard lock/defer unlock |

All mutex usage is correct. No deadlock risks identified (no nested locks, no lock ordering issues).

### TUI Thread Safety

Bubbletea's single-threaded message loop (`Update()`) handles all state mutations. No concurrent access to TUI state. `View()` methods are read-only. **No race conditions in TUI layer.**

### TOCTOU (Time-of-Check-Time-of-Use)

The atomic write pattern (write to `.tmp`, `fsync`, rename) eliminates TOCTOU for file content. However, the lack of symlink checks (HIGH-2) creates a theoretical TOCTOU between checking if the file exists and writing to it.

---

## Positive Observations

### INFO-1: Excellent Atomic Write Pattern
Every file persistence operation uses the write-sync-rename pattern with proper cleanup on failure. This prevents data corruption on crash or power loss. Consistently applied across all subsystems.

### INFO-2: Proper Error Wrapping with %w
All errors use `fmt.Errorf("context: %w", err)` enabling proper error chain inspection. No string-based error matching. Sentinel errors defined at package level.

### INFO-3: No Panics in User-Facing Code
No `panic()` calls found in TUI `Update()` or `View()` methods. All slice accesses are bounds-checked. Nil checks present where needed.

### INFO-4: Strong Obsidian Path Validation
The `sanitizeDailyNotePath()` function in `obsidian_daily.go:33-56` is a model of path validation: null byte check, `filepath.Clean()`, component-level `..` check, absolute path rejection. This pattern should be applied more broadly.

### INFO-5: Pure Go SQLite (No CGO)
Using `modernc.org/sqlite` eliminates C memory safety concerns that would come with CGO-based SQLite bindings. Good dependency choice.

---

## Recommendations Summary (Priority Order)

| Priority | Finding | Effort | Impact |
|----------|---------|--------|--------|
| 1 | **HIGH-1:** Change `~/.threedoors/` to `0o700`, all data files to `0o600` | Low | Prevents local data exposure |
| 2 | **HIGH-2:** Add symlink validation before file operations | Medium | Prevents symlink attacks |
| 3 | **MEDIUM-4:** Ensure config.yaml uses `0o600`, warn if world-readable with tokens | Low | Protects credentials |
| 4 | **MEDIUM-3:** Standardize permission model across all subsystems | Low | Consistency |
| 5 | **LOW-2:** Use `os.OpenFile(..., 0o600)` instead of `os.Create()` for .tmp files | Low | Eliminates brief exposure window |
| 6 | **MEDIUM-1:** Add explicit scanner buffer limits to all JSONL readers | Low | Defense in depth |
| 7 | **MEDIUM-2:** Add file size check before YAML reads | Low | Prevents OOM on huge files |
| 8 | CI: Pin third-party GitHub Actions to SHA hashes | Low | Supply chain hardening |
| 9 | CI: Add `govulncheck` to quality gate | Low | Catch known CVEs |

---

## Scope Exclusions

- Network infrastructure / cloud deployment security (not applicable — local TUI app)
- Physical security of the device
- macOS Gatekeeper / code signing validation (covered by release workflow, not app code)
- Third-party service APIs (GitHub, Jira, Todoist) — authentication handled correctly via env vars

---

## Methodology

1. Static analysis of all Go source files in `internal/` and `cmd/`
2. Dependency review via `go.mod`
3. GitHub Actions workflow review for supply chain risks
4. Manual code review focused on OWASP patterns adapted for CLI/TUI applications
5. Runtime attack scenario analysis (crafted input, corrupted files, symlink attacks)
6. Concurrency analysis (mutex patterns, race condition assessment)
