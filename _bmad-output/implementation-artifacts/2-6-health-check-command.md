# Story 2.6: Health Check Command

Status: review

## Story

As a user,
I want a health check command to verify Apple Notes connectivity and system status,
So that I can diagnose sync issues and ensure the app is working correctly.

## Acceptance Criteria

1. **`:health` command triggers health check:**
   - Given the application is running
   - When the user types `:health` in the command palette
   - Then the system checks: Apple Notes accessibility, database read/write, sync status, last successful sync timestamp
   - And displays results with green (OK) / red (FAIL) indicators

2. **Actionable suggestions for detected issues:**
   - Given the health check detects issues
   - When results are displayed
   - Then actionable suggestions are shown (e.g., "Grant Full Disk Access in System Settings")
   - And each check has a clear status indicator: OK (green), FAIL (red), WARN (yellow)

3. **Health check completes within performance bounds:**
   - Given the health check is triggered
   - Then all checks complete within 3 seconds (verified by timing assertion in tests, NOT enforced in code via timeout)
   - And the UI remains responsive during checks (non-blocking via tea.Cmd)
   - NOTE: The 3-second bound is a test-only SLA. Do not implement context cancellation or timeouts in RunAll(). Simply verify via a performance test that RunAll() completes within 3 seconds with realistic test data.

4. **Health check works when Apple Notes is unavailable:**
   - Given Apple Notes provider is not configured or unavailable
   - When `:health` is run
   - Then the Apple Notes check shows WARN with message "Apple Notes not configured - using text file backend"
   - And other checks (database, task file) still run and report status

5. **Help command includes `:health`:**
   - Given the user types `:help`
   - Then `:health` appears in the list of available commands

## Definition of Done

- All unit and integration tests pass (`go test ./...`)
- `gofumpt` formatting applied
- `golangci-lint run ./...` passes with zero warnings
- 70%+ test coverage on `health_checker.go`
- Health check works with both TextFileProvider and MockProvider
- All 5 acceptance criteria verified via automated tests
- No regressions in existing tests
- `:health` added to `:help` command list

## Tasks / Subtasks

- [ ] Task 1: Create HealthChecker domain logic (AC: 1, 2, 4)
  - [ ] 1.1: Create `internal/tasks/health_checker.go` with `HealthChecker` struct:
    ```go
    type HealthChecker struct {
        provider TaskProvider
    }

    func NewHealthChecker(provider TaskProvider) *HealthChecker
    ```
    **NOTE:** Do NOT include `syncEngine *SyncEngine` — it is unused. `CheckSyncStatus()` calls the package-level `LoadSyncState()` function directly. Including an unused field will cause `golangci-lint` to flag it.
  - [ ] 1.2: Define `HealthCheckResult` struct:
    ```go
    type HealthStatus string
    const (
        HealthOK   HealthStatus = "OK"
        HealthFail HealthStatus = "FAIL"
        HealthWarn HealthStatus = "WARN"
    )

    type HealthCheckItem struct {
        Name       string
        Status     HealthStatus
        Message    string
        Suggestion string // actionable fix, empty if OK
    }

    type HealthCheckResult struct {
        Items    []HealthCheckItem
        Overall  HealthStatus // worst status across all items
        Duration time.Duration
    }
    ```
  - [ ] 1.3: Implement `CheckTaskFile()` - verify tasks.yaml exists and is readable/writable:
    - Get task file path using `GetTasksFilePath()` from `file_manager.go` (returns `~/.threedoors/tasks.yaml`)
    - OK: `os.Stat(path)` succeeds AND test write check passes
    - FAIL: file missing (`os.IsNotExist`) → Suggestion: "Check ~/.threedoors/ directory permissions"
    - FAIL: file not writable → Suggestion: "Check file permissions on ~/.threedoors/tasks.yaml"
    - **Test write check implementation:**
      ```go
      // Test writability by creating a temp file in the same directory, then removing it
      tmpPath := filepath.Join(filepath.Dir(tasksPath), ".healthcheck.tmp")
      f, err := os.Create(tmpPath)
      if err != nil {
          return HealthCheckItem{Status: HealthFail, ...}
      }
      f.Close()
      os.Remove(tmpPath)
      ```
    - **NOTE:** `CheckTaskFile()` uses `GetTasksFilePath()` directly (package-level function). In tests, call `SetHomeDir(t.TempDir())` FIRST to isolate from real `~/.threedoors/`.
  - [ ] 1.4: Implement `CheckDatabaseReadWrite()` - verify task file can be loaded and parsed:
    - OK: LoadTasks via provider succeeds and returns valid tasks
    - FAIL: LoadTasks returns error → Suggestion: "Task file may be corrupt. Try backing up and recreating ~/.threedoors/tasks.yaml"
    - Include task count in message: "X tasks loaded successfully"
  - [ ] 1.5: Implement `CheckSyncStatus()` - check sync state health:
    - Call `LoadSyncState()` (package-level function from `sync_state.go`)
    - **Detecting "no sync history":** `LoadSyncState()` returns `(empty, nil)` when file is missing. Detect this by checking `state.LastSyncTime.IsZero()` — a zero time means no sync has ever occurred.
    - **Detecting corrupt file:** `LoadSyncState()` returns `(empty, non-nil-error)` when YAML is unparseable. Treat this as WARN (not FAIL).
    - OK: `!state.LastSyncTime.IsZero()` AND `time.Since(state.LastSyncTime) < 24*time.Hour`
    - WARN: `!state.LastSyncTime.IsZero()` AND `time.Since(state.LastSyncTime) >= 24*time.Hour` → Suggestion: "Press S in doors view to trigger a sync"
    - WARN: `state.LastSyncTime.IsZero()` (file missing or never synced) → Suggestion: "No sync history found. Sync will initialize on next provider connection"
    - WARN: `LoadSyncState()` returns non-nil error → Suggestion: "Sync state file may be corrupt. It will be rebuilt on next sync."
    - Include last sync timestamp in message if available: "Last sync: 2 hours ago"
    - **NOTE:** In tests, call `SetHomeDir(t.TempDir())` and use `SaveSyncState()` to create test sync state files.
  - [ ] 1.6: Implement `CheckAppleNotesAccess()` - check Apple Notes provider availability:
    - OK: Provider is an Apple Notes provider and LoadTasks succeeds
    - WARN: Provider is TextFileProvider → Suggestion: "Apple Notes not configured - using text file backend"
    - FAIL: Provider is Apple Notes but LoadTasks fails → Suggestion: "Grant Full Disk Access in System Settings > Privacy & Security"
    - Note: Since real Apple Notes provider doesn't exist yet, this should check provider type via type assertion
  - [ ] 1.7: Implement `RunAll()` method that runs all checks and returns `HealthCheckResult`:
    ```go
    func (hc *HealthChecker) RunAll() HealthCheckResult {
        start := time.Now()
        var items []HealthCheckItem
        items = append(items, hc.CheckTaskFile())
        items = append(items, hc.CheckDatabaseReadWrite())
        items = append(items, hc.CheckSyncStatus())
        items = append(items, hc.CheckAppleNotesAccess())

        overall := HealthOK
        for _, item := range items {
            if item.Status == HealthFail { overall = HealthFail; break }
            if item.Status == HealthWarn && overall == HealthOK { overall = HealthWarn }
        }
        return HealthCheckResult{Items: items, Overall: overall, Duration: time.Since(start)}
    }
    ```

