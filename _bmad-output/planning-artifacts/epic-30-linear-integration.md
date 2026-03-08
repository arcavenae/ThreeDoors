---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/requirements.md (FR116-FR119)
  - docs/prd/product-scope.md (Phase 5)
  - docs/architecture/components.md (LinearAdapter component)
  - _bmad-output/planning-artifacts/sprint-change-proposal-2026-03-07-linear-integration.md
  - docs/research/task-source-expansion-research.md
---

# ThreeDoors - Epic 30: Linear Integration

## Overview

This document provides the epic and story breakdown for Epic 30, decomposing PRD requirements FR116-FR119 and the Linear Integration architecture into implementable stories.

Linear was identified as Tier 2 priority in the task source expansion research, ranking 3rd after Todoist (Epic 25) and GitHub Issues (Epic 26). Linear has the best task model alignment of all evaluated services — 6 workflow states map cleanly to ThreeDoors statuses, priority (0-4) and estimates map to Effort, and due dates, labels, and Markdown descriptions all have direct mappings.

## Requirements Inventory

### Functional Requirements

- FR116: Linear as task source via GraphQL API with structured field mapping (title→Text, description→Context, labels→Tags, state.type→Status, priority→Effort with inversion, estimate→Effort secondary, dueDate→due date), filtered by team and assignee
- FR117: Linear auth via personal API key in config.yaml or `LINEAR_API_KEY` env var, configurable `team_ids` list for multi-team scoping
- FR118: Status mapping (triage/backlog/unstarted→todo, started→in-progress, completed→complete, cancelled→archived) and priority mapping (0-4 inverted to ThreeDoors Effort, estimate as secondary signal)
- FR119: Bidirectional sync — transition issues to team "Done" state via GraphQL mutation when complete in ThreeDoors, offline queuing via WALProvider

### Non-Functional Requirements

- Existing NFR-CQ1 through NFR-CQ5 (code quality gates) apply to all stories
- API responses must be cached locally with configurable TTL (default 5m) to avoid rate limit issues
- GraphQL queries must use cursor-based pagination for teams with >50 issues
- Rate limit: 5,000 requests/hour — adapter must respect rate limits and implement backoff

### Additional Requirements (from Architecture/Research)

- GraphQL-only API — no Go SDK available, use `github.com/hasura/go-graphql-client` or raw HTTP with typed query/response structs
- Issues require team context — queries scoped by `team_ids` configuration
- Linear priority is inverted: 1=urgent, 2=high, 3=medium, 4=low, 0=no priority
- `estimate` field (story points) serves as secondary effort signal when priority is 0
- Cursor-based pagination via `PageInfo { hasNextPage, endCursor }` pattern

### FR Coverage Map

- FR116: Story 30.1 (GraphQL client) + Story 30.2 (field mapping provider)
- FR117: Story 30.1 (auth config)
- FR118: Story 30.2 (status + priority mapping)
- FR119: Story 30.3 (bidirectional sync + WAL)

## Epic List

### Epic 30: Linear Integration

Integrate Linear as a task source for engineering teams, leveraging Linear's excellent task model alignment (rich workflow states, priority, estimates, labels, due dates) to provide the highest-fidelity task import of any ThreeDoors integration. Uses the Linear GraphQL API with typed Go structs for query construction and response parsing.

**FRs covered:** FR116, FR117, FR118, FR119
**Priority:** P2 (after Tier 1 integrations — Todoist Epic 25, GitHub Issues Epic 26)
**Dependencies:** Epic 7 (Adapter SDK — complete), Epic 13 (Multi-Source Aggregation — complete)
**Estimated Effort:** 4-5 days

---

## Stories

### Story 30.1: Linear GraphQL Client & Auth Configuration

**Priority:** P2
**Story Points:** 3
**Depends On:** Epic 7 (complete)
**FRs:** FR116, FR117

#### Description

As a ThreeDoors user who uses Linear for project management, I want to configure Linear as a task source with my API key and team selection, so that my Linear issues appear as ThreeDoors tasks.

#### Acceptance Criteria

1. **AC1:** A `LinearClient` struct in `internal/adapters/linear/client.go` wraps a GraphQL HTTP client that authenticates via `Authorization: <api_key>` header
2. **AC2:** The client supports querying Linear issues with cursor-based pagination (fetching all pages automatically for teams with >50 issues)
3. **AC3:** Configuration in `config.yaml` accepts:
   ```yaml
   providers:
     - name: linear
       settings:
         api_key: "lin_api_xxx"    # or LINEAR_API_KEY env var
         team_ids:
           - "TEAM-1"
           - "TEAM-2"
         assignee: "@me"           # optional: filter by authenticated user
         poll_interval: "5m"       # default: 5 minutes
   ```
