# Story 3.5: Daily Completion Tracking & Comparison

Status: ready-for-dev

## Story

As a user,
I want to see how many tasks I completed today compared to yesterday,
So that I can feel motivated by progress.

## Acceptance Criteria

1. **Given** the user has been using the app across multiple days, **When** any task is completed, **Then** the session display shows: "Completed today: X (yesterday: Y)" **And** if today > yesterday, a positive/motivational message is shown **And** if today = 0 (before this completion), no comparison is shown (avoids discouragement on first completion of the day)

2. **Given** the user types `:stats` in the command palette, **When** daily stats are displayed, **Then** the output includes: tasks completed today, tasks completed yesterday, doors viewed today, current streak (consecutive days with at least one completion)

3. **Given** the user completes a task, **When** the completion is the first of the day and yesterday had completions, **Then** the message shows "Completed today: 1 (yesterday: Y)" without the positive comparison message (since 1 is not necessarily > Y)

4. **Given** the user has no completion history (new user or empty completed.txt), **When** a task is completed, **Then** only "Completed today: 1" is shown (no yesterday comparison) **And** streak shows as 1

5. **Given** the user has completed tasks on consecutive days, **When** `:stats` is run, **Then** the streak counter shows the number of consecutive days ending with today (or yesterday if no completions today yet)

## Tasks / Subtasks