- [ ] Task 2: Add `:health` command to command palette (AC: 1, 5)
  - [ ] 2.1: Add `HealthCheckMsg` message type to `internal/tui/messages.go`:
    ```go
    type HealthCheckMsg struct {
        Result tasks.HealthCheckResult
    }
    ```
  - [ ] 2.2: Add `case "health":` to `executeCommand()` in `search_view.go`:
    ```go
    case "health":
        return sv.runHealthCheck()
    ```
  - [ ] 2.3: Implement `runHealthCheck()` method on SearchView:
    - Needs access to `HealthChecker` - add `healthChecker *tasks.HealthChecker` field to `SearchView`
    - Return a `tea.Cmd` that runs health check asynchronously:
    ```go
    func (sv *SearchView) runHealthCheck() tea.Cmd {
        return func() tea.Msg {
            result := sv.healthChecker.RunAll()
            return HealthCheckMsg{Result: result}
        }
    }
    ```
  - [ ] 2.4: Update `NewSearchView()` to accept `HealthChecker` parameter:
    ```go
    func NewSearchView(pool *tasks.TaskPool, tracker *tasks.SessionTracker, hc *tasks.HealthChecker) *SearchView
    ```
  - [ ] 2.5: Update all `NewSearchView()` call sites in `main_model.go` (3 locations):
    - Line ~78 (ReturnToDoorsMsg handler): `NewSearchView(m.pool, m.tracker, m.healthChecker)`
    - Line ~93 (ReturnToSearchMsg handler): `NewSearchView(m.pool, m.tracker, m.healthChecker)`
    - Line ~216 (slash key handler): `NewSearchView(m.pool, m.tracker, m.healthChecker)`
  - [ ] 2.5b: **CRITICAL** Update all `NewSearchView()` call sites in `search_view_test.go` (3 locations):
    - `newTestSearchView` helper (line ~14): add `nil` as third arg
    - `newTestSearchViewWithTracker` helper (line ~20): add `nil` as third arg
    - `TestSearchView_FilterTasks_AllStatuses` inline call (line ~97): add `nil` as third arg
  - [ ] 2.6: Add `healthChecker` field to `MainModel` and update `NewMainModel()`:
    ```go
    func NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, hc *tasks.HealthChecker) *MainModel
    ```
  - [ ] 2.6b: **CRITICAL** Update all `NewMainModel()` call sites in `main_model_test.go`:
    - `makeModel()` helper (line ~24): add `nil` as third arg
    - Any inline `NewMainModel()` calls: add `nil` as third arg
  - [ ] 2.7: Update `cmd/threedoors/main.go` to create `HealthChecker` and pass to `NewMainModel()`:
    ```go
    hc := tasks.NewHealthChecker(provider) // provider is TextFileProvider from existing setup
    model := tui.NewMainModel(pool, tracker, hc)
    ```
  - [ ] 2.8: Update `:help` text to include `:health`:
    ```go
    "Commands: :add <text>, :mood [mood], :stats, :health, :help, :quit | Keys: / search, a/w/d select, s re-roll, Enter open, m mood, q quit"
    ```

