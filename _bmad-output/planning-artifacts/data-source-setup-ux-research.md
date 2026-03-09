# Data Source Setup UX Research — Full Lifecycle Management

**Date:** 2026-03-09
**Requested by:** Project Owner (via supervisor)
**Party Mode Participants:** PM (John), Architect (Winston), UX Designer (Sally), Dev (Amelia), Innovation Strategist (Victor)
**Scope:** Research only — no implementation code

---

## Executive Summary

ThreeDoors has a mature adapter framework (8 working providers, multi-source aggregation, registry/factory pattern) but lacks the **connection lifecycle layer** — the infrastructure and UX for users to set up, monitor, manage, and troubleshoot their data source integrations. This report covers prior art, data source feasibility, TUI/CLI design patterns, credential security, and a recommended epic breakdown.

**Key finding:** The adapters exist. What's missing is the ConnectionManager infrastructure and the Sources UI/CLI. That's the real work — not more adapters.

---

## 1. Prior Art Survey — TUI/CLI Setup Wizards

### 1.1 GitHub CLI (`gh auth login`)

**Pattern:** Interactive OAuth device code flow
- Prompts: hostname → git protocol → SSH key → auth method (browser OAuth vs paste token)
- Device code flow: CLI generates one-time code, opens browser to `github.com/login/device`, polls until confirmed
- `--with-token` flag for non-interactive/CI use

