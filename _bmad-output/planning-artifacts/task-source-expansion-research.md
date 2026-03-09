# Task Source Expansion Research

## Overview

ThreeDoors currently supports three task sources: **text files** (YAML), **Obsidian** (Markdown checkboxes), and **Apple Notes** (via AppleScript). This document evaluates additional integrations that would make ThreeDoors more useful across daily workflows.

## Current Architecture Summary

### Adapter Pattern

All task sources implement the `TaskProvider` interface (`internal/core/provider.go`):

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

New adapters are registered in `cmd/threedoors/main.go` via a `Registry` of `AdapterFactory` functions. Each adapter lives in its own package under `internal/adapters/`. A contract test suite (`internal/adapters/contract.go`) validates all implementations.

### Task Model

The core `Task` struct supports: `ID`, `Text`, `Context`, `Status`, `Type`, `Effort`, `Location`, `Notes`, `Blocker`, timestamps, and `SourceProvider`. Statuses include: `todo`, `blocked`, `in-progress`, `in-review`, `complete`, `deferred`, `archived`.

### Multi-Source Aggregation

The `MultiSourceAggregator` (Epic 13) combines tasks from multiple providers, tagging each with `SourceProvider` and routing writes back to the originating provider. Read-only adapters can return `ErrReadOnly` from `MarkComplete` and be wrapped in a `FallbackProvider`. A `WALProvider` wrapper enables offline-first operation with queued retries.

### Configuration

Sources are configured in `~/.threedoors/config.yaml`:

```yaml
schema_version: 1
providers:
  - name: obsidian
    settings:
      vault_path: /path/to/vault
  - name: textfile
```

---

## Evaluated Integrations

### Tier 1 — Recommended (High Value, Clean Fit)

#### 1. Todoist

