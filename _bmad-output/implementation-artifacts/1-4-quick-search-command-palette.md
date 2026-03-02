# Story 1.4: Quick Search & Command Palette

Status: ready-for-dev

## Story

As a user,
I want to search for specific tasks and execute commands via text input,
so that I can efficiently find and act on tasks without relying solely on the three doors.

## Acceptance Criteria

1. **Search Mode Activation:** Pressing `/` from the three doors view opens a text input field at the bottom with placeholder "Search tasks... (or :command for commands)"
2. **Live Search:** As the user types characters, matching tasks display from bottom-up (live substring matching). The list updates with each keystroke. If no matches, "No tasks match '<text>'" is shown. Empty input shows no results.
3. **Search Results Navigation:** User navigates results with arrow keys (up/down), WASD (w/s), or HJKL (j/k for up/down). The selected result is highlighted. Pressing Enter opens the selected task in the detail view (same DetailView from Story 1.3).
4. **Context-Aware Return from Detail:** When a task detail was opened from search, pressing Esc from the detail view returns to the search view with search text preserved and previous selection highlighted.
5. **Exit Search:** Pressing Esc or Ctrl+C from the search input (not detail view) closes search mode and returns to three doors view.
6. **Command Mode Switch:** When the search input is empty and the user types `:` as the first character, the input switches to command mode with visual indicator (e.g., different prompt color/text).
7. **`:add <task text>`:** Adds a new task to tasks.yaml and the in-memory pool.
8. **`:mood [mood]`:** Logs a mood entry. If no mood parameter, opens the mood capture dialog (same MoodView from Story 1.3).
9. **`:stats`:** Displays session statistics (tasks completed, doors viewed, time in session).
10. **`:help`:** Displays available commands and key bindings.
11. **`:quit` / `:exit`:** Exits the application cleanly.
12. **Unknown Command:** Typing an invalid command shows "Unknown command: '<command>'. Type :help for available commands."

## Tasks / Subtasks

- [ ] Task 1: Add `charmbracelet/bubbles` dependency for textinput component (AC: 1)
  - [ ] 1.1: Run `go get github.com/charmbracelet/bubbles@v0.20.0` to add dependency
  - [ ] 1.2: Run `go mod tidy` to clean up
  - [ ] 1.3: Verify import path: `"github.com/charmbracelet/bubbles/textinput"`

- [ ] Task 2: Create SearchView component (AC: 1, 2, 3, 5)
  - [ ] 2.1: Create `internal/tui/search_view.go` with SearchView struct
  - [ ] 2.2: SearchView fields: `textInput` (textinput.Model from bubbles/textinput), `results` ([]*tasks.Task), `selectedIndex` (int), `query` (string), `pool` (*tasks.TaskPool), `tracker` (*tasks.SessionTracker), `width` (int), `isCommandMode` (bool)
  - [ ] 2.3: Implement `NewSearchView(pool, tracker)` constructor:
    - Initialize `ti := textinput.New()`
    - Set `ti.Placeholder = "Search tasks... (or :command for commands)"`
    - Call `ti.Focus()` to activate cursor
    - Set `ti.CharLimit = 256`
    - Set `ti.Width` to available width
  - [ ] 2.4: Implement `filterTasks(query string) []*tasks.Task` as a method on SearchView (NOT a separate domain service - keep in TUI layer since this is simple substring matching). Case-insensitive match: `strings.Contains(strings.ToLower(task.Text), strings.ToLower(query))`. Search ALL tasks in pool regardless of status.
  - [ ] 2.5: Implement `Update(msg) tea.Cmd` - CRITICAL: Must delegate `tea.KeyMsg` and other messages to `textInput.Update(msg)` first to handle cursor/typing, THEN handle navigation keys (up/down/j/k/w/s), Enter, Esc. Return both the textinput cmd and any SearchView cmd.
  - [ ] 2.6: Implement `View() string` - render results above input field, input at bottom. "Bottom-up" means: most relevant/best match at bottom (closest to input field), earlier results rendered above.
  - [ ] 2.7: Handle result navigation: up/down arrows, j/k keys to move selectedIndex. NOTE: w/s and WASD should NOT be navigation keys in search mode because they conflict with typing. Only use up/down arrows and j/k (vi-style) for navigation.
  - [ ] 2.8: On Enter with selected result, send `SearchResultSelectedMsg{Task: selectedTask}`
  - [ ] 2.9: On Esc (when textinput is empty or focused), send `SearchClosedMsg{}` (distinct from ReturnToDoorsMsg for clarity)
  - [ ] 2.10: Implement `SetWidth(w int)` for terminal width updates - must also update `textInput.Width`
  - [ ] 2.11: Implement `RestoreState(query string, selectedIndex int)` for context-aware return from detail - sets textInput value and re-runs filter

