# Sprint Change Proposal: Snooze/Defer as First-Class Action

**Date:** 2026-03-07
**Proposed by:** UX Research Analysis
**Change Scope:** Minor — New epic addition, no existing work affected
**Status:** Approved for implementation planning

---

## Section 1: Issue Summary

### Problem Statement

The `deferred` status exists in the ThreeDoors task model (`StatusDeferred` in `internal/core/task_status.go`) with valid transitions defined (todo->deferred, deferred->todo), but it has no UX surface in the TUI. Users cannot snooze tasks with a return date.

When a task appears in the doors but isn't actionable today, users can only give "not now" feedback or re-roll. The task keeps reappearing in future door selections, creating "I can't do this now but don't want to lose it" anxiety — a key contributor to task manager abandonment.

### Discovery Context

Identified through competitive UX research (`ux-workflow-improvements-research.md`), ranked as **P0 (High Impact, Low Effort)**. Akiflow's "snooze to tomorrow" is one of its most-used features. The existing `StatusDeferred` infrastructure confirms this was architecturally anticipated but never surfaced.

### Evidence

1. **Competitive:** Akiflow's snooze-to-tomorrow is its most-used feature
2. **Research:** "Backlog horror" is the #1 reason task managers get abandoned (73% churn within 30 days) — showing only actionable tasks directly addresses this
3. **Codebase:** `StatusDeferred` already exists with transitions, styles, tests — the backend is partially ready
4. **Design Principle #5:** "Ready means ready — the 3 doors should only show tasks the user can act on right now"

---

## Section 2: Impact Analysis

### Epic Impact

- **No existing epics modified.** This is a new Epic 28.
- **Epic 23 (CLI):** Optional future extension — `threedoors task defer` CLI command. Non-blocking.
- **Epic 24 (MCP):** Optional future extension — `defer_task` MCP tool. Non-blocking.
- **Epics 25/26 (Todoist/GitHub):** Deferred tasks already excluded from sync-back (adapters only sync `StatusComplete`). No conflict.

### Story Impact

No current or future stories require changes. The new epic is fully independent.

### Artifact Conflicts

| Artifact | Impact | Severity |
|----------|--------|----------|
| PRD (requirements.md) | Add new FRs (FR104-FR108) for snooze behavior | Low — additive |
| PRD (product-scope.md) | Add Snooze/Defer to Phase 6+ scope | Low — additive |
| Architecture (data-models.md) | Add `DeferUntil *time.Time` to Task model | Low — one field |
| Architecture (components.md) | Add SnoozeView, DeferredListView components | Low — new components |
| Architecture (core-workflows.md) | Add snooze workflow | Low — new workflow |
| Task Status (task_status.go) | Add `in-progress -> deferred` transition | Low — one line |
| Door Selection (task_pool.go) | Filter by DeferUntil in GetAvailableForDoors | Low — one condition |

### Technical Impact

- **Key binding conflict:** `S` key is currently used for re-roll (S/Down). Options: (a) Use `Z` for snooze, (b) Reassign `S` to snooze since Down arrow covers re-roll. **Recommendation:** Use `Z` key to avoid breaking existing muscle memory.
- **Data model:** Add `DeferUntil *time.Time` field to Task struct, persisted in YAML as `defer_until`.
- **Door filtering:** Modify `GetAvailableForDoors()` to exclude tasks where `DeferUntil` is non-nil and in the future.
- **Auto-return:** Tasks whose `DeferUntil` has passed should automatically transition back to `todo` status on next load.

---

## Section 3: Recommended Approach

### Selected Path: Direct Adjustment (New Epic)

Add **Epic 28: Snooze/Defer as First-Class Action** to the roadmap as a P1 epic with 4-5 stories.

### Rationale

1. **Zero disruption:** Fully independent of Epics 23-26. Can be parallelized.
2. **Existing infrastructure:** `StatusDeferred` with transitions, styles, and tests already exists. This is ~60% backend-ready.
3. **Low effort:** Estimated 4-5 stories. Core changes are: one model field, one filter condition, two TUI views.
4. **High impact:** Directly addresses Design Principle #5 ("Ready means ready") and the #1 task manager abandonment reason.

### Effort Estimate

- **Overall:** Low-Medium
- **Risk:** Low — builds on proven patterns (status transitions, TUI views, command palette)
- **Timeline Impact:** None on existing work

### Trade-offs Considered

- **Alternative: Enhance existing "not now" feedback** — Rejected. Feedback doesn't remove the task from rotation; snooze does.
- **Alternative: Add to Daily Planning Mode (Epic 28 existing)** — The daily planning "defer" option (FR98) is complementary but different — it means "leave in pool without priority." Snooze-with-date is a distinct, more powerful mechanism.

---

## Section 4: Detailed Change Proposals

### 4.1 PRD Changes

**File: docs/prd/requirements.md**

ADD after Phase 6+ Daily Planning Mode section:

