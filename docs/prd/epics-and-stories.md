---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/index.md (sharded PRD - 10 files)
  - docs/architecture/index.md (sharded Architecture - 19 files)
  - docs/prd/user-interface-design-goals.md (UX embedded in PRD)
---

# ThreeDoors - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for ThreeDoors, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

**Technical Demo Phase:**
- TD1: The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)
- TD2: The system shall read tasks from a simple local text file (~/.threedoors/tasks.txt)
- TD3: The system shall display the Three Doors interface showing three tasks selected from the text file
- TD4: The system shall allow door selection via A/Left, W/Up, D/Right keys with no initial selection after launch or re-roll
- TD5: The system shall provide a refresh mechanism via S/Down to generate a new set of three doors
- TD6: The system shall display doors with dynamic width adjustment based on terminal size
- TD7: The system shall respond to task management keystrokes: c (complete), b (blocked), i (in-progress), e (expand), f (fork), p (procrastinate)
- TD8: The system shall embed "progress over perfection" messaging in the interface
- TD9: The system shall write completed tasks to a separate file (~/.threedoors/completed.txt) with timestamp

**Full MVP Phase:**
- FR2: The system shall integrate with Apple Notes as primary task storage backend with bidirectional sync
- FR3: The system shall allow task capture with optional context (what and why) through CLI/TUI
- FR4: The system shall retrieve and display tasks from Apple Notes
- FR5: The system shall mark tasks complete, updating both app state and Apple Notes
- FR6: The system shall display user-defined values and goals persistently throughout sessions
- FR7: The system shall provide choose-your-own-adventure interactive navigation
- FR8: The system shall track daily task completion count with day-over-day comparison
- FR9: The system shall prompt user once per session for improvement suggestion
- FR10: The system shall embed enhanced "progress over perfection" messaging
- FR11: The system shall maintain a local enrichment layer (SQLite/vector DB) for metadata and cross-references
- FR12: The system shall support bidirectional sync with Apple Notes on iPhone
- FR15: The system shall provide a health check command for Apple Notes connectivity
- FR16: The system shall support quick add mode for minimal-interaction task capture
- FR18: The system shall allow door feedback options (Blocked, Not now, Needs breakdown, Other comment)
- FR19: The system shall capture and store blocker information when task marked blocked
- FR20: The system shall use door selection and feedback patterns to inform future door selection (learning)
- FR21: The system shall categorize tasks by type, effort level, and context for diverse door selection

### Non-Functional Requirements

**Technical Demo Phase:**
- TD-NFR1: Go 1.25.4+ with gofumpt formatting standards
- TD-NFR2: Bubbletea/Charm Bracelet ecosystem for TUI
- TD-NFR3: macOS primary target platform
- TD-NFR4: Local text files only, no external services or telemetry
- TD-NFR5: <100ms latency for typical operations
- TD-NFR6: Make build system with build, run, clean targets
- TD-NFR7: Graceful handling of missing or corrupted task files

**Full MVP Phase:**
- NFR1: Idiomatic Go patterns and gofumpt formatting
- NFR2: Continue Bubbletea/Charm ecosystem
- NFR3: macOS primary platform
- NFR4: Local or iCloud storage (via Apple Notes), no external telemetry
- NFR5: Local application state and enrichment data (cross-computer sync deferred)
- NFR6: <500ms latency for typical operations
- NFR7: Graceful degradation when Apple Notes unavailable
- NFR8: OS keychain for credential/token storage
- NFR9: No sensitive data logging
- NFR10: Make build system
- NFR11: Clear architectural separation (core, TUI, adapters, enrichment)
- NFR12: Data integrity during external Apple Notes modification

**Code Quality & Submission Standards (Cross-Cutting):**
- NFR-CQ1: All code must pass gofumpt formatting before submission
- NFR-CQ2: All code must pass golangci-lint with zero issues before submission
- NFR-CQ3: All branches must be rebased onto upstream/main before PR creation
- NFR-CQ4: All PRs must have clean git diff --stat showing only in-scope changes
- NFR-CQ5: All fix-up commits must be squashed before PR submission

### Additional Requirements

**From Architecture:**
- Greenfield Go project (no starter template) - go mod init
- Two-layer architecture: TUI layer (internal/tui) + Domain layer (internal/tasks)
- MVU pattern mandatory (Bubbletea enforced Elm Architecture)
- Structured YAML data format for tasks with metadata (status, notes, timestamps)
- Five-state task lifecycle: todo → blocked → in-progress → in-review → complete
- Atomic writes for all file persistence (write-to-temp, fsync, rename)
- UUID v4 for task identification
- Constructor injection for dependency management
- Repository/adapter pattern (TaskProvider interface) deferred to Epic 2
- Ring buffer for recently-shown door tracking (default size: 10)
- Fisher-Yates shuffle for random door selection
- Apple Notes integration spike required before Epic 2 implementation
- Unit tests for core domain logic (70%+ coverage target in MVP)
- Integration tests for backend adapters
- CI/CD via GitHub Actions (MVP phase)

