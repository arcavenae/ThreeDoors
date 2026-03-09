---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/requirements.md (FR104-FR109)
  - docs/prd/product-scope.md (Phase 3.5)
  - _bmad-output/planning-artifacts/architecture-snooze-defer.md
  - _bmad-output/planning-artifacts/sprint-change-proposal-2026-03-07-snooze-defer.md
  - ux-workflow-improvements-research.md
---

# ThreeDoors - Epic 28: Snooze/Defer as First-Class Action

## Overview

This document provides the epic and story breakdown for Epic 28, decomposing PRD requirements FR104-FR109 and the Snooze/Defer architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

- FR104: Z-key snooze action with quick defer options (Tomorrow, Next Week, Pick Date, Someday) from doors view and detail view
- FR105: Snooze sets `defer_until` timestamp, transitions to deferred, removes from door eligibility
- FR106: Auto-return deferred tasks to todo when defer_until passes (startup + 1-minute tea.Tick)
- FR107: `:deferred` command showing snoozed tasks with un-snooze and edit-date actions
- FR108: Session metrics logging for snooze, snooze_return, and unsnooze events
- FR109: Additional status transitions: in-progress->deferred, blocked->deferred

### Non-Functional Requirements

- Existing NFR-CQ1 through NFR-CQ5 (code quality gates) apply to all stories
- DeferUntil field must serialize/deserialize through YAML round-trip without data loss
- SnoozeView must render correctly at terminal widths from 40 to 200 columns
- Auto-return tick must not impact TUI responsiveness (<1ms per check cycle)

### Additional Requirements (from Architecture)

- DeferUntil is `*time.Time` (nullable pointer), YAML tag `defer_until,omitempty`
- Snooze dates calculated in local time, stored as UTC
- "Someday" = StatusDeferred with nil DeferUntil (indefinite, manual un-snooze only)
- No change needed to GetAvailableForDoors filter (already excludes deferred status)
- Auto-return clears DeferUntil and sets status back to todo

### FR Coverage Map

- FR104: Story 28.2 (SnoozeView + Z-key binding)
- FR105: Story 28.1 (DeferUntil field) + Story 28.2 (snooze action)
- FR106: Story 28.1 (auto-return logic)
- FR107: Story 28.3 (DeferredListView + :deferred command)
- FR108: Story 28.4 (session metrics)
- FR109: Story 28.1 (status transition additions)

## Epic List

### Epic 28: Snooze/Defer as First-Class Action

Surface the existing `deferred` status as a first-class user action with date-based snooze, making the door pool trustworthy by ensuring only actionable tasks appear. Users can snooze tasks with quick date options, view and manage their snoozed tasks, and have tasks automatically return when their defer date arrives.

**FRs covered:** FR104, FR105, FR106, FR107, FR108, FR109
**Priority:** P1
**Dependencies:** None (builds on existing StatusDeferred infrastructure)

---

## Epic 28: Snooze/Defer as First-Class Action

Users can snooze tasks they can't act on right now, removing them from the door pool until a chosen return date. This makes every door presentation trustworthy — all three doors are genuinely actionable.

### Story 28.1: DeferUntil Field, Status Transitions, and Auto-Return Logic

As a ThreeDoors user,
I want deferred tasks to have a return date and automatically come back when that date arrives,
So that I can snooze tasks without worrying about forgetting them.

**Status:** Not Started
**Priority:** P1
**Depends On:** None (foundational story)
**FRs:** FR105, FR106, FR109

**Acceptance Criteria:**

**AC 28.1.1 — DeferUntil field persists through save/load cycle**
**Given** a task with DeferUntil set to "2026-03-08T14:00:00Z"
**When** the task pool is saved to YAML and reloaded
**Then** the task's DeferUntil value is preserved exactly
**And** a task with nil DeferUntil serializes with no defer_until key in YAML

**AC 28.1.2 — New status transitions are valid**
**Given** a task with status "in-progress"
**When** the task transitions to "deferred"
**Then** the transition succeeds without error
**And** the same transition is valid from "blocked" to "deferred"

**AC 28.1.3 — Auto-return on startup**
**Given** a task with status "deferred" and DeferUntil in the past
**When** the application starts and loads the task pool
**Then** the task's status is set to "todo"
**And** the task's DeferUntil is cleared to nil
**And** the task is eligible for door selection

**AC 28.1.4 — Auto-return during session via tea.Tick**
**Given** a task with status "deferred" and DeferUntil set to 1 minute from now
**When** the 1-minute defer-return tick fires after the DeferUntil has passed
**Then** the task's status is set to "todo"
**And** the task's DeferUntil is cleared to nil
**And** the doors view refreshes to include the returned task

**AC 28.1.5 — Someday tasks do not auto-return**
**Given** a task with status "deferred" and DeferUntil is nil (Someday snooze)
**When** the auto-return check runs
**Then** the task remains in "deferred" status
**And** the task does not appear in door selection

