# Story 4.3: Mood Correlation Analysis

Status: ready-for-dev

## Story

As a user,
I want door selection to consider my current mood and historical patterns,
So that I'm shown tasks that match my current capacity and emotional state.

## Acceptance Criteria

1. **Given** a user has logged a mood entry in the current session (via 'M' key or `:mood` command), **When** doors are selected for display, **Then** the selection algorithm weights tasks based on mood-task correlation data from `patterns.json` (e.g., "stressed" → prefer quick-wins over deep-work)

2. **Given** mood-aware selection is active, **When** doors are generated, **Then** at least 1 of 3 doors does NOT match mood preference (diversity preserved — not all doors match mood)

3. **Given** no mood data exists (no mood logged this session AND no historical mood correlations in patterns.json), **When** doors are selected, **Then** the system falls back to the existing diversity-based random selection (current behavior unchanged)

4. **Given** the pattern analyzer runs, **When** sessions contain mood entries AND door selections with categorized tasks, **Then** `MoodCorrelation.PreferredType` and `MoodCorrelation.PreferredEffort` fields are populated with the most-selected TaskType and TaskEffort for that mood

5. **Given** a user types `:insights mood`, **When** the command executes, **Then** it shows a focused mood-correlation summary: per-mood breakdown of preferred task types, effort levels, avg tasks completed, and session counts

6. **Given** mood-aware selection is influencing doors, **When** doors are displayed, **Then** a subtle indicator shows the mood context (e.g., "Tuned to: focused" in dimmed text below the door set)

## Tasks / Subtasks

