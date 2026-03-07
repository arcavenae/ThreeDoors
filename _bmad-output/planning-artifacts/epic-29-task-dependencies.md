---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/requirements.md (FR110-FR115)
  - docs/prd/product-scope.md (Phase 3.5+)
  - _bmad-output/planning-artifacts/architecture-task-dependencies.md
  - _bmad-output/planning-artifacts/sprint-change-proposal-2026-03-07-task-dependencies.md
  - docs/research/ux-workflow-improvements-research.md
---

# ThreeDoors - Epic 29: Task Dependencies & Blocked-Task Filtering

## Overview

This document provides the epic and story breakdown for Epic 29, decomposing PRD requirements FR110-FR115 and the Task Dependencies architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

- FR110: `depends_on` field on tasks containing list of task IDs, stored in YAML, persisted through enrichment DB for cross-provider deps
- FR111: Auto-filter tasks with incomplete dependencies from door selection; orphaned dependency IDs treated as unmet (pessimistic)
- FR112: "Blocked by: [task text]" indicator in doors and detail views, truncated to 40 chars with "+N more" count
- FR113: Auto-unblock check on task completion, emit `dependency_unblocked` event, refresh doors
- FR114: Dependency management in detail view: `+` to add (task search/picker), `-` to remove
- FR115: Circular dependency detection via DFS, reject with user-visible error message

### Non-Functional Requirements

- Existing NFR-CQ1 through NFR-CQ5 (code quality gates) apply to all stories
- DependsOn field must serialize/deserialize through YAML round-trip without data loss
- Dependency check in GetAvailableForDoors must complete in <1ms for pools up to 1000 tasks
- Cycle detection must complete in <1ms for dependency graphs up to 100 nodes

### Additional Requirements (from Architecture)

- DependsOn is `[]string` (task ID slice), YAML tag `depends_on,omitempty`
- DependencyResolver as pure functions in `internal/core/dependency.go` (not embedded in TaskPool)
- Pessimistic orphaned dependency handling: missing task ID = unmet dependency
- No reverse index for v1 — iterate pool on completion events
- "Blocked by" indicator shows first blocker text (40 chars) + count of additional blockers

### FR Coverage Map

- FR110: Story 29.1 (DependsOn field + YAML persistence)
- FR111: Story 29.2 (door selection filter)
- FR112: Story 29.3 (TUI blocked-by indicator)
- FR113: Story 29.2 (auto-unblock on completion)
- FR114: Story 29.3 (dependency management in detail view)
- FR115: Story 29.1 (cycle detection in DependencyResolver)

## Epic List

### Epic 29: Task Dependencies & Blocked-Task Filtering

Add a native dependency graph for tasks, ensuring the Three Doors only present genuinely actionable tasks by automatically filtering those whose prerequisites are incomplete. Users can declare task dependencies, see which tasks are blocked and why, and have tasks automatically become available when their dependencies complete.

**FRs covered:** FR110, FR111, FR112, FR113, FR114, FR115
**Priority:** P1
**Dependencies:** None (builds on existing Task model and TaskPool infrastructure)

---

## Epic 29: Task Dependencies & Blocked-Task Filtering

Every door presented should be genuinely actionable. Task dependencies make the door pool trustworthy by ensuring tasks that can't be started yet are automatically held back until their prerequisites are done.

### Story 29.1: DependsOn Field, DependencyResolver, and YAML Persistence

As a ThreeDoors user,
I want to declare that a task depends on other tasks,
So that the system understands which tasks must be completed before others can begin.

**Status:** Not Started
**Priority:** P1
**Depends On:** None (foundational story)
**FRs:** FR110, FR115

**Acceptance Criteria:**

**AC 29.1.1 — DependsOn field persists through save/load cycle**
**Given** a task with DependsOn set to `["task-id-1", "task-id-2"]`
**When** the task pool is saved to YAML and reloaded
**Then** the task's DependsOn values are preserved exactly
**And** a task with nil/empty DependsOn serializes with no `depends_on` key in YAML

