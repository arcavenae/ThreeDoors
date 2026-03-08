# Epic 31: Expand/Fork Key Implementations

**Priority:** P2
**Status:** Backlog
**Dependencies:** Epic 3 (Enhanced Interaction) — complete, Epic 13 (Multi-Source Aggregation) — complete
**Estimated Effort:** 3-5 days across all stories
**Design Decision:** H9 (Expand = manual sub-tasks, Fork = variant creation)

## Epic Summary

Complete the Expand (`[E]`) and Fork (`[F]`) key actions in the TUI detail view. Expand enables manual sub-task creation with parent-child relationship tracking. Fork enables variant creation that copies relevant fields while resetting status for a fresh start. Both features have basic stub implementations that need enhancement to fulfill the specifications from Design Decision H9.

## Stories

---

### Story 31.1: Task Model ParentID Extension

**Priority:** P2
**Status:** Not Started
**Depends On:** None (foundational)
**Estimated Effort:** Small (1-2 hours)

#### Description

Add a native `ParentID` field to the `core.Task` struct to support parent-child relationships between tasks. Extend `TaskPool` with methods to query subtasks and check parent status. Update `GetAvailableForDoors()` to exclude parent tasks that have children.

#### Acceptance Criteria

- [ ] `core.Task` has `ParentID *string` field with `yaml:"parent_id,omitempty"` tag
- [ ] `TaskPool.GetSubtasks(parentID string) []*Task` returns all tasks whose ParentID matches
- [ ] `TaskPool.HasSubtasks(taskID string) bool` returns true if any task has this ID as parent
- [ ] `GetAvailableForDoors()` excludes tasks that have one or more subtasks
- [ ] YAML round-trip preserves `parent_id` field when present
- [ ] Existing tasks without `parent_id` load correctly (nil = no parent)
- [ ] Table-driven tests for GetSubtasks, HasSubtasks, and door exclusion
- [ ] `make lint && make test` pass with zero warnings

#### Technical Notes

- ParentID is a `*string` (pointer) — nil means no parent, non-nil means subtask
- No schema version bump needed — field is additive with omitempty
- Single-level nesting only: no validation that parent itself has no parent (v1 simplification)
- Use existing `TaskPool.GetAllTasks()` to iterate for subtask queries

---

### Story 31.2: Enhanced Expand — Sequential Subtask Creation

**Priority:** P2
**Status:** Not Started
**Depends On:** 31.1

#### Description

Enhance the Expand feature to create subtasks with `ParentID` set to the parent task's ID. Implement sequential subtask creation mode: after submitting one subtask (Enter), the input stays open for the next subtask. Only Esc exits expand mode. Show a running subtask count during input.

#### Acceptance Criteria

- [ ] Pressing `E` in detail view enters expand mode with prompt "New subtask (Enter to add, Esc to finish):"
- [ ] Enter creates a new task with `ParentID` set to the current task's ID
- [ ] After Enter, expand mode stays active with cleared input and updated count: "Subtask N added. Next subtask (Esc to finish):"
- [ ] Esc exits expand mode and returns to detail view
- [ ] Empty input on Enter shows flash message "Subtask text cannot be empty" and stays in expand mode
- [ ] `ExpandTaskMsg` handler in main model sets `ParentID` on the new task
- [ ] Table-driven tests for sequential subtask creation
- [ ] `make lint && make test` pass

#### Technical Notes

- Add `expandCount int` field to DetailView to track subtasks added in current session
- `handleExpandInput` on Enter: emit ExpandTaskMsg, clear input, increment count, do NOT change mode
- `handleExpandInput` on Esc: reset count, change mode to DetailModeView
- Main model handler: `newTask.ParentID = &msg.ParentTask.ID`

---

### Story 31.3: Subtask List Rendering in Detail View

**Priority:** P2
**Status:** Not Started
**Depends On:** 31.1, 31.2

#### Description

Render a subtask list in the detail view when the viewed task has children. Show each subtask's status icon and text in an indented tree format with a completion ratio summary.

#### Acceptance Criteria

- [ ] Detail view shows subtask list between task text/notes and the separator line when task has children
- [ ] Subtask list uses tree characters: `├─` for non-last items, `└─` for last item
- [ ] Each subtask shows status in brackets: `[TODO]`, `[DONE]`, `[BLOCKED]`, `[IN-PROGRESS]`
- [ ] Completion ratio displayed: "Subtasks: N/M complete"
- [ ] Subtask text truncated to fit terminal width minus indent
- [ ] No subtask section rendered for tasks without children
- [ ] Status colors applied to subtask status brackets (reuse existing StatusColor)
- [ ] Table-driven tests for rendering with 0, 1, and multiple subtasks
- [ ] `make lint && make test` pass

