# Architecture: Undo Task Completion (Epic 32)

**Date:** 2026-03-08
**Status:** Approved via Party Mode consensus
**PRD Requirements:** FR127–FR131

---

## Overview

Enable reversing task completion (`complete → todo`) by extending the existing status transition model. This is a minimal, surgical change to the table-driven transition system — no new packages, interfaces, or persistence formats required.

## Design Decisions

### DD-32.1: Reverse Transition Target

**Decision:** Only `complete → todo` is supported. Not `complete → in-progress` or `complete → blocked`.

**Rationale:** Undo returns the task to the beginning of its lifecycle. The user can then transition to `in-progress` via normal flow. Supporting multiple reverse targets adds complexity without value — the user's intent is "this task isn't actually done," not "this task is in a specific non-done state."

### DD-32.2: CompletedAt Handling

**Decision:** Clear `CompletedAt` to `nil` when transitioning from `complete` to `todo`.

**Rationale:** `CompletedAt` represents the factual state "this task is done at this time." When undoing completion, maintaining the old `CompletedAt` would violate the data model invariant that `CompletedAt` is only set when `status == complete`. The original completion timestamp is preserved in the `undo_complete` session metrics event for historical analysis.

### DD-32.3: Completed Log Immutability

**Decision:** The `completed.txt` file remains append-only. Undo does NOT remove entries.

**Rationale:** `completed.txt` is an audit trail. Modifying it on undo would break the append-only guarantees relied upon by analysis scripts (`scripts/daily_completions.sh`, `scripts/analyze_sessions.sh`). The session metrics `undo_complete` event provides the correction signal — any reporting tool can correlate completion and undo events for accurate counts.

### DD-32.4: No Time Limit

**Decision:** Undo is available regardless of when the task was completed.

**Rationale:** Accidental completion can be discovered at any time. Time-limited undo (e.g., "undo within 30 seconds") adds UX complexity (countdown timers, ephemeral UI) without meaningful benefit. The status menu approach requires deliberate user action (select task → open status menu → choose todo), which provides sufficient friction to prevent accidental undo.

### DD-32.5: Dependency Re-evaluation

**Decision:** When a completed task is undone and other tasks depend on it (FR113), dependent tasks are re-checked and may return to blocked state.

**Rationale:** The dependency system (Epic 29) uses `GetAvailableForDoors()` to filter tasks whose dependencies are not all complete. When a dependency is undone, any task that was unblocked by that completion must be re-evaluated. The existing dependency check logic handles this naturally — no special undo-aware code is needed in the dependency filter, because it already checks live status on every refresh.

---

## Component Changes

### internal/core/task_status.go

**Change:** Add `StatusTodo` to `validTransitions[StatusComplete]`

```go
// BEFORE
StatusComplete:   {},

// AFTER
StatusComplete:   {StatusTodo},
```

**Impact:** Enables `complete → todo` across all surfaces (TUI, CLI, MCP, adapters) automatically because all status change operations go through `IsValidTransition()`.

### internal/core/task.go — UpdateStatus()

**Change:** Clear `CompletedAt` when leaving complete status

```go
// Add before existing CompletedAt setting logic:
if t.Status == StatusComplete && newStatus != StatusComplete {
    t.CompletedAt = nil
}
```

**Impact:** Maintains data model invariant that `CompletedAt` is only set when status is complete or archived.

### internal/core/session_tracker.go

**Change:** Add `RecordUndoComplete()` method to log undo events

```go
func (st *SessionTracker) RecordUndoComplete(taskID string, originalCompletedAt time.Time) {
    // Log undo_complete event to JSONL session metrics
}
```

**Impact:** Enables behavioral analysis of accidental completions.

### internal/tui/ — StatusUpdateMenu

**Change:** None required. The status menu already reads from `GetValidTransitions()`, which will automatically include `todo` when the current status is `complete`.

**Impact:** The UX "just works" — users see `todo` as a valid option when viewing a completed task.

### internal/tui/ — Detail View

**Change:** Show confirmation message after undo completion (toast/status bar message).

### internal/cli/ — Status Command

**Change:** None required for existing `threedoors task status <id> <status>` command — it already validates transitions via the same `IsValidTransition()` function.

### Adapters (all)

**Change:** Ensure adapters handle tasks transitioning from `complete → todo` gracefully. For bidirectional sync adapters (Todoist, GitHub, Linear, Jira), the reverse sync (reopening a closed/completed item) is deferred to individual adapter epics. The core undo feature focuses on local state only.

---

## State Machine (Updated)

```mermaid
stateDiagram-v2
    [*] --> todo
    todo --> in-progress
    todo --> blocked
    todo --> complete
    todo --> deferred
    todo --> archived
    blocked --> todo
    blocked --> in-progress
    blocked --> complete
    in-progress --> blocked
    in-progress --> in-review
    in-progress --> complete
    in-review --> in-progress
    in-review --> complete
    complete --> todo : undo
    deferred --> todo
```

**Change from current:** Added `complete → todo` transition (labeled "undo").

---

## Data Flow

```
User Action: Select completed task → Status menu → Choose "todo"
    │
    ├─ IsValidTransition(complete, todo) → true (NEW)
    │
    ├─ Task.UpdateStatus(todo)
    │   ├─ Set status = todo
    │   ├─ Clear CompletedAt = nil
    │   ├─ Set UpdatedAt = now
    │   └─ Clear Blocker = ""
    │
    ├─ SessionTracker.RecordUndoComplete(taskID, originalCompletedAt)
    │   └─ Append undo_complete event to sessions.jsonl
    │
    ├─ TaskPool.UpdateTask(task)
    │   └─ Task now eligible for GetAvailableForDoors()
    │
    ├─ FileManager.SaveTasks() → atomic write to tasks.yaml
    │
    └─ TUI: Show confirmation toast, refresh doors view
```

---

## Testing Strategy

### Unit Tests

- `TestIsValidTransition_CompleteToTodo` — verify transition is now valid
- `TestIsValidTransition_CompleteToOthers` — verify only `todo` is valid target from complete (not `in-progress`, `blocked`, etc.)
- `TestUpdateStatus_UndoComplete_ClearsCompletedAt` — verify `CompletedAt` is nil after undo
- `TestUpdateStatus_UndoComplete_SetsUpdatedAt` — verify timestamp is updated
- `TestUpdateStatus_UndoComplete_ClearsBlocker` — verify blocker is cleared
- Regression: all existing transition tests pass unchanged

### Integration Tests

- TUI: Status menu shows `todo` for completed tasks
- CLI: `threedoors task status <id> todo` succeeds for completed tasks
- Session metrics: `undo_complete` event is logged correctly

### Edge Cases

- Undo a task completed weeks ago
- Undo a task that is a dependency for other tasks (re-blocks dependents)
- Undo a task from a bidirectional sync adapter (local undo only, no remote API call)
- Undo a task with notes — notes should be preserved

---

## Non-Goals

- General undo system (undo any action) — out of scope
- Undo history / undo stack — out of scope
- Time-limited undo with countdown UI — rejected in party mode
- Adapter-level "reopen" API calls — deferred to individual adapter epics
- Removing entries from completed.txt — rejected (immutable audit trail)
