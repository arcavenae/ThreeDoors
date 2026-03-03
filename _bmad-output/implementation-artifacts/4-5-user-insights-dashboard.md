# Story 4.5: User Insights Dashboard

Status: ready-for-dev

## Story

As a user,
I want a dashboard and insights view that shows my productivity patterns, mood-task correlations, door position preferences, completion trends, and streaks,
So that I can understand my behavior and celebrate multi-dimensional progress beyond just task counts.

## Acceptance Criteria

1. **Given** the user types `:dashboard` or `:insights` in command mode, **When** the command is executed, **Then** a full-screen insights view appears showing: completion trends (day-over-day, week-over-week), mood-productivity correlations, door position preference distribution, current streak info, and a "better than yesterday" multi-dimensional summary. **And** the view is scrollable if content exceeds terminal height. **And** pressing Esc returns to the previous view.

2. **Given** accumulated session history in sessions.jsonl, **When** the insights view loads, **Then** a pattern analyzer processes all session records to compute: daily completion counts (last 7 days), mood frequency distribution, mood-to-completion-rate correlations (e.g., "Focused: avg 4.2 tasks/session"), door position selection percentages (left/center/right), and task bypass rates. **And** a minimum of 3 sessions are required before showing correlations (cold start guard). **And** if insufficient data, a friendly message is shown: "Keep using ThreeDoors to unlock insights! (X more sessions needed)".

3. **Given** the user starts a new session, **When** the greeting is displayed in DoorsView, **Then** the greeting includes a compact "better than yesterday" multi-dimensional line showing: tasks completed today vs yesterday, current streak, and mood trend (if mood data exists). **And** the comparison is encouraging regardless of direction (e.g., "3 tasks today vs 5 yesterday — every door opened counts"). **And** if no previous session data exists, the line is omitted.

4. **Given** session data with mood entries and completion counts, **When** the pattern analyzer computes correlations, **Then** it calculates average tasks completed per session for each mood type, identifies the user's most productive mood, identifies the user's most frequent mood, and displays these as simple text summaries (not charts). **And** correlations use a minimum sample size of 2 sessions per mood type.

5. **Given** session data with door selection records, **When** door position preferences are analyzed, **Then** the analyzer calculates selection percentage for each position (left/center/right), identifies any strong bias (>50% for one position), and displays as "Left: X% | Center: Y% | Right: Z%". **And** if a bias is detected, a gentle note appears: "You tend to pick the [position] door — try mixing it up!"

6. **Given** accumulated completion history, **When** completion trends are displayed, **Then** the last 7 days are shown as a simple text-based sparkline or bar chart (using Unicode block characters), with each day's count labeled. **And** week-over-week comparison shows this week's total vs last week's total with percentage change.

## Tasks / Subtasks