- [ ] Task 1: Enhance MoodCorrelation analysis to populate PreferredType and PreferredEffort (AC: #4)
  - [ ] 1.1: In `internal/tasks/pattern_analyzer.go`, update `analyzeMoodCorrelations()` to cross-reference door selections with task categorization data. For each session with mood entries, look up each selected task's `TaskType` and `TaskEffort`. Build frequency maps per mood.
  - [ ] 1.2: **Do NOT pass `*TaskPool` directly to the analyzer** (violates separation of concerns — pool is mutable live state). Instead, create a lookup table `map[string]TaskCategoryInfo` where `TaskCategoryInfo` has `Type TaskType` and `Effort TaskEffort`. Build this map from the pool BEFORE passing to the analyzer. Add a `SetTaskCategories(categories map[string]TaskCategoryInfo)` method on `PatternAnalyzer`. Current constructor signature at `pattern_analyzer.go:72` is `func NewPatternAnalyzer() *PatternAnalyzer { return &PatternAnalyzer{} }` — keep it unchanged, use the setter instead.
     - **Call sites to update:** `cmd/threedoors/main.go:55` (background goroutine) and `internal/tui/main_model.go:81` (cached load). Both call `tasks.NewPatternAnalyzer()`. After construction, call `analyzer.SetTaskCategories(buildCategoryMap(pool))`.
     - **Define helper:** `func BuildTaskCategoryMap(pool *TaskPool) map[string]TaskCategoryInfo` in `mood_selector.go` (or `pattern_analyzer.go`).
  - [ ] 1.3: In the mood correlation loop, for each session's `DoorSelections`, match `TaskText` against the category map to get `Type` and `Effort`. Build `map[string]map[TaskType]int` (mood → type → count) and `map[string]map[TaskEffort]int` (mood → effort → count). The most frequent type/effort per mood becomes `PreferredType`/`PreferredEffort`.
  - [ ] 1.4: Handle tasks not found in category map (completed/deleted tasks): skip them, don't fail. This is expected since old sessions may reference tasks no longer in the pool.
  - [ ] 1.5: Update `Analyze()` to pass category map context to `analyzeMoodCorrelations()`.
  - [ ] 1.6: Handle partial data: if a mood has sessions but NO categorized task matches (all tasks missing from map), leave PreferredType/PreferredEffort as empty strings — don't fabricate data.

- [ ] Task 2: Implement mood-aware door selection algorithm (AC: #1, #2, #3)
  - [ ] 2.1: Create `internal/tasks/mood_selector.go` with public API `SelectDoorsWithMood(pool *TaskPool, count int, currentMood string, patterns *PatternReport) []*Task` and internal `selectDoorsWithMoodAndRand(pool *TaskPool, count int, currentMood string, patterns *PatternReport, rng *rand.Rand) []*Task` for deterministic testing.
  - [ ] 2.2: Algorithm design (single coherent flow):
    - **Step A — Fallback check:** If `currentMood` is empty OR `patterns` is nil OR no `MoodCorrelation` matches `currentMood` OR the matching correlation has empty `PreferredType` → fall back to `SelectDoors(pool, count)` (existing behavior).
    - **Step B — Generate candidates:** Same N=10 random candidate approach as `selectDoorsWithRand()` in `door_selector.go:37`. Generate 10 candidate sets via Fisher-Yates shuffle from `pool.GetAvailableForDoors()`.
    - **Step C — Score candidates:** For each candidate set, compute `CombinedScore = DiversityScore(candidates) + MoodAlignmentScore(candidates, preferredType, preferredEffort)`.
    - **Step D — MoodAlignmentScore definition:** `func MoodAlignmentScore(tasks []*Task, preferredType TaskType, preferredEffort TaskEffort) int` — for each task: +2 if task.Type matches preferredType, +1 if task.Effort matches preferredEffort. These stack (a task matching both gets +3).
    - **Step E — Diversity floor enforcement (post-scoring):** After selecting the highest-scoring candidate set, count how many tasks match the preferred type. If ALL `count` tasks match (no diversity), replace the last task with a random non-matching task from the pool. If no non-matching tasks exist in pool (all tasks are same type), accept the set as-is — can't enforce diversity when it doesn't exist.
    - **Step F — Return:** Return the final set.
  - [ ] 2.3: **Edge case — pool has fewer than `count` tasks:** `GetAvailableForDoors()` may return fewer than 3 tasks. If pool has fewer than `count` tasks and ALL match mood preference, skip diversity floor enforcement (can't swap what doesn't exist). If pool has 0 tasks, return empty slice.
  - [ ] 2.4: Keep `SelectDoors()` unchanged for backward compatibility.

- [ ] Task 3: Integrate mood-aware selection into TUI flow (AC: #1, #3, #6)
  - [ ] 3.1: In `internal/tui/doors_view.go`, add `currentMood string` field and `patternReport *tasks.PatternReport` field. Add `SetCurrentMood(mood string)` and `SetPatternReport(report *tasks.PatternReport)` methods.
  - [ ] 3.2: Modify `DoorsView.RefreshDoors()` at `doors_view.go:123` — currently calls `dv.currentDoors = tasks.SelectDoors(dv.pool, 3)`. Change to: if `dv.currentMood != ""` and `dv.patternReport != nil`, call `dv.currentDoors = tasks.SelectDoorsWithMood(dv.pool, 3, dv.currentMood, dv.patternReport)`, else keep existing `tasks.SelectDoors(dv.pool, 3)`.
  - [ ] 3.3: Access latest mood from `SessionTracker`: add `func (st *SessionTracker) LatestMood() string` in `session_tracker.go` — returns `st.metrics.MoodEntries[len(st.metrics.MoodEntries)-1].Mood` or empty string if `len(st.metrics.MoodEntries) == 0`.
  - [ ] 3.4: In `main_model.go`, the `MainModel` already has `tracker *tasks.SessionTracker` (line 44) and `patternReport *tasks.PatternReport` (line 48). When handling `ReturnToDoorsMsg` (line 163 area), before calling `m.doorsView.RefreshDoors()`, call `m.doorsView.SetCurrentMood(m.tracker.LatestMood())` and `m.doorsView.SetPatternReport(m.patternReport)`.
  - [ ] 3.5: After mood is recorded, trigger door refresh. The mood_view already sends a `ReturnToDoorsMsg` when mood selection completes (this returns to door view). The `ReturnToDoorsMsg` handler in `main_model.go:163` already calls `m.doorsView.RefreshDoors()`. So mood → return to doors → refresh already happens. Just ensure `SetCurrentMood` is called BEFORE `RefreshDoors` in that handler.
  - [ ] 3.6: Add subtle mood indicator below door set. In `doors_view.go` `View()`, after rendering the three door cards, if `dv.currentMood != ""` and `dv.patternReport != nil`, append: `lipgloss.NewStyle().Faint(true).Italic(true).Render(fmt.Sprintf("Tuned to: %s", dv.currentMood))` centered below the three doors.

- [ ] Task 4: Implement `:insights mood` subcommand (AC: #5)
  - [ ] 4.1: In `internal/tui/search_view.go`, the existing `parseCommand()` at line 85 already splits input by first space: `parts := strings.SplitN(input, " ", 2)` → `cmd` = first word, `args` = rest. So `:insights mood` → cmd="insights", args="mood". Update the `case "insights":` handler (line ~164) to check `args`: if `strings.TrimSpace(args) == "mood"`, call `tasks.FormatMoodInsights(report)` instead of `tasks.FormatInsights(report)`.
  - [ ] 4.2: In `internal/tasks/insights_formatter.go`, add `func FormatMoodInsights(report *PatternReport) string` that produces a focused mood report:
    ```
    Mood Correlation Analysis (N sessions with mood data)

    Mood        | Sessions | Preferred Type | Preferred Effort | Avg Completed
    ------------|----------|----------------|------------------|---------------
    Focused     | 12       | technical      | deep-work        | 3.2
    Stressed    | 8        | administrative | quick-win        | 2.1
    Tired       | 5        | physical       | quick-win        | 1.4
    Energized   | 7        | creative       | medium           | 2.8

    Insights:
    - Your most productive mood is "focused" (avg 3.2 tasks/session)
    - When stressed, you gravitate toward quick-wins — the system adapts to this
    - Mood data improves door selection — keep logging moods!
    ```
  - [ ] 4.3: If no mood correlations exist (empty MoodCorrelations slice), show: "No mood correlation data yet. Log moods during sessions (press M) to build patterns. Need at least 3 sessions with the same mood."
  - [ ] 4.4: Handle partial data in table: if a MoodCorrelation has empty `PreferredType`, display "-" in that column instead of empty string. Same for `PreferredEffort`.
  - [ ] 4.5: The `args` comparison should be case-insensitive: `strings.EqualFold(strings.TrimSpace(args), "mood")`. This handles `:insights Mood`, `:insights MOOD`, etc.

- [ ] Task 5: Write comprehensive unit tests (AC: #1-6)
  - [ ] 5.1: Create `internal/tasks/mood_selector_test.go` with table-driven tests using **deterministic RNG** (`rand.New(rand.NewPCG(fixedSeed, 0))`) via `selectDoorsWithMoodAndRand()` for reproducible assertions — NOT statistical tests:
    - No mood / nil patterns → returns same result as `selectDoorsWithRand()` with same seed
    - Empty string mood → falls back to diversity-only
    - Valid mood with correlation → assert exact tasks returned for given seed (deterministic)
    - Diversity floor: construct pool where all tasks match mood preference, verify at least 1 non-matching in result (or all-matching if pool has no alternatives)
    - Unknown mood (not in correlations) → falls back to random (same as no mood)
    - Pool with fewer than 3 tasks, all matching mood → no swap attempted, returns available tasks
    - Pool with 0 available tasks → returns empty slice
    - Correlation exists but PreferredType is empty → falls back to diversity-only
  - [ ] 5.2: Update `internal/tasks/pattern_analyzer_test.go`:
    - Test `analyzeMoodCorrelations()` now populates PreferredType and PreferredEffort when category map is set
    - Test with sessions having mood "focused" + technical task selections → PreferredType = "technical"
    - Test with mixed types → most frequent type wins
    - Test with no categorized tasks (empty category map) → PreferredType remains ""
    - Test with tasks not in category map (deleted tasks) → skipped gracefully, PreferredType based on remaining matches
    - **Provide pool fixture in tests:** call `analyzer.SetTaskCategories(map[string]TaskCategoryInfo{...})` before `Analyze()`
  - [ ] 5.3: Create `internal/tasks/insights_formatter_test.go` (or add to existing):
    - Test `FormatMoodInsights()` with fully populated MoodCorrelations
    - Test `FormatMoodInsights()` with empty correlations → shows "no data yet" message
    - Test `FormatMoodInsights()` with partial data (some moods have PreferredType, some don't) → shows "-" for empty fields
    - Test `FormatMoodInsights()` with nil report → shows "not enough data" message
  - [ ] 5.4: Test `SessionTracker.LatestMood()` in `session_tracker_test.go`:
    - No moods → returns ""
    - One mood → returns that mood
    - Multiple moods → returns last one
  - [ ] 5.5: Test `MoodAlignmentScore()` directly:
    - Task matches type only → +2
    - Task matches effort only → +1
    - Task matches both → +3
    - Task matches neither → +0
    - Mix of matching and non-matching → correct sum
  - [ ] 5.6: Test `:insights mood` command parsing in context:
    - `:insights mood` → calls FormatMoodInsights
    - `:insights` (no arg) → calls FormatInsights (existing behavior)
    - `:insights MOOD` → calls FormatMoodInsights (case insensitive)

- [ ] Task 6: Update existing tests and integration (AC: all)
  - [ ] 6.1: `NewPatternAnalyzer()` constructor signature is UNCHANGED (returns `*PatternAnalyzer{}`). Existing 26 test call sites in `pattern_analyzer_test.go` do NOT need updating. Only tests that test mood correlation WITH category data need to call `analyzer.SetTaskCategories()` after construction.
  - [ ] 6.2: Update `cmd/threedoors/main.go:55` — after `analyzer := tasks.NewPatternAnalyzer()`, add `analyzer.SetTaskCategories(tasks.BuildTaskCategoryMap(pool))` before calling `analyzer.Analyze()`.
  - [ ] 6.3: Update `internal/tui/main_model.go:81` — same pattern: after creating analyzer, set task categories from pool before loading patterns.
  - [ ] 6.4: Verify all existing tests still pass: `go test ./internal/tasks/... ./internal/tui/...`

## Dev Notes

### Architecture & Patterns

- **New files:** `internal/tasks/mood_selector.go`, `internal/tasks/mood_selector_test.go`
- **Modified files:** `internal/tasks/pattern_analyzer.go` (enhance mood correlations), `internal/tasks/insights_formatter.go` (add `FormatMoodInsights`), `internal/tui/doors_view.go` (mood indicator + mood field), `internal/tui/main_model.go` (mood-aware door selection), `internal/tui/search_view.go` (`:insights mood` subcommand), `internal/tasks/session_tracker.go` (add `LatestMood()` method), `cmd/threedoors/main.go` (pass pool to analyzer)
- **Pattern:** Follow existing `door_selector.go` structure — pure functions with deterministic RNG for testability. `selectDoorsWithMoodAndRand()` internal, `SelectDoorsWithMood()` public.
- **Scoring approach:** Extend existing diversity scoring, don't replace it. Combined score = diversity + mood alignment. This preserves the existing behavior when no mood is active.

### Critical Constraints

- **No external dependencies:** Pure Go stdlib only. No ML libraries.
- **Non-breaking:** `SelectDoors()` remains unchanged. `SelectDoorsWithMood()` is additive.
- **Soft preference, not hard filter:** Mood-aware selection biases toward preferred types but doesn't exclude other types. Diversity floor (at least 1 non-matching door) prevents echo chambers.
- **Graceful degradation:** No mood → diversity only. No patterns → diversity only. No categorization → diversity only. Each missing piece just reduces the mood signal, never breaks.
- **Performance:** Mood alignment scoring adds O(n) per candidate set. With N=10 candidates and 3 tasks each, this is trivial (<1ms).
- **Backward compatible:** Existing sessions.jsonl format unchanged. patterns.json gains populated PreferredType/PreferredEffort fields (previously empty strings, now filled).
- **Configurable weights are OUT OF SCOPE:** The epics AC mentions "selection weights are configurable in a learning config section" but this story uses simple fixed scoring (+2 type, +1 effort). Configurable weights can be a follow-up story if needed.
- **Pattern staleness acknowledged:** Pattern analysis runs at startup only. If user logs mood and completes tasks, mood correlations won't update until next app launch. This is by design (same as Story 4.2) and acceptable for MVP.

### Exact Code References (verified)

- **PatternAnalyzer constructor:** `internal/tasks/pattern_analyzer.go:72` — `func NewPatternAnalyzer() *PatternAnalyzer { return &PatternAnalyzer{} }`
- **PatternAnalyzer.Analyze():** `internal/tasks/pattern_analyzer.go:110`
- **analyzeMoodCorrelations():** `internal/tasks/pattern_analyzer.go:266`
- **MoodCorrelation struct:** `internal/tasks/pattern_analyzer.go:51-58` — PreferredType and PreferredEffort currently empty strings
- **DoorSelector:** `internal/tasks/door_selector.go` — `SelectDoors()` at line 30, `DiversityScore()` at line 12, `selectDoorsWithRand()` at line 37
- **DoorsView.RefreshDoors():** `internal/tui/doors_view.go:123` — `dv.currentDoors = tasks.SelectDoors(dv.pool, 3)` — THIS is what to modify
- **SessionTracker:** `internal/tasks/session_tracker.go` — `MoodEntry` at line 17, `RecordMood()` at line 117
- **MoodView:** `internal/tui/mood_view.go` — mood options at line 10 (focused, tired, stressed, energized, distracted, calm, other)
- **InsightsFormatter:** `internal/tasks/insights_formatter.go` — `FormatInsights()`, mood section at line 73
- **SearchView parseCommand():** `internal/tui/search_view.go:85` — already splits by first space: `strings.SplitN(input, " ", 2)` → cmd + args
- **SearchView executeCommand():** `internal/tui/search_view.go:98` — switch on cmd, `:insights` case at line ~164
- **DoorsView:** `internal/tui/doors_view.go` — avoidance indicator at line 178, `SetAvoidanceData()` at line 73
- **MainModel struct:** `internal/tui/main_model.go:29-55` — has `tracker *tasks.SessionTracker` (line 44), `patternReport *tasks.PatternReport` (line 48), `pool *tasks.TaskPool` (line 43)
- **MainModel.NewMainModel():** `internal/tui/main_model.go:58` — accepts `(pool, tracker, provider, hc)`
- **MainModel pattern loading:** `internal/tui/main_model.go:78-83` — loads from patterns.json, assigns at line 96
- **MainModel ReturnToDoorsMsg handler:** `internal/tui/main_model.go:163` — calls `m.doorsView.RefreshDoors()`
- **NewPatternAnalyzer call sites:** `cmd/threedoors/main.go:55` and `internal/tui/main_model.go:81` (production); 26 sites in `pattern_analyzer_test.go` (tests)
- **Main startup goroutine:** `cmd/threedoors/main.go:52-73` — pattern analysis
- **TaskType/TaskEffort enums:** `internal/tasks/task_categorization.go` — types at line 5, efforts at line 25
- **Task struct:** `internal/tasks/task.go:18-30` — Type, Effort, Location fields
- **GetConfigDirPath:** `internal/tasks/file_manager.go:34-46`
- **Test helpers:** `internal/tasks/test_helpers_test.go` — `newTestTask()`, `newCategorizedTestTask()`, `poolFromTasks()`

### What This Story Does NOT Do

- Does NOT implement persistent avoidance prompts (10+ bypasses → R/B/D/A prompt — that's Story 4.4)
- Does NOT implement goal re-evaluation prompts (that's Story 4.5)
- Does NOT modify SessionTracker recording — only reads existing mood data
- Does NOT add new mood options — uses existing mood vocabulary from mood_view.go
- Does NOT implement configurable weights yet (simple scoring approach first; config can be added in a follow-up if needed)
- Does NOT implement "better than yesterday" tracking (that's a later story)

### Project Structure Notes

- New files in `internal/tasks/`: `mood_selector.go`, `mood_selector_test.go`
- Modified in `internal/tasks/`: `pattern_analyzer.go`, `insights_formatter.go`, `session_tracker.go`
- Modified in `internal/tui/`: `doors_view.go`, `main_model.go`, `search_view.go`
- Modified in `cmd/threedoors/`: `main.go`
- No new packages or directories
- No new go.mod dependencies

### Testing Strategy Notes (for tea agent)

**CRITICAL: Test boundary clarity:**
- `mood_selector_test.go` = **pure unit tests** — no file I/O, no TUI, no Bubble Tea. Tests `MoodAlignmentScore()`, `selectDoorsWithMoodAndRand()`, `BuildTaskCategoryMap()`.
- `pattern_analyzer_test.go` additions = **unit tests with mock data** — tests mood correlation with category map.
- `insights_formatter_test.go` = **pure unit tests** — tests `FormatMoodInsights()` string output.
- `session_tracker_test.go` = **unit tests** — tests `LatestMood()`.
- **NO TUI integration tests** — DoorsView rendering, mood indicator display, and command routing are verified by running the app, not automated tests. Do not attempt to test Bubble Tea views.

**Test ordering (write in this order):**
1. `MoodAlignmentScore()` — lowest level, pure scoring function
2. `BuildTaskCategoryMap()` — builds the lookup table
3. `SetTaskCategories()` — API contract (nil map, empty map, overwrite)
4. `selectDoorsWithMoodAndRand()` — uses all above
5. `FormatMoodInsights()` — formatting output
6. `LatestMood()` — simple accessor

**Key conventions:**
- Use **table-driven tests** (Go convention throughout codebase)
- Use **deterministic RNG** via `selectDoorsWithMoodAndRand()` with fixed seed for exact assertions — NOT statistical distributions
- Use `newCategorizedTestTask()` from `test_helpers_test.go` for creating test tasks with types/efforts
- Use `poolFromTasks()` from `test_helpers_test.go` to build test pools
- Use `t.TempDir()` for any file I/O tests
- Run: `go test ./internal/tasks/... ./internal/tui/...`

**New struct definitions tea must know:**
```go
// TaskCategoryInfo holds categorization for a task, keyed by task text.
// Lives in mood_selector.go (or pattern_analyzer.go).
type TaskCategoryInfo struct {
    Type   TaskType
    Effort TaskEffort
}

// BuildTaskCategoryMap creates a lookup table from a TaskPool.
// Includes ALL tasks, even those with zero-value Type/Effort.
// Returns empty map (not nil) for nil pool or pool with no tasks.
func BuildTaskCategoryMap(pool *TaskPool) map[string]TaskCategoryInfo
```

**New helper tea must CREATE** (does not exist yet):
```go
// makeTestSession creates a minimal valid SessionMetrics for testing.
// Only MoodEntries and DoorSelections need to be populated for mood correlation tests.
// All other fields can be zero-valued.
func makeTestSession(id string, start time.Time, selections []DoorSelectionRecord, bypasses [][]string, moods []MoodEntry) SessionMetrics {
    return SessionMetrics{
        SessionID:      id,
        StartTime:      start,
        EndTime:        start.Add(15 * time.Minute),
        DoorSelections: selections,
        TaskBypasses:   bypasses,
        MoodEntries:    moods,
        MoodEntryCount: len(moods),
    }
}
```

**Exported vs unexported functions:**
- `MoodAlignmentScore()` — **EXPORTED** (capital M) for direct testing
- `selectDoorsWithMoodAndRand()` — **unexported** but testable within same package (tests are in `tasks` package)
- `SelectDoorsWithMood()` — **EXPORTED** public API
- `BuildTaskCategoryMap()` — **EXPORTED** for use in main.go and tests
- `SetTaskCategories()` — **EXPORTED** method on `PatternAnalyzer`

**Expected strings for FormatMoodInsights:**
- Nil report → `"Not enough data yet — need at least 5 sessions for insights."`
- Empty MoodCorrelations → `"No mood correlation data yet. Log moods during sessions (press M) to build patterns. Need at least 3 sessions with the same mood."`
- Populated report → starts with `"Mood Correlation Analysis"` header

### Sample Test Fixtures

**Mood-aware selection test:**
```go
// Pool with categorized tasks
pool := poolFromTasks([]*Task{
    newCategorizedTestTask("Fix login bug", TypeTechnical, EffortDeepWork, LocationWork),
    newCategorizedTestTask("Reply to emails", TypeAdministrative, EffortQuickWin, LocationAnywhere),
    newCategorizedTestTask("Design mockup", TypeCreative, EffortMedium, LocationWork),
    newCategorizedTestTask("Buy groceries", TypePhysical, EffortQuickWin, LocationErrands),
    newCategorizedTestTask("Write unit tests", TypeTechnical, EffortMedium, LocationWork),
    newCategorizedTestTask("File expenses", TypeAdministrative, EffortQuickWin, LocationAnywhere),
})

// Pattern report with mood correlations
patterns := &PatternReport{
    MoodCorrelations: []MoodCorrelation{
        {Mood: "stressed", PreferredType: "administrative", PreferredEffort: "quick-win", SessionCount: 5, AvgTasksCompleted: 2.1},
        {Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", SessionCount: 8, AvgTasksCompleted: 3.5},
    },
}

// Deterministic test with fixed seed
rng := rand.New(rand.NewPCG(42, 0))
doors := selectDoorsWithMoodAndRand(pool, 3, "stressed", patterns, rng)
// Assert exact tasks returned (deterministic with seed 42)
```

**Mood correlation analysis test:**
```go
// Use makeTestSession helper (tea must create this)
sessions := []SessionMetrics{
    makeTestSession("s1", time.Now(), []DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 0}}, nil, []MoodEntry{{Mood: "focused"}}),
    makeTestSession("s2", time.Now(), []DoorSelectionRecord{{TaskText: "Write unit tests", DoorPosition: 1}}, nil, []MoodEntry{{Mood: "focused"}}),
    makeTestSession("s3", time.Now(), []DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 0}}, nil, []MoodEntry{{Mood: "focused"}}),
    makeTestSession("s4", time.Now(), []DoorSelectionRecord{{TaskText: "Write unit tests", DoorPosition: 2}}, nil, []MoodEntry{{Mood: "focused"}}),
    makeTestSession("s5", time.Now(), []DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 1}}, nil, []MoodEntry{{Mood: "focused"}}),
}
analyzer := NewPatternAnalyzer()
analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
    "Fix login bug":    {Type: TypeTechnical, Effort: EffortDeepWork},
    "Write unit tests": {Type: TypeTechnical, Effort: EffortMedium},
})
report, err := analyzer.Analyze(sessions)
// Assert: report.MoodCorrelations[0].Mood == "focused"
// Assert: report.MoodCorrelations[0].PreferredType == "technical"
```

**Diversity floor swap test:**
```go
// All tasks are administrative — diversity floor should try to swap but can't
allAdminPool := poolFromTasks([]*Task{
    newCategorizedTestTask("Email 1", TypeAdministrative, EffortQuickWin, LocationAnywhere),
    newCategorizedTestTask("Email 2", TypeAdministrative, EffortQuickWin, LocationAnywhere),
    newCategorizedTestTask("Email 3", TypeAdministrative, EffortQuickWin, LocationAnywhere),
})
patterns := &PatternReport{
    MoodCorrelations: []MoodCorrelation{
        {Mood: "stressed", PreferredType: "administrative", PreferredEffort: "quick-win", SessionCount: 5},
    },
}
rng := rand.New(rand.NewPCG(42, 0))
doors := selectDoorsWithMoodAndRand(allAdminPool, 3, "stressed", patterns, rng)
// All 3 should be admin tasks — no swap possible since no alternatives exist
```

**SetTaskCategories edge cases:**
```go
analyzer := NewPatternAnalyzer()
// nil map — should not panic
analyzer.SetTaskCategories(nil)
// empty map
analyzer.SetTaskCategories(map[string]TaskCategoryInfo{})
// normal map
analyzer.SetTaskCategories(map[string]TaskCategoryInfo{"task": {Type: TypeTechnical}})
// overwrite — second call replaces first
analyzer.SetTaskCategories(map[string]TaskCategoryInfo{"other": {Type: TypeCreative}})
```

### Concurrency Safety Notes

- Mood-aware selection reads `PatternReport` from `atomic.Pointer` — same thread-safety as Story 4.2
- `LatestMood()` reads from `SessionTracker.metrics.MoodEntries` — only written from TUI thread (mood_view.go), read from TUI thread (door generation) — same goroutine, no race
- No new concurrency concerns beyond what Story 4.2 already handles

### References

- [Source: _bmad-output/planning-artifacts/epics.md - Epic 4, Story 4.3]
- [Source: docs/prd/requirements.md - FR20, FR21]
- [Source: docs/prd/epic-details.md - Epic 4 capabilities]
- [Source: docs/prd/user-journeys.md - Journey 5: Mood-Aware Adaptive Door Selection]
- [Source: internal/tasks/pattern_analyzer.go:51-58 - MoodCorrelation struct]
- [Source: internal/tasks/pattern_analyzer.go:266-316 - analyzeMoodCorrelations()]
- [Source: internal/tasks/door_selector.go:12-80 - SelectDoors, DiversityScore]
- [Source: internal/tasks/session_tracker.go:17-22 - MoodEntry]
- [Source: internal/tui/mood_view.go:10-18 - mood options]
- [Source: internal/tasks/insights_formatter.go:73-84 - mood patterns display]
- [Source: _bmad-output/implementation-artifacts/4-2-pattern-recognition-avoidance.md - Previous story]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
