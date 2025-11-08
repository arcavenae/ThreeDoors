# ThreeDoors Product Requirements Document (PRD)

**Document Version:** 1.1 (Technical Demo & Validation Phase)
**Last Updated:** 2025-11-07
**Project Repository:** github.com/arcaven/ThreeDoors.git

---

## Goals and Background Context

### Goals

**Technical Demo & Validation Phase (Pre-MVP):**
- Validate the Three Doors UX concept in 1 week (4-8 hours of development)
- Prove the core hypothesis: "Presenting three diverse tasks is better than presenting a list"
- Build working TUI with Bubbletea to demonstrate feasibility
- Use simple local text file for rapid task population and testing
- Gather real usage feedback before investing in complex integrations

**Full MVP Goals (Post-Validation):**
- Master BMAD methodology through authentic, real-world application
- Create a todo app that reduces friction and actually helps with organization
- Build a personal achievement partner that works with human psychology, not against it
- Enable seamless cross-context navigation across multiple platforms and tools
- Capture the full story (what AND why) to improve stakeholder communication
- Achieve measurably better personal organization than current scattered approach
- Demonstrate progress-over-perfection philosophy in both product design and development process

### Background Context

Traditional todo apps work well for already-organized people, but they're fundamentally rudimentary tools that haven't evolved alongside modern technology capabilities. While they help those who are naturally organized stay organized, they offer little support for adapting to the dynamic reality of modern life—where the same person occupies multiple roles (employee, parent, partner, learner), experiences varying moods and energy states, and faces constantly shifting priorities.

ThreeDoors recognizes that as technology has advanced, we can offer substantially more support. We can organize our organization tools themselves, bringing together tasks scattered across multiple systems. More importantly, we can adapt technology support dynamically: responding to the user's current context, role, mood, and circumstances, re-routing based on changing conditions and priorities. This PRD defines the MVP: a CLI/TUI application with Apple Notes integration that begins this journey, embodying "progress over perfection" philosophy while serving as a practical demonstration of the BMAD methodology.

### Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| 2025-11-07 | 1.0 | Initial PRD creation from project brief | John (PM Agent) |
| 2025-11-07 | 1.1 | Pivoted to Technical Demo & Validation approach (Option C): Simplified to text file storage, 1-week validation timeline, deferred Apple Notes and learning features to post-validation phases | John (PM Agent) |

---

## Requirements

### Technical Demo & Validation Phase Requirements

**Core Requirements (Week 1 - Build & Validate):**

**TD1:** The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)

**TD2:** The system shall read tasks from a simple local text file (e.g., `~/.threedoors/tasks.txt`)

**TD3:** The system shall display the Three Doors interface showing three tasks selected from the text file

**TD4:** The system shall allow users to select a door (press 1, 2, or 3) to start working on that task

**TD5:** The system shall allow users to mark the selected task as complete

**TD6:** The system shall track and display task completion count for the current session

**TD7:** The system shall provide a refresh mechanism (press R) to generate a new set of three doors

**TD8:** The system shall embed "progress over perfection" messaging in the interface

**TD9:** The system shall write completed tasks to a separate file (e.g., `~/.threedoors/completed.txt`) with timestamp

**Success Criteria for Phase:**
- Built and running within 4-8 hours of development time
- Developer uses it daily for 1 week to validate UX concept
- Three Doors selection feels meaningfully different from a simple list
- Decisions made on whether to proceed to Full MVP based on real usage

---

### Full MVP Requirements (Post-Validation - Deferred)

**Phase 2 - Apple Notes Integration:**

**FR2:** The system shall integrate with Apple Notes as the primary task storage backend, enabling bidirectional sync

**FR4:** The system shall retrieve and display tasks from Apple Notes within the application interface

**FR5:** The system shall allow users to mark tasks as complete, updating both the application state and Apple Notes

**FR12:** The system shall allow updates to tasks from either the application or directly in Apple Notes on iPhone, with changes reflected bidirectionally

**FR15:** The system shall provide a health check command to verify Apple Notes connectivity and database integrity

**Phase 3 - Enhanced Interaction & Learning:**

**FR3:** The system shall allow users to capture new tasks with optional context (what and why) through the CLI/TUI

**FR6:** The system shall display user-defined values and goals persistently throughout task work sessions

**FR7:** The system shall provide a "choose-your-own-adventure" interactive navigation flow that presents options rather than demands

**FR8:** The system shall track daily task completion count and display comparison to previous day's count

**FR9:** The system shall prompt the user once per session with: "What's one thing you could improve about this list/task/goal right now?"

**FR10:** The system shall embed "progress over perfection" messaging throughout interaction patterns and interface copy (enhanced beyond Tech Demo)

**FR16:** The system shall support a "quick add" mode for capturing tasks with minimal interaction

**FR18:** The system shall allow users to provide feedback on why a specific door isn't suitable with options: Blocked, Not now, Needs breakdown, or Other comment

**FR19:** The system shall capture and store blocker information when a task is marked as blocked

**FR20:** The system shall use door selection and feedback patterns to inform future door selection (learning which task types suit which contexts)

**FR21:** The system shall categorize tasks by type, effort level, and context to enable diverse door selection