**Token Storage:**
- Primary: System credential store (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Fallback: Plain text `~/.config/gh/hosts.yml` with warning
- `--insecure-storage` flag to force plain text

**Key takeaway:** Progressive disclosure (start with simplest question, branch). Device code flow avoids callback server complexity.

### 1.2 AWS CLI (`aws configure`)

**Pattern:** Step-by-step sequential prompts
```
AWS Access Key ID [None]: AKIA...
AWS Secret Access Key [None]: wJal...
Default region name [None]: us-east-1
Default output format [None]: json
```

**Multi-instance:** Named profiles (`--profile staging`) in `~/.aws/credentials` + `~/.aws/config`. Active profile selected via `--profile` flag or `AWS_PROFILE` env var.

**Key takeaway:** Separation of secrets from configuration into distinct files. Named profiles for unlimited multi-account support. Environment variable override for session-level switching.

### 1.3 Google Cloud (`gcloud init`)

**Pattern:** Multi-step wizard with branching
1. Configuration selection (create new or re-initialize)
2. Account authorization (browser OAuth)
3. Account selection from authenticated list
4. Project selection
5. Compute defaults (optional)

**Named configurations** as first-class concept — each bundles account + project + region. `gcloud config configurations activate <name>` switches entire context.

**Key takeaway:** Named configurations that bundle all context together. Wizard adapts steps based on what's already configured.

### 1.4 Railway CLI (`railway login`)

**Pattern:** Minimal friction browser auth
- Default: opens browser directly, done
- Headless: `--browserless` displays pairing code + URL (device code flow)
- CI: `RAILWAY_TOKEN` env var skips interactive login entirely
- Security: OAuth 2.0 + PKCE (no client secret needed for CLI)

**Key takeaway:** Absolute minimal friction for happy path. PKCE for security without embedded secrets.

### 1.5 Terraform Providers

**Pattern:** Declarative requirements + imperative credential setup
- `.tf` files declare required providers
- `terraform init` downloads and verifies provider binaries
- Credentials via env vars, shared credential files, or `terraform login`
- Lock file (`.terraform.lock.hcl`) for reproducibility

**Credential priority:** env vars > CLI config > credentials helpers

**Key takeaway:** Declarative provider requirements in code, imperative credential setup outside code. No credentials in version control by design.

### 1.6 lazygit / lazydocker

**Pattern:** External YAML config, no in-app settings panels
- Config: `~/.config/lazygit/config.yml`
- In-app: press `o` to open config in editor, restart to apply
- `?` key: contextual help showing available keybindings

**Key takeaway:** Configuration-as-code. Contextual `?` help for discoverability.

### 1.7 charmbracelet/huh (Go Form Library)

**Components:** Input, Text, Select, MultiSelect, Confirm, Note, FilePicker
**Architecture:** `Form → Group[] → Field[]` — Groups function as wizard pages
**Dynamic forms:** `WithHideFunc()` for conditional pages, `*Func` variants for dependent fields
**Validation:** Per-field `Validate(func(string) error)` with real-time feedback
**Bubbletea integration:** Implements `tea.Model` — embeds directly in existing Bubbletea apps
**Themes:** Default, Charm, Dracula, Catppuccin, Base16

**Key takeaway:** Perfect fit for ThreeDoors. Native Bubbletea integration, Group-as-page model maps directly to setup wizard steps, dynamic forms allow provider-specific field rendering.

### 1.8 Go Keyring Libraries

| Library | API | Backends | Notes |
|---------|-----|----------|-------|
| `zalando/go-keyring` | Simple Set/Get/Delete | macOS Keychain, Linux Secret Service, Windows Cred Mgr | Used by GitHub CLI |
| `99designs/keyring` | Open → Ring → Set/Get | All above + KWallet, Pass, encrypted JWT files, KeyCtl | Built for aws-vault, more backends |

**Recommendation:** `99designs/keyring` for broader backend support, especially the encrypted file fallback for headless environments.

### 1.9 Connection Lifecycle Patterns Across CLIs

**Multi-instance:** Named profiles (AWS), named configurations (gcloud), hostname-based (gh). Common pattern: config file with named sections, env var override, CLI flag override.

**Token lifecycle:**
- Access tokens auto-refresh silently
- When refresh token expires → explicit user action required with clear error
- CLI never silently re-authenticates — user must explicitly re-login

**Health monitoring:** "whoami" or status commands (e.g., `aws sts get-caller-identity`, `gcloud auth list`). Lightweight verification of active identity and token validity.

**Sync status:** On-demand via status commands (CLIs) or continuous in status bars (TUI tools).

---

## 2. Data Source Feasibility Matrix

### Summary Table

| # | Data Source | Feasibility | Auth | API Type | Go SDK | Bidirectional | Change Detection | Multi-Instance |
|---|------------|-------------|------|----------|--------|--------------|-----------------|---------------|
| 1 | **Jira** | Medium | OAuth 2.0 (3LO), API tokens | REST v3 | Community (go-jira, go-atlassian) | Yes | Webhooks + polling | Yes (Cloud + Server) |
| 2 | **Apple Notes** | Hard | macOS permissions | AppleScript | None (exec osascript) | Limited | Polling only | No |
| 3 | **Apple Reminders** | Medium | macOS permissions | EventKit via cgo | Community (go-eventkit) | Yes | EventKit notifications | Multiple lists |
| 4 | **Todoist** | Easy | OAuth 2.0, Bearer tokens | REST v1 (new unified) | None official; thin HTTP | Yes | Webhooks | Per-account |
| 5 | **Linear** | Medium | OAuth 2.0, API keys | GraphQL | Minimal community | Yes | Webhooks | Per-workspace |
| 6 | **GitHub Issues** | Easy | OAuth 2.0, PATs, Apps | REST v3 + GraphQL v4 | Official (google/go-github) | Yes | Webhooks + ETags | Yes (repos + Enterprise) |
| 7 | **Notion** | Medium | OAuth 2.0, integration tokens | REST | Community (jomei/notionapi) | Yes | Webhooks (new 2025) | Per-workspace |
| 8 | **Trello** | Easy | OAuth 1.0a, key+token | REST | Community (adlio/trello) | Yes | Webhooks | Per-board |
| 9 | **Asana** | Medium | OAuth 2.0, PATs | REST | Community (unmaintained) | Yes | Webhooks | Per-workspace |
| 10 | **ClickUp** | Medium | OAuth 2.0, API tokens | REST v2 | Community (raksul/go-clickup) | Yes | Webhooks | Per-workspace |
| 11 | **MS To Do** | Hard | OAuth 2.0 (Azure AD/MSAL) | REST (MS Graph) | Official (msgraph-sdk-go) | Yes | Webhooks (needs HTTPS) | Per-tenant |
| 12 | **Google Tasks** | Medium | OAuth 2.0 | REST | Official (google-api-go-client) | Yes | Polling only (no webhooks) | Per-account |
| 13 | **Plain text/MD** | Easy | N/A (local) | File I/O | fsnotify | Yes | fsnotify | Multiple paths |
| 14 | **Obsidian** | Done | N/A (local) | File I/O + optional REST plugin | N/A | Yes | fsnotify | Multiple vaults |

### Priority Recommendations (Party Mode Consensus)

**Tier 1 — Complete in-progress work:**
1. Todoist (Epic 25, 3/4 done)
2. Linear (Epic 30, stories defined)

**Tier 2 — Highest new user value:**
3. GitHub Issues — developer audience, official Go SDK, mature API
4. Apple Reminders — personal task users on macOS, good data model fit
5. Notion — massive user base, but schema mapping complexity

**Tier 3 — Enterprise unlock:**
6. Jira — already implemented (Epic 19)! Focus on lifecycle management
7. Asana, ClickUp, Trello — "nice to have" for broader appeal

**Tier 4 — Low ROI:**
8. Apple Notes — no task model, fragile AppleScript, macOS-only. **Not recommended.**
9. Microsoft To Do — Azure AD auth is disproportionately complex for a CLI tool
10. Google Tasks — no webhooks, barely any metadata (no priority, no labels)

### Key Feasibility Notes

**Already implemented adapters:** textfile, applenotes, jira, github, obsidian, reminders, todoist (3/4)

**OAuth vs API Token considerations:**
- OAuth: better UX (click-authorize-done), but requires device code flow implementation
- API tokens: simpler to implement, but worse UX (user must find token in settings)
- Recommendation: support both, prefer OAuth where available

**Rate limit concerns:**
- Notion: 3 req/sec — requires aggressive caching
- ClickUp: 100 req/min — restrictive for bulk operations
- Todoist: ~60 req/min — adequate for personal use
- GitHub: 5,000 req/hr — generous
- Linear: 250,000 points/hr — very generous

---

## 3. TUI Design — Setup Wizard & Sources Dashboard

### 3.1 Setup Wizard (`:connect` command)

Uses `charmbracelet/huh` with Groups-as-pages pattern.

**Step 1 — Provider Selection:**
```
┌─────────────────────────────────────────────┐
│          🔗 Connect a Data Source            │
│                                             │
│  What would you like to connect?            │
│                                             │
│  > ● Todoist                                │
│    ○ GitHub Issues                          │
│    ○ Jira Cloud                             │
│    ○ Jira Server/Data Center                │
│    ○ Apple Reminders                        │
│    ○ Linear                                 │
│    ○ Notion                                 │
│    ○ Trello                                 │
│    ○ Plain text files                       │
│                                             │
│  [Enter] Next  [Esc] Cancel                 │
└─────────────────────────────────────────────┘
```

**Step 2a — API Token Flow (Todoist, Linear API key):**
```
┌─────────────────────────────────────────────┐
│          🔗 Connect Todoist (2/4)           │
│                                             │
│  Give this connection a name:               │
│  ┌─────────────────────────────────────┐    │
│  │ Personal Todoist                    │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  API Token:                                 │
│  ┌─────────────────────────────────────┐    │
│  │ ••••••••••••••••••••                │    │
│  └─────────────────────────────────────┘    │
│  (Find at: Settings → Integrations →       │
│   Developer → API token)                    │
│                                             │
│  [Enter] Next  [Esc] Cancel                 │
└─────────────────────────────────────────────┘
```

**Step 2b — OAuth Device Code Flow (GitHub, Jira Cloud, Linear OAuth):**
```
┌─────────────────────────────────────────────┐
│          🔗 Connect GitHub Issues (2/4)     │
│                                             │
│  Opening your browser for authentication... │
│                                             │
│  If the browser doesn't open, visit:        │
│  https://github.com/login/device            │
│                                             │
│  Enter this code: ABCD-1234                 │
│                                             │
│  ⠋ Waiting for authorization...             │
│                                             │
│  [Esc] Cancel                               │
└─────────────────────────────────────────────┘
```

**Step 2c — Local Source (Plain text, Apple Reminders):**
```
┌─────────────────────────────────────────────┐
│       🔗 Connect Apple Reminders (2/3)      │
│                                             │
│  Give this connection a name:               │
│  ┌─────────────────────────────────────┐    │
│  │ Home Reminders                      │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  Which reminder lists to sync?              │
│  [x] Personal                               │
│  [x] Shopping                               │
│  [ ] Work (already connected via Jira)      │
│  [ ] Birthdays                              │
│                                             │
│  [Enter] Next  [Esc] Cancel                 │
└─────────────────────────────────────────────┘
```

**Step 3 — Sync Configuration:**
```
┌─────────────────────────────────────────────┐
│          🔗 Connect Todoist (3/4)           │
│                                             │
│  Sync mode:                                 │
│  > ● Read & write (bidirectional)           │
│    ○ Read only (import tasks, don't push)  │
│                                             │
│  Which projects?                            │
│  [x] Inbox                                  │
│  [x] Work                                   │
│  [ ] Someday/Maybe                          │
│                                             │
│  Poll interval: [30s]                       │
│                                             │
│  [Enter] Next  [Esc] Cancel                 │
└─────────────────────────────────────────────┘
```

**Step 4 — Test & Confirm:**
```
┌─────────────────────────────────────────────┐
│          🔗 Connect Todoist (4/4)           │
│                                             │
│  Testing connection...                      │
│                                             │
│  ✓ API reachable                            │
│  ✓ Token valid (user: arcaven@example.com)  │
│  ✓ Found 23 tasks across 2 projects         │
│                                             │
│  Ready to connect "Personal Todoist"        │
│                                             │
│  [Enter] Connect  [Esc] Cancel              │
└─────────────────────────────────────────────┘
```

### 3.2 Sources Dashboard (`:sources` command)

**List View:**
```
┌─────────────────────────────────────────────────────────┐
│  📡 Connected Sources                         [?] Help  │
│─────────────────────────────────────────────────────────│
│                                                         │
│  ● Work Jira          ✓ Synced 2m ago    142 tasks     │
│  ● Personal Todoist   ✓ Synced 30s ago    23 tasks     │
│  ● OSS GitHub Issues  ✓ Synced 5m ago     17 tasks     │
│  ○ Home Reminders     ⏸ Paused            31 tasks     │
│  ⚠ Linear (Team)      ⚠ Auth expired       0 tasks     │
│                                                         │
│  [a] Add source  [d] Disconnect  [p] Pause/Resume      │
│  [r] Re-sync     [t] Test        [Enter] Details       │
│  [Esc] Back                                             │
└─────────────────────────────────────────────────────────┘
```

**Status indicators:**
- `●` Green = connected and syncing normally
- `○` Gray = paused by user
- `⚠` Yellow = needs attention (auth expired, persistent errors, rate limited)
- `✗` Red = disconnected or unreachable

**Detail View (Enter on a source):**
```
┌─────────────────────────────────────────────────────────┐
│  📋 Personal Todoist                          [?] Help  │
│─────────────────────────────────────────────────────────│
│                                                         │
│  Status:      ● Connected                               │
│  Last sync:   30 seconds ago                            │
│  Tasks:       23 active, 47 completed                   │
│  Projects:    "Work", "Home", "Side Projects"           │
│  Sync mode:   Bidirectional                             │
│  Poll rate:   Every 30s                                 │
│                                                         │
│  Health Checks:                                         │
│   ✓ API reachable         ✓ Token valid                │
│   ✓ Rate limit OK (42/60) ✓ Cache fresh                │
│                                                         │
│  [e] Edit settings  [r] Re-authenticate                 │
│  [p] Pause sync     [d] Disconnect                      │
│  [l] View sync log  [Esc] Back                          │
└─────────────────────────────────────────────────────────┘
```

**Sync Log View:**
```
┌─────────────────────────────────────────────────────────┐
│  📋 Personal Todoist — Sync Log               [?] Help  │
│─────────────────────────────────────────────────────────│
│                                                         │
│  14:32:01  ✓ Synced 23 tasks (2 new, 1 completed)      │
│  14:31:31  ✓ Synced 22 tasks (no changes)              │
│  14:31:01  ✓ Synced 22 tasks (1 updated remotely)      │
│  14:30:15  ⚠ Conflict: "Buy groceries" modified both   │
│               locally and remotely — kept remote        │
│  14:30:01  ✓ Synced 22 tasks (no changes)              │
│  14:29:05  ✓ Completed "Review PR #42" synced to API   │
│                                                         │
│  [Esc] Back                                             │
└─────────────────────────────────────────────────────────┘
```

### 3.3 Main View Integration

**Status bar indicator in doors view:**
```
┌─────────────────────────────────────────────┐
│  [Door A]    [Door B]    [Door C]           │
│  Buy milk    Fix bug     Write docs         │
│                                             │
│  ─────────────────────────────────────────  │
│  ⚠ Linear auth expired — :sources to fix   │
└─────────────────────────────────────────────┘
```

Only shown when a source needs attention. Fades or dismisses like existing key hint indicators.

### 3.4 Disconnection Flow

```
┌─────────────────────────────────────────────┐
│       🔗 Disconnect "Work Jira"?            │
│                                             │
│  This will stop syncing with Work Jira.     │
│                                             │
│  What should happen to synced tasks?        │
│  > ● Keep tasks locally (mark as local)     │
│    ○ Remove synced tasks                    │
│                                             │
│  [Enter] Disconnect  [Esc] Cancel           │
└─────────────────────────────────────────────┘
```

### 3.5 Re-Authentication Flow

When a connection enters `AuthExpired` state, pressing `r` on the source triggers re-auth:

```
┌─────────────────────────────────────────────┐
│     🔗 Re-authenticate Linear (Team)        │
│                                             │
│  Your Linear authorization has expired.     │
│                                             │
│  Opening your browser for re-authorization..│
│                                             │
│  If the browser doesn't open, visit:        │
│  https://linear.app/oauth/authorize?...     │
│                                             │
│  ⠋ Waiting for authorization...             │
│                                             │
│  [Esc] Cancel                               │
└─────────────────────────────────────────────┘
```

---

## 4. CLI Design — Non-Interactive Management

### 4.1 Command Structure

```bash
# === SETUP ===
threedoors connect <provider>              # Interactive wizard
threedoors connect todoist --label "Personal" --token $TOKEN
threedoors connect github --label "OSS" --repos owner/repo1,owner/repo2
threedoors connect jira --label "Work" --server https://jira.company.com --token $TOKEN
threedoors connect reminders --label "Home" --lists "Personal,Shopping"
threedoors connect textfile --label "Notes" --path ~/tasks.yaml

# === LIST & STATUS ===
threedoors sources                         # List all connections (summary)
threedoors sources status                  # Detailed status of all connections
threedoors sources status "Work Jira"      # Detailed status of one connection

# === MANAGEMENT ===
threedoors sources test "Work Jira"        # Verify connectivity and auth
threedoors sources sync "Work Jira"        # Force immediate sync
threedoors sources pause "Home Reminders"  # Pause sync
threedoors sources resume "Home Reminders" # Resume sync
threedoors sources edit "Personal Todoist" # Re-enter setup wizard for this source
threedoors sources reauth "Linear (Team)"  # Re-authenticate (OAuth flow)
threedoors sources disconnect "Linear"     # Remove connection (prompts for task handling)
threedoors sources disconnect "Linear" --keep-tasks  # Non-interactive disconnect

# === SYNC LOG ===
threedoors sources log "Personal Todoist"            # Recent sync events
threedoors sources log "Personal Todoist" --last 50  # Last 50 events
threedoors sources log --errors                      # Only errors across all sources

# === JSON OUTPUT (for scripting) ===
threedoors sources --json                  # Machine-readable status
threedoors sources status "Work Jira" --json
```

### 4.2 Example CLI Output

```
$ threedoors sources
NAME                 PROVIDER    STATUS      LAST SYNC    TASKS
Work Jira            jira        connected   2m ago       142
Personal Todoist     todoist     connected   30s ago       23
OSS GitHub Issues    github      connected   5m ago        17
Home Reminders       reminders   paused      2h ago        31
Linear (Team)        linear      auth expired never         0

$ threedoors sources status "Work Jira"
Name:         Work Jira
Provider:     jira
Status:       ● Connected
Server:       https://jira.company.com
Last Sync:    2 minutes ago (14:30:01 UTC)
Tasks:        142 active, 891 completed
JQL Filter:   assignee = currentUser() AND status != Done
Sync Mode:    Bidirectional
Poll Rate:    Every 60s

Health Checks:
  ✓ API reachable
  ✓ Token valid (expires in 29 days)
  ✓ Rate limit OK (120/450 points used)
  ✓ Cache fresh (142 tasks cached)

$ threedoors sources test "Work Jira"
Testing Work Jira...
  ✓ DNS resolution: jira.company.com → 203.0.113.42
  ✓ TLS handshake: valid certificate (expires 2027-01-15)
  ✓ Authentication: valid API token
  ✓ Authorization: 3 accessible projects
  ✓ Rate limit: 330/450 points remaining
Connection healthy.
```

### 4.3 Environment Variable Overrides

```bash
# Provider-specific token overrides (for CI/automation)
THREEDOORS_TODOIST_TOKEN=xxx
THREEDOORS_JIRA_TOKEN=xxx
THREEDOORS_GITHUB_TOKEN=xxx     # Also respects GH_TOKEN / GITHUB_TOKEN
THREEDOORS_LINEAR_TOKEN=xxx
THREEDOORS_NOTION_TOKEN=xxx

# Connection-specific overrides (by label, slugified)
THREEDOORS_CONN_WORK_JIRA_TOKEN=xxx
THREEDOORS_CONN_PERSONAL_TODOIST_TOKEN=xxx
```

### 4.4 Config File Format

```yaml
# ~/.threedoors/config.yaml
schema_version: 3  # Bump from 2 for connection support

# Legacy single-provider fields preserved for backward compatibility
provider: textfile

# New: Named connections (replaces providers[] list)
connections:
  - id: "conn_01HXYZ"
    provider: todoist
    label: "Personal Todoist"
    settings:
      project_ids: "2331452,8842901"
      filter: ""
      poll_interval: 30s
      sync_mode: bidirectional
    # NOTE: credentials stored in system keychain, NOT here

  - id: "conn_02HXYZ"
    provider: jira
    label: "Work Jira"
    settings:
      server: "https://jira.company.com"
      jql: "assignee = currentUser() AND status != Done"
      poll_interval: 60s
      sync_mode: bidirectional

  - id: "conn_03HXYZ"
    provider: github
    label: "OSS GitHub Issues"
    settings:
      repos: "arcaven/ThreeDoors,arcaven/other-repo"
      labels: "bug,feature"
      poll_interval: 300s
      sync_mode: readonly

# Existing config preserved
theme: classic
show_key_hints: true
llm: {}
```

Credentials are stored in the system keychain under service `threedoors`, key `connection:<id>:token`.

---

## 5. Architecture Recommendations

### 5.1 Connection Manager

```go
// ConnectionState represents the lifecycle state of a data source connection.
type ConnectionState int

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateSyncing
    StateError
    StateAuthExpired
    StatePaused
)

// Connection represents a configured instance of a data source.
type Connection struct {
    ID           string            // unique instance ID (ULID)
    ProviderName string            // "jira", "todoist", "github"
    Label        string            // user-friendly: "Work Jira", "Personal Todoist"
    State        ConnectionState
    LastSync     time.Time
    LastError    string
    SyncMode     string            // "bidirectional", "readonly"
    PollInterval time.Duration
    Settings     map[string]string // provider-specific non-secret config
    TaskCount    int               // cached active task count
}

// ConnectionManager manages the lifecycle of all data source connections.
type ConnectionManager struct {
    connections map[string]*Connection
    registry    *Registry
    keyring     keyring.Keyring
    mu          sync.RWMutex
}

// Key methods:
// Add(providerName, label string, settings map[string]string) (*Connection, error)
// Remove(id string, keepTasks bool) error
// Pause(id string) error
// Resume(id string) error
// TestConnection(id string) (HealthCheckResult, error)
// ForceSync(id string) error
// ReAuthenticate(id string) error  // triggers OAuth/token refresh
// List() []*Connection
// Get(id string) (*Connection, error)
// SyncLog(id string, limit int) []SyncEvent
```

### 5.2 Credential Storage

**Priority chain:**
1. Environment variable (CI/automation) — `THREEDOORS_<PROVIDER>_TOKEN` or `THREEDOORS_CONN_<LABEL>_TOKEN`
2. System keychain (interactive use) — via `99designs/keyring`
3. Encrypted file fallback (headless Linux) — keyring's built-in JWT backend

**Rules:**
- Credentials NEVER stored in `config.yaml`
- Credentials NEVER logged or echoed (mask with `••••`)
- Token refresh happens silently; re-auth requires explicit user action
- On keychain failure, warn user and offer `--insecure-storage` flag (following `gh` pattern)

### 5.3 Sync Conflict Resolution

**Strategy (party mode consensus):**
1. **Last-writer-wins with notification** as default — simple, predictable
2. **Remote-wins for metadata** (status, priority) — source system is authoritative
3. **Local-wins for ThreeDoors-specific fields** (effort category, door assignment)
4. **Log all conflicts** in sync log for transparency
5. **Never auto-delete** — orphaned tasks (deleted remotely) get marked, user decides

### 5.4 Plugin Architecture

**Compiled-in, not pluggable** (party mode consensus). Go's `plugin` package is Linux-only and fragile. The existing Registry + Factory pattern is clean, simple, and testable:

```go
func registerBuiltinAdapters(reg *core.Registry) {
    _ = reg.Register("textfile", ...)
    _ = reg.Register("jira", jira.Factory)
    _ = reg.Register("todoist", todoist.Factory)
    _ = reg.Register("github", github.Factory)
    // Adding a new provider = implement interface + add one line here
}
```

### 5.5 OAuth Implementation

**Recommendation: Device Code Flow** (not callback server)

Device code flow advantages over callback server:
- No port conflict issues
- No firewall/proxy problems
- Works in remote SSH sessions
- Works in containers
- Simpler implementation (HTTP polling vs HTTP server)
- Same pattern used by `gh`, `railway`, `gcloud`

Flow:
1. Request device code from provider
2. Display code and URL to user
3. Open browser automatically (fallback: display URL)
4. Poll authorization endpoint until confirmed or timeout
5. Exchange for access + refresh tokens
6. Store in keychain

### 5.6 Auto-Detection of Existing Tools

**Innovation insight from party mode:** The `:connect` wizard should detect installed tools:
- Check if `gh` CLI is installed and authenticated → offer GitHub Issues connection
- Check for `~/.config/todoist/` or `TODOIST_API_TOKEN` → offer Todoist connection
- Check for Jira config files → offer Jira connection
- Check for `.obsidian/` directories → offer Obsidian vault connection

This transforms setup from "configure everything" to "we found these — connect them?"

---

## 6. Suggested Epic/Story Breakdown

### Epic A — Connection Manager Infrastructure

| Story | Description | Effort |
|-------|-------------|--------|
| A.1 | Connection state machine and ConnectionManager type | M |
| A.2 | Keyring integration (99designs/keyring) with env var fallback | M |
| A.3 | Config schema v3 migration (connections[] replacing providers[]) | M |
| A.4 | Connection CRUD operations (add/remove/pause/resume) | M |
| A.5 | Sync event logging infrastructure | S |
| A.6 | Migrate existing adapters to ConnectionManager pattern | L |

### Epic B — Sources TUI

| Story | Description | Effort |
|-------|-------------|--------|
| B.1 | Setup wizard with huh forms (provider select → config → test → confirm) | L |
| B.2 | Sources dashboard view (list, status indicators, keybindings) | L |
| B.3 | Source detail view (health checks, settings, sync stats) | M |
| B.4 | Sync log view | S |
| B.5 | Status bar integration for connection health alerts | S |
| B.6 | Disconnection flow with task preservation options | M |
| B.7 | Re-authentication flow (OAuth + token re-entry) | M |

### Epic C — Sources CLI

| Story | Description | Effort |
|-------|-------------|--------|
| C.1 | `threedoors connect <provider>` command (non-interactive flags) | M |
| C.2 | `threedoors sources` list/status/test commands | M |
| C.3 | `threedoors sources` management commands (pause/resume/sync/disconnect) | M |
| C.4 | `threedoors sources log` command | S |
| C.5 | JSON output support for all sources commands | S |

### Epic D — OAuth Device Code Flow

| Story | Description | Effort |
|-------|-------------|--------|
| D.1 | Generic device code flow client | M |
| D.2 | GitHub OAuth integration | M |
| D.3 | Linear OAuth integration | M |
| D.4 | Token refresh lifecycle (silent refresh, expiry detection) | M |

### Epic E — Sync Lifecycle & Advanced Features

| Story | Description | Effort |
|-------|-------------|--------|
| E.1 | Conflict resolution strategy with logging | M |
| E.2 | Orphaned task handling (remote deletions) | S |
| E.3 | Auto-detection of existing tools in setup wizard | M |
| E.4 | Proactive connection health notifications | S |

**Estimated total:** ~25 stories across 5 epics. Epics A and B are the critical path — they should be designed together and implemented A-first.

---

## 7. Opportunities Noted (Not In Scope)

- **Webhook receiver:** For sources that support webhooks, a local HTTP server could replace polling. But polling is simpler and works universally. Consider webhooks as an optimization in a future epic.
- **Sync dashboard command:** `threedoors dashboard` could show a full-screen live view of all sync activity across all sources. Similar to `lazydocker` container monitoring.
- **Provider marketplace:** A community-contributed provider directory. Not practical with compiled-in architecture, but could be revisited if ThreeDoors moves to a plugin model.
- **Shared connections:** Team-level connection configs that multiple users share (e.g., "Engineering Jira" with shared JQL). Requires auth-per-user but config-per-team.
- **Mobile companion:** Connection status and basic management from an iOS/Android app. Blocked on Epic 16 (iPhone, icebox).

---

## 8. Decision Record

### Adopted Approaches

| Decision | Rationale |
|----------|-----------|
| `99designs/keyring` for credential storage | More backends than `zalando/go-keyring`, encrypted file fallback for headless, built for aws-vault |
| Device code flow for OAuth | No callback server complexity, works in SSH/containers, proven pattern (gh, gcloud) |
| Compiled-in providers (not pluggable) | Go plugin package is Linux-only and fragile; Registry+Factory is clean and testable |
| `charmbracelet/huh` for setup wizard | Native Bubbletea integration, Group-as-page wizard model, dynamic forms |
| Last-writer-wins conflict resolution | Simple, predictable; remote-wins for metadata, local-wins for ThreeDoors-specific fields |
| Named connections with ULID IDs | Supports multiple instances of same provider, user-friendly labels |

### Rejected Approaches

| Approach | Reason for Rejection |
|----------|---------------------|
| OAuth callback server | Port conflicts, firewall issues, doesn't work in SSH/containers |
| `zalando/go-keyring` | Fewer backends, no encrypted file fallback for headless Linux |
| Go plugin architecture | Linux-only, fragile, poor debugging, no real ecosystem |
| Storing credentials in config.yaml | Security risk; credentials should never be in plain text config files |
| Apple Notes integration | No task model, fragile AppleScript, macOS-only, extremely poor ROI |
| In-app settings editor (like lazygit) | Setup wizard is better UX for initial config; config file editing is fine for power users |
| Webhook receiver for change detection | Adds HTTP server complexity; polling is simpler, works universally; revisit as optimization later |

---

*This report was produced as a research artifact. No implementation code was written. See party mode transcript for full discussion context.*
