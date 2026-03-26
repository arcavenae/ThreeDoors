# Party Mode: In-App Bug Reporting via `:bug` Command

**Date:** 2026-03-09
**Participants:** PM, Architect, UX Designer, Dev
**Topic:** In-app bug reporting for ThreeDoors TUI
**Rounds:** 4

---

## Round 1: Problem Framing & SOUL.md Alignment

### PM
The core problem: when users hit a bug in ThreeDoors, the friction to report it is enormous. They have to leave the TUI, open a browser, navigate to GitHub, create an issue, manually describe their environment, and try to remember what they were doing. By then, context is lost and most bugs go unreported.

The `:bug` command should feel like SOUL.md's "friend helping you" — not corporate telemetry. The key question: can we make bug reporting feel like a natural part of the TUI conversation rather than an external process?

**Risk:** Any implementation that requires GitHub auth setup before first use creates a chicken-and-egg problem — the users who most need to report bugs (new users) are least likely to have configured tokens.

### Architect
SOUL.md is clear: "No telemetry, no analytics, no phone-home." This means:
1. No automatic data collection — everything must be user-initiated
2. No silent network calls — user must explicitly choose to submit
3. Data stays local unless user actively sends it
4. No intermediate services — direct to GitHub or local file

The existing GitHub adapter (Epic 26) uses PAT auth via `GITHUB_TOKEN` env var or config.yaml. But requiring auth for bug reports creates a catch-22: users without GitHub tokens can't report bugs through the app.

**Proposal:** A tiered approach — browser-based (no auth needed) as default, API-based (if token available) as upgrade, local file as fallback.

### UX Designer
The interaction should be:
1. User types `:bug` — opens a bug report view
2. Shows what data will be included (transparency)
3. User writes their description
4. Preview the full report before sending
5. Choose submission method: open in browser / save locally
6. Confirmation with next steps

**Critical UX principle:** The user must see exactly what will be sent BEFORE anything leaves their machine. No surprises. This is the "friend" test — a friend would show you what they're sending on your behalf.

The preview step is non-negotiable. It builds trust and lets users redact anything they're uncomfortable with.

### Dev
Looking at the codebase:
- `:bug` fits naturally into `search_view.executeCommand()` switch
- A new `ViewBugReport` mode would follow existing patterns
- Environment data is available via `cli/version.go` build vars + `runtime` package
- GitHub client exists but lacks `CreateIssue()` — would need to add it
- No breadcrumb system exists today — `previousView` is single-level only

**Concern:** Adding a breadcrumb tracking system is non-trivial. It touches `MainModel.Update()` which processes every keystroke. Must be zero-cost when not capturing.

---

## Round 2: Authentication Deep Dive

### Architect
I evaluated five authentication approaches for issue creation:

**Option A: Pre-filled Browser URL (RECOMMENDED)**
- GitHub supports URL query params: `https://github.com/owner/repo/issues/new?title=...&body=...`
- Zero auth required — uses whatever GitHub session the browser has
- Body is URL-encoded; limited to ~8KB URL length (plenty for bug reports)
- Works on any OS via `open` (macOS), `xdg-open` (Linux), `start` (Windows)
- Fallback: copy URL to clipboard if browser launch fails

**Option B: PAT via existing config**
- Already have `GITHUB_TOKEN` / config.yaml infrastructure from Epic 26
- Would need `CreateIssue()` method on `GitHubClient`
- Best UX for users who already have GitHub adapter configured
- But creates auth barrier for everyone else

**Option C: OAuth Device Flow**
- GitHub supports device authorization grant (RFC 8628)
- User gets a code, opens browser, enters code at github.com/login/device
- Token stored locally for future use
- Complex implementation; overkill for bug reports

**Option D: GitHub CLI (`gh`)**
- Shell out to `gh issue create`
- Already handles auth, but requires gh CLI installed
- Not portable; SOUL.md says "no dependencies on external tools"

**Option E: Email-based**
- Some repos support email-to-issue
- Requires repo configuration; not universal
- Poor formatting control

**Recommendation:** Option A (browser URL) as primary, Option B (PAT) as automatic upgrade when token exists, local file save as offline fallback.

### PM
I agree with tiered approach. The browser URL method is genius for zero-config:
- New users can report bugs immediately
- Power users with GitHub tokens get seamless API submission
- Offline users save to file for later

**Concern about URL length:** Bug report body with environment info + breadcrumbs + description could exceed URL limits in edge cases. We need a truncation strategy.

### Dev
URL length limits in practice:
- Modern browsers handle 8,000-16,000 characters in URLs
- GitHub's issue creation endpoint handles long URLs well
- Our report would be: ~200 chars env info + ~500 chars breadcrumbs + user description
- Even with a long description, we're well under limits