- [ ] Task 1: Create PatternAnalyzer service in domain layer (AC: #2, #4, #5)
  - [ ] 1.1: Create `internal/tasks/pattern_analyzer.go` with `PatternAnalyzer` struct
  - [ ] 1.2: Implement `LoadSessions(path string) error` to parse sessions.jsonl (JSON Lines format, one `SessionMetrics` per line)
  - [ ] 1.3: Implement `GetDailyCompletions(days int) map[string]int` — returns date -> completion count for last N days
  - [ ] 1.4: Implement `GetMoodCorrelations() []MoodCorrelation` — returns mood -> avg tasks completed per session, sorted by productivity
  - [ ] 1.5: Implement `GetDoorPositionPreferences() DoorPreferences` — returns left/center/right percentages and bias detection
  - [ ] 1.6: Implement `GetWeekOverWeek() WeekComparison` — this week total vs last week total with percentage change
  - [ ] 1.7: Implement `HasSufficientData() bool` — returns true if >= 3 sessions exist
  - [ ] 1.8: Implement `GetSessionsNeeded() int` — returns how many more sessions needed for insights
  - [ ] 1.9: Implement `GetMostProductiveMood() string` — returns mood with highest avg completion
  - [ ] 1.10: Implement `GetMostFrequentMood() string` — returns mood logged most often
  - [ ] 1.11: Implement `GetMostRecentMood() string` — returns Mood field from last MoodEntry of most recent session with mood data; returns `""` if none
  - [ ] 1.12: Implement `GetBypassRate() float64` — returns percentage of door refreshes (tasks shown but not selected) across all sessions; uses `TaskBypasses` field

- [ ] Task 2: Create InsightsView TUI component (AC: #1, #6)
  - [ ] 2.1: Create `internal/tui/insights_view.go` with `InsightsView` struct
  - [ ] 2.2: Add `ViewInsights` to view mode constants in main_model.go
  - [ ] 2.3: Implement `NewInsightsView(analyzer *tasks.PatternAnalyzer, counter *tasks.CompletionCounter) *InsightsView`
  - [ ] 2.4: Implement `Update(msg tea.Msg) tea.Cmd` and `View() string` — follow HealthView sub-view pattern (do NOT implement `Init()` or full `tea.Model` interface)
  - [ ] 2.5: Implement `View()` rendering with sections: Header, Completion Trends, Mood Correlations, Door Preferences, Streak Info
  - [ ] 2.6: Implement scrolling support using viewport or manual offset for long content
  - [ ] 2.7: Implement sparkline rendering using Unicode block chars (▁▂▃▄▅▆▇█) for 7-day completion trend
  - [ ] 2.8: Handle cold start state — show friendly "need more data" message when insufficient sessions
  - [ ] 2.9: Esc key sends `ReturnToDoorsMsg{}` (reuse existing message type, do NOT create `ReturnFromInsightsMsg`)

- [ ] Task 3: Register `:dashboard` and `:insights` commands (AC: #1)
  - [ ] 3.1: Add `:dashboard` and `:insights` cases to the `switch cmd` block in `executeCommand()` in `search_view.go` (NOT `parseCommand()` — that only splits the string)
  - [ ] 3.2: Both commands return a `tea.Cmd` that sends `ShowInsightsMsg{}` — do NOT pass PatternAnalyzer to SearchView; MainModel handles the view switch
  - [ ] 3.3: Update `:help` output string to include `:dashboard` / `:insights` commands

- [ ] Task 4: Create "Better Than Yesterday" multi-dimensional greeting (AC: #3)
  - [ ] 4.1: Create `internal/tasks/greeting_insights.go` with `FormatMultiDimensionalGreeting(analyzer *PatternAnalyzer, counter *CompletionCounter) string`
  - [ ] 4.2: Format: compact single line showing tasks today vs yesterday, streak, mood trend
  - [ ] 4.3: Encouraging messaging regardless of direction — never guilt or shame
  - [ ] 4.4: Integrate into `DoorsView` greeting area — insert between `greetingStyle.Render(dv.greeting)` and the `"\n\n"` that follows it in `View()` method (~line 102 of doors_view.go)
  - [ ] 4.5: DoorsView needs access to BOTH PatternAnalyzer AND CompletionCounter — add both as fields or use a setter method. Current `NewDoorsView(pool, tracker)` signature will need updating.
  - [ ] 4.6: Omit the line entirely if no previous session data exists (graceful degradation)

- [ ] Task 5: Wire everything together in MainModel (AC: #1, #2, #3)
  - [ ] 5.1: Add `patternAnalyzer *tasks.PatternAnalyzer` field to MainModel
  - [ ] 5.2: Initialize PatternAnalyzer in `NewMainModel()` by loading from sessions.jsonl path
  - [ ] 5.3: Add `ShowInsightsMsg` handling in MainModel.Update() to switch to ViewInsights and create InsightsView
  - [ ] 5.4: Pass both PatternAnalyzer AND CompletionCounter to DoorsView for multi-dimensional greeting
  - [ ] 5.5: Handle `ReturnToDoorsMsg` when in ViewInsights (already handled for other views — just add ViewInsights to the existing switch case). Do NOT create a separate `ReturnFromInsightsMsg`.

- [ ] Task 6: Write comprehensive tests (AC: all)
  - [ ] 6.1: Create `internal/tasks/pattern_analyzer_test.go` — test all analyzer methods
  - [ ] 6.2: Test cold start guard (< 3 sessions)
  - [ ] 6.3: Test mood correlation calculations with known data
  - [ ] 6.4: Test door position preference percentages and bias detection
  - [ ] 6.5: Test daily completion aggregation for last 7 days
  - [ ] 6.6: Test week-over-week comparison with known data
  - [ ] 6.7: Test sparkline rendering with various data ranges
  - [ ] 6.8: Test greeting_insights formatting with various scenarios
  - [ ] 6.9: Test insights view rendering (snapshot or content checks)
  - [ ] 6.10: Test `:dashboard` and `:insights` command routing
  - [ ] 6.11: Test empty sessions.jsonl handling
  - [ ] 6.12: Test malformed session entries are skipped gracefully

## Dev Notes

### Critical Existing Infrastructure to Reuse

**Session Metrics (internal/tasks/session_tracker.go) — with actual json tags:**
```go
type SessionMetrics struct {
    SessionID           string                `json:"session_id"`
    StartTime           time.Time             `json:"start_time"`
    EndTime             time.Time             `json:"end_time"`
    DurationSeconds     float64               `json:"duration_seconds"`
    TasksCompleted      int                   `json:"tasks_completed"`
    DoorsViewed         int                   `json:"doors_viewed"`
    RefreshesUsed       int                   `json:"refreshes_used"`
    DetailViews         int                   `json:"detail_views"`
    NotesAdded          int                   `json:"notes_added"`
    StatusChanges       int                   `json:"status_changes"`
    MoodEntryCount      int                   `json:"mood_entries"`           // WARNING: tag is "mood_entries" not "mood_entry_count"
    TimeToFirstDoorSecs float64               `json:"time_to_first_door_seconds"` // WARNING: "seconds" not "secs"
    DoorSelections      []DoorSelectionRecord `json:"door_selections,omitempty"`
    TaskBypasses        [][]string            `json:"task_bypasses,omitempty"`
    MoodEntries         []MoodEntry           `json:"mood_entries_detail,omitempty"` // WARNING: "mood_entries_detail" not "mood_entries"
    DoorFeedback        []DoorFeedbackEntry   `json:"door_feedback,omitempty"`
    DoorFeedbackCount   int                   `json:"door_feedback_count"`
}
```

**DoorSelectionRecord:** `{DoorPosition int, TaskText string, Timestamp time.Time}`
- DoorPosition: 0=left, 1=center, 2=right

**MoodEntry:** `{Timestamp time.Time, Mood string, CustomText string}`
- Mood values: "Focused", "Tired", "Stressed", "Energized", "Distracted", "Calm", "Other"

**MetricsWriter (internal/tasks/metrics_writer.go):**
- Writes to `~/.threedoors/sessions.jsonl` — JSON Lines format, one session per line
- Called on app exit via `tracker.Finalize()` then `writer.AppendSession()`

**CompletionCounter (internal/tasks/completion_counter.go):**
- Already has: `GetTodayCount()`, `GetYesterdayCount()`, `GetStreak()`, `FormatCompletionMessage()`
- Loads from `completed.txt`, groups by date
- Has `nowFunc` injection for testing

**Command Infrastructure (internal/tui/search_view.go):**
- `parseCommand(input)` — strips `:`, splits on first space, lowercases
- Existing commands: `:add`, `:mood`, `:goals`, `:stats`, `:health`, `:help`, `:quit`
- Commands trigger flash messages or view switches via Bubbletea messages
- Add `:dashboard` and `:insights` following the same pattern as `:health`

**View Modes (internal/tui/main_model.go) — complete iota enum:**
```go
const (
    ViewDoors       ViewMode = iota // 0
    ViewDetail                      // 1
    ViewMood                        // 2
    ViewSearch                      // 3
    ViewHealth                      // 4
    ViewAddTask                     // 5
    ViewValuesGoals                 // 6
    ViewFeedback                    // 7
    ViewImprovement                 // 8
    ViewNextSteps                   // 9
    ViewInsights                    // 10 — ADD THIS
)
```

### Architecture & Pattern Compliance

- **MVU Pattern:** All state changes via Bubbletea messages. InsightsView is a sub-view (NOT a full `tea.Model`). Follow HealthView pattern: implement only `Update(msg tea.Msg) tea.Cmd` and `View() string`. Do NOT implement `Init()`. MainModel manages the lifecycle.
- **Domain-layer separation:** PatternAnalyzer lives in `internal/tasks/` (domain), InsightsView lives in `internal/tui/` (presentation)
- **Constructor injection:** PatternAnalyzer created in `NewMainModel()`, passed to views
- **Error handling:** `LoadSessions()` returns `nil` error for missing file (`os.IsNotExist`), returns wrapped error for permission denied. Malformed JSON lines are skipped silently (no error). Empty file (0 bytes) returns nil error with empty sessions. When called before `LoadSessions()`, all computation methods return zero-value results (empty slices, zero floats, `false` booleans).
- **Constructor pattern:** Create `PatternAnalyzer` INSIDE `NewMainModel()` (same pattern as `CompletionCounter` at lines 70-75 of main_model.go). Do NOT add it as a constructor parameter — that would change the signature and impact `main.go`.
- **Command wiring:** Do NOT pass `PatternAnalyzer` to `SearchView`. Instead, `:dashboard`/`:insights` commands send a `ShowInsightsMsg{}` message (no data). MainModel handles it by creating InsightsView with its own analyzer reference. This follows the `ShowMoodMsg` pattern.
- **NewMainModel current signature:** `func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, provider tasks.TaskProvider, hc *tasks.HealthChecker) *MainModel` — do not change this.
- **NewDoorsView current signature:** `func NewDoorsView(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *DoorsView` — add PatternAnalyzer and CompletionCounter as additional params or use setter methods after construction.
- **gofumpt formatting:** All code must pass `gofumpt`
- **No panics:** Use error returns
- **Existing styles:** Use styles from `internal/tui/styles.go` — `flashStyle`, `colorComplete`, `colorGreeting`, door colors for position preference display

### Data Flow

```
sessions.jsonl → PatternAnalyzer.LoadSessions() → in-memory analysis
completed.txt  → CompletionCounter (existing)   → streak + daily counts

PatternAnalyzer + CompletionCounter → InsightsView.View() → full dashboard
PatternAnalyzer + CompletionCounter → FormatMultiDimensionalGreeting() → DoorsView greeting
```

### sessions.jsonl Format

Each line is a JSON object matching the `SessionMetrics` struct. **CRITICAL: All JSON fields use explicit snake_case json tags, NOT PascalCase.**

```json
{"session_id":"uuid","start_time":"2026-03-01T10:00:00Z","end_time":"2026-03-01T10:30:00Z","duration_seconds":1800,"tasks_completed":3,"doors_viewed":8,"refreshes_used":5,"detail_views":4,"notes_added":1,"status_changes":2,"mood_entries":1,"time_to_first_door_seconds":12.5,"door_selections":[{"door_position":0,"task_text":"task1","timestamp":"..."},{"door_position":2,"task_text":"task2","timestamp":"..."}],"task_bypasses":[["bypassed1","bypassed2"]],"mood_entries_detail":[{"timestamp":"...","mood":"Focused","custom_text":""}],"door_feedback":[],"door_feedback_count":0}
```

**CRITICAL JSON Tag Gotchas (from session_tracker.go):**

| Go Field Name | JSON Tag | WARNING |
|---|---|---|
| `MoodEntryCount` | `"mood_entries"` | NOT `"mood_entry_count"` — confusing name! |
| `MoodEntries` (slice) | `"mood_entries_detail"` | NOT `"mood_entries"` — that's the count field! |
| `TimeToFirstDoorSecs` | `"time_to_first_door_seconds"` | NOT `"time_to_first_door_secs"` |
| `DoorSelectionRecord.DoorPosition` | `"door_position"` | snake_case, not PascalCase |
| `DoorSelectionRecord.TaskText` | `"task_text"` | snake_case |
| `MoodEntry.CustomText` | `"custom_text"` | snake_case, omitempty |

The `LoadSessions()` function uses `json.Unmarshal` which matches against these json tags. Using wrong field names will result in zero values.

### PatternAnalyzer Data Structures

```go
type PatternAnalyzer struct {
    sessions []SessionMetrics
    nowFunc  func() time.Time // for testability
}

// Constructors (follow CompletionCounter pattern)
func NewPatternAnalyzer() *PatternAnalyzer
func NewPatternAnalyzerWithNow(nowFunc func() time.Time) *PatternAnalyzer

type MoodCorrelation struct {
    Mood              string
    SessionCount      int     // number of sessions with this mood
    AvgTasksCompleted float64 // average tasks completed per session
}

type DoorPreferences struct {
    LeftPercent     float64 // rounded to 1 decimal place
    CenterPercent   float64 // rounded to 1 decimal place
    RightPercent    float64 // rounded to 1 decimal place
    TotalSelections int
    HasBias         bool   // true if any position > 50%
    BiasPosition    string // "left", "center", or "right" (lowercase) if HasBias
}

type WeekComparison struct {
    ThisWeekTotal int
    LastWeekTotal int
    PercentChange float64 // positive = improvement; if LastWeekTotal == 0 and ThisWeekTotal > 0, use 100.0; if both 0, use 0.0
    Direction     string  // "up", "down", "same"
}
```

### Critical Computation Rules (for test assertions)

**Week boundaries:** Weeks run Monday-Sunday (ISO 8601). "This week" = Monday of the current week through today. "Last week" = the full 7 days of the previous Monday-Sunday.

**Date grouping for daily completions:** Use `StartTime` of each session (not `EndTime`) to determine which day the session's completions count toward. Format keys as `"2006-01-02"` (Go reference date). Days with zero sessions ARE present in the map with value `0` when they fall within the requested range.

**Multi-mood sessions:** If a session has multiple mood entries (e.g., Focused and Tired), each mood is attributed the FULL `TasksCompleted` count for that session. This means a session with 5 completions and 2 moods contributes `5` to both Focused and Tired correlations. Rationale: mood was present during the session, so the full productivity is associated.

**Sessions with no mood entries:** Excluded entirely from mood correlation calculations. They do not appear as a "No Mood" category.

**Sessions with no door selections:** Excluded from door preference calculations. `TotalSelections` only counts sessions that have `len(DoorSelections) > 0`.

**Percentage rounding:** Door preference percentages are rounded to 1 decimal place using `math.Round(pct*10)/10`. They may not sum to exactly 100% due to rounding — this is acceptable.

**Cold start threshold:** `HasSufficientData()` returns `true` when `len(sessions) >= 3`. `GetSessionsNeeded()` returns `max(0, 3 - len(sessions))`.

**Thread safety:** Not required. PatternAnalyzer is only called from the Bubbletea main loop (single-threaded).

### InsightsView Struct Definition

```go
type InsightsView struct {
    analyzer *tasks.PatternAnalyzer
    counter  *tasks.CompletionCounter
    width    int // terminal width, set via SetWidth()
    scroll   int // scroll offset for long content (lines scrolled from top)
}

func NewInsightsView(analyzer *tasks.PatternAnalyzer, counter *tasks.CompletionCounter) *InsightsView
func (iv *InsightsView) SetWidth(w int) // required — HealthView and SearchView both have this
func (iv *InsightsView) Update(msg tea.Msg) tea.Cmd
func (iv *InsightsView) View() string
```

Section header strings (exact, for test assertions with `strings.Contains`):
- `"COMPLETION TRENDS (Last 7 Days)"`
- `"STREAKS"`
- `"MOOD & PRODUCTIVITY"`
- `"DOOR POSITION PREFERENCES"`
- `"Your Insights Dashboard"` (main header)
- `"Press Esc to return"` (footer)
- `"Keep using ThreeDoors to unlock insights!"` (cold start, followed by `"(X more sessions needed)"`)

### InsightsView Layout

```
┌─────────────────────────────────────────────────────┐
│  📊 Your Insights Dashboard                         │
├─────────────────────────────────────────────────────┤
│                                                     │
│  COMPLETION TRENDS (Last 7 Days)                    │
│  Mon  Tue  Wed  Thu  Fri  Sat  Sun                  │
│   ▃    ▅    █    ▂    ▄    ▆    ▁                   │
│   2    3    5    1    3    4    0                     │
│                                                     │
│  This week: 18  |  Last week: 14  |  ↑ 28%          │
│                                                     │
│  STREAKS                                            │
│  Current streak: 5 days 🔥                          │
│                                                     │
│  MOOD & PRODUCTIVITY                                │
│  Focused:    avg 4.2 tasks/session (8 sessions)     │
│  Energized:  avg 3.8 tasks/session (5 sessions)     │
│  Calm:       avg 3.1 tasks/session (4 sessions)     │
│  Tired:      avg 1.5 tasks/session (6 sessions)     │
│  Most productive mood: Focused                      │
│  Most frequent mood: Focused                        │
│                                                     │
│  DOOR POSITION PREFERENCES                          │
│  Left: 35%  |  Center: 42%  |  Right: 23%           │
│                                                     │
│  Press Esc to return                                 │
└─────────────────────────────────────────────────────┘
```

### "Better Than Yesterday" Greeting Format — Exact Specification

`FormatMultiDimensionalGreeting` returns a **deterministic** string (no randomness). The format is structured with `|` delimiters. No prose-style encouraging suffixes — the encouragement comes from the existing random greeting message above it.

**Format template (all cases use this structure):**

Case 1 — Today has completions AND yesterday data exists:
```
📈 Today: {today} | Yesterday: {yesterday} | Streak: {streak} days | Mood: {lastMood}
```
Example: `📈 Today: 3 | Yesterday: 5 | Streak: 5 days | Mood: Focused`

Case 2 — Today has zero completions, yesterday data exists:
```
📈 Yesterday: {yesterday} | Streak: {streak} days | Mood: {lastMood}
```
Example: `📈 Yesterday: 4 | Streak: 5 days | Mood: Energized`

Case 3 — No mood data available (omit Mood segment):
```
📈 Today: {today} | Yesterday: {yesterday} | Streak: {streak} days
```

Case 4 — No previous session data at all (empty sessions.jsonl):
Returns `""` (empty string — line omitted entirely from DoorsView)

Case 5 — Streak is 0:
Omit "Streak" segment from the line.

**"Mood" value:** Use the most recent mood entry from the most recent session that has mood entries. This is `GetMostRecentMood() string` — a new method returning the `Mood` field of the last `MoodEntry` from the most recent session with `len(MoodEntries) > 0`. Returns `""` if no mood data exists.

**"Streak" value:** Use `CompletionCounter.GetStreak()` (existing method).

**"Today"/"Yesterday" values:** Use `CompletionCounter.GetTodayCount()` and `GetYesterdayCount()` (existing methods).

### Sparkline Rendering

Use Unicode block characters for visual completion trend:
```go
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func sparkline(values []int) string {
    max := maxOf(values)
    if max == 0 {
        return strings.Repeat(string(sparkChars[0]), len(values))
    }
    var result strings.Builder
    for _, v := range values {
        idx := int(float64(v) / float64(max) * float64(len(sparkChars)-1))
        result.WriteRune(sparkChars[idx])
    }
    return result.String()
}
```

### File Structure Requirements

Files to create:
- `internal/tasks/pattern_analyzer.go` — PatternAnalyzer service
- `internal/tasks/pattern_analyzer_test.go` — Tests
- `internal/tasks/greeting_insights.go` — Multi-dimensional greeting formatter
- `internal/tasks/greeting_insights_test.go` — Tests
- `internal/tui/insights_view.go` — InsightsView TUI component
- `internal/tui/insights_view_test.go` — Tests

Files to modify:
- `internal/tui/main_model.go` — Add ViewInsights mode, PatternAnalyzer field, message routing
- `internal/tui/search_view.go` — Add `:dashboard` and `:insights` command handling
- `internal/tui/doors_view.go` — Add multi-dimensional greeting line
- `internal/tui/messages.go` — Add `ShowInsightsMsg` only (reuse existing `ReturnToDoorsMsg` for Esc handling)

### Testing Requirements

**Test Files:**
- `internal/tasks/pattern_analyzer_test.go` (NEW) — All analyzer logic tests
- `internal/tasks/greeting_insights_test.go` (NEW) — Greeting format tests
- `internal/tui/insights_view_test.go` (NEW) — View rendering tests

**Test Helper for Creating Test Sessions:**
```go
func makeTestSession(startTime time.Time, completed int, moods []string, doorPositions []int) SessionMetrics {
    entries := make([]MoodEntry, len(moods))
    for i, m := range moods {
        entries[i] = MoodEntry{Mood: m, Timestamp: startTime}
    }
    selections := make([]DoorSelectionRecord, len(doorPositions))
    for i, p := range doorPositions {
        selections[i] = DoorSelectionRecord{DoorPosition: p, TaskText: fmt.Sprintf("task-%d", i), Timestamp: startTime}
    }
    return SessionMetrics{
        SessionID:      uuid.New().String(),
        StartTime:      startTime,
        EndTime:        startTime.Add(30 * time.Minute),
        DurationSeconds: 1800,
        TasksCompleted: completed,
        MoodEntries:    entries,
        MoodEntryCount: len(moods),
        DoorSelections: selections,
    }
}
```

**IMPORTANT:** `startTime` parameter is required — almost all tests need sessions at specific dates for daily completions and week-over-week assertions.

**Test Helper for Writing sessions.jsonl Fixture Files:**
```go
func writeSessionsFile(t *testing.T, dir string, sessions []SessionMetrics) string {
    t.Helper()
    path := filepath.Join(dir, "sessions.jsonl")
    var buf bytes.Buffer
    for _, s := range sessions {
        data, err := json.Marshal(s)
        require.NoError(t, err) // or use t.Fatal if no testify
        buf.Write(data)
        buf.WriteByte('\n')
    }
    err := os.WriteFile(path, buf.Bytes(), 0o644)
    require.NoError(t, err)
    return path
}
```

Note: Use Go standard `testing` package (no testify). Replace `require.NoError` with:
```go
if err != nil { t.Fatalf("writeSessionsFile: %v", err) }
```

**Test time freezing:** Use `nowFunc func() time.Time` field in PatternAnalyzer for consistent date calculations in tests.

**BDD Test Scenarios:**

| # | Given | When | Then |
|---|-------|------|------|
| T1 | 5 sessions with various completions and moods | GetMoodCorrelations() called | Returns mood -> avg completions, sorted by productivity |
| T2 | < 3 sessions | HasSufficientData() called | Returns false |
| T3 | 10 sessions, 7 with door selections | GetDoorPositionPreferences() | Returns percentages summing to 100%, bias detected if >50% |
| T4 | Sessions spanning 2 weeks | GetWeekOverWeek() called | Returns correct totals and percentage change |
| T5 | 7 days of completion data | GetDailyCompletions(7) | Returns correct counts for each day |
| T6 | No sessions.jsonl file | LoadSessions() called | Returns nil error, empty data, HasSufficientData()=false |
| T7 | Malformed JSON lines mixed with valid | LoadSessions() called | Valid sessions loaded, malformed skipped |
| T8 | Session with multiple moods | Correlation calc | Each mood counted separately for its correlation |
| T9 | All door selections are center | Bias detection | HasBias=true, BiasPosition="center" |
| T10 | Equal door distribution (33/33/34) | Bias detection | HasBias=false |
| T11 | Today=3, Yesterday=5, Streak=2 | FormatMultiDimensionalGreeting() | Returns encouraging line with all dimensions |
| T12 | No previous data | FormatMultiDimensionalGreeting() | Returns empty string |
| T13 | User types `:dashboard` | Command parsed | ShowInsightsMsg sent |
| T14 | User types `:insights` | Command parsed | ShowInsightsMsg sent |
| T15 | Exactly 2 sessions per mood | Correlation with min sample | Correlations shown (2 >= minimum of 2) |

### Library & Framework Requirements

- **Go standard library only** for PatternAnalyzer — `encoding/json`, `os`, `bufio`, `time`, `sort`, `fmt`, `math`, `strings`
- **Bubbletea** — for InsightsView component
- **Lipgloss** — for styling the dashboard
- **No new dependencies** — everything needed is already in go.mod

### Previous Story Intelligence

From Story 3.5 (Daily Completion Tracking):
- CompletionCounter pattern: domain service with `nowFunc` injection, loads from file, in-memory operations
- Follow same pattern for PatternAnalyzer
- Flash message and view integration patterns

From Story 3.7 (Enhanced Navigation & Messaging):
- Navigation patterns between views
- Message routing in MainModel
- View mode enum additions

From Story 2.6 (Health Check Command):
- `:health` command pattern — triggers `HealthCheckMsg`, switches to `ViewHealth`, renders results, Esc returns
- Follow identical pattern for `:dashboard`/`:insights` → `ShowInsightsMsg` → `ViewInsights` → render → Esc returns

### Git Intelligence

Recent commits show consistent patterns:
- File naming: `snake_case.go` and `snake_case_test.go`
- All code passes `gofumpt` and `golangci-lint run ./...`
- Table-driven tests with descriptive test case names
- Constructor injection for dependencies
- Error handling: return errors, don't panic

### References

- [Source: internal/tasks/session_tracker.go] — SessionMetrics struct, DoorSelectionRecord, MoodEntry
- [Source: internal/tasks/metrics_writer.go] — sessions.jsonl write format
- [Source: internal/tasks/completion_counter.go] — CompletionCounter pattern to follow
- [Source: internal/tui/search_view.go] — Command parsing and routing
- [Source: internal/tui/main_model.go] — View mode enum, message routing
- [Source: internal/tui/health_view.go] — View pattern to follow for InsightsView
- [Source: internal/tui/doors_view.go] — Greeting rendering, integration point
- [Source: internal/tui/styles.go] — Existing style constants
- [Source: internal/tui/messages.go] — Message type definitions
- [Source: docs/architecture/data-storage-schema.md] — sessions.jsonl schema
- [Source: docs/architecture/coding-standards.md] — Code standards
- [Source: docs/architecture/test-strategy-and-standards.md] — Test standards
- [Source: docs/prd/epics-and-stories.md#Story4.4-4.6] — Related epic stories

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
