# Story 2.3: Read Tasks from Apple Notes

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want the Three Doors app to read tasks from my Apple Notes,
so that I can manage my existing Apple Notes tasks from the terminal.

## Acceptance Criteria

1. **Apple Notes provider reads tasks via AppleScript/osascript:**
   - Given Apple Notes contains a designated task note (configurable note title, default: "ThreeDoors Tasks")
   - When the application starts with Apple Notes provider configured
   - Then tasks are retrieved from Apple Notes via `osascript` using `plaintext text of note` property
   - And tasks appear within <2 seconds of startup

2. **Graceful fallback to text file backend:**
   - Given Apple Notes is not accessible (app closed, permissions denied, osascript fails, not on macOS, note doesn't exist)
   - When the application starts
   - Then the system falls back gracefully to TextFileProvider
   - And a clear message informs the user about the fallback (logged to stderr, shown in status bar)
   - And core Three Doors functionality remains fully operational

3. **Identical UX regardless of provider:**
   - Given the Apple Notes provider is active
   - When the user navigates the Three Doors interface
   - Then the experience is identical to the text file backend (same keys, same views, same flows)

4. **Note format parsing (plaintext):**
   - Given a note in Apple Notes with task lines (one task per line, optional checkbox prefix `- [ ]` / `- [x]`)
   - When tasks are parsed from the plaintext output
   - Then each non-empty content line becomes a Task with proper ID (deterministic from note title + line index), text, status mapping
   - And checked items (`- [x]`) map to StatusComplete
   - And unchecked items (`- [ ]`) or plain lines map to StatusTodo
   - And empty lines are skipped

5. **Configuration for Apple Notes provider:**
   - Given the user's config at `~/.threedoors/config.yaml`
   - When `provider: applenotes` is set (vs default `provider: textfile`)
   - Then AppleNotesProvider is used as the TaskProvider
   - And `note_title: "ThreeDoors Tasks"` configures which note to read from (overridable)

6. **Note not found handling:**
   - Given the configured note title does not exist in Apple Notes
   - When `LoadTasks()` is called
   - Then an error is returned (not an empty list)
   - And the FallbackProvider catches this and delegates to TextFileProvider
   - And stderr shows: `"Apple Notes: note 'ThreeDoors Tasks' not found. Create this note in Apple Notes to use the Apple Notes provider. Falling back to text file."`

## Definition of Done

- All unit and integration tests pass (`go test ./...`)
- `gofumpt` formatting applied to all new/modified files
- `golangci-lint run ./...` passes with zero warnings
- 80%+ test coverage on `apple_notes_provider.go` and `provider_factory.go`
- All existing tests pass (zero regressions)
- All 6 acceptance criteria verified via automated tests
- Darwin-only tests use `//go:build darwin` build tag
- No integration tests require Notes.app — all osascript calls mocked in unit tests

## Tasks / Subtasks

- [ ] Task 1: Spike — validate AppleScript `plaintext text` approach (AC: 1) **MUST COMPLETE BEFORE TASK 2**
  - [ ] 1.1: Create `scripts/spike_applenotes.sh` testing BOTH approaches:
    - `osascript -e 'tell application "Notes" to get plaintext text of note "ThreeDoors Tasks"'` (preferred — returns plain text, no HTML parsing needed)
    - `osascript -e 'tell application "Notes" to get body of note "ThreeDoors Tasks"'` (fallback — returns HTML)
  - [ ] 1.2: Validate: permissions behavior, error messages when Notes.app not running, error when note doesn't exist, auto-launch behavior
  - [ ] 1.3: Confirm `plaintext text` works and document findings in code comments. If `plaintext text` fails, fall back to `body` + HTML parsing (see HTML Parsing Notes section)

- [ ] Task 2: Implement AppleNotesProvider (AC: 1, 3, 4, 6)
  - [ ] 2.1: Create `internal/tasks/apple_notes_provider.go`:
    ```go
    // CommandExecutor abstracts osascript execution for testability.
    // Real impl wraps exec.CommandContext; tests return canned (string, error).
    type CommandExecutor func(ctx context.Context, script string) (string, error)

    // defaultExecutor runs osascript for real
    func defaultExecutor(ctx context.Context, script string) (string, error) {
        cmd := exec.CommandContext(ctx, "osascript", "-e", script)
        out, err := cmd.Output()
        return string(out), err
    }

    type AppleNotesProvider struct {
        noteTitle string
        executor  CommandExecutor // defaults to defaultExecutor
    }

    func NewAppleNotesProvider(noteTitle string) *AppleNotesProvider
    func NewAppleNotesProviderWithExecutor(noteTitle string, executor CommandExecutor) *AppleNotesProvider // for testing
    ```
    - Implement `TaskProvider` interface (LoadTasks, SaveTask, SaveTasks, DeleteTask)
    - `LoadTasks()`: execute osascript to get plaintext, parse lines into `[]*Task`
    - `SaveTask()`, `SaveTasks()`, `DeleteTask()`: return `ErrReadOnly` (write support is Story 2.4)
    - Define: `var ErrReadOnly = errors.New("apple notes provider is read-only")`
  - [ ] 2.2: Implement `executeAppleScript(script string) (string, error)` method on provider:
    - Call `p.executor(ctx, script)` — executor handles exec details
    - Set timeout via `context.WithTimeout` (2 second max)
    - Return meaningful wrapped errors with sentinel detection:
      - Note not found: osascript exits non-zero with stderr containing "Can't get note" → wrap as `fmt.Errorf("apple notes: note %q not found: %w", p.noteTitle, err)`
      - Permission denied: stderr contains "not allowed" → wrap as `fmt.Errorf("apple notes: automation permission denied: %w", err)`
      - Timeout: `context.DeadlineExceeded` → wrap as `fmt.Errorf("apple notes: osascript timed out after 2s: %w", err)`
      - Command not found: exec error → wrap as `fmt.Errorf("apple notes: osascript not found (not macOS?): %w", err)`
    - AppleScript: `tell application "Notes" to get plaintext text of note "<noteTitle>"`
  - [ ] 2.3: Implement `parseNoteBody(body string) []*Task`:
    - Split body by newlines (plaintext — no HTML stripping needed if spike confirms `plaintext text` works)
    - Skip empty lines and whitespace-only lines (`strings.TrimSpace(line) == ""`)
    - Parse checkbox prefixes (case-insensitive X):
      - `- [ ]` → StatusTodo
      - `- [x]` or `- [X]` → StatusComplete
      - Plain text (no checkbox) → StatusTodo
      - `* [ ]` or `* [x]` → treat same as dash prefix (asterisk variant)
      - Indented checkboxes (`  - [ ] nested`) → strip leading whitespace, parse normally
    - Task text = line content after checkbox prefix, trimmed; or full line if no prefix
    - Generate deterministic task ID: `uuid.NewSHA1(uuid.NameSpaceURL, []byte(p.noteTitle+":"+strconv.Itoa(lineIndex)))`
    - Set CreatedAt/UpdatedAt to `time.Now().UTC()`
    - Unicode characters in task text: pass through unchanged
  - [ ] 2.4: Interface compliance test: `var _ TaskProvider = (*AppleNotesProvider)(nil)`

- [ ] Task 3: Implement provider configuration and fallback (AC: 2, 5, 6)
  - [ ] 3.1: Create `internal/tasks/provider_config.go`:
    - Define `ProviderConfig` struct: `Provider string`, `NoteTitle string` (YAML tags: `provider`, `note_title`)
    - `LoadProviderConfig(path string) (*ProviderConfig, error)` reads from `~/.threedoors/config.yaml`
    - Default values: `provider: "textfile"`, `note_title: "ThreeDoors Tasks"`
    - If file doesn't exist, return defaults (not an error)
  - [ ] 3.2: Create `internal/tasks/provider_factory.go`:
    - `NewProviderFromConfig(config *ProviderConfig) TaskProvider` factory function
    - If `config.Provider == "applenotes"`: return `NewFallbackProvider(NewAppleNotesProvider(config.NoteTitle), NewTextFileProvider())`
    - If `config.Provider == "textfile"` or default: return `NewTextFileProvider()`
    - Unknown provider value: log warning, return TextFileProvider
  - [ ] 3.3: Implement `FallbackProvider` in `internal/tasks/fallback_provider.go`:
    ```go
    type FallbackProvider struct {
        primary  TaskProvider
        fallback TaskProvider
        usedFallback bool
        fallbackReason string
    }
    func NewFallbackProvider(primary, fallback TaskProvider) *FallbackProvider
    ```
    - `LoadTasks()`: try primary, on error log to stderr and delegate to fallback, set `usedFallback = true`
    - `SaveTask()`/`SaveTasks()`/`DeleteTask()`: if `usedFallback`, delegate to fallback; else delegate to primary
    - `IsFallback() bool` — for status bar display
    - `FallbackReason() string` — human-readable reason

- [ ] Task 4: Write comprehensive tests (AC: all) — use **table-driven tests** (Go convention)
  - [ ] 4.1: `internal/tasks/apple_notes_provider_test.go` (no build tags — all tests use mock executor):
    - **AC→Test mapping:**
      - AC1: `TestAppleNotesProvider_LoadTasks_Success`
      - AC4: `TestAppleNotesProvider_ParseNoteBody_*` (table-driven)
      - AC6: `TestAppleNotesProvider_LoadTasks_NoteNotFound`
    - **`parseNoteBody` table-driven tests** with these fixtures:
      | Input | Expected Tasks | Notes |
      |-------|---------------|-------|
      | `"- [ ] Buy milk\n- [x] Buy eggs"` | 2 tasks: milk=Todo, eggs=Complete | Basic checkboxes |
      | `"Plain task line"` | 1 task: Todo | No checkbox prefix |
      | `""` | 0 tasks | Empty note |
      | `"\n\n\n"` | 0 tasks | Only empty lines |
      | `"  \n  \t  \n"` | 0 tasks | Whitespace-only lines |
      | `"- [X] Capital X"` | 1 task: Complete | Case-insensitive X |
      | `"* [ ] Asterisk item"` | 1 task: Todo | Asterisk variant |
      | `"  - [ ] Indented"` | 1 task: Todo | Indented checkbox |
      | `"Buy milk\n\nBuy bread"` | 2 tasks | Empty line skipped |
      | `"Task with 🎯 emoji"` | 1 task with emoji | Unicode passthrough |
      | `"Title only"` | 1 task | Note with just a title |
    - Test deterministic ID: call `parseNoteBody` twice with same input → same IDs
    - Test different input → different IDs
    - **`LoadTasks` mock executor tests:**
      - Mock returns valid plaintext → verify parsed tasks
      - Mock returns error (simulating note not found) → verify wrapped error contains "not found"
      - Mock returns error (simulating permission denied) → verify wrapped error contains "permission denied"
      - Mock returns error (simulating timeout via `context.DeadlineExceeded`) → verify wrapped error contains "timed out"
      - Mock returns empty string → verify empty task slice (not error)
    - Test `SaveTask`/`SaveTasks`/`DeleteTask` return `ErrReadOnly` (`errors.Is(err, ErrReadOnly)`)
    - **Mock executor pattern:**
      ```go
      func mockExecutor(output string, err error) CommandExecutor {
          return func(ctx context.Context, script string) (string, error) {
              return output, err
          }
      }
      ```
    - **Sample plaintext fixture** (realistic osascript output):
      ```
      Shopping List
      - [ ] Buy milk
      - [x] Buy eggs
      - [ ] Buy bread

      Work Tasks
      - [ ] Review PR
      - [ ] Update docs
      ```
  - [ ] 4.2: `internal/tasks/provider_config_test.go` (use `t.TempDir()` for file isolation):
    - Test LoadProviderConfig with valid YAML written to temp dir
    - Test missing file → returns defaults, no error
    - Test invalid YAML → returns error
    - Test empty file → returns defaults
    - Test round-trip: write config YAML → load → verify fields match
    - **AC→Test:** AC5: `TestLoadProviderConfig_ValidConfig`
  - [ ] 4.3: `internal/tasks/fallback_provider_test.go` (use MockProvider from `test_helpers_test.go`):
    - **Full state matrix (table-driven):**
      | Primary LoadTasks | Primary Save | Expected Behavior |
      |-------------------|-------------|-------------------|
      | Success | N/A | Use primary results, `IsFallback()=false` |
      | Error | N/A | Use fallback results, `IsFallback()=true`, `FallbackReason()` populated |
      | Success | `ErrReadOnly` | `SaveTasks` delegates to fallback (primary is read-only) |
      | Error | Error | Fallback used for load; if fallback save also fails, return error |
    - **AC→Test:** AC2: `TestFallbackProvider_FallsBackOnPrimaryError`, AC6: `TestFallbackProvider_NoteNotFound_FallsBack`
  - [ ] 4.4: `internal/tasks/provider_factory_test.go`:
    - Test `NewProviderFromConfig` with `provider: "textfile"` → returns `*TextFileProvider`
    - Test `NewProviderFromConfig` with `provider: "applenotes"` → returns `*FallbackProvider`
    - Test `NewProviderFromConfig` with `provider: ""` → defaults to TextFileProvider
    - Test `NewProviderFromConfig` with `provider: "unknown"` → defaults to TextFileProvider with warning
  - [ ] 4.5: Verify all existing tests still pass (zero regressions): `go test ./...`

- [ ] Task 5: Integration wiring (AC: 1, 2, 3)
  - [ ] 5.1: Modify `cmd/threedoors/main.go` — **exact wiring point:**
    ```go
    // CURRENT (line ~13):
    loadedTasks, err := tasks.LoadTasks()

    // REPLACE WITH:
    cfg, err := tasks.LoadProviderConfig(filepath.Join(configPath, "config.yaml"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: config load failed: %v, using defaults\n", err)
        cfg = &tasks.ProviderConfig{Provider: "textfile"}
    }
    provider := tasks.NewProviderFromConfig(cfg)
    loadedTasks, err := provider.LoadTasks()
    ```
    - Pass `provider` to `tui.NewMainModel()` so it can use provider for saves
  - [ ] 5.2: Modify `internal/tui/main_model.go` — **exact wiring point:**
    ```go
    // CURRENT (line ~252):
    func (m *MainModel) saveTasks() error {
        allTasks := m.pool.GetAllTasks()
        return tasks.SaveTasks(allTasks)
    }

    // REPLACE WITH:
    func (m *MainModel) saveTasks() error {
        allTasks := m.pool.GetAllTasks()
        return m.provider.SaveTasks(allTasks)
    }
    ```
    - Add `provider tasks.TaskProvider` field to `MainModel` struct
    - Update `NewMainModel` constructor to accept `TaskProvider` parameter
  - [ ] 5.3: SyncEngine compatibility note: `SyncEngine.Sync()` calls `provider.LoadTasks()` which works fine. `ApplyChanges()` modifies the `TaskPool` in-memory, not the provider. The caller of `Sync()` is responsible for persisting via `provider.SaveTasks()` — if provider returns `ErrReadOnly`, the caller should catch this and skip the save (log a debug message). This is acceptable for read-only mode.
  - [ ] 5.4: Run `gofumpt -w .` and `golangci-lint run ./...`

## Dev Notes

### Spike Research (Folded from Story 2.2)

**Three approaches evaluated:**

1. **AppleScript via osascript (CHOSEN):**
   - Pros: Stable macOS API, no Full Disk Access needed, works with iCloud-synced notes, simple to call from Go via `os/exec`
   - Cons: Notes.app must be running (or will auto-launch), slower than direct DB read (~200-500ms)
   - **Preferred property:** `plaintext text of note` — returns plain text, avoids HTML parsing entirely
   - **Fallback property:** `body of note` — returns HTML, requires HTML stripping (see HTML Parsing Notes)
   - Read: `tell application "Notes" to get plaintext text of note "<title>"`
   - Write: `tell application "Notes" to set body of note "<title>" to "<html>"` (Story 2.4)
   - Risk: Automation permission prompt on first use (System Settings > Privacy > Automation)

2. **Direct SQLite read (NoteStore.sqlite):**
   - Location: `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite`
   - Verdict: Too fragile — proprietary gzip+protobuf format, undocumented schema, requires Full Disk Access

3. **MCP Server (mcp-apple-notes):**
   - Verdict: Over-engineered for direct Go CLI integration; external process dependency

**Decision: AppleScript via `plaintext text` property** — simplest, no HTML parsing, works with iCloud sync, adequate performance.

### Architecture Compliance

- **Pattern:** TaskProvider interface (adapter pattern) — `internal/tasks/provider.go`
- **Reference implementation:** `internal/tasks/text_file_provider.go` — follow same structure
- **New files go in:** `internal/tasks/` package
- **Dependencies:** Only stdlib `os/exec`, `context`, `html` — no new external packages (`github.com/google/uuid` already in go.mod)
- **Error handling:** Idiomatic Go — return `fmt.Errorf("...: %w", err)` wrapped errors
- **Testing:** `CommandExecutor` function type on struct enables mock injection (no interface needed)
- **Build tags:** Use `//go:build darwin` for any test that actually executes osascript

### HTML Parsing Notes (Only if `plaintext text` doesn't work)

If spike reveals `plaintext text` is unavailable, fall back to `body` property with HTML parsing:
- Strip all HTML tags via `regexp.MustCompile("<[^>]*>").ReplaceAllString(body, "")`
- `<br>` and `<div>` → newline (replace before stripping)
- Decode HTML entities via `html.UnescapeString()` from `html` stdlib
- Apple Notes checkboxes in HTML use `<ul class="checklist">` with specific class attributes

### Key Design Decision: Read-Only for This Story

- `SaveTask()`, `SaveTasks()`, `DeleteTask()` return `ErrReadOnly`
- Write support is Story 2.4
- `FallbackProvider` handles this: if primary is read-only and fallback is active, saves go to fallback
- SyncEngine callers must handle `ErrReadOnly` from `provider.SaveTasks()` gracefully (skip save, log debug)

### Deterministic Task IDs

- Use UUID v5 (SHA1-based, deterministic): `uuid.NewSHA1(uuid.NameSpaceURL, []byte(noteTitle+":"+strconv.Itoa(lineIndex)))`
- Same note title + same line position = same UUID every time
- Enables sync engine to track task identity across sessions
- Caveat: If user reorders lines in Apple Notes, IDs will shift — acceptable for v1

### Exact Integration Points

**`cmd/threedoors/main.go` line ~13:** Currently calls `tasks.LoadTasks()` directly. Replace with provider factory.
**`internal/tui/main_model.go` line ~252:** `saveTasks()` calls `tasks.SaveTasks()` directly. Replace with `m.provider.SaveTasks()`.
**`MainModel` struct:** Add `provider tasks.TaskProvider` field, update constructor.

### CI/Testing Strategy

- **TDD flow:** TEA creates failing test stubs first, dev makes them pass
- **All tests use mocks — no build tags needed for this story.** `CommandExecutor` returns `(string, error)` directly, so tests never invoke osascript
- **Table-driven tests** throughout — Go idiomatic pattern
- **Mock pattern:** `CommandExecutor func(ctx, script) (string, error)` — tests inject a closure returning canned output
- **File tests:** Use `t.TempDir()` for config file tests — ensures isolation, auto-cleanup
- **Reuse `MockProvider`** from `test_helpers_test.go` for FallbackProvider tests (it already implements `TaskProvider`)
- **GitHub Actions:** All tests run on ubuntu — zero darwin-only dependencies in unit tests
- **Coverage target:** 80%+ on `apple_notes_provider.go` and `provider_factory.go`

### Project Structure Notes

- All new files in `internal/tasks/` — consistent with existing codebase
- New files: `apple_notes_provider.go`, `provider_config.go`, `provider_factory.go`, `fallback_provider.go`
- Test files: `apple_notes_provider_test.go`, `provider_config_test.go`, `provider_factory_test.go`, `fallback_provider_test.go`
- Config file: `~/.threedoors/config.yaml` — new file, YAML format
- No new packages or directories

### References

- [Source: internal/tasks/provider.go] — TaskProvider interface definition
- [Source: internal/tasks/text_file_provider.go] — Reference implementation to follow
- [Source: internal/tasks/sync_engine.go] — How providers are consumed by sync engine
- [Source: internal/tasks/task.go] — Task model with all required fields
- [Source: internal/tasks/test_helpers_test.go] — Shared test helpers (MockProvider)
- [Source: cmd/threedoors/main.go:13] — Current task loading entry point to modify
- [Source: internal/tui/main_model.go:252] — Current saveTasks() to modify
- [Source: docs/architecture/high-level-architecture.md#Repository Pattern] — Adapter pattern decision
- [Source: docs/prd/epics-and-stories.md#Story 2.3] — Original story requirements
- [Source: docs/prd/epics-and-stories.md#Story 2.2] — Spike requirements folded in

### Git Intelligence

Recent patterns from Story 2.5:
- Files follow `snake_case.go` naming in `internal/tasks/`
- Test files use `_test.go` suffix in same package
- Shared test helpers in `test_helpers_test.go`
- `gofumpt` formatting required
- `golangci-lint` must pass with zero warnings

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
