# External API Integrations

ThreeDoors integrates with six external task management services via the `TaskProvider` interface (`internal/core/provider.go`). All adapters live under `internal/adapters/` and follow the same patterns: thin HTTP/SDK client, `TaskProvider` implementation, config parsing with env-var fallback, and rate-limit handling.

For sync protocol details (scheduler, circuit breaker, caching), see [task-sync-architecture.md](task-sync-architecture.md).

---

## Adapter Summary

| Adapter | API | Auth | Env Var | Package |
|---------|-----|------|---------|---------|
| Jira | REST API v3 | Basic (Cloud) / PAT (Server) | `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN` | `internal/adapters/jira/` |
| Apple Reminders | JXA via `osascript` | macOS TCC consent | (none) | `internal/adapters/reminders/` |
| Todoist | REST API v1 | Bearer token | `TODOIST_API_TOKEN` | `internal/adapters/todoist/` |
| GitHub Issues | go-github SDK (REST v3) | PAT via OAuth2 | `GITHUB_TOKEN` | `internal/adapters/github/` |
| Linear | GraphQL API | API key | `LINEAR_API_KEY` | `internal/adapters/linear/` |
| ClickUp | REST API v2 | PAT | `CLICKUP_API_TOKEN` | `internal/adapters/clickup/` |

All network-based adapters return `RateLimitError` on HTTP 429, parsed from the `Retry-After` header.

---

## 1. Jira

**API:** Atlassian Jira REST API v3 (Cloud and Server/DC).

**Endpoints:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/rest/api/3/search/jql` | Search issues via JQL |
| GET | `/rest/api/3/issue/{key}` | Single issue details |
| GET | `/rest/api/3/issue/{key}/transitions` | Valid status transitions |
| POST | `/rest/api/3/issue/{key}/transitions` | Execute status transition |

**Auth:** Cloud uses `Basic base64(email:api_token)`. Server/DC uses `Bearer <PAT>`. Auth type configured via `auth_type` setting (`basic` or `pat`).

**Config:** `url`, `auth_type` (required); `jql` (default: `assignee = currentUser() AND statusCategory != Done`), `max_results` (default: 50), `poll_interval` (default: 30s).

---

## 2. Apple Reminders

**API:** JavaScript for Automation (JXA) scripts executed via `osascript -l JavaScript`. No network API — communicates with macOS Reminders.app locally.

**Access:** Uses a `CommandExecutor` interface wrapping `osascript`. First invocation triggers a macOS TCC consent dialog for Reminders access. `HealthCheck()` verifies permission.

**Config:** `lists` (comma-separated list names, empty = all), `include_completed` (default: `false`).

---

## 3. Todoist

**API:** Todoist REST API v1.

**Endpoints:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/rest/v1/tasks` | List active tasks |
| POST | `/rest/v1/tasks/{id}/close` | Mark task complete |
| GET | `/rest/v1/projects` | List projects (health check) |

**Auth:** `Authorization: Bearer <api_token>`. Token from `TODOIST_API_TOKEN` env var or config.

**Rate limit:** 450 requests per 15 minutes.

**Config:** `api_token` (required); `project_ids` and `filter` (mutually exclusive, optional), `poll_interval` (default: 30s).

---

## 4. GitHub Issues

**API:** GitHub REST API v3 via `google/go-github` SDK (v68).

**Operations:** List open issues (paginated, filtered by assignee), update issue state, manage labels.

**Auth:** Personal Access Token via OAuth2 token source. `GITHUB_TOKEN` env var or config. Unauthenticated access supported (rate-limited).

**Rate limit:** GitHub primary rate limit (5,000 req/hr authenticated, 60 req/hr unauthenticated). 403 responses parsed as `RateLimitError`.

**Config:** `repos` (required, comma-separated `owner/repo`), `assignee` (default: `@me`), `poll_interval` (default: 5m), `in_progress_label` (default: `in-progress`), `priority_label.*` mappings.

---

## 5. Linear

**API:** Linear GraphQL API (`https://api.linear.app/graphql`).

**Operations:** Query issues with cursor-based pagination (default page size: 50), filtered by team and assignee.

**Auth:** `Authorization: <api_key>`. Key from `LINEAR_API_KEY` env var or config.

**Rate limit:** HTTP 429 with `Retry-After` header. Max 3 automatic retries.

**Config:** `api_key`, `team_ids` (required, comma-separated); `assignee` (optional), `poll_interval` (default: 5m).

---

## 6. ClickUp

**API:** ClickUp REST API v2 (`https://api.clickup.com/api/v2`).

**Endpoints:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/team/{team_id}/space` | List spaces |
| GET | `/space/{space_id}/list` | List lists in a space |
| GET | `/list/{list_id}/task` | List tasks in a list |
| PUT | `/task/{task_id}` | Update task status |

**Auth:** `Authorization: <api_token>`. Token from `CLICKUP_API_TOKEN` env var or config.

**Config:** `api_token`, `team_id` (required); at least one of `space_ids` or `list_ids`; `assignee`, `poll_interval` (default: 30s), `done_status`, `blocked_status` (for write-back).

---

## Local-Only Adapters (No External API)

These adapters read from local files or macOS services and do not make network calls:

- **textfile** (`internal/adapters/textfile/`) — YAML task files on disk, with `fsnotify` file watching
- **obsidian** (`internal/adapters/obsidian/`) — Markdown files in an Obsidian vault
- **applenotes** (`internal/adapters/applenotes/`) — Apple Notes via JXA/osascript (local IPC, no network)

---

## Common Patterns

All external adapters share these implementation patterns:

- **`TaskProvider` interface** — every adapter implements the full CRUD interface from `internal/core/provider.go`
- **`RateLimitError`** — structured error type with `RetryAfter` duration, checked via `errors.As`
- **Config parsing** — `ParseConfig(settings map[string]string)` with env-var override precedence
- **Read-only phase** — initial implementations return `core.ErrReadOnly` for write methods; write support added incrementally
- **Contract tests** — `internal/adapters/contract.go` provides `RunContractTests` for adapter conformance
- **Poll interval** — configurable per adapter; used by the sync scheduler for background refresh

## Related Documents

- [task-sync-architecture.md](task-sync-architecture.md) — sync scheduler, circuit breaker, canonical ID mapping
- [coding-standards.md](coding-standards.md) — atomic writes, error handling patterns