- [ ] Task 3: Implement Command Mode (AC: 6, 7, 8, 9, 10, 11, 12)
  - [ ] 3.1: Detect `:` as first character in textInput.Value() to switch to command mode
  - [ ] 3.2: Visual indicator for command mode (change prompt style/color using `commandModeStyle`)
  - [ ] 3.3: Implement `parseCommand(input string) (cmd string, args string)` - splits `:cmd args`
  - [ ] 3.4: Implement `:add <task text>`:
    - Create task via `tasks.NewTask(text)`
    - Send `TaskAddedMsg{Task: newTask}` to MainModel
    - MainModel handles: `pool.AddTask(task)` then `m.saveTasks()` (uses existing persistence pattern)
    - Show flash: "Task added: <truncated text>"
    - If no text provided, show flash: "Usage: :add <task text>"
  - [ ] 3.5: Implement `:mood [mood]` - if arg provided log directly, else send ShowMoodMsg
  - [ ] 3.6: Implement `:stats` - display session metrics from SessionTracker (completed count, doors viewed, session duration)
  - [ ] 3.7: Implement `:help` - display formatted help with all key bindings and commands
  - [ ] 3.8: Implement `:quit` / `:exit` - send tea.Quit
  - [ ] 3.9: Handle unknown commands with error message

- [ ] Task 4: Add new message types to messages.go (AC: all)
  - [ ] 4.1: Add `SearchResultSelectedMsg{Task *tasks.Task}` - sent when user selects a search result
  - [ ] 4.2: Add `TaskAddedMsg{Task *tasks.Task}` - sent when `:add` command creates a task
  - [ ] 4.3: Add `SearchClosedMsg{}` - sent when user presses Esc in search view (distinct from ReturnToDoorsMsg)
  - [ ] 4.4: Add `ReturnToSearchMsg{Query string; SelectedIndex int}` - sent when Esc from detail view that was opened from search

- [ ] Task 5: Integrate SearchView into MainModel (AC: 1, 4, 5)
  - [ ] 5.1: Add `ViewSearch ViewMode = iota` after `ViewMood` in the ViewMode enum (value = 3)
  - [ ] 5.2: Add `searchView *SearchView` field to MainModel
  - [ ] 5.3: Add `previousView ViewMode` field to MainModel for context-aware return
  - [ ] 5.4: Add `searchQuery string` and `searchSelectedIndex int` fields to save search state when entering detail
  - [ ] 5.5: Handle `/` key in `updateDoors()` to create SearchView and switch to ViewSearch
  - [ ] 5.6: Handle `SearchResultSelectedMsg` in Update():
    - Save search state: `m.searchQuery = m.searchView.textInput.Value()`, `m.searchSelectedIndex = m.searchView.selectedIndex`
    - Set `m.previousView = ViewSearch`
    - Create DetailView for selected task, switch to ViewDetail
  - [ ] 5.7: Handle `ReturnToSearchMsg` in Update():
    - Create or restore SearchView with `RestoreState(msg.Query, msg.SelectedIndex)`
    - Switch to ViewSearch
  - [ ] 5.8: Modify existing `ReturnToDoorsMsg` handler: check if `m.previousView == ViewSearch`, if so send `ReturnToSearchMsg` with saved state instead of going to doors
  - [ ] 5.9: Handle `SearchClosedMsg` in Update() - set viewMode to ViewDoors, nil out searchView, reset previousView
  - [ ] 5.10: Handle `TaskAddedMsg` in Update():
    - `m.pool.AddTask(msg.Task)`
    - `m.saveTasks()` (reuses existing persistence helper)
    - Show flash: "Task added"
    - Return to search view (stay in search mode after add)
  - [ ] 5.11: Add `updateSearch(msg)` method to delegate to SearchView
  - [ ] 5.12: Update `View()` method to render ViewSearch when active
  - [ ] 5.13: Update `WindowSizeMsg` handler to pass width to SearchView (including textInput width)

- [ ] Task 6: Update help text in DoorsView (AC: 1)
  - [ ] 6.1: Add `/ search` to the help text at the bottom of doors view