- [ ] Task 3: Render health check results in TUI (AC: 1, 2)
  - [ ] 3.1: Add `ViewHealth` mode to `ViewMode` enum in `main_model.go`:
    ```go
    const (
        ViewDoors ViewMode = iota
        ViewDetail
        ViewMood
        ViewSearch
        ViewHealth
    )
    ```
  - [ ] 3.2: Create `internal/tui/health_view.go` with `HealthView` struct:
    ```go
    type HealthView struct {
        result tasks.HealthCheckResult
        width  int
    }

    func NewHealthView(result tasks.HealthCheckResult) *HealthView {
        return &HealthView{result: result}
    }

    func (hv *HealthView) SetWidth(w int) {
        hv.width = w
    }
    ```
    **NOTE:** `SetWidth()` is required — `MainModel.Update()` calls `SetWidth()` on all views during `tea.WindowSizeMsg` handling. Without it, the health view won't resize on terminal resize.
  - [ ] 3.3: Implement `HealthView.View()` rendering with colored status indicators:
    - Header: "ThreeDoors - Health Check" (bold, accent color)
    - Per item: `[OK] Task File: 12 tasks loaded successfully` (green)
    - Per item: `[FAIL] Apple Notes: Cannot access Apple Notes database` (red)
    - Per item: `  → Grant Full Disk Access in System Settings` (yellow, indented suggestion)
    - Footer: `Overall: OK | Completed in 45ms`
    - Help line: `Press Esc to return`
  - [ ] 3.4: Add health check styles to `styles.go`:
    ```go
    healthOKStyle   = lipgloss.NewStyle().Foreground(colorComplete).Bold(true)  // green
    healthFailStyle = lipgloss.NewStyle().Foreground(colorBlocked).Bold(true)   // red
    healthWarnStyle = lipgloss.NewStyle().Foreground(colorInProgress).Bold(true) // yellow/orange
    healthSuggestionStyle = lipgloss.NewStyle().Foreground(colorInProgress)      // yellow
    ```
  - [ ] 3.5: Implement `HealthView.Update()` - handles Esc to return to doors:
    ```go
    func (hv *HealthView) Update(msg tea.Msg) tea.Cmd {
        if msg, ok := msg.(tea.KeyMsg); ok {
            if msg.Type == tea.KeyEscape {
                return func() tea.Msg { return ReturnToDoorsMsg{} }
            }
        }
        return nil
    }
    ```
    **NOTE:** Do NOT handle `"q"` key in HealthView. The `"q"` key triggers `tea.Quit` at the doors level (`main_model.go:190`). In HealthView, only `Esc` returns to doors. If the user presses `"q"` from HealthView, the key falls through to `MainModel.updateHealth()` which should NOT quit — it should be silently ignored (return nil). The `"q"` to quit only works from ViewDoors. Test this behavior explicitly.
  - [ ] 3.6: Handle `HealthCheckMsg` in `MainModel.Update()`:
    ```go
    case HealthCheckMsg:
        m.healthView = NewHealthView(msg.Result)
        m.healthView.SetWidth(m.width)
        m.viewMode = ViewHealth
        return m, nil
    ```
  - [ ] 3.7: Add `ViewHealth` case to `MainModel.View()` and view delegation:
    - Add `healthView *HealthView` field to `MainModel` struct
    - Add `ViewHealth` case to `View()`:
      ```go
      case ViewHealth:
          if m.healthView != nil {
              view = m.healthView.View()
          }
      ```
    - Add `updateHealth()` method:
      ```go
      func (m *MainModel) updateHealth(msg tea.Msg) (tea.Model, tea.Cmd) {
          if m.healthView == nil {
              return m, nil
          }
          cmd := m.healthView.Update(msg)
          return m, cmd
      }
      ```
    - Add `ViewHealth` case to view delegation in `Update()`:
      ```go
      case ViewHealth:
          return m.updateHealth(msg)
      ```
    - Add `WindowSizeMsg` handling for health view:
      ```go
      if m.healthView != nil {
          m.healthView.SetWidth(msg.Width)
      }
      ```
    - In `ReturnToDoorsMsg` handler, add cleanup: `m.healthView = nil`

