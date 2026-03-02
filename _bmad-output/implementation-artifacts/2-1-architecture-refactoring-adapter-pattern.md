# Story 2.1: Architecture Refactoring to Adapter Pattern

Status: ready-for-dev

## Story

As a developer,
I want the codebase refactored to use a TaskProvider interface,
so that multiple backends (text file, Apple Notes) can be swapped without changing the core logic.

## Acceptance Criteria

1. **Given** the existing codebase with direct file I/O **When** the refactoring is complete **Then** a `TaskProvider` interface exists with methods: `LoadTasks()`, `SaveTask()`, `DeleteTask()`, `MarkComplete()` **And** a `TextFileProvider` implements this interface (wrapping existing file I/O logic) **And** the MainModel and domain layer depend only on the `TaskProvider` interface, not concrete implementations **And** all existing functionality works identically through the new interface

2. **Given** the adapter pattern is in place **When** unit tests are run **Then** core domain logic can be tested with a mock `TaskProvider` **And** `TextFileProvider` has integration tests covering read, write, and error scenarios **And** test coverage for core domain logic reaches 70%+

3. **Given** the refactored architecture **When** the build is run **Then** CI/CD pipeline via GitHub Actions runs tests on every commit **And** `make test` target is added to the Makefile **And** `make lint` target runs golangci-lint

## Current State Analysis

**CRITICAL: Much of this story is already implemented.** The codebase already has:

- `TaskProvider` interface in `internal/tasks/provider.go` with `LoadTasks()`, `SaveTask()`, `SaveTasks()`, `DeleteTask()`
- `TextFileProvider` in `internal/tasks/text_file_provider.go` implementing the interface
- `MainModel` already takes `TaskProvider` via constructor injection (`main_model.go:41`)
- `saveTasks()` already uses `m.provider.SaveTasks()` (`main_model.go:253-256`)
- `cmd/threedoors/main.go` uses `NewProviderFromConfig()` factory
- `ProviderConfig`, `ProviderFactory`, `FallbackProvider` all exist
- `AppleNotesProvider` stub exists
- CI/CD pipeline with GitHub Actions exists (`.github/workflows/ci.yml`)
- `make test` and `make lint` targets exist in Makefile
- Test coverage: `internal/tasks` at 78.5%, `internal/tui` at 79.4% (both exceed 70%)

## What's Actually Needed (Delta Work)

The remaining work focuses on gap analysis and completing the acceptance criteria fully:

### Gap 1: Missing `MarkComplete()` Method
The original AC specifies `MarkComplete()` as an interface method. Current interface has `LoadTasks()`, `SaveTask()`, `SaveTasks()`, `DeleteTask()` but NO `MarkComplete()`.

### Gap 2: Integration Test Gaps for TextFileProvider
While `provider_test.go` exists, verify it has comprehensive integration tests for:
- Read from existing YAML file
- Write/save tasks
- Delete tasks by ID
- Error scenarios (missing file, corrupt file, permission errors)
- Concurrent access safety

### Gap 3: Mock TaskProvider for TUI Tests
The TUI tests (`main_model_test.go`, `doors_view_test.go`, etc.) should use a mock `TaskProvider` rather than real file operations. Verify mock exists and is used.

### Gap 4: Verify No Direct File I/O Leaks
Ensure NO code path in `internal/tui/` or domain layer calls `LoadTasks()` / `SaveTasks()` / `AppendCompleted()` directly (bypassing the provider). All I/O must go through `TaskProvider`.

## Tasks / Subtasks

- [ ] **Task 1: Add `MarkComplete()` to TaskProvider Interface** (AC: 1)
  - [ ] 1.1: Add `MarkComplete(taskID string) error` method to `TaskProvider` interface in `internal/tasks/provider.go`
  - [ ] 1.2: Implement `MarkComplete()` on `TextFileProvider` — should: load task, set status to complete, set CompletedAt timestamp, save task, append to completed.txt log
  - [ ] 1.3: Implement `MarkComplete()` on `AppleNotesProvider` (stub — return nil or not-implemented error)
  - [ ] 1.4: Implement `MarkComplete()` on `FallbackProvider` — delegate to primary/fallback
  - [ ] 1.5: Update `main_model.go` `TaskCompletedMsg` handler to use `m.provider.MarkComplete(taskID)` instead of calling `tasks.AppendCompleted()` directly