**From UX Design:**
- Three doors rendered horizontally with dynamic terminal width adjustment
- No "Door X" labels - clean, uncluttered presentation
- Door opening animation (optional) with expanded detail view
- Context-aware Esc navigation (returns to previous context: doors or search)
- Bottom-up search results display (reduces eye travel)
- Multiple navigation schemes (arrows, WASD, HJKL vi-style)
- Command palette via : prefix for power-user features
- Terminal aesthetic with warmth - Lipgloss styling with green/yellow/red status colors
- "Progress over perfection" visual language with asymmetry and celebration messaging
- 80x24 minimum terminal, responsive to larger sizes
- 256-color support minimum

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| TD1 | Epic 1 | CLI/TUI interface |
| TD2 | Epic 1 | Read tasks from text file |
| TD3 | Epic 1 | Three Doors display |
| TD4 | Epic 1 | Door selection keys |
| TD5 | Epic 1 | Refresh mechanism |
| TD6 | Epic 1 | Dynamic width adjustment |
| TD7 | Epic 1 | Task management keystrokes |
| TD8 | Epic 1 | Progress over perfection messaging |
| TD9 | Epic 1 | Completed tasks file with timestamp |
| FR2 | Epic 2 | Apple Notes bidirectional sync |
| FR3 | Epic 3 | Task capture with context |
| FR4 | Epic 2 | Retrieve/display from Apple Notes |
| FR5 | Epic 2 | Mark complete in Apple Notes |
| FR6 | Epic 3 | Persistent values/goals display |
| FR7 | Epic 3 | Choose-your-own-adventure navigation |
| FR8 | Epic 3 | Daily completion tracking |
| FR9 | Epic 3 | Session improvement prompt |
| FR10 | Epic 3 | Enhanced messaging |
| FR11 | Epic 5 | Local enrichment layer |
| FR12 | Epic 2 | Bidirectional iPhone sync |
| FR15 | Epic 2 | Health check command |
| FR16 | Epic 3 | Quick add mode |
| FR18 | Epic 3 | Door feedback options |
| FR19 | Epic 3 | Blocker capture |
| FR20 | Epic 4 | Learning from patterns |
| FR21 | Epic 4 | Task categorization |
| FR27 | Epic 8 | Obsidian vault integration |
| FR28 | Epic 8 | Obsidian bidirectional sync |
| FR29 | Epic 8 | Obsidian vault configuration |
| FR30 | Epic 8 | Obsidian daily note integration |
| FR31 | Epic 7 | Adapter registry |
| FR32 | Epic 7 | Config-driven provider selection |
| FR33 | Epic 7 | Adapter developer guide |
| FR34 | Epic 15 | Psychology research documentation |
| FR35 | Epic 14 | LLM task decomposition |
| FR36 | Epic 14 | Git repo output for coding agents |
| FR37 | Epic 14 | Configurable LLM backends |
| FR38 | Epic 10 | First-run welcome flow |
| FR39 | Epic 10 | Import from existing sources |
| FR40 | Epic 11 | Offline-first with local queue |
| FR41 | Epic 11 | Sync status indicator |
| FR42 | Epic 11 | Conflict visualization |
| FR43 | Epic 11 | Sync log |
| FR44 | Epic 12 | Local calendar reading |
| FR45 | Epic 12 | Time-contextual door selection |
| FR46 | Epic 13 | Cross-provider task aggregation |
| FR47 | Epic 13 | Duplicate detection |
| FR48 | Epic 13 | Source attribution in TUI |
| FR49 | Epic 9 | Apple Notes integration E2E tests |
| FR50 | Epic 9 | Contract tests for adapters |
| FR51 | Epic 9 | Functional E2E tests |

## Epic List

### Epic 1: Three Doors Technical Demo & Validation
**Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists. User can launch a TUI app, see three random tasks from a text file, select doors, manage task status, search tasks, track moods, and collect session metrics for concept validation.
**FRs covered:** TD1, TD2, TD3, TD4, TD5, TD6, TD7, TD8, TD9
**NFRs covered:** TD-NFR1, TD-NFR2, TD-NFR3, TD-NFR4, TD-NFR5, TD-NFR6, TD-NFR7
**Dependencies:** None (greenfield)
**Timeline:** Week 1 (4-8 hours)

### Epic 2: Foundation & Apple Notes Integration
**Goal:** Replace/augment text file backend with Apple Notes integration, enabling bidirectional sync so tasks edited on iPhone appear in the terminal and vice versa. Includes architectural refactoring to adapter pattern.
**FRs covered:** FR2, FR4, FR5, FR12, FR15
**NFRs covered:** NFR7, NFR8, NFR11, NFR12
**Dependencies:** Epic 1 (validated concept)
**Timeline:** 3-4 weeks at 2-4 hrs/week
**Prerequisites:** Apple Notes integration spike, validation gate passed

### Epic 3: Enhanced Task Capture & Interaction
**Goal:** Enrich the task management experience with quick task capture, contextual capture (what and why), persistent values/goals display, door feedback mechanisms, daily completion tracking, and session improvement prompts.
**FRs covered:** FR3, FR6, FR7, FR8, FR9, FR10, FR16, FR18, FR19
**NFRs covered:** NFR6
**Dependencies:** Epic 2 (stable backend integration)
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 4: Learning & Intelligent Door Selection
**Goal:** Use historical session metrics to analyze user patterns and adapt door selection. The system learns from selection patterns, mood correlations, and avoidance behaviors to present smarter door choices adapted to current context.
**FRs covered:** FR20, FR21
**Dependencies:** Epic 3 (sufficient usage data)
**Timeline:** 3-4 weeks at 2-4 hrs/week