For the PAT upgrade path, adding `CreateIssue()` to the existing `GitHubClient` is straightforward — the `go-github` library already has `Issues.Create()`.

### UX Designer
The submission flow should be:
1. If PAT is configured → offer "Submit directly" as primary action
2. Always offer "Open in browser" as alternative
3. Always offer "Save to file" as fallback
4. Show clear feedback on what happened

The text should be warm, not corporate:
- "Report saved to ~/.threedoors/bug-reports/" not "Error report archived to local filesystem"
- "Opening GitHub in your browser..." not "Launching default browser for issue creation"

---

## Round 3: Privacy & Breadcrumb Design

### PM
Privacy requirements are absolute per SOUL.md. Let me define the allowlist:

**SAFE to include (allowlist):**
- ThreeDoors version, commit hash, build date
- Go version
- OS name + version (runtime.GOOS, runtime.GOARCH)
- Terminal dimensions (width × height)
- Current view mode name (e.g., "DoorsView", "DetailView")
- Navigation breadcrumbs: view transitions only
- Non-text keystrokes: arrows, Enter, Esc, Tab, function keys
- Active theme name
- Number of tasks loaded (count only, not names/content)
- Number of providers configured (count only, not names)
- Session duration

**NEVER include (blocklist):**
- Task names, descriptions, or any task content
- File paths (task files, config paths)
- Provider names or configuration details
- Search queries typed by user
- Any text input content
- Tag names or values
- User's system username or home directory
- GitHub token or any credentials
- Mood entries
- Values/goals content

### Architect
For the breadcrumb system, I propose a **ring buffer** approach:

```go
type Breadcrumb struct {
    ViewMode  ViewMode
    Action    string    // "enter", "back", "key:Enter", "key:Esc"
    Timestamp time.Time
}

type BreadcrumbTrail struct {
    entries [50]Breadcrumb  // Fixed-size ring buffer, last 50 actions
    head    int
    count   int
}
```

**Design decisions:**
- **Count-bounded, not time-bounded** — 50 entries captures enough context without memory concerns. Time-bounded is harder to reason about.
- **Fixed array, not slice** — zero allocations after init, no GC pressure
- **Record in MainModel.Update()** — single capture point for all view transitions
- **Only record view transitions and non-text keys** — skip text input entirely
- **No goroutines** — synchronous append in the Update loop (single append is nanoseconds)

### Dev
Implementation approach for breadcrumb capture:

```go
// In MainModel.Update(), after processing:
func (m *MainModel) recordBreadcrumb(action string) {
    m.breadcrumbs.Add(Breadcrumb{
        ViewMode:  m.viewMode,
        Action:    action,
        Timestamp: time.Now().UTC(),
    })
}
```

Capture points:
1. View transitions: when `m.viewMode` changes → record `"view:DetailView"`
2. Non-text key events: `tea.KeyMsg` where `msg.Type != tea.KeyRunes` → record `"key:Enter"`
3. Command execution: when `:command` is parsed → record `"cmd:stats"` (command name only, no args)

**Critical:** Text input keys (`tea.KeyRunes`) are NEVER recorded. This is the privacy firewall.

### UX Designer
The breadcrumb display in the bug report should be human-readable:

```
Recent Activity (last 50 actions):
  [14:23:05] Doors View → selected door
  [14:23:06] Detail View → pressed Enter
  [14:23:08] Detail View → pressed Esc
  [14:23:08] Doors View → typed command
  [14:23:10] Search View → executed :stats
  [14:23:12] Insights View → pressed Esc
  [14:23:12] Doors View
```

Note: "typed command" not ":bug" — we don't even reveal the command prefix to avoid leaking search queries that started with `:`.

Wait, actually commands are fine to show by name since they're a fixed set. Only the *arguments* to commands should be hidden. Revised:
- `:stats` → show as `cmd:stats`
- `:add buy milk` → show as `cmd:add` (args stripped)
- `:tag work` → show as `cmd:tag` (args stripped)

---

## Round 4: Offline, Scope, and Implementation Strategy

### PM
**Offline handling:**
Save bug reports as timestamped markdown files in `~/.threedoors/bug-reports/`:
```
~/.threedoors/bug-reports/
  bug-2026-03-09T14-23-05Z.md
```

Format: GitHub-compatible markdown so users can copy-paste into an issue later, or we could add a `:bug-submit` command to submit saved reports when online.

**Scope recommendation:** This is a P2 feature. It's valuable but not blocking anything. I'd estimate 3 stories:
1. Breadcrumb tracking system (ring buffer in MainModel)
2. Bug report view (`:bug` command, environment collection, preview, description input)
3. Submission methods (browser URL, API if token, local save)