- [ ] **Task 2: Audit and Fix Direct File I/O Leaks** (AC: 1)
  - [ ] 2.1: Search all `internal/tui/*.go` files for direct calls to `tasks.LoadTasks()`, `tasks.SaveTasks()`, `tasks.AppendCompleted()`, `tasks.EnsureConfigDir()` — these should go through the provider
  - [ ] 2.2: In `main_model.go:132`, `tasks.AppendCompleted()` is called directly — this MUST be replaced with `m.provider.MarkComplete()` (from Task 1.5)
  - [ ] 2.3: Verify `cmd/threedoors/main.go` only uses provider for task loading (already correct)
  - [ ] 2.4: Document any remaining direct file calls that are intentional (e.g., config loading, metrics writing are NOT task storage and are fine)

- [ ] **Task 3: Fix Mock TaskProvider for Tests** (AC: 2) — **CRITICAL FIX REQUIRED**
  - [ ] 3.1: `MockProvider` exists in `internal/tasks/provider_test.go` but is missing `MarkComplete()`. Add `MarkComplete()` with: `CompletedIDs []string`, `CompleteErr error` fields, tracking logic
  - [ ] 3.2: **CRITICAL**: `makeModel()` in `internal/tui/main_model_test.go:22-25` uses `tasks.NewTextFileProvider()` — a REAL file provider that reads/writes actual files! This MUST be changed to use a mock. Create a `mockProvider` struct in `main_model_test.go` (or export `MockProvider` from tasks package) that implements `TaskProvider` including `MarkComplete()`
  - [ ] 3.3: Update `makeModel()` to accept an optional provider, defaulting to mock: `func makeModelWithProvider(provider tasks.TaskProvider, texts ...string) *MainModel` and update `makeModel()` to call it with a no-op mock
  - [ ] 3.4: Existing tests that call `makeModel()` with task completion (e.g., `TestDoorsView_CompletionCounter_ShowsAfterCompletion`, `TestFlashMessage_ShowsAfterCompletion`) currently trigger REAL file writes. These must be fixed to use mock
  - [ ] 3.5: The mock must track: `LoadCalls`, `SaveCalls`, `DeleteCalls`, `CompleteCalls` (counts + arguments) for precise assertion

- [ ] **Task 4: Verify TextFileProvider Integration Tests** (AC: 2)
  - [ ] 4.1: Check `provider_test.go` and `text_file_provider_test.go` (or equivalent) for coverage of: LoadTasks, SaveTask, SaveTasks, DeleteTask
  - [ ] 4.2: Add integration test for new `MarkComplete()` method
  - [ ] 4.3: Add error scenario tests: missing file, corrupt YAML, permission denied (using t.TempDir())
  - [ ] 4.4: Verify all integration tests use `t.TempDir()` for isolation

- [ ] **Task 5: Verify Coverage and CI** (AC: 2, 3)
  - [ ] 5.1: Run `go test ./internal/tasks/... -cover` — must be 70%+ (currently 78.5%)
  - [ ] 5.2: Run `make test` — must pass
  - [ ] 5.3: Run `make lint` — must pass with zero warnings
  - [ ] 5.4: Run `make fmt` — must produce no changes
  - [ ] 5.5: Verify GitHub Actions CI runs tests on every commit (already configured)

- [ ] **Task 6: Final Validation** (AC: all)
  - [ ] 6.1: Verify `MainModel` constructor only takes `TaskProvider` interface (not concrete type) — already correct
  - [ ] 6.2: Build application: `make build` — must succeed
  - [ ] 6.3: Run full test suite: `make test` — all pass
  - [ ] 6.4: Run linter: `make lint` — zero warnings
  - [ ] 6.5: Verify all existing functionality works identically (no regressions)