- [ ] Task 7: Add search-related styles (AC: 1, 2, 3, 6)
  - [ ] 7.1: Add `searchInputStyle` to styles.go - style for the text input area
  - [ ] 7.2: Add `searchResultStyle` to styles.go - style for search result items
  - [ ] 7.3: Add `searchSelectedStyle` to styles.go - style for highlighted search result
  - [ ] 7.4: Add `commandModeStyle` to styles.go - distinct visual for command mode indicator
  - [ ] 7.5: Add `statsStyle` to styles.go - style for stats display
  - [ ] 7.6: Add `helpViewStyle` to styles.go - style for help display

- [ ] Task 8: Write unit tests (AC: all)
  - [ ] 8.1: Create `internal/tui/search_view_test.go`
  - [ ] 8.2: Test `filterTasks` with various queries (exact match, partial, case-insensitive, no match, empty query)
  - [ ] 8.3: Test `parseCommand` parsing (`:add text`, `:mood`, `:stats`, `:help`, `:quit`, `:exit`, unknown command, empty command)
  - [ ] 8.4: Test navigation key handling (up/down arrows and j/k move selectedIndex correctly, bounds checking)
  - [ ] 8.5: Test Esc sends SearchClosedMsg (NOT ReturnToDoorsMsg)
  - [ ] 8.6: Test Enter on selected result sends SearchResultSelectedMsg with correct task
  - [ ] 8.7: Test command mode activation (`:` as first char sets isCommandMode)
  - [ ] 8.8: Test `:add` creates task via TaskAddedMsg with correct text
  - [ ] 8.9: Test `:add` with no text shows usage error
  - [ ] 8.10: Test `:add` persistence roundtrip: task added appears in pool AND can be saved
  - [ ] 8.11: Test RestoreState correctly restores query and selectedIndex
  - [ ] 8.12: Test SetWidth updates both view width and textInput.Width
  - [ ] 8.13: Test edge case: terminal resize while search is active
  - [ ] 8.14: Create `internal/tui/main_model_search_test.go` - test view transitions for search flow
  - [ ] 8.15: Test context-aware return: doors -> search -> detail -> search (with preserved state)
  - [ ] 8.16: Test context-aware return: doors -> detail -> doors (normal flow, no search involvement)

## Dev Notes

### Architecture Compliance

This story adds a new view component (`SearchView`) following the same patterns established in Stories 1.1-1.3. The architecture is consistent:

**Required pattern:**
- New SearchView follows the same component pattern as DoorsView, DetailView, MoodView
- Uses Bubbletea MVU with message-based communication
- No direct view-to-view calls; all routing through MainModel
- Constructor injection for dependencies (pool, tracker)

**View routing update:**
```
MainModel.viewMode:
  ViewDoors   -> (press /)   -> ViewSearch
  ViewSearch  -> (Enter)     -> ViewDetail (with previousView = ViewSearch)
  ViewDetail  -> (Esc)       -> ViewSearch (if previousView == ViewSearch, restore state)
  ViewSearch  -> (Esc)       -> ViewDoors
```

[Source: docs/architecture/components.md, docs/architecture/high-level-architecture.md]

### New Dependency Required

**`github.com/charmbracelet/bubbles`** v0.20.0 - for textinput component:
- Import: `"github.com/charmbracelet/bubbles/textinput"`
- Initialize: `ti := textinput.New()` then `ti.Focus()` to activate cursor
- Set: `ti.Placeholder`, `ti.CharLimit = 256`, `ti.Width`

```bash
go get github.com/charmbracelet/bubbles@v0.20.0
go mod tidy
```

**CRITICAL: `textinput.Model` is a value type** - store as `textinput.Model` NOT `*textinput.Model`. Update pattern:
```go
var cmd tea.Cmd
sv.textInput, cmd = sv.textInput.Update(msg)
```
You MUST delegate all `tea.Msg` to `textInput.Update()` before handling your own keys, or typing/cursor won't work.

[Source: docs/architecture/tech-stack.md]

### Key Implementation Details

**Search Algorithm:**
- Case-insensitive substring match: `strings.Contains(strings.ToLower(task.Text), strings.ToLower(query))`
- Search ALL tasks in pool (including blocked, in-progress, etc. - not just available-for-doors)
- Results ordered by relevance: exact prefix match first, then substring match
- No fuzzy matching needed for tech demo

**Bottom-Up Results Display:**
- Results rendered above the input field
- Most relevant result at the bottom (closest to input)
- Scroll up for more results
- Selected result highlighted with `searchSelectedStyle`

**Command Parsing:**
```go
func parseCommand(input string) (cmd string, args string) {
    input = strings.TrimPrefix(input, ":")
    parts := strings.SplitN(input, " ", 2)
    cmd = strings.ToLower(parts[0])
    if len(parts) > 1 {
        args = parts[1]
    }
    return
}
```

