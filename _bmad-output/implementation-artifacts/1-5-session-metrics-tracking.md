# Story 1.5: Session Metrics Tracking

Status: ready-for-dev

## Story

As a developer validating the Three Doors concept,
I want objective session metrics collected automatically,
so that I can make a data-informed validation decision instead of relying solely on subjective impressions.

## Acceptance Criteria

1. **SessionTracker component** in `internal/tasks/session_tracker.go` - ALREADY EXISTS
   - Tracks session_id, start/end times, duration
   - Tracks behavioral counters: tasks_completed, doors_viewed, refreshes_used, detail_views, notes_added, status_changes, mood_entries
   - Tracks time_to_first_door_seconds (key friction metric)
   - Door selection patterns (left=0, center=1, right=2)
   - Task bypass tracking (tasks shown but not selected before refresh)
   - Status change details
   - Mood tracking with timestamps
   - Constructor `NewSessionTracker()` initializes with UUID and current timestamp
   - All recording methods implemented and integrated

2. **MetricsWriter component** in `internal/tasks/metrics_writer.go` - ALREADY EXISTS
   - Constructor `NewMetricsWriter(baseDir string)` sets path to sessions.jsonl
   - Method `AppendSession(metrics *SessionMetrics)` writes JSON line to file
   - Creates file if doesn't exist, appends to existing

3. **SessionTracker integrated into MainModel** - ALREADY EXISTS
   - MainModel includes tracker field
   - SessionTracker passed to DoorsView, DetailView, SearchView
   - No UI changes (completely invisible to user)

4. **Recording calls integrated into views** - ALREADY EXISTS
   - Door selection calls RecordDoorSelection()
   - Refresh calls RecordRefresh()
   - Mood capture calls RecordMood()
   - Detail view calls RecordDetailView()
   - Status changes call RecordStatusChange() and RecordTaskCompleted()

5. **Session persistence on app exit** - ALREADY EXISTS
   - cmd/threedoors/main.go calls Finalize() and AppendSession() on clean exit
   - Write failures logged as warning to stderr

6. **Metrics file format validated** - NEEDS VERIFICATION
   - Each line must be valid JSON (parseable by `jq`)
   - File is append-only, human-readable

7. **Performance requirements met** - NEEDS VERIFICATION
   - Recording adds <1ms overhead per event
   - No UI lag or stuttering observed

8. **Analysis Scripts Created** - NOT YET IMPLEMENTED
   - `scripts/analyze_sessions.sh` - Session summary and averages
   - `scripts/daily_completions.sh` - Daily completion counts from completed.txt
   - `scripts/validation_decision.sh` - Automated validation criteria evaluation

9. **MetricsWriter tests** - NOT YET IMPLEMENTED
   - Unit tests for AppendSession (file creation, appending, error handling)
   - Integration test verifying JSON Lines format is parseable by jq

10. **Comprehensive SessionTracker tests** - PARTIALLY EXISTS (needs expansion)
    - Existing tests cover: NewSessionTracker, RecordDoorSelection, RecordRefresh, RecordMood, Finalize
    - Missing: RecordDetailView, RecordNoteAdded, RecordStatusChange, RecordTaskCompleted, RecordDoorViewed edge cases
    - **Concurrency is NOT in scope** - Bubbletea uses single-threaded MVU update loop, no mutex needed

11. **Integration test: SessionTracker → MetricsWriter pipeline** - NOT YET IMPLEMENTED
    - Full pipeline test: NewSessionTracker() → record events → Finalize() → MetricsWriter.AppendSession() → read file → verify JSON
    - This is the critical path and must be tested end-to-end

12. **Shell script test fixtures and validation** - NOT YET IMPLEMENTED
    - Sample `scripts/testdata/sessions.jsonl` with known data for script testing
    - Sample `scripts/testdata/completed.txt` with known data for daily_completions testing
    - `make test-scripts` target that runs scripts against test fixtures and verifies output

13. **Definition of Done**
    - `go test ./...` passes with zero failures
    - `golangci-lint run ./...` passes with zero warnings
    - All three analysis scripts execute successfully against test fixture data
    - `make analyze` works correctly
    - `make test-scripts` passes
    - Coverage for `internal/tasks` remains at or above 70%