**Phase 4 - Data Layer & Enrichment:**

**FR11:** The system shall maintain a local enrichment layer (SQLite and/or vector database) for metadata, cross-references, and relationships that cannot be stored in source systems

### Non-Functional Requirements

**Technical Demo Phase:**

**TD-NFR1:** The system shall be built in Go 1.25.4+ using idiomatic patterns and gofumpt formatting standards

**TD-NFR2:** The system shall use the Bubbletea/Charm Bracelet ecosystem for TUI implementation

**TD-NFR3:** The system shall operate on macOS as the primary target platform

**TD-NFR4:** The system shall store all data in local text files (`~/.threedoors/` directory) with no external services or telemetry

**TD-NFR5:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <100ms for typical operations given simple file I/O)

**TD-NFR6:** The system shall use Make as the build system with simple targets: `build`, `run`, `clean`

**TD-NFR7:** The system shall gracefully handle missing or corrupted task files by creating defaults

---

**Full MVP Phase (Post-Validation - Deferred):**

**NFR1:** The system shall maintain idiomatic Go patterns and gofumpt formatting standards

**NFR2:** The system shall continue using Bubbletea/Charm Bracelet ecosystem

**NFR3:** The system shall operate on macOS as primary platform

**NFR4:** The system shall store all user data locally or in user's iCloud (via Apple Notes), with no external telemetry or tracking

**NFR5:** The system shall store application state and enrichment data locally (cross-computer sync deferred to post-MVP)

**NFR6:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <500ms for typical operations)

**NFR7:** The system shall provide graceful degradation when Apple Notes integration is unavailable, maintaining core functionality

**NFR8:** The system shall implement secure credential storage using OS keychain for any API keys or authentication tokens

**NFR9:** The system shall never log sensitive user data or credentials

**NFR10:** The system shall use Make as the build system

**NFR11:** The system shall maintain clear architectural separation between core engine, TUI layer, integration adapters, and enrichment storage

**NFR12:** The system shall maintain data integrity even when Apple Notes is modified externally while app is running

---

## User Interface Design Goals

### Overall UX Vision

ThreeDoors presents as a conversational partner rather than a demanding taskmaster. The central interface metaphor is literal: **three doors, three tasks, three different on-ramps to action**. At each session start, the user is presented with three carefully selected tasks that are very different from each other—different types of activities, different effort levels, different contexts—but all represent good starting points based on priorities. This design serves dual purposes: it gets the user in the habit of doing *something* (reducing inertia), and it teaches the tool about the user's current state by observing which types of tasks they gravitate toward or avoid.

The interface should feel like opening a dialogue, not confronting a backlog. Users are greeted with options that respect their current capacity—whether focused, overwhelmed, or stuck—and celebrate any choice as progress.

### Key Interaction Paradigms

**The Three Doors (Primary Interaction):**
The main interface presents three tasks simultaneously as entry points. These tasks should be:
- **Intentionally diverse** - Different types of activities (e.g., creative vs. administrative vs. physical, or high-focus vs. low-friction vs. context-switching)
- **Small at first** - Especially in early usage, doors should present approachable tasks to build momentum
- **All viable next steps** - Each represents a legitimate priority, not filler options
- **Learning opportunities** - User's choice (or avoidance) informs the system about current mood, energy, and capacity

Over time, the system learns: "On Tuesday mornings, user picks Door 1 (focused work). On Friday afternoons, user picks Door 3 (quick wins). User never picks administrative tasks before 10am."

**Door Refresh & Feedback (MVP Core):**
- **Refresh/New Doors** - Simple keystroke (e.g., 'R' or 'N') to generate a new set of three doors if current options don't appeal. No judgment, no friction—just new options.
- **Door Feedback** - Option to indicate why a door isn't suitable (basic MVP options):
  - "Blocked" - Task cannot proceed (captures blocker)
  - "Not now" - Task is valid but doesn't fit current mood/context (teaches system about state)
  - "Needs breakdown" - Task is too big/unclear (MVP: flag for later attention; Post-MVP: may trigger breakdown assistance)
  - "Other comment" - Freeform note about the task (refactoring, context, etc.)

These interactions serve dual purposes: give users control (preventing feeling trapped) and provide rich learning signal to the system about task suitability, blockers, and user state.

**Choose-Your-Own-Adventure Navigation:**
Beyond the three doors, other decision points present 3-5 contextual options rather than requiring command memorization. Options adapt based on state and history.

**Progressive Disclosure:**
Start simple, reveal complexity only when needed. Quick add mode for speed, expanded capture for context when desired. Don't force decisions upfront.

**Persistent Context:**
Values/goals remain visible (but unobtrusive) throughout the session—likely as a subtle header or footer—reminding users of the "why" while working on the "what."

**Encouraging Tone:**
All messaging embodies "progress over perfection." Copy celebrates any action ("You picked a door and started. That's what matters.").

### Core Screens and Views

From a product perspective, these are the critical views necessary to deliver MVP value:

1. **Three Doors Dashboard (Primary Interface)** - Session entry point presenting three diverse tasks as "doors" to choose, with minimal surrounding context. Core question: "Which door feels right today?" Includes refresh option and per-door feedback mechanism.