**AC 28.1.6 — Deferred tasks excluded from doors (existing behavior verified)**
**Given** a task with status "deferred" (with or without DeferUntil)
**When** GetAvailableForDoors() is called
**Then** the deferred task is not in the returned list

**Tasks:**

1. Add `DeferUntil *time.Time` field to Task struct in `internal/core/task.go` with YAML tag `defer_until,omitempty`
2. Add `StatusInProgress -> StatusDeferred` and `StatusBlocked -> StatusDeferred` transitions in `internal/core/task_status.go`
3. Implement `CheckDeferredReturns(pool *TaskPool) int` function that iterates deferred tasks and returns expired ones to todo
4. Call CheckDeferredReturns on startup in task pool initialization
5. Add DeferReturnTickMsg type and 1-minute tea.Every tick in MainModel
6. Handle DeferReturnTickMsg in MainModel.Update to run CheckDeferredReturns and refresh doors
7. Write unit tests: DeferUntil serialization round-trip, new transitions, auto-return logic, Someday handling, boundary conditions
8. Write regression test: verify GetAvailableForDoors still excludes deferred tasks (existing behavior)
9. Run `make fmt && make lint && make test` — all must pass

---

### Story 28.2: Snooze TUI View and Z-Key Binding

As a ThreeDoors user,
I want to press Z on a door or in the detail view to quickly snooze a task with date options,
So that I can remove non-actionable tasks from my doors without losing them.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 28.1 (DeferUntil field and status transitions must exist)
**FRs:** FR104, FR105

**Acceptance Criteria:**

**AC 28.2.1 — Z key opens SnoozeView from doors**
**Given** the user is in the doors view with a door selected
**When** the user presses the Z key
**Then** a SnoozeView overlay appears showing four options: Tomorrow, Next Week, Pick Date, Someday

**AC 28.2.2 — Z key opens SnoozeView from detail view**
**Given** the user is viewing a task in the detail view
**When** the user presses the Z key
**Then** a SnoozeView overlay appears showing the same four options

**AC 28.2.3 — Tomorrow snooze calculates correct date**
**Given** the SnoozeView is open
**When** the user selects "Tomorrow" and presses Enter
**Then** the task's DeferUntil is set to the next day at 09:00 local time (stored as UTC)
**And** the task's status transitions to "deferred"
**And** the view returns to doors with the snoozed task removed

**AC 28.2.4 — Next Week snooze calculates correct date**
**Given** the SnoozeView is open
**When** the user selects "Next Week" and presses Enter
**Then** the task's DeferUntil is set to the next Monday at 09:00 local time (stored as UTC)
**And** the task's status transitions to "deferred"

**AC 28.2.5 — Pick Date accepts user input**
**Given** the SnoozeView is open
**When** the user selects "Pick Date" and presses Enter
**Then** a text input appears for date entry in YYYY-MM-DD format
**And** on valid date entry and Enter, the task is snoozed until that date at 09:00 local time
**And** on invalid date format, an error message is shown and input remains active

**AC 28.2.6 — Someday snooze sets indefinite deferral**
**Given** the SnoozeView is open
**When** the user selects "Someday" and presses Enter
**Then** the task's DeferUntil is set to nil
**And** the task's status transitions to "deferred"

**AC 28.2.7 — ESC cancels snooze**
**Given** the SnoozeView is open
**When** the user presses ESC
**Then** the view closes without any changes to the task
**And** the user returns to their previous view (doors or detail)

**AC 28.2.8 — Z key ignored without door selection**
**Given** the user is in the doors view with no door selected
**When** the user presses the Z key
**Then** nothing happens (no SnoozeView opens)

**Tasks:**

1. Create `internal/tui/snooze_view.go` with SnoozeView struct, NewSnoozeView, Update, View
2. Implement date calculation functions: tomorrow9am(), nextMonday9am(), parsePickDate(input)
3. Add Z-key handler in DoorsView.Update — open SnoozeView when door is selected
4. Add Z-key handler in TaskDetailView.Update — open SnoozeView for current task
5. Add "[Z] Snooze" to TaskDetailView.View() options bar
6. Define TaskSnoozedMsg and handle in MainModel (save, refresh doors)
7. Add SnoozeView routing in MainModel (open/close, overlay management)
8. Write unit tests: date calculation functions, SnoozeView state transitions, Z-key routing
9. Write golden file tests for SnoozeView rendered output at standard widths
10. Run `make fmt && make lint && make test` — all must pass

---

### Story 28.3: Deferred List View and :deferred Command

As a ThreeDoors user,
I want to see all my snoozed tasks in one place and be able to un-snooze or change their return dates,
So that I can manage my deferred tasks and bring them back early if needed.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 28.1 (DeferUntil field must exist)
**Can parallel with:** Story 28.2

**FRs:** FR107

**Acceptance Criteria:**

**AC 28.3.1 — :deferred command opens DeferredListView**
**Given** the user is in any view with the command palette available
**When** the user types `:deferred` and presses Enter
**Then** the DeferredListView opens showing all tasks with status "deferred"

