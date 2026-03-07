# Task Sync Architecture

## Overview

This document extends the ThreeDoors architecture to support Jira and Apple Reminders as task sources, and introduces sync protocol improvements (scheduler, circuit breaker, canonical ID mapping) for reliable multi-provider operation.

All new adapters implement the existing `TaskProvider` interface (`internal/core/provider.go`) and integrate via the `Registry` (`internal/core/registry.go`).

---

## 1. Jira Adapter

### Package Structure

```
internal/adapters/jira/
├── jira_client.go           # Thin HTTP client for Jira REST API v3
├── jira_client_test.go      # HTTP test server-based unit tests
├── jira_provider.go         # TaskProvider implementation
├── jira_provider_test.go    # Unit tests with mock client
├── field_mapping.go         # Jira issue → core.Task mapping
├── field_mapping_test.go    # Field mapping unit tests
└── config.go                # Jira-specific config types
```

### HTTP Client

A thin client using `net/http` + `encoding/json`. No third-party SDK dependency.

**Endpoints used:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/rest/api/3/search/jql` | Search issues via JQL (cursor-paginated) |
| GET | `/rest/api/3/issue/{key}` | Get single issue details |
| GET | `/rest/api/3/issue/{key}/transitions` | Discover valid status transitions |
| POST | `/rest/api/3/issue/{key}/transitions` | Execute status transition |

**Authentication:**

```go
type AuthConfig struct {
    Type     string // "basic" | "pat"
    URL      string // Jira base URL
    Email    string // Cloud only
    APIToken string // Cloud: API token, Server: PAT
}
```

- Cloud: `Authorization: Basic base64(email:api_token)`
- Server/DC: `Authorization: Bearer <PAT>`
- Credentials sourced from env vars (`JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`) or config.yaml settings.

**Rate Limit Handling:**

```go
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
    resp, err := c.httpClient.Do(req.WithContext(ctx))
    if err != nil {
        return nil, fmt.Errorf("jira request %s: %w", req.URL.Path, err)
    }
    if resp.StatusCode == http.StatusTooManyRequests {
        retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
        // Caller handles retry with exponential backoff + jitter
        return nil, &RateLimitError{RetryAfter: retryAfter}
    }
    return resp, nil
}
```

### JiraProvider

```go
type JiraProvider struct {
    client  *Client
    jql     string
    mapping *FieldMapper
    cache   *TaskCache
}
```

**Phase 1 (Read-Only):**

| Method | Behavior |
|--------|----------|
| `Name()` | Returns `"jira"` |
| `LoadTasks()` | Execute JQL, paginate, map to `[]*core.Task`, update local cache |
| `SaveTask()` | Returns `core.ErrReadOnly` |
| `SaveTasks()` | Returns `core.ErrReadOnly` |
| `DeleteTask()` | Returns `core.ErrReadOnly` |
| `MarkComplete()` | Returns `core.ErrReadOnly` |
| `Watch()` | Returns `nil` |
| `HealthCheck()` | Test API connectivity with approximate-count endpoint |

**Phase 2 (Bidirectional):**

| Method | Behavior |
|--------|----------|
| `MarkComplete(taskID)` | GET transitions → find "Done" transition → POST transition |
| WAL wrapping | `core.NewWALProvider(jiraProvider)` for offline queuing |

### Field Mapping

```go
type FieldMapper struct {
    StatusMap   map[string]core.TaskStatus // Jira status name → ThreeDoors status
    EffortMap   map[string]core.TaskEffort // Jira priority name → ThreeDoors effort
    UseCategory bool                       // Fall back to statusCategory if status name not mapped
}
```

**Default status mapping (via statusCategory):**

| Jira `statusCategory.key` | ThreeDoors `TaskStatus` |
|---------------------------|------------------------|
| `new` | `todo` |
| `indeterminate` | `in-progress` |
| `done` | `complete` |

**Default effort mapping (via priority):**

| Jira Priority | ThreeDoors Effort |
|---------------|------------------|
| Highest, High | `deep-work` |
| Medium | `medium` |
| Low, Lowest | `quick-win` |

**Task ID:** Jira issue key (e.g., `PROJ-42`) used directly as `Task.ID`.

**Context field:** `fmt.Sprintf("[%s] %s", project.Key, strings.Join(labels, ", "))`.

### Configuration

```yaml
providers:
  - name: jira
    settings:
      url: "https://company.atlassian.net"
      auth_type: "basic"
      jql: "assignee = currentUser() AND sprint in openSprints() AND statusCategory != Done"
      max_results: "50"
      poll_interval: "30s"