2. **Task List View** - Full task display when user wants to see beyond the three doors, with filtering and status

3. **Quick Add Flow** - Minimal-friction task capture (possibly single input field)

4. **Extended Capture Flow** - Optional deeper capture including "why" context and task metadata (effort, type, context)

5. **Values/Goals Setup** - Initial and ongoing management of user-defined values that guide prioritization

6. **Progress View** - Visualization showing "better than yesterday" metrics and door choice patterns over time (e.g., "You've opened 5 doors this week, up from 3 last week" and "You tend to pick Door 1 in mornings, Door 3 in afternoons")

7. **Health Check View** - Diagnostic display showing Apple Notes connectivity and sync status

8. **Improvement Prompt** - End-of-session single question asking for one improvement

### Accessibility

**None** - MVP focuses on terminal interface for single user (developer). Accessibility requirements deferred to future phases when/if user base expands beyond CLI-comfortable users.

### Branding

**Terminal Aesthetic with Warmth:**
Leverage Charm Bracelet/Bubbletea's capabilities for styled terminal UI—think clean, readable typography with subtle use of color for status indication (green for progress, yellow for prompts, red sparingly for errors).

**Three Doors Visual Metaphor:**
The main interface could literally render three visual "doors" in ASCII art or styled terminal boxes:
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   DOOR 1    │  │   DOOR 2    │  │   DOOR 3    │
│             │  │             │  │             │
│  [Task A]   │  │  [Task B]   │  │  [Task C]   │
│  Quick win  │  │  Deep work  │  │  Creative   │
│  ~5min      │  │  ~30min     │  │  ~15min     │
└─────────────┘  └─────────────┘  └─────────────┘