### Architect
**Implementation dependencies:**
- Story 1 (breadcrumbs) has zero dependencies — pure internal addition
- Story 2 (view) depends on Story 1, plus needs `dist/version.go` exposed to TUI
- Story 3 (submission) depends on Story 2, optionally extends `GitHubClient` with `CreateIssue()`

**Architecture decisions:**

1. **Ring buffer size: 50 entries** — captures ~2-5 minutes of interaction. Enough for context, not enough to be surveillance.

2. **Breadcrumb as opt-in at code level** — the breadcrumb system always runs (it's just a ring buffer in memory), but it's only *read* when the user explicitly invokes `:bug`. No persistence, no logging, no export except through the bug report flow.

3. **Report format: GitHub-flavored markdown** — renders correctly in browser URL, API submission, and local file save. Single format for all paths.

4. **Target repo: hardcoded `arcavenae/ThreeDoors`** — this is ThreeDoors' own bug reporter. No configuration needed. If someone forks and wants to report to their own repo, they can change it in code.

### Dev
**Effort estimate:**
- Breadcrumb system: small (ring buffer + capture points in Update)
- Bug report view: medium (new ViewMode, text input for description, preview rendering)
- Submission: medium (URL construction, browser launch, optional API, file save)

**Technical risks:**
1. URL encoding of markdown in browser URLs — need to handle special characters
2. `os/exec` for browser launch varies by platform — but Go has good cross-platform support
3. The MainModel.Update() function is already large (1600+ lines) — adding breadcrumb recording needs to be minimal

**Recommendation:** Add breadcrumb recording as a single method call at the TOP of Update(), before the switch statement. This ensures every message is captured regardless of which view handles it.

### UX Designer
**Final interaction flow:**

```
User types: :bug

┌─────────────────────────────────────────┐
│  🐛 Report a Bug                        │
│                                         │
│  What went wrong?                       │
│  ┌─────────────────────────────────────┐│
│  │ (text input area)                   ││
│  │                                     ││
│  └─────────────────────────────────────┘│
│                                         │
│  The following will be included:        │
│  • ThreeDoors v0.8.2 (abc1234)         │
│  • macOS 15.3 (arm64), Go 1.25.4      │
│  • Terminal: 120×40                     │
│  • Current view: DoorsView             │
│  • Last 50 navigation actions          │
│  • Theme: Classic                       │
│  • 12 tasks loaded, 2 providers        │
│                                         │
│  ⚠ No task names, content, or personal │
│    data will be included.              │
│                                         │
│  [Enter] Preview & Submit              │
│  [Esc] Cancel                          │
└─────────────────────────────────────────┘
```

Then preview screen shows the full markdown, with options:
- `[b]` Open in browser (default — works for everyone)
- `[s]` Submit via GitHub API (only shown if token configured)
- `[f]` Save to file
- `[Esc]` Cancel

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| PM-1 | Submission method | Browser URL (primary), API (upgrade), file (fallback) | OAuth device flow (too complex), gh CLI (external dependency), email (poor formatting) | Zero-config primary path; progressive enhancement for power users |
| PM-2 | Feature priority | P2 | P1 | Valuable but not blocking; no user requests yet |
| ARCH-1 | Breadcrumb storage | Ring buffer, 50 entries, count-bounded | Time-bounded (harder to reason about), unbounded (memory risk) | Fixed memory, sufficient context, zero allocations |
| ARCH-2 | Target repo | Hardcoded `arcavenae/ThreeDoors` | Configurable (YAGNI) | Single-product reporter; forks can change in code |
| ARCH-3 | Report format | GitHub-flavored markdown | JSON (not human-readable), plain text (poor formatting) | Works across all submission paths; renders on GitHub |
| ARCH-4 | Breadcrumb capture | Always-on ring buffer, read only on :bug | Opt-in toggle (complexity), persistent log (privacy) | Zero-cost memory-only buffer; no persistence without user action |
| UX-1 | Preview requirement | Mandatory preview before any submission | Direct submit (trust violation) | SOUL.md alignment — user must see what's sent |
| UX-2 | Privacy display | Explicit allowlist shown in report view + blocklist disclaimer | Hidden details (erodes trust) | Transparency builds trust; friend-helping-friend feel |
| UX-3 | Breadcrumb display format | Human-readable timestamps + view names | Raw data dump (intimidating), no breadcrumbs shown (loses value) | Users can understand and verify what's included |
| DEV-1 | Breadcrumb integration point | Top of MainModel.Update(), single method call | Per-view recording (fragile), middleware (overengineered) | Centralizes capture; impossible to miss events |
| DEV-2 | Text input privacy | Never record tea.KeyRunes | Record all keys with redaction (fragile) | Blocklist approach risks leaks; allowlist is safer |
| DEV-3 | Story decomposition | 3 stories: breadcrumbs, view, submission | Single story (too large), 5 stories (too granular) | Clean dependency chain; each is independently testable |