```

---

## 2. Apple Reminders Adapter

### Package Structure

```
internal/adapters/reminders/
├── reminders_provider.go       # TaskProvider implementation
├── reminders_provider_test.go  # Unit tests with mock executor
├── jxa_scripts.go              # JXA script templates
├── field_mapping.go            # Reminder ↔ Task field conversion
├── field_mapping_test.go       # Mapping tests
└── config.go                   # Config types
```

### JXA Access via CommandExecutor

Reuses the `CommandExecutor` interface pattern from the Apple Notes adapter:

```go
type CommandExecutor interface {
    Execute(ctx context.Context, script string) (string, error)
}

type OSAScriptExecutor struct{}

func (e *OSAScriptExecutor) Execute(ctx context.Context, script string) (string, error) {
    cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript", "-e", script)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("osascript: %w", err)
    }
    return strings.TrimSpace(string(output)), nil
}
```

JXA scripts return JSON via `JSON.stringify()`, avoiding the brittle text parsing issues from the Apple Notes adapter.

### RemindersProvider

```go
type RemindersProvider struct {
    executor CommandExecutor
    lists    []string // empty = all lists
    timeout  time.Duration
}
```

**Phase 1 (Read-Only):**

| Method | Behavior |
|--------|----------|
| `Name()` | Returns `"reminders"` |
| `LoadTasks()` | JXA read incomplete reminders, parse JSON, map to `[]*core.Task` |
| `SaveTask()` | Returns `core.ErrReadOnly` |
| `MarkComplete()` | Returns `core.ErrReadOnly` |
| `Watch()` | Returns `nil` |
| `HealthCheck()` | Lightweight JXA read to verify Reminders access |

**Phase 2 (Write Support):**

| Method | Behavior |
|--------|----------|
| `SaveTask(task)` | JXA create or update reminder by ID |
| `MarkComplete(taskID)` | JXA set `completed = true` |
| `DeleteTask(taskID)` | JXA delete reminder |

### Field Mapping

| Reminder Field | Task Field | Mapping |
|---------------|-----------|---------|
| `id` | `ID` | Direct (stable persistent `x-apple-reminder://` URI) |
| `name` | `Text` | Direct |
| `body` | `Notes[0].Text` | Map to first TaskNote |
| `completed` | `Status` | `true` → `complete`, `false` → `todo` |
| `priority` | `Effort` | 1-4 → `deep-work`, 5 → `medium`, 6-9 → `quick-win`, 0 → empty |
| `creationDate` | `CreatedAt` | ISO 8601 parse |
| `modificationDate` | `UpdatedAt` | ISO 8601 parse |
| `completionDate` | `CompletedAt` | ISO 8601 parse (nil if not completed) |
| list name | `SourceProvider` | `"reminders:<list-name>"` |

### Configuration

```yaml
providers:
  - name: reminders
    settings:
      lists: "Work,ThreeDoors"  # comma-separated, empty = all
      include_completed: "false"
```

### TCC Permissions

First `osascript` call targeting Reminders triggers a macOS consent dialog. The `HealthCheck()` method attempts a lightweight read and returns guidance if permission is denied.

---

## 3. Todoist Adapter

### Package Structure