4. **AC4:** Env var `LINEAR_API_KEY` takes precedence over config file `api_key` when both are set
5. **AC5:** `HealthCheck()` verifies API connectivity by querying the `viewer` field (authenticated user info)
6. **AC6:** GraphQL queries use typed Go structs (not raw string queries) for compile-time safety
7. **AC7:** Rate limit handling: respect `Retry-After` headers and implement exponential backoff

#### Technical Notes

- Use `github.com/hasura/go-graphql-client` for typed GraphQL queries, OR build a thin HTTP client with `encoding/json` for query/response marshaling. Decision: prefer the hasura library for type safety and pagination helpers, but evaluate at implementation time.
- Define a `GraphQLClient` interface for testability:
  ```go
  type GraphQLClient interface {
      Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
      Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error
  }
  ```
- Linear GraphQL endpoint: `https://api.linear.app/graphql`
- Auth header format: `Authorization: <api_key>` (no "Bearer" prefix)

#### Tasks

- [ ] Create `internal/adapters/linear/` package directory
- [ ] Define `GraphQLClient` interface and implement with hasura client or raw HTTP
- [ ] Define typed query structs for `issues`, `viewer`, and `workflowStates`
- [ ] Implement cursor-based pagination wrapper
- [ ] Implement `LinearConfig` parsing from config.yaml with env var override
- [ ] Implement `HealthCheck()` via `viewer` query
- [ ] Implement rate limit backoff
- [ ] Write unit tests with mocked GraphQLClient interface
- [ ] Write integration test with `httptest.NewServer` serving canned GraphQL responses

---

### Story 30.2: Read-Only Linear Provider with Field Mapping

**Priority:** P2
**Story Points:** 3
**Depends On:** Story 30.1
**FRs:** FR116, FR118

#### Description

As a ThreeDoors user, I want my Linear issues to appear as properly mapped ThreeDoors tasks with correct status, effort, tags, and due dates, so that I can manage my engineering work through the Three Doors interface.

#### Acceptance Criteria

1. **AC1:** `LinearProvider` implements `core.TaskProvider` with `LoadTasks()` returning all issues from configured teams, mapped to ThreeDoors `Task` structs
2. **AC2:** Field mapping:
   - `issue.title` → `Task.Text`
   - `issue.description` → `Task.Context` (Markdown preserved)
   - `issue.labels[].name` → `Task.Tags`
   - `issue.dueDate` → `Task.DueDate`
   - `issue.identifier` (e.g., "TEAM-123") → `Task.ID` prefix for source identification
3. **AC3:** Status mapping (via `issue.state.type`):
   - `triage` → `todo`
   - `backlog` → `todo`
   - `unstarted` → `todo`
   - `started` → `in-progress`
   - `completed` → `complete`
   - `cancelled` → `archived`
4. **AC4:** Effort mapping:
   - Linear priority (1=urgent → Effort 4, 2=high → Effort 3, 3=medium → Effort 2, 4=low → Effort 1)
   - When priority is 0 (no priority), use `estimate` as secondary signal (if present)
5. **AC5:** `SaveTask()` and `DeleteTask()` return `core.ErrReadOnly` — this is a read-only provider in this story
6. **AC6:** `Watch()` returns a channel that polls for issue changes at `poll_interval` intervals
7. **AC7:** Local cache with configurable TTL (default 5m) at `~/.threedoors/linear-cache.yaml` to reduce API calls
8. **AC8:** Source badge: `[LN]` for TUI display
9. **AC9:** Issues filtered by `assignee` config when specified (default: all team members)

#### Technical Notes

- Query `workflowStates` for each team to build the status mapping table dynamically (state names vary per team, but `state.type` is consistent)
- Cache invalidation: full reload on poll, partial updates if Linear supports `updatedAfter` filter
- `Task.SourceProvider` set to `"linear"` for write routing

#### Tasks

- [ ] Implement `LinearProvider` struct with `LoadTasks()`, `SaveTask()`, `DeleteTask()`, `MarkComplete()`
- [ ] Implement status mapping from `state.type` enum
- [ ] Implement priority inversion mapping (Linear 1-4 → ThreeDoors 4-1)
- [ ] Implement estimate-as-effort fallback when priority is 0
- [ ] Implement assignee filtering
- [ ] Implement `Watch()` polling with configurable interval
- [ ] Implement local YAML cache with TTL
- [ ] Register provider in adapter registry (`internal/adapters/registry.go`)
- [ ] Write unit tests for field mapping (table-driven)
- [ ] Write unit tests for status mapping (all 6 state types)
- [ ] Write unit tests for effort mapping (priority + estimate combinations)

---

### Story 30.3: Bidirectional Sync & WAL Integration

**Priority:** P2
**Story Points:** 3
**Depends On:** Story 30.2
**FRs:** FR119

#### Description

As a ThreeDoors user, when I complete a task that originated from Linear, I want the corresponding Linear issue to be transitioned to "Done" automatically, so that my work status stays synchronized across both systems.

#### Acceptance Criteria

