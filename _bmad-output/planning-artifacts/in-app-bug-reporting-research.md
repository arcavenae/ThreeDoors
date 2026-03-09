# Research: In-App Bug Reporting via `:bug` Command

**Date:** 2026-03-09
**Status:** Research complete — ready for epic/story creation
**Party Mode:** [in-app-bug-reporting-party-mode.md](in-app-bug-reporting-party-mode.md)

---

## Problem Statement

When ThreeDoors users encounter a bug, the reporting friction is high:
1. Leave the TUI
2. Open browser, navigate to GitHub Issues
3. Manually describe environment (OS, version, terminal)
4. Try to reconstruct what they were doing from memory
5. Write the issue

By step 3, most users give up. Context is lost. Bugs go unreported.

**Goal:** A `:bug` command that lets users report bugs without leaving the TUI, with automatic environment context and a navigation breadcrumb trail — while strictly respecting SOUL.md's privacy-first principles.

---

## SOUL.md Alignment Analysis

### Supports
- **"Friend helping report a problem"** — the `:bug` command is like telling a friend "this broke" and they help you write it up
- **"Reduce friction to starting"** — reporting should be as easy as `:bug it crashed when I pressed Enter`
- **"Every interaction should feel deliberate"** — user initiates report, sees preview, chooses how to send
- **"Local-first, privacy-always"** — data stays on machine until user explicitly sends it

### Tension Points
- **"No telemetry, no analytics, no phone-home"** — breadcrumb tracking could feel like telemetry if not handled carefully
- **Resolution:** Breadcrumbs are memory-only (never persisted), never transmitted without explicit user action, and the user sees exactly what will be sent in a mandatory preview step
- **Key test:** Would a privacy-conscious friend be comfortable with this? Yes — it's the equivalent of "what were you doing when it happened?" not "let me install a keylogger"

### Verdict: ALIGNED
The feature passes the SOUL.md test when implemented with:
1. Mandatory preview before any data leaves the machine
2. Strict allowlist (never blocklist) for included data
3. User controls submission method and timing
4. No persistence of breadcrumbs — memory-only ring buffer

---

## Architecture Analysis

### Current Codebase State

| Component | Status | Relevant Code |
|-----------|--------|---------------|
| Command parsing | Ready | `search_view.go:executeCommand()` — add `:bug` case |
| View system | Ready | `main_model.go` ViewMode enum — add `ViewBugReport` |
| Environment info | Partial | `cli/version.go` has version/commit; needs TUI exposure |
| GitHub client | Partial | `github_client.go` has auth; lacks `CreateIssue()` |
| Breadcrumb system | Missing | Only `previousView` field exists — no trail |
| Browser launch | Missing | Need `os/exec` wrapper for `open`/`xdg-open` |

### Proposed Architecture

```
MainModel.Update()
    │
    ├── BreadcrumbTrail.Record()  ← Always captures (ring buffer)
    │       [50-entry fixed array, view transitions + non-text keys]
    │
    └── ":bug" command
            │
            ├── CollectEnvironment()  ← Version, OS, terminal, counts
            ├── BreadcrumbTrail.Format()  ← Human-readable trail
            └── → ViewBugReport
                    │
                    ├── Description input (textarea)
                    ├── Environment preview (readonly)
                    ├── Breadcrumb preview (readonly)
                    │
                    └── Submit options:
                        ├── [b] Browser URL (zero-auth)
                        ├── [s] GitHub API (if token)
                        └── [f] Save to file (always)
```

---

## Challenge 1: Authentication

### Evaluated Options

| Option | Auth Required | Complexity | Portability | Recommended |
|--------|--------------|------------|-------------|-------------|
| **A. Browser URL** | None (uses browser session) | Low | Universal | **PRIMARY** |
| **B. PAT via config** | `GITHUB_TOKEN` | Low (exists) | Universal | **UPGRADE** |
| C. OAuth Device Flow | One-time browser auth | High | Universal | No — overkill |
| D. `gh` CLI | `gh auth login` | Medium | Requires gh | No — external dep |
| E. Email | None | Low | Requires repo config | No — poor format |

### Recommended: Tiered Approach

**Tier 1 — Browser URL (default, zero-config):**
```
https://github.com/arcaven/ThreeDoors/issues/new?
  title=Bug%3A+<user-title>&
  body=<url-encoded-markdown>&
  labels=type.bug
```

