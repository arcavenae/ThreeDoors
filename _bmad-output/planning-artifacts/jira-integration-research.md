# Jira Integration Research for ThreeDoors

**Date:** 2026-03-03
**Status:** Research Complete
**Author:** fancy-hawk (worker agent)

## Executive Summary

ThreeDoors can integrate with Jira as a task source via the existing `TaskProvider` adapter pattern. A `JiraProvider` would query Jira issues via JQL, map them to ThreeDoors `Task` structs, and optionally transition issues when tasks are completed. The existing `WALProvider` wrapper already supports offline-first queuing, making it a natural fit for network-dependent backends.

---

## 1. Current Architecture

### TaskProvider Interface

Defined in `internal/core/provider.go`:

```go
type TaskProvider interface {
    Name() string
    LoadTasks() ([]*Task, error)
    SaveTask(task *Task) error
    SaveTasks(tasks []*Task) error
    DeleteTask(taskID string) error
    MarkComplete(taskID string) error
    Watch() <-chan ChangeEvent
    HealthCheck() HealthCheckResult
}
```

### Task Model

Defined in `internal/core/task.go`. Key fields relevant to Jira mapping:

| ThreeDoors Field | Type | Jira Equivalent |
|---|---|---|
| `ID` | `string` | Issue key (e.g., `PROJ-42`) |
| `Text` | `string` (1-500 chars) | `summary` |
| `Context` | `string` | `project.key` + labels |
| `Status` | `TaskStatus` | `status.statusCategory` |
| `Type` | `TaskType` | `issuetype.name` (mapped) |
| `Effort` | `TaskEffort` | `priority.name` or story points (mapped) |
| `Location` | `TaskLocation` | Not applicable (leave empty) |
| `Notes` | `[]TaskNote` | Comments (read-only) |
| `Blocker` | `string` | Blocked link or flagged status |
| `CreatedAt` | `time.Time` | `created` |
| `UpdatedAt` | `time.Time` | `updated` |
| `CompletedAt` | `*time.Time` | Resolution date |
| `SourceProvider` | `string` | `"jira"` |

### Existing Infrastructure That Helps

- **`WALProvider`** (`internal/core/wal_provider.go`): Wraps any `TaskProvider` with write-ahead logging. Failed writes are queued to `~/.threedoors/sync-queue.jsonl` and retried with exponential backoff. This solves offline-first behavior out of the box.
- **`FallbackProvider`** (`internal/core/fallback_provider.go`): Wraps primary + fallback. If Jira is unreachable on startup, falls back to local text file.
- **`MultiSourceAggregator`** (`internal/core/aggregator.go`): Merges tasks from multiple providers. A user could run Jira + local text file simultaneously.
- **`Registry`** (`internal/core/registry.go`): Factory-based adapter registration. New adapters register via `reg.Register("jira", factory)`.
- **Contract tests** (`internal/adapters/contract.go`): `RunContractTests(t, factory)` validates any new adapter against the full interface contract.

---

## 2. Jira REST API

### Search Endpoint (Current)

The legacy `GET /rest/api/3/search` has been **removed from Jira Cloud**. The current endpoint is:

```
POST /rest/api/3/search/jql
Content-Type: application/json

{
  "jql": "assignee = currentUser() AND sprint in openSprints() AND statusCategory != Done",
  "fields": ["summary", "status", "priority", "assignee", "labels", "issuetype", "created", "updated"],
  "maxResults": 50
}
```

**Pagination**: Cursor-based via `nextPageToken` (no `startAt`). The `total` field is gone; use `isLast: true` to detect the final page.

### Transitioning Issues

Status changes require the transitions API (direct field setting is not supported):

```
GET  /rest/api/3/issue/{key}/transitions     # Discover valid transitions
POST /rest/api/3/issue/{key}/transitions     # Execute: {"transition": {"id": "31"}}
```

Returns **204 No Content** on success. A **409 Conflict** means a concurrent transition is in progress — retry with backoff.

### Other Relevant Endpoints