## Dev Notes

### Architecture Patterns & Constraints

- **Two-layer architecture**: TUI layer (`internal/tui/`) imports domain layer (`internal/tasks/`). NEVER the reverse.
- **Bubbletea Elm Architecture (MVU)**: All state changes through Update(), all rendering through View()
- **Constructor injection**: `NewMainModel(pool, tracker, provider)` — provider is an interface
- **gofumpt formatting**: Run before every commit
- **golangci-lint**: Must pass with zero warnings
- **Atomic writes**: All file persistence uses write-to-temp, fsync, rename pattern (in file_manager.go)

### Current Source Tree

```
cmd/threedoors/
  main.go                           # Entry point — uses NewProviderFromConfig()
  main_test.go
internal/
  tasks/
    provider.go                     # TaskProvider interface (MODIFY: add MarkComplete)
    text_file_provider.go           # TextFileProvider (MODIFY: add MarkComplete)
    apple_notes_provider.go         # AppleNotesProvider stub (MODIFY: add MarkComplete stub)
    fallback_provider.go            # FallbackProvider (MODIFY: add MarkComplete)
    provider_config.go              # ProviderConfig + LoadProviderConfig
    provider_factory.go             # NewProviderFromConfig factory
    task.go                         # Task struct with ID, Text, Status, Blocker, Notes, timestamps
    task_status.go                  # Status constants, IsValidTransition()
    task_pool.go                    # TaskPool — in-memory collection with ring buffer
    door_selection.go               # SelectDoors() — Fisher-Yates shuffle
    door_selector.go                # DoorSelector interface
    file_manager.go                 # LoadTasks/SaveTasks/AppendCompleted file I/O
    session_tracker.go              # SessionTracker — metrics recording
    metrics_writer.go               # MetricsWriter — sessions.jsonl
    sync_engine.go                  # Sync engine (future)
    sync_state.go                   # Sync state tracking (future)
    test_helpers_test.go            # Shared test utilities
    *_test.go                       # Test files
  tui/
    main_model.go                   # MainModel — root model (MODIFY: use MarkComplete)
    doors_view.go                   # DoorsView — three doors rendering
    detail_view.go                  # DetailView — task details + status menu
    mood_view.go                    # MoodView — mood capture
    search_view.go                  # SearchView — search + command palette
    messages.go                     # Shared message types
    styles.go                       # Lipgloss styles
    *_test.go                       # Test files
.github/workflows/ci.yml           # GitHub Actions CI (already configured)
Makefile                            # build, run, clean, fmt, lint, test (already has all targets)
```

### Files to MODIFY

| File | Changes |
|------|---------|
| `internal/tasks/provider.go` | Add `MarkComplete(taskID string) error` to interface |
| `internal/tasks/text_file_provider.go` | Implement `MarkComplete()` |
| `internal/tasks/apple_notes_provider.go` | Add `MarkComplete()` stub |
| `internal/tasks/fallback_provider.go` | Add `MarkComplete()` delegation |
| `internal/tui/main_model.go` | Replace `tasks.AppendCompleted()` call with `m.provider.MarkComplete()` |
| `internal/tasks/provider_test.go` | Add `MarkComplete()` test |
| `internal/tasks/test_helpers_test.go` | Update mock if needed |

### Files NOT to Touch

- `cmd/threedoors/main.go` — Already correctly uses provider
- `internal/tasks/task.go` — No changes needed
- `internal/tasks/task_status.go` — No changes needed
- `internal/tasks/task_pool.go` — No changes needed
- `internal/tasks/file_manager.go` — Underlying I/O functions stay (wrapped by provider)
- `internal/tui/doors_view.go` — No changes needed
- `internal/tui/detail_view.go` — No changes needed
- `internal/tui/search_view.go` — No changes needed
- `internal/tui/styles.go` — No changes needed
- `Makefile` — Already has all required targets
- `.github/workflows/ci.yml` — Already configured

