---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - docs/prd/index.md (sharded PRD - 14 files, v2.0 with 9 party mode recommendations)
  - docs/architecture/index.md (sharded Architecture v2.0 - 19 files)
  - docs/prd/user-interface-design-goals.md (UX embedded in PRD)
  - docs/sprint-status-report.md (Epics 1-3 complete, 22 stories implemented)
regeneratedFrom: "PRD v2.0 + Architecture v2.0 (post-party-mode-recommendations)"
---

# ThreeDoors - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for ThreeDoors, decomposing the requirements from the PRD v2.0, UX Design, and Architecture v2.0 into implementable stories. This is a regeneration reflecting the 9 party mode recommendations integrated into the PRD and architecture.

**Implementation Status:** Epics 0-3 are COMPLETE (40 stories, 73 merged PRs). Epics 4-8, 10-12, 14-15, 18 have implemented stories but stale status markers. See gap analysis report for details.

## Requirements Inventory

### Functional Requirements

**Technical Demo Phase (COMPLETE):**
- TD1: The system shall provide a CLI/TUI interface optimized for terminal emulators (iTerm2 and similar)
- TD2: The system shall read tasks from a simple local text file (~/.threedoors/tasks.txt)
- TD3: The system shall display the Three Doors interface showing three tasks selected from the text file
- TD4: The system shall allow door selection via A/Left, W/Up, D/Right keys with no initial selection after launch or re-roll
- TD5: The system shall provide a refresh mechanism via S/Down to generate a new set of three doors
- TD6: The system shall display doors with dynamic width adjustment based on terminal size
- TD7: The system shall respond to task management keystrokes: c (complete), b (blocked), i (in-progress), e (expand), f (fork), p (procrastinate)
- TD8: The system shall embed "progress over perfection" messaging in the interface
- TD9: The system shall write completed tasks to a separate file (~/.threedoors/completed.txt) with timestamp

**Phase 2 - Apple Notes Integration (COMPLETE):**
- FR2: The system shall integrate with Apple Notes as primary task storage backend with bidirectional sync
- FR4: The system shall retrieve and display tasks from Apple Notes
- FR5: The system shall mark tasks complete, updating both app state and Apple Notes
- FR12: The system shall support bidirectional sync with Apple Notes on iPhone
- FR15: The system shall provide a health check command for Apple Notes connectivity

**Phase 3 - Enhanced Interaction & Learning (PARTIALLY COMPLETE):**
- FR3: The system shall allow task capture with optional context (what and why) through CLI/TUI ✅
- FR6: The system shall display user-defined values and goals persistently throughout sessions ✅
- FR7: The system shall provide choose-your-own-adventure interactive navigation ✅
- FR8: The system shall track daily task completion count with day-over-day comparison ✅
- FR9: The system shall prompt user once per session for improvement suggestion ✅
- FR10: The system shall embed enhanced "progress over perfection" messaging ✅
- FR16: The system shall support quick add mode for minimal-interaction task capture ✅
- FR18: The system shall allow door feedback options (Blocked, Not now, Needs breakdown, Other comment) ✅
- FR19: The system shall capture and store blocker information when task marked blocked ✅
- FR20: The system shall use door selection and feedback patterns to inform future door selection (learning) ⏳ Epic 4
- FR21: The system shall categorize tasks by type, effort level, and context for diverse door selection ⏳ Epic 4

**Phase 4 - Distribution & Packaging (COMPLETE):**
- FR22: macOS binaries code-signed with Apple Developer certificate ✅ (Story 5.1)
- FR23: Notarized with Apple's notarization service ✅ (Story 5.1)
- FR24: Installable via Homebrew tap ✅ (Story 5.1)
- FR25: DMG or pkg installer as alternative ✅ (Story 5.1)
- FR26: Automated release process ✅ (Story 5.1)

**Phase 5 - Data Layer & Enrichment:**
- FR11: The system shall maintain a local enrichment layer for metadata and cross-references ⏳ Epic 6

**Phase 6+ - Party Mode Recommendations (Accepted):**