Press 1, 2, or 3 to enter  |  R to refresh  |  B to mark blocked
```

**"Progress Over Perfection" Visual Language:**
Use asymmetry, incomplete progress bars, and "good enough" indicators. The three doors might be slightly different sizes or styles, reinforcing that perfection isn't required—just pick one and start.

### Target Device and Platforms

**Primary: macOS Terminal Emulators (iTerm2, Terminal.app, Alacritty)**
- CLI/TUI optimized for 80x24 minimum, responsive to larger terminal sizes
- Assumes modern terminal with 256-color support minimum
- Keyboard-driven navigation (arrow keys, vim-style hjkl, number keys 1-3 for door selection)

**Secondary: Remote Terminal Access**
- Should function over SSH connections (for future Geodesic/remote environment access)
- ASCII fallback for constrained environments

**Mobile Access (Indirect):**
- No dedicated mobile UI in MVP
- Mobile interaction happens through Apple Notes app directly (view/edit tasks on iPhone)
- Sync bidirectionally when user returns to terminal interface

---

## Technical Assumptions

### Technical Demo Phase Architecture

**Decision:** Minimal monolithic application with simple text file I/O

**Rationale:**
- **Speed to validation**: Build and test in 4-8 hours
- **Simple is fast**: No database, no complex integrations, no abstractions until needed
- **Easy external task population**: Text files can be edited with any editor, populated from scripts, etc.
- **Prove the concept first**: Validate Three Doors UX before investing in infrastructure
- **Low risk**: Can throw away and rebuild if concept fails validation

**Tech Demo Structure:**
```
ThreeDoors/
├── cmd/
│   └── threedoors/        # Main application (single file initially)
├── internal/
│   ├── tui/              # Bubbletea Three Doors interface
│   └── tasks/            # Simple file I/O (read tasks.txt, write completed.txt)
├── docs/                  # Documentation (including this PRD)
├── .bmad-core/           # BMAD methodology artifacts
├── Makefile              # Simple build: build, run, clean
└── README.md             # Quick start guide
```

**Data Files (created at runtime in `~/.threedoors/`):**
```
~/.threedoors/
├── tasks.txt             # One task per line (user can edit directly)
├── completed.txt         # Completed tasks with timestamps
└── config.txt            # Optional: Simple key=value config (if needed)
```

---

### Full MVP Architecture (Post-Validation - Deferred)

**Structure evolves to:**
```
ThreeDoors/
├── cmd/                    # CLI entry points
│   └── threedoors/        # Main application
├── internal/              # Private application code
│   ├── core/             # Core domain logic
│   ├── tui/              # Bubbletea interface components
│   ├── integrations/     # Adapter implementations
│   │   ├── textfile/    # Text file backend (from Tech Demo)
│   │   └── applenotes/  # Apple Notes integration
│   ├── enrichment/       # Local enrichment storage
│   └── learning/         # Door selection & pattern tracking
├── pkg/                   # Public, reusable packages (if any)
├── docs/                  # Documentation (including this PRD)
├── .bmad-core/           # BMAD methodology artifacts
└── Makefile              # Build automation
```

### Service Architecture

**Technical Demo Phase:**

**Decision:** Single-layer CLI/TUI application with direct file I/O

**Rationale:**
- **No abstractions yet**: Build for one thing (text files), refactor when adding second thing
- **Validate UX first**: Door selection algorithm is the innovation, not the data layer
- **Fast iteration**: Change anything without navigating architecture layers

**Demo Architecture:**
- **TUI Layer (Bubbletea)** - Three Doors interface, keyboard handling, rendering
- **Direct File I/O** - Read tasks.txt, write completed.txt, no abstraction layer
- **Simple Door Selection** - Random selection of 3 tasks from available pool (no learning/categorization yet)

---

**Full MVP Phase (Post-Validation - Deferred):**

**Decision:** Layered architecture with pluggable integration adapters

**Architecture Layers:**
1. **TUI Layer (Bubbletea)** - User interaction, rendering, keyboard handling
2. **Core Domain Logic** - Task management, door selection algorithm, progress tracking
3. **Integration Adapters** - Abstract interface with concrete implementations (text file, Apple Notes, others later)
4. **Enrichment Storage** - Metadata, cross-references, learning patterns not stored in source systems
5. **Configuration & State** - User preferences, values/goals, application state

**Key Architectural Principles:**
- Core domain logic has NO dependencies on specific integrations (dependency inversion)
- Integrations implement common `TaskProvider` interface
- Enrichment layer wraps tasks from any source with additional metadata
- TUI layer depends only on core domain, not specific integrations

### Testing Requirements

**Technical Demo Phase:**

**Decision:** Manual testing only - validate UX through real use

**Demo Testing Approach:**
- **No automated tests for Tech Demo** - premature given throwaway prototype nature
- **Manual testing** via daily use for 1 week
- **Success measurement**: Does Three Doors feel better than a list? Yes/No decision point
- **Quality gate**: If it crashes or feels bad to use, iterate or abandon concept

**Rationale:**
- 4-8 hours to build entire demo - testing infrastructure would consume half that time
- Real usage is the test: if developer won't use it daily, concept fails regardless of test coverage
- Can add tests when/if proceeding to Full MVP

---

**Full MVP Phase (Post-Validation - Deferred):**

**Testing Scope:**
- **Unit tests** for core domain logic (door selection algorithm, categorization, progress tracking)
- **Integration tests** for backend adapters (text file, Apple Notes)
- **Manual testing** for TUI interactions (Bubbletea testing framework is immature)

**Test Coverage Goals:**
- Core domain logic: 70%+ coverage (pragmatic, not perfectionist)
- Integration adapters: Critical paths covered (read, write, sync scenarios)
- TUI layer: Manual testing via developer use

**Testing Strategy:**
- Table-driven tests (idiomatic Go pattern)
- Test fixtures for data structures
- Mock `TaskProvider` interface for testing core logic without real integrations
- CI/CD runs tests on every commit (GitHub Actions)

**Deferred for Post-MVP:**
- End-to-end testing framework
- Property-based testing for door selection algorithm
- Performance/load testing

### Additional Technical Assumptions and Requests

**Technical Demo Phase Assumptions:**

**Text File Format:**
- **Simple line-delimited format**: One task per line in `tasks.txt`
- **Completed format**: `[timestamp] task description` in `completed.txt`
- **No metadata yet**: Task is just text; no categories, priorities, or context for Tech Demo
- **Easy population**: User can edit files with any text editor, generate from scripts, copy-paste, etc.

**Door Selection Algorithm (Tech Demo):**
- **Random selection**: Pick 3 random tasks from available pool
- **Simple diversity**: Ensure no duplicates in the three doors
- **No intelligence yet**: No learning, no categorization, no context awareness
- **Validation goal**: Prove that having 3 options reduces friction vs. scrolling a full list

**File I/O:**
- **Go standard library**: Use `os`, `bufio`, `io/ioutil` - no external dependencies for file operations
- **Error handling**: Create files with defaults if missing; graceful degradation if corrupted
- **Concurrency**: Not a concern for single-user local files

---

**Full MVP Phase Assumptions (Post-Validation - Deferred):**

**Apple Notes Integration:**
- **Options Identified (2025):**
  1. **DarwinKit (github.com/progrium/darwinkit)** - Native macOS API access from Go; requires translating Objective-C patterns; full API control but higher complexity
  2. **Direct SQLite Database Access** - Apple Notes stores data in `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite`; note content is gzipped protocol buffers in `ZICNOTEDATA.ZDATA` column; read-only safe, write risks corruption
  3. **AppleScript Bridge** - Use `os/exec` to invoke AppleScript; simpler than native APIs; proven approach (see `sballin/alfred-search-notes-app`)
  4. **Existing MCP Server** - `mcp-apple-notes` server exists for Apple Notes integration; could potentially leverage this instead of building from scratch
- **Assumption:** Multiple viable paths exist; choice depends on read-only vs. read-write needs, complexity tolerance, and reliability requirements (WILL REQUIRE VALIDATION when implementing Phase 2)
- **Spike Required:** Evaluate options before implementing Apple Notes integration
- **Preferred Exploration Order:** Start with Option 4 (MCP server) or Option 2 (SQLite read-only), fall back to Option 3 (AppleScript) if bidirectional sync required, reserve Option 1 (DarwinKit) for complex scenarios

**Cloud Storage for Cross-Computer Sync (DEFERRED - Not MVP):**
- **Status:** Cross-computer sync is deferred post-MVP; single-computer local storage is sufficient for initial development and use
- **Future Exploration:** When implementing sync, explore alternatives to monolithic SQLite file:
  - Individual JSON/YAML files per task or per day (more granular, better suited for file-based cloud sync)
  - Conflict-free Replicated Data Types (CRDTs) for eventual consistency
  - Event sourcing with append-only logs
  - Cloud-native solutions (S3, Firebase, etc.) if local-first constraint relaxes
- **Awareness:** Monolithic SQLite on cloud storage (iCloud/Google Drive) is known problematic—corruption risk, locking issues, slow sync
- **MVP Decision:** Store enrichment data locally only; revisit sync architecture when/if multi-computer use becomes actual need

**Go Language & Ecosystem (Tech Demo):**
- **Language:** Go 1.25.4+ (current stable as of November 2025)
- **Formatting:** `gofumpt` (run before commits)
- **Linting:** Skip for Tech Demo (adds no validation value at this stage)
- **Dependency Management:** Go modules
- **TUI Framework:** Bubbletea + Lipgloss (styling) - minimal Bubbles components, only if needed

**Data Storage (Tech Demo):**
- **Storage:** Plain text files in `~/.threedoors/`
- **No database**: Not needed for line-delimited text
- **No configuration file initially**: Hardcode paths, add config only if needed

**Build & Development (Tech Demo):**
- **Build System:** Minimal Makefile
  ```makefile
  build:
      go build -o bin/threedoors cmd/threedoors/main.go

  run: build
      ./bin/threedoors

  clean:
      rm -rf bin/
  ```
- **Development Workflow:** Direct iteration on macOS
- **No CI/CD for Tech Demo**: Overkill for validation prototype

**Performance Expectations (Tech Demo):**
- **File I/O**: <10ms to read tasks.txt (even with 100+ tasks)
- **Door selection**: <1ms for random selection from array
- **TUI rendering**: Bubbletea handles 60fps, not a concern
- **Startup time**: <100ms total from launch to Three Doors display

**Security & Privacy (Tech Demo):**
- **Local files only**: No network, no external services
- **No logging**: Not even metadata for Tech Demo
- **File permissions**: Standard user file permissions on `~/.threedoors/`

---

**Full MVP Phase (Post-Validation - Deferred):**

**Go Language & Ecosystem:**
- **Language:** Go 1.25.4+
- **Formatting:** `gofumpt`
- **Linting:** `golangci-lint` with standard rule set
- **Dependency Management:** Go modules
- **TUI Framework:** Bubbletea + Lipgloss + Bubbles

**Data Storage:**
- **Primary:** Apple Notes (user-facing tasks) or text file backend
- **Enrichment:** SQLite for metadata (door feedback, blockers, categorization, learning patterns)
- **Configuration:** YAML or TOML for user preferences, values/goals
- **Location:** `~/.config/threedoors/` (XDG Base Directory spec on Linux, macOS equivalent)

**Build & Development:**
- **Build System:** Makefile with full targets (build, test, lint, install)
- **CI/CD:** GitHub Actions running tests on every commit
- **Development Workflow:** Direct iteration on macOS

**Performance Expectations:**
- Door selection algorithm: <100ms to choose 3 tasks from up to 1000 total tasks
- Backend sync: <2 seconds for typical data set
- TUI rendering: 60fps equivalent for smooth interaction

**Deferred Technical Decisions (Post-MVP):**
- Cross-computer sync architecture (see deferred section above)
- LLM provider integration architecture (local vs. cloud, which providers)
- Additional integration adapters (Jira, Linear, Google Calendar, etc.)
- Remote access agent for Geodesic environments
- Vector database for semantic task search
- Voice interface integration

---

## Epic List

### Phase 1: Technical Demo & Validation (Immediate - Week 1)

**Epic 1: Three Doors Technical Demo**
- **Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists
- **Timeline:** 1 week (4-8 hours development time)
- **Deliverables:** Working CLI/TUI showing Three Doors, reading from text file, marking tasks complete, basic progress tracking
- **Success Criteria:**
  - Developer uses tool daily for 1 week
  - Three Doors selection feels meaningfully different from scrolling a list
  - Decision point reached: proceed to Full MVP or pivot/abandon
- **Tech Stack:** Go 1.25.4+, Bubbletea/Lipgloss, local text files
- **Risk:** UX concept might not feel better than simple list; easy to pivot if fails

---

### Phase 2: Post-Validation Roadmap (Conditional on Phase 1 Success)

**DECISION GATE:** Only proceed with these epics if Technical Demo validates the Three Doors concept through real usage.

**Epic 2: Foundation & Apple Notes Integration**
- **Goal:** Replace text file backend with Apple Notes integration, enabling mobile task editing while maintaining Three Doors UX
- **Prerequisites:** Epic 1 success; Apple Notes integration spike completed
- **Deliverables:**
  - Refactor to adapter pattern (text file + Apple Notes backends)
  - Bidirectional sync with Apple Notes
  - Health check command for Notes connectivity
  - Migration path from text files to Notes
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week (includes spike + implementation)
- **Risk:** Apple Notes integration complexity could exceed estimates; fallback to improved text file backend

**Epic 3: Enhanced Interaction & Task Context**
- **Goal:** Add task capture, values/goals display, and basic feedback mechanisms to improve task management workflow
- **Prerequisites:** Epic 2 complete (stable backend integration)
- **Deliverables:**
  - Quick add mode for task capture
  - Extended capture with "why" context
  - Values/goals setup and persistent display
  - Door feedback options (Blocked, Not now, Needs breakdown)
  - Blocker tracking
  - Improvement prompt at session end
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Feature creep; maintain focus on minimal valuable additions

**Epic 4: Learning & Intelligent Door Selection**
- **Goal:** Implement pattern tracking and learning to make door selection context-aware and adaptive to user preferences
- **Prerequisites:** Epic 3 complete (enough usage data to learn from)
- **Deliverables:**
  - Task categorization (type, effort level, context)
  - Door selection pattern tracking
  - Learning algorithm that adapts based on user choices
  - Progress view showing door choice patterns over time
  - "Better than yesterday" multi-dimensional tracking
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week
- **Risk:** Algorithm complexity; may need to simplify learning approach

**Epic 5: Data Layer & Enrichment (Optional)**
- **Goal:** Add enrichment storage layer for metadata that cannot live in source systems
- **Prerequisites:** Epic 4 complete; proven need for enrichment beyond what backends support
- **Deliverables:**
  - SQLite enrichment database
  - Cross-reference tracking (tasks across multiple systems)
  - Metadata not supported by Apple Notes (categories, learning patterns, etc.)
  - Data migration and backup tooling
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** May be YAGNI; consider deferring indefinitely if not clearly needed

---

### Phase 3: Future Expansion (12+ months out)

**Epic 6+: Additional Integrations** (Jira, Linear, Google Calendar, Slack, etc.)
**Epic 7+: Cross-Computer Sync** (Implement alternative to monolithic SQLite on cloud storage)
**Epic 8+: LLM Integration** (Task breakdown assistance, assumption challenging, dependency collapse)
**Epic 9+: Advanced Features** (Voice interface, mobile app, web interface, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---

## Epic Details

### Epic 1: Three Doors Technical Demo

**Epic Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists.

**Scope:** CLI/TUI application that reads tasks from a text file, presents three random tasks as "doors," allows selection and completion, and tracks progress.

---

#### Story 1.1: Project Setup & Basic Bubbletea App

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

---

#### Story 1.2: File I/O for Tasks

**As a** user,
**I want** the app to read my tasks from a simple text file,
**so that** I can populate tasks easily outside the app.

**Acceptance Criteria:**
1. Application creates `~/.threedoors/` directory on first run if it doesn't exist
2. Application reads tasks from `~/.threedoors/tasks.txt` (one task per line)
3. If `tasks.txt` doesn't exist, create it with 5 sample tasks as examples
4. Empty lines and lines starting with `#` are ignored (comments)
5. Application displays count of loaded tasks (e.g., "Loaded 12 tasks")
6. Gracefully handles file read errors with helpful error message