### Key Design Decisions

1. **MarkComplete() Signature**: `MarkComplete(taskID string) error` — Takes task ID, handles status update + completion log internally. This encapsulates the "mark complete" workflow that currently spans multiple direct calls in `main_model.go:130-143`.

2. **TextFileProvider.MarkComplete() Implementation**:
```go
func (p *TextFileProvider) MarkComplete(taskID string) error {
    allTasks, err := LoadTasks()
    if err != nil {
        return fmt.Errorf("mark complete: load tasks: %w", err)
    }
    var target *Task
    for _, t := range allTasks {
        if t.ID == taskID {
            target = t
            break
        }
    }
    if target == nil {
        return fmt.Errorf("mark complete: task %s not found", taskID)
    }
    // CRITICAL: Validate status transition before completing
    if !IsValidTransition(target.Status, StatusComplete) {
        return fmt.Errorf("mark complete: invalid transition from %s to complete", target.Status)
    }
    // Update status
    target.Status = StatusComplete
    now := time.Now().UTC()
    target.CompletedAt = &now
    // Remove from active tasks
    filtered := make([]*Task, 0, len(allTasks))
    for _, t := range allTasks {
        if t.ID != taskID {
            filtered = append(filtered, t)
        }
    }
    if err := SaveTasks(filtered); err != nil {
        return fmt.Errorf("mark complete: save tasks: %w", err)
    }
    // Append to completed log — if this fails, task is still saved as complete
    // This is acceptable: the completed.txt log is informational, not authoritative
    if err := AppendCompleted(target); err != nil {
        return fmt.Errorf("mark complete: append completed: %w", err)
    }
    return nil
}
```

3. **MainModel Change — Pool vs Provider Interaction Sequence**:
   The `TaskCompletedMsg` handler in `main_model.go:130-143` currently does:
   - `tasks.AppendCompleted(msg.Task)` — direct file call (leak!)
   - `m.pool.RemoveTask(msg.Task.ID)`
   - `m.provider.SaveTasks(allTasks)` (via saveTasks())

   After refactoring, the **exact sequence** must be:
   ```go
   // Step 1: Provider handles file persistence (save + log)
   if err := m.provider.MarkComplete(msg.Task.ID); err != nil {
       // Show error in flash, DO NOT remove from pool on failure
       fmt.Fprintf(os.Stderr, "warning: failed to mark complete: %v\n", err)
       m.flash = "Error completing task"
       return m, ClearFlashCmd()
   }
   // Step 2: Only remove from in-memory pool AFTER provider succeeds
   m.pool.RemoveTask(msg.Task.ID)
   // Step 3: Increment counter, set flash, refresh doors
   m.doorsView.IncrementCompleted()
   m.flash = celebrationMessages[rand.IntN(len(celebrationMessages))]
   ```
   **CRITICAL**: If `MarkComplete()` fails, the task MUST remain in the pool. Only remove from pool after provider confirms success.
   **CRITICAL**: The `saveTasks()` call is NO LONGER NEEDED for completion — `MarkComplete()` handles it internally. Remove the redundant save.

4. **Backward Compatibility**: `LoadTasks()`, `SaveTasks()`, `AppendCompleted()` in `file_manager.go` remain as package-level functions. They're called internally by `TextFileProvider`. No external code should call them directly.

5. **Error Handling for Partial Success**: If `MarkComplete()` succeeds in saving but fails in `AppendCompleted()`, the error is still returned. The TUI handler should log it as a warning but NOT revert the task status. The completed.txt log is informational; the YAML file is the source of truth.

6. **Status Transition Validation**: `MarkComplete()` MUST call `IsValidTransition(currentStatus, StatusComplete)` before changing status. Not all statuses can transition to complete (e.g., `StatusBlocked` cannot go directly to `StatusComplete` per the transition matrix).

### Previous Story Intelligence