*Obsidian Integration (P0 - #2 Integration):*
- FR27: Integrate with Obsidian vaults as task storage backend ⏳ Epic 8
- FR28: Bidirectional sync with Obsidian vault files ⏳ Epic 8
- FR29: Obsidian vault configuration via config.yaml ⏳ Epic 8
- FR30: Obsidian daily notes integration ⏳ Epic 8

*Plugin/Adapter SDK:*
- FR31: Adapter registry with runtime discovery and loading ⏳ Epic 7
- FR32: Config-driven provider selection via config.yaml ⏳ Epic 7
- FR33: Adapter developer guide and interface specification ⏳ Epic 7

*Psychology Research & Validation:*
- FR34: Document evidence base for Three Doors choice architecture ⏳ Epic 15

*LLM Task Decomposition & Agent Action Queue:*
- FR35: LLM-powered task decomposition ⏳ Epic 14
- FR36: Output to git repository for coding agents ⏳ Epic 14
- FR37: Configurable LLM backends (local and cloud) ⏳ Epic 14

*First-Run Onboarding Experience:*
- FR38: First-run welcome flow with values/goals setup ⏳ Epic 10
- FR39: Import from existing task sources during onboarding ⏳ Epic 10

*Sync Observability & Offline-First:*
- FR40: Offline-first operation with local change queue ⏳ Epic 11
- FR41: Sync status indicator in TUI per provider ⏳ Epic 11
- FR42: Conflict visualization for sync conflicts ⏳ Epic 11
- FR43: Sync log for debugging ⏳ Epic 11

*Calendar Awareness (Local-First, No OAuth):*
- FR44: Read local calendar sources only ⏳ Epic 12
- FR45: Time-contextual door selection ⏳ Epic 12

*Multi-Source Task Aggregation:*
- FR46: Unified cross-provider task pool ⏳ Epic 13
- FR47: Duplicate detection across providers ⏳ Epic 13
- FR48: Source attribution in TUI ⏳ Epic 13

*Testing Strategy:*
- FR49: Apple Notes integration E2E tests ⏳ Epic 9
- FR50: Contract tests for adapter compliance ⏳ Epic 9
- FR51: Functional E2E tests for user workflows ⏳ Epic 9

*Docker E2E & Headless TUI Testing (Party Mode):*
- FR52: Headless TUI test harness using teatest for automated interaction testing ⏳ Epic 18
- FR53: Golden file snapshot tests for TUI visual regression detection ⏳ Epic 18
- FR54: Docker-based reproducible test environment for E2E test execution ⏳ Epic 18

### Non-Functional Requirements

**Technical Demo Phase (COMPLETE):**
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
- NFR3: macOS primary platform with signed/notarized binaries
- NFR4: Local or iCloud storage (via Apple Notes), no external telemetry
- NFR5: Local application state and enrichment data (cross-computer sync deferred)
- NFR6: <500ms latency for typical operations
- NFR7: Graceful degradation when Apple Notes unavailable
- NFR8: OS keychain for credential/token storage
- NFR9: No sensitive data logging
- NFR10: Make build system
- NFR11: Clear architectural separation (core, TUI, adapters, enrichment)
- NFR12: Data integrity during external Apple Notes modification
- NFR13: <100ms response time for adapter operations (read/write/sync)
- NFR14: Offline-first operation; core functionality without network; sync queued and replayed
- NFR15: No OAuth or cloud API credentials for calendar; local sources only
- NFR16: CI coverage gates ensuring no regression below thresholds

**Code Quality & Submission Standards (Cross-Cutting):**
- NFR-CQ1: All code must pass gofumpt formatting before submission
- NFR-CQ2: All code must pass golangci-lint with zero issues before submission
- NFR-CQ3: All branches must be rebased onto upstream/main before PR creation
- NFR-CQ4: All PRs must have clean git diff --stat showing only in-scope changes
- NFR-CQ5: All fix-up commits must be squashed before PR submission

### Additional Requirements

**From Architecture v2.0:**
- Greenfield Go project (no starter template) - go mod init
- Phase 1: Two-layer architecture: TUI layer (internal/tui) + Domain layer (internal/tasks)
- Phase 2-3: Five-layer architecture: TUI, Core Domain, Adapter Layer, Sync Engine, Intelligence Layer
- MVU pattern mandatory (Bubbletea enforced Elm Architecture)
- Structured YAML data format for tasks with metadata (status, notes, timestamps)
- Five-state task lifecycle: todo → blocked → in-progress → in-review → complete
- Atomic writes for all file persistence (write-to-temp, fsync, rename)
- UUID v4 for task identification
- Constructor injection for dependency management
- TaskProvider interface for adapter pattern (established in Epic 2)
- Adapter Registry with config-driven runtime discovery (Epic 7)
- Offline-first queue pattern with async replay (Epic 11)
- Multi-source aggregation with cross-provider dedup (Epic 13)
- Intelligence layer with opt-in feature gates (Epics 12, 14)
- Ring buffer for recently-shown door tracking (default size: 10)
- Fisher-Yates shuffle for random door selection
- Apple Notes integration via AppleScript bridge (established in Epic 2)
- Unit tests for core domain logic (70%+ coverage target)
- Integration tests for backend adapters
- CI/CD via GitHub Actions

**From UX Design:**
- Three doors rendered horizontally with dynamic width adjustment
- No "Door X" labels (reduce visual clutter)
- Context-aware Esc key behavior (return to previous screen maintaining state)
- Bottom-up search results display
- Multiple navigation schemes (arrows, WASD, HJKL)
- Live substring matching for search
- Command palette (: prefix) for power-user features
- Source attribution badges for multi-provider tasks
- Sync status indicator in footer area
- Onboarding wizard with skip option at every step

### FR Coverage Map

| Requirement | Epic | Description |
|------------|------|-------------|
| (cross-cutting) | Epic 0 ✅ | Infrastructure & Process Backfill (COMPLETE) |
| TD1-TD9 | Epic 1 ✅ | Three Doors Technical Demo (COMPLETE) |
| FR2, FR4, FR5, FR12, FR15 | Epic 2 ✅ | Apple Notes Integration (COMPLETE) |
| FR3, FR6-FR10, FR16, FR18, FR19 | Epic 3 ✅ | Enhanced Interaction (COMPLETE) |
| FR20, FR21 | Epic 4 | Learning & Intelligent Door Selection |
| FR22-FR26 | Epic 5 ✅ | macOS Distribution & Packaging (COMPLETE) |
| FR11 | Epic 6 | Data Layer & Enrichment |
| FR31, FR32, FR33 | Epic 7 | Plugin/Adapter SDK & Registry |
| FR27, FR28, FR29, FR30 | Epic 8 | Obsidian Integration |
| FR49, FR50, FR51 | Epic 9 | Testing Strategy & Quality Gates |
| FR38, FR39 | Epic 10 | First-Run Onboarding |
| FR40, FR41, FR42, FR43 | Epic 11 | Sync Observability & Offline-First |
| FR44, FR45 | Epic 12 | Calendar Awareness |
| FR46, FR47, FR48 | Epic 13 | Multi-Source Aggregation |
| FR35, FR36, FR37 | Epic 14 | LLM Task Decomposition |
| FR34 | Epic 15 | Psychology Research & Validation |

## Epic List

### Epic 0: Infrastructure & Process (Backfill) ✅ COMPLETE
Retroactive stories covering CI, documentation, tooling, quality standards, and research work from 29 unstory'd PRs.
**FRs covered:** None (cross-cutting infrastructure)
**Status:** All 19 stories complete (retroactive). See `docs/analysis/pr-story-gap-analysis.md`.

### Epic 1: Three Doors Technical Demo ✅ COMPLETE
Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept.
**FRs covered:** TD1-TD9
**Status:** All 7 stories implemented and merged.

### Epic 2: Foundation & Apple Notes Integration ✅ COMPLETE
Replace text file backend with Apple Notes integration via adapter pattern.
**FRs covered:** FR2, FR4, FR5, FR12, FR15
**Status:** All 6 stories implemented and merged.

### Epic 3: Enhanced Interaction & Task Context ✅ COMPLETE
Add task capture, values/goals, feedback mechanisms, and navigation improvements.
**FRs covered:** FR3, FR6, FR7, FR8, FR9, FR10, FR16, FR18, FR19
**Status:** All 7 stories implemented and merged.

### Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging)
Refactor core architecture, harden adapters, establish test infrastructure, and resolve tech debt from rapid Epic 1-3 implementation to prepare for Epic 4+ work.
**FRs covered:** None (infrastructure/quality — enables FR20-FR51)
**Prerequisites:** Epic 3 complete ✅
**Blocks:** Epic 4 (partially), Epic 7, Epic 8, Epic 9, Epic 11

### Epic 4: Learning & Intelligent Door Selection
Use historical session metrics to analyze user patterns and adapt door selection.
**FRs covered:** FR20, FR21
**Prerequisites:** Epic 3 complete ✅, Epic 3.5 stories 3.5.5/3.5.6 complete, sufficient usage data

### Epic 5: macOS Distribution & Packaging ✅ COMPLETE
Code signing, notarization, Homebrew tap, and pkg installer.
**FRs covered:** FR22-FR26
**Status:** Story 5.1 consolidated and implemented.

### Epic 6: Data Layer & Enrichment (Optional)
SQLite enrichment database for metadata beyond what backends support.
**FRs covered:** FR11
**Prerequisites:** Epic 4 complete, proven need

### Epic 7: Plugin/Adapter SDK & Registry
Formalize adapter pattern into plugin SDK with registry and developer guide.
**FRs covered:** FR31, FR32, FR33
**Prerequisites:** Epic 2 ✅

### Epic 8: Obsidian Integration (P0 - #2 Integration)
Add Obsidian vault as second task storage backend.
**FRs covered:** FR27, FR28, FR29, FR30
**Prerequisites:** Epic 7

### Epic 9: Testing Strategy & Quality Gates
Comprehensive testing infrastructure with integration, contract, E2E tests.
**FRs covered:** FR49, FR50, FR51
**Prerequisites:** Epic 2 ✅, Epic 7

### Epic 10: First-Run Onboarding Experience
Guided welcome flow for new users.
**FRs covered:** FR38, FR39
**Prerequisites:** Epic 3 ✅

### Epic 11: Sync Observability & Offline-First
Offline-first local change queue, sync status, conflict resolution.
**FRs covered:** FR40, FR41, FR42, FR43
**Prerequisites:** Epic 2 ✅

### Epic 12: Calendar Awareness (Local-First, No OAuth)
Time-contextual door selection from local calendar sources.
**FRs covered:** FR44, FR45
**Prerequisites:** Epic 4

### Epic 13: Multi-Source Task Aggregation View
Unified cross-provider task pool with dedup and source attribution.
**FRs covered:** FR46, FR47, FR48
**Prerequisites:** Epic 7, Epic 8 or additional adapters

### Epic 14: LLM Task Decomposition & Agent Action Queue
LLM-powered task breakdown for coding agent pickup.
**FRs covered:** FR35, FR36, FR37
**Prerequisites:** Epic 3+ ✅

### Epic 15: Psychology Research & Validation
Evidence base for ThreeDoors design decisions.
**FRs covered:** FR34
**Prerequisites:** None

---

## Epic 0: Infrastructure & Process (Backfill) ✅ COMPLETE

**Epic Goal:** Retroactively track infrastructure, documentation, tooling, and process work that was performed outside of story-level planning. These backfill stories capture work from 29 merged PRs that had no backing story.

**Status:** COMPLETE — All work already shipped. Stories created retroactively for traceability.

**Origin:** PR-Story Gap Analysis (2026-03-03), see `docs/analysis/pr-story-gap-analysis.md`

### Story 0.1: BMAD Framework Setup ✅

As a developer,
I want the BMAD method framework installed and configured,
So that the project has structured agent workflows for planning and implementation.

**Status:** Done (PR #1)

**Acceptance Criteria:**
- **AC1:** BMAD slash commands, agent definitions, and task templates are installed
- **AC2:** Project documentation framework is initialized
- **AC3:** All BMAD agents are functional and invocable

### Story 0.2: Epics & Stories Breakdown ✅

As a product manager,
I want the PRD decomposed into epics and implementable stories,
So that development work is planned and trackable.

**Status:** Done (PR #6)

**Acceptance Criteria:**
- **AC1:** All functional requirements mapped to epics
- **AC2:** Each epic has stories with acceptance criteria
- **AC3:** Story dependencies documented

### Story 0.3: README Documentation ✅

As a user,
I want installation instructions, usage docs, and keybinding reference in the README,
So that I can install and use ThreeDoors without additional help.

**Status:** Done (PRs #11, #69, #71)

**Acceptance Criteria:**
- **AC1:** Installation options documented (binary, Homebrew, source)
- **AC2:** Usage instructions with keybinding reference
- **AC3:** Data directory and configuration documented
- **AC4:** Existing formatting (emojis, structure) preserved during updates

### Story 0.4: GitHub Release Automation ✅

As a developer,
I want automated GitHub Releases with compiled binaries on merge to main,
So that users can download releases without manual packaging.

**Status:** Done (PR #12)

**Acceptance Criteria:**
- **AC1:** CI creates prerelease GitHub Release on merge to main
- **AC2:** Binaries compiled for target platforms
- **AC3:** Release tagged with version from binary

### Story 0.5: CI Test Coverage Reporting ✅

As a developer,
I want test coverage reported in CI,
So that I can track coverage trends and enforce minimums.

**Status:** Done (PR #9)

**Acceptance Criteria:**
- **AC1:** CI runs tests with `-coverprofile`
- **AC2:** Coverage summary displayed in CI output
- **AC3:** No CI failures from coverage reporting itself

### Story 0.6: PRD Validation & Expansion ✅

As a product owner,
I want the PRD validated against BMAD standards and expanded with party mode recommendations,
So that the requirements are comprehensive and well-structured.

**Status:** Done (PRs #26, #34, #36)

**Acceptance Criteria:**
- **AC1:** PRD passes BMAD 13-step validation
- **AC2:** Executive summary, user journeys, and product scope sections present
- **AC3:** Party mode recommendations integrated (FR27–FR51, NFR13–NFR16)
- **AC4:** Epic 5 (macOS distribution) requirements added (FR22–FR26)

### Story 0.7: Architecture v2.0 Documentation ✅

As a developer,
I want architecture documentation updated to reflect the expanded PRD,
So that implementation decisions are aligned with requirements.

**Status:** Done (PR #38)

**Acceptance Criteria:**
- **AC1:** 5-layer architecture documented (TUI, Core, Adapter, Sync, Intelligence)
- **AC2:** All 9 party mode recommendations reflected in architecture
- **AC3:** Component diagrams and data flow updated

### Story 0.8: Epic Regeneration & Bridging Stories ✅

As a product manager,
I want epics regenerated from PRD v2.0 with bridging stories for technical debt,
So that the story backlog reflects current requirements.

**Status:** Done (PR #39)

**Acceptance Criteria:**
- **AC1:** All epics regenerated from PRD v2.0
- **AC2:** Epic 3.5 (Platform Readiness) added with 8 bridging stories
- **AC3:** Epic 4 detailed with 6 stories
- **AC4:** Total story count updated

### Story 0.9: PR Quality Standards & Checklists ✅

As a developer,
I want standardized pre-PR submission checklists and quality NFRs,
So that fix-up PRs are prevented before submission.

**Status:** Done (PRs #32, #33, #51)

**Acceptance Criteria:**
- **AC1:** Pre-PR checklist added to all story files
- **AC2:** NFR-CQ1 through NFR-CQ5 defined in PRD
- **AC3:** Quality ACs (AC-Q1–AC-Q8) documented
- **AC4:** Coding standards updated with pre-PR checklist

### Story 0.10: Sprint Status Auditing ✅

As a scrum master,
I want sprint status audited against actual merged PRs,
So that story statuses are accurate and trustworthy.

**Status:** Done (PR #37)

**Acceptance Criteria:**
- **AC1:** All epics audited against merged PRs
- **AC2:** Stale story metadata corrected
- **AC3:** Stories without dedicated .story.md files identified
- **AC4:** Sprint status report generated

### Story 0.11: AI Tooling Research ✅

As a developer,
I want AI tooling patterns researched and documented,
So that agent workflows are optimized for this project.

**Status:** Done (PR #35)

**Acceptance Criteria:**
- **AC1:** CLAUDE.md, SOUL.md, and custom skills proposed
- **AC2:** DRY analysis across documentation completed
- **AC3:** Quality root cause analysis across PRs performed

### Story 0.12: CLAUDE.md & Quality Gate Integration ✅

As a developer,
I want a project-level CLAUDE.md with Go quality rules and quality gates in all stories,
So that AI agents consistently produce idiomatic, high-quality Go code.

**Status:** Done (PRs #50, #52)

**Acceptance Criteria:**
- **AC1:** CLAUDE.md with 10 idiomatic Go rules, error handling, testing standards
- **AC2:** Quality gates (AC-Q1–AC-Q8) added to all 41 unimplemented stories
- **AC3:** Common AI mistake patterns documented

### Story 0.13: Implementation Workflow Tooling ✅

As a developer,
I want a reusable /implement-story workflow command,
So that story implementation follows a consistent 8-phase process.

**Status:** Done (PR #48)

**Acceptance Criteria:**
- **AC1:** Custom slash command created at .claude/commands/implement-story.md
- **AC2:** 8-phase workflow codified (SM → party mode → TEA → DEV → simplify → review → PR)
- **AC3:** Command is invocable and produces consistent output

### Story 0.14: Code Signing Research ✅

As a developer,
I want the state of macOS code signing investigated and documented,
So that unsigned build issues are understood and resolvable.

**Status:** Done (PR #46)

**Acceptance Criteria:**
- **AC1:** CI signing infrastructure state documented
- **AC2:** Missing configuration identified (SIGNING_ENABLED variable)
- **AC3:** Steps to enable signing documented

### Story 0.15: Mobile App Research & Planning ✅

As a product owner,
I want iPhone mobile app feasibility researched and planned,
So that mobile expansion is informed by technical analysis.

**Status:** Done (PR #47)

**Acceptance Criteria:**
- **AC1:** Framework choice evaluated (SwiftUI recommended)
- **AC2:** Go backend sharing strategy documented
- **AC3:** Epic 16 with 7 stories added to PRD

### Story 0.16: CI/Distribution Fix-ups ✅

As a developer,
I want CI secret names aligned and notarization timeouts configured correctly,
So that the release pipeline works reliably.

**Status:** Done (PRs #10, #61, #67, #76)

**Acceptance Criteria:**
- **AC1:** CI workflow secret names match repository secret names
- **AC2:** Notarization timeout set to ≥60 minutes (Apple recommendation)
- **AC3:** All code passes gofumpt before merge
- **AC4:** Remaining manual signing setup steps documented

### Story 0.17: Story 1.3 Test Backfill ✅

As a developer,
I want comprehensive TUI tests for Story 1.3,
So that the door selection and status management features have adequate test coverage.

**Status:** Done (PR #7)

**Acceptance Criteria:**
- **AC1:** ≥76 TUI tests covering door selection, status transitions, detail view
- **AC2:** ≥90% TUI test coverage achieved
- **AC3:** Tests pass in CI

### Story 0.18: Story 8.1 Quality Gate Test Backfill ✅

As a developer,
I want AC-Q6 input sanitization tests for the Obsidian adapter,
So that the quality gate requirement is verified.

**Status:** Done (PR #74)

**Acceptance Criteria:**
- **AC1:** Special characters in filenames tested
- **AC2:** HTML/quotes in task text tested
- **AC3:** Emoji content tested
- **AC4:** Escape characters tested

### Story 0.19: Headless TUI Testing Epic Planning ✅

As a product owner,
I want Epic 18 (Docker E2E & Headless TUI Testing) planned with stories and requirements,
So that TUI testing infrastructure has a clear implementation path.

**Status:** Done (PR #60)

**Acceptance Criteria:**
- **AC1:** Epic 18 added to PRD with 5 stories (18.1–18.5)
- **AC2:** FR52–FR54 added to functional requirements
- **AC3:** Distinction from Epic 9 testing scope documented

---

## Epic 1: Three Doors Technical Demo ✅ COMPLETE

**Epic Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists.

**Status:** COMPLETE — All stories implemented and merged across 34 PRs.

### Story 1.1: Project Setup & Basic Bubbletea App ✅

As a developer,
I want a working Go project with Bubbletea framework,
So that I have a foundation for building the Three Doors TUI.

**Status:** Done (PR #2)

### Story 1.2: Display Three Doors from a Task File ✅

As a developer,
I want the application to read tasks from a text file and display three of them as "doors",
So that I can see the core interface of the application.

**Status:** Done (PR #4)

### Story 1.3: Door Selection & Task Status Management ✅

As a user,
I want to select a door and update the task's status,
So that I can take action on tasks and track my progress.

**Status:** Done (PRs #5, #7)

### Story 1.3a: Quick Search & Command Palette ✅

As a user,
I want to quickly search for specific tasks and execute commands via a text input interface,
So that I can efficiently find and act on tasks without scrolling through the three doors.

**Status:** Done (PR #13)

### Story 1.5: Session Metrics Tracking ✅

As a developer validating the Three Doors concept,
I want objective session metrics collected automatically,
So that I can make a data-informed decision at the validation gate.

**Status:** Done (PR #16)

### Story 1.6: Essential Polish ✅

As a user,
I want the app to feel polished enough to use daily,
So that I enjoy the validation experience.

**Status:** Done (PR #18)

### Story 1.7: CI/CD Pipeline & Alpha Release ✅

As a developer,
I want automated builds, tests, and releases,
So that quality is maintained and releases are consistent.

**Status:** Done (PR #8)

---

## Epic 2: Foundation & Apple Notes Integration ✅ COMPLETE

**Epic Goal:** Replace text file backend with Apple Notes integration, enabling mobile task editing while maintaining Three Doors UX.

**Status:** COMPLETE — All stories implemented and merged.

### Story 2.1: Architecture Refactoring - Adapter Pattern ✅

As a developer,
I want the codebase refactored to use a TaskProvider adapter pattern,
So that multiple backends can be plugged in.

**Status:** Done (PR #20)

### Story 2.2: Apple Notes Integration Spike ✅

As a developer,
I want to evaluate Apple Notes integration approaches,
So that I can choose the best technical path.

**Status:** Done (PR #22)

### Story 2.3: Read Tasks from Apple Notes ✅

As a user,
I want my Apple Notes tasks displayed in Three Doors,
So that I can use my existing task list.

**Status:** Done (PR #17)

### Story 2.4: Write Task Updates to Apple Notes ✅

As a user,
I want task status changes reflected back in Apple Notes,
So that my tasks stay synchronized.

**Status:** Done (PR #21)

### Story 2.5: Bidirectional Sync Engine ✅

As a user,
I want changes in Apple Notes reflected in ThreeDoors and vice versa,
So that I can edit tasks from either place.

**Status:** Done (PR #15)

### Story 2.6: Health Check Command ✅

As a user,
I want to verify Apple Notes connectivity,
So that I can diagnose sync issues.

**Status:** Done (PR #19)

---

## Epic 3: Enhanced Interaction & Task Context ✅ COMPLETE

**Epic Goal:** Add task capture, values/goals display, and feedback mechanisms to improve task management workflow.

**Status:** COMPLETE — All stories implemented and merged.

### Story 3.1: Quick Add Mode ✅

As a user,
I want to add tasks with minimal friction,
So that capturing new tasks doesn't interrupt my flow.

**Status:** Done (PR #23)

### Story 3.2: Extended Task Capture with Context ✅

As a user,
I want to capture task context (what and why),
So that I remember why tasks are important.

**Status:** Done (PR #24)

### Story 3.3: Values & Goals Setup and Display ✅

As a user,
I want to see my values and goals while working,
So that I stay aligned with what matters.

**Status:** Done (PR #25)

### Story 3.4: Door Feedback Options ✅

As a user,
I want to provide feedback on why a door doesn't suit me,
So that the system can learn my preferences.

**Status:** Done (PR #27)

### Story 3.5: Daily Completion Tracking & Comparison ✅

As a user,
I want to see my daily completion count compared to yesterday,
So that I can see my progress trend.

**Status:** Done (PR #28)

### Story 3.6: Session Improvement Prompt ✅

As a user,
I want a gentle prompt for improvement at session end,
So that I continuously refine my workflow.

**Status:** Done (PR #29)

### Story 3.7: Enhanced Navigation & Messaging ✅

As a user,
I want improved navigation and "progress over perfection" messaging,
So that the app feels cohesive and encouraging.

**Status:** Done (PR #31)

---

## Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging)

**Epic Goal:** Refactor core architecture, harden adapters, establish test infrastructure, and resolve technical debt from rapid Epic 1-3 implementation. This bridging epic prepares the codebase for Epic 4+ work by establishing the architectural foundations specified in Architecture v2.0.

**Prerequisites:** Epic 3 complete ✅
**Blocks:** Epic 4 (stories 3.5.5, 3.5.6), Epic 7 (stories 3.5.1, 3.5.2, 3.5.3), Epic 9 (story 3.5.7), Epic 11 (story 3.5.4)
**Origin:** Party mode bridging discussion (2026-03-02)

### Story 3.5.1: Core Domain Extraction

As a developer,
I want `internal/tasks` split into `internal/core` (domain logic) and separate adapter packages,
So that the architecture follows the five-layer design specified in Architecture v2.0 and enables the Plugin SDK (Epic 7).

**Acceptance Criteria:**

**Given** the current `internal/tasks/` package with ~2,100 LOC across 12 files
**When** the refactoring is complete
**Then** `internal/core/` contains: TaskPool, DoorSelector, StatusManager, SessionTracker (domain logic only)
**And** `internal/adapters/textfile/` contains the YAML file adapter (extracted from FileManager)
**And** `internal/adapters/applenotes/` contains the Apple Notes adapter
**And** `internal/tui/` depends only on `internal/core/`, not on adapter implementations (dependency inversion)
**And** all existing tests pass without modification (behavior-preserving refactor)
**And** no user-facing behavior changes

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.2: TaskProvider Interface Hardening

As a developer building future integrations,
I want the TaskProvider interface formalized with Watch(), HealthCheck(), and ChangeEvent patterns,
So that the adapter SDK (Epic 7) has a stable, well-defined contract.

**Acceptance Criteria:**

**Given** the current TaskProvider interface from Epic 2
**When** hardening is complete
**Then** `TaskProvider` interface includes: Name(), Load(), Save(), Delete(), Watch(), HealthCheck() methods
**And** `ChangeEvent` struct defined with Type (Created/Updated/Deleted), TaskID, Task, Source fields
**And** contract test stubs created in `internal/adapters/contract_test.go` (placeholder for Epic 9)
**And** existing text file and Apple Notes adapters updated to implement the hardened interface
**And** interface documented with godoc comments

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.3: Config.yaml Schema & Migration Spike

As a developer,
I want a spike on config.yaml schema design and migration path,
So that Epic 7's config-driven provider selection has a validated foundation.

**Acceptance Criteria:**

**Given** the current scattered configuration (hardcoded paths, text files)
**When** the spike is complete
**Then** `docs/spikes/config-schema.md` documents: proposed config.yaml schema, provider section design, migration path from current config
**And** spike verifies zero-friction upgrade: existing users without config.yaml default to current behavior (text file provider)
**And** sample config.yaml drafted with commented provider examples
**And** spike identifies any breaking changes and mitigation strategies

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.4: Apple Notes Adapter Hardening

As a user relying on Apple Notes sync,
I want the adapter to handle errors gracefully with timeouts and retries,
So that sync is reliable before more adapters are added.

**Acceptance Criteria:**

**Given** the current Apple Notes adapter using os/exec for AppleScript
**When** hardening is complete
**Then** all AppleScript calls have configurable timeout (default: 10s)
**And** transient failures retry with exponential backoff (max 3 retries)
**And** errors are categorized: transient (retry), permanent (fail fast), configuration (user action needed)
**And** error messages are user-friendly and actionable
**And** adapter logs sync operations for debugging (respects NFR9 - no sensitive data)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.5: Baseline Regression Test Suite

As a developer preparing for Epic 4 (Learning),
I want baseline tests for the current door selection and task management behavior,
So that the learning engine (Epic 4) can be validated against known-good behavior.

**Acceptance Criteria:**

**Given** the current random door selection algorithm
**When** baseline tests are created
**Then** table-driven tests cover: random selection from pool, Fisher-Yates diversity, recently-shown ring buffer exclusion, empty/small pool edge cases
**And** status management tests cover: all valid state transitions, invalid transition rejection, completion flow
**And** task pool tests cover: load, filter by status, add, remove, update operations
**And** tests serve as regression suite when Epic 4 modifies selection algorithm
**And** all tests pass on current codebase
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.6: Session Metrics Reader Library

As a developer building Epic 4 (Learning),
I want a reusable library for reading and parsing session metrics,
So that Epic 4 stories can focus on learning logic rather than I/O.

**Acceptance Criteria:**

**Given** session metrics stored in `~/.threedoors/sessions.jsonl`
**When** the reader library is created
**Then** `internal/core/metrics/reader.go` provides: ReadAll(), ReadSince(time), ReadLast(n) methods
**And** each method returns typed `SessionMetrics` structs (not raw JSON)
**And** handles corrupted/malformed lines gracefully (skip with warning, don't fail)
**And** unit tests cover: empty file, single session, multiple sessions, corrupted lines
**And** library is dependency-free (no external packages beyond stdlib)
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources
**And** test assertions verify actual outcomes, not just absence of errors

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.7: Adapter Test Scaffolding & CI Coverage Floor

As a developer,
I want test infrastructure scaffolding and CI coverage enforcement,
So that Epic 9 (Testing Strategy) has a foundation and coverage doesn't erode.

**Acceptance Criteria:**

**Given** the current CI pipeline without coverage enforcement
**When** scaffolding is complete
**Then** test fixture directory `testdata/` created with sample data for adapter testing
**And** mock/stub helpers created in `internal/testing/` for common test patterns
**And** CI pipeline updated to measure coverage (`go test -coverprofile`) and fail if below threshold (set to current level)
**And** coverage report posted as PR comment
**And** `internal/adapters/contract_test.go` scaffolding ready for Epic 9 to fill
**And** CI runs `golangci-lint run ./...` with zero issues required (errcheck, staticcheck included)
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.8: Validation Gate Decision Documentation

As the product team,
I want the Phase 1 validation results formally documented,
So that the proceed-to-MVP decision is recorded and learnings inform Epic 4.

**Acceptance Criteria:**

**Given** Phase 1 (Technical Demo) has been used daily
**When** documentation is complete
**Then** `docs/validation-gate-results.md` documents: validation period, usage patterns, friction reduction evidence from session metrics
**And** UX lessons learned captured (what worked, what surprised, what to improve)
**And** formal "proceed to MVP" decision recorded with rationale
**And** recommendations for Epic 4 learning algorithm based on observed patterns
**And** document references actual session metrics data as evidence

**Quality Gate (AC-Q5):** All changed files directly related to this story's ACs | scope-checked ✓ | See Appendix for full BDD criteria.

---

## Epic 4: Learning & Intelligent Door Selection

**Epic Goal:** Use historical session metrics (captured in Epic 1 Story 1.5) to analyze user patterns and adapt door selection to improve task engagement and completion rates.

**Prerequisites:** Epic 3 complete ✅, sufficient usage data collected
**FRs covered:** FR20, FR21

### Story 4.1: Task Categorization & Tagging

As a user,
I want my tasks automatically categorized by type, effort, and context,
So that the system can present diverse door selections.

**Acceptance Criteria:**

**Given** a task pool with uncategorized tasks
**When** the categorization engine processes them
**Then** each task receives type (creative, administrative, technical, physical), effort (quick-win, medium, deep-work), and context (home, work, errands) labels
**And** categorization is heuristic-based (keyword matching, task text analysis) without requiring user input
**And** users can override or correct auto-categorization via `:tag` command
**And** categories are persisted in task metadata (YAML)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Categorization output must be deterministic for the same input; sorted collections where ordering matters.

### Story 4.2: Session Metrics Pattern Analysis

As a developer,
I want to analyze historical session metrics for user behavior patterns,
So that the learning engine has data to work with.

**Acceptance Criteria:**

**Given** accumulated session metrics in sessions.jsonl
**When** the pattern analyzer runs
**Then** it identifies: door position preferences (left/center/right bias), task type selection vs bypass rates, time-of-day patterns, mood-task correlation coefficients, and avoidance patterns (tasks shown 3+ times without selection)
**And** results are stored in a patterns cache file (patterns.json)
**And** analysis runs on app startup (background, non-blocking)
**And** minimum 5 sessions required before generating patterns (cold start guard)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 4.3: Mood-Aware Adaptive Door Selection

As a user,
I want door selection to consider my current mood and historical patterns,
So that I'm shown tasks that match my current capacity.

**Acceptance Criteria:**

**Given** a user has logged a mood entry (or has recent mood history)
**When** doors are selected for display
**Then** the selection algorithm weights tasks based on mood-task correlation data (e.g., "stressed" → prefer quick-wins over deep-work)
**And** the algorithm still includes diversity (not all doors match mood preference)
**And** if no mood data exists, falls back to random selection (current behavior)
**And** selection weights are configurable in a learning config section

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Adaptive selection must use seeded randomness or documented non-determinism; anti-repeat guards required; time.Now() called once per selection operation.

### Story 4.4: Avoidance Detection & User Insights

As a user,
I want to be gently informed about my avoidance patterns,
So that I can make conscious decisions about deferred tasks.

**Acceptance Criteria:**

**Given** a task has been shown in doors 5+ times without selection
**When** that task appears in doors again
**Then** a subtle indicator appears (e.g., "You've seen this task 7 times")
**And** the system does NOT nag or guilt — framing is informational
**And** a `:insights` command shows a summary of patterns ("When stressed, you avoid technical tasks")
**And** persistent avoidance (10+ bypasses) triggers a gentle prompt: "This task keeps appearing. Would you like to: [R]econsider, [B]reak down, [D]efer, [A]rchive?"

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Avoidance counts must be deterministic; bypass tracking sorted by count; time.Now() called once per session.

### Story 4.5: Goal Re-evaluation Prompts

As a user,
I want gentle prompts to reconsider goals when persistent avoidance patterns emerge,
So that my task list stays aligned with what I actually want to do.

**Acceptance Criteria:**

**Given** a pattern of avoidance for tasks related to a specific goal/value
**When** avoidance exceeds threshold (configurable, default: 3 related tasks avoided 5+ times each)
**Then** at session start, a non-blocking prompt appears: "Some [goal] tasks have been deferred repeatedly. Would you like to review your [goal] priorities?"
**And** user can dismiss with a single keypress
**And** re-evaluation prompt shown at most once per week per goal
**And** prompt links to `:goals` command for editing

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 4.6: "Better Than Yesterday" Multi-Dimensional Tracking

As a user,
I want to see progress across multiple dimensions,
So that I celebrate improvement beyond just task count.

**Acceptance Criteria:**

**Given** accumulated session history
**When** a new session starts
**Then** the greeting includes multi-dimensional comparison: tasks completed, doors opened, mood trend, avoidance reduction, and streaks
**And** comparison is day-over-day and week-over-week
**And** messaging is encouraging regardless of direction ("3 tasks today vs 5 yesterday — every door opened counts")
**And** dimensions are displayed compactly (single line or expandable)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Multi-dimensional comparisons must use consistent time base; time.Now() called once per session start; streak calculations deterministic.

---

## Epic 5: macOS Distribution & Packaging ✅ COMPLETE

**Epic Goal:** Provide a trusted, seamless installation experience on macOS.

**Status:** COMPLETE — Story 5.1 consolidated signing, notarization, Homebrew, and pkg (PR #30).

### Story 5.1: CI Code Signing, Notarization, Homebrew & pkg ✅

As a macOS user,
I want signed, notarized binaries installable via Homebrew or pkg,
So that Gatekeeper allows execution without security warnings.

**Status:** Done (PR #30)

---

## Epic 6: Data Layer & Enrichment (Optional)

**Epic Goal:** Add enrichment storage layer for metadata that cannot live in source systems.

**Prerequisites:** Epic 4 complete, proven need for enrichment beyond what backends support
**FRs covered:** FR11

### Story 6.1: SQLite Enrichment Database Setup

As a developer,
I want a local SQLite database for enrichment metadata,
So that cross-reference tracking and learning patterns have persistent storage.

**Acceptance Criteria:**

**Given** the application starts
**When** enrichment storage is needed (learning patterns, cross-references)
**Then** a SQLite database is created at `~/.threedoors/enrichment.db`
**And** schema includes tables for: task enrichment (categories, learning data), cross-references (task links across providers), and user preferences
**And** database is created lazily (only when first enrichment write occurs)
**And** migrations are version-tracked for schema evolution

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q7 (Error Handling):** All database operations must check error returns including db.Close(); use fmt.Errorf("context: %w", err) wrapping; no silently discarded errors.
**Atomic Writes:** Database writes must use transactions; file-based operations use write-to-tmp, sync, rename pattern.

### Story 6.2: Cross-Reference Tracking

As a user with multiple task sources,
I want tasks linked across providers,
So that related items are connected regardless of source.

**Acceptance Criteria:**

**Given** a task exists in multiple providers (or is related to tasks in other providers)
**When** the user links them via `:link` command or automatic detection
**Then** cross-references are stored in enrichment.db
**And** linked tasks show a "linked" indicator in task detail view
**And** navigating to linked tasks is supported from detail view

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q7 (Error Handling):** All database operations must check error returns; cross-reference writes use transactions.

---

## Epic 7: Plugin/Adapter SDK & Registry

**Epic Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide.

**Prerequisites:** Epic 2 ✅ (adapter pattern established)
**FRs covered:** FR31, FR32, FR33

### Story 7.1: Adapter Registry & Runtime Discovery

As a developer building integrations,
I want a formal adapter registry that discovers and loads task providers at runtime,
So that new integrations can be added without modifying core application code.

**Acceptance Criteria:**

**Given** the application starts
**When** the adapter registry initializes
**Then** it discovers all registered TaskProvider implementations
**And** adapters register via `registry.Register(name, factory)` pattern
**And** failed adapter initialization logs warning and continues with other adapters
**And** registry exposes `ListProviders()`, `GetProvider(name)`, and `ActiveProviders()` methods
**And** existing text file and Apple Notes adapters are migrated to registry pattern

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 7.2: Config-Driven Provider Selection

As a user with multiple task sources,
I want to configure active backends via `~/.threedoors/config.yaml`,
So that I can choose which task providers are active without code changes.

**Acceptance Criteria:**

**Given** a config.yaml with `providers:` section
**When** the application starts
**Then** only configured providers are loaded and activated
**And** provider-specific settings (paths, credentials) passed to adapter factory
**And** missing config.yaml falls back to text file provider (backward compatible)
**And** sample config.yaml generated on first run with commented examples

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 7.3: Adapter Developer Guide & Contract Tests

As an integration developer,
I want a clear guide and contract test suite for building adapters,
So that I can create new task provider integrations with confidence.

**Acceptance Criteria:**

**Given** a developer wants to build a new adapter
**When** they follow the developer guide
**Then** `docs/adapter-developer-guide.md` covers: TaskProvider interface spec, registration, config schema, testing
**And** contract test suite in `internal/adapters/contract_test.go` validates any TaskProvider
**And** tests cover: CRUD operations, error handling, concurrent access, interface compliance
**And** contract test suite is reusable (adapters import and run against their implementation)
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

---

## Epic 8: Obsidian Integration (P0 - #2 Integration)

**Epic Goal:** Add Obsidian vault as second task storage backend. Local-first Markdown integration with bidirectional sync.

**Prerequisites:** Epic 7 (adapter SDK)
**FRs covered:** FR27, FR28, FR29, FR30

### Story 8.1: Obsidian Vault Reader/Writer Adapter

As a user who manages tasks in Obsidian,
I want ThreeDoors to read and write tasks from my Obsidian vault,
So that I can use Three Doors with my existing Obsidian workflow.

**Acceptance Criteria:**

**Given** a configured Obsidian vault path
**When** the adapter loads
**Then** `ObsidianAdapter` implements `TaskProvider` interface
**And** reads Markdown files from configured vault folder
**And** parses task items using Obsidian checkbox syntax (`- [ ]`, `- [x]`, `- [/]`)
**And** supports Obsidian task metadata (due dates, tags, priorities)
**And** writes task status changes back using atomic file operations
**And** passes adapter contract test suite

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q6 (Input Sanitization):** File paths and task content from Obsidian vault must be sanitized; test cases with special characters in filenames and task text.
**Atomic Writes:** File write operations must use write-to-tmp, sync, rename pattern per coding-standards.md.

### Story 8.2: Obsidian Bidirectional Sync

As an Obsidian user,
I want changes made in Obsidian reflected in ThreeDoors and vice versa,
So that my tasks stay in sync regardless of where I edit them.

**Acceptance Criteria:**

**Given** a configured Obsidian vault
**When** files are modified externally
**Then** file watcher detects changes and re-parses affected files
**And** task pool updates without full reload
**And** concurrent edit handling uses last-write-wins with conflict logging
**And** sync latency under 2 seconds

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 8.3: Obsidian Vault Configuration

As a user,
I want to configure my Obsidian vault path and structure via config.yaml,
So that ThreeDoors integrates with my specific vault.

**Acceptance Criteria:**

**Given** config.yaml with `obsidian:` section
**When** the application starts
**Then** vault path is validated (exists, readable, writable)
**And** invalid vault path produces clear error and fallback to other providers
**And** supports configurable task folder and file pattern (glob)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 8.4: Obsidian Daily Note Integration

As an Obsidian daily notes user,
I want ThreeDoors to read/write tasks from my daily notes,
So that tasks captured in daily notes appear in Three Doors.

**Acceptance Criteria:**

**Given** daily notes enabled in config
**When** the adapter loads
**Then** reads tasks from today's daily note file
**And** quick-add tasks can be appended under configurable heading
**And** supports common date formats (`YYYY-MM-DD.md`, etc.)
**And** missing daily note handled gracefully

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q6 (Input Sanitization):** Daily note file paths and heading content must be sanitized; test cases with special characters in date formats and heading names.

---

## Epic 9: Testing Strategy & Quality Gates

**Epic Goal:** Establish comprehensive testing infrastructure ensuring reliability as the adapter ecosystem grows.

**Prerequisites:** Epic 2 ✅, Epic 7
**FRs covered:** FR49, FR50, FR51

### Story 9.1: Apple Notes Integration E2E Tests

As a developer,
I want end-to-end tests for Apple Notes integration,
So that regressions in the sync pipeline are caught automatically.

**Acceptance Criteria:**

**Given** a test environment with mock AppleScript responses
**When** E2E tests run
**Then** tests validate: note creation, task read, task update, bidirectional sync, error handling
**And** tests cover: connectivity failure, partial sync, concurrent modification
**And** test fixtures in `testdata/applenotes/` for reproducible scenarios
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 9.2: Contract Tests for Adapter Compliance

As an adapter developer,
I want a reusable contract test suite,
So that all adapters behave consistently.

**Acceptance Criteria:**

**Given** a TaskProvider implementation
**When** contract tests run
**Then** tests validate: CRUD operations, error handling, concurrent access, interface compliance
**And** each adapter runs the contract suite in its own test file
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 9.3: Performance Benchmarks

As a developer,
I want automated performance benchmarks,
So that <100ms NFR is validated and regressions caught.

**Acceptance Criteria:**

**Given** benchmark suite using Go's `testing.B`
**When** benchmarks run
**Then** adapter read, write, sync, and door selection are benchmarked
**And** results compared against <100ms threshold (NFR13)
**And** CI runs benchmarks on every PR

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 9.4: Functional E2E Tests

As a developer,
I want functional E2E tests covering full user workflows,
So that the complete user experience is validated.

**Acceptance Criteria:**

**Given** a test environment
**When** E2E tests run
**Then** tests exercise: launch → view doors → select door → manage task → exit
**And** session metrics generation verified
**And** search, command palette, mood tracking workflows covered
**And** uses Bubbletea's `teatest` package for TUI testing
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 9.5: CI Coverage Gates

As the team,
I want CI coverage gates,
So that code quality doesn't regress.

**Acceptance Criteria:**

**Given** CI pipeline
**When** a PR is submitted
**Then** coverage measurement runs (`go test -coverprofile`)
**And** PRs reducing coverage below threshold are blocked
**And** coverage report posted as PR comment
**And** CI runs full pre-PR verification checklist (gofumpt, golangci-lint with errcheck/staticcheck, go test, scope check) per coding-standards.md

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

---

## Epic 10: First-Run Onboarding Experience

**Epic Goal:** Provide a guided welcome flow for new users.

**Prerequisites:** Epic 3 ✅
**FRs covered:** FR38, FR39

### Story 10.1: Welcome Flow & Three Doors Explanation

As a new user,
I want a guided welcome on first launch,
So that I understand the Three Doors concept.

**Acceptance Criteria:**

**Given** first-run detected (no `~/.threedoors/` directory)
**When** the application launches
**Then** welcome screen with branding and concept explanation displays
**And** interactive key bindings walkthrough lets user try keys
**And** skip option available at every step
**And** onboarding state persisted (`onboarding_complete: true` in config)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 10.2: Values/Goals Setup & Task Import

As a new user,
I want to set up values/goals and import tasks during onboarding,
So that the tool is immediately useful.

**Acceptance Criteria:**

**Given** onboarding flow reaches setup step
**When** user enters values/goals
**Then** values persist to config.yaml
**And** import detection for common task sources (text, Markdown)
**And** import preview shows tasks before importing
**And** step is skippable; manual import via `:import` command later

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

---

## Epic 11: Sync Observability & Offline-First

**Epic Goal:** Ensure robust offline-first operation with sync visibility and conflict resolution.

**Prerequisites:** Epic 2 ✅
**FRs covered:** FR40, FR41, FR42, FR43

### Story 11.1: Offline-First Local Change Queue

As a user working without connectivity,
I want all changes queued locally and replayed when available,
So that I never lose work.

**Acceptance Criteria:**

**Given** a provider is unavailable
**When** the user makes changes
**Then** changes are written to WAL (`~/.threedoors/sync-queue.jsonl`)
**And** queue replays in order when connectivity restored
**And** failed replays retry with exponential backoff
**And** core functionality unaffected by sync state

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q7 (Error Handling):** WAL file operations must check all error returns including f.Close(); replay failures must be logged with context via %w wrapping.
**Atomic Writes:** WAL writes must use write-to-tmp, sync, rename pattern per coding-standards.md.

### Story 11.2: Sync Status Indicator

As a user,
I want to see sync status per provider in the TUI,
So that I know my changes are synchronized.

**Acceptance Criteria:**

**Given** multiple providers configured
**When** the TUI displays
**Then** status bar shows per-provider state (✓ synced, ↻ syncing, ⏳ pending, ✗ error)
**And** updates in real-time
**And** minimal screen real estate

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 11.3: Conflict Visualization & Sync Log

As a user encountering sync conflicts,
I want to see and resolve them,
So that I trust the sync system.

**Acceptance Criteria:**

**Given** a sync conflict is detected
**When** the user views the conflict
**Then** local vs remote versions shown side-by-side
**And** resolution options: keep local, keep remote, keep both
**And** `:synclog` command shows chronological operations
**And** sync log rotated at 1MB

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q7 (Error Handling):** Sync log file operations must check all error returns; conflict resolution must propagate errors with context.

---

## Epic 12: Calendar Awareness (Local-First, No OAuth)

**Epic Goal:** Add time-contextual door selection from local calendar sources only.

**Prerequisites:** Epic 4
**FRs covered:** FR44, FR45

### Story 12.1: Local Calendar Source Reader

As a user,
I want ThreeDoors to read my local calendar,
So that it understands my available time.

**Acceptance Criteria:**

**Given** calendar sources configured in config.yaml
**When** the calendar reader initializes
**Then** macOS Calendar.app events read via AppleScript (no OAuth)
**And** .ics file parser for configured paths
**And** CalDAV cache reader from `~/Library/Calendars/`
**And** graceful fallback when sources unavailable

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q6 (Input Sanitization):** AppleScript calls for Calendar.app must escape all user/event data; test cases with special characters in event titles and calendar names.

### Story 12.2: Time-Contextual Door Selection

As a user with calendar awareness,
I want doors to suggest time-appropriate tasks,
So that I'm not shown deep-work when I have a meeting in 15 minutes.

**Acceptance Criteria:**

**Given** calendar events available
**When** doors are selected
**Then** selection considers next event time
**And** short blocks prefer quick tasks
**And** no calendar data = standard selection
**And** time context shown in TUI ("Next event in 45 min")

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Time-contextual selection must call time.Now() once per selection operation; task ordering deterministic for same time window.

---

## Epic 13: Multi-Source Task Aggregation View

**Epic Goal:** Unified cross-provider task pool with dedup and source attribution.

**Prerequisites:** Epic 7, Epic 8 or additional adapters
**FRs covered:** FR46, FR47, FR48

### Story 13.1: Cross-Provider Task Pool Aggregation

As a user with multiple task sources,
I want all tasks merged into a single pool,
So that I see everything without switching sources.

**Acceptance Criteria:**

**Given** multiple providers configured
**When** the task pool loads
**Then** tasks collected from all active providers
**And** unified pool used for door selection, search, all views
**And** provider failures isolated (one failing doesn't block others)
**And** task pool maintains provider origin metadata

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 13.2: Duplicate Detection & Source Attribution

As a user with overlapping sources,
I want duplicates flagged and sources shown,
So that I don't work on the same task twice.

**Acceptance Criteria:**

**Given** tasks from multiple providers
**When** aggregation runs
**Then** fuzzy text matching identifies potential duplicates
**And** duplicates shown with indicator ("Possible duplicate")
**And** user can merge or dismiss duplicate flags
**And** source badges show in door view, search, and detail view

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q8 (Determinism):** Duplicate detection ordering must be deterministic; fuzzy match results sorted by score.

---

## Epic 14: LLM Task Decomposition & Agent Action Queue

**Epic Goal:** Enable LLM-powered task decomposition for coding agent pickup.

**Prerequisites:** Epic 3+ ✅
**FRs covered:** FR35, FR36, FR37

### Story 14.1: LLM Task Decomposition Spike

As a developer,
I want to spike on LLM task decomposition feasibility,
So that we understand the approach before full implementation.

**Acceptance Criteria:**

**Given** a spike investigation
**When** completed
**Then** `docs/spikes/llm-decomposition.md` covers prompt engineering, output schema, git automation
**And** tests multiple providers (local: Ollama; cloud: Claude API)
**And** agent handoff protocol drafted
**And** recommendation: build vs wait, local vs cloud, effort estimate

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 14.2: Agent Action Queue Integration

As a developer using ThreeDoors with coding agents,
I want decomposed tasks output to git repos,
So that task decomposition flows into automated implementation.

**Acceptance Criteria:**

**Given** a user initiates task decomposition
**When** the LLM processes the task
**Then** output follows BMAD story file structure
**And** stories written to configurable repo path
**And** git operations: branch creation, commit, optional PR creation
**And** configurable LLM backend via config.yaml

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.
**AC-Q6 (Input Sanitization):** Git operations (branch names, commit messages) must sanitize user-provided task content; shell command construction must escape all interpolated values; test cases with special characters.

---

## Epic 15: Psychology Research & Validation

**Epic Goal:** Build evidence base for ThreeDoors design decisions.

**Prerequisites:** None (can run in parallel)
**FRs covered:** FR34

### Story 15.1: Choice Architecture Literature Review

As the product team,
I want a literature review on the Three Doors choice architecture,
So that design decisions are grounded in behavioral science.

**Acceptance Criteria:**

**Given** research task
**When** review completed
**Then** `docs/research/choice-architecture.md` covers choice overload, paradox of choice, decision fatigue
**And** specific evidence for why 3 options
**And** comparable systems analysis
**And** practical recommendations

**Quality Gate (AC-Q5):** All changed files directly related to this story's ACs | scope-checked ✓ | See Appendix for full BDD criteria.

### Story 15.2: Mood-Task Correlation & Procrastination Research

As the product team,
I want research on mood-task correlation and procrastination interventions,
So that Epic 4's learning algorithm is evidence-informed.

**Acceptance Criteria:**

**Given** research task
**When** review completed
**Then** `docs/research/mood-correlation.md` and `docs/research/procrastination.md` produced
**And** evidence assessment for "progress over perfection"
**And** actionable recommendations for Epic 4
**And** bibliography with accessible references

**Quality Gate (AC-Q5):** All changed files directly related to this story's ACs | scope-checked ✓ | See Appendix for full BDD criteria.

---

## Epic 18: Docker E2E & Headless TUI Testing Infrastructure

**Epic Goal:** Establish reproducible, automated end-to-end testing for the TUI application using Docker containers for environment isolation and Bubbletea's `teatest` package for headless interaction testing — eliminating manual TUI testing as the sole E2E validation method.

**Prerequisites:** Epic 3 ✅, Epic 9 (Stories 9.4, 9.5)
**FRs covered:** FR49, FR51 (extends Epic 9's scope with concrete implementation approach)
**Origin:** Party mode testing infrastructure discussion (2026-03-02). Party mode consensus identified two critical gaps: (1) no reproducible test environment — tests depend on developer machine state, and (2) TUI testing is entirely manual — 10% of the test pyramid has zero automation.

**Why a separate epic from Epic 9:** Epic 9 defines *what* to test (Apple Notes E2E, contract tests, performance benchmarks, functional E2E, CI gates). This epic defines *how* to test the TUI layer specifically — the Docker infrastructure and headless testing tooling that Epic 9 Story 9.4 depends on but doesn't specify.

### Story 18.1: Headless TUI Test Harness with teatest

As a developer,
I want a headless TUI test harness using Bubbletea's `teatest` package,
So that I can write automated tests that interact with the full TUI without a real terminal.

**Acceptance Criteria:**

**Given** the `teatest` package (`github.com/charmbracelet/x/exp/teatest`) is added to `go.mod`
**When** a test creates a `teatest.NewTestModel` with the root TUI model
**Then** the test can send key messages (`tea.KeyMsg`) to simulate user input
**And** the test can retrieve `FinalOutput` and `FinalModel` for assertions
**And** `lipgloss.SetColorProfile(termenv.Ascii)` is enforced for deterministic output
**And** test helper `NewTestApp(t *testing.T, opts ...TestOption) *teatest.TestModel` is provided in `internal/tui/testhelpers_test.go`
**And** helper accepts options: `WithTermSize(w, h int)`, `WithTaskFile(path string)`, `WithConfig(cfg Config)`
**And** at least 3 smoke tests demonstrate the harness: app launch, door display, and quit
**And** tests use stdlib `testing` package only (no testify); table-driven for >2 cases; t.Helper() in helpers; t.Cleanup() for resources

**Technical Notes:**
- `teatest` creates a pseudo-TTY internally — no Docker needed for basic headless tests
- Fixed terminal size (`teatest.WithInitialTermSize(80, 24)`) ensures reproducible layout
- The harness wraps the existing `tui.NewModel()` constructor — no TUI code changes needed
- Test fixtures use `t.TempDir()` for task files and config

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 18.2: Golden File Snapshot Tests for TUI Views

As a developer,
I want golden file tests that capture expected TUI output,
So that visual regressions in the Three Doors interface are caught automatically.

**Acceptance Criteria:**

**Given** the headless test harness from Story 18.1
**When** golden file tests run
**Then** `FinalOutput` is compared against `.golden` files in `internal/tui/testdata/`
**And** golden files cover: main doors view (3 tasks), empty state (0 tasks), too-few-tasks state (1-2 tasks), door selection highlight, status bar with values/goals, help overlay
**And** `.gitattributes` includes `*.golden -text` to prevent line-ending conversion
**And** golden files are regenerated via `go test ./internal/tui/... -update`
**And** CI runs golden file comparison (without `-update`) to catch regressions
**And** at least 6 golden file scenarios covering the views listed above

**Technical Notes:**
- Golden files are the teatest-recommended approach for View() output testing
- ASCII color profile ensures golden files are portable across terminals
- Golden file diffs in CI provide clear visual indication of what changed
- Keep golden files focused on layout structure, not exact styling (ASCII mode helps)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 18.3: Input Sequence Replay Tests for User Workflows

As a developer,
I want automated tests that replay user input sequences against the TUI,
So that complete user workflows (launch → select → manage → exit) are validated end-to-end.

**Acceptance Criteria:**

**Given** the headless test harness from Story 18.1
**When** workflow replay tests run
**Then** tests exercise these user journeys via `tm.Send(tea.KeyMsg{...})` sequences:
  1. Launch → view 3 doors → select door (A key) → verify selection
  2. Launch → re-roll doors (S key) → verify new doors displayed
  3. Launch → select door → complete task (C key) → verify task removed from pool
  4. Launch → select door → mark blocked (B key) → enter blocker text → verify
  5. Launch → quick add (N key) → type task → submit → verify task in pool
  6. Launch → open help (?) → verify help overlay → close help (Esc)
**And** each workflow asserts on `FinalModel` state (not just output text)
**And** workflows use `teatest.WaitFor` for intermediate state assertions where needed
**And** test task files are created via `t.TempDir()` with known task sets
**And** tests use stdlib `testing` package only; table-driven for workflow variants

**Technical Notes:**
- Input replays test the full Bubbletea Update() → View() cycle
- Model assertions are more stable than output assertions for workflow correctness
- `WaitFor` with timeout prevents tests from hanging on unexpected state
- Each workflow should complete in <2s (set `teatest.WithFinalTimeout(2*time.Second)`)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 18.4: Docker Test Environment for Reproducible E2E

As a developer,
I want a Docker-based test environment,
So that E2E tests run identically on any machine and in CI regardless of host OS or installed tools.

**Acceptance Criteria:**

**Given** a `Dockerfile.test` in the repository root
**When** `make test-docker` is run
**Then** a Docker image is built with: Go toolchain, gofumpt, golangci-lint, and all test dependencies
**And** the full test suite (`go test ./... -v -count=1`) runs inside the container
**And** golden file tests and workflow replay tests from Stories 18.2-18.3 pass inside Docker
**And** test results and coverage report are written to a mounted volume (`./test-results/`)
**And** `docker-compose.test.yml` defines the test service with: build context, volume mounts for source and results, environment variables for test configuration
**And** the container uses a non-root user for test execution
**And** image build time is <2 minutes on a cold build (use multi-stage build with cached Go modules)
**And** `make test-docker` exits with the same exit code as the test suite

**Technical Notes:**
- Multi-stage Dockerfile: stage 1 installs tools + caches `go mod download`, stage 2 copies source and runs tests
- Docker provides the pseudo-TTY that teatest needs — no special terminal setup required
- Volume mount for source code enables fast iteration without rebuilding the image
- CI can use the same Docker image, ensuring dev/CI environment parity
- No macOS-specific dependencies in Docker (Apple Notes tests are excluded via build tags)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 18.5: CI Integration for Docker E2E Tests

As the team,
I want Docker E2E tests running automatically in CI,
So that TUI regressions are caught on every PR without relying on manual testing.

**Acceptance Criteria:**

**Given** the Docker test environment from Story 18.4
**When** a PR is submitted
**Then** a new CI job `test-docker-e2e` runs the Docker test suite
**And** the job uses `docker-compose.test.yml` to run tests
**And** test results are uploaded as CI artifacts
**And** golden file diffs (if any) are included in the CI output for review
**And** the job runs in parallel with existing `quality-gate` and `build` jobs
**And** the job completes in <5 minutes (Docker layer caching via GitHub Actions cache)
**And** the job fails the PR check if any E2E test fails
**And** `.github/workflows/ci.yml` is updated with the new job

**Technical Notes:**
- GitHub Actions supports Docker natively — use `docker compose run` in a step
- Cache Docker layers via `actions/cache` with `docker buildx` for fast rebuilds
- Separate job (not step) allows parallel execution with existing quality gates
- Apple Notes integration tests remain macOS-only; Docker E2E covers everything else

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

---

## Appendix: PR-Analysis-Derived Quality Acceptance Criteria

> **Source:** Systematic analysis of all 49 PRs (#1–#49) in arcaven/ThreeDoors, examining every delta between initial PR submission and final merge. These ACs are derived from recurring defect patterns and MUST be included in all future stories.

### Issue Categorization Summary

Analysis of 49 PRs found 18 PRs (37%) required post-submission changes. The remaining 31 PRs (63%) merged cleanly on first submission. Issue breakdown by category:

| Category | PRs Affected | % of Issues | Root Cause |
|----------|-------------|-------------|------------|
| **Lint/static analysis** (errcheck + staticcheck) | #16, #42, #43, #44, #45 | 23% | Code not linted before push |
| **Logic/correctness bugs** | #14, #17, #18, #19, #44 | 16% | Insufficient edge-case thinking in ACs |
| **Merge conflicts** | #3, #5, #19, #23, #42 | 16% | Stale branches, no pre-PR rebase |
| **gofumpt formatting** | #9, #23, #24, #42 | 13% | Formatter not run before push |
| **Missing test coverage** | #5, #7, #16, #20 | 13% | No coverage gate in story ACs |
| **Silently ignored errors** | #16, #17 | 6% | No errcheck enforcement in ACs |
| **Duplicate/wasted work** | #14, #49 | 6% | Parallel agents implementing same story |
| **Security vulnerabilities** | #17 | 3% | No input sanitization AC |
| **Scope creep** | #5 | 3% | No scope-limiting AC |

### Mandatory Quality ACs for All Future Stories

Every story in Epics 3.5–18 MUST include the following acceptance criteria in addition to feature-specific ACs. These are NON-NEGOTIABLE and derived from empirical PR failure data. Each story references these gates via a compact **Quality Gate** line; this appendix provides the authoritative BDD definitions.

#### AC-Q1: Formatting Gate (PRs #9, #23, #24, #42)

```
GIVEN code changes are ready for PR
WHEN `gofumpt -l .` is executed from the repository root
THEN zero files are listed (all files are properly formatted)
```

#### AC-Q2: Full Lint Gate (PRs #16, #42, #43, #44, #45)

```
GIVEN code changes are ready for PR
WHEN `golangci-lint run ./...` is executed
THEN zero issues are reported
AND specifically: no errcheck violations (all error return values checked, including f.Close(), os.Remove(), os.WriteFile())
AND specifically: no staticcheck QF1012 violations (never use WriteString(fmt.Sprintf(...)), always use fmt.Fprintf())
AND specifically: no staticcheck S1009 violations (no redundant nil checks before len())
AND specifically: no staticcheck S1011 violations (use append(slice, other...) not loops)
```

#### AC-Q3: Test Coverage Gate (PRs #5, #7, #16, #20)

```
GIVEN code changes are ready for PR
WHEN `go test ./...` is executed
THEN all tests pass
AND new code paths have corresponding test cases
AND no existing test files are deleted without equivalent replacement coverage
AND test assertions verify actual outcomes (not just "no error")
```

#### AC-Q4: Rebase Gate (PRs #3, #5, #19, #23, #42)

```
GIVEN code changes are ready for PR
WHEN the branch is compared against upstream/main
THEN the branch is rebased onto the latest upstream/main with zero conflicts
AND `gofumpt -l .` still produces zero output after rebase (rebase can introduce formatting drift)
```

#### AC-Q5: Scope Gate (PR #5)

```
GIVEN code changes are ready for PR
WHEN `git diff --stat` is reviewed
THEN all changed files are directly related to the story's acceptance criteria
AND no unrelated directories or configuration files are included
```

#### AC-Q6: Input Sanitization Gate (PR #17)

```
GIVEN the story involves constructing dynamic commands, scripts, or queries (AppleScript, SQL, shell, etc.)
WHEN user-provided or external data is interpolated into the command
THEN all interpolated values are properly escaped/sanitized for the target language
AND injection test cases are included for special characters (quotes, backslashes, semicolons)
```

#### AC-Q7: Error Handling Gate (PRs #16, #17)

```
GIVEN code changes include function calls that return errors
WHEN reviewing the code diff
THEN no error return values are silently discarded (assigned to `_` or ignored)
AND deferred Close() calls on writable files use error-checking patterns
AND error messages include context via fmt.Errorf("context: %w", err) wrapping
```

#### AC-Q8: Determinism Gate (PRs #14, #18)

```
GIVEN code changes involve ordering, randomization, or time-dependent behavior
WHEN the same inputs are provided
THEN outputs are deterministic (sorted collections, seeded randomness, or documented non-determinism)
AND randomized outputs have anti-repeat guards (no consecutive identical selections)
AND time.Now() is called once per logical operation, not inside loops
```

### Per-Story Defect Tracing

The following maps each affected story to the specific PR issues it produced:

| Story | PR(s) | Issues Found | Missing AC That Would Have Prevented It |
|-------|-------|-------------|----------------------------------------|
| 1.1 | #2 | 26 latent lint issues (discovered in PR #8) | AC-Q2 (lint gate) |
| 1.2 | #4 | Latent lint issues | AC-Q2 (lint gate) |
| 1.3 | #3→#5, #7 | Out-of-order impl, merge conflicts, deleted 324 test lines, scope creep (agents/ dir) | AC-Q3 (test gate), AC-Q4 (rebase gate), AC-Q5 (scope gate) |
| 1.3a | #14 | Non-deterministic ordering, state mutation bug, duplicate of #13 | AC-Q8 (determinism gate) |
| 1.5 | #16 | 3 CI failures: errcheck, staticcheck S1009, Makefile error swallowing | AC-Q2 (lint gate), AC-Q7 (error gate) |
| 1.6 | #18 | Consecutive greeting repeats | AC-Q8 (determinism gate) |
| 1.7 | #8, #9, #10 | CI itself introduced; PR #9 merged with gofumpt failure → PR #10 hotfix | AC-Q1 (formatting gate) |
| 2.1 | #20 | Missing provider tests, weak assertions, %s vs %q in errors | AC-Q3 (test gate), AC-Q7 (error gate) |
| 2.3 | #17 | AppleScript injection, silently ignored error, time consistency bug | AC-Q6 (input sanitization), AC-Q7 (error gate), AC-Q8 (determinism gate) |
| 2.6 | #19 | Stale view state, wrong test target (file vs dir), 2 rounds of merge conflicts | AC-Q3 (test gate), AC-Q4 (rebase gate) |
| 3.1 | #23 | gofumpt after rebase, merge conflict | AC-Q1 (formatting gate), AC-Q4 (rebase gate) |
| 3.2 | #24 | gofumpt formatting failure | AC-Q1 (formatting gate) |
| 4.2 | #43 | 8 errcheck violations, 3 CI failures | AC-Q2 (lint gate) |
| 4.3 | #44 | staticcheck QF1012 + S1009, logic bugs (duplicate task, case-sensitive mood) | AC-Q2 (lint gate) |
| 4.4 | #45, #49 | staticcheck S1011 + QF1012, duplicate PR from parallel agent | AC-Q2 (lint gate) |
| 4.5 | #42 | 4 CI failures, 5-file merge conflict, gofumpt + errcheck + QF1012 (fixed incrementally) | AC-Q1, AC-Q2, AC-Q4 (all gates) |