**AC 29.1.2 — HasUnmetDependencies returns true for incomplete deps**
**Given** task A depends on task B (status: todo) and task C (status: complete)
**When** HasUnmetDependencies is called for task A
**Then** it returns true (task B is not complete)

**AC 29.1.3 — HasUnmetDependencies returns false when all deps complete**
**Given** task A depends on task B (status: complete) and task C (status: complete)
**When** HasUnmetDependencies is called for task A
**Then** it returns false

**AC 29.1.4 — Orphaned dependency treated as unmet (pessimistic)**
**Given** task A depends on task ID "nonexistent-id" not in the pool
**When** HasUnmetDependencies is called for task A
**Then** it returns true (pessimistic handling)

**AC 29.1.5 — GetBlockingDependencies returns correct blockers**
**Given** task A depends on tasks B (todo), C (complete), D (in-progress)
**When** GetBlockingDependencies is called for task A
**Then** it returns [B, D] (only incomplete dependencies)
**And** orphaned IDs return a placeholder with text "[deleted task]"

**AC 29.1.6 — Circular dependency detected (direct cycle)**
**Given** task A depends on task B
**When** adding a dependency from task B to task A
**Then** WouldCreateCycle returns true

**AC 29.1.7 — Circular dependency detected (transitive cycle)**
**Given** task A depends on B, B depends on C
**When** adding a dependency from task C to task A
**Then** WouldCreateCycle returns true

**AC 29.1.8 — Self-dependency detected**
**Given** task A exists
**When** adding a dependency from task A to task A
**Then** WouldCreateCycle returns true

**AC 29.1.9 — No false positive cycle detection**
**Given** task A depends on B, task C depends on D (independent chains)
**When** adding a dependency from task A to task C
**Then** WouldCreateCycle returns false

**AC 29.1.10 — Tasks with no dependencies unaffected**
**Given** task A has empty DependsOn
**When** HasUnmetDependencies is called for task A
**Then** it returns false
**And** GetBlockingDependencies returns nil

**Tasks:**

1. Add `DependsOn []string` field to Task struct in `internal/core/task.go` with YAML tag `depends_on,omitempty`
2. Create `internal/core/dependency.go` with `HasUnmetDependencies`, `GetBlockingDependencies`, `WouldCreateCycle`, `GetNewlyUnblockedTasks` functions
3. Write comprehensive unit tests in `internal/core/dependency_test.go`: table-driven tests for all resolver functions
4. Write YAML round-trip test: DependsOn serialization/deserialization
5. Write edge case tests: empty deps, self-dependency, orphaned IDs, large fan-out (10+ deps)
6. Run `make fmt && make lint && make test` — all must pass

---

### Story 29.2: Door Selection Filter and Auto-Unblock on Completion

As a ThreeDoors user,
I want tasks with incomplete dependencies automatically hidden from doors and tasks to become available when their dependencies complete,
So that every door I see is genuinely actionable.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 29.1 (DependencyResolver must exist)
**FRs:** FR111, FR113

**Acceptance Criteria:**

**AC 29.2.1 — Dependency-blocked tasks excluded from doors**
**Given** task A (status: todo) depends on task B (status: todo)
**When** GetAvailableForDoors() is called
**Then** task A is NOT in the returned list
**And** task B IS in the returned list (it has no dependencies)

**AC 29.2.2 — Task with all deps complete appears in doors**
**Given** task A (status: todo) depends on task B (status: complete)
**When** GetAvailableForDoors() is called
**Then** task A IS in the returned list

**AC 29.2.3 — Auto-unblock on dependency completion**
**Given** task A depends on task B (the only dependency)
**And** task B is marked complete
**When** the auto-unblock check runs
**Then** GetNewlyUnblockedTasks returns [task A]
**And** a DependencyUnblockedMsg is emitted
**And** the doors view refreshes to potentially include task A

**AC 29.2.4 — Cascading unblock**
**Given** task A depends on task B, task B depends on task C
**And** task C is marked complete
**When** the auto-unblock check runs
**Then** task B becomes available for doors (its dependency C is complete)
**And** task A remains blocked (task B is not yet complete)
**When** task B is later marked complete
**Then** task A becomes available for doors