**From Story 1.6 (Essential Polish):**
- All views have polished styling with per-door colors, greeting messages, celebration messages
- `flashStyle.Render()` for temporary messages
- `ClearFlashCmd()` for timed message clearing (3 seconds)
- All tests pass, coverage well above 70%

**From Story 1.5 (Session Metrics):**
- `SessionTracker` with UUID, event tracking, JSON persistence
- `MetricsWriter` for sessions.jsonl
- Non-blocking metrics (<1ms overhead)

**Patterns to maintain:**
- Value receivers for Bubbletea Update/View
- Pointer receivers for mutation
- `t.TempDir()` for test isolation
- Table-driven tests with descriptive names
- `fmt.Errorf("context: %w", err)` for error wrapping
- `math/rand/v2` for random numbers (Go 1.25.4+)

### Git Intelligence

Recent commits show the project follows:
- Commit format: `feat: <description>` / `fix: <description>` / `test: <description>`
- Single commit per story
- All tests must pass before commit
- gofumpt formatting required

### Testing Requirements

**New tests needed:**
- `TestTextFileProvider_MarkComplete_Success` — mark task complete, verify removed from active YAML, present in completed.txt with timestamp
- `TestTextFileProvider_MarkComplete_NotFound` — attempt on nonexistent ID, verify error contains "not found"
- `TestTextFileProvider_MarkComplete_AlreadyComplete` — attempt on already-complete task, verify returns invalid transition error
- `TestTextFileProvider_MarkComplete_InvalidTransition` — attempt from StatusBlocked, verify returns error
- `TestTextFileProvider_MarkComplete_PartialFailure` — test behavior when SaveTasks succeeds but AppendCompleted fails (mock file system)
- `TestMockProvider_MarkComplete` — verify mock records the call with correct task ID
- `TestMainModel_TaskCompleted_PoolRemainsOnProviderFailure` — if MarkComplete() fails, verify task is still in pool
- `TestMainModel_TaskCompleted_PoolRemovedOnSuccess` — if MarkComplete() succeeds, verify task removed from pool
- `TestNoDirectAppendCompletedCalls` — grep/verify that `tasks.AppendCompleted` is NOT called from any `internal/tui/*.go` file (negative assertion, can be a build constraint or code review check)
- Update `TestFlashMessage_ShowsAfterCompletion` if it depends on `tasks.AppendCompleted` call pattern

**Existing tests that must NOT break:**
- All 78.5% of `internal/tasks` tests
- All 79.4% of `internal/tui` tests
- `make test` green, `make lint` clean

### BDD Acceptance Test Scenarios

```gherkin
# AC 1: TaskProvider interface with MarkComplete
Given the TaskProvider interface
When inspected
Then it has methods: LoadTasks, SaveTask, SaveTasks, DeleteTask, MarkComplete

# AC 1: TextFileProvider implements MarkComplete
Given a TextFileProvider with tasks loaded
When MarkComplete is called with a valid task ID
Then the task is removed from active tasks YAML file
And the task is appended to completed.txt with timestamp
And no error is returned

# AC 1: MainModel uses provider exclusively
Given the MainModel receives a TaskCompletedMsg
When processing the completion
Then it calls m.provider.MarkComplete(taskID)
And it does NOT call tasks.AppendCompleted() directly

# AC 2: Mock provider for testing
Given TUI tests need a TaskProvider
When tests create a MainModel
Then a mock TaskProvider is injected
And mock records all method calls for assertion

# AC 2: Coverage meets target
Given the test suite runs
When coverage is measured
Then internal/tasks/ coverage is >= 70%
And internal/tui/ coverage is >= 70%

# AC 3: CI/CD pipeline
Given code is pushed to the repository
When GitHub Actions runs
Then tests execute automatically
And linting runs automatically
And build succeeds
```

### TEA Test Architecture Guidance

**Test File Locations for New Tests:**
- `internal/tasks/provider_test.go` — Add `MarkComplete()` tests for MockProvider + TextFileProvider
- `internal/tui/main_model_test.go` — Fix `makeModel()` to use mock, add pool-integrity tests