| Attribute | Details |
|---|---|
| **API** | REST (unified v1, launched 2025) |
| **Auth** | Personal API token or OAuth 2.0 |
| **Rate Limits** | 450 requests / 15 minutes |
| **Go SDK** | Third-party ([go-todoist](https://github.com/TreelightSoftware/go-todoist)) — targets deprecated v2; raw HTTP recommended for v1 |
| **User Base** | 50M+ users; popular with individuals and small teams |

**Task Model Mapping:**

| ThreeDoors Field | Todoist Field | Quality |
|---|---|---|
| Text | `content` | Direct |
| Context | `description` | Direct |
| Status | `is_completed` (bool) | Limited — binary only |
| Effort | `priority` (1–4, inverted scale) | Usable with mapping |
| Due date | `due.date` / `due.datetime` | Good |
| Tags | `labels` (string array) | Direct |

**Why Tier 1:** Todoist is the most popular standalone task manager. Simple API key auth, clean REST API, and good field coverage. The binary status model is limiting but workable — ThreeDoors only needs to surface incomplete tasks. Priority maps to Effort with a scale inversion. Implementation effort is low given the existing adapter pattern.

**Estimated Effort:** 2–3 days. HTTP client, auth config, field mapping, contract tests.

**Considerations:**
- Priority scale is inverted (4 = urgent in Todoist)
- The go-todoist library targets deprecated API v2; building a thin HTTP client against v1 is preferable
- No in-progress status — all non-completed tasks map to `todo`
- Subtasks are available via `parent_id` but not needed for initial integration

---

#### 2. GitHub Issues

| Attribute | Details |
|---|---|
| **API** | REST v3 + GraphQL v4 |
| **Auth** | Personal Access Token (PAT) or OAuth 2.0 |
| **Rate Limits** | 5,000 requests/hour (authenticated) |
| **Go SDK** | Official: [google/go-github](https://github.com/google/go-github) |
| **User Base** | 100M+ developers |

**Task Model Mapping:**

| ThreeDoors Field | GitHub Field | Quality |
|---|---|---|
| Text | `title` | Direct |
| Context | `body` (Markdown) | Direct |
| Status | `state` (open/closed) | Limited — binary |
| Effort | None natively | Missing — label convention possible |
| Due date | `milestone.due_on` | Indirect |
| Tags | `labels` (name + color) | Direct |

**Why Tier 1:** ThreeDoors targets developers, and GitHub Issues is where engineering tasks already live. The official Go SDK is excellent and actively maintained. The `assignee` filter lets users pull only their own issues. Despite lacking native priority and due dates, the developer audience overlap is maximum.

**Estimated Effort:** 2–3 days. go-github handles most complexity. Main work: repo/org config, assignee filtering, label-to-effort mapping convention.

**Considerations:**
- Issues are repo-scoped — config needs `repos` list or org-level query
- No native priority — could map specific label names (e.g., `priority:high`) to Effort
- GitHub Projects v2 adds richer fields but requires separate GraphQL API
- Binary status (open/closed) — could check for "in progress" label
- Consider filtering: assigned to user, open state, specific labels

---

### Tier 2 — Worthwhile (Good Fit With Caveats)

#### 3. Linear

| Attribute | Details |
|---|---|
| **API** | GraphQL only |
| **Auth** | Personal API key or OAuth 2.0 |
| **Rate Limits** | 5,000 requests/hour |
| **Go SDK** | None — use generic GraphQL client |
| **User Base** | Popular with engineering teams, growing rapidly |

**Task Model Mapping:**

| ThreeDoors Field | Linear Field | Quality |
|---|---|---|
| Text | `title` | Direct |
| Context | `description` (Markdown) | Direct |
| Status | `state.type` (6 workflow states) | Excellent |
| Effort | `priority` (0–4) + `estimate` | Good |
| Due date | `dueDate` | Direct |
| Tags | `labels` | Direct |

**Why Tier 2 (not Tier 1):** Linear has the best task model alignment of any service evaluated — rich statuses, priority, estimates, labels, and due dates all map cleanly. However, GraphQL-only with no Go SDK means writing typed queries and handling cursor pagination manually. Issues require a `team` context. The target audience (engineering teams) overlaps well with ThreeDoors users.

**Estimated Effort:** 4–5 days. GraphQL query construction, pagination, team discovery, field mapping.

---

#### 4. Microsoft To Do

| Attribute | Details |
|---|---|
| **API** | REST via Microsoft Graph |
| **Auth** | OAuth 2.0 with Azure AD |
| **Rate Limits** | Undisclosed (429 with Retry-After) |
| **Go SDK** | Official: [msgraph-sdk-go](https://github.com/microsoftgraph/msgraph-sdk-go) |
| **User Base** | Massive (bundled with Microsoft 365) |

**Task Model Mapping:**

| ThreeDoors Field | MS To Do Field | Quality |
|---|---|---|
| Text | `title` | Direct |
| Context | `body.content` | Direct (HTML/text) |
| Status | `status` (5 states) | Excellent — best of all evaluated |
| Effort | `importance` (low/normal/high) | Usable |
| Due date | `dueDateTime` | Direct (timezone-aware) |
| Tags | `categories` | Direct |

**Why Tier 2:** MS To Do has the richest native status model (notStarted, inProgress, completed, waitingOnOthers, deferred) — a near-perfect match for ThreeDoors statuses. Official Go SDK exists. However, Azure AD OAuth setup is the most complex auth flow of all evaluated services, which increases both implementation effort and user friction.

**Estimated Effort:** 5–6 days. Azure AD OAuth flow, Graph API setup, token refresh handling.

---

#### 5. ClickUp

| Attribute | Details |
|---|---|
| **API** | REST v2 |
| **Auth** | Personal API token or OAuth 2.0 |
| **Rate Limits** | 100 req/min (free–business); up to 10,000 (enterprise) |
| **Go SDK** | None |
| **User Base** | 10M+ users across teams and individuals |

**Task Model Mapping:**

| ThreeDoors Field | ClickUp Field | Quality |
|---|---|---|
| Text | `name` | Direct |
| Context | `markdown_description` | Direct (Markdown) |
| Status | `status.status` | Good (list-defined) |
| Effort | `priority` (4 levels) + `time_estimate` | Good |
| Due date | `due_date` (Unix ms) | Direct |
| Tags | `tags` | Direct |

**Why Tier 2:** Best overall field coverage — native priority, Markdown descriptions, tags, and time estimates. However, no Go SDK means raw HTTP, and statuses are list-defined (not a fixed enum), requiring list-level discovery.

**Estimated Effort:** 4–5 days. HTTP client, workspace/space/list navigation, status discovery.

---

### Tier 3 — Lower Priority (Workable With Gaps)

#### 6. Trello

| Attribute | Details |
|---|---|
| **API** | REST |
| **Auth** | API key + OAuth 1.0a token |
| **Rate Limits** | 300 req/10sec per key; 100 req/10sec per token |
| **Go SDK** | [adlio/trello](https://github.com/adlio/trello) |

**Mapping Quality:** Title and labels map well. Status is modeled as the card's list (indirect). No native priority or due date concept beyond a simple `due` field. Board-centric model requires board selection.

**Estimated Effort:** 3–4 days.

**Why Lower Priority:** Declining popularity relative to newer tools. No priority field. List-based status requires convention mapping. OAuth 1.0a is dated.

---

#### 7. Asana

| Attribute | Details |
|---|---|
| **API** | REST |
| **Auth** | PAT or OAuth 2.0 |
| **Rate Limits** | ~1,500 req/min (premium) |
| **Go SDK** | Third-party, low activity |

**Mapping Quality:** Title, description, due date, and tags map well. No native priority (requires paid custom fields). Status is binary (completed bool) with section-based bucketing.

**Estimated Effort:** 4–5 days.

**Why Lower Priority:** Priority requires paid plan. Binary status. Inactive Go SDK means raw HTTP likely needed. `opt_fields` selection pattern adds API usage complexity.

---

#### 8. Google Tasks

| Attribute | Details |
|---|---|
| **API** | REST v1 |
| **Auth** | OAuth 2.0 only |
| **Rate Limits** | 50,000 queries/day |
| **Go SDK** | Official: `google.golang.org/api/tasks/v1` |

**Mapping Quality:** Title and notes only. No priority, no tags, no labels. Binary status. Due date is date-only (no time component). Extremely minimal model.

**Estimated Effort:** 3–4 days (mostly OAuth setup).

**Why Lower Priority:** Despite the official Go SDK and Gmail integration, the task model is too minimal — no priority, no tags, no rich statuses. Useful only as a basic sync target. OAuth-only auth adds user friction.

---

### Tier 4 — Not Recommended for Initial Work

#### 9. Jira

**Why Not:** ADF (Atlassian Document Format) for descriptions instead of plain text/Markdown. Project-scoped with variable custom fields per instance. Enterprise-grade complexity. Burst rate limit enforcement changes (March 2026). Estimated 6–8 days effort.

#### 10. Notion Databases

**Why Not:** No guaranteed schema — every user's database has different property names and types. Requires dynamic schema discovery per connection. 3 req/sec rate limit is tight. Breaking API changes in 2025-09-03 version. Estimated 7–10 days effort.

---

## Recommended Implementation Order

Based on user base overlap, implementation effort, and field mapping quality:

| Priority | Integration | Rationale |
|---|---|---|
| **1st** | **Todoist** | Largest personal task manager user base, simplest API, clean mapping |
| **2nd** | **GitHub Issues** | Maximum developer audience overlap, official Go SDK |
| **3rd** | **Linear** | Best field alignment, strong in engineering teams |
| **4th** | **Microsoft To Do** | Best status model, massive user base, but complex auth |
| **5th** | **ClickUp** | Best overall field coverage, growing user base |

### Implementation Notes

**For all API-based adapters:**

1. **Auth configuration** — Store API tokens in `~/.threedoors/config.yaml` settings. For OAuth flows, implement a local callback server (like `gh auth login` does) and store refresh tokens securely.

2. **Read-only first** — Initial implementations should be read-only (return `ErrReadOnly` from write methods, wrap in `FallbackProvider`). Write-back can be added later.

3. **Caching** — API-based adapters should cache loaded tasks with a TTL to avoid hitting rate limits on every TUI refresh. The `Watch()` method could use webhooks or polling with configurable intervals.

4. **Offline support** — Wrap in `WALProvider` to queue status changes when offline.

5. **Status mapping** — Define a standard mapping table per adapter:
   - Binary (completed/not): `completed` → `complete`, everything else → `todo`
   - Rich (like Linear/MS To Do): Map to ThreeDoors statuses directly

6. **Config pattern** — Follow existing adapter config:
   ```yaml
   providers:
     - name: todoist
       settings:
         api_token: "your-token-here"
         filter: "today | overdue"  # optional Todoist filter
     - name: github
       settings:
         token: "ghp_xxx"
         repos: "owner/repo1,owner/repo2"
         assignee: "@me"
   ```

7. **Contract tests** — All adapters must pass `adapters.RunContractTests`. For API-based adapters, use interface-based mocking (similar to Apple Notes' `CommandExecutor` pattern).

---

## Appendix: API Comparison Matrix

| Service | Auth Complexity | Task Model Richness | Go SDK Quality | Rate Limits | Overall Fit |
|---|---|---|---|---|---|
| Todoist | Low (API key) | Good | None (raw HTTP) | Moderate | Excellent |
| GitHub Issues | Low (PAT) | Limited | Official, excellent | Generous | Excellent |
| Linear | Low (API key) | Excellent | None (GraphQL) | Generous | Good |
| MS To Do | High (Azure AD) | Excellent | Official | Unknown | Good |
| ClickUp | Low (API key) | Excellent | None | Tight (free) | Good |
| Trello | Medium (OAuth 1) | Limited | Community | Generous | Fair |
| Asana | Low (PAT) | Moderate | Community (inactive) | Moderate | Fair |
| Google Tasks | High (OAuth 2) | Minimal | Official | Generous | Poor |
| Jira | Medium (API token) | Excessive | Community | Complex | Poor |
| Notion | Medium (integration token) | Variable | Community | Tight (3/sec) | Poor |