- [ ] Task 4: Write comprehensive tests (AC: all)
  - [ ] 4.1: Create `internal/tasks/health_checker_test.go` with `package tasks` (NOT `package tasks_test`):
    **CRITICAL:** Must use `package tasks` so that `MockProvider` from `provider_test.go` is accessible. All health checker test files must be in the same package.
    **CRITICAL:** Always pair `SetHomeDir()` with cleanup: `SetHomeDir(t.TempDir()); defer SetHomeDir("")` — this follows the pattern in `sync_state_test.go` lines 13, 71, 108.
    ```go
    package tasks

    // Test helper - always use this for consistent setup
    func setupHealthTestDir(t *testing.T) string {
        t.Helper()
        tmpDir := t.TempDir()
        SetHomeDir(tmpDir)
        t.Cleanup(func() { SetHomeDir("") })
        return tmpDir
    }

    func newTestHealthChecker(provider TaskProvider) *HealthChecker {
        return NewHealthChecker(provider)
    }

    // Test helper for building HealthCheckResult fixtures
    func newTestHealthResult(items ...HealthCheckItem) HealthCheckResult {
        overall := HealthOK
        for _, item := range items {
            if item.Status == HealthFail { overall = HealthFail; break }
            if item.Status == HealthWarn && overall == HealthOK { overall = HealthWarn }
        }
        return HealthCheckResult{Items: items, Overall: overall, Duration: 10 * time.Millisecond}
    }
    ```
  - [ ] 4.2: Unit tests for `CheckTaskFile()`:
    **CRITICAL:** Every test must call `setupHealthTestDir(t)` first to isolate from real `~/.threedoors/`.
    **NOTE on "file missing" vs "dir missing":** When `SetHomeDir` is called with a temp dir, the `.threedoors/` subdir does NOT exist yet. `CheckTaskFile()` should handle both "directory missing" and "file missing" as the same FAIL case — both mean the tasks file doesn't exist. The suggestion should say "directory permissions" in either case.
    | Test Case | Setup | Expected Status | Expected Suggestion |
    |---|---|---|---|
    | File exists and writable | `setupHealthTestDir(t)`, then `SaveTasks([]*Task{...})` to create dir + file | OK | empty |
    | File missing (no dir) | `setupHealthTestDir(t)` only, no file created | FAIL | contains "directory permissions" |
    | File not writable | See full setup below | FAIL | contains "file permissions" |
    **Full setup for read-only test:**
    ```go
    func TestCheckTaskFile_NotWritable(t *testing.T) {
        if os.Getuid() == 0 { t.Skip("test requires non-root user") }
        setupHealthTestDir(t)
        testTasks := []*Task{newTestTask("a", "Task A", StatusTodo, baseTime)}
        if err := SaveTasks(testTasks); err != nil { t.Fatal(err) }
        path, _ := GetTasksFilePath()
        _ = os.Chmod(path, 0444)
        defer os.Chmod(path, 0644) // cleanup so t.TempDir() can delete it
        hc := newTestHealthChecker(&MockProvider{Tasks: testTasks})
        item := hc.CheckTaskFile()
        // assert item.Status == HealthFail
        // assert strings.Contains(item.Suggestion, "file permissions")
    }
    ```
  - [ ] 4.3: Unit tests for `CheckDatabaseReadWrite()`:
    | Test Case | MockProvider Setup | Expected Status | Expected Message |
    |---|---|---|---|
    | Provider loads OK, 5 tasks | Tasks: 5 tasks, LoadErr: nil | OK | contains "5 tasks loaded successfully" |
    | Provider load error | LoadErr: errors.New("disk err") | FAIL | contains "corrupt" in Suggestion |
    | Provider loads 0 tasks | Tasks: empty, LoadErr: nil | OK | contains "0 tasks loaded successfully" |
  - [ ] 4.4: Unit tests for `CheckSyncStatus()`:
    **CRITICAL:** Every test must call `setupHealthTestDir(t)` first. Use `SaveSyncState()` to create test data.
    **NOTE:** `SaveSyncState()` calls `EnsureConfigDir()` which creates `~/.threedoors/` automatically.
    | Test Case | SyncState Setup | Expected Status | Expected Message/Suggestion |
    |---|---|---|---|
    | Recent sync (<24h) | `SaveSyncState(SyncState{LastSyncTime: time.Now().Add(-1*time.Hour), TaskSnapshots: map[string]TaskSnapshot{}})` | OK | Message contains "Last sync:" |
    | Old sync (>24h) | `SaveSyncState(SyncState{LastSyncTime: time.Now().Add(-48*time.Hour), TaskSnapshots: map[string]TaskSnapshot{}})` | WARN | Suggestion contains "Press S" |
    | No sync state file | `setupHealthTestDir(t)` only | WARN | Suggestion contains "No sync history" |
    | Corrupt sync state | Write garbage bytes to `<tmpDir>/.threedoors/sync_state.yaml` | WARN | Suggestion contains "corrupt" |
    **Full setup for corrupt test:**
    ```go
    func TestCheckSyncStatus_CorruptFile(t *testing.T) {
        tmpDir := setupHealthTestDir(t)
        configPath := filepath.Join(tmpDir, ".threedoors")
        _ = os.MkdirAll(configPath, 0o755)
        _ = os.WriteFile(filepath.Join(configPath, "sync_state.yaml"), []byte("{{{{not yaml"), 0o644)
        hc := newTestHealthChecker(&MockProvider{})
        item := hc.CheckSyncStatus()
        // assert item.Status == HealthWarn
        // assert strings.Contains(item.Suggestion, "corrupt")
    }
    ```
  - [ ] 4.5: Unit tests for `CheckAppleNotesAccess()`:
    | Test Case | Provider Type | Expected Status | Expected Message |
    |---|---|---|---|
    | TextFileProvider | `&TextFileProvider{}` | WARN | contains "Apple Notes not configured" |
    | MockProvider (success) | `&MockProvider{Tasks: [...]}` | OK | contains "accessible" |
    | MockProvider (failure) | `&MockProvider{LoadErr: err}` | FAIL | Suggestion contains "Full Disk Access" |
    **NOTE on AC 4:** Assert that TextFileProvider case Message contains exact string "Apple Notes not configured - using text file backend" to match AC 4.
  - [ ] 4.6: Integration test for `RunAll()`:
    **CRITICAL setup for Overall OK:** Must set up ALL prerequisites:
    ```go
    func TestRunAll_Integration(t *testing.T) {
        tmpDir := t.TempDir()
        SetHomeDir(tmpDir)
        // 1. Create tasks file so CheckTaskFile passes
        testTasks := []*Task{newTestTask("a", "Task A", StatusTodo, time.Now())}
        SaveTasks(testTasks) // creates tasks.yaml
        // 2. Create sync state so CheckSyncStatus passes
        SaveSyncState(SyncState{LastSyncTime: time.Now().UTC()})
        // 3. MockProvider with tasks so CheckDatabaseReadWrite passes
        provider := &MockProvider{Tasks: testTasks}
        hc := NewHealthChecker(provider)
        result := hc.RunAll()
        // MockProvider is not TextFileProvider → CheckAppleNotesAccess returns OK
        assert 4 items, Overall == HealthOK, Duration > 0
    }
    ```
  - [ ] 4.7: Test overall status determination:
    | Items | Expected Overall |
    |---|---|
    | [OK, OK, OK, OK] | OK |
    | [OK, WARN, OK, OK] | WARN |
    | [OK, OK, FAIL, OK] | FAIL |
    | [WARN, WARN, FAIL, OK] | FAIL |
  - [ ] 4.8: Create `internal/tui/health_view_test.go` with `package tui` (NOT `package tui_test`):
    **CRITICAL:** Must use `package tui` to access unexported fields like `width` for SetWidth test.
    ```go
    package tui

    // Test helper
    func newTestHealthView(items ...tasks.HealthCheckItem) *HealthView {
        result := tasks.HealthCheckResult{Items: items, Overall: tasks.HealthOK, Duration: 42 * time.Millisecond}
        hv := NewHealthView(result)
        hv.SetWidth(80)
        return hv
    }
    ```
    **Assertion format for View() tests:** Use these exact substrings to validate rendering:
    - Status indicators: `strings.Contains(view, "[OK]")`, `strings.Contains(view, "[FAIL]")`, `strings.Contains(view, "[WARN]")`
    - Suggestion lines: `strings.Contains(view, "  →")` (two spaces then arrow character →)
    - Overall footer: `strings.Contains(view, "Overall:")` and `strings.Contains(view, "Completed in")`

    Test cases:
    - `TestHealthView_View_RendersOKItem` — view contains `[OK]` and item Name
    - `TestHealthView_View_RendersFAILItem` — view contains `[FAIL]` and `  →` with suggestion text
    - `TestHealthView_View_RendersWARNItem` — view contains `[WARN]` and `  →` with suggestion text
    - `TestHealthView_View_RendersOverallAndDuration` — view contains `Overall:` and `Completed in`
    - `TestHealthView_Update_EscReturnsToDoorsMsg` — `Update(tea.KeyMsg{Type: tea.KeyEscape})` returns non-nil cmd that produces `ReturnToDoorsMsg`
    - `TestHealthView_Update_QKeyReturnsNil` — `Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})` returns nil
    - `TestHealthView_EmptyResult` — zero items, view still contains `Overall: OK`
    - `TestHealthView_SetWidth` — `hv.SetWidth(120)`, assert `hv.width == 120`
  - [ ] 4.9: Test `:help` includes `:health`:
    - In `search_view_test.go`, add test that `:help` FlashMsg.Text contains substring `:health`
  - [ ] 4.10: Test `:health` command returns non-nil `tea.Cmd` (verifies async execution):
    - Create SearchView with a HealthChecker, type `:health`, press Enter
    - Assert `executeCommand()` returns a non-nil `tea.Cmd` (not a direct `HealthCheckMsg`)
  - [ ] 4.11: Test `MainModel.Update(HealthCheckMsg)` wires to ViewHealth (in `main_model_test.go`):
    ```go
    func TestMainModel_HealthCheckMsg_SwitchesToViewHealth(t *testing.T) {
        pool := NewTestPool() // use existing test helper
        m := NewMainModel(pool, nil, nil)
        result := tasks.HealthCheckResult{
            Items: []tasks.HealthCheckItem{{Name: "Test", Status: tasks.HealthOK, Message: "ok"}},
            Overall: tasks.HealthOK, Duration: time.Millisecond,
        }
        m.Update(HealthCheckMsg{Result: result})
        // assert m.viewMode == ViewHealth
        // assert m.healthView != nil
    }
    ```
  - [ ] 4.12: Test `"q"` key does NOT quit from ViewHealth (in `main_model_test.go`):
    ```go
    func TestMainModel_HealthView_QKeyDoesNotQuit(t *testing.T) {
        pool := NewTestPool()
        m := NewMainModel(pool, nil, nil)
        // Set up health view
        m.Update(HealthCheckMsg{Result: tasks.HealthCheckResult{Overall: tasks.HealthOK}})
        // Press "q"
        _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
        // Assert cmd is NOT tea.Quit — cmd should be nil
        // Assert m.viewMode is still ViewHealth (not changed)
    }
    ```
  - [ ] 4.13: Performance test for `RunAll()` — verifies AC 3 (3-second SLA):
    ```go
    func TestRunAll_Performance(t *testing.T) {
        setupHealthTestDir(t)
        // Create prerequisites for all checks to pass
        testTasks := make([]*Task, 50)
        for i := range testTasks {
            testTasks[i] = newTestTask(fmt.Sprintf("id-%d", i), fmt.Sprintf("Task %d", i), StatusTodo, time.Now())
        }
        SaveTasks(testTasks)
        SaveSyncState(SyncState{LastSyncTime: time.Now().UTC(), TaskSnapshots: make(map[string]TaskSnapshot)})
        provider := &MockProvider{Tasks: testTasks}
        hc := NewHealthChecker(provider)
        result := hc.RunAll()
        if result.Duration > 3*time.Second {
            t.Errorf("RunAll took %v, expected < 3s", result.Duration)
        }
    }
    ```
  - [ ] 4.14: **TEA owns `search_view_test.go` test helper updates:**
    When TEA writes tests 4.9 and 4.10, TEA must also update the existing helpers in `search_view_test.go` to pass the new third argument:
    - `newTestSearchView`: change `NewSearchView(pool, nil)` → `NewSearchView(pool, nil, nil)`
    - `newTestSearchViewWithTracker`: change `NewSearchView(pool, tracker)` → `NewSearchView(pool, tracker, nil)`
    - Inline call in `TestSearchView_FilterTasks_AllStatuses`: add `nil` third arg
    These changes are a TEA responsibility (test file ownership) and must be done before adding new health tests to avoid compile errors.