---

#### Story 1.3: Three Doors Display

**As a** user,
**I want** to see three tasks displayed as "doors,"
**so that** I can quickly choose what to work on without scanning a long list.

**Acceptance Criteria:**
1. Three tasks are randomly selected from loaded tasks
2. Tasks are displayed in three visual "boxes" (ASCII art or Lipgloss styled borders)
3. Each door is labeled: "Door 1", "Door 2", "Door 3"
4. Task text is displayed inside each door (truncated if too long, max ~40 chars)
5. Instructions displayed at bottom: "Press 1, 2, or 3 to select | R to refresh | Q to quit"
6. No duplicate tasks appear in the three doors simultaneously
7. If fewer than 3 tasks available, show what's available (handle edge case gracefully)

---

#### Story 1.4: Door Selection & Task Completion

**As a** user,
**I want** to select a door and mark the task as complete,
**so that** I can make progress on my tasks.

**Acceptance Criteria:**
1. Pressing 1, 2, or 3 selects the corresponding door
2. Selected task is highlighted/indicated visually
3. Prompt appears: "Working on: [task text] - Press C to complete, B to go back"
4. Pressing C marks task as complete
5. Completed task is removed from available task pool
6. New set of three doors is displayed automatically after completion
7. Session completion count increments and displays (e.g., "Completed this session: 3")

