# Sprint Change Proposal: Undo Task Completion

**Date:** 2026-03-08
**Triggered By:** Validation Gate Results (docs/validation-gate-results.md) — "No undo for task completion"
**Change Type:** New Epic Proposal
**Scope Classification:** Minor

---

## Section 1: Issue Summary

### Problem Statement

Once a task is marked as complete in ThreeDoors, there is no way to reverse it through the application. The status transition model (`internal/core/task_status.go`) defines `StatusComplete` as a terminal state with zero valid outgoing transitions (`StatusComplete: {}`). Users who accidentally complete a task must manually edit the YAML file to recover it — a process that breaks the friction-free philosophy of the product.

### Discovery Context

This limitation was formally documented during the Phase 1 Validation Gate review (2026-03-03) in the "What Needs Improvement" section: "No undo for task completion. Once a task is marked done, recovery requires manual file editing. The status transition model validates forward progress but doesn't support correction."

The validation gate classified this as an "implementation issue, not a concept flaw" and recommended it be addressed via "status transition history."

### Evidence

- `internal/core/task_status.go:24` — `StatusComplete: {}` (empty transition list)
- `docs/validation-gate-results.md:141-142` — Explicit pain point identification
- `docs/architecture/data-models.md:46` — State machine shows `complete → [*]` as terminal

---

## Section 2: Impact Analysis

### Epic Impact

- **No existing epics affected** — This is a new capability
- **New Epic 32 required** — Undo Task Completion
- **Future integration epics (25, 26, 30)** — Todoist, GitHub Issues, and Linear adapters will eventually need "uncomplete" API operations, but this can be addressed incrementally within those epics

### Story Impact

- No current stories require modification
- 3-4 new stories needed in Epic 32

### Artifact Conflicts

| Artifact | Conflict | Action Needed |
|----------|----------|---------------|
| PRD (requirements.md) | No undo requirement exists | Add new FR for undo capability |
| Architecture (data-models.md) | State machine shows complete as terminal | Update state diagram to include `complete → todo` |
| Architecture (components.md) | StatusManager docs don't mention undo | Update component docs |
| Code (task_status.go) | `StatusComplete: {}` | Add `StatusTodo` to valid transitions |
| Code (task.go) | `UpdateStatus()` doesn't clear `CompletedAt` on reverse | Add logic to clear `CompletedAt` when leaving complete |

### Technical Impact

- **Core change:** One line in `validTransitions` map + `CompletedAt` clearing logic
- **TUI:** StatusUpdateMenu will automatically surface `todo` as valid target when viewing completed tasks
- **CLI (Epic 23):** `threedoors task status <id> todo` will work once transition is valid
- **Adapters:** Each adapter's `MarkComplete` has no reverse; adapter-level uncomplete is a separate concern
- **Session metrics:** Should log undo events for behavioral analysis

---

## Section 3: Recommended Approach

### Selected Path: Direct Adjustment (New Epic)

Create a focused Epic 32: Undo Task Completion with 3 stories:

1. **Status model change** — Add `complete → todo` transition, clear `CompletedAt`, add session event logging
2. **TUI undo experience** — Confirmation dialog, visual feedback, undo from detail view and doors view
3. **Adapter undo support** — Ensure adapters handle re-opened tasks correctly in bidirectional sync

### Rationale

- **Minimal code change:** The transition table is the single source of truth — adding one entry enables the feature across all surfaces
- **Zero architectural risk:** Uses existing infrastructure (status transitions, TUI status menu, session metrics)
- **Validated need:** Explicitly identified in validation gate review
- **Low effort, high value:** Quality-of-life improvement that prevents data loss from accidental completion

### Effort & Risk

- **Effort:** Low (2-3 stories, ~1-2 days of implementation)
- **Risk:** Low (well-understood change, isolated impact, comprehensive existing tests)
- **Timeline impact:** None — can be parallelized with other active work

---

## Section 4: Detailed Change Proposals

### 4.1: Status Transition Model

```
File: internal/core/task_status.go
Section: validTransitions map

OLD:
StatusComplete:   {},

NEW:
StatusComplete:   {StatusTodo},
```

**Rationale:** Enable reverse transition from complete to todo. Only `todo` is allowed (not `in-progress` or other states) to keep the undo semantic clean — the task returns to the beginning of its lifecycle.

### 4.2: Task.UpdateStatus() — CompletedAt Clearing

```
File: internal/core/task.go
Section: UpdateStatus method

ADD (after existing CompletedAt setting logic):
if t.Status == StatusComplete && newStatus != StatusComplete {
    t.CompletedAt = nil
}
```

**Rationale:** When a task is un-completed, the `CompletedAt` timestamp should be cleared to maintain data model integrity.

### 4.3: PRD — New Functional Requirement

```
Section: Phase 6+ / Quality of Life

ADD:
**FR127:** The system shall support undoing task completion by transitioning
completed tasks back to `todo` status, clearing the `CompletedAt` timestamp,
removing the task from the completed log display, and logging an `undo_complete`
event in session metrics
```

### 4.4: Architecture — State Machine Update

```
File: docs/architecture/data-models.md
Section: Valid Transitions

OLD:
complete → [*]

NEW:
complete → todo (undo)
```

---

## Section 5: Implementation Handoff

### Change Scope: Minor

This change can be implemented directly by the development team without PO/SM or PM/Architect involvement beyond standard review.

### Handoff Plan

| Role | Responsibility |
|------|---------------|
| Dev (Worker Agent) | Implement Epic 32 stories via `/implement-story` |
| Merge Queue | Standard review and merge process |
| PR Shepherd | Standard rebase/conflict resolution |

### Success Criteria

1. `complete → todo` transition works in TUI, CLI, and programmatic API
2. `CompletedAt` is cleared when undoing completion
3. All existing status transition tests pass
4. New tests cover the undo transition path
5. Session metrics log `undo_complete` events
6. Adapters gracefully handle re-opened tasks (at minimum, don't crash)

---

## Approval

**Recommendation:** Proceed with Epic 32: Undo Task Completion as a P1 addition to the roadmap.

**Next Steps:**
1. Update PRD with FR127
2. Create architecture document for undo feature
3. Create Epic 32 with stories
4. Update ROADMAP.md
5. Begin implementation via worker agents
