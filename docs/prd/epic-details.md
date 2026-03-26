---
title: Epic Details
section: Epic Specifications
lastUpdated: '2026-03-06'
---

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
1. Go module initialized with `go mod init github.com/arcavenae/ThreeDoors`
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
1. `../../_bmad-output/planning-artifacts/config-schema.md` documents proposed schema, migration path
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
**I want** to install ThreeDoors via `brew install arcavenae/tap/threedoors`,
**so that** I get automatic updates and a standard macOS installation experience.

**Acceptance Criteria:**
1. A separate GitHub repository `arcavenae/homebrew-tap` is created with a Homebrew formula
2. Formula downloads the correct signed binary for the user's architecture (arm64 or amd64)
3. Formula includes SHA256 checksums for integrity verification
4. `brew install arcavenae/tap/threedoors` installs the binary to the Homebrew prefix
5. `brew upgrade arcavenae/tap/threedoors` upgrades to the latest version
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
1. Spike document in `../../_bmad-output/planning-artifacts/llm-decomposition.md`
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
1. Document at `../../_bmad-output/planning-artifacts/choice-architecture.md`
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
1. `../../_bmad-output/planning-artifacts/mood-correlation.md` covering mood-productivity models
2. `../../_bmad-output/planning-artifacts/procrastination.md` covering intervention mechanisms
3. Evidence assessment for "progress over perfection" as motivational framework
4. Actionable recommendations for Epic 4 adaptive algorithm design
5. Bibliography with accessible references

**Estimated Time:** 4-6 hours (research)

---

## Epic 16: iPhone Mobile App (SwiftUI) 🆕

**Epic Goal:** Bring the Three Doors experience to iPhone with a native SwiftUI app that syncs tasks via the same Apple Notes document used by the desktop TUI. The mobile app provides the core Three Doors experience — see three doors, pick one, take action — optimized for touch interaction.

**Origin:** Party mode mobile app discussion (2026-03-02)
**Research:** See `../../_bmad-output/planning-artifacts/mobile-app-research.md` for full analysis of technology choices.

**Prerequisites:** Epic 2 ✅ (Apple Notes integration established)
**Tech Stack:** Swift 5.9+, SwiftUI, iCloud Drive, Xcode 16+, iOS 17+ target

**Key Design Decisions:**
- **Native SwiftUI** over React Native/Flutter/PWA — ThreeDoors is Apple-ecosystem only, and SwiftUI provides seamless iCloud/Apple Notes integration
- **Protocol-level code sharing** — Port Go interfaces (TaskProvider, Task model, SyncEngine patterns) to Swift protocols rather than using gomobile
- **Apple Notes as shared backend** — Both TUI and mobile read/write the same Apple Notes document; iCloud syncs automatically
- **Swipeable card carousel** — Three Doors translates to swipeable cards with tap-to-open, pull-to-refresh, and swipe-to-complete gestures

---

### Story 16.1: SwiftUI Project Setup & CI

**As a** developer,
**I want** a working SwiftUI project with CI pipeline,
**so that** I have a foundation for building the Three Doors mobile app.

**Acceptance Criteria:**
1. Xcode project created at `mobile/ThreeDoors/` with SwiftUI lifecycle
2. Target: iOS 17+, iPhone only (iPad layout deferred)
3. Basic app shell renders "ThreeDoors" header with app icon placeholder
4. GitHub Actions CI workflow for building and testing the Swift project
5. `.gitignore` configured for Xcode project artifacts
6. SwiftUI previews configured for development
7. App compiles and runs in iOS Simulator without errors

**Estimated Time:** 60-90 minutes

---

### Story 16.2: Task Provider Protocol & Apple Notes Reader

**As a** mobile user,
**I want** the app to read tasks from the same Apple Notes document used by the desktop TUI,
**so that** my tasks are consistent across devices.