## Dev Notes

### Architecture & Design Patterns

**Command Palette Integration:** The `:health` command follows the exact same pattern as `:stats` in `search_view.go`. Commands are dispatched in `executeCommand()` via a `switch` statement (line 95). The new `case "health":` returns a `tea.Cmd` that runs asynchronously.

**HealthChecker as Domain Logic:** The `HealthChecker` lives in `internal/tasks/` because it checks task-related infrastructure (providers, sync state, task files). It is NOT a TUI concern. The TUI only handles rendering the results.

**View Pattern:** Health check results use a dedicated `HealthView` (like `DetailView`, `MoodView`). This allows the user to read the full report and dismiss with Esc, rather than using a flash message that auto-dismisses.

**Provider Type Detection:** Since the real Apple Notes provider doesn't exist yet, `CheckAppleNotesAccess()` uses Go type assertions to detect the provider type:
```go
switch hc.provider.(type) {
case *TextFileProvider:
    return HealthCheckItem{Status: HealthWarn, Message: "Apple Notes not configured"}
default:
    // Try loading to verify connectivity
    _, err := hc.provider.LoadTasks()
    ...
}
```

### Current Codebase Context

**Existing Command Pattern** (`internal/tui/search_view.go:91-136`):
```go
func (sv *SearchView) executeCommand() tea.Cmd {
    cmd, args := parseCommand(sv.textInput.Value())
    switch cmd {
    case "add": ...
    case "mood": ...
    case "stats": return sv.showStats()
    case "help": ...
    case "quit", "exit": ...
    default: ... "Unknown command"
    }
}
```

