# Epic 32: Undo Task Completion

**Priority:** P1
**Status:** Backlog
**PRD Requirements:** FR127–FR131
**Architecture:** architecture-undo-task-completion.md
**Party Mode:** party-mode-undo-task-completion-2026-03-08.md
**Sprint Change Proposal:** sprint-change-proposal-2026-03-08-undo-task-completion.md

---

## Epic Summary

Enable users to reverse accidental task completion by adding a `complete → todo` status transition. Addresses a validated pain point from the Phase 1 Validation Gate review: "No undo for task completion. Once a task is marked done, recovery requires manual file editing."

## Design Decisions

- **DD-32.1:** Only `complete → todo` transition supported (not `complete → in-progress`)
- **DD-32.2:** `CompletedAt` cleared to nil on undo
- **DD-32.3:** `completed.txt` remains immutable; undo tracked via session metrics
- **DD-32.4:** No time limit on undo
- **DD-32.5:** Dependency re-evaluation triggers automatically via existing filters

## Stories

### Story 32.1: Status Model — Complete-to-Todo Transition

**Priority:** P1
**Depends On:** None
**Estimated Effort:** Small

#### Description

Add `complete → todo` to the status transition model and handle `CompletedAt` clearing when a task is un-completed. This is the foundational change that enables undo across all surfaces.

#### Acceptance Criteria

1. `IsValidTransition(StatusComplete, StatusTodo)` returns `true`
2. `IsValidTransition(StatusComplete, StatusInProgress)` returns `false` (no other reverse transitions added)
3. `IsValidTransition(StatusComplete, StatusBlocked)` returns `false`
4. `Task.UpdateStatus(StatusTodo)` succeeds when current status is `complete`
5. After undo, `CompletedAt` is `nil`
6. After undo, `UpdatedAt` is set to current UTC time
7. After undo, `Blocker` is empty string
8. After undo, `Status` is `todo`
9. Task notes are preserved through the undo transition
10. All existing status transition tests pass unchanged (regression)
11. `GetValidTransitions(StatusComplete)` returns `[StatusTodo]`

#### Technical Notes

- Modify `validTransitions` map in `internal/core/task_status.go`
- Add `CompletedAt` clearing logic in `Task.UpdateStatus()` in `internal/core/task.go`
- Add table-driven tests for the new transition in `internal/core/task_status_test.go`
- Add tests for `CompletedAt` clearing in `internal/core/task_test.go`

#### Tasks

- [ ] Add `StatusTodo` to `validTransitions[StatusComplete]` in `task_status.go`
- [ ] Add `CompletedAt` clearing logic when leaving complete status in `task.go`
- [ ] Add unit tests for `complete → todo` transition validity
- [ ] Add unit tests for `CompletedAt` clearing on undo
- [ ] Add unit tests verifying other reverse transitions remain invalid
- [ ] Verify all existing tests pass (regression)
- [ ] Run `make fmt && make lint && make test`

---

### Story 32.2: Session Metrics — Undo Complete Event Logging

**Priority:** P1
**Depends On:** 32.1
**Estimated Effort:** Small

#### Description

Log `undo_complete` events in session metrics when a task completion is reversed. Capture the task ID, original completion timestamp, and time elapsed since completion for behavioral analysis of accidental completions.

#### Acceptance Criteria

1. When a task transitions from `complete` to `todo`, an `undo_complete` event is appended to the JSONL session log
2. The event includes: task ID, original `CompletedAt` timestamp, time elapsed since completion, and current session timestamp
3. The event type is `undo_complete` (consistent with existing event naming: `snooze`, `snooze_return`)
4. The `completed.txt` file is NOT modified when undo occurs
5. Session metrics reader (`ReadAll`, `ReadSince`, `ReadLast`) correctly parse `undo_complete` events
6. Event logging does not block the TUI update loop (async via `tea.Cmd`)

#### Technical Notes

- Add `RecordUndoComplete(taskID string, originalCompletedAt time.Time)` to `SessionTracker`
- Follow existing event logging patterns in `internal/core/session_tracker.go`
- The session tracker's JSONL format is append-only, matching `completed.txt` immutability
- Test with existing metrics reader infrastructure

#### Tasks

- [ ] Add `undo_complete` event type constant
- [ ] Add `RecordUndoComplete()` method to `SessionTracker`
- [ ] Wire up undo event logging in the TUI status change handler
- [ ] Add unit tests for event logging
- [ ] Verify `completed.txt` is not modified on undo
- [ ] Run `make fmt && make lint && make test`

---

### Story 32.3: TUI & CLI Undo Experience

**Priority:** P1
**Depends On:** 32.1, 32.2
**Estimated Effort:** Small

#### Description

Ensure the TUI and CLI surfaces provide a smooth undo experience. The TUI status menu should automatically show `todo` as a valid option when viewing completed tasks (this happens naturally via `GetValidTransitions`). Add a confirmation toast message after undo. Ensure the CLI `task status` command supports the transition.

#### Acceptance Criteria

1. In TUI detail view of a completed task, pressing `S` (status menu) shows `todo` as a valid transition option
2. After selecting `todo` from the status menu of a completed task, the task returns to `todo` status
3. A brief confirmation message is shown: "Task uncompleted — returned to todo"
4. After undo, the doors view refreshes and the uncompleted task is eligible for door selection
5. CLI command `threedoors task status <id> todo` succeeds when the task is currently `complete`
6. CLI outputs confirmation message when undo succeeds
7. If the undone task was a dependency for other tasks, those dependents are re-evaluated (tested with dependency filter)

#### Technical Notes

- TUI StatusUpdateMenu reads from `GetValidTransitions()` — no menu code changes needed
- Add toast/status bar message to TUI after undo
- CLI `status` subcommand already validates via `IsValidTransition` — should work automatically
- Test dependency re-evaluation: complete task A (unblocks B), undo task A (re-blocks B)

#### Tasks

- [ ] Verify TUI status menu shows `todo` for completed tasks (may need no code change)
- [ ] Add confirmation toast/message after undo in TUI
- [ ] Verify CLI `task status` command handles undo
- [ ] Add TUI integration test for undo flow
- [ ] Add dependency re-evaluation test (if Epic 29 is implemented)
- [ ] Run `make fmt && make lint && make test`

---

## Story Dependency Graph

```
32.1 (Status Model)
  ├── 32.2 (Session Metrics)
  │     └── 32.3 (TUI & CLI)
  └── 32.3 (TUI & CLI)
```

## Epic Completion Criteria

1. Users can undo task completion from both TUI and CLI
2. `CompletedAt` is properly cleared on undo
3. Session metrics capture `undo_complete` events
4. `completed.txt` remains immutable
5. All existing tests pass (no regressions)
6. Dependency re-evaluation works correctly when a dependency is undone