- [ ] Task 1: Create CompletionCounter service in domain layer (AC: #1, #2, #4, #5)
  - [ ] 1.1: Create `internal/tasks/completion_counter.go` with `CompletionCounter` struct including `nowFunc func() time.Time` field for testability (default: `time.Now`)
  - [ ] 1.2: Implement `LoadFromFile(path string) error` to parse `completed.txt` format: `[YYYY-MM-DD HH:MM:SS] uuid | text`. Handle missing file gracefully (return nil error with zero counts, do NOT create the file). Skip malformed lines.
  - [ ] 1.3: Implement `GetTodayCount() int` — count completions where date matches today
  - [ ] 1.4: Implement `GetYesterdayCount() int` — count completions where date matches yesterday
  - [ ] 1.5: Implement `GetStreak() int` — walk backward from most recent day with completions, counting consecutive days
  - [ ] 1.6: Implement `IncrementToday()` — called when a new completion happens in current session (avoids re-reading file)
  - [ ] 1.7: Implement `FormatCompletionMessage() string` — returns the "Completed today: X (yesterday: Y)" string with conditional positive message

- [ ] Task 2: Integrate CompletionCounter into MainModel (AC: #1, #3)
  - [ ] 2.1: Add `completionCounter *tasks.CompletionCounter` field to `MainModel` in `internal/tui/main_model.go`
  - [ ] 2.2: Initialize CompletionCounter in `NewMainModel()` by loading from completed.txt path via `tasks.GetConfigDirPath() + "/completed.txt"`
  - [ ] 2.3: In `TaskCompletedMsg` handler (main_model.go ~line 206), after `provider.MarkComplete()`, call `completionCounter.IncrementToday()`
  - [ ] 2.4: After increment, set flash message using `completionCounter.FormatCompletionMessage()` instead of (or appended to) the existing celebration message

- [ ] Task 3: Enhance `:stats` command (AC: #2)
  - [ ] 3.1: Modify `showStats()` in `internal/tui/search_view.go` to accept or access the CompletionCounter
  - [ ] 3.2: Update stats output format to: "Stats | Today: X | Yesterday: Y | Doors viewed: Z | Streak: N days"
  - [ ] 3.3: Pass CompletionCounter reference from MainModel to SearchView — add `counter *tasks.CompletionCounter` field to `SearchView` struct and update `NewSearchView()` constructor signature and call site in `NewMainModel()`

- [ ] Task 4: Write unit tests (AC: #1, #2, #3, #4, #5)
  - [ ] 4.1: Test `LoadFromFile()` with sample completed.txt data — verify correct date grouping
  - [ ] 4.2: Test `GetTodayCount()` returns correct count for today's date
  - [ ] 4.3: Test `GetYesterdayCount()` returns correct count for yesterday's date
  - [ ] 4.4: Test `GetStreak()` — consecutive days returns correct count
  - [ ] 4.5: Test `GetStreak()` — gap in days resets streak
  - [ ] 4.6: Test `GetStreak()` — no data returns 0
  - [ ] 4.7: Test `GetStreak()` — today has completions, streak includes today
  - [ ] 4.8: Test `GetStreak()` — today has no completions but yesterday does, streak still counts up to yesterday
  - [ ] 4.9: Test `IncrementToday()` correctly updates in-memory count
  - [ ] 4.10: Test `FormatCompletionMessage()` — today > yesterday shows positive message
  - [ ] 4.11: Test `FormatCompletionMessage()` — today = 0 before increment shows no comparison
  - [ ] 4.12: Test `FormatCompletionMessage()` — no yesterday data shows "Completed today: X" only
  - [ ] 4.13: Test enhanced `:stats` output includes today, yesterday, streak
  - [ ] 4.14: Test `LoadFromFile()` with empty file returns zero counts
  - [ ] 4.15: Test `LoadFromFile()` with malformed lines are skipped gracefully
  - [ ] 4.16: Test `LoadFromFile()` with non-existent file returns nil error and zero counts
  - [ ] 4.17: Test `GetStreak()` — no completions in last 3+ days returns 0 even if older completions exist
  - [ ] 4.18: Test with injected `nowFunc` to verify midnight boundary behavior
  - [ ] 4.19: Test `FormatCompletionMessage()` appends to celebration message correctly

## Dev Notes

### Party Mode #1 Findings: Dev Readiness Review

The following gaps were identified by the agent team and addressed:

1. **SearchView constructor wiring (Dev):** SearchView currently takes `tracker *SessionTracker`. The `CompletionCounter` must be added as an additional field. Look at `NewSearchView()` constructor signature and add `counter *tasks.CompletionCounter` parameter. Update call site in `NewMainModel()`.

2. **Celebration message integration (Dev):** `FormatCompletionMessage()` output should be APPENDED to the existing random celebration message (from `styles.go` `CelebrationMessages` pool), not replace it. Format: `"[celebration] | Completed today: X (yesterday: Y)"`. The celebration message provides emotional reinforcement, the counter provides data.

3. **Missing file handling (SM):** `LoadFromFile()` must handle the case where `completed.txt` does not exist (new user). Return nil error with zero counts — do NOT create the file. Use `os.IsNotExist()` check.

4. **Exact :stats output format (SM):**
   ```
   Stats | Today: 3 | Yesterday: 2 | Doors: 12 | Streak: 5 days
   ```
   Single-line format consistent with existing flash message style.

5. **Streak edge case definition (QA/Architect):** Streak = consecutive calendar days (using `time.Now().Local()` date) with at least 1 completion, walking backward from today. If today has 0 completions, walk backward from yesterday instead. If the most recent completion day is more than 1 day ago from today, streak = 0. This avoids timezone ambiguity — always use local time.

6. **Time injection for testability (Architect):** `CompletionCounter` should accept a `nowFunc func() time.Time` field (defaulting to `time.Now`) to allow tests to control the current time. This prevents flaky tests around midnight boundaries.

7. **Performance note (Architect):** For now, full file read on init is fine. Add a code comment noting that if completed.txt grows to 10k+ lines, consider reading only the last N lines or indexing by date.

### Critical Discovery: Existing Completion Infrastructure

The codebase already has the building blocks:

- **`internal/tasks/file_manager.go`** — `AppendCompleted()` writes to `completed.txt` with format `[YYYY-MM-DD HH:MM:SS] uuid | text`
- **`internal/tasks/text_file_provider.go`** — `MarkComplete()` calls `AppendCompleted()` and removes from active pool
- **`internal/tui/main_model.go:~206`** — `TaskCompletedMsg` handler calls `provider.MarkComplete(task.ID)`, increments `doorsView.completedCount`, shows celebration flash
- **`internal/tui/search_view.go:187-199`** — Existing `showStats()` uses `sv.tracker.Finalize()` for session-only stats
- **`internal/tasks/session_tracker.go`** — Tracks session-level metrics (TasksCompleted, DetailViews, RefreshesUsed)
- **`internal/tui/doors_view.go`** — Has `completedCount` field and `IncrementCompleted()` method for session count

### completed.txt Format

```
[2026-03-01 10:15:30] uuid-001 | Write architecture document
[2026-03-01 11:20:00] uuid-002 | Review PRD
[2026-03-02 09:00:00] uuid-004 | Fix login bug
```

**Parsing strategy:** Extract date from `[YYYY-MM-DD` prefix (first 11 chars after `[`), group by date string, count per day.

### What Needs to Be Built

1. **`CompletionCounter`** — New domain-layer service (`internal/tasks/completion_counter.go`) that:
   - Parses completed.txt on initialization
   - Groups completions by date
   - Calculates today/yesterday counts and streak
   - Provides in-memory increment for current session additions
   - Formats display messages with conditional positive reinforcement

2. **MainModel integration** — Wire CompletionCounter into the task completion flow to show daily comparison on each completion

3. **Enhanced `:stats`** — Update search_view.go `showStats()` to include daily and streak data from CompletionCounter

### Architecture & Pattern Compliance

- **MVU Pattern (Bubbletea Elm Architecture):** All state changes via messages. No direct mutation outside `Update()`.
- **Domain-layer separation:** CompletionCounter lives in `internal/tasks/` (domain layer), not `internal/tui/` (presentation layer)
- **Constructor injection:** CompletionCounter created in `NewMainModel()` and passed to views that need it
- **Atomic writes:** CompletionCounter only reads completed.txt; writes are handled by existing `AppendCompleted()` in file_manager.go
- **Error handling:** LoadFromFile should return error, but caller should handle gracefully (log warning, use zero counts) — never crash
- **gofumpt formatting:** All code must pass `gofumpt` formatting check
- **No panics:** Use error returns, not panics

### File Structure Requirements

Files to create:
- `internal/tasks/completion_counter.go` — CompletionCounter service
- `internal/tasks/completion_counter_test.go` — Tests for CompletionCounter

Files to modify:
- `internal/tui/main_model.go` — Add completionCounter field, initialize in NewMainModel, use in TaskCompletedMsg handler
- `internal/tui/search_view.go` — Update showStats() to include daily comparison and streak
- `internal/tui/messages.go` — Add DailyStatsMsg if needed (may not be needed if using flash messages directly)

### Testing Requirements

**Test Files:**
- `internal/tasks/completion_counter_test.go` (NEW) — All CompletionCounter logic tests
- `internal/tui/search_view_test.go` (MODIFY) — Test enhanced :stats output
- `internal/tui/main_model_test.go` (MODIFY) — Test completion flash includes daily comparison

**Test Helper Patterns (from existing codebase):**
- Use `t.TempDir()` for test completed.txt files
- Use `os.WriteFile()` to create test data with specific dates
- Use `time.Now()` and `time.Now().AddDate(0, 0, -1)` for today/yesterday in test data
- Table-driven tests following existing patterns
- No mocking frameworks — use Go's built-in `testing` package
- Coverage target: 70%+ for new code

**BDD Test Scenarios:**

| # | Given | When | Then | File |
|---|-------|------|------|------|
| T1 | completed.txt has 3 entries today, 2 yesterday | LoadFromFile called | GetTodayCount()=3, GetYesterdayCount()=2 | completion_counter_test.go |
| T2 | completed.txt is empty | LoadFromFile called | GetTodayCount()=0, GetYesterdayCount()=0, GetStreak()=0 | completion_counter_test.go |
| T3 | Completions on 5 consecutive days ending today | GetStreak() called | Returns 5 | completion_counter_test.go |
| T4 | Completions on 3 days, gap, then 2 more days ending today | GetStreak() called | Returns 2 (only current streak) | completion_counter_test.go |
| T5 | No completions today, 3 consecutive days ending yesterday | GetStreak() called | Returns 3 (streak up to yesterday) | completion_counter_test.go |
| T6 | Today=3, Yesterday=2 | FormatCompletionMessage() called | Returns "Completed today: 3 (yesterday: 2) - You're on a roll!" or similar | completion_counter_test.go |
| T7 | Today=1, Yesterday=0 (no yesterday data) | FormatCompletionMessage() called | Returns "Completed today: 1" (no comparison) | completion_counter_test.go |
| T8 | Today=0 (pre-increment state) | FormatCompletionMessage() called | Returns "" or minimal message (no comparison) | completion_counter_test.go |
| T9 | Counter initialized with 2 today | IncrementToday() called | GetTodayCount() returns 3 | completion_counter_test.go |
| T10 | completed.txt has malformed lines mixed with valid | LoadFromFile called | Valid entries counted, malformed skipped, no error | completion_counter_test.go |
| T11 | MainModel receives TaskCompletedMsg | Completion processed | Flash includes daily comparison text | main_model_test.go |
| T12 | User types :stats | Stats displayed | Output includes today, yesterday, streak | search_view_test.go |
| T13 | Today=2, Yesterday=5 | FormatCompletionMessage() called | Returns "Completed today: 2 (yesterday: 5)" without positive message | completion_counter_test.go |
| T14 | No completions in last 3 days, completions before that | GetStreak() called | Returns 0 | completion_counter_test.go |
| T15 | Completions only today (first ever day) | GetStreak() called | Returns 1 | completion_counter_test.go |

### Party Mode #2 Findings: Test Readiness Review

The following gaps were identified to enable TEA to create tests before dev implements:

1. **Explicit Constructor Signatures:**
   ```go
   // Primary constructor
   func NewCompletionCounter() *CompletionCounter

   // Constructor with time injection for testing
   func NewCompletionCounterWithNow(nowFunc func() time.Time) *CompletionCounter

   // All public methods:
   func (cc *CompletionCounter) LoadFromFile(path string) error
   func (cc *CompletionCounter) GetTodayCount() int
   func (cc *CompletionCounter) GetYesterdayCount() int
   func (cc *CompletionCounter) GetStreak() int
   func (cc *CompletionCounter) IncrementToday()
   func (cc *CompletionCounter) FormatCompletionMessage() string
   ```

2. **Internal struct shape (for test construction):**
   ```go
   type CompletionCounter struct {
       dateCounts map[string]int  // "2026-03-02" -> 3
       nowFunc    func() time.Time
   }
   ```

3. **Exact FormatCompletionMessage output rules:**
   - If today count = 0: return `""` (empty string)
   - If today > 0 and no yesterday data (yesterday = 0): return `"Completed today: X"`
   - If today > 0 and yesterday > 0 and today > yesterday: return `"Completed today: X (yesterday: Y) - [random positive message]"`
   - If today > 0 and yesterday > 0 and today <= yesterday: return `"Completed today: X (yesterday: Y)"`
   - Note: The celebration message prefix is added by the CALLER (MainModel), not by FormatCompletionMessage. This keeps the counter pure domain logic.

4. **Test helper for creating test data:**
   ```go
   // In completion_counter_test.go
   func writeCompletedFile(t *testing.T, dir string, entries map[string][]string) string {
       // entries: date -> []taskTexts, e.g. {"2026-03-01": {"task1", "task2"}}
       // Writes proper formatted completed.txt to dir
       // Returns file path
   }
   ```

5. **Test time freezing pattern:**
   ```go
   frozenNow := time.Date(2026, 3, 2, 14, 0, 0, 0, time.Local)
   cc := NewCompletionCounterWithNow(func() time.Time { return frozenNow })
   ```

6. **Permission error handling:** `LoadFromFile` should return error on permission issues (not swallow them). Only swallow `os.IsNotExist`. Tests should verify this distinction.

### Library & Framework Requirements

- **Go standard library only** for CompletionCounter — `time`, `os`, `bufio`, `strings`, `fmt`
- **Bubbletea** (`github.com/charmbracelet/bubbletea`) — already imported, MVU framework
- **Lipgloss** (`github.com/charmbracelet/lipgloss`) — already imported, for styling
- Use existing styles from `internal/tui/styles.go` — specifically `flashStyle`, `colorComplete` for positive messages

### Previous Story Intelligence

From Story 3.1 (Quick Add Mode):
- Added `ViewAddTask` view mode and `AddTaskView` component
- Pattern: new view modes added to `MainModel` enum, message types in `messages.go`
- Flash messages use `FlashMsg{Text: "..."}` + `ClearFlashCmd()` pattern
- `:` shortcut from doors view opens command mode

From Story 3.4 (Door Feedback Options):
- Added `ViewFeedback` view mode and feedback component
- Pattern: feedback data stored with task ID and timestamp
- Recent addition — code patterns are fresh and consistent

### Git Intelligence

Recent commits show:
- Story 3.1-3.4 implemented following pattern: feat commit → review fixes
- All code passes `gofumpt` formatting and `golangci-lint run ./...`
- Tests follow table-driven pattern with `makePool()` and `makeModel()` helpers
- File naming convention: `snake_case.go` and `snake_case_test.go`

### Positive Completion Messages Pool

When today > yesterday, use varied messages from a pool (similar to existing celebration messages in styles.go):
- "You're ahead of yesterday!"
- "Beating yesterday's pace!"
- "On a roll!"
- "Momentum building!"

Keep them short and consistent with existing "progress over perfection" messaging tone.

### References

- [Source: internal/tasks/file_manager.go#AppendCompleted] — Writes to completed.txt
- [Source: internal/tasks/text_file_provider.go#MarkComplete] — Completion flow
- [Source: internal/tui/main_model.go#TaskCompletedMsg] — Completion message handler
- [Source: internal/tui/search_view.go#showStats] — Current :stats implementation
- [Source: internal/tasks/session_tracker.go] — Session metrics tracking
- [Source: internal/tui/doors_view.go#IncrementCompleted] — Session completion counter
- [Source: internal/tui/styles.go] — Styling patterns and celebration messages
- [Source: internal/tui/messages.go] — Message type definitions
- [Source: docs/architecture/data-storage-schema.md] — Data file format specifications
- [Source: docs/architecture/coding-standards.md] — Coding standards
- [Source: docs/architecture/test-strategy-and-standards.md] — Testing standards

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