GitHub supports issue creation via URL query parameters. The user's existing browser session provides authentication. URL length limit is ~8,000-16,000 chars in modern browsers — our reports will be well under 2,000 chars.

Platform-specific browser launch:
- macOS: `open <url>`
- Linux: `xdg-open <url>`
- Windows: `start <url>` (future, if needed)

**Tier 2 — GitHub API (automatic upgrade):**
If user has `GITHUB_TOKEN` configured (for Epic 26 GitHub Issues adapter or otherwise), offer direct API submission via `go-github` `Issues.Create()`. User still sees preview first.

**Tier 3 — Local file (offline fallback):**
Save as `~/.threedoors/bug-reports/bug-<timestamp>.md` in GitHub-compatible markdown. User can copy-paste into an issue later, or a future `:bug-submit` command could batch-submit saved reports.

---

## Challenge 2: Privacy — Data Allowlist

### Strict Allowlist (SAFE to include)

| Category | Data | Example |
|----------|------|---------|
| App version | Version, commit, build date | `v0.8.2 (abc1234, 2026-03-09)` |
| Runtime | Go version | `go1.25.4` |
| Platform | OS, architecture | `darwin/arm64` |
| Terminal | Width × height | `120×40` |
| Current state | View mode name | `DoorsView` |
| Theme | Active theme name | `Classic` |
| Task counts | Total tasks, provider count | `12 tasks, 2 providers` |
| Session | Session duration | `14m32s` |
| Navigation | View transitions, non-text keys | `DoorsView → key:Enter → DetailView` |

### Strict Blocklist (NEVER include)

| Category | Why |
|----------|-----|
| Task names/content | Personal data |
| File paths | Reveals directory structure |
| Provider names/config | Could reveal services used |
| Search queries | User text input |
| Any text input content | Privacy violation |
| Tag names/values | Task categorization is personal |
| Username/home directory | PII |
| Credentials/tokens | Security |
| Mood entries | Sensitive personal data |
| Values/goals content | Deeply personal |

### Privacy Implementation

The privacy firewall is at the **breadcrumb capture level**, not the **report generation level**:

```go
// In MainModel.Update():
case tea.KeyMsg:
    if msg.Type == tea.KeyRunes {
        // NEVER record text input — this IS the privacy firewall
        return
    }
    m.breadcrumbs.Record("key:" + msg.String())
```

This means even if a bug in report generation accidentally tried to include text, the data was never captured in the first place. Defense in depth.

---

## Challenge 3: Breadcrumb System Design

### Ring Buffer Specification

```go
const BreadcrumbCapacity = 50

type BreadcrumbEntry struct {
    ViewMode  string    // "DoorsView", "DetailView", etc.
    Action    string    // "key:Enter", "view:Detail", "cmd:stats"
    Timestamp time.Time // UTC
}

type BreadcrumbTrail struct {
    entries [BreadcrumbCapacity]BreadcrumbEntry
    head    int  // next write position
    count   int  // entries used (up to BreadcrumbCapacity)
}
```

### Why Count-Bounded (not Time-Bounded)

| Criterion | Count-bounded (50) | Time-bounded (5 min) |
|-----------|-------------------|---------------------|
| Memory | Fixed, predictable | Variable |
| Context quality | Always 50 actions | Could be 3 or 300 |
| Implementation | Simple modular arithmetic | Timer management |
| Edge cases | None | What if user is idle? |

### Capture Points

1. **View transitions:** When `m.viewMode` changes → `"view:<NewViewName>"`
2. **Non-text keys:** `tea.KeyMsg` where `msg.Type != tea.KeyRunes` → `"key:<keyname>"`
3. **Command execution:** When `:command` is parsed → `"cmd:<command>"` (name only, args stripped)
4. **Window resize:** `tea.WindowSizeMsg` → `"resize:<w>x<h>"`

### Human-Readable Format

```markdown
### Recent Activity (last 50 actions)

| Time | View | Action |
|------|------|--------|
| 14:23:05 | DoorsView | key:a (select door) |
| 14:23:06 | DetailView | view transition |
| 14:23:08 | DetailView | key:Enter |
| 14:23:10 | DoorsView | key:Esc (back) |
| 14:23:12 | DoorsView | cmd:stats |
| 14:23:14 | InsightsView | view transition |
```

---

## Challenge 4: Offline Handling

### Local Save Format