| Endpoint | Purpose |
|---|---|
| `GET /rest/api/3/issue/{key}` | Get single issue details |
| `GET /rest/api/3/field` | Discover custom field IDs (story points, sprint) |
| `POST /rest/api/3/search/approximate-count` | Get result count without full payload |

---

## 3. Authentication

### Jira Cloud

**API Token + Basic Auth** (recommended for personal/CLI tools):

```
Authorization: Basic base64(email:api_token)
```

Tokens generated at `https://id.atlassian.com/manage/api-tokens`. Works through SSO/2FA.

**OAuth 2.0 (3LO)** (recommended for multi-user apps):

1. Register app in Atlassian Developer Console
2. Redirect user to `https://auth.atlassian.com/authorize`
3. Exchange code for access + refresh tokens
4. `Authorization: Bearer <access_token>`

Relevant scopes: `read:jira-work`, `write:jira-work`, `read:jira-user`

### Jira Server / Data Center

**Personal Access Tokens** (PAT, available since Jira 8.14):

```
Authorization: Bearer <PAT>
```

### Recommendation for ThreeDoors

Start with **API Token + Basic Auth** for Cloud and **PAT + Bearer** for Server/DC. These are the simplest approaches for a personal CLI tool. OAuth 2.0 can be added later if needed for distribution.

**Credential storage**: Environment variables (`JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`) or a config entry in `~/.threedoors/config.yaml`. Never store credentials in task YAML files.

| Method | Jira Cloud | Jira Server/DC |
|---|---|---|
| API Token + Basic | Yes (primary) | No |
| PAT + Bearer | No | Yes (8.14+) |
| OAuth 2.0 (3LO) | Yes (multi-user) | Via plugin |

---

## 4. JQL Queries for ThreeDoors

ThreeDoors shows 3 tasks at a time from the "available" pool. Useful default JQL queries:

```sql
-- My current sprint work
assignee = currentUser() AND sprint in openSprints() AND statusCategory != Done

-- My open issues across all projects
assignee = currentUser() AND statusCategory != Done ORDER BY priority ASC

-- High priority items
assignee = currentUser() AND priority in (Highest, High) AND statusCategory != Done

-- Recently updated (for refresh)
assignee = currentUser() AND updated >= -1h AND statusCategory != Done
```

The JQL query should be **user-configurable** in `config.yaml` so users can filter to their workflow. A sensible default is the current sprint query.

### Proposed Config

```yaml
providers:
  - name: jira
    settings:
      url: "https://company.atlassian.net"
      auth_type: "basic"           # "basic" | "pat" | "oauth"
      jql: "assignee = currentUser() AND sprint in openSprints() AND statusCategory != Done"
      max_results: "50"            # per page
      poll_interval: "30s"         # background refresh interval
      field_mapping_priority: "effort"  # map Jira priority to ThreeDoors effort
```

---

## 5. Field Mapping Design

### Status Mapping

Jira uses `statusCategory` (3 categories) which maps cleanly:

| Jira `statusCategory.key` | ThreeDoors `TaskStatus` |
|---|---|
| `new` | `todo` |
| `indeterminate` | `in-progress` |
| `done` | `complete` |

For finer-grained mapping, inspect `status.name`:

| Jira `status.name` | ThreeDoors `TaskStatus` |
|---|---|
| "To Do", "Open", "Backlog" | `todo` |
| "In Progress", "In Development" | `in-progress` |
| "In Review", "In QA" | `in-review` |
| "Blocked", "Impediment" | `blocked` |
| "Done", "Closed", "Resolved" | `complete` |

The mapping should be configurable, with `statusCategory` as the fallback.

### Priority/Effort Mapping

| Jira Priority | ThreeDoors Effort |
|---|---|
| Highest, High | `deep-work` |
| Medium | `medium` |
| Low, Lowest | `quick-win` |

Alternative: Map story points instead (if available):
- 1-2 points: `quick-win`
- 3-5 points: `medium`
- 8+ points: `deep-work`

### Type Mapping

| Jira Issue Type | ThreeDoors Type |
|---|---|
| Story, Task | `technical` (default) |
| Bug | `technical` |
| Sub-task | `technical` |
| Epic | `deep-work` effort (not a type mapping) |