## Tasks / Subtasks

- [ ] Task 1: Create test helper and fixtures (AC: #9, #11, #12)
  - [ ] 1.1: Create `testSessionMetrics()` helper in `internal/tasks/metrics_writer_test.go` that returns a fully populated `*SessionMetrics` for reuse across tests
  - [ ] 1.2: Create `scripts/testdata/sessions.jsonl` with 6+ sample sessions (known values for script validation)
  - [ ] 1.3: Create `scripts/testdata/completed.txt` with sample completion entries matching format: `[YYYY-MM-DD HH:MM:SS] task_id | task_text`

- [ ] Task 2: Add MetricsWriter tests (AC: #9)
  - [ ] 2.1: Create `internal/tasks/metrics_writer_test.go` using `testSessionMetrics()` helper
  - [ ] 2.2: Test AppendSession creates file if not exists
  - [ ] 2.3: Test AppendSession appends to existing file (multiple sessions - verify 2+ JSON lines)
  - [ ] 2.4: Test output is valid JSON Lines (each line parseable via `json.Unmarshal`)
  - [ ] 2.5: Test error handling for non-existent parent directory (e.g., `/nonexistent/dir/sessions.jsonl`)

- [ ] Task 3: Add integration test: SessionTracker → MetricsWriter pipeline (AC: #11)
  - [ ] 3.1: Test full pipeline: NewSessionTracker() → RecordDoorSelection + RecordRefresh + RecordMood + RecordTaskCompleted → Finalize() → MetricsWriter.AppendSession() → read file → json.Unmarshal → verify all fields populated correctly

- [ ] Task 4: Expand SessionTracker tests (AC: #10)
  - [ ] 4.1: Add test for RecordDoorViewed (first door captures time-to-first-door, subsequent doesn't overwrite)
  - [ ] 4.2: Add test for RecordDoorSelection calling RecordDoorViewed (verify DoorsViewed increments correctly, calling both separately causes double-increment - document this is by design)
  - [ ] 4.3: Add test for RecordDetailView
  - [ ] 4.4: Add test for RecordNoteAdded
  - [ ] 4.5: Add test for RecordStatusChange
  - [ ] 4.6: Add test for RecordTaskCompleted
  - [ ] 4.7: Add test for empty TaskBypasses (RecordRefresh with empty slice - verify no nil entry appended)

- [ ] Task 5: Create analysis scripts (AC: #8)
  - [ ] 5.1: Create `scripts/analyze_sessions.sh` - reads sessions.jsonl, handles zero-division for empty door_selections array, outputs session count, average duration, average completions, average refreshes, door position distribution, mood distribution
  - [ ] 5.2: Create `scripts/daily_completions.sh` - reads completed.txt (format: `[YYYY-MM-DD HH:MM:SS] task_id | task_text`), aggregates by date, shows daily counts and trends. MUST read `file_manager.go` AppendCompleted() to verify actual format.
  - [ ] 5.3: Create `scripts/validation_decision.sh` - evaluates: min 5 sessions, avg completion > 0, avg time-to-first-door < 10s, engagement = percentage of sessions where `detail_views > 0` >= 50%, outputs PASS/FAIL with per-criterion reasoning
  - [ ] 5.4: Make all scripts executable (chmod +x)

- [ ] Task 6: Add Makefile targets and script tests (AC: #12)
  - [ ] 6.1: Add `make analyze` target that runs all three scripts against `~/.threedoors/` data (with `THREEDOORS_DIR` env var override for custom paths)
  - [ ] 6.2: Add `make test-scripts` target that runs all three scripts against `scripts/testdata/` fixtures and verifies non-empty output and exit code 0
  - [ ] 6.3: Verify `make analyze` and `make test-scripts` both work

- [ ] Task 7: Verify integration and performance (AC: #6, #7)
  - [ ] 7.1: Run `go test ./...` - all tests pass
  - [ ] 7.2: Run `golangci-lint run ./...` - zero warnings
  - [ ] 7.3: Run `make test-scripts` - all script tests pass
  - [ ] 7.4: Verify sessions.jsonl output is valid via `jq . sessions.jsonl`

## Dev Notes

### What Already Exists (DO NOT RECREATE)

The core session tracking infrastructure was implemented as part of Stories 1.3 and 1.4. The following files exist and are fully functional:

- **`internal/tasks/session_tracker.go`** - Complete SessionTracker with all recording methods
- **`internal/tasks/session_tracker_test.go`** - Basic tests (5 test functions)
- **`internal/tasks/metrics_writer.go`** - Complete MetricsWriter with AppendSession
- **`cmd/threedoors/main.go`** - Already calls tracker.Finalize() and writer.AppendSession() on exit
- **`internal/tui/main_model.go`** - MainModel already has `tracker *tasks.SessionTracker` field
- **`internal/tui/doors_view.go`** - Already calls RecordDoorSelection, RecordRefresh
- **`internal/tui/detail_view.go`** - Already calls RecordDetailView, RecordStatusChange, RecordTaskCompleted
- **`internal/tui/mood_view.go`** - Already calls RecordMood (via MoodCapturedMsg in MainModel)
- **`internal/tui/search_view.go`** - Already has :stats command showing session metrics

### What This Story MUST Create

1. **Three shell scripts** in `scripts/` directory
2. **Shell script test fixtures** in `scripts/testdata/`
3. **MetricsWriter test file** `internal/tasks/metrics_writer_test.go` (with reusable `testSessionMetrics()` helper)
4. **Integration test** for SessionTracker → MetricsWriter pipeline
5. **Expanded SessionTracker tests** in existing `internal/tasks/session_tracker_test.go`
6. **Makefile targets** for `analyze` and `test-scripts`

### Critical Design Decisions (DO NOT CHANGE)

- **Crash data loss is ACCEPTABLE** for Tech Demo phase. No signal handler needed. Session data is only persisted on clean exit via Bubbletea quit flow. Do NOT add SIGTERM/SIGINT handlers.
- **Concurrency is NOT in scope.** Bubbletea's MVU is single-threaded. Do NOT add mutexes to SessionTracker.
- **MetricsWriter intentionally skips fsync.** It uses `os.O_APPEND` mode for an append-only log. This is correct — do NOT "fix" it by adding atomic write pattern. The architecture doc's fsync mandate applies to `tasks.yaml` (read-write), not append-only logs.
- **RecordDoorSelection calls RecordDoorViewed internally.** Calling both separately WILL double-increment DoorsViewed. This is by design — direct RecordDoorViewed() is for legacy/fallback and RecordDoorSelection() is the preferred entry point.

### completed.txt Format (VERIFY BEFORE SCRIPTING)

The `daily_completions.sh` script parses `~/.threedoors/completed.txt`. Before writing the script, **read `internal/tasks/file_manager.go` method `AppendCompleted()`** to verify the exact line format. Expected format per architecture docs:
```
[YYYY-MM-DD HH:MM:SS] task_id | task_text
```
If the actual implementation differs, use the actual format.

### Architecture Compliance

- **File placement:** Scripts go in `scripts/` at project root
- **Test files:** `*_test.go` alongside source files in same package
- **Go version:** 1.25.4+
- **Formatting:** `gofumpt` before commit
- **Linting:** `golangci-lint run ./...` must pass
- **Test style:** Table-driven tests preferred
- **Temp dirs:** Use `t.TempDir()` for test files
- **Error wrapping:** Use `%w` verb in fmt.Errorf
- **Timestamps:** Always UTC (`time.Now().UTC()`)

### Script Requirements

All scripts must:
- Use `#!/usr/bin/env bash` shebang
- Use `set -euo pipefail` for safety
- Check for required tools (jq) and files
- Output human-readable summaries
- Exit cleanly with appropriate codes
- Handle empty/missing data files gracefully

**`scripts/analyze_sessions.sh`** should output:
```
ThreeDoors Session Analysis
===========================
Total sessions: N
Average duration: X.X minutes
Average tasks completed: X.X per session
Average refreshes: X.X per session
Average detail views: X.X per session
Average time to first door: X.Xs
Door position preference: Left=N% Center=N% Right=N%
Mood distribution: Focused=N Tired=N ...
```

**`scripts/daily_completions.sh`** should output:
```
Daily Completion Report
=======================
Date       | Completed | Cumulative
2026-03-01 | 5         | 5
2026-03-02 | 3         | 8
...
```

**`scripts/validation_decision.sh`** should evaluate:
- Minimum 5 sessions recorded
- Average completion rate > 0 tasks/session
- Average time-to-first-door < 10 seconds
- Engagement: percentage of sessions where `detail_views > 0` must be >= 50%
- Output: PASS/FAIL with per-criterion pass/fail reasoning

### Script Edge Cases (CRITICAL)

All scripts must handle these edge cases without crashing:
- **Empty/missing data file**: Print "No data found" and exit 0
- **Zero-division**: When calculating averages over empty arrays (e.g., door_selections is empty), default to 0 or "N/A"
- **Sessions with -1 time_to_first_door**: These are sessions where no door was selected. Filter them from time-to-first-door averages.
- **Mixed data quality**: Some sessions may have empty arrays for door_selections, task_bypasses, or mood_entries_detail. Handle gracefully.

### Library & Framework Requirements

- **jq** is required for shell scripts (JSON processing)
- **Go standard library** `testing` package for tests
- **`t.TempDir()`** for file-based test isolation
- **`encoding/json`** for JSON marshaling verification in tests
- No new Go dependencies needed

### File Structure

```
ThreeDoors/
├── scripts/                          # NEW directory
│   ├── analyze_sessions.sh           # NEW - session analysis
│   ├── daily_completions.sh          # NEW - daily completion counts
│   ├── validation_decision.sh        # NEW - validation criteria check
│   └── testdata/                     # NEW - test fixtures for scripts
│       ├── sessions.jsonl            # NEW - 6+ sample sessions with known values
│       └── completed.txt             # NEW - sample completions with known dates
├── internal/tasks/
│   ├── metrics_writer.go             # EXISTS - no changes needed
│   ├── metrics_writer_test.go        # NEW - MetricsWriter tests + integration test
│   ├── session_tracker.go            # EXISTS - no changes needed
│   └── session_tracker_test.go       # EXISTS - expand with more tests
└── Makefile                          # EXISTS - add 'analyze' and 'test-scripts' targets
```

### Makefile Target Specifications

**`make analyze`**: Runs all three analysis scripts against real user data. Accepts `THREEDOORS_DIR` env var to override default `~/.threedoors/` path.
```makefile
THREEDOORS_DIR ?= $(HOME)/.threedoors
analyze:
	@./scripts/analyze_sessions.sh $(THREEDOORS_DIR)/sessions.jsonl
	@./scripts/daily_completions.sh $(THREEDOORS_DIR)/completed.txt
	@./scripts/validation_decision.sh $(THREEDOORS_DIR)/sessions.jsonl
```

**`make test-scripts`**: Runs all three scripts against `scripts/testdata/` fixtures, verifies exit code 0 and non-empty output.

### jq Dependency Note

All analysis scripts require `jq`. GitHub Actions runners (ubuntu-latest) include `jq` by default. Scripts must check for `jq` at start and exit with clear error message if missing: `"Error: jq is required but not installed. Install with: brew install jq"`

### Testing Requirements

- **Regression baseline FIRST:** Run `go test ./...` BEFORE any changes to confirm green baseline
- All new tests must pass: `go test ./...`
- MetricsWriter tests must use `t.TempDir()` for isolation
- Table-driven tests for multiple scenarios
- Test that JSON output is valid and parseable
- Coverage target: 70%+ for internal/tasks
- **All Go tests stay in `package tasks`** (same package, not external `tasks_test`) to access private fields — matches existing pattern in `session_tracker_test.go`
- **Integration test goes in `metrics_writer_test.go`** since MetricsWriter is the pipeline endpoint
- **Expanded SessionTracker tests go in existing `session_tracker_test.go`** — do NOT create a new file
- **Shell script tests are Makefile-based** (bash, not Go) — run script against testdata, check exit code 0 and non-empty stdout
- **Makefile `test-scripts` must `chmod +x scripts/*.sh`** before running (Git may not preserve execute bits)

### Canonical Test Data Values

Use these EXACT values in BOTH the Go `testSessionMetrics()` helper AND `scripts/testdata/sessions.jsonl` fixtures for consistency:

**Session A (typical usage):**
- session_id: "test-session-aaa"
- duration_seconds: 300 (5 minutes)
- tasks_completed: 2
- doors_viewed: 5
- refreshes_used: 3
- detail_views: 2
- time_to_first_door_seconds: 1.5
- door_selections: [{position: 0, text: "Task A"}, {position: 2, text: "Task B"}]
- mood_entries_detail: [{mood: "Focused"}]

**Session B (power user):**
- session_id: "test-session-bbb"
- duration_seconds: 900 (15 minutes)
- tasks_completed: 5
- doors_viewed: 12
- refreshes_used: 7
- detail_views: 6
- time_to_first_door_seconds: 0.8
- door_selections: [{position: 1}, {position: 0}, {position: 2}, {position: 1}, {position: 1}]
- mood_entries_detail: [{mood: "Energized"}, {mood: "Focused"}]

**Session C (quick browse, no selection):**
- session_id: "test-session-ccc"
- duration_seconds: 30
- tasks_completed: 0
- doors_viewed: 0
- refreshes_used: 1
- detail_views: 0
- time_to_first_door_seconds: -1 (no door selected)
- door_selections: [] (empty)
- mood_entries_detail: []

**Session D-F:** Additional sessions with varied values to reach minimum 6 for validation_decision.sh testing.

**`scripts/testdata/completed.txt` fixture data:**
```
[2026-03-01 10:15:30] uuid-001 | Write architecture document
[2026-03-01 11:20:00] uuid-002 | Review PRD
[2026-03-01 14:30:45] uuid-003 | Implement Story 1.1
[2026-03-02 09:00:00] uuid-004 | Fix login bug
[2026-03-02 16:45:00] uuid-005 | Update dependencies
```

### :stats Command vs analyze_sessions.sh Scope

The `:stats` command in search_view.go shows **current session** metrics only (completed, detail views, refreshes). The `analyze_sessions.sh` script shows **aggregate metrics across all sessions**. These are intentionally different scopes — `:stats` is real-time, scripts are historical analysis. No consistency issue.

### Previous Story Intelligence

Stories 1.1-1.4 established:
- Two-layer architecture (TUI + Domain) - DO NOT add new packages
- Constructor injection pattern for dependencies
- Bubbletea MVU pattern with custom message types
- Atomic file writes for persistence (but MetricsWriter uses append mode which is correct for append-only log)
- `gofumpt` formatting enforced
- Table-driven test pattern

### Git Intelligence

Recent commits show:
- Story 1.4 added search_view.go with :stats command already showing metrics
- CI/CD pipeline added with GitHub Actions
- Test coverage reporting added
- All code passes gofumpt and golangci-lint

### JSON Lines Format Reference

Each line in `sessions.jsonl` looks like:
```json
{"session_id":"uuid","start_time":"2026-03-02T06:00:00Z","end_time":"2026-03-02T06:15:00Z","duration_seconds":900,"tasks_completed":3,"doors_viewed":10,"refreshes_used":5,"detail_views":4,"notes_added":1,"status_changes":5,"mood_entries":2,"time_to_first_door_seconds":2.5,"door_selections":[{"timestamp":"...","door_position":1,"task_text":"..."}],"task_bypasses":[["task1","task2","task3"]],"mood_entries_detail":[{"timestamp":"...","mood":"Focused","custom_text":""}]}
```

### References

- [Source: docs/prd/epic-details.md#story-15-session-metrics-tracking] - Full AC specification
- [Source: docs/architecture/data-storage-schema.md] - Storage format standards
- [Source: docs/architecture/test-strategy-and-standards.md] - Test patterns and coverage targets
- [Source: docs/architecture/coding-standards.md] - Coding rules and naming conventions
- [Source: docs/architecture/source-tree.md] - File placement rules
- [Source: docs/prd/epics-and-stories.md#story-15] - Original story with full BDD criteria

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

### Completion Notes List

### File List
