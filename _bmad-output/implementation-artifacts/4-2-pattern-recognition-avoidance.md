# Story 4.2: Session Metrics Pattern Analysis & Avoidance Detection

Status: ready-for-dev

## Story

As a developer,
I want to analyze historical session metrics for user behavior patterns,
So that the learning engine has data to work with and users gain insight into their habits.

## Acceptance Criteria

1. **Given** accumulated session metrics in sessions.jsonl, **When** the pattern analyzer runs, **Then** it identifies: door position preferences (left/center/right bias), task type selection vs bypass rates, time-of-day patterns, mood-task correlation coefficients, and avoidance patterns (tasks shown 3+ times without selection)

2. **Given** pattern analysis results, **When** analysis completes, **Then** results are stored in a patterns cache file (patterns.json) in `~/.threedoors/`

3. **Given** the app starts up, **When** sufficient session data exists, **Then** analysis runs on app startup (background, non-blocking) using a goroutine

4. **Given** fewer than 5 sessions in sessions.jsonl, **When** pattern analysis is triggered, **Then** it returns early with no patterns generated (cold start guard)

5. **Given** a task has been shown in doors 5+ times without selection (tracked via TaskBypasses in sessions.jsonl), **When** that task appears in doors again, **Then** a subtle indicator appears (e.g., "Seen 7 times") below the task text on the door card

6. **Given** a user types `:insights`, **When** the command executes, **Then** it shows a summary of patterns: door position preference, most selected task types, most bypassed task types, avoidance list (tasks bypassed 5+ times), and mood-task correlations (e.g., "When stressed, you tend to select quick-win tasks")

## Tasks / Subtasks