**Existing View Modes** (`internal/tui/main_model.go:12-19`):
```go
const (
    ViewDoors ViewMode = iota    // 0
    ViewDetail                    // 1
    ViewMood                      // 2
    ViewSearch                    // 3
)
```
Add `ViewHealth = 4` after ViewSearch.

**Existing Message Types** (`internal/tui/messages.go`):
- ReturnToDoorsMsg, TaskUpdatedMsg, ShowMoodMsg, MoodCapturedMsg, TaskCompletedMsg
- FlashMsg, ClearFlashMsg, SearchResultSelectedMsg, TaskAddedMsg, SearchClosedMsg, ReturnToSearchMsg
- Add: `HealthCheckMsg`

**Existing Styles** (`internal/tui/styles.go`):
- Colors: colorTodo(252), colorInProgress(214), colorBlocked(196), colorInReview(39), colorComplete(82), colorAccent(63), colorSelected(86)
- Reuse: colorComplete for OK (green), colorBlocked for FAIL (red), colorInProgress for WARN (yellow)

**SearchView Constructor Changes:**
Currently: `NewSearchView(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *SearchView`
Needs: `NewSearchView(pool *tasks.TaskPool, tracker *tasks.SessionTracker, hc *tasks.HealthChecker) *SearchView`

Called from `main_model.go` at 3 locations:
- Line 78: `m.searchView = NewSearchView(m.pool, m.tracker)` (ReturnToDoorsMsg)
- Line 93: `m.searchView = NewSearchView(m.pool, m.tracker)` (ReturnToSearchMsg)
- Line 216: `m.searchView = NewSearchView(m.pool, m.tracker)` (slash key)