```
internal/adapters/todoist/
├── todoist_client.go           # Thin HTTP client for Todoist REST API v1
├── todoist_client_test.go      # HTTP test server-based unit tests
├── todoist_provider.go         # TaskProvider implementation
├── todoist_provider_test.go    # Unit tests with mock client
├── field_mapping.go            # Todoist task → core.Task mapping
├── field_mapping_test.go       # Field mapping unit tests
└── config.go                   # Todoist-specific config types
```

### HTTP Client

A thin client using `net/http` + `encoding/json` targeting Todoist REST API v1. No third-party SDK (the go-todoist library targets the deprecated v2 API).

**Endpoints used:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/rest/v1/tasks` | List active (non-completed) tasks |
| POST | `/rest/v1/tasks/{id}/close` | Mark task as complete |
| GET | `/rest/v1/projects` | List projects (for config validation) |

**Authentication:**

```go
type AuthConfig struct {
    APIToken string // Personal API token
}
```

- All requests: `Authorization: Bearer <api_token>`
- Token sourced from env var (`TODOIST_API_TOKEN`) or config.yaml settings

**Rate Limit Handling:**

Todoist enforces 450 requests per 15 minutes. The adapter reuses the same `RateLimitError` pattern from the Jira adapter:

```go
if resp.StatusCode == http.StatusTooManyRequests {
    retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
    return nil, &RateLimitError{RetryAfter: retryAfter}
}
```

### TodoistProvider

```go
type TodoistProvider struct {
    client     *Client
    projectIDs []string     // filter by project IDs (optional)
    filter     string       // Todoist filter expression (optional, mutually exclusive with projectIDs)
    mapping    *FieldMapper
    cache      *TaskCache
}
```

**Phase 1 (Read-Only):**

| Method | Behavior |
|--------|----------|
| `Name()` | Returns `"todoist"` |
| `LoadTasks()` | GET tasks, filter deleted, map to `[]*core.Task`, update local cache |
| `SaveTask()` | Returns `core.ErrReadOnly` |
| `SaveTasks()` | Returns `core.ErrReadOnly` |
| `DeleteTask()` | Returns `core.ErrReadOnly` |
| `MarkComplete()` | Returns `core.ErrReadOnly` |
| `Watch()` | Returns `nil` (Todoist has no webhook/push support for personal tokens) |
| `HealthCheck()` | GET projects to verify API connectivity |

**Phase 2 (Bidirectional):**

| Method | Behavior |
|--------|----------|
| `MarkComplete(taskID)` | POST `/rest/v1/tasks/{id}/close` |
| WAL wrapping | `core.NewWALProvider(todoistProvider)` for offline queuing |

### Field Mapping

```go
type FieldMapper struct {
    // No configurable mappings needed — Todoist model is fixed
}
```

**Default field mapping:**

| Todoist Field | Task Field | Mapping |
|--------------|-----------|---------|
| `id` | `ID` | Direct (string) |
| `content` | `Text` | Direct |
| `description` | `Context` | Direct |
| `is_completed` | `Status` | `true` → `complete`, `false` → `todo` |
| `priority` | `Effort` | Inverted scale (see table below) |
| `labels` | (future) | Direct string array |
| `due.date` / `due.datetime` | (future) | ISO date parse |
| `project_id` | `SourceProvider` | `"todoist:<project_name>"` |
| `is_deleted` | (filtered) | `true` → excluded from results |

**Priority-to-Effort mapping (inverted scale):**

| Todoist Priority | Todoist Meaning | ThreeDoors Effort |
|-----------------|----------------|------------------|
| 0 | No priority | `quick-win` (lowest) |
| 1 | Normal | `quick-win` |
| 2 | High | `medium` |
| 3 | Urgent | `deep-work` |
| 4 | Critical | `deep-work` (highest) |

### Configuration

```yaml
providers:
  - name: todoist
    settings:
      api_token: "${TODOIST_API_TOKEN}"  # or literal token
      project_ids: "project-id-1,project-id-2"  # optional, comma-separated
      filter: "today | overdue"  # optional, mutually exclusive with project_ids
      poll_interval: "60s"