```markdown
# Bug Report — ThreeDoors

**Submitted:** 2026-03-09T14:23:05Z
**Status:** Saved locally (not yet submitted)

## Description

<user's description here>

## Environment

- **ThreeDoors:** v0.8.2 (abc1234, built 2026-03-09)
- **Go:** go1.25.4
- **OS:** darwin/arm64 (macOS 15.3)
- **Terminal:** 120×40
- **Theme:** Classic
- **Tasks:** 12 loaded from 2 providers
- **Session:** 14m32s

## Recent Activity

<breadcrumb table here>
```

### Save Location
`~/.threedoors/bug-reports/bug-<RFC3339-timestamp>.md`

Created on demand (directory auto-created on first save). Files are GitHub-flavored markdown — user can drag-and-drop or copy-paste into a GitHub issue.

### Future Enhancement (Out of Scope)
A `:bug-submit` command could scan `~/.threedoors/bug-reports/` for unsent reports and batch-submit them. This is a natural follow-up but not part of the initial implementation.

---

## Challenge 5: Implementation Scope

### Proposed Epic Structure

**Epic NN: In-App Bug Reporting** (P2)

| Story | Title | Effort | Depends On |
|-------|-------|--------|------------|
| NN.1 | Breadcrumb Tracking System | S | None |
| NN.2 | Bug Report View & Environment Collection | M | NN.1 |
| NN.3 | Submission Methods (Browser, API, File) | M | NN.2 |

### Story Sketches

**NN.1 — Breadcrumb Tracking System**
- `BreadcrumbTrail` ring buffer type with `Record()` and `Format()` methods
- Integration into `MainModel.Update()` at top of function
- Capture view transitions, non-text keys, command names (no args)
- Privacy: `tea.KeyRunes` never captured
- Tests: ring buffer overflow, privacy filtering, format output

**NN.2 — Bug Report View & Environment Collection**
- New `ViewBugReport` mode with text input for description
- `CollectEnvironment()` function gathering safe data per allowlist
- Preview screen showing full report before submission
- Privacy disclaimer visible at all times
- `:bug` command wired into `search_view.executeCommand()`
- Tests: environment collection, view rendering, allowlist verification

**NN.3 — Submission Methods**
- Browser URL construction with proper encoding
- Platform-specific browser launch (`open`/`xdg-open`)
- Optional `GitHubClient.CreateIssue()` when token available
- Local file save to `~/.threedoors/bug-reports/`
- Error handling for each path (browser fail → offer clipboard copy, API fail → offer browser, etc.)
- Tests: URL encoding, file save, error cascading

### Dependencies
- None from other epics — fully standalone
- Enhances Epic 26 (GitHub integration) but doesn't require it

### Technical Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| URL encoding edge cases | Low | Standard `net/url` library handles this |
| Browser launch failure | Low | Fallback to clipboard copy |
| MainModel.Update() bloat | Medium | Single method call, no branching |
| Privacy leak in breadcrumbs | High | Allowlist at capture level (defense in depth) |
| Report too large for URL | Low | Truncate breadcrumbs if > 6KB |

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| BUG-1 | Primary submission | Browser URL (pre-filled) | OAuth device flow, gh CLI, email | Zero-config; no auth required; universal |
| BUG-2 | Auth upgrade path | PAT via existing config | Separate token, OAuth | Reuses Epic 26 infrastructure |
| BUG-3 | Offline fallback | Local markdown file | Queue for later submission | Simple; user controls timing |
| BUG-4 | Breadcrumb storage | Ring buffer, 50 entries | Time-bounded, unbounded, persistent | Fixed memory; sufficient context; privacy-safe |
| BUG-5 | Privacy approach | Strict allowlist at capture level | Blocklist at report level | Defense in depth; can't leak what wasn't captured |
| BUG-6 | Preview requirement | Mandatory before any submission | Optional | SOUL.md trust alignment |
| BUG-7 | Target repo | Hardcoded `arcaven/ThreeDoors` | Configurable | Single-product reporter; YAGNI |
| BUG-8 | Feature priority | P2 | P1 | Valuable but no blocking issues or user requests |
| BUG-9 | Story count | 3 stories | 1 (too large), 5 (too granular) | Clean dependency chain; each independently testable |
| BUG-10 | Report format | GitHub-flavored markdown | JSON, plain text | Renders on GitHub; works across all submission paths |
