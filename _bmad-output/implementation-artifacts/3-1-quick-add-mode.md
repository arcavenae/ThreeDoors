# Story 3.1: Quick Add Mode

Status: ready-for-dev

## Story

As a user,
I want to quickly add tasks with minimal interaction,
So that I can capture ideas without breaking flow.

## Acceptance Criteria

1. **Given** the user is in any view, **When** the user types `:add <task text>` in the command palette, **Then** a new task is created with the given text **And** the task status is set to `todo` **And** the task is immediately available in the Three Doors pool **And** the task is persisted to the active backend **And** a brief confirmation flash is shown ("Task added") that auto-clears after the standard flash duration

2. **Given** the user types `:add` without text, **When** Enter is pressed, **Then** an inline text input prompts for task text with placeholder "Enter task text..." **And** text is limited to 500 characters (per data model spec) **And** pressing Enter with non-empty text (after trim) creates the task and returns to the previous view (doors or search) **And** pressing Enter with empty/whitespace-only text shows "Task text cannot be empty" flash **And** pressing Esc cancels the add operation and returns to previous view

3. **Given** the user is in the doors view, **When** the user presses `:` (colon), **Then** the search/command palette opens in command mode (same as pressing `/` then typing `:`) — this ensures "from any view" works without requiring the user to know about `/` first

4. **Given** a task is added (via either `:add <text>` or inline prompt), **When** the task pool was previously empty, **Then** the new task appears in the doors view when returning (doors should refresh to include it)

## Tasks / Subtasks