```

Note: `project_ids` and `filter` are mutually exclusive. If both are specified, the adapter returns a configuration error at initialization.

### Testing Strategy

- **Unit tests:** `httptest.NewServer` returning canned Todoist REST API v1 responses
- **Field mapping tests:** Table-driven tests covering all 5 priority values (0-4) mapping to correct Effort values
- **Contract tests:** `adapters.RunContractTests` with mock HTTP server backing
- **Rate limit tests:** Verify 429 handling with Retry-After header (reuse Jira pattern)
- **Deleted task tests:** Verify `is_deleted == true` tasks are filtered from results
- **Config validation tests:** Verify mutual exclusivity of `project_ids` and `filter`
- **Integration tests:** `//go:build integration` — require real Todoist account (manual only)

---

## 4. Sync Protocol Architecture

### 3.1 Sync Scheduler

Replace on-demand sync with a background scheduler managing independent per-provider sync loops.

```
┌─────────────────────────────────────────────────────────┐
│                    SyncScheduler                        │
│                                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │
│  │  TextFile   │ │   Apple     │ │    Jira     │      │
│  │   Loop      │ │  Reminders  │ │    Loop     │      │
│  │             │ │   Loop      │ │             │      │
│  │  fsnotify   │ │  poll 60s   │ │  poll 30s   │      │
│  │  primary    │ │  adaptive   │ │  adaptive   │      │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘      │
│         │               │               │              │
│         └───────────────┼───────────────┘              │
│                         ▼                              │
│               results chan SyncResult                   │
└─────────────────────────────────────────────────────────┘
```

```go
// internal/core/sync_scheduler.go

type SyncScheduler struct {
    loops   map[string]*ProviderLoop
    results chan SyncResult
    stop    chan struct{}
}

type ProviderLoop struct {
    provider TaskProvider
    circuit  *CircuitBreaker
    interval *AdaptiveInterval
}

type AdaptiveInterval struct {
    current    time.Duration
    min        time.Duration // e.g., 30s
    max        time.Duration // e.g., 30m
    multiplier float64       // e.g., 2.0
    jitter     float64       // e.g., 0.2 (±20%)
}
```

**Behavior:**
- Each loop runs as an independent goroutine
- Providers with `Watch()` support use the channel as primary trigger; polling runs concurrently as fallback
- On sync success: reset interval to minimum
- On sync failure: multiply interval, up to maximum
- Jitter (±20%) prevents thundering herd on recovery

### 3.2 Circuit Breaker

```go
// internal/core/circuit_breaker.go

type CircuitState int

const (
    CircuitClosed   CircuitState = iota // healthy, forward all requests
    CircuitOpen                          // tripped, fast-fail
    CircuitHalfOpen                      // probing, allow one request
)

type CircuitBreaker struct {
    state         CircuitState
    failureCount  int
    failureThreshold int           // default 5
    failureWindow    time.Duration // default 2m
    probeInterval    time.Duration // starts at 30s, doubles, max 30m
    lastFailure      time.Time
    lastProbe        time.Time
    mu               sync.Mutex
}
```

**State transitions:**
- Closed → Open: after `failureThreshold` consecutive failures within `failureWindow`
- Open → Half-Open: after `probeInterval` elapses
- Half-Open → Closed: on successful probe
- Half-Open → Open: on failed probe (double `probeInterval`)

**Integration with existing components:**
- `SyncStatusTracker` reads circuit state for per-provider UI display
- `WALProvider` queues writes when circuit is Open
- `MultiSourceAggregator.LoadTasks()` returns cached tasks for Open providers

### 3.3 Canonical ID Mapping

```go
// Addition to internal/core/task.go

type SourceRef struct {
    Provider string `yaml:"provider" json:"provider"`
    NativeID string `yaml:"native_id" json:"native_id"`
}
```

The `Task` struct gains a `SourceRefs []SourceRef` field. This enables:
- Write routing to all providers that know a task
- Permanent dedup links across providers
- Temp ID mapping for optimistic creation

**Migration:** Existing `SourceProvider string` field maps to `SourceRefs[0]`. Backward-compatible — if `SourceRefs` is empty, fall back to `SourceProvider`.