**AC 28.3.2 — Tasks sorted by return date**
**Given** the DeferredListView is open with multiple snoozed tasks
**When** the view renders
**Then** tasks are sorted by DeferUntil ascending (soonest first)
**And** tasks with nil DeferUntil ("Someday") appear last
**And** each task shows its text (truncated if needed), return date or "Someday", and time remaining

**AC 28.3.3 — Un-snooze returns task to todo**
**Given** the DeferredListView is open with tasks listed
**When** the user navigates to a task and presses 'u'
**Then** the task's status is set to "todo"
**And** the task's DeferUntil is cleared to nil
**And** the task is removed from the deferred list
**And** the task becomes eligible for door selection

**AC 28.3.4 — Edit date reopens SnoozeView**
**Given** the DeferredListView is open with tasks listed
**When** the user navigates to a task and presses 'e'
**Then** the SnoozeView opens for that task with the same four options
**And** on snooze confirmation, the deferred list refreshes with the updated date

**AC 28.3.5 — ESC returns to previous view**
**Given** the DeferredListView is open
**When** the user presses ESC
**Then** the view closes and returns to the previous view

**AC 28.3.6 — Empty state message**
**Given** there are no deferred tasks
**When** the user opens `:deferred`
**Then** a message is shown: "No snoozed tasks. Use Z on a door to snooze a task."

**Tasks:**

1. Create `internal/tui/deferred_list_view.go` with DeferredListView struct, NewDeferredListView, Update, View
2. Implement task sorting: by DeferUntil ascending, nil values last
3. Implement relative time display ("Tomorrow", "3 days", "Someday")
4. Add 'u' key handler for un-snooze (set todo, clear DeferUntil)
5. Add 'e' key handler to open SnoozeView for selected task
6. Add j/k and arrow key navigation
7. Register `:deferred` command in command palette routing
8. Add DeferredListView routing in MainModel
9. Write unit tests: sorting logic, un-snooze behavior, empty state
10. Write golden file tests for DeferredListView rendered output
11. Run `make fmt && make lint && make test` — all must pass

---

### Story 28.4: Session Metrics Logging for Snooze Events

As a ThreeDoors developer/analyst,
I want snooze events logged to session metrics,
So that usage patterns can be analyzed and the feature can be improved over time.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 28.2 (snooze action must exist), Story 28.1 (auto-return must exist)
**FRs:** FR108

**Acceptance Criteria:**

**AC 28.4.1 — Snooze events logged**
**Given** a user snoozes a task via the Z key
**When** the snooze is confirmed
**Then** a `snooze` event is appended to the JSONL session metrics log
**And** the event includes: task_id, defer_until (null for Someday), option ("tomorrow"/"next_week"/"pick_date"/"someday"), timestamp

**AC 28.4.2 — Auto-return events logged**
**Given** a deferred task's DeferUntil has passed
**When** the auto-return logic transitions it back to todo
**Then** a `snooze_return` event is appended to the session metrics log
**And** the event includes: task_id, timestamp

**AC 28.4.3 — Manual un-snooze events logged**
**Given** a user un-snoozes a task from the DeferredListView
**When** the un-snooze is confirmed
**Then** an `unsnooze` event is appended to the session metrics log
**And** the event includes: task_id, timestamp

**AC 28.4.4 — Metrics format consistent with existing events**
**Given** the existing session metrics use JSONL format at `~/.threedoors/sessions.jsonl`
**When** snooze events are logged
**Then** they follow the same JSONL structure and field conventions as existing event types

**Tasks:**

1. Define snooze/snooze_return/unsnooze event types in session metrics module
2. Log snooze event in MainModel's TaskSnoozedMsg handler
3. Log snooze_return event in CheckDeferredReturns when tasks auto-return
4. Log unsnooze event in DeferredListView's un-snooze handler
5. Write unit tests: verify event structure, field presence, JSONL format
6. Run `make fmt && make lint && make test` — all must pass

---

## Story Dependency Graph

```
28.1 (DeferUntil + Auto-Return + Transitions)
  |
  ├── 28.2 (SnoozeView + Z-Key)
  |     |
  |     └── 28.4 (Session Metrics) ← also depends on 28.1
  |
  └── 28.3 (DeferredListView + :deferred)
```

Stories 28.2 and 28.3 can be implemented in parallel after 28.1 completes.
Story 28.4 depends on both 28.1 and 28.2 but is lightweight.

## Estimated Scope

| Story | Files Changed | Files Created | Complexity |
|-------|--------------|---------------|------------|
| 28.1 | 3 (task.go, task_status.go, main_model.go) | 0 | Low |
| 28.2 | 3 (doors_view.go, task_detail_view.go, main_model.go) | 1 (snooze_view.go) | Medium |
| 28.3 | 2 (main_model.go, command_palette.go) | 1 (deferred_list_view.go) | Medium |
| 28.4 | 3 (main_model.go, deferred_list_view.go, session metrics) | 0 | Low |