Type mapping is lossy — Jira issue types don't map cleanly to ThreeDoors' creative/admin/technical/physical categories. Best approach: use labels or a custom field, or leave `Type` empty and let users categorize locally.

### Context Field

Combine project key and labels:

```go
context := fmt.Sprintf("[%s] %s", issue.Fields.Project.Key, strings.Join(issue.Fields.Labels, ", "))
// e.g., "[PROJ] backend, auth"
```

### Task ID Strategy

Use the Jira issue key as the ThreeDoors task ID: `PROJ-42`. This is human-readable, unique, and enables direct lookups. The existing `Task.Validate()` allows any non-empty string as ID, so Jira keys are compatible.

---

## 6. Bidirectional Sync

### ThreeDoors -> Jira (MarkComplete)

When a user completes a task in ThreeDoors:

1. `MarkComplete(taskID)` is called with the Jira issue key
2. `GET /rest/api/3/issue/{key}/transitions` to discover valid transitions
3. Find a transition where `to.statusCategory.key == "done"`
4. `POST /rest/api/3/issue/{key}/transitions` with that transition ID
5. If Jira is unreachable, the `WALProvider` wrapper queues the operation

### Jira -> ThreeDoors (LoadTasks)

On each `LoadTasks()` call:

1. Execute the configured JQL query via `POST /rest/api/3/search/jql`
2. Paginate through all results
3. Map each issue to a `*core.Task`
4. Return the full set (ThreeDoors' `TaskPool` handles filtering)

### Conflict Resolution

- **Jira is the source of truth** for status. If a task was transitioned in Jira between ThreeDoors loads, the new status is accepted.
- **Local-only fields** (ThreeDoors `Type`, `Effort`, `Location`, `Notes`) are not synced back to Jira. They could be cached locally in a sidecar file (`~/.threedoors/jira-local-metadata.yaml`).
- **WAL replay**: When replaying queued transitions, check current status first. If the issue is already in "Done", skip the transition.

---

## 7. Rate Limiting

### Jira Cloud Rate Limits (as of March 2026)

Three independent systems:

| System | Limit | Response |
|---|---|---|
| Points-based hourly quota | 65K-500K points/hr (varies by plan) | 429 |
| Burst API limits | 100 req/sec GET, 50 req/sec PUT/DELETE | 429 |
| Per-issue write limits | 20 writes/2s, 100 writes/30s | 429 |

### Recommended Strategy

```go
// Respect Retry-After header
if resp.StatusCode == http.StatusTooManyRequests {
    retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
    time.Sleep(retryAfter)
    // retry with exponential backoff + jitter
}
```

For a personal TUI tool fetching ~50 issues on load + occasional transitions, rate limits are unlikely to be hit. But the implementation must handle 429 gracefully.

---

## 8. Go Libraries

### Option A: `andygrunwald/go-jira` v1 (stable, widely used)

```go
import jira "github.com/andygrunwald/go-jira"

tp := jira.BasicAuthTransport{Username: email, Password: token}
client, _ := jira.NewClient(tp.Client(), baseURL)
issues, _, _ := client.Issue.Search(jql, &jira.SearchOptions{MaxResults: 50})
```

- Pros: Battle-tested, supports Cloud + Server/DC, large community
- Cons: Uses deprecated search endpoint; v2 is unstable

### Option B: `ctreminiom/go-atlassian` v2 (Cloud-focused)

```go
import v3 "github.com/ctreminiom/go-atlassian/v2/jira/v3"

client, _ := v3.New(nil, baseURL)
client.Auth.SetBasicAuth(email, token)
```

- Pros: Native v3 API support, Agile/Sprint APIs built-in
- Cons: Cloud-only, smaller community

### Option C: Raw `net/http` (no dependency)

Build a thin client using `net/http` + `encoding/json`. This aligns with ThreeDoors' Go proverb: "A little copying is better than a little dependency."

- Pros: Zero dependencies, full control over new search/jql endpoint, smaller binary
- Cons: More code to write and maintain

### Recommendation

**Option C (raw `net/http`)** for initial implementation. ThreeDoors only needs ~4 API calls (search, get issue, get transitions, do transition). A thin client is simpler than importing a full SDK. If complexity grows, migrate to `go-atlassian` v2.

---

## 9. Offline-First Behavior

The existing `WALProvider` already handles this:

```go
// In provider setup:
jiraProvider := NewJiraProvider(config)
walProvider := core.NewWALProvider(jiraProvider)
```

When Jira is unreachable:
- `LoadTasks()` returns cached results from last successful fetch (cached in a local sidecar file)
- Write operations (`MarkComplete`, `SaveTask`) are queued to the WAL
- On next successful connection, `ReplayPending()` drains the queue

### Local Cache

Add a local cache file (`~/.threedoors/jira-cache.yaml`) for offline reads:
- Updated on every successful `LoadTasks()`
- Read as fallback when Jira is unreachable
- Include a TTL/timestamp so stale data is flagged in the UI

---

## 10. Proposed Implementation Plan

### Phase 1: Read-Only Jira Provider

1. Create `internal/adapters/jira/` package
2. Implement thin HTTP client (`jira_client.go`) with:
   - Basic Auth + PAT support
   - `SearchJQL(ctx, jql, fields, maxResults, pageToken)` method
   - Rate limit handling (429 + Retry-After)
   - Cursor-based pagination
3. Implement `JiraProvider` (`jira_provider.go`):
   - `LoadTasks()` — JQL search, map to `[]*core.Task`
   - `SaveTask()` / `SaveTasks()` / `DeleteTask()` — return `ErrReadOnly`
   - `MarkComplete()` — return `ErrReadOnly`
   - `Watch()` — return `nil`
   - `HealthCheck()` — test API connectivity
4. Register in `RegisterBuiltinAdapters()`
5. Add contract tests + unit tests with HTTP test server
6. Add config support for Jira settings

**Estimated scope**: ~500-700 lines of Go + ~400 lines of tests

### Phase 2: Bidirectional Sync

1. Implement `MarkComplete()` via transitions API
2. Add local metadata cache for ThreeDoors-only fields
3. Wrap with `WALProvider` for offline queuing
4. Add `FallbackProvider` wrapping for graceful degradation

**Estimated scope**: ~300 lines of Go + ~200 lines of tests

### Phase 3: Enhanced Features

1. Background polling via `tea.Tick` command
2. Configurable status/priority mapping in config
3. Custom field discovery (`GET /rest/api/3/field`)
4. OAuth 2.0 support for multi-user scenarios
5. `Watch()` implementation via polling (not webhooks — impractical for TUI)

---

## 11. Open Questions

1. **Should Jira tasks be editable in ThreeDoors?** Editing `Text` would require `PUT /rest/api/3/issue/{key}` to update the summary. This adds complexity — recommend read-only for Phase 1.
2. **How to handle Jira's ADF description format?** The `description` field returns Atlassian Document Format JSON, not plain text. Options: ignore it, extract text nodes only, or render a simplified version.
3. **Story points vs priority for effort mapping?** Story points require discovering a custom field ID per instance. Priority is simpler but less accurate.
4. **Should the local metadata cache persist ThreeDoors-specific categorizations?** This would let users tag Jira tasks with Type/Location/Effort locally.
5. **Multi-project support?** Should the JQL query be the only filter, or should we support explicit project key configuration?

---

## References

- [Jira Cloud REST API v3 — Issue Search](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/)
- [Jira Cloud REST API — Authentication](https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/)
- [Jira Cloud Rate Limiting](https://developer.atlassian.com/cloud/jira/platform/rate-limiting/)
- [Jira Webhooks](https://developer.atlassian.com/cloud/jira/platform/webhooks/)
- [go-jira library](https://github.com/andygrunwald/go-jira)
- [go-atlassian library](https://github.com/ctreminiom/go-atlassian)
- [New search/jql endpoint migration](https://community.atlassian.com/forums/Jira-articles/How-to-use-the-new-Jira-cloud-issue-search-API/ba-p/3006109)