**Acceptance Criteria:**
1. `TaskProvider` Swift protocol defined mirroring Go's `TaskProvider` interface (loadTasks, saveTask, deleteTask, markComplete)
2. `Task` Swift struct defined with Codable conformance (id, text, status, notes, createdAt, updatedAt)
3. `TaskStatus` enum matches Go version (todo, blocked, inProgress, inReview, complete)
4. `AppleNotesProvider` implementation reads tasks from Apple Notes
5. Checkbox format parsing matches TUI: `- [ ] task` (todo), `- [x] task` (complete)
6. Deterministic UUID generation matches Go implementation (`noteTitle:lineIndex` based)
7. Note title configurable (matches TUI's config)
8. Error handling for Notes access permission denied
9. Unit tests with mock note content

**Estimated Time:** 120-150 minutes

---

### Story 16.3: Three Doors Card Carousel

**As a** mobile user,
**I want** to see three task cards I can swipe through,
**so that** I get the Three Doors experience on my phone.

**Acceptance Criteria:**
1. Three task cards displayed as a horizontal swipeable carousel (TabView with PageTabViewStyle or custom)
2. Each card shows task text, status badge, and creation date
3. Cards use distinct visual styling consistent with TUI door aesthetic
4. Current card indicator (dots or similar) shows position
5. Smooth swipe animation between cards
6. Empty state handled ("No tasks found — add tasks in Apple Notes")
7. Loading state while fetching from Apple Notes
8. Card layout adapts to different iPhone screen sizes

**Estimated Time:** 90-120 minutes

---

### Story 16.4: Door Detail & Status Actions

**As a** mobile user,
**I want** to tap a card to see task details and change its status,
**so that** I can take action on tasks from my phone.

**Acceptance Criteria:**
1. Tapping a card opens a detail view with full task text, notes, status, timestamps
2. Detail view includes action buttons: Complete, Blocked, In Progress, In Review
3. Status change writes back to Apple Notes (same checkbox format)
4. Success haptic feedback on status change
5. "Progress over perfection" toast shown after completing a task
6. Detail view dismissible via swipe-down gesture or close button
7. After status change, returns to carousel with updated card
8. Optimistic UI update with rollback on write failure

**Estimated Time:** 90-120 minutes

---

### Story 16.5: Session Metrics & iCloud Sync

**As a** developer analyzing usage patterns,
**I want** mobile session metrics compatible with the desktop JSONL format,
**so that** mobile and desktop sessions can be analyzed together.

**Acceptance Criteria:**
1. `SessionTracker` Swift class mirrors Go's SessionTracker (session_id, start/end, behavioral counters)
2. Metrics recorded: doors_viewed, tasks_completed, refreshes, status_changes, card_swipes
3. Session data appended to `sessions.jsonl` in app's iCloud Drive container
4. JSONL format matches Go's MetricsWriter output schema exactly
5. iCloud Drive sync configured for `~/.threedoors/` equivalent directory
6. Config file (`config.yaml`) readable from shared iCloud Drive location
7. Metrics written on app background/termination (UIApplication lifecycle)
8. Offline metrics cached locally, synced when iCloud available

**Estimated Time:** 120-150 minutes

---

### Story 16.6: Swipe Gestures & Pull-to-Refresh

**As a** mobile user,
**I want** intuitive gestures for common actions,
**so that** the app feels native and fast to use.

**Acceptance Criteria:**
1. **Pull-to-refresh**: Pull down on carousel to generate new set of three doors
2. **Swipe right on card**: Quick-complete gesture (with confirmation haptic)
3. **Swipe left on card**: Defer/skip gesture (marks as "not now")
4. **Long press on card**: Opens context menu with all status options
5. Gesture animations smooth and responsive (60 FPS)
6. Gesture hints shown on first use (onboarding overlay)
7. Undo option shown briefly after swipe-to-complete (5 second window)
8. Pull-to-refresh triggers Apple Notes re-read

**Estimated Time:** 90-120 minutes

---

### Story 16.7: Polish & TestFlight Distribution

**As a** developer,
**I want** the app polished and distributed via TestFlight,
**so that** it can be tested on real devices before wider release.

**Acceptance Criteria:**
1. App icon designed and configured (Three Doors motif)
2. Launch screen configured with branding
3. Dark mode support (matches system setting)
4. Accessibility: VoiceOver labels on all interactive elements
5. Accessibility: Dynamic Type support for text sizing
6. App configured in App Store Connect for TestFlight
7. Privacy labels configured: "Data Not Collected" (tasks stay in Apple Notes)
8. TestFlight build uploaded and available for testing
9. Minimum iOS version validated (iOS 17+)
10. No crashes on iPhone SE (smallest screen) through iPhone 16 Pro Max (largest screen)

**Estimated Time:** 120-150 minutes

---

## Epic 17: Door Theme System ✅ COMPLETE

**Epic Goal:** Replace the uniform rounded-border door appearance with visually distinct themed doors using ASCII/ANSI art frames, with user-selectable themes via onboarding, settings view, and config.yaml.

**Status:** COMPLETE — All 6 stories implemented and merged (PRs #119, #120, #121, #123, #124, #122).

**Scope:** Theme type definitions, registry, three new theme implementations (Modern, Sci-Fi, Shoji), Classic theme wrapper, DoorsView integration, onboarding theme picker, settings `:theme` command, config persistence, and golden file tests.

**Research:** See `../../_bmad-output/planning-artifacts/door-themes-research.md` (8 ANSI mockups, feasibility matrix), `../../_bmad-output/planning-artifacts/door-themes-analyst-review.md` (analyst assessment), `../../_bmad-output/planning-artifacts/door-themes-party-mode.md` (architecture decisions).

---

### Story 17.1: Theme Types, Registry, and Classic Theme Wrapper

**As a** developer,
**I want** a DoorTheme type, ThemeColors struct, theme registry, and a Classic theme that wraps the current rendering,
**so that** the theme infrastructure is in place and existing behavior is preserved as a theme option.

**Acceptance Criteria:**
1. `DoorTheme` struct defined in `internal/tui/themes/theme.go` with fields: Name, Description, Render function, Colors (ThemeColors), MinWidth
2. `ThemeColors` struct defined with Frame, Fill, Accent, Selected fields (all `lipgloss.Color`)
3. Theme registry in `internal/tui/themes/registry.go` as `map[string]DoorTheme` with lookup helper
4. `DefaultThemeName` constant set to `"modern"`
5. Classic theme in `internal/tui/themes/classic.go` that wraps current Lipgloss `doorStyle`/`selectedDoorStyle` rendering
6. Classic theme produces output identical to current door rendering (verified by test)
7. All code passes `make fmt` and `make lint`

**Estimated Time:** 60-90 minutes

---

### Story 17.2: Modern, Sci-Fi, and Shoji Theme Implementations

**As a** user,
**I want** three visually distinct door themes to choose from,
**so that** my Three Doors interface has personality and visual variety.

**Acceptance Criteria:**
1. Modern/Minimalist theme (`modern.go`): clean single-line frame, minimal `●` doorknob, generous whitespace
2. Sci-Fi/Spaceship theme (`scifi.go`): double-line outer frame (`╔╗╚╝═║`), side rails with `░▓` shade, upper content panel + lower control panel
3. Japanese Shoji theme (`shoji.go`): grid pattern using `┼─│┬┴├┤` characters, task text overlaid on central cells
4. All themes render correctly at widths from their declared MinWidth up to 60+ characters
5. All themes handle the `selected` flag by adjusting frame colors
6. All themes word-wrap content text to fit within their interior content area
7. Only Unicode characters from box-drawing, block elements, and geometric shapes ranges used (NFR17)
8. All code passes `make fmt` and `make lint`

**Estimated Time:** 120-180 minutes

**Dependencies:** Story 17.1

---

### Story 17.3: DoorsView Integration — Load Theme from Config

**As a** user,
**I want** my selected door theme applied to the Three Doors display,
**so that** the doors render with my chosen visual style.

**Acceptance Criteria:**
1. `DoorsView` struct gains a `theme themes.DoorTheme` field
2. Theme loaded from `config.yaml` `theme` key at DoorsView initialization
3. `View()` method uses `theme.Render()` instead of `doorStyle.Render()`
4. Invalid or missing theme config falls back to `DefaultThemeName` ("modern") with warning logged
5. Terminal width checked against `theme.MinWidth`; falls back to Classic theme when too narrow
6. Existing per-door color system replaced by theme's ThemeColors when a non-classic theme is active
7. Door number labels overlaid consistently regardless of active theme
8. All existing TUI tests continue to pass
9. All code passes `make fmt` and `make lint`

**Estimated Time:** 60-90 minutes

**Dependencies:** Stories 17.1, 17.2

---

### Story 17.4: Theme Picker in Onboarding Flow

**As a** new user,
**I want** to browse and select a door theme during first-run onboarding,
**so that** I can personalize my Three Doors experience from the start.

**Acceptance Criteria:**
1. Theme picker step added to the first-run onboarding flow (after values/goals, before key bindings walkthrough)
2. Picker displays doors rendered with each available theme in a horizontal preview
3. Left/right arrow keys browse between themes; current theme name and description shown
4. Enter confirms selection; Escape or "Skip" defaults to Modern/Minimalist
5. Selected theme written to `config.yaml`
6. Picker handles narrow terminals gracefully (vertical layout or one-at-a-time display)
7. All code passes `make fmt` and `make lint`

**Estimated Time:** 90-120 minutes

**Dependencies:** Story 17.3, Epic 10 (onboarding infrastructure — can stub if not yet implemented)

---

### Story 17.5: Settings View — `:theme` Command with Preview

**As a** user,
**I want** to change my door theme from within the TUI at any time,
**so that** I can try different themes without editing config files.

**Acceptance Criteria:**
1. `:theme` command registered in command palette
2. Command opens theme picker view (reuses component from Story 17.4)
3. Current theme highlighted in the picker
4. Theme change takes effect immediately (no restart required)
5. New theme selection persisted to `config.yaml`
6. `:theme` command listed in `:help` output
7. All code passes `make fmt` and `make lint`

**Estimated Time:** 60-90 minutes

**Dependencies:** Stories 17.3, 17.4

---

### Story 17.6: Golden File Tests for All Themes

**As a** developer,
**I want** golden file tests for every door theme,
**so that** visual regressions are caught automatically.

**Acceptance Criteria:**
1. Golden file test for each theme (Classic, Modern, Sci-Fi, Shoji) at 28-char and 40-char widths
2. Both selected and unselected states tested per theme
3. Golden files stored in `internal/tui/themes/testdata/`
4. Tests use `go test -update` flag to regenerate golden files
5. Width boundary tests: each theme at MinWidth (should render) and MinWidth-1 (should indicate fallback)
6. Content wrapping tests: short (1 line), medium (2-3 lines), and long (5+ lines) task text
7. All tests pass with `make test`
8. All code passes `make fmt` and `make lint`

**Estimated Time:** 60-90 minutes

**Dependencies:** Stories 17.1, 17.2

---

## Epic 18: Docker E2E & Headless TUI Testing Infrastructure ✅ COMPLETE

**Epic Goal:** Establish reproducible, automated E2E testing using Docker containers and Bubbletea's `teatest` package for headless TUI interaction testing with golden file snapshots.

**Scope:** Headless TUI test harness via `teatest`, golden file snapshot tests for visual regression, workflow replay tests, Docker-based reproducible test environment, and CI integration.

**Status:** COMPLETE — All 5 stories implemented and merged

---

### Story 18.1: Headless TUI Test Harness with teatest

**As a** developer,
**I want** a headless TUI test harness using `teatest`,
**so that** I can write automated tests that interact with the full TUI without a real terminal.

**Acceptance Criteria:**
1. `teatest` package added to `go.mod`
2. Tests can create a `teatest.NewTestModel` with the root TUI model
3. Tests can send key messages to simulate user input
4. Test helper `NewTestApp` provided with options for term size, task file, provider
5. At least 3 smoke tests: app launch, door display, quit
6. ASCII color profile for deterministic output

---

### Story 18.2: Golden File Snapshot Tests

**Acceptance Criteria:**
1. Golden file infrastructure for comparing rendered TUI output against stored references
2. `--update` flag for regenerating golden files
3. Tests at multiple terminal widths

### Story 18.3: Workflow Replay Tests

**Acceptance Criteria:**
1. Multi-step interaction tests (navigate, select, complete)
2. Full user workflow coverage

### Story 18.4: Docker Environment

**Acceptance Criteria:**
1. `Dockerfile.test` + `docker-compose.test.yml` for reproducible test environment
2. Consistent test results across machines

### Story 18.5: CI Integration

**Acceptance Criteria:**
1. Docker E2E tests integrated into CI pipeline
2. Golden file drift detection

---

## Epic 19: Jira Integration

**Epic Goal:** Integrate Jira as a task source, enabling developers to see their Jira issues as ThreeDoors tasks. Phase 1 is read-only; Phase 2 adds bidirectional sync via the transitions API.

**Prerequisites:** Epic 7 (adapter SDK), Epic 11 (sync observability), Epic 13 (multi-source aggregation)

---

### Story 19.1: Jira HTTP Client

**As a** developer,
**I want** a thin HTTP client for the Jira REST API v3,
**so that** the JiraProvider can query and transition issues without a third-party SDK dependency.

**Acceptance Criteria:**
1. `Client` struct in `internal/adapters/jira/jira_client.go` with `NewClient(config AuthConfig) *Client`
2. Basic Auth (Cloud) and PAT Bearer (Server/DC) authentication support
3. `SearchJQL(ctx, jql, fields, maxResults, pageToken) (*SearchResult, error)` method using `POST /rest/api/3/search/jql`
4. Cursor-based pagination support via `nextPageToken`
5. HTTP 429 handling: parse `Retry-After` header, return `*RateLimitError`
6. Unit tests using `httptest.NewServer` with canned responses
7. No third-party dependencies beyond stdlib

**Estimated Time:** 90-120 minutes

**Dependencies:** None — foundation for Epic 19

---

### Story 19.2: Jira Read-Only Provider

**As a** ThreeDoors user with Jira,
**I want** my Jira issues to appear as tasks in ThreeDoors,
**so that** I can use the Three Doors selection for my sprint work.

**Acceptance Criteria:**
1. `JiraProvider` in `internal/adapters/jira/jira_provider.go` implementing `core.TaskProvider`
2. `LoadTasks()` executes configured JQL, paginates results, maps to `[]*core.Task`
3. Field mapping: issue key → ID, summary → Text, statusCategory → Status, priority → Effort, project+labels → Context
4. `SaveTask/SaveTasks/DeleteTask/MarkComplete` return `core.ErrReadOnly`
5. `Watch()` returns `nil`; `HealthCheck()` tests API connectivity
6. Adapter factory registered in `Registry` as `"jira"`
7. Contract tests pass via `adapters.RunContractTests`
8. Table-driven field mapping tests for all status/priority combinations

**Estimated Time:** 90-120 minutes

**Dependencies:** Story 19.1

---

### Story 19.3: Jira Bidirectional Sync

**As a** ThreeDoors user,
**I want** completing a Jira task in ThreeDoors to transition the issue to Done in Jira,
**so that** my task status stays synchronized.

**Acceptance Criteria:**
1. `MarkComplete(taskID)` implementation: GET transitions → find Done transition → POST transition
2. Handle 409 Conflict (concurrent transition) with retry
3. WAL wrapping: `core.NewWALProvider(jiraProvider)` for offline queuing
4. FallbackProvider wrapping for graceful degradation when Jira is unreachable
5. Local cache file (`~/.threedoors/jira-cache.yaml`) updated on successful LoadTasks, used as fallback
6. Unit tests for transition discovery, execution, conflict handling, and WAL replay

**Estimated Time:** 90-120 minutes

**Dependencies:** Stories 19.1, 19.2

---

### Story 19.4: Jira Config and Registration

**As a** ThreeDoors user,
**I want** to configure Jira integration in my config.yaml,
**so that** I can connect to my Jira instance with my preferred JQL filter.

**Acceptance Criteria:**
1. Config section for Jira in `~/.threedoors/config.yaml`: url, auth_type, jql, max_results, poll_interval
2. Environment variable fallback: `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`
3. Config validation: required fields (url, auth_type), URL format, auth_type enum
4. Adapter factory wired to config parsing
5. Registration in `RegisterBuiltinAdapters()`
6. Unit tests for config parsing, validation, env var fallback

**Estimated Time:** 60-90 minutes

**Dependencies:** Story 19.2

---

## Epic 20: Apple Reminders Integration

**Epic Goal:** Add Apple Reminders as a task source with full CRUD support, leveraging its structured data model for a higher-quality integration than Apple Notes.

**Prerequisites:** Epic 7 (adapter SDK), macOS only

---

### Story 20.1: Reminders JXA Scripts and CommandExecutor

**As a** developer,
**I want** JXA scripts for reading, creating, updating, completing, and deleting reminders,
**so that** the RemindersProvider has a reliable access layer for Apple Reminders.

**Acceptance Criteria:**
1. `CommandExecutor` interface in `internal/adapters/reminders/` (reuse pattern from applenotes)
2. `OSAScriptExecutor` implementation using `osascript -l JavaScript`
3. JXA script: read incomplete reminders from specified lists as JSON array
4. JXA script: read all reminder list names as JSON array
5. JXA script: complete a reminder by ID
6. JXA script: create a new reminder with name, body, priority, list
7. JXA script: delete a reminder by ID
8. Unit tests with mock executor returning canned JSON responses

**Estimated Time:** 90-120 minutes

**Dependencies:** None — foundation for Epic 20

---

### Story 20.2: Reminders Read-Only Provider

**As a** ThreeDoors user on macOS,
**I want** my Apple Reminders to appear as tasks in ThreeDoors,
**so that** I can use Three Doors selection for my reminder lists.

**Acceptance Criteria:**
1. `RemindersProvider` in `internal/adapters/reminders/reminders_provider.go` implementing `core.TaskProvider`
2. `LoadTasks()` reads incomplete reminders via JXA, maps to `[]*core.Task`
3. Field mapping: id → ID (stable persistent URI), name → Text, body → Notes, priority → Effort, completed → Status
4. Configurable list filtering (empty = all lists)
5. `SaveTask/SaveTasks/DeleteTask/MarkComplete` return `core.ErrReadOnly`
6. `HealthCheck()` attempts lightweight read, reports TCC permission status
7. Contract tests pass via `adapters.RunContractTests`
8. Table-driven field mapping tests

**Estimated Time:** 90-120 minutes

**Dependencies:** Story 20.1

---

### Story 20.3: Reminders Write Support

**As a** ThreeDoors user,
**I want** to complete, create, and delete reminders from within ThreeDoors,
**so that** changes sync back to Apple Reminders on all my devices via iCloud.

**Acceptance Criteria:**
1. `MarkComplete(taskID)` sets `completed = true` via JXA
2. `SaveTask(task)` creates new reminder if ID is empty, updates existing if ID matches
3. `DeleteTask(taskID)` removes reminder via JXA
4. Error categorization: permission denied, reminder not found, timeout
5. Retry logic with configurable attempts and backoff (reuse pattern from applenotes)
6. Full contract test compliance (stable IDs enable complete round-trip testing)

**Estimated Time:** 90-120 minutes

**Dependencies:** Stories 20.1, 20.2

---

### Story 20.4: Reminders Config, Registration, and Health Check

**As a** ThreeDoors user,
**I want** to configure Apple Reminders in config.yaml with list filtering,
**so that** I only see reminders from my work-related lists.

**Acceptance Criteria:**
1. Config section: lists (comma-separated), include_completed (bool)
2. Adapter factory wired to config parsing
3. Registration in `RegisterBuiltinAdapters()` as `"reminders"`
4. `HealthCheck()` returns clear guidance when Reminders access is denied
5. Platform guard: `//go:build darwin` — adapter only available on macOS
6. Unit tests for config parsing and validation

**Estimated Time:** 60-90 minutes

**Dependencies:** Story 20.2

---

## Epic 21: Sync Protocol Hardening

**Epic Goal:** Harden the sync architecture for reliable multi-provider operation with background scheduling, fault isolation, and cross-provider identity mapping.

**Prerequisites:** Epic 11 (sync observability), Epic 13 (multi-source aggregation)

---

### Story 21.1: Sync Scheduler with Per-Provider Loops

**As a** ThreeDoors user with multiple task sources,
**I want** background sync to run automatically per provider,
**so that** I don't have to interact with the app to discover remote changes.

**Acceptance Criteria:**
1. `SyncScheduler` struct in `internal/core/sync_scheduler.go` managing per-provider `ProviderLoop` goroutines
2. Each loop runs independently with configurable poll interval
3. Hybrid trigger: `Watch()` channel as primary, polling as fallback (concurrent)
4. `AdaptiveInterval`: reset to min on success, multiply on failure (up to max), ±20% jitter
5. Results fan-in to a single channel consumed by the TUI via `tea.Cmd`
6. `Start(ctx)` and `Stop()` lifecycle methods
7. Unit tests with fake clock, mock providers, deterministic scheduling
8. No goroutine leaks on Stop() (verified by test)

**Estimated Time:** 120-150 minutes

**Dependencies:** None within this epic — but requires Epic 11 SyncEngine

---

### Story 21.2: Circuit Breaker per Provider

**As a** ThreeDoors user,
**I want** a failing provider to be isolated without affecting other providers,
**so that** one unreachable service doesn't degrade my entire task view.

**Acceptance Criteria:**
1. `CircuitBreaker` struct in `internal/core/circuit_breaker.go` with Closed/Open/Half-Open states
2. Closed → Open after 5 consecutive failures within 2-minute window
3. Open → Half-Open after probe interval (starts 30s, doubles, max 30m)
4. Half-Open → Closed on successful probe; → Open on failed probe
5. Integration with `SyncStatusTracker`: expose circuit state per provider
6. `MultiSourceAggregator` uses circuit state to return cached tasks for Open providers
7. Table-driven state transition tests
8. Thread-safe (sync.Mutex protected)

**Estimated Time:** 90-120 minutes

**Dependencies:** None within this epic

---

### Story 21.3: Canonical ID Mapping (SourceRef)

**As a** ThreeDoors user with tasks in multiple providers,
**I want** tasks to be permanently linked across providers,
**so that** deduplication is reliable and write-back works to all sources.

**Acceptance Criteria:**
1. `SourceRef` struct added to `internal/core/task.go`: Provider (string), NativeID (string)
2. `Task` gains `SourceRefs []SourceRef` field with YAML/JSON serialization
3. Backward compatibility: if `SourceRefs` is empty, fall back to `SourceProvider`
4. Identity resolution: lookup by SourceRef → match, or heuristic dedup → link
5. Write routing uses `SourceRefs` to update all providers that know a task
6. Schema version bump to 2 with migration function for existing tasks
7. Unit tests for SourceRef matching, migration, backward compatibility

**Estimated Time:** 90-120 minutes

**Dependencies:** None within this epic

---

### Story 21.4: Sync Dashboard Enhancements

**As a** ThreeDoors user,
**I want** to see per-provider sync health, staleness, and pending queue status,
**so that** I know which data is current and which may be stale.

**Acceptance Criteria:**
1. `ProviderSyncStatus` gains: CircuitState, RetryIn, StaleSince, SyncCount24h, ErrorCount24h
2. TUI sync status line shows circuit state icons: ✓ (closed), ✗ (open), ↻ (half-open)
3. Staleness indicator: tasks from providers exceeding staleness threshold are visually annotated
4. WAL pending count displayed: ⏳ WAL pending (N items, oldest Xm)
5. Unit tests for status rendering at all circuit states
6. No new TUI views — extends existing `SyncStatusTracker` display

**Estimated Time:** 60-90 minutes

**Dependencies:** Stories 21.1, 21.2

---

## Epic 22: Self-Driving Development Pipeline

**Epic Goal:** Enable ThreeDoors tasks to directly trigger multiclaude worker agents, creating a closed loop where the app dispatches its own development work and tracks results (PRs, CI status) back in the TUI.

**Scope:** DevDispatch data model, file-based queue persistence, dispatch engine wrapping multiclaude CLI, TUI dispatch key binding and command, dev queue view, worker status polling, auto-generated follow-up tasks, and safety guardrails.

**Status:** COMPLETE — All 8 stories implemented and merged (PRs #149, #152, #163, #162, #161, #164, #159, #160)

---

### Story 22.1: DevDispatch Data Model & Queue Persistence

**As a** ThreeDoors user,
**I want** a persistent dev dispatch queue,
**so that** dispatched development tasks survive TUI restarts.

**Acceptance Criteria:**
1. `DevDispatch` model with fields: ID, TaskID, Description, Status (pending/dispatched/completed/failed), WorkerName, PRNumber, CIStatus, timestamps
2. File-based queue persistence at `~/.threedoors/dev-queue.yaml`
3. Queue CRUD operations: add, list, update status, remove
4. Atomic writes for queue file updates

---

### Story 22.2: Dispatch Engine (multiclaude CLI Wrapper)

**As a** ThreeDoors user,
**I want** to dispatch development tasks to multiclaude workers,
**so that** coding work can be automated from within the TUI.

**Acceptance Criteria:**
1. `DispatchEngine` wraps multiclaude CLI: `CreateWorker`, `ListWorkers`, `GetHistory`, `RemoveWorker`
2. Worker creation generates task description from selected task context
3. Error handling for multiclaude CLI failures (not installed, not configured, etc.)
4. Thread-safe dispatch operations

---

### Story 22.3: TUI Dispatch Key Binding & Command

**As a** ThreeDoors user,
**I want** to dispatch a task for development from the detail view,
**so that** I can trigger development work without leaving the TUI.

**Acceptance Criteria:**
1. `x` key in detail view triggers dispatch workflow
2. `:dispatch` command in command palette
3. Confirmation dialog before dispatching
4. Visual feedback on successful dispatch

---

### Story 22.4: Dev Queue View

**As a** ThreeDoors user,
**I want** to see and manage my dev dispatch queue,
**so that** I can track which tasks are being worked on.

**Acceptance Criteria:**
1. Dev queue view showing all queue items with status indicators
2. Approve, kill, and remove actions on queue items
3. Status icons for pending/dispatched/completed/failed

---

### Story 22.5: Worker Status Polling

**As a** ThreeDoors user,
**I want** automatic status updates on dispatched workers,
**so that** I know when work is complete.

**Acceptance Criteria:**
1. `tea.Tick` polling at 30-second intervals for active workers
2. Status updates reflected in queue view and status bar
3. PR number captured when worker creates a PR

---

### Story 22.6: Auto-Generated Follow-Up Tasks

**As a** ThreeDoors user,
**I want** follow-up tasks automatically created when workers complete,
**so that** I don't forget to review PRs and fix CI.

**Acceptance Criteria:**
1. On worker completion, generate "Review PR #NNN" task
2. On CI failure, generate "Fix CI for PR #NNN" task
3. Follow-up tasks added to main task pool

---

### Story 22.7: Optional Story File Generation

**As a** ThreeDoors user,
**I want** optional story file generation for dispatched tasks,
**so that** complex tasks get proper story-driven development.

**Acceptance Criteria:**
1. Optional story generation via existing `AgentService` (Epic 14)
2. Toggle in dispatch dialog
3. Generated story file path included in worker task description

---

### Story 22.8: Safety Guardrails

**As a** ThreeDoors user,
**I want** safety limits on development dispatch,
**so that** I don't accidentally overwhelm the system.

**Acceptance Criteria:**
1. Max concurrent workers limit (configurable, default 3)
2. Approval gate for dispatches (configurable)
3. Rate limiting (min interval between dispatches)
4. Audit log of all dispatch operations

---

## Epic 23: CLI Interface

**Epic Goal:** Provide a complete non-TUI CLI interface for ThreeDoors that serves both human power users (scriptable task management) and LLM agents (structured JSON output).

**Scope:** Cobra-based CLI scaffold with `--json` persistent flag, task CRUD commands, door selection commands, session and analytics commands, shell completions.

**Status:** COMPLETE — All 11 stories implemented and merged (PRs #170-#195, #225)

---

### Story 23.1: CLI Scaffold & Output Formatter

**As a** developer or LLM agent,
**I want** a Cobra-based CLI with JSON output support,
**so that** I can script ThreeDoors operations.

**Acceptance Criteria:**
1. Cobra root command `threedoors` with `--json` persistent flag
2. Output formatter interface supporting human-readable and JSON modes
3. Version, help, and completion subcommands
4. Shell completions for bash, zsh, fish, PowerShell

---

### Stories 23.2-23.11: Task CRUD, Door Commands, Session Commands

Stories 23.2-23.11 implement the full CLI command surface: `task list`, `task show`, `task add`, `task update`, `task complete`, `doors`, `doors pick`, `doors roll`, session commands, and analytics commands. Each follows the same pattern: Cobra command definition, domain layer integration, human and JSON output formatting, and test coverage.

---

## Epic 24: MCP/LLM Integration Server

**Epic Goal:** Expose ThreeDoors functionality as an MCP (Model Context Protocol) server enabling LLM agents to interact with tasks programmatically.

**Scope:** MCP server with tool definitions for task management, structured JSON responses, resource endpoints for task context, integration with existing TaskProvider infrastructure.

**Status:** COMPLETE — All 8 stories implemented and merged (PRs #177-#197)

---

### Stories 24.1-24.8: MCP Server Implementation

Stories implement the full MCP server: tool registration and discovery, task management tools (list, show, add, update, complete), door interaction tools (roll, pick), resource endpoints for task context and session data, and integration testing. The server runs as a subprocess invoked by LLM agents via the MCP protocol.

---

## Epic 25: Todoist Integration

**Epic Goal:** Integrate Todoist as a task source via REST API v1 with thin HTTP client, read-only then bidirectional sync.

**Scope:** Todoist REST API v1 client with API token auth, read-only provider with field mapping (priority-to-effort, project filtering), bidirectional sync with WAL integration, contract tests.

**Status:** COMPLETE — All 4 stories implemented and merged (PRs #308, #321, plus Stories 25.3 & 25.4)

---

### Story 25.1: Todoist HTTP Client & Auth

**Acceptance Criteria:**
1. Thin HTTP client for Todoist REST API v1
2. API token authentication via config.yaml/env var
3. Task listing with project filtering
4. Error handling with rate limit respect (NFR20)

### Story 25.2: Read-Only Provider

**Acceptance Criteria:**
1. `TodoistProvider` implementing `TaskProvider` interface
2. Priority-to-effort mapping (P1→High, P2→Medium, etc.)
3. Project filtering configuration
4. Local cache with TTL (NFR21)

### Story 25.3: Bidirectional Sync

**Acceptance Criteria:**
1. Complete tasks via Todoist API
2. WAL offline queuing for write operations
3. Circuit breaker integration (NFR23)

### Story 25.4: Contract Tests

**Acceptance Criteria:**
1. Contract tests with mocked HTTP responses
2. Integration tests verifying field mapping
3. Edge case coverage (empty projects, API errors)

---

## Epic 26: GitHub Issues Integration

**Epic Goal:** Integrate GitHub Issues as a task source, enabling developers to see their GitHub issues as ThreeDoors tasks.

**Scope:** GitHub API client via go-github SDK, read-only provider with issue-to-task mapping, bidirectional sync (close issues), contract tests.

**Status:** COMPLETE — All 4 stories implemented and merged (PRs #201-#205)

---

### Stories 26.1-26.4: GitHub Issues Provider

Follows the standard 4-story integration pattern: SDK client, read-only provider with field mapping (labels→tags, milestones→context, assignee filtering), bidirectional sync (close issues on task completion), and contract tests with mocked GitHub API responses.

---

## Epic 27: Daily Planning Mode

**Epic Goal:** Add a guided daily planning ritual that transforms ThreeDoors from a reactive task picker into a proactive morning engagement tool, driving long-term retention through structured planning sessions.

**Scope:** Planning data model with session-scoped `+focus` tag, review incomplete tasks flow, focus selection flow, energy level matching, planning session metrics, focus-aware door selection scoring, CLI and TUI commands.

**Status:** COMPLETE — All 5 stories implemented and merged

**Requirements:** FR97-FR103 (already in requirements.md)

---

### Stories 27.1-27.5: Daily Planning Implementation

Five stories implementing the full planning flow: 27.1 (planning data model and review flow), 27.2 (focus selection with energy matching), 27.3 (planning metrics logging), 27.4 (focus-aware door scoring), 27.5 (CLI `threedoors plan` and TUI `:plan` commands).

---

## Epic 28: Snooze/Defer as First-Class Action

**Epic Goal:** Surface existing `StatusDeferred` as a first-class user action with date-based snooze and auto-return.

**Scope:** Z-key snooze action, defer_until timestamp, auto-return logic, `:deferred` command, additional status transitions, session metrics logging.

**Status:** COMPLETE — All 4 stories implemented and merged (PRs #310, #338, #358, #355)

**Requirements:** FR104-FR109 (already in requirements.md)

---

### Stories 28.1-28.4: Snooze/Defer Implementation

Four stories: 28.1 (Z-key action and SnoozeView with quick options), 28.2 (auto-return via startup check + tea.Tick), 28.3 (`:deferred` command with un-snooze and edit), 28.4 (status transitions and metrics).

---

## Epic 29: Task Dependencies & Blocked-Task Filtering

**Epic Goal:** Native dependency graph support. Blocks tasks with unmet dependencies from door selection.

**Scope:** `depends_on` field, automatic filtering, pessimistic orphan handling, blocked-by indicators, auto-unblock, dependency management UI, circular dependency detection, metrics logging, DependencyResolver.

**Status:** COMPLETE — All 4 stories done (PRs #307, #319, #340, #356)

**Requirements:** FR110-FR115 (already in requirements.md)

---

### Stories 29.1-29.4: Dependency Implementation

Four stories: 29.1 (depends_on field and door filtering), 29.2 (blocked-by indicators and auto-unblock), 29.3 (dependency management UI with +/- keys), 29.4 (circular detection and DependencyResolver).

---

## Epic 30: Linear Integration

**Epic Goal:** Integrate Linear as a task source via GraphQL API with full field mapping and bidirectional sync.

**Scope:** Linear GraphQL client, typed queries with cursor-based pagination, API key auth, read-only provider with full field mapping, bidirectional sync via GraphQL mutations, WAL offline queuing, contract tests.

**Status:** COMPLETE — All 4 stories done (PRs #446, #699, #705, #709)

**Requirements:** FR116-FR119 (referenced in epic-list.md)

---

### Stories 30.1-30.4: Linear Integration

Follows the standard 4-story integration pattern: 30.1 (GraphQL client with typed queries and pagination), 30.2 (read-only provider with status/priority/effort/label/due-date mapping), 30.3 (bidirectional sync with complete-task mutation and WAL), 30.4 (contract tests with mocked GraphQL server).

---

## Epic 31: Expand/Fork Key Implementations

**Epic Goal:** Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9.

**Scope:** E key expand mode, parent_id field, subtask list rendering, parent exclusion from doors, F key fork with field copy/reset semantics, fork cross-references via enrichment DB.

**Status:** COMPLETE — All 5 stories done (PRs #698, #701, #708, #714, #718)

**Requirements:** FR120-FR126 (already in requirements.md)

---

### Stories 31.1-31.5: Expand/Fork Implementation

Five stories: 31.1 (E key expand mode with sequential subtask creation), 31.2 (parent_id field and subtask list rendering), 31.3 (parent exclusion from door selection), 31.4 (F key fork with copy/reset semantics), 31.5 (fork cross-references and enrichment DB integration).

---

## Epic 32: Undo Task Completion

**Epic Goal:** Allow reversing accidental task completion via `complete → todo` transition.

**Scope:** Status transition, CompletedAt clearing, session metrics logging, immutable completed log preservation, dependency re-evaluation.

**Status:** COMPLETE — All 3 stories done

**Requirements:** FR127-FR131 (already in requirements.md)

---

### Stories 32.1-32.3: Undo Implementation

Three stories: 32.1 (complete→todo transition and CompletedAt clearing), 32.2 (undo_complete metrics event), 32.3 (dependency re-evaluation on undo).

---

## Epic 33: Seasonal Door Theme Variants

**Epic Goal:** Time-based seasonal theme variants that auto-switch based on current date, extending Epic 17's theme infrastructure.

**Scope:** Seasonal theme variant system, date-based auto-switching, seasonal theme definitions, configuration options.

**Status:** COMPLETE — All 4 stories done

---

### Stories 33.1-33.4: Seasonal Themes

Four stories implementing seasonal theme variants that auto-switch based on calendar date, extending the Epic 17 theme registry with time-aware variant selection.

---

## Epic 34: SOUL.md + Custom Development Skills

**Epic Goal:** Create SOUL.md project philosophy document and 4 custom Claude Code slash commands to improve AI agent alignment and developer workflow.

**Scope:** SOUL.md at project root, `/pre-pr` validation automation, `/validate-adapter` compliance checking, `/check-patterns` violation scanning, `/new-story` template generator.

**Status:** COMPLETE — All 4 stories implemented and merged (PRs #222, #224, #228, #230)

**Requirements:** NFR-DX1 through NFR-DX6 (already in requirements.md)

---

### Story 34.1: SOUL.md Project Philosophy Document

**Acceptance Criteria:**
1. SOUL.md at project root with philosophy, design principles, behavioral guidelines
2. Core tenets: progress over perfection, work with human nature, three doors not three hundred, local-first privacy-always, meet users where they are, solo human agent team

### Story 34.2: /pre-pr Validation Slash Command

**Acceptance Criteria:**
1. 8-step pre-PR validation: branch freshness, formatting, linting, tests, race detection, dead code, scope review, commit cleanliness

### Story 34.3: /validate-adapter Slash Command

**Acceptance Criteria:**
1. TaskProvider interface compliance check
2. Error wrapping pattern validation
3. Factory registration, test coverage, atomic write usage checks

### Story 34.4: /check-patterns and /new-story Slash Commands

**Acceptance Criteria:**
1. `/check-patterns`: scans for design pattern violations (direct status mutation, direct file writes, fmt.Println in TUI, panics, etc.)
2. `/new-story`: generates story files from standard template

---

## Epic 35: Door Visual Appearance — Door-Like Proportions

**Epic Goal:** Redesign all door themes to visually read as actual doors using portrait orientation, panel dividers, asymmetric handles, and threshold/floor lines.

**Scope:** DoorAnatomy helper type, height-aware Render signature, all themes updated with door-like proportions, compact mode fallback, shadow/depth effects, golden file test regeneration, accessibility validation.

**Status:** COMPLETE — All 7 stories implemented and merged (PRs #226, #229, #234, #236, #237, #238, #239)

**Requirements:** FR138-FR147

---

### Stories 35.1-35.7: Door Visual Redesign

Seven stories: 35.1 (DoorAnatomy type and height-aware rendering), 35.2 (classic theme door-like proportions), 35.3 (modern theme), 35.4 (minimal theme), 35.5 (art-deco theme), 35.6 (compact mode fallback for short terminals), 35.7 (shadow/depth effects and golden file regeneration).

---

## Epic 36: Door Selection Interaction Feedback

**Epic Goal:** Make door selection feel responsive and satisfying by enhancing visual feedback contrast, adding deselect toggle, and ensuring universal quit.

**Scope:** Selection highlight contrast enhancement, deselect toggle (press same key to deselect), universal quit handler, selection animation feedback.

**Status:** COMPLETE — All 4 stories done

**Requirements:** FR148-FR151

---

### Stories 36.1-36.4: Selection Feedback

Four stories: 36.1 (high-contrast selection highlight), 36.2 (deselect toggle — pressing same door key deselects), 36.3 (universal quit handler), 36.4 (selection animation feedback).

---

## Epic 37: Persistent BMAD Agent Infrastructure

**Epic Goal:** Enable autonomous project governance by adding persistent BMAD agents and cron jobs that maintain story status, ROADMAP accuracy, architecture doc currency, and quality metrics.

**Scope:** Agent definition files for project-watchdog and arch-watchdog, cron configuration for SM sprint health and QA coverage audit, agent communication architecture, monitoring and evaluation framework.

**Status:** COMPLETE — All 4 stories implemented (PR #271, PR #280, PR #279, PR #281)

---

### Story 37.1: Project-Watchdog Agent Definition

**Acceptance Criteria:**
1. Agent definition file at `agents/project-watchdog.md`
2. PM role: merged PR monitoring, story status updates, ROADMAP sync, PRD drift detection
3. Responsibility+WHY format per D-10

### Story 37.2: Arch-Watchdog Agent Definition

**Acceptance Criteria:**
1. Agent definition file at `agents/arch-watchdog.md`
2. Architect role: code-to-architecture-doc alignment, undocumented pattern detection, drift flagging
3. Responsibility+WHY format

### Story 37.3: SM Sprint Health & QA Coverage Cron

**Acceptance Criteria:**
1. Cron setup documentation at `docs/quality/cron-setup.md`
2. SM role: 4-hourly sprint status, blocked story detection, stale PR alerts
3. QA role: weekly coverage audit, regression flagging

### Story 37.4: Agent Communication Architecture

**Acceptance Criteria:**
1. Architecture doc at `_bmad-output/planning-artifacts/architecture-persistent-agent-infrastructure.md`
2. Monitoring framework at `docs/operations/agent-evaluation.md`
3. multiclaude message bus patterns documented

---

## Epic 38: Dual Homebrew Distribution

**Epic Goal:** Establish parallel Homebrew distribution channels (stable + alpha) with signing parity, publishing controls, verification, and retention management.

**Scope:** Alpha Homebrew formula, publishing toggle, code signing and notarization, alpha formula verification, alpha release retention cleanup.

**Status:** COMPLETE — All 6 stories done

---

### Stories 38.1-38.6: Dual Homebrew

Six stories implementing parallel distribution: alpha formula (`threedoors-a.rb`), publishing toggle (`vars.ALPHA_TAP_ENABLED`), code signing for stable releases, alpha formula verification via tap CI, alpha retention cleanup (keep last 30), documentation and testing.

---

## Epic 39: Keybinding Display System

**Epic Goal:** Add toggleable keybinding discoverability to the TUI: context-sensitive bottom bar, full keybinding overlay, and inline door key indicators unified under `h` toggle.

**Scope:** Compile-time keybinding registry, concise bottom bar, full-screen overlay, `h` toggle unifying door key indicators and bar, terminal size adaptation, context-sensitive content, `:hints` command.

**Status:** COMPLETE — 12/13 stories done, 1 cancelled (D-140: auto-fade rejected)

---

### Stories 39.1-39.13: Keybinding Display

Thirteen stories (one cancelled): 39.1 (keybinding registry), 39.2 (bottom bar rendering), 39.3 (full overlay with `?` key), 39.4 (terminal size adaptation), 39.5 (context-sensitive bar content), 39.6 (door key indicators `[a]`/`[w]`/`[d]`), 39.7 (`h` toggle unification per D-137), 39.8-39.11 (refinements and edge cases), 39.12 (auto-fade — CANCELLED per D-140), 39.13 (`:hints` command alias).

---

## Epic 40: Beautiful Stats Display

**Epic Goal:** Transform the insights dashboard from plain text into a visually delightful display using Lipgloss styled panels, gradient sparklines, bar charts, fun facts, heatmaps, and milestone celebrations.

**Scope:** Three-phase implementation: Phase 1 (dashboard shell, sparklines, fun facts), Phase 2 (bar charts, heatmap, hidden metrics, animated counters, tab navigation), Phase 3 (theme-matched colors, milestone celebrations).

**Status:** COMPLETE — All 10 stories done

---

### Stories 40.1-40.10: Beautiful Stats

Ten stories across three phases: 40.1 (Lipgloss bordered dashboard shell), 40.2 (gradient sparklines with colorblind-safe palette), 40.3 (fun facts engine), 40.4 (mood correlation bar charts), 40.5 (GitHub-style activity heatmap), 40.6 (surface hidden session metrics), 40.7 (animated counter reveals), 40.8 (tab navigation for Overview/Detail), 40.9 (theme-matched color palettes), 40.10 (milestone celebrations with observation language).

---

## Epic 41: Charm Ecosystem Adoption & TUI Polish

**Epic Goal:** Systematically adopt underutilized charmbracelet ecosystem components to reduce custom code maintenance, improve UX consistency, and deliver on SOUL.md's "physical objects" promise.

**Scope:** bubbles/spinner for async feedback, lipgloss layout composition, bubbles/viewport for scroll views, harmonica spring-physics spike, adaptive color profile support.

**Status:** COMPLETE — All 6 stories done

---

### Stories 41.1-41.6: Charm Ecosystem

Six stories: 41.1 (bubbles/spinner for provider operations), 41.2 (lipgloss.JoinVertical + Place for layout), 41.3 (bubbles/viewport replacing 3 custom scroll implementations), 41.4 (harmonica spring-physics door transition spike), 41.5 (adaptive color profile for terminal-aware degradation), 41.6 (integration testing and polish).

---

## Epic 42: Application Security Hardening

**Epic Goal:** Remediate all actionable findings from the application security audit — standardize file permissions, add symlink validation, enforce input size limits, protect credentials, and harden CI supply chain.

**Scope:** File permission standardization, symlink validation, file size limits, credential exposure protection, CI supply chain hardening.

**Status:** COMPLETE — All 5 stories done

---

### Story 42.1: File Permission Standardization

**Acceptance Criteria:**
1. All directory creation uses 0o700 permissions
2. All file creation uses 0o600 permissions
3. Startup migration for existing installs with incorrect permissions

### Story 42.2: Symlink Validation

**Acceptance Criteria:**
1. `os.Lstat()` check on startup for config/data directories
2. Symlink validation before file writes
3. Warning log for unexpected symlinks

### Story 42.3: Input Size Limits

**Acceptance Criteria:**
1. File size limits before YAML reads (prevent DoS via large files)
2. Explicit scanner buffer limits on all JSONL readers
3. Configurable limits with safe defaults

### Story 42.4: Credential Protection

**Acceptance Criteria:**
1. Credential exposure warning on startup (if tokens visible in config)
2. `yaml:"-"` tag on all token/secret fields
3. Config file permission check

### Story 42.5: CI Supply Chain Hardening

**Acceptance Criteria:**
1. SHA-pinned third-party GitHub Actions
2. govulncheck added to CI quality gate
3. Dependency review for known vulnerabilities

---

## Epic 43: Connection Manager Infrastructure

**Epic Goal:** Build the connection lifecycle layer for data source integrations: state machine, credential storage, config schema v3, CRUD operations, sync event logging, and adapter migration.

**Scope:** ConnectionManager with 7-state machine, CredentialStore with keychain/env/file fallback, config schema v3 with `connections:` array, connection CRUD, JSONL sync event logging, adapter migration.

**Status:** COMPLETE — All 6 stories implemented and merged (PRs #428, #442, #467, #526, #439, #540)

---

### Story 43.1: ConnectionManager State Machine

**Acceptance Criteria:**
1. `ConnectionManager` type with 7 states: Disconnected, Connecting, Connected, Syncing, Error, AuthExpired, Paused
2. Valid state transitions defined and enforced
3. Thread-safe state management

### Story 43.2: CredentialStore

**Acceptance Criteria:**
1. System keychain integration via 99designs/keyring
2. Env var fallback for CI/headless environments
3. Encrypted file fallback as last resort
4. Credential CRUD: store, retrieve, delete, rotate

### Story 43.3: Config Schema v3

**Acceptance Criteria:**
1. `connections:` array with named connections and ULID IDs
2. Auto-migration from config schema v2
3. Backward compatibility with existing configs

### Story 43.4: Connection CRUD Operations

**Acceptance Criteria:**
1. Add, remove, pause, resume, test, force-sync operations
2. Validation on all mutations
3. Event emission for state changes

### Story 43.5: Sync Event Logging

**Acceptance Criteria:**
1. JSONL sync event log per connection
2. Rolling retention policy
3. Event types: sync_start, sync_complete, sync_error, auth_expired

### Story 43.6: Adapter Migration

**Acceptance Criteria:**
1. All 8 existing adapters wrapped in ConnectionManager pattern
2. Backward-compatible initialization
3. Migration test coverage

---

## Epic 44: Sources TUI

**Epic Goal:** TUI interfaces for data source management: setup wizard, sources dashboard, source detail view, sync log view, status bar alerts, disconnection flow, and re-authentication flow.

**Scope:** 4-step setup wizard using charmbracelet/huh, sources dashboard, source detail view with health checks, sync log view, status bar health alerts, disconnect and re-auth flows.

**Status:** COMPLETE — All 7 stories done

---

### Stories 44.1-44.7: Sources TUI

Seven stories: 44.1 (`:connect` setup wizard with huh forms), 44.2 (sources dashboard with status indicators), 44.3 (source detail view with health checks and sync stats), 44.4 (sync log view with scrollable event history), 44.5 (status bar alerts for connections needing attention), 44.6 (disconnect flow with task preservation), 44.7 (re-authentication flow).

---

## Epic 45: Sources CLI

**Epic Goal:** Non-interactive CLI commands for data source management with consistent `--json` output for scripting and CI/automation.

**Scope:** `threedoors connect <provider>` with per-provider flags, `threedoors sources` commands (list, status, test, pause/resume/sync/disconnect, log), JSON output.

**Status:** COMPLETE — All 6 stories done

---

### Stories 45.1-45.6: Sources CLI

Six stories: 45.1 (`threedoors connect <provider>` with flag sets), 45.2 (`threedoors sources list` and `status`), 45.3 (`threedoors sources test` and health checks), 45.4 (`threedoors sources pause/resume/sync/disconnect`), 45.5 (`threedoors sources log` with filtering), 45.6 (JSON output and integration testing).

---

## Epic 46: OAuth Device Code Flow

**Epic Goal:** Generic OAuth device code flow client (RFC 8628) for browser-based authentication, with provider-specific integrations for GitHub and Linear.

**Scope:** Reusable device code flow client, GitHub OAuth integration with PAT fallback, Linear OAuth/API key integration, silent token refresh with re-auth on expiry.

**Status:** COMPLETE — All 4 stories done

---

### Stories 46.1-46.4: OAuth Implementation

Four stories: 46.1 (generic device code flow client per RFC 8628), 46.2 (GitHub OAuth with device code + PAT fallback), 46.3 (Linear OAuth/API key integration), 46.4 (silent token refresh lifecycle).

---

## Epic 47: Sync Lifecycle & Advanced Features

**Epic Goal:** Advanced sync features: conflict resolution, orphaned task handling, auto-detection of installed tools, and proactive connection health notifications.

**Scope:** ConflictResolver with field-level strategy, orphaned task marking and management UI, installed tool detection, predictive warnings.

**Status:** COMPLETE — All 4 stories done

---

### Stories 47.1-47.4: Sync Lifecycle

Four stories: 47.1 (ConflictResolver with remote-wins for metadata, local-wins for ThreeDoors fields), 47.2 (orphaned task marking and management), 47.3 (installed tool auto-detection for setup wizard), 47.4 (predictive warnings for token expiry, rate limits, error streaks).

---

## Epic 48: Door-Like Doors — Visual Door Metaphor Enhancement

**Epic Goal:** Transform rectangular card/panel doors into visually convincing doors using side-mounted handles, hinge marks, threshold lines, crack-of-light selection feedback, and handle turn micro-animations.

**Scope:** Side-mounted handle at right edge + hinge marks on left edge, continuous threshold/floor line, crack of light selection effect, handle turn micro-animation.

**Status:** COMPLETE — All 4 stories done

---

### Stories 48.1-48.4: Door-Like Doors

Four stories: 48.1 (side-mounted handles and hinge marks for asymmetry), 48.2 (continuous threshold/floor line grounding doors), 48.3 (crack-of-light effect on selection — door becomes "ajar"), 48.4 (handle turn micro-animation synced with spring physics).

---

## Epic 49: ThreeDoors Doctor — Self-Diagnosis Command

**Epic Goal:** Comprehensive self-diagnosis command (`threedoors doctor`) with flutter-style category-based output, conservative auto-repair, and channel-aware version checking.

**Scope:** Doctor command skeleton with DoctorChecker framework, 6 check categories, channel-aware version checking, conservative auto-repair, verbose/category/JSON modes.

**Status:** COMPLETE — All 10 stories done

---

### Stories 49.1-49.10: Doctor Command

Ten stories implementing the full doctor infrastructure: 49.1 (command skeleton and DoctorChecker framework), 49.2 (Environment checks), 49.3 (Task Data checks), 49.4 (Providers checks), 49.5 (Sessions checks), 49.6 (Sync checks), 49.7 (Database checks), 49.8 (channel-aware version checking with 24h cache), 49.9 (conservative auto-repair via `--fix`), 49.10 (verbose mode, category filter, JSON output).

---

## Epic 50: In-App Bug Reporting

**Epic Goal:** Add a `:bug` command for frictionless in-app bug reporting with navigation breadcrumb trail, automatic environment context, mandatory preview, and tiered submission.

**Scope:** Ring buffer breadcrumb tracking, bug report view, three submission paths (browser URL, GitHub API, local file), strict privacy allowlist.

**Status:** COMPLETE — All 3 stories implemented and merged (PRs #478, #624, #649)

---

### Story 50.1: Breadcrumb Tracking

**Acceptance Criteria:**
1. Ring buffer tracking 50 entries (view transitions + non-text keys)
2. Privacy-safe: task content, search queries, personal data never collected
3. Strict allowlist at capture level

### Story 50.2: Bug Report View

**Acceptance Criteria:**
1. Text description input
2. Environment summary (OS, terminal, Go version, ThreeDoors version)
3. Mandatory preview before submission

### Story 50.3: Submission Paths

**Acceptance Criteria:**
1. Browser URL: zero-auth, opens pre-filled GitHub issue URL
2. GitHub API: PAT-based direct issue creation (optional upgrade)
3. Local file: offline fallback saving report to disk

---

## Epic 51: SLAES — Self-Learning Agentic Engineering System

**Epic Goal:** Build a continuous improvement meta-system with a persistent `retrospector` agent that monitors PR merges, detects process waste, audits doc consistency, analyzes CI/conflict patterns, and files improvement recommendations.

**Scope:** Retrospector agent definition, JSONL findings log, saga detection, doc consistency audit, BOARD.md recommendation pipeline, merge conflict rate analysis, CI failure taxonomy, research lifecycle tracking, weekly trend reporting, 5 Watchmen safeguards.

**Status:** COMPLETE — All 11 stories implemented and merged (PRs #460-#465, #505-#509, #608)

---

### Stories 51.1-51.11: SLAES Implementation

Eleven stories across three phases:
- Phase 0 (51.1-51.2): Retrospector agent definition in responsibility+WHY format; rewrite 5 operational agent definitions with incident-hardened guardrails
- Phase 1 (51.3-51.7): JSONL findings log, per-merge lightweight retro, saga detection, doc consistency audit, BOARD.md recommendation pipeline with confidence scoring
- Phase 2 (51.8-51.11): Merge conflict rate analysis, CI failure taxonomy, research lifecycle tracking, PR creation authority, weekly trend reporting

5 Watchmen safeguards: no self-modification, audit trail, confidence scoring, periodic human review, kill switch (3 rejections → read-only).

---

## Epic 52: Envoy Three-Layer Firewall

**Epic Goal:** Restructure the envoy agent definition into a formal three-layer firewall architecture for issue screening.

**Scope:** Three-layer firewall (Layer 1: syntax/spam filter, Layer 2: scope/relevance check, Layer 3: impact assessment), clear entry/exit criteria per layer.

**Status:** COMPLETE — All 4 stories implemented and merged (PRs #515, #517, #514, #516)

---

### Stories 52.1-52.4: Envoy Firewall

Four stories: 52.1 (Layer 1: syntax and spam filtering), 52.2 (Layer 2: scope and relevance checking against ROADMAP), 52.3 (Layer 3: impact assessment and priority classification), 52.4 (integration and testing of full firewall pipeline).

---

## Epic 53: Remote Collaboration — multiclaude Cross-Machine Access

**Epic Goal:** Document and enable remote collaboration with multiclaude via SSH, with future MCP bridge support.

**Scope:** SSH remote access documentation, tmux session management for remote agents, MCP bridge architecture design, security considerations.

**Status:** COMPLETE — All 5 stories done (PRs #613, #615, #665, #691, #693)

---

### Stories 53.1-53.5: Remote Collaboration

Five stories: 53.1 (SSH tunnel setup and documentation), 53.2 (tmux session management for remote agent attachment), 53.3 (MCP bridge architecture design), 53.4 (security hardening for remote access), 53.5 (integration testing and operational guide).

---

## Epic 54: Gemini Research Supervisor — Deep Research Agent Infrastructure

**Epic Goal:** Deploy a persistent research-supervisor agent that wraps the official Gemini CLI (`@google/gemini-cli`) with OAuth authentication, providing web-grounded research with context packaging, result shielding, and dual-tier budget management.

**Scope:** Research-supervisor agent definition, Gemini CLI installation and OAuth setup, `scripts/gemini-research.sh` wrapper, `GEMINI.md` project context file, context packaging system (8 curated bundles, 60KB budget), three-layer result shielding, dual-tier rate limiting (50 Pro/day + 1,000 Flash/day).

**Status:** COMPLETE — All 5 stories done (PRs #537, #538, #664, #689, #690)

---

### Story 54.1: Research-Supervisor Agent Definition

**Acceptance Criteria:**
1. Agent definition file at `agents/research-supervisor.md` in Responsibility+WHY format
2. Polling loop with message bus integration
3. Request protocol for inter-agent research dispatch
4. Authority matrix defining research scope boundaries

### Story 54.2: Gemini CLI Installation, OAuth Setup & Wrapper Script

**Acceptance Criteria:**
1. `@google/gemini-cli` installed via npm
2. OAuth authentication flow documented
3. `scripts/gemini-research.sh` wrapper with structured JSON output
4. No API key required — OAuth tokens cached and auto-refreshed

### Story 54.3: Context Packaging & Prompt Engineering

**Acceptance Criteria:**
1. 8 curated context bundles (architecture, PRD, stories, etc.) within 60KB budget
2. Keyword auto-detection for bundle selection
3. `--include-directories` integration for Gemini CLI file context
4. Priority shedding order for oversized contexts

### Story 54.4: Result Shielding & Artifact Storage

**Acceptance Criteria:**
1. Three-layer shielding: executive summary → detailed report → raw JSON
2. Executive summary fits in agent context window
3. Full reports stored on disk at `_bmad-output/research-results/`
4. JSON response parsing from `gemini --output-format json`

### Story 54.5: Rate Limiting, Budget Management & Model Selection

**Acceptance Criteria:**
1. Daily query count tracking per model tier (budget.json)
2. Pro vs Flash model selection (deep analysis vs quick lookups)
3. Priority queue with deduplication of similar queries
4. Reserve 5 Pro queries after 6pm UTC
5. Batch optimization for related queries

---

## Epic 55: CI Optimization Phase 1

**Epic Goal:** Reduce PR CI wall clock time from 3m33s to ~2m08s through CI configuration changes only — no test code modifications.

**Scope:** Docker E2E moved to push-only, golangci-lint version fix, benchmark path filtering, `make test-fast` target, CI Go build cache verification.

**Status:** COMPLETE — All 3 stories done

---

### Stories 55.1-55.3: CI Optimization

Three stories: 55.1 (Docker E2E push-only + lint version fix), 55.2 (benchmark path filtering — only run when core code changes), 55.3 (local dev acceleration with `make test-fast` ~10s target).

---

## Epic 56: Door Visual Redesign — Three-Layer Depth System

**Epic Goal:** Transform door rendering from imperceptible wireframe shadows into solid, 3D-feeling surfaces using background fill, bevel lighting, and gradient shadow.

**Scope:** ThemeColors extension with depth fields, background fill for all themes, bevel lighting (lighter top/left, darker bottom/right), shadow overhaul with gradient (▓▒░), panel zone shading, width-adaptive shadow.

**Status:** COMPLETE — All 5 stories done

---

### Stories 56.1-56.5: Door Visual Redesign

Five stories: 56.1 (ThemeColors extension with FillLower, Highlight, ShadowEdge, ShadowNear, ShadowFar), 56.2 (background fill for all 8 theme door interiors), 56.3 (bevel lighting for raised-surface perception), 56.4 (shadow overhaul with gradient and per-theme rendering), 56.5 (panel zone shading and width-adaptive shadow).

---

## Epic 57: LLM CLI Services

**Epic Goal:** Enable ThreeDoors to invoke LLM CLI tools (Claude CLI, Gemini CLI, Ollama CLI) as subprocess-based service providers for intelligent task operations.

**Scope:** CLIProvider + CLISpec model implementing LLMBackend interface, pre-built specs for 4 CLI tools, auto-discovery with fallback chain, TaskExtractor/TaskEnricher/TaskBreakdown services, TUI and CLI commands, `threedoors llm status`.

**Status:** COMPLETE — All 8 stories done

---

### Stories 57.1-57.8: LLM CLI Services

Eight stories: 57.1 (CLIProvider and CLISpec declarative model), 57.2 (pre-built specs for Claude/Gemini/Ollama/custom CLIs), 57.3 (auto-discovery and fallback chain), 57.4 (TaskExtractor service and `:extract` TUI command), 57.5 (CLI `threedoors extract` with file/clipboard/stdin), 57.6 (TaskEnricher with before/after diff TUI), 57.7 (TaskBreakdown CLI backend extension), 57.8 (`threedoors llm status` command).

---

## Epic 58: Supervisor Shift Handover — Context-Aware Supervisor Rotation

**Epic Goal:** Detect supervisor context window degradation, serialize operational state, and transfer control to a fresh supervisor instance while workers continue uninterrupted.

**Scope:** Shift clock with three-tier thresholds, rolling state snapshot, handover orchestrator with 5-step protocol, supervisor startup with state file, emergency handover, handover history, manual trigger command.

**Status:** COMPLETE — All 7 stories done

---

### Stories 58.1-58.7: Supervisor Shift Handover

Seven stories in two phases:
- MVP (58.1-58.4): Shift clock (daemon-side transcript monitoring), rolling state snapshot (YAML), handover orchestrator (request→delta→spawn→ready→kill), supervisor startup with state file
- Hardening (58.5-58.7): Emergency handover (120s timeout force-kill), handover history (archived state files, JSONL log), manual trigger (`multiclaude supervisor handover`)

---

## Epic 59: Full-Terminal Vertical Layout

**Epic Goal:** Transform ThreeDoors from a content-driven partial-terminal app into a full-terminal experience using AltScreen, layout engine, and graceful degradation.

**Scope:** AltScreen for full-terminal ownership, layout engine (fixed header + flex middle + fixed footer), door height capping, perceptual centering, help view dynamic page size, SetHeight propagation, header/footer extraction, breakpoint-based degradation.

**Status:** COMPLETE — All 2 stories done

---

### Stories 59.1-59.2: Full-Terminal Layout

Two stories per D-120: 59.1 (MVP — AltScreen, layout engine, door height cap, perceptual centering, degradation breakpoints), 59.2 (follow-up — SetHeight propagation to all views, header/footer extraction, help view dynamic sizing).

---

## Epic 60: README Overhaul

**Epic Goal:** Polish the README with centered badge clusters, table of contents, foldable reference sections, updated feature list, and visual demo section.

**Scope:** Centered badge cluster, table of contents, foldable `<details>` sections, feature list audit, visual demo section.

**Status:** COMPLETE — All 5 stories done

---

### Stories 60.1-60.5: README Overhaul

Five stories: 60.1 (centered badge cluster), 60.2 (table of contents with anchor links), 60.3 (foldable reference sections), 60.4 (feature list audit covering all completed epics), 60.5 (visual demo section with screenshots).

---

## Epic 61: GitHub Pages User Guide

**Epic Goal:** Publish ThreeDoors documentation as a professional GitHub Pages site using MkDocs + Material for MkDocs.

**Scope:** MkDocs infrastructure, GitHub Actions deployment workflow, getting started guides, core user guide, per-integration setup guides, CLI reference, advanced guides, dark/light mode, search.

**Status:** COMPLETE — All 4 stories done

---

### Stories 61.1-61.4: GitHub Pages

Four stories: 61.1 (MkDocs + Material infrastructure and GitHub Actions workflow), 61.2 (getting started and core user guide pages), 61.3 (per-integration setup guides for 8 providers), 61.4 (CLI reference, config reference, advanced guides, troubleshooting).

---

## Epic 62: Retrospector Agent Reliability

**Epic Goal:** Fix three infrastructure reliability issues preventing the retrospector agent (SLAES) from operating as designed.

**Scope:** File-based fallback inbox for messaging, recommendation queue with batch pipeline to BOARD.md, JSON checkpoint for state persistence.

**Status:** COMPLETE — All 3 stories done

---

### Stories 62.1-62.3: Retrospector Reliability

Three stories: 62.1 (file-based fallback inbox with identity verification — D-176), 62.2 (recommendation queue file with batch pipeline to BOARD.md via project-watchdog — D-177), 62.3 (JSON checkpoint file for analytical state persistence across restarts — D-178).

---

## Epic 63: ClickUp Integration

**Epic Goal:** Integrate ClickUp as a task source following the established adapter pattern.

**Scope:** ClickUp REST API v2 client with token auth, read-only provider with field mapping, bidirectional sync with WAL and circuit breaker, contract tests.

**Status:** COMPLETE — All 4 stories done (PRs #706, #719, #727, #728)

---

### Stories 63.1-63.4: ClickUp Integration

Follows the standard 4-story integration pattern: 63.1 (REST API v2 client with token auth), 63.2 (read-only provider with status/priority/due-date/tag mapping), 63.3 (bidirectional sync with WAL integration and circuit breaker), 63.4 (contract tests and integration testing).

---

## Epic 65: CLI Test Coverage Hardening

**Epic Goal:** Increase `internal/cli` package test coverage from 34.8% to ≥70%, addressing the critical coverage gap identified by the TEA audit.

**Scope:** Core CLI path tests, subcommand tests, remaining command tests.

**Status:** COMPLETE — All 3 stories done

---

### Story 65.1: Core CLI Path Tests

**Acceptance Criteria:**
1. Tests for bootstrap, root command, doors command, loadTaskPool, output formatting
2. Coverage increase target: +15-20%

### Story 65.2: Subcommand Test Coverage

**Acceptance Criteria:**
1. Tests for config, mood, health, stats, plan, version subcommands
2. Coverage increase target: +15-20%

### Story 65.3: Remaining Command Tests

**Acceptance Criteria:**
1. Tests for task, sources, connect, extract, interactive commands
2. Final coverage ≥70%

## Epic 67: Retrospector Operational Data Pipeline

**Epic Goal:** Automate periodic sync of retrospector operational data (`docs/operations/`) to git via a cron-triggered project-watchdog pipeline, so worker agents in isolated worktrees have access to current operational data.

**Scope:** CronCreate setup for `SYNC_OPERATIONAL_DATA` trigger, supervisor standing orders for cron creation and staleness verification. The project-watchdog handler already exists (added during Epic 62 reliability work).

**Status:** Not Started (0/1 done)

**Prerequisites:** Epic 62 (Retrospector Agent Reliability) — COMPLETE

**Triggered by:** Party mode research: `_bmad-output/planning-artifacts/retrospector-data-pipeline-party-mode.md`

**NFRs covered:** NFR-ODP1 through NFR-ODP5

---

### Story 67.1: Cron-Triggered Operational Data Sync Pipeline

**As a** multiclaude operator,
**I want** retrospector operational data to be automatically committed to git every 3 hours,
**so that** worker agents in isolated worktrees can access current operational data and the data is durable in version control.

**Acceptance Criteria:**
1. Supervisor startup checklist includes a `CronCreate("0 */3 * * *", "multiclaude message send project-watchdog SYNC_OPERATIONAL_DATA")` entry
2. project-watchdog receives `SYNC_OPERATIONAL_DATA` messages and creates data-sync PRs when `docs/operations/` has changes (handler already exists — verify it works)
3. If no files changed, no commit or PR is created (idempotency)
4. Supervisor has a standing order to periodically verify data freshness: `git log --oneline docs/operations/ --since="8 hours ago"` — investigate if no commits and retrospector is active
5. All changes go through PR workflow (branch protection compliance)
6. Data sync PRs use the standard `data-sync/<timestamp>` branch naming convention

**Technical Notes:**
- project-watchdog definition already contains the `SYNC_OPERATIONAL_DATA Response Protocol` (lines 80-121 of `agents/project-watchdog.md`)
- This story primarily involves supervisor configuration (cron setup, standing orders) and verification
- Cron interval: every 3 hours (party mode recommended 2-4 hours; 3 hours balances freshness vs. PR noise)
- Staleness SLA: 6 hours (2x commit interval)

**Estimated Time:** 30-45 minutes

---

## Epic 66: CLI/TUI Adapter Wiring Parity (PROVISIONAL)

**Epic Goal:** Fix three gaps where implemented adapter code is not properly connected to CLI and TUI entry points, ensuring all registered adapters are accessible through both the CLI and TUI connect flows with proper flag validation.

**Scope:** CLI adapter registration ordering bug (critical), ClickUp connect wiring, and provider spec parity across CLI/TUI. All work targets `cmd/threedoors/main.go`, `internal/cli/connect.go`, and `internal/tui/connect_wizard.go`.

**Triggered by:** Unwired features audit (2026-03-13) — see `_bmad-output/planning-artifacts/unwired-features-audit.md`

**FRs covered:** FR152, FR153, FR154, FR155, FR156, NFR24

---

### Story 66.1: CLI Adapter Registration Fix

**As a** user with a non-textfile provider configured,
**I want** CLI commands to work correctly,
**so that** I can manage my tasks from the command line regardless of which provider I use.

**Acceptance Criteria:**
1. `registerBuiltinAdapters(core.DefaultRegistry())` is called before the CLI/TUI routing branch in `main()`
2. CLI commands (`task list`, `doors`, etc.) work correctly with non-textfile providers
3. Regression test verifies the adapter registry is populated before CLI command execution
4. TUI path continues to work correctly
5. `go test -race ./cmd/threedoors/... ./internal/cli/...` passes

**Estimated Time:** 30 minutes

---

### Story 66.2: ClickUp Connect Wiring

**As a** ClickUp user,
**I want** to connect my ClickUp workspace via `threedoors connect clickup` or the TUI `:connect` wizard,
**so that** I can use ThreeDoors with my ClickUp tasks.

**Acceptance Criteria:**
1. `clickup` added to CLI `knownProviderSpecs` with correct flag spec
2. `clickup` added to TUI `DefaultProviderSpecs()`
3. `threedoors connect clickup --label "My ClickUp" --token <token> --list-id <id>` works end-to-end
4. TUI `:connect` wizard shows ClickUp as selectable

**Estimated Time:** 30 minutes

**Dependencies:** Story 66.1

---

### Story 66.3: Provider Spec Parity & Validation

**As a** user connecting any supported provider via CLI,
**I want** required flags to be validated,
**so that** I get clear error messages instead of silently incomplete configurations.

**Acceptance Criteria:**
1. `knownProviderSpecs` has entries for all 9 registered providers
2. Required flags enforced per provider
3. Parity test verifies `registerBuiltinAdapters()`, `knownProviderSpecs`, `ValidArgs`, and `DefaultProviderSpecs()` are in sync
4. Connect command help text lists all supported providers

**Estimated Time:** 60-90 minutes

**Dependencies:** Story 66.2

---

## Epic 64: Cross-Computer Sync

**Epic Goal:** Enable task data synchronization across multiple computers using a Git-based sync transport with device identity, conflict resolution, and offline queue reconciliation.

**Scope:** Architecture research spike (ADR), device identity and registration system, Git-based sync transport, cross-machine conflict resolution (last-writer-wins with vector clocks), offline queue and reconciliation, E2E test suite.

**Status:** COMPLETE — All 6 stories done (PRs #715, #721, #731, #734, #743, #748)

---

### Story 64.1: Architecture Research Spike

**Acceptance Criteria:**
1. ADR document evaluating 3+ sync approaches with pros/cons/rejected rationale
2. Transport mechanism chosen (git-based sync, cloud intermediary, P2P, or hybrid)
3. Conflict resolution strategy chosen
4. Device identity mechanism designed
5. Sync scope defined (which data syncs, which doesn't)
6. Security model documented
7. Party mode artifact saved

### Story 64.2: Device Identity & Registration

**Acceptance Criteria:**
1. UUID v5 device identity seeded from machine ID + install path
2. `DeviceRegistry` interface with `LocalDeviceRegistry` implementation
3. CLI `threedoors device` commands for listing and managing devices
4. Auto-registration on first run

### Story 64.3: Sync Transport Layer

**Acceptance Criteria:**
1. Git-based sync transport implementation
2. `GitDeviceRegistry` extending local registry to sync repo
3. Sync setup wizard for initial configuration

### Story 64.4: Conflict Resolution

**Acceptance Criteria:**
1. Vector clock or timestamp-based conflict detection
2. Resolution strategy per ADR (field-level or document-level)
3. Conflict visualization for manual resolution

### Story 64.5: Offline Queue & Reconciliation

**Acceptance Criteria:**
1. Offline change queue with replay on reconnection
2. Reconciliation of divergent state
3. Merge of independent offline changes

### Story 64.6: E2E Test Suite

**Acceptance Criteria:**
1. Multi-device sync simulation tests
2. Conflict scenario coverage
3. Offline/online transition tests

---

## Epic 69: TUI MainModel Decomposition

**Epic Goal:** Refactor `internal/tui/main_model.go` (2991 lines, 32 ViewModes) from a monolithic god object into focused files using Go's same-receiver-different-file pattern, reducing merge conflict surface and improving maintainability.

**Scope:** Four-phase extraction: view transition/navigation logic, source/sync view controllers, planning/task management controllers, auxiliary views and command dispatch.

**Status:** In Progress (1/4 done)

---

### Story 69.1: Extract View Transition & Navigation Logic

**As a** ThreeDoors developer,
**I want** view transition and navigation logic extracted into `view_navigation.go`,
**so that** navigation behavior is isolated, testable, and reduces merge conflicts.

**Acceptance Criteria:**
1. New file `internal/tui/view_navigation.go` with `setViewMode()`, `goBack()`, and view stack management
2. `main_model.go` reduced by ~200-300 lines
3. All existing tests pass unchanged
4. `go test -race ./internal/tui/...` passes
5. No user-visible behavior changes

### Story 69.2: Extract Source/Sync View Controllers

**Acceptance Criteria:**
1. Source dashboard, source detail, sync log view methods extracted
2. Related Update/View dispatch cases moved to new file

### Story 69.3: Extract Planning & Task Management Controllers

**Acceptance Criteria:**
1. Planning mode, deferred view, dependency management methods extracted
2. Related Update/View dispatch cases moved

### Story 69.4: Extract Auxiliary Views & Command Dispatch

**Acceptance Criteria:**
1. Help, bug report, settings, command palette handlers extracted
2. Command dispatch logic consolidated

---

## Epic 70: Completion History & Progress View

**Epic Goal:** New `:history` TUI view and `threedoors history` CLI command for browsing completed tasks with aggregated statistics, celebrating user progress per SOUL.md's "Progress Over Perfection" principle.

**Scope:** Completion data reader/aggregator, TUI history view with filtering and stats, CLI history command with JSON output.

**Status:** In Progress (1/3 done)

---

### Story 70.1: Completion Data Reader & Aggregator

**As a** ThreeDoors user,
**I want** a data layer that reads and aggregates my completed tasks,
**so that** the history view can display what I've accomplished.

**Acceptance Criteria:**
1. `CompletionRecord` type with Title, CompletedAt, Source, TaskID fields
2. `CompletionReader` with `Read()`, `Today()`, `ThisWeek()` methods
3. Reads from `completed.txt` append-only log
4. `CompletionStats` aggregation (totals, streaks, averages)

### Story 70.2: History TUI View

**Acceptance Criteria:**
1. `:history` command opens scrollable completion history
2. Filter by timeframe (today, this week, all time)
3. Aggregated stats display (total, streak, per-day average)
4. Source attribution per entry

### Story 70.3: History CLI Command

**Acceptance Criteria:**
1. `threedoors history` command with `--json` output
2. `--days N` flag for timeframe filtering
3. Human-readable and JSON output modes
4. Stats summary included in output

---

## Epic 72: Operationalize GitHub Label Usage

**Epic Goal:** Wire GitHub label application into agent workflows so that PRs are routinely labeled, issue labeling is resilient to envoy downtime, and mutual exclusivity is enforced. Closes the gap between label infrastructure (27-label taxonomy, authority matrix) and actual label usage (zero PR labels, inconsistent issue labels).

**Scope:** Agent definition text changes (`agents/merge-queue.md`, `agents/envoy.md`), operational doc updates (`docs/operations/label-authority.md`), supervisor memory updates, one GitHub API label creation, and one-time retroactive cleanup of existing issues.

**Status:** Not Started (0/4 stories)

**Reference:** PR #806 (gap analysis), D-184, D-185

---

### Story 72.1: Merge-Queue PR Labeling

**As a** project maintainer,
**I want** the merge-queue agent to automatically apply type, scope, and agent labels to PRs during merge validation,
**so that** all merged PRs are consistently classified for filtering and dashboard queries.

**Acceptance Criteria:**
1. PR labeling section added to `agents/merge-queue.md` with title-prefix inference mapping
2. `scope.in-scope` applied for PRs referencing stories
3. `agent.worker` applied for PRs from `work/*` branches
4. Label authority matrix updated for merge-queue
5. Merge-queue Labels table updated with new entries

---

### Story 72.2: Envoy Label Resilience

**As a** project maintainer,
**I want** the envoy agent to catch up on missed labels after restart, warn about context exhaustion, and enforce label mutual exclusivity,
**so that** issue labeling is reliable regardless of envoy downtime.

**Acceptance Criteria:**
1. Startup catch-up scan added to envoy "Your rhythm" section
2. Context Exhaustion Risk section added (matching merge-queue/pr-shepherd format)
3. Mutual exclusivity enforcement instructions added

---

### Story 72.3: Supervisor Label Discipline & Missing Label Creation

**As a** project maintainer,
**I want** the supervisor to apply labels on dispatch/scope/closure and the missing `resolution.wontfix` label to be created,
**so that** the full 27-label taxonomy is operational.

**Acceptance Criteria:**
1. Supervisor label discipline documented (agent.worker on dispatch, scope.* on decisions, resolution.* on closure)
2. `resolution.wontfix` label created on GitHub with correct color and description
3. Label authority summary table updated

---

### Story 72.4: Retroactive Label Cleanup

**As a** project maintainer,
**I want** all currently unlabeled open issues to receive appropriate labels and mutual exclusivity violations fixed,
**so that** the GitHub issue dashboard is clean and accurate.

**Acceptance Criteria:**
1. All open issues have at minimum `triage.*` and `type.*` labels
2. Issue #330 mutual exclusivity violation fixed (remove `triage.in-progress`)
3. Issue #296 mutual exclusivity violation fixed (remove `type.bug`)
4. No remaining mutual exclusivity violations across open issues

---