```
## Phase 6+ - Snooze/Defer as First-Class Action (Accepted)

**Snooze/Defer:**

**FR104:** The system shall provide a snooze action accessible via the `Z` key when a door
is selected, presenting quick defer options: Tomorrow, Next Week, Pick Date, and Someday

**FR105:** When a task is snoozed, the system shall set a `defer_until` timestamp on the
task, transition its status to `deferred`, and immediately remove it from door selection
eligibility

**FR106:** Deferred tasks shall automatically return to `todo` status when their
`defer_until` date arrives (checked on application startup and periodically during sessions)

**FR107:** The system shall provide a `:deferred` command in the command palette that
displays all currently snoozed tasks with their return dates, allowing users to un-snooze
(return to todo immediately) or change the snooze date

**FR108:** The system shall log snooze events (task ID, defer-until date, snooze option
selected) as a `snooze` event type in the JSONL session metrics log
```

**File: docs/prd/product-scope.md**

ADD to Phase 6+ section:

```
- Snooze/Defer: Z-key quick snooze with Tomorrow/Next Week/Pick Date/Someday options,
  deferred task filtering from doors, `:deferred` list view, auto-return on due date
```

### 4.2 Architecture Changes

**File: docs/architecture/data-models.md**

ADD `DeferUntil` field to Task model:

```
OLD:
- `completedAt`: `*time.Time` - When marked complete (nil if not complete)

NEW:
- `completedAt`: `*time.Time` - When marked complete (nil if not complete)
- `deferUntil`: `*time.Time` - When deferred task should return to pool (nil if not deferred)
```

ADD to YAML storage format example:

```yaml
defer_until: 2026-03-08T09:00:00Z  # null if not deferred
```

**File: docs/architecture/components.md**

ADD new TUI components:

```
### Component: SnoozeView (Epic 28)

**Responsibility:** Quick date picker for snoozing tasks from the doors view.

**Key Interfaces:**
- `NewSnoozeView(task *Task) *SnoozeView`
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
- `View() string`

**Key Behaviors:**
- Four options: Tomorrow (9am), Next Week (Monday 9am), Pick Date (calendar picker), Someday (no date, indefinite defer)
- Arrow keys navigate, Enter confirms, ESC cancels
- On confirm: sets DeferUntil, transitions status to deferred, returns to doors

### Component: DeferredListView (Epic 28)

**Responsibility:** Display all currently snoozed tasks with return dates.

**Key Interfaces:**
- `NewDeferredListView(pool *TaskPool) *DeferredListView`
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
- `View() string`

**Key Behaviors:**
- Lists deferred tasks sorted by return date (soonest first)
- 'u' un-snoozes selected task (returns to todo immediately)
- 'e' edits snooze date
- ESC returns to previous view
- Accessible via `:deferred` command
```

### 4.3 Task Status Transition Change

**File: internal/core/task_status.go**

```
OLD:
StatusInProgress: {StatusBlocked, StatusInReview, StatusComplete},

NEW:
StatusInProgress: {StatusBlocked, StatusInReview, StatusComplete, StatusDeferred},
```

Rationale: Users mid-task should be able to snooze. "I started this but can't finish today."

### 4.4 Door Selection Filter Change

**File: internal/core/task_pool.go — GetAvailableForDoors()**

```
OLD:
if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {

NEW:
if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {
    // Skip deferred tasks that haven't reached their return date
    if t.DeferUntil != nil && t.DeferUntil.After(time.Now().UTC()) {
        continue
    }
```

Note: Tasks with `DeferUntil` in the past AND status still `deferred` should be auto-returned to `todo` on startup.

---

## Section 5: Implementation Handoff

### Change Scope: Minor

Direct implementation by development team. No backlog reorganization needed.

### Handoff Plan

| Role | Responsibility |
|------|---------------|
| **PM Agent** | Create Epic 28 with stories in epics-and-stories breakdown |
| **Architect Agent** | Create detailed architecture doc for snooze/defer |
| **Dev Workers** | Implement stories via `/implement-story` |
| **QA** | Validate acceptance criteria, test defer date filtering |

### Success Criteria

1. `Z` key on selected door opens snooze picker
2. Snoozed tasks disappear from door rotation until return date
3. `:deferred` command shows all snoozed tasks
4. Auto-return works on application startup
5. Session metrics log snooze events
6. All existing tests continue to pass

### Suggested Story Breakdown (4 stories)

1. **28.1 — DeferUntil Field & Auto-Return Logic** — Add field to Task model, YAML persistence, auto-return on startup, update GetAvailableForDoors filter
2. **28.2 — Snooze TUI View & Z-Key Binding** — SnoozeView component with Tomorrow/Next Week/Pick Date/Someday, wire to doors view
3. **28.3 — Deferred List View & Un-snooze** — DeferredListView component, `:deferred` command, un-snooze and edit-date actions
4. **28.4 — Session Metrics & CLI Integration** — Log snooze events to JSONL, optional `threedoors task defer` CLI command

---

## Approval

**Approved for implementation planning.** Proceed to PRD update, architecture creation, and epic/story breakdown.