**MainModel Constructor Changes:**
Currently: `NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *MainModel`
Needs: `NewMainModel(pool *tasks.TaskPool, tracker *tasks.SessionTracker, hc *tasks.HealthChecker) *MainModel`

Called from `cmd/threedoors/main.go`.

### New Files This Story Creates

```
/internal/tasks/health_checker.go       # HealthChecker, HealthCheckResult, HealthCheckItem types
/internal/tasks/health_checker_test.go  # Comprehensive health check tests
/internal/tui/health_view.go           # HealthView TUI component
/internal/tui/health_view_test.go      # HealthView rendering and interaction tests
```

### Files This Story Modifies

```
/internal/tui/search_view.go      # Add "health" case to executeCommand(), add healthChecker field
/internal/tui/messages.go          # Add HealthCheckMsg type
/internal/tui/styles.go            # Add health check styles (healthOKStyle, healthFailStyle, healthWarnStyle)
/internal/tui/main_model.go       # Add ViewHealth mode, HealthCheckMsg handler, healthChecker field, healthView field
/cmd/threedoors/main.go           # Create HealthChecker, pass to NewMainModel()
```

### Technical Requirements

1. **HealthChecker is injectable** - uses TaskProvider interface, testable with MockProvider
2. **All checks are non-blocking** - RunAll() runs via tea.Cmd (Bubbletea async pattern)
3. **Atomic file checks** - use `os.Stat()` and test writes to temp file in same directory
4. **Error wrapping** - `fmt.Errorf("health: check task file: %w", err)`
5. **No panics** - all errors produce FAIL status items, never crash
6. **UTC timestamps** - sync state timestamps compared in UTC
7. **Config directory** - use existing `GetConfigDirPath()` / `EnsureConfigDir()` from file_manager.go

### Coding Standards

- Go 1.25.4+
- `gofumpt` formatting before every commit
- `golangci-lint run ./...` must pass with zero warnings
- Import ordering: stdlib -> external -> internal
- Table-driven tests with descriptive names
- No mocking frameworks - use interfaces and simple stubs
- Exported types/funcs: PascalCase
- Private types/funcs: camelCase

### Testing Standards

- **Domain logic (internal/tasks):** 70%+ coverage target on health_checker.go
- **TUI layer (internal/tui):** Test health_view rendering and key handling
- Use `t.TempDir()` for file I/O tests
- **CRITICAL: Call `SetHomeDir(t.TempDir())` at the start of EVERY test that touches file system** (CheckTaskFile, CheckSyncStatus, RunAll integration) to isolate from real `~/.threedoors/`
- Table-driven tests for multiple scenarios
- MockProvider for provider checks (defined in existing `provider_test.go`, accessible because test file uses `package tasks`)
- **All health checker test files MUST use `package tasks` (not `package tasks_test`)** to access MockProvider and internal helpers
- Skip read-only file tests when running as root: `if os.Getuid() == 0 { t.Skip("...") }`

