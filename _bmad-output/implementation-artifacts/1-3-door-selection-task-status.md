# Story 1.3: Door Selection & Task Status Management

Status: ready-for-dev

## Story

As a user,
I want to select a door and update the task's status,
so that I can take action on tasks and track my progress.

## Acceptance Criteria

1. Pressing A/Left Arrow, W/Up Arrow, or D/Right Arrow selects the corresponding door (left, center, right)
2. Selected task is highlighted/indicated visually
3. **Door Opening & Expanded Detail View:**
   - When door is selected (or Enter is pressed), door presents optional animation as if opening
   - Selected door shifts to left position and expands to fill the screen
   - Task detail view displays:
     - Task text (full, not truncated)
     - Any existing task metadata/details (status, notes, timestamps, etc.)
     - Status action menu with all available options
   - **Esc** key closes the door and returns to three doors view
4. Status action menu shows available options:
   - **C**: Mark as Complete
   - **B**: Mark as Blocked
   - **I**: Mark as In Progress
   - **E**: Expand (break into more tasks)
   - **F**: Fork (clone/split task)
   - **P**: Procrastinate (defer task)
   - **R**: Flag for Rework
   - **M**: Log Mood/Context (also available from door view without selection)
   - **Esc**: Close door and return to previous screen
5. Pressing any status key (within expanded detail view) applies that status to the selected task
6. Pressing **M** from door view (no task selection needed) opens mood capture dialog:
   - Multiple choice options: "Focused", "Tired", "Stressed", "Energized", "Distracted", "Calm", "Other"
   - If "Other" selected, prompt for custom text input (word or phrase)
   - Mood entries are timestamped and recorded in session metrics
   - Returns to door view immediately after capture
7. Completed tasks are removed from available task pool (in-memory) and appended to `~/.threedoors/completed.txt` with timestamp
8. Blocked/deferred/rework tasks remain in the pool but are tagged with status
9. New set of three doors is displayed automatically after any status change (door closes and returns to three doors view)
10. Session completion count increments for completed tasks (e.g., "Completed this session: 3")
11. "Progress over perfection" message shown after completing a task
12. Door selection is tracked: which door position (left/center/right) was selected
13. Task bypass is tracked: doors shown but not selected before refresh
14. Mood entries are tracked with timestamps for correlation with task selection patterns

## Tasks / Subtasks