- [ ] Task 1: Implement `:add` with inline prompt mode (AC: #2)
  - [ ] 1.1: Add `ViewAddTask` view mode enum value to `internal/tui/main_model.go`
  - [ ] 1.2: Create `AddTaskView` component in `internal/tui/add_task_view.go` with textinput (500 char limit, placeholder "Enter task text...")
  - [ ] 1.3: Handle `AddTaskPromptMsg` in `MainModel.Update()` to enter inline add mode, preserving `previousView` for return
  - [ ] 1.4: Handle Enter to create task: validate non-empty after trim, emit `TaskAddedMsg` on success, show "Task text cannot be empty" flash on empty
  - [ ] 1.5: Handle Esc to cancel: return to previous view (doors or restored search state)
  - [ ] 1.6: Wire up the `:add` (no args) command in `SearchView.executeCommand()` to emit `AddTaskPromptMsg`
- [ ] Task 2: Ensure `:add <text>` with args works correctly from any view (AC: #1)
  - [ ] 2.1: Verify existing `:add <text>` handler in `search_view.go` correctly creates task via `tasks.NewTask(args)` and emits `TaskAddedMsg`
  - [ ] 2.2: Verify `MainModel.Update()` handler for `TaskAddedMsg` adds to pool, persists via `saveTasks()`, and shows flash "Task added"
  - [ ] 2.3: Ensure the flash message auto-clears via existing `ClearFlashCmd()` pattern
  - [ ] 2.4: Verify task status is set to `todo` by `tasks.NewTask()`
- [ ] Task 3: Add `:` shortcut from doors view to enter command mode (AC: #3)
  - [ ] 3.1: In `MainModel.updateDoors()`, handle `:` keypress to open search view in command mode (pre-populate input with `:`)
- [ ] Task 4: Write unit tests (AC: #1, #2, #3, #4)
  - [ ] 4.1: Test `AddTaskView` creation, text input, Enter to submit, Esc to cancel
  - [ ] 4.2: Test `:add <text>` command produces correct `TaskAddedMsg` with task status `todo`
  - [ ] 4.3: Test `:add` without args produces `AddTaskPromptMsg`
  - [ ] 4.4: Test `MainModel` handles `TaskAddedMsg` (pool addition, persistence, flash)
  - [ ] 4.5: Test `MainModel` handles `AddTaskPromptMsg` (view transition, previousView preserved)
  - [ ] 4.6: Test `AddTaskView` cancel returns to correct previous view
  - [ ] 4.7: Test empty/whitespace-only text validation in `AddTaskView`
  - [ ] 4.8: Test `:` from doors view opens command mode
  - [ ] 4.9: Test adding task when pool is empty → doors refresh includes new task
  - [ ] 4.10: Test 500-char limit enforcement on inline text input

## Dev Notes

### Critical Discovery: `:add <text>` Already Implemented

The `:add <text>` command (AC #1) is **already fully implemented** in the current codebase:

- **`internal/tui/search_view.go:96-105`** — `executeCommand()` handles the `"add"` case: when args are present, it creates a `tasks.NewTask(args)` and emits `TaskAddedMsg{Task: newTask}`
- **`internal/tui/main_model.go:122-128`** — `Update()` handles `TaskAddedMsg`: adds to pool via `m.pool.AddTask(msg.Task)`, saves via `m.saveTasks()`, sets flash "Task added", and auto-clears with `ClearFlashCmd()`
- **`internal/tui/messages.go:45-48`** — `TaskAddedMsg` struct already defined

**What's NOT implemented:** When `:add` is typed without args, it currently shows a usage hint `"Usage: :add <task text>"` (search_view.go:98-100) instead of prompting inline for text. This is the primary work for this story.

### What Needs to Be Built

1. **`AddTaskView`** — A new inline text input view similar to `MoodView` that:
   - Shows a text input with placeholder "Enter task text..."
   - On Enter: creates the task and emits `TaskAddedMsg`
   - On Esc: cancels and returns to previous view via `ReturnToDoorsMsg` or restored search state

2. **`AddTaskPromptMsg`** — New message type to trigger the inline add mode

3. **`ViewAddTask` mode** — New view mode in `MainModel` enum

4. **Wire up in `MainModel.Update()`** — Handle `AddTaskPromptMsg` to transition to add task view, and wire the new view's update/view methods

### Architecture & Pattern Compliance

- **MVU Pattern (Bubbletea Elm Architecture):** All state changes via messages. No direct mutation outside `Update()`.
- **View components:** Follow existing pattern (see `MoodView`, `DetailView`): struct with `Update(msg) tea.Cmd`, `View() string`, `SetWidth(int)`
- **Message types:** Define in `internal/tui/messages.go` following existing patterns
- **File naming:** `add_task_view.go` and `add_task_view_test.go` in `internal/tui/`
- **Atomic writes:** Task persistence uses `m.saveTasks()` which delegates to `m.provider.SaveTasks()` — already handles atomic writes
- **TaskProvider interface:** `SaveTasks([]*Task) error` persists to active backend (text file or Apple Notes via fallback)

### File Structure Requirements

Files to create:
- `internal/tui/add_task_view.go` — New AddTaskView component
- `internal/tui/add_task_view_test.go` — Tests for AddTaskView

Files to modify:
- `internal/tui/messages.go` — Add `AddTaskPromptMsg` message type
- `internal/tui/main_model.go` — Add `ViewAddTask` enum, handle `AddTaskPromptMsg`, wire AddTaskView update/view, add `:` shortcut in `updateDoors()`
- `internal/tui/search_view.go` — Change `:add` (no args) to emit `AddTaskPromptMsg` instead of usage hint
- `internal/tui/main_model_test.go` — Add tests for new message handling and `:` shortcut

### Testing Requirements

**Test Files:**
- `internal/tui/add_task_view_test.go` (NEW) — Tests for AddTaskView component
- `internal/tui/search_view_test.go` (MODIFY) — Add tests for `:add` without args emitting `AddTaskPromptMsg`
- `internal/tui/main_model_test.go` (MODIFY) — Add tests for `AddTaskPromptMsg` handling, `:` shortcut, `TaskAddedMsg` with empty pool

**Test Helper Patterns (from existing codebase):**
- Use `makePool(texts...)` and `makeModel(texts...)` helpers from `main_model_test.go`
- Use `keyMsg(s string)` helper for simulating key presses: `keyMsg("enter")`, `keyMsg("esc")`, `keyMsg(":")`, etc.
- For rune input to textinput: `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")}`
- Assert view transitions via `m.viewMode` field
- Assert messages by type-asserting the return of `Update()` tea.Cmd execution

**BDD Test Scenarios:**

| # | Given | When | Then | File |
|---|-------|------|------|------|
| T1 | AddTaskView is active | User types "buy milk" and presses Enter | `TaskAddedMsg{Task}` emitted with text "buy milk", status `todo` | add_task_view_test.go |
| T2 | AddTaskView is active | User presses Esc | Returns to previous view (doors or search) | add_task_view_test.go |
| T3 | AddTaskView is active | User presses Enter with empty input | Flash "Task text cannot be empty", stays in add view | add_task_view_test.go |
| T4 | AddTaskView is active | User types 501 chars | Input limited to 500 chars | add_task_view_test.go |
| T5 | SearchView, user types `:add buy milk` | Presses Enter | `TaskAddedMsg` emitted with text "buy milk" | search_view_test.go |
| T6 | SearchView, user types `:add` (no args) | Presses Enter | `AddTaskPromptMsg` emitted | search_view_test.go |
| T7 | MainModel in doors view | Receives `AddTaskPromptMsg` | viewMode → `ViewAddTask`, AddTaskView created | main_model_test.go |
| T8 | MainModel in any view | Receives `TaskAddedMsg` | Task added to pool, `saveTasks()` called, flash "Task added" | main_model_test.go |
| T9 | MainModel in doors view | User presses `:` | viewMode → `ViewSearch`, search input pre-populated with `:` | main_model_test.go |
| T10 | MainModel with empty pool | Receives `TaskAddedMsg` | Pool count is 1, doors refresh can display the task | main_model_test.go |
| T11 | AddTaskView is active with text | User presses Enter with "  " (whitespace only) | Flash "Task text cannot be empty", stays in add view | add_task_view_test.go |

**Testing standards:**
- **Table-driven tests** following existing patterns in `search_view_test.go` and `main_model_test.go`
- **No mocking frameworks** — use Go's built-in `testing` package
- **MockProvider pattern:** For persistence tests, use `tasks.NewTextFileProvider()` with `t.TempDir()` or create a simple `mockProvider` struct implementing `TaskProvider` that records calls
- **Coverage target:** 70%+ for new code

### Library & Framework Requirements

- **Bubbletea** (`github.com/charmbracelet/bubbletea`) — MVU framework, already imported
- **Bubbles textinput** (`github.com/charmbracelet/bubbles/textinput`) — Text input component, already used in `SearchView`
- **Lipgloss** (`github.com/charmbracelet/lipgloss`) — Styling, already imported
- Use existing styles from `internal/tui/styles.go` for consistency

### Previous Story Intelligence

From Story 1.4 (Quick Search & Command Palette) which established the command palette:
- Command parsing uses `parseCommand()` function splitting on first space
- Commands return `tea.Cmd` closures that emit message types
- Flash messages use `FlashMsg` + `ClearFlashCmd()` pattern with auto-clear timer
- Search view maintains state for context-aware return from detail view

From Story 1.6 (Essential Polish):
- Lipgloss styling applied consistently across all views
- `headerStyle`, `helpStyle`, `flashStyle`, `commandModeStyle` etc. defined in `styles.go`
- Celebration messages for task completion in `styles.go`

### Git Intelligence

Recent commits show:
- Stories 2.3, 2.5 added Apple Notes provider and bidirectional sync
- Story 1.5 added session metrics tracking
- Story 1.6 added essential polish and styling
- All follow the pattern: feat commit → code review findings fix commit
- Code must pass `gofumpt` formatting and `golangci-lint run ./...`

### References

- [Source: internal/tui/search_view.go#executeCommand] — Existing `:add` command handler
- [Source: internal/tui/main_model.go#TaskAddedMsg] — Existing task added handler
- [Source: internal/tui/messages.go] — Message type definitions
- [Source: internal/tui/mood_view.go] — Pattern reference for simple input view
- [Source: internal/tasks/task.go#NewTask] — Task constructor
- [Source: internal/tasks/task_pool.go#AddTask] — Pool addition method
- [Source: docs/architecture/components.md] — Component architecture
- [Source: docs/architecture/coding-standards.md] — Coding standards
- [Source: docs/architecture/test-strategy-and-standards.md] — Testing standards
- [Source: docs/prd/epics-and-stories.md#Story-3.1] — Original story definition

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
