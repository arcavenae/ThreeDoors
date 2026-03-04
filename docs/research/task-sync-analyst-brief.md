# Task Sync Integration — Analyst Brief

**Date:** 2026-03-03
**Sources:** jira-integration-research.md, task-source-expansion-research.md, apple-reminders-integration-research.md, sync-architecture-scaling-research.md, ux-workflow-improvements-research.md

---

## 1. Key Findings

### Jira Integration

The existing `TaskProvider` interface and infrastructure (`WALProvider`, `FallbackProvider`, `MultiSourceAggregator`, `Registry`) make Jira integration straightforward at the adapter level. The research recommends:

- **Raw `net/http` client** over third-party SDKs — ThreeDoors only needs ~4 Jira API calls (search via JQL, get issue, get transitions, do transition). This aligns with the Go proverb "a little copying is better than a little dependency."
- **Phase 1 read-only**, Phase 2 bidirectional. Read-only returns `ErrReadOnly` from write methods; `FallbackProvider` handles graceful degradation.
- **Status mapping via `statusCategory`** (3 categories: new→todo, indeterminate→in-progress, done→complete) with optional fine-grained `status.name` overrides.
- **Auth: API Token + Basic Auth** (Cloud) and PAT + Bearer (Server/DC). Environment variables or config.yaml — never in task files.
- **Jira issue key as Task.ID** (e.g., `PROJ-42`) — human-readable, unique, compatible with existing validation.
- **WALProvider wrapping** for offline-first: queues failed transitions, replays on reconnection.
- **Estimated scope**: ~500-700 lines Go + ~400 lines tests (Phase 1).

### Apple Reminders Integration

Apple Reminders is a significantly better fit than Apple Notes for task management due to its structured data model (title, notes, due date, priority, completion status, persistent IDs).

- **JXA via `osascript`** recommended over AppleScript (native JSON output via `JSON.stringify`) or EventKit/cgo (build complexity).
- **Reuse `CommandExecutor` pattern** from Apple Notes adapter for testability and CI portability.
- **Stable persistent IDs** (`x-apple-reminder://...`) — eliminates the position-based ID fragility of the Apple Notes adapter.
- **Full CRUD support** possible (unlike Apple Notes which is read-only for `MarkComplete`): SaveTask creates/updates, MarkComplete sets completed flag, DeleteTask removes.
- **Priority mapping**: Reminders 1-4=high, 5=medium, 6-9=low, 0=none → ThreeDoors effort levels.
- **Performance**: ~300-500ms per operation (within NFR6 budget).
- **Phase 1 read-only**, Phase 2 write support, Phase 3 optional EventKit migration behind build tag.

### Generic Task Sync Protocol

The sync architecture research identifies 6 scaling challenges and proposes solutions:

1. **Sync Scheduler** — Per-provider independent sync loops with adaptive intervals (backoff on failure, reset on success). Hybrid push (Watch channel) + polling fallback.
2. **Circuit Breaker** — Per-provider health state tracking (Closed→Open→Half-Open). Prevents cascading failures; integrates with `SyncStatusTracker` for UI display.
3. **Canonical ID Mapping** — `SourceRef` type linking internal UUID to provider-native IDs. Solves cross-provider dedup permanently.
4. **Property-Level Conflict Resolution** — Field-level last-write-wins (vs current task-level). Preserves both title-change-in-source-A and status-change-in-source-B.
5. **Dashboard Enhancements** — Staleness indicators, circuit state display, WAL pending count.
6. **Webhook/Push Support** — Already supported via `Watch() <-chan ChangeEvent`; the hybrid push+poll pattern requires no sync engine changes.

### Task Source Expansion Priority

The research ranks future integrations by user base overlap, effort, and field mapping quality:

| Priority | Integration | Effort | Rationale |
|----------|-----------|--------|-----------|
| 1st | Todoist | 2-3 days | Largest personal task manager, simplest API |
| 2nd | GitHub Issues | 2-3 days | Maximum developer overlap, official Go SDK |
| 3rd | Linear | 4-5 days | Best field alignment, GraphQL-only |
| 4th | Microsoft To Do | 5-6 days | Best status model, complex Azure AD auth |
| 5th | ClickUp | 4-5 days | Best field coverage, growing user base |