- [ ] Task 1: Restore source code and build infrastructure (AC: prerequisite)
  - [ ] 1.1: Restore go.mod and go.sum from commit 3d2486b (run `git checkout 3d2486b -- go.mod go.sum` or restore + `go mod tidy`)
  - [ ] 1.2: Restore cmd/threedoors/main.go and main_test.go
  - [ ] 1.3: Restore internal/tasks/*.go files (task.go, file_manager.go, file_manager_test.go, session_tracker.go, metrics_writer.go)
  - [ ] 1.4: Create Makefile with targets: build, run, clean, fmt, lint, test (per architecture spec)
  - [ ] 1.5: Add `github.com/charmbracelet/bubbles` dependency (needed for textarea in notes/mood input)
  - [ ] 1.6: Add `gopkg.in/yaml.v3` dependency (needed for YAML task persistence)
  - [ ] 1.7: Run `go mod tidy` and verify project compiles
  - [ ] 1.8: Run existing tests to verify baseline
  - [ ] 1.9: Fix deprecated `rand.Seed` - upgrade to `math/rand/v2` or use global rand (Go 1.20+ auto-seeds)

- [ ] Task 2: Upgrade Task model from simple text to full lifecycle (AC: 3, 5, 7, 8)
  - [ ] 2.1: Extend Task struct with id (UUID), status (TaskStatus enum), notes, blocker, timestamps
  - [ ] 2.2: Implement TaskStatus type with constants (todo, blocked, in-progress, in-review, complete)
  - [ ] 2.3: Implement NewTask(text) constructor with UUID generation
  - [ ] 2.4: Implement status transition validation (IsValidTransition, UpdateStatus)
  - [ ] 2.5: Implement AddNote, SetBlocker methods
  - [ ] 2.6: Update LoadTasks to parse plain text into Task structs with UUIDs and default status "todo"
  - [ ] 2.7: Implement SaveTasks for persisting task state back to file
  - [ ] 2.8: Implement AppendCompleted for completed.txt logging
  - [ ] 2.9: Write unit tests for Task model and status transitions

- [ ] Task 3: Refactor main.go into proper architecture (AC: all)
  - [ ] 3.1: Create internal/tui/ package directory
  - [ ] 3.2: Create internal/tui/messages.go with all message types: SelectDoorMsg, ReturnToDoorsMsg, TaskUpdatedMsg, RefreshDoorsMsg, MoodCaptureMsg
  - [ ] 3.3: Create MainModel as root Bubbletea model with view routing ("doors" | "detail" | "mood")
  - [ ] 3.4: Extract DoorsView component from current main.go View/Update logic into internal/tui/doors_view.go
  - [ ] 3.5: Remove panic in getThreeRandomDoors - handle < 3 tasks gracefully (show 1-2 doors or "no tasks" message)
  - [ ] 3.6: Update cmd/threedoors/main.go to be minimal - just creates MainModel and runs
  - [ ] 3.7: Write tests for view routing and message handling

- [ ] Task 4: Implement TaskDetailView (AC: 3, 4, 5)
  - [ ] 4.1: Create TaskDetailView component in internal/tui/detail_view.go
  - [ ] 4.2: Render full task text, status, notes, timestamps
  - [ ] 4.3: Display status action menu (C/B/I/E/F/P/R/M/Esc)
  - [ ] 4.4: Handle key presses for status changes
  - [ ] 4.5: Handle Esc to return to doors view (send ReturnToDoorsMsg)
  - [ ] 4.6: On status change, persist to file, remove completed tasks from pool, refresh doors
  - [ ] 4.7: Optional: door opening animation effect

- [ ] Task 5: Implement completion flow (AC: 7, 9, 10, 11)
  - [ ] 5.1: When C pressed, mark task complete, remove from pool
  - [ ] 5.2: Append completed task to ~/.threedoors/completed.txt with format `[YYYY-MM-DD HH:MM:SS] task_id | task_text`
  - [ ] 5.3: Increment session completion counter displayed in footer
  - [ ] 5.4: Show "Progress over perfection" message for 2-3 seconds after completion (use tea.Tick)
  - [ ] 5.5: Auto-refresh doors after status change
  - [ ] 5.6: Handle edge case: what if completing last available task? Show "All tasks done!" celebration

- [ ] Task 6: Implement Mood Capture (AC: 6, 14)
  - [ ] 6.1: Create MoodCaptureView component
  - [ ] 6.2: Handle M key from doors view (no selection needed)
  - [ ] 6.3: Display mood options: Focused, Tired, Stressed, Energized, Distracted, Calm, Other
  - [ ] 6.4: If "Other", show text input for custom mood
  - [ ] 6.5: Record mood with timestamp in SessionTracker
  - [ ] 6.6: Return to doors view after capture

- [ ] Task 7: Implement Expand/Fork/Procrastinate/Rework (AC: 4, 5, 8)
  - [ ] 7.1: E (Expand): prompt for sub-tasks text, create new tasks, original marked complete
  - [ ] 7.2: F (Fork): clone task, append " (fork)" to copy
  - [ ] 7.3: P (Procrastinate): mark with "deferred" status, keep in pool
  - [ ] 7.4: R (Rework): flag task for rework status

- [ ] Task 8: Integrate session tracking (AC: 12, 13, 14)
  - [ ] 8.1: Update SessionTracker with enhanced methods: RecordDoorSelection(position, taskText), RecordMood(mood, customText)
  - [ ] 8.2: Track door position selections in DoorsView
  - [ ] 8.3: Track task bypasses on refresh (doors shown but not selected)
  - [ ] 8.4: Integrate SessionTracker into MainModel, pass to views
  - [ ] 8.5: Persist session on app exit

- [ ] Task 9: Blocked task handling (AC: 8)
  - [ ] 9.1: When B pressed, show text input for optional blocker note (use bubbles/textinput)
  - [ ] 9.2: Store blocker text on task
  - [ ] 9.3: Display blocker info in task detail view with visual indicator (red text)

- [ ] Task 10: Data format migration to YAML (AC: 3, 7, 8)
  - [ ] 10.1: Migrate task storage from tasks.txt to tasks.yaml format per architecture spec
  - [ ] 10.2: Implement backward compatibility: if tasks.txt exists but tasks.yaml doesn't, migrate automatically
  - [ ] 10.3: Use atomic write pattern for SaveTasks (write .tmp, fsync, rename)
  - [ ] 10.4: Add gopkg.in/yaml.v3 for YAML marshaling/unmarshaling

- [ ] Task 11: Integration and edge case tests (AC: all)
  - [ ] 11.1: Integration test: select door -> view detail -> change status -> verify file persistence
  - [ ] 11.2: Test completed.txt format validation
  - [ ] 11.3: Edge case: fewer than 3 tasks available (1-2 tasks, 0 tasks)
  - [ ] 11.4: Edge case: all tasks completed
  - [ ] 11.5: Test TaskDetailView key handling
  - [ ] 11.6: Test mood capture flow

## Dev Notes

### CRITICAL: Source Code Recovery Required

The source code from Stories 1.1 and 1.2 was deleted in commit `af12e90`. All Go files must be restored from commit `3d2486b` before any work begins. The files to restore are:

```
cmd/threedoors/main.go
cmd/threedoors/main_test.go
go.mod (with go.sum)
internal/tasks/task.go
internal/tasks/file_manager.go
internal/tasks/file_manager_test.go
internal/tasks/session_tracker.go
internal/tasks/metrics_writer.go
```

After restoring, run `go mod tidy` and `go test ./...` to verify baseline.

### Architecture Compliance

This story requires a significant refactor from the current monolithic main.go into the proper two-layer architecture defined in the architecture docs.

**Required architecture:**
- **TUI Layer** (`internal/tui/`): MainModel, DoorsView, TaskDetailView, MoodCaptureView
- **Domain Layer** (`internal/tasks/`): Enhanced Task model, FileManager, StatusManager, TaskPool, DoorSelector

**Key patterns to follow:**
- Model-View-Update (MVU) pattern via Bubbletea (Elm Architecture)
- Constructor injection for dependencies
- Message-based communication between views (SelectDoorMsg, ReturnToDoorsMsg, etc.)
- View routing in MainModel based on currentView state

[Source: docs/architecture/components.md, docs/architecture/high-level-architecture.md]

### Task Model Upgrade

The current Task struct is minimal (`Text string`). It must be upgraded to the full model:

```go
type TaskStatus string

const (
    StatusTodo       TaskStatus = "todo"
    StatusBlocked    TaskStatus = "blocked"
    StatusInProgress TaskStatus = "in-progress"
    StatusInReview   TaskStatus = "in-review"
    StatusComplete   TaskStatus = "complete"
)

type Task struct {
    ID          string      `yaml:"id"`
    Text        string      `yaml:"text"`
    Status      TaskStatus  `yaml:"status"`
    Notes       []TaskNote  `yaml:"notes"`
    Blocker     string      `yaml:"blocker"`
    CreatedAt   time.Time   `yaml:"created_at"`
    UpdatedAt   time.Time   `yaml:"updated_at"`
    CompletedAt *time.Time  `yaml:"completed_at"`
}
```

**Status transition rules** (state machine):
- todo -> in-progress, blocked, complete
- blocked -> todo, in-progress, complete
- in-progress -> blocked, in-review, complete
- in-review -> in-progress, complete
- Any state -> complete (force complete)

[Source: docs/architecture/data-models.md]

### File Persistence - MIGRATE TO YAML

**Current:** Plain text `tasks.txt` with one task per line.
**Target:** Migrate to `tasks.yaml` format per architecture spec. This is required because the Task model now has status, notes, blocker, timestamps that cannot be stored in plain text.

**Migration strategy:**
1. On startup, check if `tasks.yaml` exists
2. If not but `tasks.txt` exists, auto-migrate: read text lines, create Task structs with UUIDs and "todo" status, write to tasks.yaml
3. After successful migration, rename tasks.txt to tasks.txt.bak

**FileManager needs:**
- `LoadTasks() (*TaskPool, error)` - reads tasks.yaml, returns populated TaskPool
- `SaveTasks(pool *TaskPool) error` - persists TaskPool to tasks.yaml (atomic write)
- `AppendCompleted(task *Task) error` - appends `[YYYY-MM-DD HH:MM:SS] task_id | task_text` to completed.txt
- `InitializeFiles() error` - create ~/.threedoors/ dir and sample tasks.yaml if missing

**Atomic write pattern required for SaveTasks:**
1. Marshal TaskPool to YAML
2. Write to `tasks.yaml.tmp`
3. `fsync()` to flush to disk
4. Atomic rename `tasks.yaml.tmp` -> `tasks.yaml`

**Add dependency:** `gopkg.in/yaml.v3` v3.0.1

[Source: docs/architecture/data-storage-schema.md, docs/architecture/coding-standards.md]

### Coding Standards

- Go 1.25.4 strict
- `gofumpt` formatting required
- `golangci-lint` must pass
- NO `fmt.Println` in TUI code - all rendering through Bubbletea `View()` method
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- No panics in production code (remove the panic in current getThreeRandomDoors)
- Constructor functions: `NewXxx()` pattern
- PascalCase for exports, camelCase for private

[Source: docs/architecture/coding-standards.md]

### Testing Standards

- **Coverage goals:** 70%+ for internal/tasks, 20%+ for internal/tui, 50%+ overall
- **Test patterns:** Table-driven tests preferred
- Use `t.TempDir()` for file operations (never touch real ~/.threedoors)
- Test status transition validation exhaustively
- Test Task model creation and methods
- Test FileManager read/write/complete operations

[Source: docs/architecture/test-strategy-and-standards.md]

### TEA Test Architecture Specification

**Test file mapping:**
| Source File | Test File | Priority |
|---|---|---|
| internal/tasks/task.go | internal/tasks/task_test.go | HIGH |
| internal/tasks/task_status.go | internal/tasks/task_status_test.go | HIGH |
| internal/tasks/file_manager.go | internal/tasks/file_manager_test.go | HIGH |
| internal/tasks/task_pool.go | internal/tasks/task_pool_test.go | HIGH |
| internal/tasks/door_selector.go | internal/tasks/door_selector_test.go | MEDIUM |
| internal/tasks/session_tracker.go | internal/tasks/session_tracker_test.go | MEDIUM |
| internal/tasks/metrics_writer.go | internal/tasks/metrics_writer_test.go | MEDIUM |
| internal/tui/main_model.go | internal/tui/main_model_test.go | LOW |
| internal/tui/doors_view.go | internal/tui/doors_view_test.go | LOW |
| internal/tui/detail_view.go | internal/tui/detail_view_test.go | LOW |

**Interface definitions for testability:**
```go
// TaskStore abstracts file persistence for testing
type TaskStore interface {
    LoadTasks() ([]*Task, error)
    SaveTasks(tasks []*Task) error
    AppendCompleted(task *Task) error
}

// MetricsRecorder abstracts session tracking for testing
type MetricsRecorder interface {
    RecordDoorViewed()
    RecordDoorSelection(position int, taskText string)
    RecordRefresh(doorTasks []string)
    RecordDetailView()
    RecordTaskCompleted(taskText string)
    RecordStatusChange(status string, taskText string)
    RecordMood(mood string, customText string)
    Finalize() *SessionMetrics
}
```

**Status transition test matrix (table-driven):**
```go
// All valid transitions
{"todo", "in-progress", true},
{"todo", "blocked", true},
{"todo", "complete", true},
{"blocked", "todo", true},
{"blocked", "in-progress", true},
{"blocked", "complete", true},
{"in-progress", "blocked", true},
{"in-progress", "in-review", true},
{"in-progress", "complete", true},
{"in-review", "in-progress", true},
{"in-review", "complete", true},
// Invalid transitions
{"todo", "in-review", false},
{"blocked", "in-review", false},
{"in-review", "todo", false},
{"in-review", "blocked", false},
{"complete", "todo", false},
{"complete", "in-progress", false},
{"complete", "blocked", false},
{"complete", "in-review", false},
// Same status (no-op, should be valid)
{"todo", "todo", true},
```

**YAML roundtrip test:**
```
Given a TaskPool with 3 tasks (various statuses, notes, blockers)
When SaveTasks is called then LoadTasks is called
Then the loaded tasks match the original tasks exactly (field-by-field comparison)
```

**Acceptance test scenarios (BDD format):**

**AC1 - Door Selection:**
```
Given three doors are displayed
When user presses 'a' (or left arrow)
Then the left door is highlighted with selected style
And no status change occurs until Enter is pressed
```

**AC3 - Detail View:**
```
Given a door is selected
When user presses Enter (or the selection key activates detail)
Then the view transitions to TaskDetailView
And full task text, status, timestamps are displayed
And status action menu is visible
When user presses Esc
Then view returns to three doors
```

**AC5 - Status Change:**
```
Given TaskDetailView is showing a task with status "todo"
When user presses 'C' (complete)
Then task status changes to "complete"
And task is removed from pool
And completed.txt is appended
And view returns to doors with new selection
```

**AC6 - Mood Capture:**
```
Given three doors are displayed (no selection needed)
When user presses 'M'
Then mood selection dialog appears with 7 options
When user presses '1' (Focused)
Then mood is recorded with timestamp
And view returns to doors
```

**AC7 - Completed Task Persistence:**
```
Given a task "Write tests" with ID "abc-123"
When task is marked complete
Then completed.txt contains line matching: [YYYY-MM-DD HH:MM:SS] abc-123 | Write tests
And task is no longer in tasks.yaml
```

**Edge case tests:**
```
Given only 2 tasks exist in tasks.yaml
When app starts
Then 2 doors are displayed (not 3)
And no panic occurs

Given only 1 task exists
When user completes it
Then "All tasks done!" message is displayed
And doors view shows empty state

Given tasks.txt exists but tasks.yaml doesn't
When app starts
Then tasks are migrated from txt to yaml
And tasks.txt is renamed to tasks.txt.bak
```

**Test fixtures (create in testdata/ or use t.TempDir):**
- `testdata/sample_tasks.yaml` - 5 tasks in various statuses
- `testdata/empty_tasks.yaml` - valid YAML with empty tasks array
- `testdata/single_task.yaml` - 1 task for edge case testing
- `testdata/tasks_with_notes.yaml` - tasks with notes and blockers

**TUI testing approach:**
- Use manual message-based testing (create model, send tea.KeyMsg, assert state)
- Do NOT use teatest package (adds complexity, low ROI for tech demo)
- Test state transitions, not rendering output

### Tech Stack

| Component | Technology | Version |
|---|---|---|
| Language | Go | 1.25.4 |
| TUI Framework | Bubbletea | 1.2.4+ (currently 1.3.10) |
| Styling | Lipgloss | 1.0.0+ (currently 1.1.0) |
| Components | Bubbles | 0.20.0 |
| UUID | github.com/google/uuid | 1.6.0 |
| Build | Make | System |

[Source: docs/architecture/tech-stack.md]

### UX Styling Specifications

**Status colors (Lipgloss):**
- TODO: White/default (`lipgloss.Color("252")`)
- IN-PROGRESS: Yellow (`lipgloss.Color("214")`)
- BLOCKED: Red (`lipgloss.Color("196")`)
- IN-REVIEW: Blue (`lipgloss.Color("39")`)
- COMPLETE: Green (`lipgloss.Color("82")`)

**Task Detail View layout:**
- Full-width bordered box (lipgloss.RoundedBorder)
- Header: "TASK DETAILS" with accent color
- Status shown with color + emoji indicator
- Notes listed chronologically with timestamps
- Action menu at bottom with key hints

**"Progress over perfection" message:**
- Shown for 2-3 seconds after task completion (use `tea.Tick` command)
- Displayed in green, centered below the doors
- Format: "Progress over perfection. Just pick one and start."

**Mood capture dialog:**
- Centered overlay/popup style
- Numbered list: 1. Focused, 2. Tired, 3. Stressed, 4. Energized, 5. Distracted, 6. Calm, 7. Other
- Press number key to select
- If "Other", show textinput field
- Brief confirmation message on capture

**Blocker note input:**
- Uses `bubbles/textinput` component
- Prompt: "Blocker reason (optional, Enter to skip):"
- Max 200 characters
- Enter submits, Esc cancels

### Key Design Decisions

- **Door opening animation is optional** but provides delightful visual feedback
- **Expanded detail view** shifts door left and fills screen for focused interaction
- **Context-aware return** (Esc) maintains navigation state (critical for search integration in Story 1.3a)
- All status changes tracked for future pattern analysis
- Door position preferences (left vs center vs right) captured for learning
- Tasks that are expanded or forked create new task entries in tasks.txt
- Blocked tasks should prompt for optional blocker note
- **Mood tracking is casual and low-friction** - accessible anytime via 'M' key without needing task selection
- Multiple choice moods keep capture quick; custom text option allows nuanced expression

### Previous Story Intelligence

**From Story 1.2 (completed):**
- Task model is currently `type Task struct { Text string }` - needs full upgrade
- FileManager exists with LoadTasks(), GetConfigDirPath(), SetHomeDir() for testing
- Main.go has working door display with AWSD/arrow key selection
- Tests exist for FileManager and basic quit functionality
- Warning: session_tracker.go exists (from Story 1.1 QA failure) - it has basic SessionMetrics but needs enhancement for door position tracking, mood, and bypass tracking
- metrics_writer.go exists with basic AppendSession functionality

**From Story 1.1 QA notes:**
- session_tracker.go was untracked and caused QA failure
- DO integrate it properly in this story rather than ignoring it

### Git Intelligence

Recent commits show documentation-only changes. Source code was removed in commit af12e90.
Previous code patterns:
- Standard Go module layout
- Bubbletea MVU pattern in main.go
- Lipgloss for styling with dynamic terminal width
- rand.Seed + rand.Perm for random door selection (should upgrade to math/rand/v2)

### Project Structure Notes

**Target source tree (from architecture):**
```
cmd/threedoors/
  main.go                 # Entry point, minimal - creates MainModel and runs

internal/tui/
  main_model.go           # Root Bubbletea model, view routing
  doors_view.go           # Three doors display
  detail_view.go          # Task detail + status actions
  mood_view.go            # Mood capture dialog
  messages.go             # Shared message types

internal/tasks/
  task.go                 # Task model with full lifecycle
  task_status.go          # TaskStatus enum and transitions
  file_manager.go         # File I/O (load/save/complete)
  task_pool.go            # In-memory task collection
  door_selector.go        # Random selection algorithm
  session_tracker.go      # Session metrics tracking
  metrics_writer.go       # JSONL persistence
```

[Source: docs/architecture/source-tree.md]

### References

- [Source: docs/architecture/components.md] - Component definitions and interactions
- [Source: docs/architecture/core-workflows.md] - Sequence diagrams for all workflows
- [Source: docs/architecture/data-models.md] - Task, TaskStatus, TaskPool, DoorSelection models
- [Source: docs/architecture/data-storage-schema.md] - YAML and completed.txt formats
- [Source: docs/architecture/coding-standards.md] - Go coding standards and patterns
- [Source: docs/architecture/test-strategy-and-standards.md] - Testing approach and coverage goals
- [Source: docs/architecture/tech-stack.md] - Technology versions
- [Source: docs/architecture/source-tree.md] - Directory structure
- [Source: docs/prd/epic-details.md#Story 1.3] - Full acceptance criteria
- [Source: docs/stories/1.2.story.md] - Previous story completion notes

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
