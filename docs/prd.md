# ThreeDoors Product Requirements Document (PRD)

**Document Version:** 1.0
**Last Updated:** 2025-11-07
**Project Repository:** github.com/arcaven/ThreeDoors.git

---

## Goals and Background Context

### Goals

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

---

## Requirements

### Functional Requirements

**FR1:** The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)

**FR2:** The system shall integrate with Apple Notes as the primary task storage backend, enabling bidirectional sync

**FR3:** The system shall allow users to capture new tasks with optional context (what and why) through the CLI/TUI

**FR4:** The system shall retrieve and display tasks from Apple Notes within the application interface

**FR5:** The system shall allow users to mark tasks as complete, updating both the application state and Apple Notes

**FR6:** The system shall display user-defined values and goals persistently throughout task work sessions

**FR7:** The system shall provide a "choose-your-own-adventure" interactive navigation flow that presents options rather than demands

**FR8:** The system shall track daily task completion count and display comparison to previous day's count

**FR9:** The system shall prompt the user once per session with: "What's one thing you could improve about this list/task/goal right now?"

**FR10:** The system shall embed "progress over perfection" messaging throughout interaction patterns and interface copy

**FR11:** The system shall maintain a local enrichment layer (SQLite and/or vector database) for metadata, cross-references, and relationships that cannot be stored in source systems

**FR12:** The system shall allow updates to tasks from either the application or directly in Apple Notes on iPhone, with changes reflected bidirectionally

**FR15:** The system shall provide a health check command to verify Apple Notes connectivity and database integrity

**FR16:** The system shall support a "quick add" mode for capturing tasks with minimal interaction

**FR17:** The system shall provide a refresh mechanism to generate a new set of three doors when current options don't appeal to the user

**FR18:** The system shall allow users to provide feedback on why a specific door isn't suitable with options: Blocked, Not now, Needs breakdown, or Other comment

**FR19:** The system shall capture and store blocker information when a task is marked as blocked

**FR20:** The system shall use door selection and feedback patterns to inform future door selection (learning which task types suit which contexts)

**FR21:** The system shall categorize tasks by type, effort level, and context to enable diverse door selection

### Non-Functional Requirements

**NFR1:** The system shall be built in Go using idiomatic patterns and gofumpt formatting standards

**NFR2:** The system shall use the Bubbletea/Charm Bracelet ecosystem for TUI implementation

**NFR3:** The system shall operate on macOS as the primary and MVP target platform

**NFR4:** The system shall store all user data locally or in the user's iCloud (via Apple Notes), with no external telemetry or tracking

**NFR5:** The system shall store application state and enrichment data on iCloud Drive or Google Drive to enable cross-computer sync

**NFR6:** The system shall handle concurrent access scenarios gracefully to prevent SQLite database corruption when opened on multiple computers

**NFR7:** The system shall respond to user interactions within the CLI/TUI with minimal latency (target: <500ms for typical operations)

**NFR8:** The system shall provide graceful degradation when Apple Notes integration is unavailable, maintaining core functionality

**NFR9:** The system shall implement secure credential storage using OS keychain for any API keys or authentication tokens

**NFR10:** The system shall never log sensitive user data or credentials

**NFR11:** The system shall use Make or Task as the build system

**NFR12:** The system shall maintain clear architectural separation between core engine, TUI layer, integration adapters, and enrichment storage

**NFR13:** The system shall maintain data integrity even when Apple Notes is modified externally while app is running

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

### Repository Structure: Monorepo

**Decision:** Single repository containing all ThreeDoors components.

**Rationale:**
- Single-developer project benefits from simplified workflow (one clone, one build, one set of issues)
- All components (core engine, TUI, Apple Notes adapter, enrichment layer) are tightly coupled in MVP
- No need for independent versioning or deployment of separate services
- Easier to maintain consistency across codebase during rapid iteration
- Can restructure later if project evolves to need separation

**Structure:**
```
ThreeDoors/
├── cmd/                    # CLI entry points
│   └── threedoors/        # Main application
├── internal/              # Private application code
│   ├── core/             # Core domain logic
│   ├── tui/              # Bubbletea interface components
│   ├── integrations/     # Adapter implementations
│   │   └── applenotes/  # Apple Notes integration
│   ├── enrichment/       # Local enrichment storage
│   └── learning/         # Door selection & pattern tracking
├── pkg/                   # Public, reusable packages (if any)
├── docs/                  # Documentation (including this PRD)
├── .bmad-core/           # BMAD methodology artifacts
└── Makefile              # Build automation
```

### Service Architecture

**Decision:** Monolithic CLI/TUI application with pluggable integration adapters

**Rationale:**
- MVP serves single user on single machine—no need for distributed architecture
- CLI/TUI constraint means no web server, no API endpoints for MVP
- Pluggable adapter pattern allows future integrations (Jira, Linear, etc.) without core rewrites
- Local-first architecture: all data processing happens on user's machine
- Enrichment layer co-located with application, stored locally for MVP