### Epic 5: Data Layer & Enrichment (Optional)
**Goal:** Add enrichment storage layer for cross-system metadata, richer task relationships, and persistent enrichment data spanning multiple task sources. Only implement if clear need emerges.
**FRs covered:** FR11
**NFRs covered:** NFR5
**Dependencies:** Epic 4 (proven need)
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 7: Plugin/Adapter SDK & Registry
**Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide.
**FRs covered:** FR31, FR32, FR33
**NFRs covered:** NFR11
**Dependencies:** Epic 2 (adapter pattern established)
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 8: Obsidian Integration (P0 - #2 Integration)
**Goal:** Add Obsidian vault as second task storage backend. Local-first Markdown with bidirectional sync.
**FRs covered:** FR27, FR28, FR29, FR30
**Dependencies:** Epic 7 (adapter SDK)
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 9: Testing Strategy & Quality Gates
**Goal:** Comprehensive testing infrastructure with integration, contract, performance, and E2E tests.
**FRs covered:** FR49, FR50, FR51
**NFRs covered:** NFR13, NFR16
**Dependencies:** Epic 2, Epic 7
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 10: First-Run Onboarding Experience
**Goal:** Guided welcome flow for new users.
**FRs covered:** FR38, FR39
**Dependencies:** Epic 3
**Timeline:** 1-2 weeks at 2-4 hrs/week

### Epic 11: Sync Observability & Offline-First
**Goal:** Robust offline-first operation with sync status visibility, conflict visualization, and debugging.
**FRs covered:** FR40, FR41, FR42, FR43
**NFRs covered:** NFR14
**Dependencies:** Epic 2
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 12: Calendar Awareness (Local-First, No OAuth)
**Goal:** Time-contextual door selection from local calendar sources only.
**FRs covered:** FR44, FR45
**NFRs covered:** NFR15
**Dependencies:** Epic 4
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 13: Multi-Source Task Aggregation View
**Goal:** Unified cross-provider task pool with dedup detection and source attribution.
**FRs covered:** FR46, FR47, FR48
**Dependencies:** Epic 7, Epic 8+
**Timeline:** 2-3 weeks at 2-4 hrs/week

### Epic 14: LLM Task Decomposition & Agent Action Queue (Future)
**Goal:** LLM-powered task breakdown with git repo output for coding agent pickup.
**FRs covered:** FR35, FR36, FR37
**Dependencies:** Epic 3+
**Timeline:** 3-4 weeks at 2-4 hrs/week (spike-driven)

### Epic 15: Psychology Research & Validation (Parallel Track)
**Goal:** Evidence base for ThreeDoors design decisions.
**FRs covered:** FR34
**Dependencies:** None (parallel research track)
**Timeline:** Ongoing (2-4 hrs/week)

### Epic 16+: Additional Integrations & Advanced Features (Future)
**Goal:** Jira, Linear, Slack, cross-computer sync, voice interface, mobile apps.
**FRs covered:** Future FRs (not yet specified)
**Dependencies:** Epic 7+ (stable adapter SDK)
**Timeline:** 12+ months out

---

## Epic 1: Three Doors Technical Demo & Validation

Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists.

### Story 1.1: Project Setup & Basic Bubbletea App

As a developer,
I want a working Go project with Bubbletea framework,
So that I have a foundation for building the Three Doors TUI.

**Acceptance Criteria:**

**Given** the developer has Go 1.25.4+ installed
**When** they run `go mod init github.com/arcaven/ThreeDoors`
**Then** a Go module is initialized with proper module path
**And** Bubbletea 1.2.4+ and Lipgloss 1.0.0+ dependencies are added via go get

**Given** the project is built and run
**When** the user launches the application
**Then** a basic TUI renders "ThreeDoors - Technical Demo" header
**And** the application responds to 'q' keypress to quit
**And** the application responds to Ctrl+C to quit

**Given** the project structure
**When** a Makefile exists with `build`, `run`, and `clean` targets
**Then** `make build` compiles to `bin/threedoors`
**And** `make run` builds and runs the application
**And** `make clean` removes build artifacts

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 1.2: Display Three Doors from Task File

As a developer,
I want the application to read tasks from a text file and display three as "doors",
So that I can see the core Three Doors interface.

**Acceptance Criteria:**

**Given** the application starts
**When** `~/.threedoors/tasks.txt` exists with tasks (one per line)
**Then** three randomly selected tasks are displayed horizontally across the terminal
**And** doors dynamically adjust width based on terminal size
**And** no "Door X" labels are shown

**Given** the application starts
**When** `~/.threedoors/tasks.txt` does not exist
**Then** the file is created with 3-5 sample tasks
**And** three of those sample tasks are displayed as doors

**Given** three doors are displayed
**When** no door has been selected or doors were just re-rolled
**Then** no door is highlighted/selected (neutral state)

**Given** three doors are displayed
**When** the user presses 'a' or left arrow
**Then** the left door is selected/highlighted

**Given** three doors are displayed
**When** the user presses 'w' or up arrow
**Then** the center door is selected/highlighted

**Given** three doors are displayed
**When** the user presses 'd' or right arrow
**Then** the right door is selected/highlighted

**Given** three doors are displayed
**When** the user presses 's' or down arrow
**Then** a new set of three random tasks replaces the current doors
**And** no door is selected after re-roll

**Given** the application is running
**When** the user presses 'q' or Ctrl+C
**Then** the application exits cleanly

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 1.3: Door Selection & Task Status Management

As a user,
I want to select a door and update the task's status,
So that I can take action on tasks and track my progress.

**Acceptance Criteria:**

**Given** a door is selected (highlighted)
**When** the user presses Enter or the selection key again
**Then** a door opening animation plays (optional)
**And** the selected door shifts left and expands to fill the screen as a detail view
**And** the detail view shows: full task text, task metadata/status, and status action menu

**Given** the user is in the detail view
**When** the user presses 'c'
**Then** the task is marked as Complete
**And** the task is removed from the available pool
**And** the task is appended to `~/.threedoors/completed.txt` with timestamp
**And** the session completion count increments
**And** a "progress over perfection" celebration message is shown
**And** the view returns to three doors with a new set

**Given** the user is in the detail view
**When** the user presses 'b'
**Then** the task is marked as Blocked
**And** the user is prompted for an optional blocker note
**And** the task remains in the pool tagged with blocked status

**Given** the user is in the detail view
**When** the user presses 'i', 'e', 'f', 'p', or 'r'
**Then** the corresponding status is applied (In Progress, Expand, Fork, Procrastinate, Rework)
**And** the view returns to three doors with a new set

**Given** the user is in the detail view
**When** the user presses 'm'
**Then** a mood capture dialog appears with options: Focused, Tired, Stressed, Energized, Distracted, Calm, Other
**And** if "Other" is selected, a text input prompts for custom mood
**And** the mood entry is timestamped and recorded

**Given** the user is in the detail view
**When** the user presses Esc
**Then** the detail view closes and returns to the three doors view

**Given** the user is in the three doors view (no door selected)
**When** the user presses 'm'
**Then** the mood capture dialog opens without requiring door selection

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 1.4: Quick Search & Command Palette

As a user,
I want to search for specific tasks and execute commands via text input,
So that I can efficiently find and act on tasks without relying solely on the three doors.

**Acceptance Criteria:**

**Given** the user is in the three doors view
**When** the user presses '/'
**Then** a text input field appears at the bottom with placeholder "Search tasks... (or :command for commands)"

**Given** the search input is active
**When** the user types characters
**Then** matching tasks display from bottom-up (live substring matching)
**And** the list updates with each keystroke
**And** if no matches, "No tasks match '<text>'" is shown
**And** empty input shows no results

**Given** search results are displayed
**When** the user navigates with arrow keys, WASD, or HJKL (vi-style)
**Then** the selected result is highlighted
**And** pressing Enter opens the selected task in the detail view (same as Story 1.3)

**Given** a task detail was opened from search
**When** the user presses Esc from the detail view
**Then** the search view returns with search text preserved and previous selection highlighted

**Given** the search input is active (not in detail view)
**When** the user presses Esc or Ctrl+C
**Then** search mode closes and returns to three doors view

**Given** the search input is empty
**When** the user types ':' as the first character
**Then** the input switches to command mode with visual indicator

**Given** command mode is active
**When** the user types `:add <task text>` and presses Enter
**Then** a new task is added to tasks.txt

**Given** command mode is active
**When** the user types `:mood [mood]` and presses Enter
**Then** a mood is logged (prompts for selection if no mood parameter given)

**Given** command mode is active
**When** the user types `:stats` and presses Enter
**Then** session statistics are displayed (tasks completed, doors viewed, time in session)

**Given** command mode is active
**When** the user types `:help` and presses Enter
**Then** available commands and key bindings are displayed

**Given** command mode is active
**When** the user types `:quit` or `:exit` and presses Enter
**Then** the application exits cleanly

**Given** command mode is active
**When** the user types an invalid command and presses Enter
**Then** "Unknown command: '<command>'. Type :help for available commands." is shown

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 1.5: Session Metrics Tracking

As a developer validating the Three Doors concept,
I want objective session metrics collected automatically,
So that I can make a data-informed validation decision.

**Acceptance Criteria:**

**Given** the application starts
**When** a new session begins
**Then** a SessionTracker is initialized with UUID and current timestamp

**Given** a session is active
**When** a door is viewed, selected, refreshed, or a status change occurs
**Then** the corresponding event is recorded silently (no UI impact)
**And** door position selections are tracked (left=0, center=1, right=2)
**And** task bypass tracking records tasks shown but not selected before refresh
**And** mood entries are captured with timestamps

**Given** the session tracker is recording
**When** any recording operation occurs
**Then** it adds <1ms overhead per event
**And** no UI lag or stuttering is observable

**Given** the application exits cleanly
**When** the session is finalized
**Then** session metrics are written as a JSON line to `~/.threedoors/sessions.jsonl`
**And** each line is valid JSON parseable by `jq`
**And** the file is append-only

**Given** a file write failure occurs during metrics persistence
**When** the write fails
**Then** a warning is logged to stderr
**And** the application does not crash
**And** no error dialog is shown to the user

**Given** analysis scripts exist
**When** `scripts/analyze_sessions.sh` is run
**Then** session summaries and averages are displayed

**Given** analysis scripts exist
**When** `scripts/validation_decision.sh` is run
**Then** automated validation criteria are evaluated against collected data

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 1.6: Essential Polish

As a user,
I want the app to feel polished enough to use daily,
So that I enjoy the validation experience.

**Acceptance Criteria:**

**Given** the application is running
**When** any screen is rendered
**Then** Lipgloss styling is applied with distinct colors: doors have their own color, success messages are green, prompts are yellow/blue

**Given** the application starts or a task is completed
**When** the corresponding event occurs
**Then** a "Progress over perfection" message is displayed (startup greeting or post-completion)

**Given** the user interacts with the application
**When** any action is performed
**Then** the response feels responsive and smooth with no noticeable lag

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 2: Foundation & Apple Notes Integration

Replace/augment text file backend with Apple Notes integration, enabling bidirectional sync so tasks edited on iPhone appear in the terminal and vice versa.

### Story 2.1: Architecture Refactoring to Adapter Pattern

As a developer,
I want the codebase refactored to use a TaskProvider interface,
So that multiple backends (text file, Apple Notes) can be swapped without changing the core logic.

**Acceptance Criteria:**

**Given** the existing codebase with direct file I/O
**When** the refactoring is complete
**Then** a `TaskProvider` interface exists with methods: `LoadTasks()`, `SaveTask()`, `DeleteTask()`, `MarkComplete()`
**And** a `TextFileProvider` implements this interface (wrapping existing file I/O logic)
**And** the MainModel and domain layer depend only on the `TaskProvider` interface, not concrete implementations
**And** all existing functionality works identically through the new interface

**Given** the adapter pattern is in place
**When** unit tests are run
**Then** core domain logic can be tested with a mock `TaskProvider`
**And** `TextFileProvider` has integration tests covering read, write, and error scenarios
**And** test coverage for core domain logic reaches 70%+

**Given** the refactored architecture
**When** the build is run
**Then** CI/CD pipeline via GitHub Actions runs tests on every commit
**And** `make test` target is added to the Makefile
**And** `make lint` target runs golangci-lint

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 2.2: Apple Notes Integration Spike

As a developer,
I want to evaluate Apple Notes integration approaches,
So that I can choose the best method for bidirectional sync.

**Acceptance Criteria:**

**Given** the spike begins
**When** evaluating integration options
**Then** at least 3 approaches are tested: MCP server (mcp-apple-notes), direct SQLite read, AppleScript bridge
**And** each approach is evaluated for: read capability, write capability, reliability, complexity

**Given** the spike is complete
**When** results are documented
**Then** a spike report exists documenting: chosen approach, pros/cons, risks, estimated effort
**And** a proof-of-concept demonstrates reading at least one note from Apple Notes
**And** the chosen approach is validated for both read and write operations

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 2.3: Read Tasks from Apple Notes

As a user,
I want the Three Doors app to read tasks from my Apple Notes,
So that I can manage my existing Apple Notes tasks from the terminal.

**Acceptance Criteria:**

**Given** Apple Notes contains a designated task note
**When** the application starts with Apple Notes provider configured
**Then** tasks are retrieved from Apple Notes and displayed in the Three Doors interface
**And** tasks appear within <2 seconds of startup

**Given** Apple Notes is not accessible (app closed, permissions denied)
**When** the application starts
**Then** the system falls back gracefully to text file backend
**And** a clear message informs the user about the fallback
**And** core Three Doors functionality remains fully operational

**Given** the Apple Notes provider is active
**When** the user navigates the Three Doors interface
**Then** the experience is identical to the text file backend (same keys, same views, same flows)

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Verify AppleScript injection safety: Ensure all note titles and task text passed to osascript are properly escaped
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 2.4: Write Task Updates to Apple Notes

As a user,
I want task status changes in Three Doors to sync back to Apple Notes,
So that my iPhone shows the latest task state.

**Acceptance Criteria:**

**Given** a task from Apple Notes is displayed in Three Doors
**When** the user marks the task as complete, blocked, in-progress, or any other status
**Then** the status change is written back to Apple Notes within 5 seconds
**And** the change is visible when viewing the note on iPhone

**Given** a write operation to Apple Notes fails
**When** the error occurs
**Then** the change is cached locally for retry
**And** a non-intrusive warning is shown to the user
**And** the local state reflects the intended change

**Given** the user completes a task in Three Doors
**When** the completion is synced
**Then** the task is marked as complete in Apple Notes (not deleted)
**And** the task appears in the local completed.txt log with timestamp

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Verify AppleScript injection safety: Ensure all note titles and task text passed to osascript are properly escaped
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 2.5: Bidirectional Sync

As a user,
I want tasks edited on my iPhone to appear updated in Three Doors,
So that I can manage tasks from either device seamlessly.

**Acceptance Criteria:**

**Given** a task was modified in Apple Notes on iPhone
**When** the user opens or refreshes Three Doors
**Then** the modified task appears with the latest content
**And** no duplicate tasks are created

**Given** a new task was added in Apple Notes on iPhone
**When** the user opens or refreshes Three Doors
**Then** the new task appears in the available pool

**Given** a task was deleted in Apple Notes on iPhone
**When** the user opens or refreshes Three Doors
**Then** the task is removed from the available pool
**And** no error is displayed

**Given** conflicting changes (edited in both Apple Notes and Three Doors)
**When** sync occurs
**Then** the most recent change wins (last-write-wins)
**And** the user is informed if their local change was overridden

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 2.6: Health Check Command

As a user,
I want a health check command to verify Apple Notes connectivity,
So that I can diagnose sync issues.

**Acceptance Criteria:**

**Given** the application is running
**When** the user types `:health` in the command palette
**Then** the system checks: Apple Notes accessibility, database read/write, sync status, last successful sync timestamp
**And** displays results with green (OK) / red (FAIL) indicators

**Given** the health check detects issues
**When** results are displayed
**Then** actionable suggestions are shown (e.g., "Grant Full Disk Access in System Settings")

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 3: Enhanced Task Capture & Interaction

Enrich the task management experience with task capture, values/goals, door feedback, and progress tracking.

### Story 3.1: Quick Add Mode

As a user,
I want to quickly add tasks with minimal interaction,
So that I can capture ideas without breaking flow.

**Acceptance Criteria:**

**Given** the user is in any view
**When** the user types `:add <task text>` in the command palette
**Then** a new task is created with the given text
**And** the task is immediately available in the Three Doors pool
**And** the task is persisted to the active backend
**And** a brief confirmation is shown ("Task added")

**Given** the user types `:add` without text
**When** Enter is pressed
**Then** an inline text input prompts for task text
**And** pressing Enter again creates the task
**And** pressing Esc cancels

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.2: Extended Task Capture with Context

As a user,
I want to capture not just what a task is but why it matters,
So that I can make better decisions about task priority and alignment with goals.

**Acceptance Criteria:**

**Given** the user wants to add a task with context
**When** the user types `:add-ctx` or `:add --why` in the command palette
**Then** a multi-step capture flow begins:
**And** Step 1: "What's the task?" - single line text input
**And** Step 2: "Why does this matter?" - optional context input (Enter to skip)
**And** the task is created with both text and context stored

**Given** a task has context/why information
**When** the task appears in the detail view
**Then** the "why" context is displayed below the task text

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.3: Values & Goals Setup and Display

As a user,
I want to define and see my values and goals during task sessions,
So that I stay aligned with what matters most.

**Acceptance Criteria:**

**Given** the user has no values/goals configured
**When** the user types `:goals` in the command palette
**Then** a setup flow guides them through entering 1-5 values or goals
**And** values are persisted to `~/.threedoors/config.yaml` (or equivalent)

**Given** the user has values/goals configured
**When** any screen is displayed
**Then** values/goals appear as a subtle header or footer (not intrusive)
**And** they remain visible across door selection, detail view, and search

**Given** the user wants to edit values/goals
**When** the user types `:goals edit` in the command palette
**Then** the existing values are displayed for editing (add, remove, reorder)

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.4: Door Feedback Options

As a user,
I want to provide feedback on why a door isn't suitable,
So that the system can learn my preferences over time.

**Acceptance Criteria:**

**Given** a door is selected in the three doors view
**When** the user presses 'n' (not now) or opens the feedback menu
**Then** options are displayed: "Blocked", "Not now", "Needs breakdown", "Other comment"
**And** the feedback is recorded with the task ID and timestamp

**Given** the user selects "Needs breakdown"
**When** the feedback is submitted
**Then** the task is flagged for future breakdown
**And** the doors refresh with a new set

**Given** the user selects "Other comment"
**When** the feedback option is chosen
**Then** a text input appears for freeform feedback
**And** the comment is stored with the task

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.5: Daily Completion Tracking & Comparison

As a user,
I want to see how many tasks I completed today compared to yesterday,
So that I can feel motivated by progress.

**Acceptance Criteria:**

**Given** the user has been using the app across multiple days
**When** any task is completed
**Then** the session display shows: "Completed today: X (yesterday: Y)"
**And** if today > yesterday, a positive message is shown
**And** if today = 0, no comparison is shown (avoids discouragement)

**Given** the user types `:stats` in the command palette
**When** daily stats are displayed
**Then** the output includes: tasks completed today, tasks completed yesterday, doors viewed today, current streak (consecutive days with completions)

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.6: Session Improvement Prompt

As a user,
I want to be prompted for one improvement idea per session,
So that I continuously refine my task management approach.

**Acceptance Criteria:**

**Given** the user has been in a session for at least 5 minutes or completed 1+ tasks
**When** the user exits the application (q or :quit)
**Then** a prompt appears: "What's one thing you could improve about this list/task/goal right now?"
**And** the user can type a response and press Enter to save
**And** pressing Esc skips the prompt

**Given** the user provides an improvement suggestion
**When** the response is saved
**Then** it is appended to `~/.threedoors/improvements.txt` with timestamp and session ID

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 3.7: Enhanced Navigation & Messaging

As a user,
I want richer choose-your-own-adventure navigation and enhanced "progress over perfection" messaging,
So that the app feels like a supportive partner rather than a demanding taskmaster.

**Acceptance Criteria:**

**Given** the user performs various actions (complete task, open door, refresh, add task)
**When** the action completes
**Then** contextual, encouraging messages are shown (varying, not always the same)
**And** messages embody "progress over perfection" philosophy
**And** messages celebrate any action as progress

**Given** the user is at a decision point (e.g., after completing a task)
**When** options are presented
**Then** 3-5 contextual next steps are shown (not just "return to doors")
**And** options adapt based on state (e.g., "add another task", "review blocked tasks", "check stats", "take a break")

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 4: Learning & Intelligent Door Selection

Use historical session metrics to analyze patterns and adapt door selection to user context.

### Story 4.1: Task Categorization

As a user,
I want tasks to be categorized by type, effort, and context,
So that the door selection can present diverse options.

**Acceptance Criteria:**

**Given** a task exists in the system
**When** the task is created or edited
**Then** optional categorization fields are available: type (creative, administrative, technical, physical), effort (quick-win, medium, deep-work), context (home, work, anywhere)

**Given** tasks have categories
**When** three doors are selected
**Then** the algorithm prefers diversity: different types, different effort levels when possible
**And** if insufficient variety exists, random selection is used as fallback

**Given** categories are not set on a task
**When** the task appears in doors
**Then** it is treated as "uncategorized" and can appear in any door

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 4.2: Pattern Recognition & Avoidance Detection

As a user,
I want the system to recognize my selection patterns,
So that I get insights into my work habits.

**Acceptance Criteria:**

**Given** sufficient session data exists (10+ sessions)
**When** pattern analysis runs
**Then** the system identifies: tasks/types consistently selected vs bypassed, time-of-day preferences, avoidance patterns (tasks shown 5+ times but never selected)

**Given** avoidance is detected for a task
**When** the task appears in doors
**Then** a subtle indicator shows it has been frequently bypassed (not judgmental)

**Given** the user types `:insights` in the command palette
**When** pattern data is available
**Then** insights are displayed: "You tend to pick quick-wins in the morning", "Task X has been shown 8 times without selection", "Your most productive days are Tuesdays"

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 4.3: Mood Correlation Analysis

As a user,
I want the system to correlate my mood with task selection,
So that I understand how emotions affect my productivity.

**Acceptance Criteria:**

**Given** mood data and task selection data from 10+ sessions
**When** the user types `:insights mood` in the command palette
**Then** correlations are displayed: "When stressed, you avoid complex tasks", "Your highest completion rate is when feeling focused", "You tend to log 'tired' on Fridays"

**Given** the user logs a mood
**When** doors are being selected
**Then** the current mood is factored into door selection as a soft preference (not hard filter)
**And** if mood is "tired", simpler/quicker tasks are slightly preferred
**And** if mood is "focused", deeper work tasks are slightly preferred

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 4.4: Adaptive Door Selection Algorithm

As a user,
I want door selection to adapt to my patterns over time,
So that I'm presented with tasks I'm more likely to engage with.

**Acceptance Criteria:**

**Given** historical data from pattern recognition and mood correlation
**When** three doors are selected
**Then** the algorithm balances: user's current mood preference, task diversity, avoidance pattern awareness, time-of-day patterns
**And** at least one "stretch" task (something the user tends to avoid) is included when possible

**Given** the adaptive algorithm is active
**When** the user refreshes doors
**Then** the new set reflects different aspects of the algorithm's recommendations
**And** no task appears in consecutive door sets (ring buffer still applies)

**Given** the adaptive algorithm is active
**When** persistent avoidance of certain task types is detected across 3+ sessions
**Then** a gentle prompt suggests goal re-evaluation: "You've been skipping [category] tasks. Want to review if they still align with your goals?"

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 4.5: User Insights Dashboard

As a user,
I want a comprehensive insights view of my work patterns,
So that I can make informed decisions about my productivity.

**Acceptance Criteria:**

**Given** sufficient historical data
**When** the user types `:dashboard` or `:insights all` in the command palette
**Then** a summary view displays: completion trends (daily/weekly), mood-productivity correlations, door position preferences, most/least engaged task categories, streak information, "better than yesterday" multi-dimensional tracking

**Given** the dashboard is displayed
**When** the user presses Esc
**Then** the view returns to the previous context

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 5: Data Layer & Enrichment (Optional)

Add enrichment storage layer for cross-system metadata and richer task relationships.

### Story 5.1: SQLite Enrichment Database Setup

As a developer,
I want a SQLite database for storing enrichment metadata,
So that cross-system task relationships and learning patterns can be persisted efficiently.

**Acceptance Criteria:**

**Given** the enrichment layer is being set up
**When** the database is initialized
**Then** a SQLite database is created at `~/.threedoors/enrichment.db`
**And** schema includes tables for: task_metadata (categories, enrichment tags), cross_references (links between tasks across systems), learning_patterns (algorithm weights, pattern data), feedback_history (door feedback, mood correlations)

**Given** the database exists
**When** the application starts
**Then** the enrichment layer loads in parallel with the task provider
**And** startup time increases by no more than 100ms

**Given** a database migration is needed
**When** the schema version changes
**Then** automatic migration runs on startup
**And** existing data is preserved

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 5.2: Cross-Reference Tracking

As a user,
I want tasks from different sources to be linked,
So that I can see relationships across systems.

**Acceptance Criteria:**

**Given** tasks exist from multiple backends (text file + Apple Notes)
**When** the user identifies two related tasks
**Then** a cross-reference can be created between them via `:link <task1> <task2>`
**And** linked tasks show a reference indicator in the detail view

**Given** a task with cross-references is viewed
**When** the detail view is displayed
**Then** linked tasks are listed with their source system and status

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 5.3: Data Migration & Backup

As a user,
I want data migration tools and backup capability,
So that I don't lose enrichment data during upgrades or system changes.

**Acceptance Criteria:**

**Given** enrichment data exists
**When** the user types `:backup` in the command palette
**Then** a timestamped backup is created at `~/.threedoors/backups/enrichment-<timestamp>.db`

**Given** a backup exists
**When** the user types `:restore <backup-file>`
**Then** the enrichment database is replaced with the backup
**And** a confirmation prompt prevents accidental restores

**Given** a new version changes the schema
**When** the application starts
**Then** automatic migration preserves all existing data
**And** a migration log records what changed

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 7: Plugin/Adapter SDK & Registry

Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide. Unblocks all future integrations.

### Story 7.1: Adapter Registry & Runtime Discovery

As a developer building integrations,
I want a formal adapter registry that discovers and loads task providers at runtime,
So that new integrations can be added without modifying core application code.

**Acceptance Criteria:**

**Given** the adapter registry is initialized
**When** the application starts
**Then** it discovers all registered TaskProvider implementations
**And** loads them based on configuration

**Given** an adapter is registered
**When** it fails to initialize
**Then** the system logs a warning and continues with other adapters
**And** graceful degradation is maintained

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 7.2: Config-Driven Provider Selection

As a user with multiple task sources,
I want to configure active backends via `~/.threedoors/config.yaml`,
So that I can choose which task providers are active without code changes.

**Acceptance Criteria:**

**Given** a config.yaml exists with provider configuration
**When** the application starts
**Then** only configured providers are loaded and activated
**And** provider-specific settings are passed to each adapter

**Given** no config.yaml exists
**When** the application starts
**Then** it falls back to the default text file provider
**And** a sample config.yaml is generated for reference

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 7.3: Adapter Developer Guide

As an integration developer,
I want a clear guide and interface specification for building adapters,
So that I can create new task provider integrations with confidence.

**Acceptance Criteria:**

**Given** the developer guide exists
**When** a developer reads it
**Then** it covers: TaskProvider interface spec, registration process, config schema, testing requirements, and example adapter implementation

**Given** the contract test suite exists
**When** a new adapter is developed
**Then** it can validate compliance by running the contract test suite against its implementation

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 8: Obsidian Integration (P0 - #2 Integration)

Add Obsidian vault as second task storage backend after Apple Notes. Local-first Markdown integration with bidirectional sync.

### Story 8.1: Obsidian Vault Reader/Writer Adapter

As a user who manages tasks in Obsidian,
I want ThreeDoors to read and write tasks from my Obsidian vault,
So that I can use Three Doors with my existing Obsidian workflow.

**Acceptance Criteria:**

**Given** an Obsidian vault path is configured
**When** the application starts
**Then** it reads Markdown files from the configured vault folder
**And** parses task items (checkbox syntax: `- [ ]`, `- [x]`) from the files

**Given** a task is completed in ThreeDoors
**When** the status change is persisted
**Then** the corresponding Markdown file is updated with the new checkbox state
**And** file writes use atomic operations to prevent corruption

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 8.2: Obsidian Bidirectional Sync

As an Obsidian user,
I want changes made in Obsidian to be reflected in ThreeDoors and vice versa,
So that my tasks stay in sync regardless of where I edit them.

**Acceptance Criteria:**

**Given** a vault file is modified externally (in Obsidian)
**When** ThreeDoors refreshes or polls for changes
**Then** the updated tasks are reflected in the Three Doors interface
**And** no data is lost from concurrent edits

**Given** ThreeDoors modifies a task
**When** the change is written to the vault
**Then** Obsidian reflects the change on its next file reload

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 8.3: Obsidian Vault Configuration

As a user,
I want to configure my Obsidian vault path, target folder, and file naming via config.yaml,
So that ThreeDoors integrates with my specific vault structure.

**Acceptance Criteria:**

**Given** config.yaml contains Obsidian provider settings
**When** the application starts
**Then** it uses the configured vault path, folder, and naming conventions
**And** validates the vault path exists and is accessible

**Given** an invalid vault path is configured
**When** the application starts
**Then** a clear error message indicates the vault path issue
**And** the application falls back to other configured providers

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 8.4: Obsidian Daily Note Integration

As an Obsidian user who uses daily notes,
I want ThreeDoors to read/write tasks from my daily note files,
So that tasks captured in daily notes appear in Three Doors and vice versa.

**Acceptance Criteria:**

**Given** daily note integration is enabled in config
**When** the application loads tasks
**Then** it also reads tasks from today's daily note file
**And** uses the configured daily note path pattern (e.g., `YYYY-MM-DD.md`)

**Given** a task is added via ThreeDoors quick add
**When** daily note mode is active
**Then** the task is appended to today's daily note under a configurable heading

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 9: Testing Strategy & Quality Gates

Establish comprehensive testing infrastructure with integration, contract, performance, and E2E tests.

### Story 9.1: Apple Notes Integration E2E Tests

As a developer,
I want end-to-end tests for the Apple Notes integration workflow,
So that regressions in the sync pipeline are caught automatically.

**Acceptance Criteria:**

**Given** the test suite runs
**When** Apple Notes integration tests execute
**Then** they validate: note creation, task read, task update, bidirectional sync, and error handling
**And** tests use mock/stub AppleScript responses for CI compatibility

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 9.2: Contract Tests for Adapter Compliance

As an adapter developer,
I want a contract test suite that validates any TaskProvider implementation,
So that all adapters behave consistently regardless of backend.

**Acceptance Criteria:**

**Given** a TaskProvider implementation exists
**When** the contract test suite runs against it
**Then** it validates: CRUD operations, error handling, concurrent access safety, and interface compliance
**And** the test suite is reusable across all adapter implementations

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 9.3: Performance Benchmarks

As a developer,
I want automated performance benchmarks validating the <100ms NFR,
So that performance regressions are caught before they reach users.

**Acceptance Criteria:**

**Given** the benchmark suite runs
**When** adapter operations (read, write, sync) are benchmarked
**Then** results are compared against the <100ms threshold (NFR13)
**And** benchmark results are reported in CI output
**And** regressions beyond threshold fail the CI pipeline

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 9.4: Functional E2E Tests

As a developer,
I want functional end-to-end tests covering full user workflows,
So that the complete user experience is validated automatically.

**Acceptance Criteria:**

**Given** the E2E test suite runs
**When** full user workflows are exercised (launch → select door → manage task → exit)
**Then** each workflow completes successfully
**And** session metrics are correctly generated

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 9.5: CI Coverage Gates

As a team,
I want CI coverage gates that prevent test coverage from regressing,
So that code quality is maintained as the codebase grows.

**Acceptance Criteria:**

**Given** a PR is submitted
**When** CI runs the test suite
**Then** coverage is measured and compared against the established threshold
**And** PRs that reduce coverage below the threshold are blocked
**And** coverage reports are generated and accessible

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 10: First-Run Onboarding Experience

Guided welcome flow for new users to set up values/goals, understand Three Doors, learn key bindings, and optionally import existing tasks.

### Story 10.1: Welcome Flow & Three Doors Explanation

As a new user,
I want a guided welcome experience on first launch,
So that I understand the Three Doors concept and feel confident using the tool.

**Acceptance Criteria:**

**Given** the application launches for the first time (no `~/.threedoors/` directory exists)
**When** the welcome flow starts
**Then** it explains the Three Doors concept (choice architecture, why 3 options)
**And** walks through key bindings with interactive examples
**And** the user can skip the walkthrough at any time

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 10.2: Values/Goals Setup & Task Import

As a new user,
I want to set up my values/goals and import existing tasks during onboarding,
So that the tool is immediately useful with my real data.

**Acceptance Criteria:**

**Given** the welcome flow reaches the setup step
**When** the user is prompted for values/goals
**Then** they can enter values and goals that will be displayed during sessions (per FR6)

**Given** the import step is reached
**When** existing task sources are detected (text files, other tools)
**Then** the user can select sources to import from
**And** imported tasks populate the task pool

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 11: Sync Observability & Offline-First

Robust offline-first operation with local change queue, sync status visibility, conflict visualization, and sync debugging.

### Story 11.1: Offline-First Local Change Queue

As a user working without connectivity,
I want all changes queued locally and replayed when sync targets are available,
So that I never lose work due to connectivity issues.

**Acceptance Criteria:**

**Given** a sync target is unavailable
**When** the user makes changes (complete, add, update tasks)
**Then** changes are queued in a local write-ahead log
**And** core functionality remains fully operational

**Given** a sync target becomes available
**When** the queue is replayed
**Then** all queued changes are applied in order
**And** failures are retried with exponential backoff

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 11.2: Sync Status Indicator

As a user,
I want to see the sync status of each provider in the TUI,
So that I know whether my changes are synchronized.

**Acceptance Criteria:**

**Given** the TUI is displayed
**When** sync providers are configured
**Then** a status indicator shows per-provider sync state (synced, syncing, pending, error)
**And** the indicator updates in real-time as sync operations complete

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 11.3: Conflict Visualization & Sync Log

As a user encountering sync conflicts,
I want to see what conflicted and review a sync log for debugging,
So that I can resolve issues and trust the sync system.

**Acceptance Criteria:**

**Given** a sync conflict is detected
**When** the conflict visualization is shown
**Then** it displays both local and remote versions of the conflicting item
**And** provides resolution options (keep local, keep remote, keep both)

**Given** sync operations occur
**When** the user types `:synclog` in the command palette
**Then** a chronological sync log is displayed with timestamps, operations, and outcomes

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 12: Calendar Awareness (Local-First, No OAuth)

Time-contextual door selection by reading local calendar sources. No OAuth, no cloud APIs.

### Story 12.1: Local Calendar Source Reader

As a user,
I want ThreeDoors to read my local calendar to understand my available time,
So that doors can suggest tasks appropriate for my current time context.

**Acceptance Criteria:**

**Given** calendar integration is enabled in config
**When** the application loads
**Then** it reads events from macOS Calendar.app via AppleScript
**And/or** parses .ics files from configured paths
**And/or** reads CalDAV cache from local filesystem
**And** no OAuth or cloud API calls are made

**Given** calendar reading fails
**When** the application continues
**Then** it falls back to non-time-contextual door selection
**And** logs a warning about calendar unavailability

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 12.2: Time-Contextual Door Selection

As a user with calendar awareness enabled,
I want doors to suggest tasks that fit my available time blocks,
So that I'm not shown a 2-hour task when I have a meeting in 15 minutes.

**Acceptance Criteria:**

**Given** calendar data indicates a short time block (< 30 min)
**When** doors are generated
**Then** the algorithm prefers quick tasks over long ones

**Given** calendar data indicates a large open block
**When** doors are generated
**Then** the algorithm includes tasks of any estimated duration

**Given** no calendar data is available
**When** doors are generated
**Then** the standard selection algorithm is used (no degradation)

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 13: Multi-Source Task Aggregation View

Unified cross-provider task pool with dedup detection and source attribution.

### Story 13.1: Cross-Provider Task Pool Aggregation

As a user with multiple task sources,
I want all tasks aggregated into a single pool for Three Doors selection,
So that I see tasks from all my sources without switching between them.

**Acceptance Criteria:**

**Given** multiple providers are configured and active
**When** the task pool is loaded
**Then** tasks from all providers are merged into a single pool
**And** the Three Doors selection draws from the unified pool

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 13.2: Duplicate Detection & Source Attribution

As a user with overlapping task sources,
I want duplicates flagged and each task's source clearly shown,
So that I don't work on the same task twice and know where each task lives.

**Acceptance Criteria:**

**Given** tasks are aggregated from multiple providers
**When** potential duplicates are detected (fuzzy text matching)
**Then** they are flagged with a visual indicator
**And** the user can merge or dismiss duplicate flags

**Given** a task is displayed in any view (doors, search, detail)
**When** multiple providers are active
**Then** the task's source provider is shown as a badge or label

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 14: LLM Task Decomposition & Agent Action Queue (Future)

LLM-powered task breakdown with git repo output for coding agent pickup. Spike-first approach.

### Story 14.1: LLM Task Decomposition Spike

As a developer,
I want to spike on LLM-powered task decomposition,
So that we understand the feasibility, prompt engineering, and output quality before committing to full implementation.

**Acceptance Criteria:**

**Given** a user selects a task for decomposition
**When** the LLM spike is triggered
**Then** it generates BMAD-style stories/specs from the task description
**And** outputs follow a defined schema

**Spike deliverables:**
- Prompt engineering experiments with multiple LLM providers
- Output schema definition for stories/specs
- Git automation proof-of-concept (writing to repo structure)
- Agent handoff protocol draft (how Claude Code / multiclaude picks up work)
- Local vs cloud LLM comparison
- Recommendation document for full implementation

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

### Story 14.2: Agent Action Queue Integration

As a developer using ThreeDoors with coding agents,
I want decomposed tasks output to a git repo structure that coding agents can pick up,
So that task decomposition flows directly into automated implementation.

**Acceptance Criteria:**

**Given** the LLM generates stories/specs from a task
**When** the output is committed to the repo
**Then** it follows the BMAD story file structure
**And** multiclaude / Claude Code can discover and pick up the stories
**And** the task in ThreeDoors is updated with a link to the generated work

#### Pre-PR Submission Checklist

- [ ] Rebase onto latest main: `git fetch upstream main && git rebase upstream/main`
- [ ] Run gofumpt: `gofumpt -l .` — verify no output
- [ ] Run golangci-lint: `golangci-lint run ./...` — verify 0 issues
- [ ] Run all tests: `go test ./... -count=1` — verify 0 failures
- [ ] Check for dead code: `go vet ./...`
- [ ] Verify no out-of-scope files: Review `git diff --stat`
- [ ] Single clean commit preferred: Squash fix-ups before pushing

---

## Epic 15: Psychology Research & Validation (Parallel Track)

Evidence base for ThreeDoors design decisions through literature review and validation studies.

### Story 15.1: Choice Architecture Literature Review

As the product team,
I want a literature review documenting the evidence for the Three Doors choice architecture,
So that design decisions are grounded in behavioral science.

**Acceptance Criteria:**

**Given** the literature review is complete
**When** it is documented in `docs/research/choice-architecture.md`
**Then** it covers: why 3 options (choice overload research), paradox of choice, decision fatigue, and comparable systems
**And** includes citations and practical implications for ThreeDoors design

### Story 15.2: Mood-Task Correlation & Procrastination Research

As the product team,
I want research on mood-task correlation models and procrastination interventions,
So that Epic 4's learning algorithm is informed by evidence.

**Acceptance Criteria:**

**Given** the research is complete
**When** it is documented in `docs/research/mood-correlation.md` and `docs/research/procrastination.md`
**Then** it covers: mood-productivity correlations, procrastination intervention mechanisms, "progress over perfection" as motivational framework
**And** provides actionable recommendations for Epic 4 implementation

---

## Epic 16+: Additional Integrations & Advanced Features (Future)

*Stories to be defined when Phase 4 planning begins. Potential integrations include Jira, Linear, Slack, cross-computer sync, voice interface, and mobile apps. Each integration will follow the adapter SDK established in Epic 7.*