### Dependencies on Prior Stories

- **Story 2.5** (Bidirectional Sync): `SyncState`, `SyncEngine`, `TaskProvider`, `TextFileProvider` - **all exist in codebase**
- **Story 1.4** (Command Palette): `executeCommand()` switch, `parseCommand()`, `FlashMsg` - **all exist in codebase**
- **Stories 2.2-2.4** (Apple Notes): Real provider not yet implemented - **health check uses type assertion to detect, works with MockProvider**

### Edge Cases to Handle

1. **No tasks file at all:** First launch scenario - CheckTaskFile returns FAIL with helpful message
2. **Empty tasks file:** Valid but empty - CheckDatabaseReadWrite returns OK with "0 tasks loaded"
3. **Corrupt tasks.yaml:** Provider.LoadTasks() errors - CheckDatabaseReadWrite returns FAIL
4. **SyncState missing:** First launch - CheckSyncStatus returns WARN "No sync history"
5. **SyncState corrupt:** Unparseable YAML - CheckSyncStatus returns WARN
6. **Provider is nil:** Edge case - all provider-dependent checks return FAIL "No provider configured"
7. **Very slow provider:** Health check should still complete within 3 seconds (consider timeout)
8. **Concurrent health check:** User runs :health while sync in progress - should not deadlock

### Previous Story Intelligence

From Story 2.5 (Bidirectional Sync) dev notes:
- `TaskProvider` interface with `LoadTasks()`, `SaveTask()`, `SaveTasks()`, `DeleteTask()`
- `TextFileProvider` wraps existing `file_manager.go` functions
- `MockProvider` in test files with `LoadErr`, `SaveErr`, `DeleteErr` fields for failure simulation
- `SyncState` with `LastSyncTime` and `TaskSnapshots` map
- `LoadSyncState()` returns empty state if file missing (not an error)
- Atomic write pattern: write to `.tmp`, fsync, rename
- `GetConfigDirPath()` and `EnsureConfigDir()` for `~/.threedoors/` paths

### Git Intelligence

Recent commits show:
- Story 2.5 established `TaskProvider` interface, `SyncEngine`, `SyncState` patterns
- Tests use table-driven approach with shared fixtures
- `gofumpt` formatting is enforced in CI
- Provider compliance tests verify interface implementation

### References

- [Source: docs/prd/epics-and-stories.md#Story 2.6: Health Check Command]
- [Source: docs/architecture/error-handling-strategy.md]
- [Source: docs/architecture/external-apis.md]
- [Source: docs/architecture/test-strategy-and-standards.md]
- [Source: docs/architecture/coding-standards.md]
- [Source: internal/tui/search_view.go - command palette architecture]
- [Source: internal/tasks/provider.go - TaskProvider interface]
- [Source: internal/tasks/sync_state.go - SyncState model]
- [Source: internal/tasks/sync_engine.go - SyncEngine patterns]

## Dev Agent Record

### Agent Model Used

claude-opus-4-6

### Debug Log References

None - clean implementation.

### Completion Notes List

- Implemented HealthChecker domain logic in internal/tasks/health_checker.go
- Four health checks: TaskFile, DatabaseReadWrite, SyncStatus, AppleNotesAccess
- CheckTaskFile uses os.OpenFile to verify writability of tasks.yaml
- CheckSyncStatus uses LoadSyncState() + LastSyncTime.IsZero() to detect "no history"
- CheckAppleNotesAccess uses Go type assertion to detect TextFileProvider vs other providers
- Created HealthView TUI component with colored [OK]/[FAIL]/[WARN] indicators
- Added :health command to command palette following existing :stats pattern
- Updated NewSearchView and NewMainModel signatures to accept HealthChecker
- Updated all test call sites (search_view_test, main_model_test, main_test)
- All 41+ tests pass, golangci-lint clean, gofumpt applied

### File List

**New files:**
- internal/tasks/health_checker.go
- internal/tasks/health_checker_test.go
- internal/tui/health_view.go
- internal/tui/health_view_test.go

**Modified files:**
- internal/tui/search_view.go (added :health command, healthChecker field, runHealthCheck method)
- internal/tui/messages.go (added HealthCheckMsg type)
- internal/tui/styles.go (added health check styles)
- internal/tui/main_model.go (added ViewHealth, healthView, healthChecker, updateHealth, HealthCheckMsg handler)
- cmd/threedoors/main.go (create HealthChecker, pass to NewMainModel)
- internal/tui/search_view_test.go (updated NewSearchView calls for new signature)
- internal/tui/main_model_test.go (updated NewMainModel calls for new signature)
- cmd/threedoors/main_test.go (updated NewMainModel call for new signature)
- _bmad-output/implementation-artifacts/2-6-health-check-command.md (story file updates)