**Architecture Layers:**
1. **TUI Layer (Bubbletea)** - User interaction, rendering, keyboard handling
2. **Core Domain Logic** - Task management, door selection algorithm, progress tracking
3. **Integration Adapters** - Abstract interface with concrete implementations (Apple Notes first, others later)
4. **Enrichment Storage** - Metadata, cross-references, learning patterns not stored in source systems (local storage for MVP)
5. **Configuration & State** - User preferences, values/goals, application state

**Key Architectural Principles:**
- Core domain logic has NO dependencies on specific integrations (dependency inversion)
- Integrations implement common `TaskProvider` interface
- Enrichment layer wraps tasks from any source with additional metadata
- TUI layer depends only on core domain, not specific integrations

### Testing Requirements

**Decision:** Unit + Integration testing focused on core logic and critical paths

**MVP Testing Scope:**
- **Unit tests** for core domain logic (door selection algorithm, categorization, progress tracking)
- **Integration tests** for Apple Notes adapter (mocked initially, real integration in CI if feasible)
- **Manual testing** for TUI interactions (Bubbletea testing framework is immature)
- **Health check command (FR15)** serves as smoke test for deployment validation

**Test Coverage Goals:**
- Core domain logic: 80%+ coverage
- Integration adapters: Critical paths covered (read, write, sync scenarios)
- TUI layer: Manual testing via developer use

**Testing Strategy:**
- Table-driven tests (idiomatic Go pattern)
- Test fixtures for Apple Notes data structures
- Mock `TaskProvider` interface for testing core logic without real integrations
- CI/CD runs tests on every commit (GitHub Actions)

**Deferred for Post-MVP:**
- End-to-end testing framework
- Property-based testing for door selection algorithm
- Performance/load testing
- Accessibility testing (N/A for MVP given CLI/TUI constraint)

### Additional Technical Assumptions and Requests

**Apple Notes Integration:**
- **Primary Approach:** Investigate native macOS frameworks accessible from Go (potentially via cgo with Foundation/AppKit bindings, or existing Go libraries wrapping Notes functionality)
- **Fallback:** AppleScript bridge via `os/exec` if native integration proves impractical
- **Assumption:** Some viable path exists to read/write Apple Notes from Go (REQUIRES VALIDATION - highest risk technical assumption)
- **Spike Required:** First development session must validate Apple Notes integration feasibility, starting with native options before falling back to AppleScript
- **Research Areas:** Go cgo macOS bindings, existing Go libraries for Notes access, Notes SQLite database structure if accessible

**Cloud Storage for Cross-Computer Sync (DEFERRED - Not MVP):**
- **Status:** Cross-computer sync is deferred post-MVP; single-computer local storage is sufficient for initial development and use
- **Future Exploration:** When implementing sync, explore alternatives to monolithic SQLite file:
  - Individual JSON/YAML files per task or per day (more granular, better suited for file-based cloud sync)
  - Conflict-free Replicated Data Types (CRDTs) for eventual consistency
  - Event sourcing with append-only logs
  - Cloud-native solutions (S3, Firebase, etc.) if local-first constraint relaxes
- **Awareness:** Monolithic SQLite on cloud storage (iCloud/Google Drive) is known problematic—corruption risk, locking issues, slow sync
- **MVP Decision:** Store enrichment data locally only; revisit sync architecture when/if multi-computer use becomes actual need

**Go Language & Ecosystem:**
- **Language:** Go 1.25.4+ (latest stable)
- **Formatting:** `gofumpt` (stricter than `gofmt`)
- **Linting:** `golangci-lint` with standard rule set
- **Dependency Management:** Go modules
- **TUI Framework:** Bubbletea + Lipgloss (styling) + Bubbles (components) from Charm ecosystem

**Data Storage:**
- **Primary:** Apple Notes (user-facing tasks)
- **Enrichment:** SQLite for metadata (door feedback, blockers, categorization, learning patterns) - local storage only for MVP
- **Future Consideration:** Vector database for semantic search (deferred post-MVP unless need emerges)
- **Configuration:** YAML or TOML file for user preferences, values/goals
- **Location:** Local user directory (`~/.config/threedoors/` or similar) for MVP

**Build & Development:**
- **Build System:** Makefile (familiar, simple, no additional dependencies)
- **Commands:** `make build`, `make test`, `make lint`, `make install`
- **Development Workflow:** Direct iteration on macOS (primary platform)
- **No Docker for MVP** - native development only, containerization deferred

**Security & Privacy:**
- **No external services** for MVP (all local processing)
- **Future LLM integration** will require API key management via OS keychain
- **No logging of task content** - only metadata (counts, timestamps, categories)
- **Apple Notes access** uses standard macOS permissions (user grants on first run)

**Performance Expectations:**
- Door selection algorithm: <100ms to choose 3 tasks from up to 1000 total tasks
- Apple Notes sync: <2 seconds for typical data set (50-100 notes)
- TUI rendering: 60fps equivalent for smooth interaction (Bubbletea handles this)
- Acceptable degradation: Slower performance with 1000+ tasks is acceptable for MVP

**Deferred Technical Decisions (Post-MVP):**
- Cross-computer sync architecture (see deferred section above)
- LLM provider integration architecture (local vs. cloud, which providers)
- Additional integration adapters (Jira, Linear, Google Calendar, etc.)
- Remote access agent for Geodesic environments
- Vector database for semantic task search
- Voice interface integration

---