---

#### Story 1.5: Completed Tasks Tracking

**As a** user,
**I want** completed tasks saved to a file with timestamps,
**so that** I have a record of what I've accomplished.

**Acceptance Criteria:**
1. Completed tasks are appended to `~/.threedoors/completed.txt`
2. Format: `[YYYY-MM-DD HH:MM:SS] task description`
3. File is created if it doesn't exist
4. Session completion count is displayed on screen (e.g., "✓ Completed this session: 3")
5. "Progress over perfection" message shown after completing a task (e.g., "Nice! Any progress is good progress.")

---

#### Story 1.6: Door Refresh Mechanism

**As a** user,
**I want** to refresh the three doors if none appeal to me,
**so that** I have control over my options without feeling trapped.

**Acceptance Criteria:**
1. Pressing R generates a new set of three doors
2. New selection is different from current selection (no duplicates of currently shown tasks)
3. Random selection ensures variety over multiple refreshes
4. Refresh count is tracked and displayed (optional, for learning about user behavior)
5. Edge case: If 3 or fewer tasks remain total, show message "All available tasks are already showing"

---

#### Story 1.7: Polish, Styling & Edge Case Handling

**As a** user,
**I want** the app to feel polished and handle edge cases gracefully,
**so that** I enjoy using it and don't encounter confusing states.