- [ ] Task 1: Create pattern analyzer engine (AC: #1, #4)
  - [ ] 1.1: Create `internal/tasks/pattern_analyzer.go` with `PatternAnalyzer` struct
  - [ ] 1.2: Implement `ReadSessions(path string) ([]SessionMetrics, error)` — read sessions.jsonl, parse each line as JSON into `SessionMetrics`. Use `bufio.Scanner` + `json.Unmarshal`. Handle empty/missing file gracefully (return empty slice, nil error).
  - [ ] 1.3: Implement cold start guard: `func (pa *PatternAnalyzer) Analyze(sessions []SessionMetrics) (*PatternReport, error)` — if `len(sessions) < 5`, return nil PatternReport with no error
  - [ ] 1.4: Define `PatternReport` struct:
    ```go
    type PatternReport struct {
        GeneratedAt         time.Time                    `json:"generated_at"`
        SessionCount        int                          `json:"session_count"`
        DoorPositionBias    DoorPositionStats            `json:"door_position_bias"`
        TaskTypeStats       map[string]TypeSelectionRate `json:"task_type_stats"`
        TimeOfDayPatterns   []TimeOfDayPattern           `json:"time_of_day_patterns"`
        MoodCorrelations    []MoodCorrelation            `json:"mood_correlations"`
        AvoidanceList       []AvoidanceEntry             `json:"avoidance_list"`
    }
    ```
  - [ ] 1.5: Define supporting structs:
    ```go
    type DoorPositionStats struct {
        LeftCount    int     `json:"left_count"`
        CenterCount  int     `json:"center_count"`
        RightCount   int     `json:"right_count"`
        TotalSelections int  `json:"total_selections"`
        PreferredPosition string `json:"preferred_position"` // "left", "center", "right", or "none"
    }

    type TypeSelectionRate struct {
        TimesShown    int     `json:"times_shown"`
        TimesSelected int     `json:"times_selected"`
        TimesBypassed int     `json:"times_bypassed"`
        SelectionRate float64 `json:"selection_rate"` // selected/shown
    }

    type TimeOfDayPattern struct {
        Period           string  `json:"period"` // "morning", "afternoon", "evening", "night"
        SessionCount     int     `json:"session_count"`
        AvgTasksCompleted float64 `json:"avg_tasks_completed"`
        AvgDuration      float64 `json:"avg_duration_minutes"`
    }

    type MoodCorrelation struct {
        Mood              string  `json:"mood"`
        SessionCount      int     `json:"session_count"`
        PreferredType     string  `json:"preferred_type"`     // most selected TaskType
        PreferredEffort   string  `json:"preferred_effort"`   // most selected TaskEffort
        AvgTasksCompleted float64 `json:"avg_tasks_completed"`
    }

    type AvoidanceEntry struct {
        TaskText     string `json:"task_text"`
        TimesBypassed int   `json:"times_bypassed"`
        TimesShown   int    `json:"times_shown"`
        NeverSelected bool  `json:"never_selected"`
    }
    ```

- [ ] Task 2: Implement each analysis dimension (AC: #1)
  - [ ] 2.1: **Door position bias** — iterate all `DoorSelections` across sessions, count position 0/1/2 frequency. Calculate preferred position (>40% of selections = bias, else "none").
  - [ ] 2.2: **Task type selection vs bypass rates** — for each session, cross-reference `DoorSelections` (selected tasks) with `TaskBypasses` (bypassed tasks). Since `DoorSelections` records `TaskText` and `TaskBypasses` records task text arrays, match by text. Categorization data (Type/Effort/Location) is on the Task struct — the analyzer needs access to the current task pool OR must infer type from the task text. **Decision:** Store task type in DoorSelectionRecord in future, but for now, match task text against current pool to look up type. If task not found in pool (deleted/completed), skip it.
  - [ ] 2.3: **Time-of-day patterns** — group sessions by StartTime hour: morning (5-11), afternoon (12-16), evening (17-20), night (21-4). Calculate avg tasks completed and avg duration per period.
  - [ ] 2.4: **Mood-task correlations** — for sessions with MoodEntries, correlate the mood text with which TaskTypes were selected in that session. Group by mood, find most-selected type and effort. Requires minimum 3 sessions with same mood to report a correlation.
  - [ ] 2.5: **Avoidance detection** — build a map of task text → bypass count by iterating all TaskBypasses arrays across all sessions. Any task bypassed 3+ times (per AC) is an avoidance candidate. Include in AvoidanceList.

- [ ] Task 3: Implement patterns cache persistence (AC: #2)
  - [ ] 3.1: Implement `SavePatterns(report *PatternReport, path string) error` — marshal to JSON (indented for readability), write atomically to `~/.threedoors/patterns.json` (same atomic write pattern as file_manager.go: write to .tmp, rename).
  - [ ] 3.2: Implement `LoadPatterns(path string) (*PatternReport, error)` — read patterns.json, unmarshal. Handle missing file (return nil, nil). Handle corrupt file (return nil, error).
  - [ ] 3.3: Cache invalidation: if patterns.json `GeneratedAt` is older than the latest session in sessions.jsonl, re-analyze. Compare `GeneratedAt` against last session's `EndTime`.

- [ ] Task 4: Integrate analysis into app startup (AC: #3)
  - [ ] 4.1: In `cmd/threedoors/main.go`, after creating the SessionTracker (line 49), launch pattern analysis in a goroutine. The goroutine should also load the result into a shared pointer for the TUI to use:
    ```go
    var patternsReport atomic.Pointer[tasks.PatternReport]
    go func() {
        analyzer := tasks.NewPatternAnalyzer()
        configDir, _ := tasks.GetConfigDirPath()
        sessionsPath := filepath.Join(configDir, "sessions.jsonl")
        patternsPath := filepath.Join(configDir, "patterns.json")

        // Try loading cached patterns first
        cached, _ := analyzer.LoadPatterns(patternsPath)

        // Check if re-analysis needed
        sessions, err := analyzer.ReadSessions(sessionsPath)
        if err != nil { return }

        if cached != nil && !analyzer.NeedsReanalysis(cached, sessions) {
            patternsReport.Store(cached)
            return
        }

        report, err := analyzer.Analyze(sessions)
        if err != nil || report == nil { return }
        _ = analyzer.SavePatterns(report, patternsPath)
        patternsReport.Store(report)
    }()
    ```
  - [ ] 4.2: Pass `&patternsReport` to `NewMainModel()` so the TUI can access patterns. Add `patterns *atomic.Pointer[tasks.PatternReport]` field to MainModel. This is thread-safe for concurrent read/write.
  - [ ] 4.3: The goroutine must not block the main TUI. Errors silently ignored (same pattern as metrics writer).
  - [ ] 4.4: Implement `NeedsReanalysis(cached *PatternReport, sessions []SessionMetrics) bool` — returns true if cached is nil, or if `len(sessions) > cached.SessionCount`, or if last session EndTime > cached.GeneratedAt.

- [ ] Task 5: Implement avoidance indicator on door cards (AC: #5)
  - [ ] 5.1: Add `patterns *atomic.Pointer[tasks.PatternReport]` field to `DoorsView` struct. Pass from MainModel during construction.
  - [ ] 5.2: Create helper `func (dv *DoorsView) getBypassCount(taskText string) int` — safely loads PatternReport from atomic pointer, searches AvoidanceList for matching task text, returns bypass count (0 if not found or patterns nil).
  - [ ] 5.3: In door card rendering in `doors_view.go` (after the `categoryBadge` call around line 160), if `getBypassCount(task.Text) >= 5`, append a subtle dimmed line using lipgloss: `lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Seen %d times", count))`.
  - [ ] 5.4: The indicator must be non-judgmental — purely informational, no guilt language. Use "Seen" not "Avoided" or "Skipped".

- [ ] Task 6: Implement `:insights` command (AC: #6)
  - [ ] 6.1: Add `case "insights":` in `search_view.go` `executeCommand()` (around line 163, after existing commands). Pass patterns pointer from SearchView field.
  - [ ] 6.2: Add `patterns *atomic.Pointer[tasks.PatternReport]` field to `SearchView` struct. Update `NewSearchView()` constructor to accept it. Update ALL call sites in main_model.go.
  - [ ] 6.3: Read patterns from the atomic pointer (same data the goroutine populated). If nil, show "Not enough data yet — need at least 5 sessions for insights."
  - [ ] 6.4: Create `internal/tasks/insights_formatter.go` with `FormatInsights(report *PatternReport) string` — pure function that formats the multi-line insights text. This keeps formatting logic testable outside the TUI:
    ```
    Session Insights (N sessions analyzed)

    Door Preference: You tend to pick the [left/center/right] door (X%)

    Task Types: Most selected: [type] (X%) | Most bypassed: [type] (X%)

    Best Time: You complete the most tasks in the [morning/afternoon/evening] (avg X/session)

    Mood Patterns:
    - When [mood]: prefer [type] tasks, avg [N] completed

    Avoidance Alert: [N] tasks bypassed 5+ times
    - "[task text]" (bypassed N times)
    ```
  - [ ] 6.5: Display insights via `FlashMsg` for short output (< 5 lines) or create an `InsightsViewMsg` that shows a scrollable detail view (reuse existing detail view pattern from `detail_view.go`). **Decision:** Use a dedicated view for rich multi-line output — create `insights_view.go` following the same Model/Init/Update/View pattern as other views, or simply render as a formatted string in the existing detail view area.
  - [ ] 6.6: If no patterns available, the message should encourage continued use: "Keep going! After 5 sessions, you'll start seeing patterns."
  - [ ] 6.7: Handle edge cases in formatting: no mood data → skip "Mood Patterns" section; no avoidance → skip "Avoidance Alert" section; no door preference → show "No strong door preference detected"

- [ ] Task 7: Write unit tests (AC: #1-6)
  - [ ] 7.1: Create `internal/tasks/pattern_analyzer_test.go` with table-driven tests
  - [ ] 7.2: Test `ReadSessions()`:
    - Valid sessions.jsonl with 3 entries → returns 3 SessionMetrics
    - Empty file → returns empty slice, nil error
    - Missing file → returns empty slice, nil error
    - Malformed JSON line → skips line, continues parsing
  - [ ] 7.3: Test `Analyze()` cold start guard:
    - 4 sessions → returns nil PatternReport
    - 5 sessions → returns valid PatternReport
  - [ ] 7.4: Test door position bias:
    - 10 selections all position 0 → preferred "left" with 100%
    - Even distribution → preferred "none"
    - 6 left, 2 center, 2 right → preferred "left"
  - [ ] 7.5: Test time-of-day patterns:
    - Sessions at 9am, 10am, 11am → all "morning" period
    - Sessions at 9am, 2pm, 8pm → morning, afternoon, evening
  - [ ] 7.6: Test avoidance detection:
    - Task "Buy groceries" bypassed in 3 sessions → appears in AvoidanceList with count 3
    - Task bypassed only once → not in list
  - [ ] 7.7: Test mood correlations:
    - 3 sessions with mood "focused" selecting technical tasks → correlation: focused → technical
    - Only 1 session with mood "tired" → no correlation reported (minimum 3)
  - [ ] 7.8: Test `SavePatterns()`/`LoadPatterns()` round-trip
  - [ ] 7.9: Test cache invalidation: old patterns.json + newer sessions → re-analyze
  - [ ] 7.10: Test `:insights` command formatting with sample PatternReport

## Dev Notes

### Architecture & Patterns

- **Location:** All new code in `internal/tasks/` package. New files: `pattern_analyzer.go`, `pattern_analyzer_test.go`
- **Pattern:** Follow `completion_counter.go` / `health_checker.go` — struct with constructor, methods on receiver
- **Session data:** Read from `~/.threedoors/sessions.jsonl` (JSON Lines format). Each line is a `SessionMetrics` struct (defined in `session_tracker.go:31-50`)
- **Key data fields for analysis:**
  - `DoorSelections []DoorSelectionRecord` — position (0/1/2), task text, timestamp
  - `TaskBypasses [][]string` — arrays of bypassed task texts per refresh
  - `MoodEntries []MoodEntry` — mood text + timestamp
  - `StartTime` / `EndTime` — for time-of-day analysis
  - `TasksCompleted` — for productivity correlation
- **Persistence:** patterns.json uses same atomic write pattern as file_manager.go (write .tmp, rename)
- **Startup integration:** Fire-and-forget goroutine in `cmd/threedoors/main.go` after line 49 (after tracker creation)

### Critical Constraints

- **No external dependencies:** Pure Go. Use `encoding/json`, `bufio`, `os`, `math`, `time`, `sort` from stdlib.
- **Non-blocking:** Pattern analysis MUST run in a goroutine. No impact on TUI startup time.
- **Offline-first:** Everything local (NFR4, NFR14).
- **Backward compatible:** Missing sessions.jsonl or patterns.json is not an error.
- **Performance:** Analysis of ~100 sessions should complete in <100ms.
- **Cold start guard:** Minimum 5 sessions before generating any patterns.

### Existing Code References

- **SessionMetrics struct:** `internal/tasks/session_tracker.go:31-50`
- **DoorSelectionRecord:** `internal/tasks/session_tracker.go:24-29` (Timestamp, DoorPosition int, TaskText string)
- **TaskBypasses:** `[][]string` — each inner array = task texts shown and bypassed during one refresh
- **MoodEntry:** `internal/tasks/session_tracker.go:17-22` (Timestamp, Mood string, CustomText string)
- **MetricsWriter.AppendSession:** `internal/tasks/metrics_writer.go:26-43` — writes to sessions.jsonl
- **GetConfigDirPath:** `internal/tasks/file_manager.go:34-46` → `~/.threedoors/`
- **TaskType/TaskEffort/TaskLocation:** `internal/tasks/task_categorization.go` (from PR #40)
- **Door card rendering:** `internal/tui/doors_view.go` — `categoryBadge()` at lines 29-44, door rendering around line 160
- **Command handling:** `internal/tui/search_view.go:99` — `executeCommand()` switch
- **FlashMsg:** `internal/tui/messages.go` — for simple feedback
- **Test helpers:** `internal/tasks/test_helpers_test.go` — `newTestTask()`, `newCategorizedTestTask()`, `poolFromTasks()`

### What This Story Does NOT Do

- Does NOT modify the door selection algorithm (that's Story 4.3 — mood-aware selection)
- Does NOT implement goal re-evaluation prompts (that's Story 4.5)
- Does NOT modify SessionTracker or MetricsWriter — only reads existing data
- Does NOT add new session recording fields — works with existing sessions.jsonl format
- Does NOT implement persistent avoidance prompts (10+ bypasses → R/B/D/A prompt, that's Story 4.4 continuation)

### Project Structure Notes

- New files in `internal/tasks/`: `pattern_analyzer.go`, `pattern_analyzer_test.go`, `insights_formatter.go`, `insights_formatter_test.go`
- Modified files: `cmd/threedoors/main.go` (startup goroutine + atomic pointer), `internal/tui/doors_view.go` (avoidance indicator + patterns field), `internal/tui/search_view.go` (`:insights` command + patterns field), `internal/tui/main_model.go` (pass patterns pointer to views, update NewSearchView/DoorsView constructors)
- No new packages or directories
- No new go.mod dependencies (sync/atomic is stdlib)

### Testing Strategy Notes (for tea agent)

- Use **table-driven tests** (Go convention throughout codebase)
- Create test fixtures: sample sessions.jsonl content as string constants in test file
- Test each analysis dimension independently with focused test functions
- Test cold start guard (< 5 sessions)
- Test edge cases: empty sessions, sessions with no door selections, sessions with no mood entries
- Test file I/O: missing file, empty file, corrupt JSON, partial JSON line
- Test patterns.json round-trip persistence
- Use `t.TempDir()` for file I/O tests (same as metrics_writer_test.go)
- Test `FormatInsights()` output formatting with various PatternReport states
- Test `NeedsReanalysis()` logic
- Test thread-safety: concurrent read of atomic.Pointer while goroutine writes
- Run: `go test ./internal/tasks/... ./internal/tui/...`

### Sample Test Fixture (sessions.jsonl format)

Each line in sessions.jsonl is a JSON object matching `SessionMetrics`:
```json
{"session_id":"sess-001","start_time":"2025-11-10T09:00:00Z","end_time":"2025-11-10T09:15:00Z","duration_seconds":900,"tasks_completed":2,"doors_viewed":3,"refreshes_used":1,"detail_views":1,"notes_added":0,"status_changes":1,"mood_entries":1,"time_to_first_door_seconds":2.5,"door_selections":[{"timestamp":"2025-11-10T09:05:00Z","door_position":0,"task_text":"Fix login bug"}],"task_bypasses":[["Buy groceries","Write report"]],"mood_entries_detail":[{"timestamp":"2025-11-10T09:01:00Z","mood":"focused","custom_text":""}],"door_feedback":[],"door_feedback_count":0}
```

Use `makeTestSession(id string, start time.Time, selections []DoorSelectionRecord, bypasses [][]string, moods []MoodEntry) SessionMetrics` helper to build test fixtures programmatically rather than parsing JSON strings.

### Concurrency Safety Notes

- `sync/atomic.Pointer[PatternReport]` is used to share data between the background goroutine and the TUI thread
- The goroutine calls `patternsReport.Store(report)` once analysis is complete
- TUI reads via `patternsReport.Load()` — returns nil before analysis completes (safe)
- No mutex needed — atomic.Pointer provides single-writer, multiple-reader safety
- patterns.json file write from goroutine and file read from `:insights` could race — but since `:insights` reads the atomic pointer (not the file), this is safe. File I/O only happens once at startup.

### References

- [Source: docs/prd/epics-and-stories.md - Epic 4, Story 4.2 lines 630-643]
- [Source: docs/prd/epics-and-stories.md - Epic 4, Story 4.4 lines 660-673 (avoidance detection)]
- [Source: docs/prd/requirements.md - FR20, FR21]
- [Source: internal/tasks/session_tracker.go:31-50 - SessionMetrics struct]
- [Source: internal/tasks/session_tracker.go:24-29 - DoorSelectionRecord]
- [Source: internal/tasks/metrics_writer.go:26-43 - AppendSession]
- [Source: internal/tasks/file_manager.go:34-46 - GetConfigDirPath]
- [Source: internal/tasks/task_categorization.go - TaskType/TaskEffort/TaskLocation]
- [Source: internal/tui/doors_view.go:29-44 - categoryBadge rendering]
- [Source: internal/tui/search_view.go:99 - executeCommand switch]
- [Source: cmd/threedoors/main.go:49 - SessionTracker creation]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