#### Technical Notes

- Use `dv.pool.GetSubtasks(dv.task.ID)` to fetch children
- Render between notes section and separator line in `View()`
- Use `lipgloss` for status coloring — reuse `StatusColor()` function
- Truncate subtask text to `w - 10` (accounting for indent + status)

---

### Story 31.4: Enhanced Fork — Variant Creation with ForkTask Factory

**Priority:** P2
**Status:** Not Started
**Depends On:** None (can be implemented in parallel with 31.1-31.3)

#### Description

Replace the current Fork implementation (simple `NewTask` copy) with a `ForkTask` factory method that creates a proper variant: preserving Text, Context, Effort, and Tags while resetting Status to todo, clearing Blocker and Notes, and adding a "Forked from" note. Emit a `TaskForkedMsg` so the main model can create an enrichment DB cross-reference.

#### Acceptance Criteria

- [ ] `core.ForkTask(original *Task) *Task` factory method exists
- [ ] ForkTask preserves: Text, Context, Effort, Tags (defensive copy of Tags slice)
- [ ] ForkTask resets: Status to `todo`, Blocker to `""`, Notes to empty
- [ ] ForkTask sets fresh timestamps (CreatedAt, UpdatedAt = time.Now().UTC())
- [ ] ForkTask generates new UUID (not copied from original)
- [ ] ForkTask adds note: "Forked from: [text truncated to 60 chars]"
- [ ] ForkTask does NOT copy ParentID (fork is independent)
- [ ] `TaskForkedMsg{Original, Variant}` message type exists
- [ ] Detail view `F` key emits `TaskForkedMsg` instead of `TaskAddedMsg`
- [ ] Main model handles `TaskForkedMsg`: adds variant to pool, persists, creates enrichment cross-reference with `forked-from` relationship
- [ ] Flash message displayed: "Forked! Variant added to pool"
- [ ] Table-driven tests for ForkTask field semantics
- [ ] Integration test: fork creates cross-reference in enrichment DB
- [ ] `make lint && make test` pass

#### Technical Notes

- `ForkTask` lives in `internal/core/task.go` alongside `NewTask`
- Tags defensive copy: `forked.Tags = append([]string{}, original.Tags...)`
- Cross-reference creation in main model handler, NOT in core package
- Enrichment DB cross-reference: `SourceTaskID=original.ID`, `TargetTaskID=variant.ID`, `Relationship="forked-from"`
- If enrichment DB is nil, skip cross-reference (fork still works without it)

---

### Story 31.5: Design Decision H9 Status Update

**Priority:** P2
**Status:** Not Started
**Depends On:** 31.1-31.4

#### Description

Update `docs/design-decisions-needed.md` to mark Decision H9 as implemented (Epic 31). Update story file statuses as implementation progresses.

#### Acceptance Criteria

- [ ] H9 entry updated: "Implementation: Epic 31 (Expand/Fork Key Implementations)"
- [ ] H9 blocked status updated: "Blocked: None — Epic 31 created"
- [ ] All story files (31.1-31.5) have correct status reflecting implementation state
- [ ] No other changes to design-decisions-needed.md

---

## Epic Dependencies Diagram

```
31.1 (ParentID) ──→ 31.2 (Expand) ──→ 31.3 (Rendering)
                                    └──→ 31.5 (Docs)
31.4 (Fork) ─────────────────────────┘
```

Stories 31.1 and 31.4 can be implemented in parallel. Story 31.2 depends on 31.1. Story 31.3 depends on both 31.1 and 31.2. Story 31.5 is a documentation update after implementation.

## ROADMAP Integration

Add to ROADMAP.md under Active Epics:

```markdown
### Epic 31: Expand/Fork Key Implementations (P2) — 0/5 stories done

Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 31.1 | Task Model ParentID Extension | Not Started | P2 | None |
| 31.2 | Enhanced Expand — Sequential Subtask Creation | Not Started | P2 | 31.1 |
| 31.3 | Subtask List Rendering in Detail View | Not Started | P2 | 31.1, 31.2 |
| 31.4 | Enhanced Fork — Variant Creation with ForkTask Factory | Not Started | P2 | None |
| 31.5 | Design Decision H9 Status Update | Not Started | P2 | 31.1-31.4 |
```
