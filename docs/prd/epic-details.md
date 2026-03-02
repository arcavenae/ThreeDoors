# Epic Details

## Epic 1: Three Doors Technical Demo

**Epic Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists.

**Scope:** CLI/TUI application that reads tasks from a text file, presents three random tasks as "doors," allows refresh and selection, marks tasks complete, and tracks progress.

**Story Sequence Optimization:** Stories reordered to validate refresh UX before completion (moved 1.6→1.4). Non-essential features simplified or made optional to focus on core validation.

---

### Story 1.1: Project Setup & Basic Bubbletea App

**As a** developer,
**I want** a working Go project with Bubbletea framework,
**so that** I have a foundation for building the Three Doors TUI.

**Acceptance Criteria:**
1. Go module initialized with `go mod init github.com/arcaven/ThreeDoors`
2. Bubbletea and Lipgloss dependencies added
3. Basic TUI application renders "ThreeDoors - Technical Demo" header
4. Application responds to 'q' keypress to quit
5. `Makefile` with `build`, `run`, and `clean` targets works
6. Application compiles and runs without errors

**Estimated Time:** 30-45 minutes

---

### Story 1.2: Display Three Doors from a Task File

**As a** developer,
**I want** the application to read tasks from a text file and display three of them as "doors",
**so that** I can see the core interface of the application.

**Acceptance Criteria:**
1. On startup, the application reads tasks from `~/.threedoors/tasks.txt`
2. If `tasks.txt` does not exist, it is created with a default set of 3-5 sample tasks
3. A `FileManager` component is created in `internal/tasks/file_manager.go` to handle file reading and creation
4. A `Task` model is defined in `internal/tasks/task.go` to represent a task (text content)
5. The main application view displays three randomly selected tasks from the file
6. The display shows three randomly selected tasks, distributed horizontally from left to right, without "Door X" labels
7. The application responds to 'a' or 'left arrow' for the left door, 'w' or 'up arrow' for the middle/center door, and 'd' or 'right arrow' for the right door
8. Pressing 's' or 'down arrow' re-rolls the doors, presenting a new set of three tasks
9. Initially, or after re-rolling, no door is selected/active
10. The application responds to the following keystrokes for task management (functionality to be implemented in future stories):
    - 'c': Mark selected task as complete
    - 'b': Mark selected task as blocked
    - 'i': Mark selected task as in progress
    - 'e': Expand selected task (into more tasks)
    - 'f': Fork selected task (clone/split)
    - 'p': Procrastinate/avoid selected task
11. The application still quits on 'q' or 'ctrl+c'

**Key Design Decisions:**
- Three doors are rendered horizontally, each occupying approximately 1/3rd of the terminal width (dynamic adjustment)
- No "Door X" labels displayed to reduce visual clutter
- No door is selected initially or after re-rolling to avoid bias
- Arrow keys provided as alternative to WASD for accessibility

**Estimated Time:** 60-90 minutes

---

### Story 1.3: Door Selection & Task Status Management

**As a** user,
**I want** to select a door and update the task's status,
**so that** I can take action on tasks and track my progress.

**Acceptance Criteria:**
1. Pressing A/Left Arrow, W/Up Arrow, or D/Right Arrow selects the corresponding door (left, center, right)
2. Selected task is highlighted/indicated visually
3. **Door Opening Animation & Expanded Detail View:**
   - When door is selected (or Enter is pressed), door presents optional animation as if opening
   - Selected door shifts to left position and expands to fill the screen
   - Task detail view displays:
     - Task text (full, not truncated)
     - Any existing task metadata/details (status, notes, timestamps, etc.)
     - Status action menu with all available options
   - **Esc** key closes the door and returns to previous screen (context-aware):
     - If opened from three doors view → returns to three doors view
     - If opened from search (Story 1.3a) → returns to search view with text preserved and previous selection highlighted
4. Status action menu shows available options:
   - **C**: Mark as Complete
   - **B**: Mark as Blocked
   - **I**: Mark as In Progress
   - **E**: Expand (break into more tasks)
   - **F**: Fork (clone/split task)
   - **P**: Procrastinate (defer task)
   - **R**: Flag for Rework
   - **M**: Log Mood/Context (also available from door view without selection)
   - **Esc**: Close door and return to previous screen
5. Pressing any status key (within expanded detail view) applies that status to the selected task
6. Pressing **M** from door view (no task selection needed) opens mood capture dialog:
   - Multiple choice options: "Focused", "Tired", "Stressed", "Energized", "Distracted", "Calm", "Other"
   - If "Other" selected, prompt for custom text input (word or phrase)
   - Mood entries are timestamped and recorded in session metrics
   - Returns to door view immediately after capture
7. Completed tasks are removed from available task pool (in-memory) and appended to `~/.threedoors/completed.txt` with timestamp
8. Blocked/deferred/rework tasks remain in the pool but are tagged with status
9. New set of three doors is displayed automatically after any status change (door closes and returns to three doors view)
10. Session completion count increments for completed tasks (e.g., "✓ Completed this session: 3")
11. "Progress over perfection" message shown after completing a task
12. Door selection is tracked: which door position (left/center/right) was selected
13. Task bypass is tracked: doors shown but not selected before refresh
14. Mood entries are tracked with timestamps for correlation with task selection patterns

**Key Design Decisions:**
- **Door opening animation is optional** but provides delightful visual feedback
- **Expanded detail view** shifts door left and fills screen for focused interaction
- **Context-aware return** (Esc) maintains navigation state (critical for search integration in 1.3a)
- All status changes should be tracked for future pattern analysis
- Door position preferences (left vs center vs right) captured for learning
- Tasks that are expanded or forked create new task entries in tasks.txt
- Blocked tasks should prompt for optional blocker note
- **Mood tracking is casual and low-friction** - accessible anytime via 'M' key without needing task selection
- Mood data provides crucial context for correlating emotional state with task selection behavior
- Multiple choice moods keep capture quick; custom text option allows nuanced expression