1. **AC1:** `MarkComplete(taskID)` sends a GraphQL mutation to transition the issue to the team's `completed` workflow state
2. **AC2:** The mutation discovers the correct "completed" state ID dynamically by querying `workflowStates` for the issue's team and finding the state with `type: "completed"`
3. **AC3:** Write operations are wrapped in `WALProvider` for offline queuing — if the API is unreachable, the mutation is queued and retried
4. **AC4:** `SaveTask()` supports updating the issue `title` and `description` via GraphQL mutation (but not status — status changes go through `MarkComplete`)
5. **AC5:** Failed mutations are logged with the task ID and error for sync observability
6. **AC6:** `DeleteTask()` remains `ErrReadOnly` — Linear issue deletion is destructive and not supported
7. **AC7:** Sync status is reported via `HealthCheck()` including last successful sync time and pending WAL entries

#### Technical Notes

- GraphQL mutation for state transition:
  ```graphql
  mutation IssueUpdate($id: String!, $stateId: String!) {
    issueUpdate(id: $id, input: { stateId: $stateId }) {
      success
      issue { id state { name type } }
    }
  }
  ```
- Must query `workflowStates(filter: { team: { id: { eq: $teamId } } })` to discover the correct completed state ID
- Cache the team→completedStateId mapping to avoid re-querying on every completion

#### Tasks

- [ ] Implement `MarkComplete()` with GraphQL mutation for state transition
- [ ] Implement dynamic completed-state discovery per team
- [ ] Cache team→completedStateId mapping with TTL
- [ ] Integrate WALProvider wrapper for offline queuing
- [ ] Implement `SaveTask()` for title/description updates
- [ ] Add sync status reporting to `HealthCheck()`
- [ ] Write unit tests for mutation construction
- [ ] Write unit tests for WAL integration (queue, replay, dequeue)
- [ ] Write integration test with mocked GraphQL server for full sync flow

---

### Story 30.4: Contract Tests & Integration Testing

**Priority:** P2
**Story Points:** 2
**Depends On:** Story 30.2
**FRs:** FR116 (validation)

#### Description

As a ThreeDoors developer, I want comprehensive contract and integration tests for the Linear adapter, so that I can verify compliance with the TaskProvider interface and catch regressions in field mapping and sync behavior.

#### Acceptance Criteria

1. **AC1:** `LinearProvider` passes the full `adapters.RunContractTests` suite using a mocked `GraphQLClient` interface
2. **AC2:** Contract tests validate all `TaskProvider` methods: `LoadTasks`, `SaveTask`, `DeleteTask`, `MarkComplete`, `Watch`, `HealthCheck`
3. **AC3:** Field mapping tests cover all Linear field types with table-driven test cases:
   - All 6 `state.type` values → correct ThreeDoors status
   - All 5 priority values (0-4) → correct Effort
   - Estimate fallback when priority is 0
   - Label arrays → Tags
   - Due date parsing → DueDate
4. **AC4:** Integration test with `httptest.NewServer` serving canned GraphQL responses validates end-to-end query/response flow
5. **AC5:** Error handling tests: auth failure (401), rate limit (429 with Retry-After), network timeout, malformed response
6. **AC6:** Pagination test: mock server returns paginated results, verify all pages fetched
7. **AC7:** All tests pass with `go test -race ./internal/adapters/linear/...`

#### Technical Notes

- Use the `GraphQLClient` interface from Story 30.1 for mocking — no external dependencies in tests
- Canned responses should match actual Linear API response shapes
- Contract test suite is in `internal/adapters/contract.go`

#### Tasks

- [ ] Create `internal/adapters/linear/linear_test.go` with contract test runner
- [ ] Create mock `GraphQLClient` implementation with canned responses
- [ ] Write table-driven field mapping tests (status, priority, effort, labels, dates)
- [ ] Write pagination test with multi-page mock responses
- [ ] Write error handling tests (401, 429, timeout, malformed)
- [ ] Write `httptest.NewServer` integration test for full query/response flow
- [ ] Verify `go test -race` passes
- [ ] Run `adapters.RunContractTests` with LinearProvider

---

## Dependency Graph

```
Epic 7 (Adapter SDK) ──── COMPLETE
        │
        ▼
  Story 30.1: GraphQL Client & Auth
        │
        ▼
  Story 30.2: Read-Only Provider ──────┐
        │                              │
        ▼                              ▼
  Story 30.3: Bidirectional Sync   Story 30.4: Contract Tests
```

## Validation Checklist

- [x] All PRD requirements (FR116-FR119) are covered by at least one story
- [x] Stories follow the established adapter pattern (Epics 25, 26)
- [x] 4-story breakdown matches Tier 1 integration epics
- [x] Story points total: 11 (within 4-5 day estimate)
- [x] Dependencies are explicit and all prerequisites are complete
- [x] Contract test story ensures adapter compliance
- [x] GraphQL-specific concerns addressed (typed queries, pagination, mutations)
- [x] Offline-first operation via WALProvider in sync story