**Context-Aware Navigation Stack:**
- MainModel tracks `previousView ViewMode` to know where to return
- When opening detail from search: `previousView = ViewSearch`, save `searchQuery` and `searchSelectedIndex`
- DetailView already sends `ReturnToDoorsMsg` on Esc - MainModel intercepts this: if `previousView == ViewSearch`, send `ReturnToSearchMsg{Query, SelectedIndex}` instead
- When `previousView == ViewDoors` (normal door->detail flow): handle `ReturnToDoorsMsg` as before
- `SearchClosedMsg` is distinct from `ReturnToDoorsMsg` - used when user explicitly exits search with Esc

**Message flow diagram:**
```
Doors --(/ key)--> Search --(Enter on result)--> Detail --(Esc)--> Search (restored)
Doors --(Enter on door)--> Detail --(Esc)--> Doors (normal)
Search --(Esc)--> Doors (SearchClosedMsg)
```

**`:add` Implementation:**
```go
// In SearchView command handling:
case "add":
    if args == "" {
        // Show error: "Usage: :add <task text>"
        return showFlash("Usage: :add <task text>")
    }
    newTask := tasks.NewTask(args)
    return func() tea.Msg { return TaskAddedMsg{Task: newTask} }
```

**`:stats` Display:**
```
Session Statistics:
  Duration: 12m 34s
  Tasks completed: 3
  Doors viewed: 15
  Doors refreshed: 4
  Moods logged: 1
```

**`:help` Display:**
```
Key Bindings:
  a/left   - Select left door
  w/up     - Select center door
  d/right  - Select right door
  s/down   - Re-roll doors
  Enter    - Open selected door
  m        - Log mood
  /        - Search tasks
  q        - Quit

Commands (type : to enter):
  :add <text>  - Add new task
  :mood [mood] - Log mood
  :stats       - Show session stats
  :help        - Show this help
  :quit        - Exit application
```

### Previous Story Intelligence