**Acceptance Criteria:**
1. Lipgloss styling applied: colors for doors, headers, messages (green for success, yellow for prompts)
2. "Progress over perfection" messaging embedded (e.g., on startup, after completion)
3. Edge case: All tasks completed shows celebratory message and option to add more tasks
4. Edge case: Only 1-2 tasks remaining displays them gracefully (not three empty doors)
5. Edge case: Empty tasks.txt shows helpful message: "Add tasks to ~/.threedoors/tasks.txt to get started!"
6. Application feels responsive and smooth (no noticeable lag)
7. README.md created with quick start instructions

---

### Epic 2-5: Post-Validation Epics (Placeholder)

**Note:** These epics are placeholders for post-validation planning. Detailed stories will be created only if Epic 1 successfully validates the Three Doors concept.

**Epic 2: Foundation & Apple Notes Integration**
*Stories to be defined after Epic 1 validation and Apple Notes integration spike*

**Epic 3: Enhanced Interaction & Task Context**
*Stories to be defined based on learnings from Epic 2 usage patterns*

**Epic 4: Learning & Intelligent Door Selection**
*Stories to be defined based on sufficient usage data from Epic 3*

**Epic 5: Data Layer & Enrichment (Optional)**
*Stories to be defined only if clear need emerges from Epic 4*

---

## Checklist Results Report

### Executive Summary

**Overall PRD Completeness:** 95%

**MVP Scope Appropriateness:** Just Right (pivoted to Technical Demo & Validation approach)

**Readiness for Architecture Phase:** Ready

**Most Critical Observations:**
- Excellent pivot to Technical Demo validates concept before major investment
- Clear phased approach with decision gate prevents premature optimization
- Technical assumptions well-researched (Go 1.25.4, Apple Notes options, Context7 MCP)
- Story breakdown appropriately sized for 4-8 hour timeline
- Minor gaps don't block progress

---

### Category Analysis

| Category                         | Status  | Critical Issues |
| -------------------------------- | ------- | --------------- |
| 1. Problem Definition & Context  | PASS    | None |
| 2. MVP Scope Definition          | PASS    | None - excellent scope discipline |
| 3. User Experience Requirements  | PASS    | None |
| 4. Functional Requirements       | PASS    | None |
| 5. Non-Functional Requirements   | PASS    | None |
| 6. Epic & Story Structure        | PASS    | None |
| 7. Technical Guidance            | PASS    | None - well-researched |
| 8. Cross-Functional Requirements | PARTIAL | Minor: Task data model could be more explicit |
| 9. Clarity & Communication       | PASS    | None |

---

### Top Issues by Priority

**BLOCKERS:** None

**HIGH:** None

**MEDIUM:**
1. **Task data model**: While simple (line of text), could explicitly document what constitutes a task (format, max length, special characters handling)
2. **Post-validation decision criteria**: "Feels better than a list" is subjective; could add specific measurement approaches (e.g., count refreshes per session, time to select task, subjective 1-10 rating)

**LOW:**
1. **Visual mockups**: Three Doors described in text but no ASCII mockup; helpful but not blocking
2. **Error message catalog**: Could pre-define friendly error messages for common scenarios

---

### MVP Scope Assessment

**Scope Appropriateness: EXCELLENT**

**Strengths:**
- Technical Demo approach validates core hypothesis before investing in complexity
- 4-8 hour timeline is achievable and realistic
- Text file storage removes Apple Notes integration risk from critical path
- Clear success criteria (daily use for 1 week)
- Decision gate prevents continuing down wrong path

**Potential Cuts (if needed):**
- Story 1.7 (Polish & Styling) could be reduced if time-constrained; core UX works without perfect styling

**Missing Features (none essential for Tech Demo):**
- All deferred features appropriately moved to post-validation phases

**Complexity Concerns:**
- Bubbletea learning curve is only complexity; mitigated by simple requirements

**Timeline Realism:**
- 4-8 hours for 7 stories: ~30-60 min per story with buffer
- Very realistic given simplicity of text file I/O and clear acceptance criteria

---

### Technical Readiness

**Technical Constraints: CLEAR**

**Specified Constraints:**
- Go 1.25.4+
- Bubbletea + Lipgloss
- macOS primary platform
- Text files in `~/.threedoors/`
- No external dependencies for Tech Demo

**Identified Technical Risks:**
- **LOW**: Bubbletea learning curve (mitigated: good documentation, simple use case)
- **DEFERRED**: Apple Notes integration (appropriate - not needed for Tech Demo)

**Areas Needing Architect Investigation:**
- None for Tech Demo phase
- Apple Notes integration options documented for future investigation (4 approaches identified with Context7 MCP)

**Architecture Guidance Quality:**
- Architecture section provides clear direction
- "No abstractions yet" guidance prevents over-engineering
- Post-validation architecture evolution path documented

---

### Recommendations

**For Immediate Next Steps:**