Note: The task-source research rates Jira as Tier 4 (high complexity, ADF descriptions) while the dedicated Jira research shows it's feasible with the thin-client approach (~4 API calls). The dedicated research takes precedence.

### UX Workflow Improvements

Relevant to sync integration:
- **Quick Capture CLI** (P0) — non-interactive `threedoors add` from any terminal
- **Task Dependencies** (P1) — `depends_on` field for blocked-task filtering
- **Snooze/Defer** (P0) — defer date for removing tasks from rotation temporarily

These UX improvements complement sync integration by making the unified task pool more actionable.

---

## 2. PRD Recommendations

### New Functional Requirements

| ID | Requirement | Source |
|----|------------|--------|
| FR63 | Jira read-only adapter: query issues via JQL, map to Task model | Jira research |
| FR64 | Jira status mapping: configurable statusCategory/status.name → TaskStatus | Jira research |
| FR65 | Jira auth config: API Token + Basic Auth (Cloud), PAT + Bearer (Server/DC) | Jira research |
| FR66 | Jira bidirectional sync: MarkComplete via transitions API, WAL queuing | Jira research |
| FR67 | Apple Reminders read adapter: JXA-based, structured field mapping | Reminders research |
| FR68 | Apple Reminders write support: SaveTask, MarkComplete, DeleteTask | Reminders research |
| FR69 | Apple Reminders list filtering: configurable list names in config.yaml | Reminders research |
| FR70 | Sync scheduler: per-provider independent sync loops with adaptive intervals | Sync research |
| FR71 | Circuit breaker: per-provider health tracking with automatic recovery | Sync research |
| FR72 | Canonical ID mapping: SourceRef linking internal IDs to provider-native IDs | Sync research |

### New Non-Functional Requirements

| ID | Requirement | Source |
|----|------------|--------|
| NFR20 | API-based adapters must handle HTTP 429 with Retry-After header respect | Jira research |
| NFR21 | API-based adapters must cache loaded tasks with configurable TTL | Sync research |
| NFR22 | Credential storage must use environment variables or config.yaml — never in task files | Jira research |
| NFR23 | Circuit breaker must trip after 5 consecutive failures; probe at configurable interval | Sync research |

---

## 3. Architecture Recommendations

### New Epics

**Epic 19: Jira Integration** (Phase 4)
- Story 19.1: Jira HTTP client (auth, search, pagination, rate limits)
- Story 19.2: JiraProvider read-only (LoadTasks, field mapping, contract tests)
- Story 19.3: Jira bidirectional sync (MarkComplete via transitions, WAL wrapping)
- Story 19.4: Jira config and registration

**Epic 20: Apple Reminders Integration** (Phase 4)
- Story 20.1: Reminders JXA scripts and CommandExecutor
- Story 20.2: RemindersProvider read-only (LoadTasks, field mapping)
- Story 20.3: Reminders write support (SaveTask, MarkComplete, DeleteTask)
- Story 20.4: Reminders config, registration, and health check

**Epic 21: Sync Protocol Hardening** (Phase 4)
- Story 21.1: Sync scheduler with per-provider loops
- Story 21.2: Circuit breaker per provider
- Story 21.3: Canonical ID mapping (SourceRef)
- Story 21.4: Property-level conflict resolution

### Implementation Order

1. **Epic 21 (Sync Protocol)** first — the scheduler and circuit breaker are prerequisites for reliable multi-provider operation
2. **Epic 20 (Apple Reminders)** second — lower complexity, local-only, validates adapter patterns
3. **Epic 19 (Jira)** third — network-dependent, benefits from sync scheduler and circuit breaker

---

## 4. Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Jira ADF description format | Low | Ignore descriptions initially; extract text nodes later |
| JXA deprecation by Apple | Medium | EventKit migration path behind build tag |
| Property-level conflict resolution breaks sync state format | Medium | Schema versioning with migration function |
| Rate limit enforcement changes (Jira March 2026) | Low | Retry-After header handling covers all cases |
| Clock skew across providers | Medium | Use local receipt time as authoritative timestamp |