**AC 29.2.5 — Auto-unblock only fires for newly unblocked tasks**
**Given** task A depends on tasks B and C
**And** task B is marked complete (but C is still todo)
**When** the auto-unblock check runs
**Then** GetNewlyUnblockedTasks returns empty (A still has unmet dep C)

**AC 29.2.6 — Tasks without dependencies unaffected by filter**
**Given** tasks with empty DependsOn and status todo
**When** GetAvailableForDoors() is called
**Then** those tasks are included as before (no regression)

**Tasks:**

1. Modify `GetAvailableForDoors()` in `internal/core/task_pool.go` to call `HasUnmetDependencies()` and skip tasks with unmet deps
2. Add `DependencyUnblockedMsg` type in `internal/tui/messages.go` (or appropriate location)
3. Add auto-unblock check in MainModel when a task transitions to StatusComplete
4. Handle `DependencyUnblockedMsg` in MainModel.Update: log event, refresh doors
5. Write unit tests: GetAvailableForDoors with dependency-blocked tasks
6. Write integration test: complete dep -> verify dependent reappears in available list
7. Write cascade test: A->B->C chain unblocking
8. Write regression test: tasks without deps still appear (no behavioral change)
9. Run `make fmt && make lint && make test` — all must pass

---

### Story 29.3: TUI Blocked-By Indicator and Dependency Management

As a ThreeDoors user,
I want to see which tasks are blocking my work and manage dependencies from the detail view,
So that I understand task relationships and can organize my work effectively.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 29.1 (DependencyResolver must exist)
**Can parallel with:** Story 29.2
**FRs:** FR112, FR114, FR115

**Acceptance Criteria:**

**AC 29.3.1 — Blocked-by indicator on doors**
**Given** task A has incomplete dependencies, with task B as the first blocker
**When** task A appears in a context where blocked-by is rendered (e.g., detail view, or a future enhanced doors view)
**Then** a "Blocked by: [task B text truncated to 40 chars]" line is displayed
**And** if there are additional blockers, "+N more" is appended

**AC 29.3.2 — Blocked-by indicator with single blocker**
**Given** task A depends only on task B (incomplete)
**When** the blocked-by indicator renders
**Then** it shows "Blocked by: [B's text]" with no "+N more" suffix

**AC 29.3.3 — Blocked-by indicator with multiple blockers**
**Given** task A depends on tasks B, C, D (all incomplete)
**When** the blocked-by indicator renders
**Then** it shows "Blocked by: [B's text] (+2 more)"

**AC 29.3.4 — Dependency list in detail view**
**Given** the user is viewing task A in the detail view
**And** task A has dependencies [B (complete), C (todo)]
**When** the detail view renders
**Then** dependencies are shown below notes with checkbox indicators:
- `[x] B's text` (complete)
- `[ ] C's text` (incomplete, marked as blocking)

**AC 29.3.5 — Add dependency via + key**
**Given** the user is viewing a task in the detail view
**When** the user presses `+`
**Then** a task search/picker opens showing available tasks
**And** selecting a task adds it to DependsOn
**And** the dependency list updates immediately

**AC 29.3.6 — Circular dependency rejected with error**
**Given** the user is adding a dependency that would create a cycle
**When** the add operation is attempted
**Then** an error message "Cannot add dependency: would create a circular chain" is displayed
**And** the dependency is not added

**AC 29.3.7 — Remove dependency via - key**
**Given** the user is viewing a task with dependencies in the detail view
**And** a dependency is selected/highlighted in the list
**When** the user presses `-`
**Then** the selected dependency is removed from DependsOn
**And** the dependency list updates immediately

**AC 29.3.8 — Empty dependency list**
**Given** a task with no dependencies
**When** the detail view renders
**Then** no dependency section is shown (or a minimal "No dependencies" note)

**Tasks:**