**From Story 1.3 (completed, merged PR #5):**
- MainModel has clean view routing with ViewDoors, ViewDetail, ViewMood
- Message-based communication works well (TaskCompletedMsg, TaskUpdatedMsg, etc.)
- DetailView component pattern is established - SearchView should follow same pattern
- Flash message system (FlashMsg + ClearFlashCmd) is available for confirmations
- saveTasks() helper exists in MainModel for persistence
- DoorsView exposes pool, tracker, currentDoors, selectedDoorIndex
- SetWidth pattern used across all views

**From Story 1.3 code review (PR #7):**
- Comprehensive TUI test suite was added covering detail_view, doors_view, main_model, mood_view, styles
- Tests use message-based testing pattern: create model, send tea.KeyMsg, assert state
- Follow same test patterns for SearchView tests

**Key patterns from existing code:**
- Views return `tea.Cmd` from `Update()`, not `(tea.Model, tea.Cmd)`
- MainModel delegates to views and handles returned commands
- All views have `SetWidth(w int)` method
- Constructor pattern: `NewXxxView(dependencies...) *XxxView`

### Git Intelligence

Recent commits:
- `0a99cfa` Merge PR #8 - Story 1.7 CI/CD Pipeline & Alpha Release
- `52ca2d1` test: Add comprehensive TUI test suite for Story 1.3
- `fb5528f` fix: resolve duplicate imports after rebase onto Story 1.2

**Established file patterns:**
- Source: `internal/tui/xxx_view.go` + `internal/tui/xxx_view_test.go`
- Messages in: `internal/tui/messages.go`
- Styles in: `internal/tui/styles.go`
- Domain code in: `internal/tasks/`

### Testing Standards

- **Coverage goals:** 70%+ for search/command logic, 20%+ for TUI rendering
- **Test patterns:** Table-driven tests (see existing `*_test.go` files)
- Use message-based TUI testing: create model, send tea.KeyMsg, assert state changes
- Do NOT use teatest package

**Test file mapping:**

| Source File | Test File | Priority |
|---|---|---|
| internal/tui/search_view.go | internal/tui/search_view_test.go | HIGH |
| internal/tui/main_model.go | internal/tui/main_model_test.go (extend) | HIGH |
| internal/tui/messages.go | (no tests needed - pure data types) | N/A |
| internal/tui/styles.go | (no tests needed - style definitions) | N/A |

**Test helper for creating populated TaskPool:**
```go
func testPool(texts ...string) *tasks.TaskPool {
    pool := tasks.NewTaskPool()
    for _, text := range texts {
        pool.AddTask(tasks.NewTask(text))
    }
    return pool
}
```

**Test fixture data (use in search tests):**
- Pool with 5 tasks in various statuses: "Write unit tests" (todo), "Fix login bug" (in-progress), "Review PR for auth" (in-review), "Deploy to staging" (blocked), "Update README" (todo)
- This provides: substring match ("unit"), case sensitivity ("fix" vs "Fix"), status variety, multi-word search ("login bug")

**Simulating typing in textinput during tests:**
```go
// Send individual key runes to simulate typing
for _, r := range "search query" {
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
    sv.Update(msg)
}
// Or send the textinput value directly for unit tests:
sv.textInput.SetValue("search query")
```

**BDD Acceptance Test Scenarios:**

**AC1 - Search Mode Activation:**
```
Given the user is in the three doors view
When the user presses '/'
Then a text input field appears at the bottom
And the input has placeholder "Search tasks... (or :command for commands)"
And the cursor is active in the input field
```

**AC2 - Live Search:**
```
Given search mode is active with pool containing ["Write tests", "Fix bug", "Review PR"]
When the user types "te"
Then "Write tests" appears in the results
And "Fix bug" and "Review PR" do NOT appear
When the user clears the input
Then no results are shown
```

**AC3 - Results Navigation:**
```
Given search results show ["Write tests", "Write docs"] (2 matches)
When the user presses down arrow
Then the first result is highlighted
When the user presses down arrow again
Then the second result is highlighted
When the user presses Enter
Then SearchResultSelectedMsg is sent with the highlighted task
```

**AC4 - Context-Aware Return:**
```
Given the user opened a task from search (query="test", selectedIndex=1)
When the user presses Esc in detail view
Then the search view is restored
And the search input contains "test"
And the previously selected result (index 1) is highlighted
```

**AC5 - Exit Search:**
```
Given search mode is active
When the user presses Esc
Then SearchClosedMsg is sent
And the view returns to three doors
```

**AC6 - Command Mode:**
```
Given search input is empty
When the user types ":"
Then command mode is visually indicated (style change)
When the user types ":add Buy groceries" and presses Enter
Then TaskAddedMsg is sent with task text "Buy groceries"
```

**AC12 - Unknown Command:**
```
Given command mode is active
When the user types ":foo" and presses Enter
Then flash message shows "Unknown command: 'foo'. Type :help for available commands."
```

**Edge Case Tests:**
```
Given search mode is active and the pool is empty
When the user types "anything"
Then "No tasks match 'anything'" is shown

Given search mode is active
When the user types special chars "[]().*+"
Then no crash occurs (literal string match, not regex)

Given command mode is active
When the user types ":add" with no text and presses Enter
Then flash message shows "Usage: :add <task text>"

Given search mode is active
When terminal resizes
Then search input width updates correctly
And results re-render at new width
```

[Source: docs/architecture/test-strategy-and-standards.md]

### Coding Standards

- Go 1.25.4 strict
- `gofumpt` formatting required
- `golangci-lint` must pass
- NO `fmt.Println` in TUI code - all rendering through Bubbletea `View()` method
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- No panics in production code
- Constructor functions: `NewXxx()` pattern
- PascalCase for exports, camelCase for private

[Source: docs/architecture/coding-standards.md]

### Project Structure Notes

**New files for Story 1.4:**
```
internal/tui/
  search_view.go           # NEW - Search + Command Palette component
  search_view_test.go      # NEW - Tests for search functionality
```

**Modified files:**
```
internal/tui/
  main_model.go            # ADD ViewSearch mode, search integration
  messages.go              # ADD SearchResultSelectedMsg, TaskAddedMsg, ShowStatsMsg, ShowHelpMsg
  styles.go                # ADD search-related styles
  doors_view.go            # ADD / key hint to help text
go.mod                     # ADD charmbracelet/bubbles dependency
```

### References

- [Source: docs/prd/epics-and-stories.md#Story 1.4] - Full acceptance criteria
- [Source: docs/architecture/components.md] - Component definitions
- [Source: docs/architecture/high-level-architecture.md] - Two-layer architecture
- [Source: docs/architecture/data-models.md] - Task, TaskPool models
- [Source: docs/architecture/coding-standards.md] - Go coding standards
- [Source: docs/architecture/test-strategy-and-standards.md] - Testing approach
- [Source: docs/architecture/tech-stack.md] - Technology versions
- [Source: docs/architecture/source-tree.md] - Directory structure
- [Source: _bmad-output/implementation-artifacts/1-3-door-selection-task-status.md] - Previous story

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