### 3.4 Local Cache Strategy

API-based adapters cache loaded tasks in `~/.threedoors/<provider>-cache.yaml`:
- Updated on every successful `LoadTasks()`
- Read as fallback when provider is unreachable (circuit Open)
- Include `last_updated` timestamp for staleness detection
- TTL configurable per provider (default: 5 minutes for file-based, 15 minutes for network)

---

## 4. Registration and Initialization

New adapters register via the existing `Registry`:

```go
// cmd/threedoors/main.go (or registration helper)

func RegisterBuiltinAdapters(reg *core.Registry) {
    // Existing adapters
    reg.Register("textfile", textfile.Factory)
    reg.Register("applenotes", applenotes.Factory)
    reg.Register("obsidian", obsidian.Factory)

    // New adapters
    reg.Register("jira", jira.Factory)
    reg.Register("reminders", reminders.Factory)
    reg.Register("todoist", todoist.Factory)
}
```

Config-driven provider selection in `~/.threedoors/config.yaml`:

```yaml
schema_version: 2
providers:
  - name: textfile
  - name: jira
    settings:
      url: "https://company.atlassian.net"
      auth_type: "basic"
      jql: "assignee = currentUser() AND statusCategory != Done"
  - name: reminders
    settings:
      lists: "Work"
  - name: todoist
    settings:
      api_token: "${TODOIST_API_TOKEN}"
      filter: "today | overdue"
```

---

## 5. Testing Strategy

### Jira Adapter

- **Unit tests:** `httptest.NewServer` returning canned Jira API responses
- **Field mapping tests:** Table-driven tests for all status/priority/effort mappings
- **Contract tests:** `adapters.RunContractTests` with mock HTTP server backing
- **Rate limit tests:** Verify 429 handling with Retry-After header
- **Integration tests:** `//go:build integration` — require real Jira instance (manual only)

### Apple Reminders Adapter

- **Unit tests:** Mock `CommandExecutor` returning canned JSON
- **Field mapping tests:** Table-driven for all priority/status/effort mappings
- **Contract tests:** Full `RunContractTests` suite (stable IDs enable full round-trip)
- **Integration tests:** `//go:build integration && darwin` — require macOS with Reminders access

### Sync Protocol

- **Sync scheduler tests:** Fake clock, mock providers, deterministic scheduling
- **Circuit breaker tests:** Table-driven state transition tests
- **Canonical ID tests:** SourceRef linking, dedup resolution, migration from SourceProvider

---

## 6. Dependencies

| Component | Depends On | Blocked By |
|-----------|-----------|------------|
| Jira HTTP client | `net/http`, `encoding/json` | Nothing (zero external deps) |
| JiraProvider | Jira HTTP client, `core.TaskProvider` | HTTP client |
| RemindersProvider | `CommandExecutor`, `core.TaskProvider` | Nothing |
| Todoist HTTP client | `net/http`, `encoding/json` | Nothing (zero external deps) |
| TodoistProvider | Todoist HTTP client, `core.TaskProvider` | HTTP client |
| SyncScheduler | `TaskProvider.Watch()`, `SyncEngine` | Epic 11 (sync observability) |
| CircuitBreaker | `SyncStatusTracker` | Nothing (self-contained) |
| SourceRef (canonical ID) | `core.Task` model change | Nothing |

---

## 7. Open Questions

1. **Jira ADF descriptions:** Ignore for Phase 1, or extract text nodes? Recommend ignore — ThreeDoors uses `Text` (summary) not description.
2. **Apple Reminders priority vs. due date:** Map priority to Effort. Due date could go in Context field for now, or be deferred to a future Task model extension.
3. **SourceRef migration:** Schema version bump from 1 to 2. Migration function runs on first load of old-format tasks. Need to handle both formats during transition.
4. **Sync scheduler as tea.Cmd:** The scheduler should emit results via `tea.Cmd` for TUI integration. This keeps the Bubbletea update loop responsive.