1. Add blocked-by indicator rendering in `internal/tui/doors_view.go` using `GetBlockingDependencies()`
2. Add dependency list section to `internal/tui/task_detail_view.go` — checkbox-style display
3. Implement `+` key handler in detail view: open task search/picker, validate cycle, add dep
4. Implement `-` key handler in detail view: remove selected dependency
5. Implement cycle rejection with user-visible error message
6. Add dependency cursor/selection state to TaskDetailView for navigating the dep list
7. Handle TaskUpdatedMsg for dependency changes (save pool)
8. Write unit tests: blocked-by text formatting, truncation, "+N more" logic
9. Write golden file tests: blocked-by indicator rendering at standard widths
10. Write golden file tests: detail view dependency list rendering
11. Run `make fmt && make lint && make test` — all must pass

---

### Story 29.4: Session Metrics Logging for Dependency Events

As a ThreeDoors developer/analyst,
I want dependency events logged to session metrics,
So that usage patterns can be analyzed and the feature can be improved over time.

**Status:** Not Started
**Priority:** P1
**Depends On:** Story 29.2 (auto-unblock must exist), Story 29.3 (add/remove dep must exist)
**FRs:** FR113 (event logging aspect)

**Acceptance Criteria:**

**AC 29.4.1 — Dependency added events logged**
**Given** a user adds a dependency via the + key
**When** the dependency is successfully added
**Then** a `dependency_added` event is appended to the JSONL session metrics log
**And** the event includes: task_id, dependency_id, timestamp

**AC 29.4.2 — Dependency removed events logged**
**Given** a user removes a dependency via the - key
**When** the dependency is successfully removed
**Then** a `dependency_removed` event is appended to the session metrics log
**And** the event includes: task_id, dependency_id, timestamp

**AC 29.4.3 — Auto-unblock events logged**
**Given** a task's dependencies are all complete after another task finishes
**When** the auto-unblock logic detects the newly unblocked task
**Then** a `dependency_unblocked` event is appended to the session metrics log
**And** the event includes: task_id, completed_dependency_id, timestamp

**AC 29.4.4 — Cycle rejection events logged**
**Given** a user attempts to add a dependency that would create a cycle
**When** the cycle is detected and rejected
**Then** a `dependency_cycle_rejected` event is appended to the session metrics log
**And** the event includes: task_id, attempted_dependency_id, timestamp

**AC 29.4.5 — Metrics format consistent with existing events**
**Given** the existing session metrics use JSONL format at `~/.threedoors/sessions.jsonl`
**When** dependency events are logged
**Then** they follow the same JSONL structure and field conventions as existing event types

**Tasks:**

1. Define dependency event types in session metrics module (dependency_added, dependency_removed, dependency_unblocked, dependency_cycle_rejected)
2. Log dependency_added event in TaskDetailView's + key handler
3. Log dependency_removed event in TaskDetailView's - key handler
4. Log dependency_unblocked event in MainModel's DependencyUnblockedMsg handler
5. Log dependency_cycle_rejected event when WouldCreateCycle returns true
6. Write unit tests: verify event structure, field presence, JSONL format
7. Run `make fmt && make lint && make test` — all must pass

---

## Story Dependency Graph

```
29.1 (DependsOn Field + DependencyResolver)
  |
  +-- 29.2 (Door Filter + Auto-Unblock)
  |     |
  |     +-- 29.4 (Session Metrics) <-- also depends on 29.3
  |
  +-- 29.3 (TUI Indicators + Dependency Management)
        |
        +-- 29.4 (Session Metrics)
```

Stories 29.2 and 29.3 can be implemented in parallel after 29.1 completes.
Story 29.4 depends on both 29.2 and 29.3 but is lightweight.

## Estimated Scope

| Story | Files Changed | Files Created | Complexity |
|-------|--------------|---------------|------------|
| 29.1 | 1 (task.go) | 2 (dependency.go, dependency_test.go) | Medium |
| 29.2 | 2 (task_pool.go, main_model.go) | 0 | Medium |
| 29.3 | 2 (doors_view.go, task_detail_view.go) | 0 | Medium-High |
| 29.4 | 3 (main_model.go, task_detail_view.go, session metrics) | 0 | Low |