**Estimated Time:** 90-120 minutes (including door animation, expanded detail view, mood capture UI)

---

### Story 1.3a: Quick Search & Command Palette

**As a** user,
**I want** to quickly search for specific tasks and execute commands via a text input interface,
**so that** I can efficiently find and act on tasks without scrolling through the three doors.

**Acceptance Criteria:**

**Search Mode (Default):**
1. Pressing **/** key from three doors view opens text input field at bottom of screen
2. Text input appears with placeholder text: "Search tasks... (or :command for commands)"
3. As user types, live substring matching filters task list:
   - Matching tasks display from **bottom-up** extending up the screen
   - List updates continuously as each character is entered
   - If no matches found, show message: "No tasks match '<search text>'"
   - Empty input shows no results (doesn't show all tasks)
4. **Navigation within search results:**
   - **Arrow keys** (up/down/left/right): Navigate through search results
   - **A/S/D/W keys**: Navigate through search results (s=down, w=up, a/d reserved for future horizontal navigation)
   - **H/J/K/L keys** (vi-style): Navigate through search results (j=down, k=up, h/l reserved)
   - Selected result is highlighted
   - Enter key: Opens selected task in expanded detail view (same as Story 1.3 door selection)
5. **Task detail from search:**
   - When task is opened via Enter, shows same expanded detail view as Story 1.3
   - **Esc** from detail view returns to search with:
     - Search text preserved in input field
     - Previously selected task still highlighted
     - User can continue searching or refine search
6. **Exit search mode:**
   - **Esc** key (when in search input, not in task detail) clears search and returns to three doors view
   - **Ctrl+C** also exits search mode

**Command Mode (vi-style):**
7. Typing **:** as first character in empty text input switches to command mode
8. Command mode indicated by visual cue (e.g., prompt changes to ":")
9. **Available commands:**
   - **:add <task text>** - Add new task to tasks.txt
   - **:edit** - Edit current task list file directly (opens in $EDITOR or vim)
   - **:mood [mood]** - Quick mood log (optional mood parameter, otherwise prompts)
   - **:stats** - Show session statistics (tasks completed, doors viewed, time in session, etc.)
   - **:chat** - Open AI chat interface for task-related questions/help (deferred implementation)
   - **:help** - Display available commands and key bindings
   - **:quit** or **:exit** - Exit application
10. Commands execute on Enter key
11. Invalid commands show error: "Unknown command: '<command>'. Type :help for available commands."
12. **Esc** exits command mode and returns to three doors view

**Key Design Decisions:**
- **Bottom-up list display** reduces eye travel distance from input field
- **Multiple navigation schemes** (arrows, WASD, HJKL) accommodate different user preferences
- **Live substring matching** provides instant feedback
- **Context preservation** (search text + selection) when returning from task detail is critical for UX
- **Command palette** (`:`) provides power-user features without cluttering main UI
- **:chat command deferred** to post-validation (placeholder for future AI integration)
- Search only matches task text (no metadata search in Tech Demo phase)

**Estimated Time:** 90-120 minutes (search mode + navigation + command parsing)

---

### Story 1.5: Session Metrics Tracking

**As a** developer validating the Three Doors concept,
**I want** objective session metrics collected automatically,
**so that** I can make a data-informed decision at the validation gate instead of relying solely on subjective impressions.

**Priority:** HIGH (enables objective validation)

**Context:** The validation decision gate asks: "Does Three Doors reduce friction vs. traditional task lists?" Without metrics, this is purely subjective. This story adds lightweight, non-intrusive tracking to provide objective evidence. The metrics should be invisible to the user during normal operation - no prompts, no UI changes, just silent background collection.

**Future Pattern Analysis Foundation:** Capturing door selection patterns (left vs center vs right), bypass behaviors (which tasks are skipped), status change patterns (what gets blocked, procrastinated, or completed), and **mood/emotional context** creates the data foundation for Epic 4 (Learning & Intelligent Door Selection). Over time, the application will use this data to:
- Predict which types of tasks the user tends to skip or avoid
- Identify patterns that indicate blockers or discouragement
- **Correlate emotional state with task selection behavior** (e.g., "stressed" → avoids complex tasks)
- Surface insights to help users understand their work patterns
- Adapt door selection to encourage balanced progress across different task types
- Help users re-evaluate goals when persistent avoidance patterns emerge

**Acceptance Criteria:**

1. **SessionTracker component created** in `internal/tasks/session_tracker.go`
   - Tracks session_id, start/end times, duration
   - Tracks behavioral counters: tasks_completed, doors_viewed, refreshes_used, detail_views, notes_added, status_changes, mood_entries
   - Tracks time_to_first_door_seconds (key friction metric)
   - **NEW: Door selection patterns** - tracks which door position selected (left=0, center=1, right=2) per selection
   - **NEW: Task bypass tracking** - records tasks shown in doors but not selected before refresh
   - **NEW: Status change details** - records what status was applied (complete, blocked, in-progress, expand, fork, procrastinate, rework)
   - **NEW: Task content capture** - stores task text with each interaction for future pattern analysis
   - **NEW: Mood tracking** - captures timestamped mood entries (predefined or custom text) for correlation with task behavior
   - Constructor `NewSessionTracker()` initializes with UUID and current timestamp
   - Methods: `RecordDoorViewed(doorPosition int)`, `RecordRefresh(doorTasks []string)`, `RecordDetailView()`, `RecordTaskCompleted(taskText string)`, `RecordNoteAdded()`, `RecordStatusChange(status string, taskText string)`, `RecordDoorSelection(doorPosition int, taskText string)`, `RecordMood(mood string, customText string)`
   - Method `Finalize()` calculates duration and returns metrics for persistence
   - Mood entries stored as array: `[{timestamp, mood, custom_text}]`

2. **MetricsWriter component created** in `internal/tasks/metrics_writer.go`
   - Constructor `NewMetricsWriter(baseDir string)` sets path to sessions.jsonl
   - Method `AppendSession(metrics *SessionMetrics)` writes JSON line to file
   - Creates file if doesn't exist, appends to existing file
   - Returns error on failure (caller logs warning, doesn't crash)

3. **SessionTracker integrated into MainModel**
   - MainModel includes sessionTracker field
   - SessionTracker passed to DoorsView and TaskDetailView constructors
   - No UI changes (completely invisible to user)

4. **Recording calls integrated into DoorsView**
   - Door selection (1/2/3) calls `RecordDoorViewed()`
   - Refresh (R) calls `RecordRefresh()`
   - Mood capture (M) calls `RecordMood()` with selected mood and optional custom text
   - Recording happens before transitioning to detail view

5. **Recording calls integrated into TaskDetailView**
   - Constructor calls `RecordDetailView()` on initialization
   - Note addition calls `RecordNoteAdded()`
   - Status change calls `RecordStatusChange()`
   - Completion calls both `RecordStatusChange()` and `RecordTaskCompleted()`

6. **Session persistence on app exit**
   - `cmd/threedoors/main.go` calls `Finalize()` and `AppendSession()` on clean exit
   - Write failures logged as warning to stderr, don't block exit
   - File created: `~/.threedoors/sessions.jsonl` (JSON Lines format)

7. **Metrics file format validated**
   - Each line is valid JSON (parseable by `jq`)
   - File is append-only, human-readable
   - Manual verification: run app 2-3 times, verify lines in sessions.jsonl

8. **Performance requirements met**
   - Recording adds <1ms overhead per event
   - No UI lag or stuttering observed
   - Memory overhead negligible (<1KB per session)

9. **Error handling implemented**
   - File write failures don't crash app (warning logged)
   - JSON marshal failures don't crash app
   - No error dialogs shown to user

**Analysis Scripts Created:**
- `scripts/analyze_sessions.sh` - Session summary and averages
- `scripts/daily_completions.sh` - Daily completion counts from completed.txt
- `scripts/validation_decision.sh` - Automated validation criteria evaluation

**Deferred to Epic 4 (Learning & Intelligent Door Selection):**
- Pattern analysis algorithms (which task types are consistently avoided)
- ML-based task selection optimization
- User insight reports ("You tend to skip tasks containing X")
- Adaptive door selection based on historical patterns
- Goal re-evaluation prompts when persistent avoidance detected
- Task categorization and tagging for pattern recognition

**Deferred to Future:**
- Daily aggregation reports (manual analysis via scripts)
- In-app metrics visualization
- Friction score prompts (manual logging recommended)
- Metrics export to other formats
- Historical cleanup/rotation

**Rationale:** Provides objective data to answer "Does Three Doors reduce friction?" Metrics enable data-informed decision at validation gate rather than relying solely on subjective feel. Enhanced tracking of door selection patterns, task bypass behaviors, and **mood/emotional context** creates the data foundation needed for future learning and adaptation features (Epic 4). Mood correlation will help identify which emotional states lead to productive task selection vs avoidance.

**Estimated Time:** 85-100 minutes (enhanced tracking + mood capture)

---

### Story 1.6: Essential Polish (SIMPLIFIED)

**As a** user,
**I want** the app to feel polished enough to use daily,
**so that** I enjoy the validation experience.

**Acceptance Criteria:**
1. Lipgloss styling applied: distinct colors for doors, success messages (green), prompts (yellow/blue)
2. "Progress over perfection" message embedded in interface (startup greeting or post-completion)
3. Application feels responsive and smooth (no noticeable lag)

**Deferred to MVP:**
- README.md (you're the only user for validation)
- Extensive edge case handling (all tasks completed celebration, 1-2 tasks remaining display logic)
- Advanced error messaging

**Rationale for Simplification:** Focus on making the core experience pleasant. Edge cases are unlikely to be hit during 1-week validation. README isn't needed when you built it.

**Estimated Time:** 20-30 minutes

---

## Epic 2: Foundation & Apple Notes Integration ✅ COMPLETE

**Status:** All 6 stories implemented and merged. See `docs/sprint-status-report.md` for details.
- Story 2.1: Architecture Refactoring - Adapter Pattern (PR #20)
- Story 2.2: Apple Notes Integration Spike (PR #22)
- Story 2.3: Read Tasks from Apple Notes (PR #17)
- Story 2.4: Write Task Updates to Apple Notes (PR #21)
- Story 2.5: Bidirectional Sync Engine (PR #15)
- Story 2.6: Health Check Command (PR #19)

## Epic 3: Enhanced Interaction & Task Context ✅ COMPLETE

**Status:** All 7 stories implemented and merged. See `docs/sprint-status-report.md` for details.
- Story 3.1: Quick Add Mode (PR #23)
- Story 3.2: Extended Task Capture with Context (PR #24)
- Story 3.3: Values & Goals Setup and Display (PR #25)
- Story 3.4: Door Feedback Options (PR #27)
- Story 3.5: Daily Completion Tracking & Comparison (PR #28)
- Story 3.6: Session Improvement Prompt (PR #29)
- Story 3.7: Enhanced Navigation & Messaging (PR #31)

---

## Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging) 🆕

**Epic Goal:** Refactor core architecture, harden adapters, establish test infrastructure, and resolve technical debt from rapid Epic 1-3 implementation. Prepares the codebase for Epic 4+ by establishing Architecture v2.0 foundations.

**Origin:** Party mode bridging discussion (2026-03-02)
**Prerequisites:** Epic 3 complete ✅
**Blocks:** Epic 4 (partially), Epic 7, Epic 8, Epic 9, Epic 11

### Story 3.5.1: Core Domain Extraction

**As a** developer,
**I want** `internal/tasks` split into `internal/core` and separate adapter packages,
**So that** the architecture follows the five-layer design and enables the Plugin SDK (Epic 7).

**Acceptance Criteria:**
1. `internal/core/` contains: TaskPool, DoorSelector, StatusManager, SessionTracker
2. `internal/adapters/textfile/` contains the YAML file adapter
3. `internal/adapters/applenotes/` contains the Apple Notes adapter
4. `internal/tui/` depends only on `internal/core/` (dependency inversion)
5. All existing tests pass (behavior-preserving refactor)
6. No user-facing behavior changes

### Story 3.5.2: TaskProvider Interface Hardening

**As a** developer building future integrations,
**I want** the TaskProvider interface formalized with Watch(), HealthCheck(), ChangeEvent,
**So that** the adapter SDK (Epic 7) has a stable contract.

**Acceptance Criteria:**
1. `TaskProvider` interface includes: Name(), Load(), Save(), Delete(), Watch(), HealthCheck()
2. `ChangeEvent` struct defined with Type, TaskID, Task, Source fields
3. Contract test stubs created in `internal/adapters/contract_test.go`
4. Existing adapters updated to implement hardened interface
5. Interface documented with godoc comments

### Story 3.5.3: Config.yaml Schema & Migration Spike

**As a** developer,
**I want** a spike on config.yaml schema and migration path,
**So that** Epic 7's config-driven provider selection has a validated foundation.

**Acceptance Criteria:**
1. `docs/spikes/config-schema.md` documents proposed schema, migration path
2. Zero-friction upgrade verified (no config.yaml defaults to current behavior)
3. Sample config.yaml drafted with commented provider examples
4. Breaking changes identified with mitigation strategies

### Story 3.5.4: Apple Notes Adapter Hardening

**As a** user relying on Apple Notes sync,
**I want** the adapter to handle errors gracefully with timeouts and retries,
**So that** sync is reliable before more adapters land.

**Acceptance Criteria:**
1. All AppleScript calls have configurable timeout (default: 10s)
2. Transient failures retry with exponential backoff (max 3 retries)
3. Errors categorized: transient, permanent, configuration
4. Error messages are user-friendly and actionable
5. No sensitive data in adapter logs (NFR9)

### Story 3.5.5: Baseline Regression Test Suite

**As a** developer preparing for Epic 4,
**I want** baseline tests for current door selection and task management,
**So that** the learning engine can be validated against known-good behavior.

**Acceptance Criteria:**
1. Table-driven tests for random selection, Fisher-Yates, ring buffer, edge cases
2. Status management tests for all state transitions
3. Task pool tests for load, filter, add, remove, update
4. Tests pass on current codebase

### Story 3.5.6: Session Metrics Reader Library

**As a** developer building Epic 4,
**I want** a reusable library for reading session metrics,
**So that** Epic 4 stories can focus on learning logic.

**Acceptance Criteria:**
1. `internal/core/metrics/reader.go` with ReadAll(), ReadSince(), ReadLast() methods
2. Returns typed SessionMetrics structs
3. Handles corrupted lines gracefully
4. Unit tests cover empty, single, multiple sessions, corrupted data

### Story 3.5.7: Adapter Test Scaffolding & CI Coverage Floor

**As a** developer,
**I want** test infrastructure and CI coverage enforcement,
**So that** Epic 9 has a foundation and coverage doesn't erode.

**Acceptance Criteria:**
1. `testdata/` directory with sample adapter test data
2. `internal/testing/` with mock/stub helpers
3. CI measures coverage and fails below threshold
4. Coverage report posted as PR comment

### Story 3.5.8: Validation Gate Decision Documentation

**As the** product team,
**I want** Phase 1 validation results documented,
**So that** the proceed-to-MVP decision is recorded.

**Acceptance Criteria:**
1. `docs/validation-gate-results.md` with validation period, usage patterns, evidence
2. UX lessons learned captured
3. Formal proceed-to-MVP decision with rationale
4. Recommendations for Epic 4 based on observed patterns

---

**Epic 4: Learning & Intelligent Door Selection**

**Goal:** Use historical session metrics (captured in Story 1.4) to analyze user patterns and adapt door selection to improve task engagement and completion rates.

**Key Capabilities to Build:**
- **Pattern Recognition:** Analyze which types of tasks are consistently selected vs bypassed
- **Mood Correlation Analysis:** Identify which emotional states (focused, stressed, tired, etc.) correlate with task selection, avoidance, or completion patterns
- **Avoidance Detection:** Flag tasks or task patterns that are repeatedly shown but never selected
- **Status Pattern Analysis:** Track which task types tend to get blocked, procrastinated, or reworked (correlated with mood state)
- **Adaptive Selection:** Adjust door algorithm based on current mood state and historical patterns (e.g., show simpler tasks when user logs "tired")
- **User Insights:** Surface reports like "When stressed, you tend to avoid complex technical tasks" or "Your highest completion rate is when feeling 'focused'"
- **Goal Re-evaluation Prompts:** When persistent avoidance detected (especially with specific mood patterns), suggest user review if tasks align with goals
- **Encouragement System:** Gently encourage work on task types that haven't been touched in a while, with mood-aware messaging
- **Position Preference Analysis (Minor):** Track if user has bias toward certain door positions (left/center/right)

**Data Foundation:** Epic 1 Story 1.4 creates the metrics infrastructure capturing door position selections, task bypasses, status changes, and **mood/emotional context**—all essential for pattern analysis. Mood tracking enables correlation between emotional state and work behavior, allowing adaptive task selection based on current user state.

*Detailed stories to be defined based on sufficient usage data from Epic 3*

**Epic 5: macOS Distribution & Packaging**

**Goal:** Provide a trusted, seamless installation experience on macOS by signing, notarizing, and packaging the binary so macOS Gatekeeper does not quarantine it on download. This removes the biggest adoption friction for new users.

**Context:** Currently, Go binaries built in CI are unsigned and untrusted. When users download them from GitHub Releases, macOS Gatekeeper quarantines the file, requiring users to manually allow execution via System Preferences > Security & Privacy. This is a poor first-run experience that undermines trust and creates unnecessary friction.

**Independence:** This epic has no dependencies on other feature epics and can be implemented at any time. It is a cross-cutting infrastructure concern.

**Stories:**

### Story 5.1: CI Code Signing & Notarization

**As a** macOS user downloading ThreeDoors,
**I want** the binary to be signed and notarized,
**so that** macOS Gatekeeper allows execution without security warnings or quarantine.

**Acceptance Criteria:**
1. CI pipeline signs darwin/arm64 and darwin/amd64 binaries with a valid Apple Developer ID Application certificate
2. Signed binaries are submitted to Apple's notarization service and stapled
3. `codesign --verify` and `spctl --assess` pass on the resulting binaries
4. GitHub Releases contain only signed+notarized macOS binaries (Linux binaries remain unsigned)
5. Signing secrets (certificate, password, Apple ID credentials, team ID) are stored as GitHub Actions encrypted secrets
6. Signing step fails gracefully with clear error if secrets are not configured (e.g., in forks)

**Implementation Guidance:**
- Use `gon` or direct `codesign`/`xcrun notarytool` in CI
- Apple Developer ID Application certificate exported as .p12, base64-encoded in secrets
- Notarization requires Apple ID with app-specific password and Team ID
- Consider using `macos-latest` runner for the signing step (codesign requires macOS)

### Story 5.2: Homebrew Tap Formula

**As a** macOS user,
**I want** to install ThreeDoors via `brew install arcaven/tap/threedoors`,
**so that** I get automatic updates and a standard macOS installation experience.

**Acceptance Criteria:**
1. A separate GitHub repository `arcaven/homebrew-tap` is created with a Homebrew formula
2. Formula downloads the correct signed binary for the user's architecture (arm64 or amd64)
3. Formula includes SHA256 checksums for integrity verification
4. `brew install arcaven/tap/threedoors` installs the binary to the Homebrew prefix
5. `brew upgrade arcaven/tap/threedoors` upgrades to the latest version
6. CI pipeline automatically updates the Homebrew formula on new releases (SHA256 and version)

**Implementation Guidance:**
- Homebrew formula is a Ruby file in the `homebrew-tap` repo
- Use `on_arm` / `on_intel` blocks for architecture-specific URLs
- CI can use `brew bump-formula-pr` or directly update the formula file via GitHub API
- Include `test` block in formula that runs `threedoors --version` or `threedoors health`

### Story 5.3: DMG/pkg Installer

**As a** macOS user who prefers graphical installation,
**I want** a DMG or pkg installer available for download,
**so that** I can install ThreeDoors without using the terminal or Homebrew.

**Acceptance Criteria:**
1. CI generates a signed .pkg installer containing the signed+notarized binary
2. The .pkg installs `threedoors` to `/usr/local/bin/`
3. The .pkg is also notarized with Apple
4. The installer is uploaded to GitHub Releases alongside the raw binaries
5. Double-clicking the .pkg on macOS launches the standard macOS installer UI

**Implementation Guidance:**
- Use `pkgbuild` and `productbuild` (available on macOS runners) to create the .pkg
- Sign the .pkg with Developer ID Installer certificate
- Notarize the .pkg separately from the binary
- DMG is an alternative but .pkg is simpler for CLI tools (no drag-to-install UX needed)

---

**Epic 6: Data Layer & Enrichment (Optional)**
*Stories to be defined only if clear need emerges from Epic 4*

---

## Epic 7: Plugin/Adapter SDK & Registry

**Epic Goal:** Formalize the adapter pattern (established in Epic 2) into a plugin SDK with runtime registry, config-driven provider selection, and developer documentation. This epic unblocks all future integrations by providing a stable, well-documented foundation.

**Scope:** Adapter registry, config.yaml-driven provider management, contract test suite, developer guide.

---

### Story 7.1: Adapter Registry & Runtime Discovery

**As a** developer building integrations,
**I want** a formal adapter registry that discovers and loads task providers at runtime,
**so that** new integrations can be added without modifying core application code.

**Acceptance Criteria:**
1. `AdapterRegistry` component created in `internal/adapters/registry.go`
2. Registry discovers all registered `TaskProvider` implementations at startup
3. Adapters register themselves via `registry.Register(name, factory)` pattern
4. Failed adapter initialization logs a warning and continues with other adapters
5. Registry exposes `ListProviders()`, `GetProvider(name)`, and `ActiveProviders()` methods
6. Existing text file and Apple Notes adapters migrated to use registry pattern

**Estimated Time:** 90-120 minutes

---

### Story 7.2: Config-Driven Provider Selection

**As a** user with multiple task sources,
**I want** to configure active backends via `~/.threedoors/config.yaml`,
**so that** I can choose which task providers are active without code changes.

**Acceptance Criteria:**
1. Config parser reads `~/.threedoors/config.yaml` for provider configuration
2. YAML schema supports `providers:` section with per-provider settings
3. Only configured providers are loaded and activated at startup
4. Provider-specific settings (paths, credentials, options) passed to adapter factory
5. Missing config.yaml falls back to text file provider (backward compatible)
6. Sample config.yaml generated on first run with commented-out provider examples

**Estimated Time:** 60-90 minutes

---

### Story 7.3: Adapter Developer Guide & Contract Tests

**As an** integration developer,
**I want** a clear guide and contract test suite for building adapters,
**so that** I can create new task provider integrations with confidence.

**Acceptance Criteria:**
1. Developer guide created at `docs/adapter-developer-guide.md`
2. Guide covers: `TaskProvider` interface spec, registration process, config schema, testing requirements
3. Example adapter implementation included (or reference to text file adapter)
4. Contract test suite in `internal/adapters/contract_test.go` validates any `TaskProvider`
5. Contract tests cover: CRUD operations, error handling, concurrent access, interface compliance
6. Contract test suite is reusable - adapters import and run it against their implementation

**Estimated Time:** 120-150 minutes

---

## Epic 8: Obsidian Integration (P0 - #2 Integration)

**Epic Goal:** Add Obsidian vault as the second task storage backend after Apple Notes. Obsidian's local-first Markdown approach makes it a natural fit for ThreeDoors' local-first philosophy.

**Scope:** Vault reader/writer, bidirectional sync, vault configuration, daily note integration.

---

### Story 8.1: Obsidian Vault Reader/Writer Adapter

**As a** user who manages tasks in Obsidian,
**I want** ThreeDoors to read and write tasks from my Obsidian vault,
**so that** I can use Three Doors with my existing Obsidian workflow.

**Acceptance Criteria:**
1. `ObsidianAdapter` implements `TaskProvider` interface
2. Reads Markdown files from configured vault folder
3. Parses task items using Obsidian checkbox syntax (`- [ ]`, `- [x]`, `- [/]`)
4. Supports Obsidian task metadata (due dates, tags, priorities in `📅`, `#tag`, `⏫` format)
5. Writes task status changes back to Markdown files using atomic file operations
6. Passes adapter contract test suite (Story 7.3)

**Estimated Time:** 120-150 minutes

---

### Story 8.2: Obsidian Bidirectional Sync

**As an** Obsidian user,
**I want** changes made in Obsidian reflected in ThreeDoors and vice versa,
**so that** my tasks stay in sync regardless of where I edit them.

**Acceptance Criteria:**
1. File watcher (fsnotify or polling) detects external changes to vault files
2. Changed files are re-parsed and task pool updated without full reload
3. ThreeDoors writes use atomic operations to prevent mid-write corruption
4. Concurrent edit handling: last-write-wins with conflict detection logging
5. Sync latency under 2 seconds for file change detection

**Estimated Time:** 90-120 minutes

---

### Story 8.3: Obsidian Vault Configuration

**As a** user,
**I want** to configure my Obsidian vault path, target folder, and file naming via config.yaml,
**so that** ThreeDoors integrates with my specific vault structure.

**Acceptance Criteria:**
1. Config.yaml supports `obsidian:` section with `vault_path`, `task_folder`, `file_pattern` settings
2. Vault path validated on startup (exists, readable, writable)
3. Invalid vault path produces clear error message and fallback to other providers
4. Default task folder is vault root; configurable to any subfolder
5. File pattern supports glob matching (e.g., `*.md`, `tasks/*.md`)

**Estimated Time:** 45-60 minutes

---

### Story 8.4: Obsidian Daily Note Integration

**As an** Obsidian user who uses daily notes,
**I want** ThreeDoors to read/write tasks from my daily note files,
**so that** tasks captured in daily notes appear in Three Doors.

**Acceptance Criteria:**
1. Config supports `daily_notes:` section with `enabled`, `folder`, `format` (date pattern)
2. Reads tasks from today's daily note file
3. Quick-add tasks can be appended to today's daily note under configurable heading
4. Supports common daily note formats: `YYYY-MM-DD.md`, `YYYY/MM/YYYY-MM-DD.md`
5. Missing daily note file handled gracefully (no error, just no tasks from that source)

**Estimated Time:** 60-90 minutes

---

## Epic 9: Testing Strategy & Quality Gates

**Epic Goal:** Establish comprehensive testing infrastructure ensuring reliability as the adapter ecosystem grows. Covers integration tests, contract tests, performance benchmarks, E2E tests, and CI coverage gates.

**Scope:** Apple Notes E2E, adapter contract tests, performance benchmarks, functional E2E, CI gates.

---

### Story 9.1: Apple Notes Integration E2E Tests

**As a** developer,
**I want** end-to-end tests for the Apple Notes integration workflow,
**so that** regressions in the sync pipeline are caught automatically.

**Acceptance Criteria:**
1. E2E test suite in `internal/adapters/applenotes/e2e_test.go`
2. Tests validate: note creation, task read, task update, bidirectional sync, error handling
3. Uses mock/stub AppleScript responses for CI compatibility (no real Apple Notes needed)
4. Tests cover: connectivity failure, partial sync, concurrent modification
5. Test fixtures in `testdata/applenotes/` for reproducible scenarios

**Estimated Time:** 120-150 minutes

---

### Story 9.2: Contract Tests for Adapter Compliance

**As an** adapter developer,
**I want** a reusable contract test suite validating any TaskProvider implementation,
**so that** all adapters behave consistently regardless of backend.

**Acceptance Criteria:**
1. Contract test suite in `internal/adapters/contract_test.go`
2. Tests: Create, Read, Update, Delete operations
3. Tests: Error handling (not found, permission denied, timeout)
4. Tests: Concurrent access safety
5. Tests: Interface compliance (all methods implemented correctly)
6. Each adapter runs the contract suite in its own test file

**Estimated Time:** 90-120 minutes

---

### Story 9.3: Performance Benchmarks

**As a** developer,
**I want** automated performance benchmarks validating the <100ms NFR,
**so that** performance regressions are caught before they reach users.

**Acceptance Criteria:**
1. Benchmark suite using Go's `testing.B` framework
2. Benchmarks for: adapter read, adapter write, adapter sync, door selection algorithm
3. Results compared against <100ms threshold (NFR13)
4. CI integration: benchmarks run on every PR
5. Benchmark results stored for trend analysis

**Estimated Time:** 60-90 minutes

---

### Story 9.4: Functional E2E Tests

**As a** developer,
**I want** functional end-to-end tests covering full user workflows,
**so that** the complete user experience is validated automatically.

**Acceptance Criteria:**
1. E2E tests exercise: launch → view doors → select door → manage task → exit
2. Tests verify session metrics are correctly generated
3. Tests cover: search, command palette, mood tracking workflows
4. Uses Bubbletea's `teatest` package for TUI testing
5. Tests run in CI without requiring a real terminal

**Estimated Time:** 120-150 minutes

---

### Story 9.5: CI Coverage Gates

**As a** team,
**I want** CI coverage gates preventing test coverage regression,
**so that** code quality is maintained as the codebase grows.

**Acceptance Criteria:**
1. Coverage measurement added to CI pipeline (`go test -coverprofile`)
2. Coverage threshold configured (starting at current coverage level)
3. PRs that reduce coverage below threshold are blocked
4. Coverage report generated and posted as PR comment
5. Threshold documented and adjustable in CI config

**Estimated Time:** 45-60 minutes

---

## Epic 10: First-Run Onboarding Experience

**Epic Goal:** Provide a guided welcome flow for new users, reducing time-to-value by explaining the Three Doors concept, setting up values/goals, and importing existing tasks.

**Scope:** Welcome flow, concept explanation, key bindings walkthrough, values/goals setup, task import.

---

### Story 10.1: Welcome Flow & Three Doors Explanation

**As a** new user,
**I want** a guided welcome experience on first launch,
**so that** I understand the Three Doors concept and feel confident using the tool.

**Acceptance Criteria:**
1. First-run detection based on absence of `~/.threedoors/` directory
2. Welcome screen with ThreeDoors branding and concept explanation
3. Interactive key bindings walkthrough (show keys, let user try them)
4. Skip option available at every step
5. Onboarding state persisted so it doesn't repeat on subsequent launches

**Estimated Time:** 90-120 minutes

---

### Story 10.2: Values/Goals Setup & Task Import

**As a** new user,
**I want** to set up my values/goals and import existing tasks during onboarding,
**so that** the tool is immediately useful with my real data.

**Acceptance Criteria:**
1. Values/goals input screen during onboarding (feeds into FR6 persistent display)
2. Import detection for common task sources (text files, Markdown files)
3. Import preview showing tasks to be imported
4. Imported tasks populate the task pool
5. Import step skippable; manual import available later via `:import` command

**Estimated Time:** 60-90 minutes

---

## Epic 11: Sync Observability & Offline-First

**Epic Goal:** Ensure robust offline-first operation with local change queue, sync status visibility in the TUI, conflict visualization, and sync debugging tools.

**Scope:** Offline queue, sync status indicator, conflict resolution UI, sync log.

---

### Story 11.1: Offline-First Local Change Queue

**As a** user working without connectivity,
**I want** all changes queued locally and replayed when sync targets are available,
**so that** I never lose work due to connectivity issues.

**Acceptance Criteria:**
1. Write-ahead log (WAL) in `~/.threedoors/sync-queue.jsonl` for pending changes
2. All adapter write operations go through the queue
3. Queue replay on connectivity restoration with ordered application
4. Failed replays retry with exponential backoff
5. Queue size limit with oldest-first eviction (configurable, default 10000 entries)
6. Core functionality (door selection, local task management) unaffected by sync state

**Estimated Time:** 120-150 minutes

---

### Story 11.2: Sync Status Indicator

**As a** user,
**I want** to see sync status per provider in the TUI,
**so that** I know whether my changes are synchronized.

**Acceptance Criteria:**
1. Status bar area in TUI shows per-provider sync state
2. States: ✓ synced, ↻ syncing, ⏳ pending (N items), ✗ error
3. Real-time updates as sync operations complete
4. Clicking/selecting the indicator shows last sync timestamp
5. Minimal screen real estate usage (icon + provider name)

**Estimated Time:** 60-90 minutes

---

### Story 11.3: Conflict Visualization & Sync Log

**As a** user encountering sync conflicts,
**I want** to see what conflicted and review a sync log,
**so that** I can resolve issues and trust the sync system.

**Acceptance Criteria:**
1. Conflict notification appears when detected during sync
2. Conflict detail view shows local vs remote versions side-by-side
3. Resolution options: keep local, keep remote, keep both
4. `:synclog` command shows chronological sync operations with timestamps
5. Sync log persisted to `~/.threedoors/sync.log` (rotated at 1MB)

**Estimated Time:** 90-120 minutes

---

## Epic 12: Calendar Awareness (Local-First, No OAuth)

**Epic Goal:** Add time-contextual door selection by reading local calendar sources. Strictly no OAuth, no cloud APIs - local data only.

**Scope:** macOS Calendar.app reader, .ics parser, CalDAV cache reader, time-aware door algorithm.

---

### Story 12.1: Local Calendar Source Reader

**As a** user,
**I want** ThreeDoors to read my local calendar,
**so that** it understands my available time for task-appropriate door selection.

**Acceptance Criteria:**
1. macOS Calendar.app events read via AppleScript (no OAuth)
2. .ics file parser for configured paths
3. CalDAV cache reader from local filesystem (`~/Library/Calendars/`)
4. Calendar events parsed into time blocks (start, end, title)
5. Config.yaml `calendar:` section for enabling sources and paths
6. Graceful fallback when calendar sources are unavailable

**Estimated Time:** 120-150 minutes

---

### Story 12.2: Time-Contextual Door Selection

**As a** user with calendar awareness enabled,
**I want** doors to suggest tasks fitting my available time,
**so that** I'm not shown a 2-hour task when I have a meeting in 15 minutes.

**Acceptance Criteria:**
1. Door selection algorithm considers next event time when choosing tasks
2. Short time blocks (< 30 min) prefer quick tasks (if effort metadata available)
3. Large open blocks include tasks of any duration
4. No calendar data = standard selection (graceful degradation)
5. Time context shown in TUI (e.g., "Next event in 45 min")

**Estimated Time:** 90-120 minutes

---

## Epic 13: Multi-Source Task Aggregation View

**Epic Goal:** Unified cross-provider task pool with duplicate detection and source attribution, enabling users to see all their tasks from all configured sources in one place.

**Scope:** Task pool aggregation, dedup detection, source attribution badges.

---

### Story 13.1: Cross-Provider Task Pool Aggregation

**As a** user with multiple task sources,
**I want** all tasks merged into a single pool for Three Doors selection,
**so that** I see tasks from all sources without switching between them.

**Acceptance Criteria:**
1. Task pool collects tasks from all active providers via registry
2. Unified pool used for door selection, search, and all task views
3. Provider load failures are isolated (one failing provider doesn't block others)
4. Refresh operation re-queries all active providers
5. Task pool maintains provider origin metadata for attribution

**Estimated Time:** 60-90 minutes

---

### Story 13.2: Duplicate Detection & Source Attribution

**As a** user with overlapping task sources,
**I want** duplicates flagged and each task's source clearly shown,
**so that** I don't work on the same task twice and know where each task lives.

**Acceptance Criteria:**
1. Fuzzy text matching identifies potential duplicates across providers
2. Duplicate pairs shown with visual indicator (e.g., "⚠ Possible duplicate")
3. User can merge or dismiss duplicate flags
4. Source provider shown as badge in door view, search results, and detail view
5. Badge format: provider icon/abbreviation (e.g., "📝" for text, "🍎" for Apple Notes, "💎" for Obsidian)

**Estimated Time:** 90-120 minutes

---

## Epic 14: LLM Task Decomposition & Agent Action Queue

**Epic Goal:** Enable LLM-powered task decomposition where selected tasks are broken into implementable stories/specs, output to git repos for coding agent (Claude Code, multiclaude) pickup.

**Scope:** Spike-first approach. Prompt engineering, output schema, git automation, agent handoff.

---

### Story 14.1: LLM Task Decomposition Spike

**As a** developer,
**I want** to spike on LLM-powered task decomposition,
**so that** we understand feasibility before committing to full implementation.

**Acceptance Criteria:**
1. Spike document in `docs/spikes/llm-decomposition.md`
2. Covers: prompt engineering experiments, output schema definition, git automation PoC
3. Tests multiple LLM providers (local: Ollama/llama.cpp; cloud: Claude API)
4. Agent handoff protocol drafted (how Claude Code / multiclaude discovers work)
5. Recommendation: build vs wait, local vs cloud, estimated effort for full implementation

**Estimated Time:** 3-4 hours (spike)

---

### Story 14.2: Agent Action Queue Integration

**As a** developer using ThreeDoors with coding agents,
**I want** decomposed tasks output to a git repo structure for agent pickup,
**so that** task decomposition flows into automated implementation.

**Acceptance Criteria:**
1. LLM output follows BMAD story file structure
2. Stories written to configurable repo path
3. Git operations: branch creation, commit, optional PR creation
4. ThreeDoors task updated with link to generated stories
5. Configurable LLM backend (local/cloud) via config.yaml

**Estimated Time:** 120-150 minutes

---

## Epic 15: Psychology Research & Validation

**Epic Goal:** Build evidence base for ThreeDoors design decisions through literature review. Findings feed into Epic 4's learning algorithm design.

**Scope:** Choice architecture review, mood-task correlation, procrastination research, motivational framework evidence.

---

### Story 15.1: Choice Architecture Literature Review

**As the** product team,
**I want** a literature review on the Three Doors choice architecture,
**so that** design decisions are grounded in behavioral science.

**Acceptance Criteria:**
1. Document at `docs/research/choice-architecture.md`
2. Covers: choice overload research (Iyengar & Lepper), paradox of choice, decision fatigue
3. Specific evidence for why 3 options (not 2, 4, or 5)
4. Comparable systems analysis (Tinder-like interfaces, binary choices, etc.)
5. Practical implications and recommendations for ThreeDoors

**Estimated Time:** 4-6 hours (research)

---

### Story 15.2: Mood-Task Correlation & Procrastination Research

**As the** product team,
**I want** research on mood-task correlation and procrastination interventions,
**so that** Epic 4's learning algorithm is evidence-informed.

**Acceptance Criteria:**
1. `docs/research/mood-correlation.md` covering mood-productivity models
2. `docs/research/procrastination.md` covering intervention mechanisms
3. Evidence assessment for "progress over perfection" as motivational framework
4. Actionable recommendations for Epic 4 adaptive algorithm design
5. Bibliography with accessible references

**Estimated Time:** 4-6 hours (research)

---