**MockProvider Enhancement (in `internal/tasks/provider_test.go`):**
```go
type MockProvider struct {
    Tasks        []*Task
    SavedTasks   []*Task
    DeletedIDs   []string
    CompletedIDs []string     // NEW: track MarkComplete calls
    LoadErr      error
    SaveErr      error
    DeleteErr    error
    CompleteErr  error         // NEW: configurable error
    LoadDelay    time.Duration
}

func (m *MockProvider) MarkComplete(taskID string) error {
    if m.CompleteErr != nil {
        return m.CompleteErr
    }
    m.CompletedIDs = append(m.CompletedIDs, taskID)
    return nil
}
```

**TUI Test Mock (in `internal/tui/main_model_test.go`):**
```go
// testProvider is a no-op TaskProvider for TUI testing.
type testProvider struct {
    completedIDs []string
    completeErr  error
    savedTasks   []*tasks.Task
    saveErr      error
}

func (p *testProvider) LoadTasks() ([]*tasks.Task, error) { return nil, nil }
func (p *testProvider) SaveTask(t *tasks.Task) error      { return p.saveErr }
func (p *testProvider) SaveTasks(ts []*tasks.Task) error {
    p.savedTasks = append(p.savedTasks, ts...)
    return p.saveErr
}
func (p *testProvider) DeleteTask(id string) error         { return nil }
func (p *testProvider) MarkComplete(id string) error {
    if p.completeErr != nil {
        return p.completeErr
    }
    p.completedIDs = append(p.completedIDs, id)
    return nil
}

func makeModel(texts ...string) *MainModel {
    return makeModelWithProvider(&testProvider{}, texts...)
}

func makeModelWithProvider(provider tasks.TaskProvider, texts ...string) *MainModel {
    pool := makePool(texts...)
    tracker := tasks.NewSessionTracker()
    return NewMainModel(pool, tracker, provider)
}
```

**Key Test Scenarios for TEA:**
1. `TestMarkComplete_ValidatesTransition` — Table-driven: test each status → complete transition against the status matrix
2. `TestMarkComplete_RemovesFromYAML_AddsToCompleted` — Integration test with real file I/O using `t.TempDir()`
3. `TestMainModel_CompletionWithMockProvider` — Verify mock's `CompletedIDs` has the right task ID after completion flow
4. `TestMainModel_CompletionFailure_TaskStaysInPool` — Set `completeErr` on mock, verify pool still has task
5. `TestMainModel_CompletionSuccess_TaskRemovedFromPool` — Verify pool.GetAllTasks() no longer has the completed task

### Definition of Done

- `MarkComplete()` method added to `TaskProvider` interface
- All implementations updated (TextFileProvider, AppleNotesProvider, FallbackProvider)
- `main_model.go` uses provider.MarkComplete() instead of direct file calls
- No direct file I/O calls from TUI layer (all through provider)
- All tests pass: `make test` green
- Code formatted: `make fmt` produces no changes
- Lint clean: `make lint` passes with zero warnings
- Test coverage >= 70% for both packages
- All existing functionality works identically (no regressions)

### Project Structure Notes

- Changes are minimal — mostly adding one method to existing interfaces/implementations
- This is a refinement of already-implemented adapter pattern, not a full rewrite
- Maintains alignment with `docs/architecture/source-tree.md`

### References

- [Source: docs/prd/epics-and-stories.md#Story 2.1] — Original acceptance criteria
- [Source: internal/tasks/provider.go] — Current TaskProvider interface
- [Source: internal/tasks/text_file_provider.go] — Current TextFileProvider implementation
- [Source: internal/tui/main_model.go:130-143] — Current TaskCompletedMsg handler with direct file call
- [Source: internal/tasks/file_manager.go] — Underlying file I/O functions
- [Source: docs/architecture/coding-standards.md] — Naming, formatting, linting rules
- [Source: docs/architecture/test-strategy-and-standards.md] — Testing philosophy and coverage targets

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
