# Architecture: Expand/Fork Key Implementations (Epic 31)

**Date:** 2026-03-08
**Status:** Proposed
**Decision Source:** Design Decision H9, Party Mode 2026-03-08

---

## Overview

This document defines the technical architecture for completing the Expand (manual sub-task creation) and Fork (variant creation) features in the ThreeDoors TUI detail view. Both features have basic stub implementations that need enhancement to fulfill the specifications from Design Decision H9.

---

## 1. Task Model Extension: ParentID

### Decision: Native Field on Task Struct

The `ParentID` field belongs on the `core.Task` struct, not in the enrichment DB's `CrossReference` system.

**Rationale:**
- Parent-child is a core domain relationship, not optional metadata
- `TaskPool` must answer "give me children of X" without enrichment DB dependency
- Maintains clean separation: `internal/core` has zero imports on enrichment
- Backward-compatible: optional YAML field with `omitempty`

### Schema Change

```go
// In internal/core/task.go
type Task struct {
    // ... existing fields ...
    ParentID *string `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
}
```

### YAML Format

```yaml
tasks:
  - id: parent-uuid
    text: "Write architecture document"
    status: todo
    # no parent_id field — this is a root task

  - id: child-uuid
    text: "Draft data models section"
    status: in-progress
    parent_id: parent-uuid  # links to parent
```

### TaskPool Extensions

```go
// GetSubtasks returns all tasks whose ParentID matches the given task ID.
func (p *TaskPool) GetSubtasks(parentID string) []*Task

// HasSubtasks returns true if any task in the pool has this task as parent.
func (p *TaskPool) HasSubtasks(taskID string) bool
```

### Door Selection Filter

`GetAvailableForDoors()` adds a new exclusion rule:
- If `HasSubtasks(task.ID)` is true, exclude the task from door selection
- This communicates "you decomposed this task, work the pieces"

---

## 2. Expand Feature Architecture

### Current State

```
User presses E → DetailModeExpandInput → single text input → Enter creates ExpandTaskMsg → main model creates unlinked task
```

### Target State

```
User presses E → DetailModeExpandInput → text input → Enter creates subtask (ParentID set) + stays in expand mode → Esc exits
```

### Message Types

```go
// ExpandTaskMsg is emitted when a subtask is created via Expand.
type ExpandTaskMsg struct {
    ParentTask  *core.Task
    NewTaskText string
}
```

No new message types needed — the existing `ExpandTaskMsg` is sufficient. The main model handler sets `ParentID` on the new task.

### Main Model Handler

```go
case ExpandTaskMsg:
    newTask := core.NewTask(msg.NewTaskText)
    parentID := msg.ParentTask.ID
    newTask.ParentID = &parentID
    // Add to pool, persist...
```

### Sequential Input Mode

The `handleExpandInput` method changes behavior on Enter:
- Creates the subtask
- Clears the input buffer
- Increments a subtask counter for display
- Does NOT exit `DetailModeExpandInput`
- Only Esc exits back to `DetailModeView`

### Detail View Rendering

When viewing a parent task, render subtask list:

```
Write architecture document

  ├─ [TODO]  Draft high-level overview
  ├─ [DONE] Data models section
  └─ [TODO]  Components section

Subtasks: 1/3 complete
```

Implementation: `DetailView.View()` checks `pool.GetSubtasks(task.ID)` and renders the list between task text and the separator line.

---

## 3. Fork Feature Architecture

### Current State

```
User presses F → core.NewTask(dv.task.Text) → TaskAddedMsg emitted
```

### Target State

```
User presses F → core.ForkTask(dv.task) → TaskForkedMsg emitted → main model creates enrichment cross-ref
```

### ForkTask Factory

```go
// ForkTask creates a variant of the given task.
// Preserves: Text, Context, Effort, Tags
// Resets: Status (todo), Blocker (""), Notes (empty), timestamps (now)
// Adds: Note "Forked from: [truncated text]"
// Does NOT copy: ParentID
func ForkTask(original *Task) *Task {
    forked := NewTask(original.Text)
    forked.Context = original.Context
    forked.Effort = original.Effort
    forked.Tags = append([]string{}, original.Tags...)  // defensive copy

    truncated := original.Text
    if len(truncated) > 60 {
        truncated = truncated[:57] + "..."
    }
    forked.AddNote("Forked from: " + truncated)

    return forked
}
```

### New Message Type

```go
// TaskForkedMsg is emitted when a task variant is created via Fork.
type TaskForkedMsg struct {
    Original *core.Task
    Variant  *core.Task
}
```

### Cross-Reference (Enrichment Layer)

The main model's `TaskForkedMsg` handler creates a cross-reference:

```go
case TaskForkedMsg:
    // Add variant to pool and persist
    // Then create enrichment cross-reference
    ref := &enrichment.CrossReference{
        SourceTaskID: msg.Original.ID,
        TargetTaskID: msg.Variant.ID,
        SourceSystem: "local",
        Relationship: "forked-from",
    }
    enrichDB.AddCrossReference(ref)
```

This keeps `internal/core` free of enrichment DB awareness.

---

## 4. Component Changes Summary

| Component | Change | Blast Radius |
|-----------|--------|-------------|
| `core.Task` | Add `ParentID *string` field | Low — optional field, backward-compatible |
| `core.TaskPool` | Add `GetSubtasks()`, `HasSubtasks()` methods | Low — new methods only |
| `core.TaskPool.GetAvailableForDoors()` | Add parent exclusion filter | Low — additive filter |
| `core.ForkTask()` | New factory function | None — new code |
| `tui.DetailView` | Sequential expand mode, subtask rendering | Medium — modifies existing view |
| `tui.DetailView` | Fork uses ForkTask + TaskForkedMsg | Low — replaces 2 lines |
| `tui.MainModel` | Handle TaskForkedMsg, set ParentID on expand | Low — new message handler |
| YAML schema | Add `parent_id` field | Low — optional, omitempty |

---

## 5. Testing Strategy

### Unit Tests
- `TestForkTask` — verifies field preservation/reset semantics
- `TestGetSubtasks` — returns correct children, empty for no children
- `TestHasSubtasks` — returns true/false correctly
- `TestGetAvailableForDoors_ExcludesParents` — parents with children excluded

### Integration Tests
- Expand creates subtask with correct ParentID
- Sequential expand creates multiple subtasks
- Fork creates variant with correct field values
- YAML round-trip with parent_id field (backward compatibility)

### TUI Tests
- DetailView renders subtask list correctly
- Expand mode stays open after Enter, exits on Esc
- Fork emits TaskForkedMsg (not TaskAddedMsg)

---

## 6. Design Constraints

1. **Single-level nesting only** — subtasks cannot have their own subtasks (v1 simplification)
2. **No property inheritance** — subtasks are independent work items
3. **No auto-completion** — parent never auto-completes when all children finish
4. **Completion ratio is display-only** — no enforcement of "complete all subtasks first"
5. **ParentID is immutable** — once set, a subtask cannot be re-parented

---

## 7. Migration & Backward Compatibility

- Existing tasks without `parent_id` field load correctly (nil pointer = no parent)
- No schema version bump needed — field is additive with `omitempty`
- No migration script required
- Existing `ExpandTaskMsg` handling remains compatible (ParentID set in handler, not in message)