1. **Proceed to Development** - PRD is ready for Story 1.1 implementation
   - All blockers resolved
   - Technical stack clear
   - Acceptance criteria testable

2. **Consider Adding (Optional - Not Blocking):**
   - Quick ASCII mockup of Three Doors layout (5 min exercise, helps visualize before coding)
   - Explicit task data model: `Task = string (max 200 chars, UTF-8)` just for completeness
   - Decision criteria template for post-validation: "Rate 1-10: Did Three Doors reduce friction? Would you continue using?"

3. **Track During Development:**
   - Actual time per story (validate 4-8 hour estimate)
   - Bubbletea learning curve challenges (inform future estimates)
   - User experience insights during validation week (feed into Epic 2+ planning)

**For Post-Validation (If Epic 1 Succeeds):**

4. **Before Epic 2:**
   - Run Apple Notes integration spike (evaluate 4 identified options)
   - Define explicit success criteria from Epic 1 learnings

5. **Documentation:**
   - Capture Epic 1 retrospective learnings
   - Update PRD with actual Technical Demo results before proceeding to Epic 2

---

### Strengths Worth Highlighting

1. **Pragmatic Scope Management**: Pivot to Technical Demo demonstrates excellent product thinking - validate before investing
2. **Research Quality**: Technical assumptions informed by current, accurate information (Go 1.25.4, Apple Notes options via Context7 MCP)
3. **Risk Mitigation**: Text file approach removes highest risk (Apple Notes) from critical path
4. **Clear Decision Gates**: Explicit decision point after validation prevents sunk cost fallacy
5. **Story Quality**: Acceptance criteria are specific, testable, and sized appropriately
6. **BMAD Alignment**: Process demonstrates "progress over perfection" philosophy the product espouses

---

### Final Decision

**✅ READY FOR DEVELOPMENT**

The PRD is comprehensive, properly structured, and ready for immediate implementation of Epic 1. The Technical Demo & Validation approach mitigates risk excellently while maintaining the vision for future expansion.

**Recommended Next Action:** Begin Story 1.1 (Project Setup & Basic Bubbletea App)

---

## Next Steps

### For Developer: Begin Technical Demo Implementation

**Objective:** Implement Epic 1 (Three Doors Technical Demo) following the user stories in sequence.

**Starting Point:** Story 1.1 - Project Setup & Basic Bubbletea App

**Recommended Approach:**

1. **Review the PRD thoroughly**, especially:
   - Technical Demo Requirements (TD1-TD9)
   - Epic 1 Stories (1.1 through 1.7)
   - Technical Assumptions for Tech Demo Phase
   - Acceptance Criteria for each story

2. **Set up development environment:**
   - Ensure Go 1.25.4+ is installed
   - Choose your preferred editor/IDE with Go support
   - Prepare terminal emulator (iTerm2 or similar)

3. **Execute stories sequentially:**
   - Complete Story 1.1 fully (all acceptance criteria met) before moving to 1.2
   - Each story builds on the previous one
   - Time-box each story to ~30-60 minutes; if significantly over, reassess approach

4. **Track progress:**
   - Note actual time spent per story (validates estimates)
   - Document any challenges encountered (especially Bubbletea learning curve)
   - Capture UX insights during daily use

5. **Validation phase:**
   - After completing Story 1.7, use the app daily for 1 week
   - Observe: Does Three Doors reduce friction vs. scrolling a list?
   - Document decision criteria results (proceed to Epic 2 or pivot?)

**Quick Start Prompt for Story 1.1:**

```bash
# Navigate to project directory
cd ~/work/simple-todo

# Initialize Go module (if not already done)
go mod init github.com/arcaven/ThreeDoors

# Add Bubbletea and Lipgloss dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest

# Create initial structure
mkdir -p cmd/threedoors
mkdir -p internal/tui

# Create basic main.go following Bubbletea "Hello World" pattern
# Target: App renders "ThreeDoors - Technical Demo" header and responds to 'q' to quit
```

**Reference Resources:**
- Bubbletea Tutorial: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- Lipgloss Examples: https://github.com/charmbracelet/lipgloss/tree/master/examples
- Go 1.25 Release Notes: https://go.dev/doc/go1.25

---

### Post-Validation: Next PRD Iteration

**If Technical Demo Succeeds (Decision Gate Passed):**

Create Epic 2 detailed stories for Apple Notes Integration:
- Run Apple Notes integration spike (evaluate 4 identified options with Context7 MCP)
- Define specific success criteria based on Epic 1 learnings
- Refine requirements based on actual usage patterns from validation week
- Update PRD version to 2.0 with Epic 2 details

**If Technical Demo Fails (Pivot Needed):**

Retrospective and reassessment:
- Document what didn't work about Three Doors concept
- Identify alternative approaches to the original problem (reduce todo app friction)
- Consider: Was the problem statement correct? Was the solution wrong? Both?
- Decide: Iterate on Three Doors design or pursue different solution entirely

---

*This PRD embodies "progress over perfection" - it's comprehensive enough to start building, flexible enough to adapt based on learnings, and structured to prevent premature investment in unvalidated concepts.*

---

**Document Complete**

*Generated using BMAD-METHOD™ framework*
*PM Agent: John*
*Session Date: 2025-11-07*
