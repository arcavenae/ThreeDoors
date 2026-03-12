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

**Implementation Status:** Epics 1-15, 3.5, 17-28, 32-41, 43, 45, 48-49, 52, 55 are COMPLETE. Epic 29 is 3/4 (29.3 In Review). Epic 0 is partial (12/19). Epic 16 is ICEBOX. Epic 42 (4/5), Epic 44 (6/7), Epic 46 (1/4), Epic 51 (5/11), Epic 54 (2/5) IN PROGRESS. Epics 30-31, 47, 50, 53, 58-59 NOT STARTED or IN PROGRESS. 590+ merged PRs total. Last audit: 2026-03-12.

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

*MCP/LLM Integration Server:*
- FR81: MCP server binary with stdio and SSE transports for LLM client connectivity ⏳ Epic 24
- FR82: Read-only task resources and structured query tools via MCP protocol ⏳ Epic 24
- FR83: Security middleware with rate limiting, audit logging, input validation ⏳ Epic 24
- FR84: Proposal/approval pattern for LLM-suggested task enrichments ⏳ Epic 24
- FR85: TUI proposal review view for approving/rejecting LLM suggestions ⏳ Epic 24
- FR86: Pattern mining and mood-execution analytics via MCP ⏳ Epic 24
- FR87: Task relationship graphs and cross-provider dependency mapping ⏳ Epic 24
- FR88: MCP prompt templates and advanced interaction tools (prioritization, workload, what-if) ⏳ Epic 24

*Docker E2E & Headless TUI Testing (Party Mode):*
- FR52: Headless TUI test harness using teatest for automated interaction testing ✅ Epic 18 (Story 18.1, PR #64)
- FR53: Golden file snapshot tests for TUI visual regression detection ✅ Epic 18 (Story 18.2, PR #86)
- FR54: Docker-based reproducible test environment for E2E test execution ✅ Epic 18 (Stories 18.4 PR #104, 18.5 PR #107)

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
| (cross-cutting) | Epic 0 | Infrastructure & Process Backfill (12/19 complete) |
| TD1-TD9 | Epic 1 ✅ | Three Doors Technical Demo (COMPLETE) |
| FR2, FR4, FR5, FR12, FR15 | Epic 2 ✅ | Apple Notes Integration (COMPLETE) |
| FR3, FR6-FR10, FR16, FR18, FR19 | Epic 3 ✅ | Enhanced Interaction (COMPLETE) |
| FR20, FR21 | Epic 4 ✅ | Learning & Intelligent Door Selection (COMPLETE) |
| FR22-FR26 | Epic 5 ✅ | macOS Distribution & Packaging (COMPLETE) |
| FR11 | Epic 6 ✅ | Data Layer & Enrichment (COMPLETE — 2/2 stories, optional epic) |
| FR31, FR32, FR33 | Epic 7 ✅ | Plugin/Adapter SDK & Registry (COMPLETE) |
| FR27, FR28, FR29, FR30 | Epic 8 ✅ | Obsidian Integration (COMPLETE) |
| FR49, FR50, FR51 | Epic 9 ✅ | Testing Strategy & Quality Gates (COMPLETE) |
| FR38, FR39 | Epic 10 ✅ | First-Run Onboarding (COMPLETE) |
| FR40, FR41, FR42, FR43 | Epic 11 ✅ | Sync Observability & Offline-First (COMPLETE) |
| FR44, FR45 | Epic 12 ✅ | Calendar Awareness (COMPLETE) |
| FR46, FR47, FR48 | Epic 13 ✅ | Multi-Source Aggregation (COMPLETE) |
| FR35, FR36, FR37 | Epic 14 ✅ | LLM Task Decomposition (COMPLETE) |
| FR34 | Epic 15 ✅ | Psychology Research & Validation (COMPLETE) |
| (mobile-specific) | Epic 16 | iPhone Mobile App (NOT STARTED) |
| FR55-FR62 | Epic 17 ✅ | Door Theme System (COMPLETE) |
| FR63-FR66 | Epic 19 ✅ | Jira Integration (COMPLETE) |
| FR67-FR69 | Epic 20 ✅ | Apple Reminders Integration (COMPLETE) |
| FR70-FR72 | Epic 21 ✅ | Sync Protocol Hardening (COMPLETE) |
| FR73-FR80 | Epic 22 ✅ | Self-Driving Development Pipeline (COMPLETE) |
| FR81-FR88 | Epic 24 ✅ | MCP/LLM Integration Server (COMPLETE) |
| FR89-FR92 | Epic 25 | Todoist Integration (COMPLETE) |
| FR93-FR96 | Epic 26 ✅ | GitHub Issues Integration (COMPLETE) |
| FR97-FR103 | Epic 27 | Daily Planning Mode (COMPLETE) |
| FR104-FR107 | Epic 28 | Snooze/Defer (NOT STARTED) |
| FR108-FR111 | Epic 29 | Task Dependencies (NOT STARTED) |
| FR116-FR119 | Epic 30 | Linear Integration (NOT STARTED) |
| FR120-FR126 | Epic 31 | Expand/Fork Key Implementations (NOT STARTED) |
| FR127-FR131 | Epic 32 | Undo Task Completion (COMPLETE) |
| FR132-FR137 | Epic 33 | Seasonal Door Theme Variants (COMPLETE) |
| FR148-FR151 | Epic 36 | Door Selection Feedback (COMPLETE) |

## Epic List

### Epic 0: Infrastructure & Process (Backfill)
Retroactive stories covering CI, documentation, tooling, quality standards, and research work from 29 unstory'd PRs. Now also includes forward-looking infrastructure improvements and test coverage hardening from TEA audit (R-001).
**FRs covered:** None (cross-cutting infrastructure)
**Status:** 12 of 22 stories complete. Stories 0.29, 0.50, 0.51, 0.52, 0.53 not started.

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

### Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging) ✅ COMPLETE
Refactor core architecture, harden adapters, establish test infrastructure, and resolve tech debt from rapid Epic 1-3 implementation to prepare for Epic 4+ work.
**FRs covered:** None (infrastructure/quality — enables FR20-FR51)
**Prerequisites:** Epic 3 complete ✅
**Status:** All 8 stories complete (PRs #90-#97).

### Epic 4: Learning & Intelligent Door Selection ✅ COMPLETE
Use historical session metrics to analyze user patterns and adapt door selection.
**FRs covered:** FR20, FR21
**Prerequisites:** Epic 3 complete ✅, Epic 3.5 stories 3.5.5/3.5.6 complete ✅
**Status:** All 6 stories complete (PRs #40, #42-#45, #82).

### Epic 5: macOS Distribution & Packaging ✅ COMPLETE
Code signing, notarization, Homebrew tap, and pkg installer.
**FRs covered:** FR22-FR26
**Status:** Story 5.1 consolidated and implemented (PR #30).

### Epic 6: Data Layer & Enrichment (Optional) ✅ COMPLETE
SQLite enrichment database for metadata beyond what backends support.
**FRs covered:** FR11
**Status:** All 2 stories complete (PRs #53, #56). Note: PR #53 titled "Story 5.1" but implements Epic 6 Story 6.1.

### Epic 7: Plugin/Adapter SDK & Registry ✅ COMPLETE
Formalize adapter pattern into plugin SDK with registry and developer guide.
**FRs covered:** FR31, FR32, FR33
**Prerequisites:** Epic 2 ✅
**Status:** All 3 stories complete (PRs #68, #70, #72).

### Epic 8: Obsidian Integration (P0 - #2 Integration) ✅ COMPLETE
Add Obsidian vault as second task storage backend.
**FRs covered:** FR27, FR28, FR29, FR30
**Prerequisites:** Epic 7 ✅
**Status:** All 4 stories complete (PRs #73, #75, #77, #79).

### Epic 9: Testing Strategy & Quality Gates ✅ COMPLETE
Comprehensive testing infrastructure with integration, contract, E2E tests.
**FRs covered:** FR49, FR50, FR51
**Prerequisites:** Epic 2 ✅, Epic 7 ✅
**Status:** All 5 stories complete (PRs #83, #89, #142, #103, #102).

### Epic 10: First-Run Onboarding Experience ✅ COMPLETE
Guided welcome flow for new users.
**FRs covered:** FR38, FR39
**Prerequisites:** Epic 3 ✅
**Status:** All 2 stories complete (PRs #55, #59).

### Epic 11: Sync Observability & Offline-First ✅ COMPLETE
Offline-first local change queue, sync status, conflict resolution.
**FRs covered:** FR40, FR41, FR42, FR43
**Prerequisites:** Epic 2 ✅
**Status:** All 3 stories complete (PRs #62, #66, #85).

### Epic 12: Calendar Awareness (Local-First, No OAuth) ✅ COMPLETE
Time-contextual door selection from local calendar sources.
**FRs covered:** FR44, FR45
**Prerequisites:** Epic 4 ✅
**Status:** All 2 stories complete (PRs #65, #81).

### Epic 13: Multi-Source Task Aggregation View ✅ COMPLETE
Unified cross-provider task pool with dedup and source attribution.
**FRs covered:** FR46, FR47, FR48
**Prerequisites:** Epic 7 ✅, Epic 8 ✅
**Status:** All 2 stories complete (PRs #84, #143).

### Epic 14: LLM Task Decomposition & Agent Action Queue ✅ COMPLETE
LLM-powered task breakdown for coding agent pickup.
**FRs covered:** FR35, FR36, FR37
**Prerequisites:** Epic 3+ ✅
**Status:** All 2 stories complete (PRs #63, #87).

### Epic 15: Psychology Research & Validation ✅ COMPLETE
Evidence base for ThreeDoors design decisions.
**FRs covered:** FR34
**Prerequisites:** None
**Status:** All 2 stories complete (PRs #54, #58).

### Epic 16: iPhone Mobile App (SwiftUI) — NOT STARTED
Native SwiftUI iPhone app with Three Doors card carousel.
**FRs covered:** Mobile-specific (not yet in PRD FRs)
**Prerequisites:** Epic 2 ✅
**Status:** Not Started. 7 stories planned (16.1-16.7). See `docs/prd/epic-details.md`.

### Epic 17: Door Theme System ✅ COMPLETE
Visually distinct themed doors with user-selectable themes.
**FRs covered:** FR55-FR62
**Prerequisites:** Epic 3 ✅, Epic 10 ✅
**Status:** All 6 stories complete (PRs #119, #120, #121, #123, #124, #122).

### Epic 19: Jira Integration ✅ COMPLETE
Jira as a task source with read-only adapter and bidirectional sync.
**FRs covered:** FR63-FR66
**Prerequisites:** Epic 7 ✅, Epic 11 ✅, Epic 13 ✅
**Status:** All 4 stories complete (PRs #132, #138, #150, #153).

### Epic 20: Apple Reminders Integration ✅ COMPLETE
Apple Reminders as a task source with full CRUD support.
**FRs covered:** FR67-FR69
**Prerequisites:** Epic 7 ✅
**Status:** All 4 stories complete (PRs #137, #148, #155, #158).

### Epic 21: Sync Protocol Hardening ✅ COMPLETE
Background sync scheduling, circuit breakers, and cross-provider identity mapping.
**FRs covered:** FR70-FR72
**Prerequisites:** Epic 11 ✅, Epic 13 ✅
**Status:** All 4 stories complete (PRs #139, #132, #151, #157).

### Epic 23: CLI Interface ✅ COMPLETE
Complete non-TUI CLI interface for ThreeDoors serving both human power users and LLM agents.
**FRs covered:** FR97-FR131
**Prerequisites:** None
**Status:** All 11 stories complete (PRs #161-#192, #225). Includes bug fix Story 23.11 (PR #225).

### Epic 24: MCP/LLM Integration Server ✅ COMPLETE
Expose ThreeDoors task management to LLMs via Model Context Protocol. Read-only queries, controlled enrichment proposals, analytics mining, and relationship graphs.
**FRs covered:** FR81-FR88
**Prerequisites:** Epic 13 ✅ (Multi-Source Aggregation), Epic 6 ✅ (Enrichment DB)
**Status:** All 8 stories complete (PRs #164-#196). Research at `../../_bmad-output/planning-artifacts/llm-integration-mcp.md`.

### Epic 25: Todoist Integration ✅ COMPLETE
Todoist as a task source with thin HTTP client against REST API v1, read-only adapter, bidirectional sync.
**FRs covered:** FR89-FR92
**Prerequisites:** Epic 7 ✅ (Adapter SDK), Epic 13 ✅ (Multi-Source Aggregation), Epic 21 ✅ (Sync Protocol Hardening)
**Status:** All 4 stories complete (PRs #308, #321, plus Stories 25.3 & 25.4). Research at `../../_bmad-output/planning-artifacts/task-source-expansion-research.md`.

### Epic 26: GitHub Issues Integration ✅ COMPLETE
GitHub Issues as a task source for developer workflows using the official go-github SDK. Label-based priority/status conventions.
**FRs covered:** FR93-FR96
**Prerequisites:** Epic 7 ✅ (Adapter SDK), Epic 13 ✅ (Multi-Source Aggregation), Epic 21 ✅ (Sync Protocol Hardening)
**Status:** All 4 stories complete (PRs #201-#205). Research at `../../_bmad-output/planning-artifacts/task-source-expansion-research.md`.

### Epic 34: SOUL.md + Custom Development Skills ✅ COMPLETE
Project philosophy document, custom Claude Code slash commands, story template updates, and retroactive spec alignment.
**FRs covered:** FR148-FR151 (project tooling)
**Prerequisites:** None
**Status:** All 4 stories complete (PRs #222, #224, #228, #230).

### Epic 35: Door Visual Appearance — Door-Like Proportions ✅ COMPLETE
Redesign all door themes to visually read as actual doors with portrait orientation, panel dividers, handles, and thresholds.
**FRs covered:** FR138-FR147
**Prerequisites:** Epic 17 ✅ (Door Theme System)
**Status:** All 7 stories complete (PRs #226, #229, #234, #236, #237, #238, #239).

### Epic 33: Seasonal Door Theme Variants — COMPLETE
Time-based seasonal theme variants that auto-switch based on the current date, extending the Door Theme System with visual variety.
**FRs covered:** FR132-FR137
**Prerequisites:** Epic 17 ✅ (Door Theme System)
**Status:** Not Started. 4 stories planned (33.1-33.4). Research at `../../_bmad-output/planning-artifacts/door-themes-analyst-review.md`.

---

## Epic 0: Infrastructure & Process (Backfill)

**Epic Goal:** Retroactively track infrastructure, documentation, tooling, and process work that was performed outside of story-level planning. These backfill stories capture work from 29 merged PRs that had no backing story. Now also includes forward-looking infrastructure improvements and test coverage hardening from the TEA audit (R-001).

**Status:** 12 of 22 stories complete. Stories 0.29, 0.50, 0.51, 0.52, 0.53 not started.

**Origin:** PR-Story Gap Analysis (2026-03-03), see `../../_bmad-output/planning-artifacts/pr-story-gap-analysis.md`

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

### Story 0.20: CI Churn Reduction — Branch Protection & Merge Queue Optimization

As a development team running multiple parallel agents,
I want CI to run efficiently without cascading reruns,
So that PRs merge quickly without wasting 5-10x CI runs per PR.

**Status:** Done (PR #260)

**Acceptance Criteria:**
- **AC1:** Branch protection "Require branches to be up to date" DISABLED (Phase 1)
- **AC2:** Branch protection still requires CI green and no force push (Phase 1)
- **AC3:** ADR documenting rationale written (Phase 1)
- **AC4:** merge-queue and pr-shepherd agents updated (Phase 1)
- **AC5:** CI workflows use path-based triggers for docs-only PRs (Phase 2)
- **AC6:** Evaluate and implement GitHub Native Merge Queue if feasible (Phase 3)

### Story 0.36: CI Circuit Breaker — Post-Merge Main Branch Monitoring

As the merge-queue agent merging PRs without requiring up-to-date branches,
I need to detect push-to-main CI failures immediately after each merge,
So that I can halt further merges before cascading breakage onto a red main.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** Merge-queue agent checks push-to-main CI status after each merge via `gh run list`
- **AC2:** Emergency mode entered on main CI failure: halt merges, message supervisor, label PR
- **AC3:** Automatic recovery when subsequent push-to-main CI succeeds
- **AC4:** Timeout handling: 10-minute timeout does not block merges
- **AC5:** Agent prompt updated with explicit post-merge CI check workflow

### Story 0.37: CI Efficiency Metrics — Track Runs Per Merged PR

As a project maintainer running parallel CI workflows,
I want to track CI efficiency metrics (runs per merged PR, churn ratio),
So that I can measure the impact of CI churn reduction and know when to reconsider deferred options.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** `scripts/ci-metrics.sh` computes total runs, merged PRs, runs-per-PR ratio, docs-skip count, main-failure count
- **AC2:** Script uses `gh` CLI only — no external dependencies
- **AC3:** Human-readable summary and JSON output modes
- **AC4:** Baseline comparison against pre-optimization metrics (5-10 runs/PR)
- **AC5:** ADR-0030 re-entry gate warning when main CI failures exceed 3/week

### Story 0.44: Scoped Label Migration — Rename and Create GitHub Labels

As the ThreeDoors project maintainer,
I want all GitHub labels migrated to scoped `.` separator format with the finalized 27-label taxonomy,
So that agents and humans have consistent, queryable, namespaced labels for issue and PR management.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** All 11 colon-separated labels renamed to `.` format (triage, priority, scope scopes)
- **AC2:** All 8 unscoped labels renamed to scoped equivalents (bug→type.bug, etc.)
- **AC3:** 7 new labels created (status.blocked, status.stale, status.do-not-merge, agent.envoy, agent.worker, contrib.good-first-issue, contrib.help-wanted)
- **AC4:** `multiclaude` and `ux` labels deleted
- **AC5:** All label colors match finalized scheme
- **AC6:** `gh label list` shows exactly 28 labels (27 scoped + dependencies)
- **AC7:** No issues or PRs lost label associations during migration

### Story 0.45: Agent Definition Updates for Scoped Labels

As the ThreeDoors agent infrastructure maintainer,
I want all agent definition files updated to reference the new scoped label names,
So that agents use the correct label names after the migration (Story 0.44).

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** `agents/merge-queue.md` updated with scoped label names and `status.do-not-merge` merge gate
- **AC2:** `agents/envoy.md` updated with scoped label names and authority documentation
- **AC3:** `agents/pr-shepherd.md` updated with scoped label names
- **AC4:** No behavioral changes introduced — only label name references updated

### Story 0.46: Label Authority & Triage Flow Documentation

As the ThreeDoors project maintainer,
I want a documented label authority matrix and triage flow,
So that all agents follow consistent labeling practices and the triage process is unambiguous.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** Label authority matrix documenting who can set/remove each label scope
- **AC2:** End-to-end triage flow documented with label state transitions
- **AC3:** Fast-track shortcut flow documented
- **AC4:** Mutual exclusivity rules documented per scope
- **AC5:** Convention enforcement documented (agents remove old labels before applying new)

### Story 0.21: Homebrew Public Distribution — Custom Tap, CI Hardening, and homebrew-core Submission

As the ThreeDoors maintainer,
I want to distribute ThreeDoors via Homebrew with a clear path from custom tap to homebrew-core,
So that users can install ThreeDoors with a single `brew install` command.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** MIT LICENSE file added and GoReleaser config created (Phase 1)
- **AC2:** `arcaven/homebrew-threedoors` tap repository created (Phase 1)
- **AC3:** GoReleaser GitHub Actions workflow triggers on semver tags (Phase 1)
- **AC4:** `brew tap arcaven/threedoors && brew install threedoors` works (Phase 1)
- **AC5:** CI runs `brew audit`, `brew install --build-from-source`, `brew test` (Phase 2)
- **AC6:** Cosign signing and SLSA provenance enabled for releases (Phase 2)
- **AC7:** Source-build formula submitted and accepted to homebrew-core (Phase 3)

### Story 0.22: Fix Homebrew Strict Audit Failures

As the ThreeDoors maintainer,
I want the Homebrew tap CI to pass `brew audit --strict` and `brew style` for all formulas,
So that formula updates don't break CI and we maintain homebrew-core readiness.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** Formula template uses `if Hardware::CPU.arm?` / `else` instead of `on_arm`/`on_intel`
- **AC2:** Formula `desc` does not start with the formula name
- **AC3:** Formula `test do` block does not pass redundant `0` to `shell_output`
- **AC4:** `brew audit --strict` passes for all formulas in tap CI
- **AC5:** `brew style` passes for all formulas in tap CI
- **AC6:** Formula update automation generates compliant formulas

### Story 0.24: Renovate + Dependabot Automated Dependency Management

As the ThreeDoors maintainer,
I want automated dependency management via Renovate (Go deps) and Dependabot (GitHub Actions),
So that dependencies stay up to date with security as the top priority and minimal manual effort.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** `renovate.json` configured for Go module dependency management with OSV vulnerability scanning
- **AC2:** Renovate groups patch updates weekly, charmbracelet ecosystem together, security updates individually
- **AC3:** Auto-merge enabled for patches and non-breaking minors; majors require human review
- **AC4:** `.github/dependabot.yml` configured for GitHub Actions version pinning (monthly grouped)
- **AC5:** Go modules ecosystem excluded from Dependabot (Renovate handles this)
- **AC6:** Merge-queue agent accepts `dependencies` labeled PRs as in-scope infrastructure
- **AC7:** Renovate schedule avoids active development hours (weekdays 6-8 AM UTC)

### Story 0.28: Issue Tracker File Structure, Authority Configuration & Initial Content

As the ThreeDoors envoy agent,
I want a local issue tracker file (`docs/issue-tracker.md`) with proper structure, authority tier configuration, and initial content,
So that I can track issue lifecycle, detect duplicates, and route issues based on reporter authority.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** `docs/issue-tracker.md` exists with authority tiers header, Open Issues table (8 columns), and Recently Resolved table
- **AC2:** Authority tiers configured: `arcaven` as owner, empty contributors list
- **AC3:** Issue status lifecycle: `open` → `triaged` → `story-created` → `pr-open` → `resolved`
- **AC4:** Current open GitHub issues populated in tracker
- **AC5:** Recently closed issues (last 50, 90-day window) populated
- **AC6:** SOUL.md Alignment Reference section with three-category classification and common misalignment patterns

### Story 0.29: Envoy Operations Guide & Integration Documentation

As the ThreeDoors development team,
I want an envoy operations guide documenting patrol workflows, cross-agent protocols, staleness thresholds, and communication patterns,
So that the envoy agent can operate consistently per team consensus.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** `docs/envoy-operations.md` exists with patrol cycle workflow, cross-agent protocols, triage authority matrix
- **AC2:** Patrol cycle documented as 7-step workflow with PR-to-issue linkage patterns
- **AC3:** Cross-agent communication protocols documented for supervisor, merge-queue, pr-shepherd, workers
- **AC4:** Three-tier authority routing rules table (owner, contributor, community)
- **AC5:** Staleness thresholds documented (14d/30d/21d) with escalation templates
- **AC6:** Reporter communication milestone templates for all 5 stages
- **AC7:** Duplicate detection and direction alignment handling documented

### Story 0.32: Help Display UX — Dedicated Help View

As a ThreeDoors user,
I want `:help` to show a persistent, readable, categorized help screen,
So that I can learn keybindings and commands without content disappearing or running off-screen.

**Status:** Not Started

**Acceptance Criteria:**
- **AC1:** New `ViewHelp` view mode with `HelpView` struct (Update/View/SetWidth)
- **AC2:** `:help` transitions to persistent ViewHelp (not FlashMsg)
- **AC3:** Two-column width-aware layout, works in 80-column terminals
- **AC4:** Categorized sections: Navigation, Task Actions, Commands, Search
- **AC5:** Scrollable via j/k, PgUp/PgDn; dismissed via Esc/q
- **AC6:** `?` global keybinding opens help from any non-text-input view
- **AC7:** Unit tests, golden file test, race detector passes

### Story 0.34: Fix 'q' Key in Sub-Views — Go Back Instead of Quit

As a user navigating a sub-view (dashboard, health, synclog, etc.),
I want pressing 'q' to return me to the doors view,
So that 'q' means "close what I'm looking at" — quit at root, back in sub-views.

**Status:** Ready

**Acceptance Criteria:**
- **AC1:** Universal quit handler (main_model.go:910-913) removed
- **AC2:** Sub-views (ViewInsights, ViewHealth, ViewSyncLog, ViewNextSteps, ViewAvoidancePrompt) treat 'q' as go-back
- **AC3:** Doors view still quits on 'q' (existing handler unchanged)
- **AC4:** Text input views unchanged ('q' = text input)
- **AC5:** Keybindings updated to show 'q: back' in sub-views
- **AC6:** Tests updated, race detector passes

### Story 0.51: TUI View Rendering Benchmarks

As a developer,
I want benchmarks for TUI `View()` rendering in complex views,
So that I can detect responsiveness regressions before they affect users.

**Status:** Not Started | **Priority:** P2

**Acceptance Criteria:**
- **AC1:** Benchmark functions exist for `DoorsView.View()`, `DashboardView.View()`, `StatsView.View()`, and `SourcesView.View()`
- **AC2:** Each benchmark uses realistic model state (populated task pool, active theme, multiple sources)
- **AC3:** Benchmarks run with `go test -bench=. ./internal/tui/...` and produce stable results
- **AC4:** Baseline results captured in `internal/tui/testdata/benchmarks-baseline.txt`
- **AC5:** All benchmarks pass with `-race` flag
- **AC6:** No new dependencies — stdlib `testing.B` only

### Story 0.52: Multi-Adapter Integration Tests

As a developer,
I want integration tests that exercise sync conflict resolution across real (simulated) adapter pairs,
So that I can be confident multi-source scenarios work correctly end-to-end.

**Status:** Not Started | **Priority:** P2

**Acceptance Criteria:**
- **AC1:** Integration test exercising two adapters syncing the same task pool with conflicting edits
- **AC2:** Test covers last-writer-wins conflict resolution
- **AC3:** Test covers orphaned task detection when a task is deleted from one source
- **AC4:** Test covers field-level conflict (title changed in A, status changed in B)
- **AC5:** Tests use mock adapters built from `TaskProvider` interface
- **AC6:** Tests are table-driven with named scenarios
- **AC7:** All tests pass with `-race` flag

### Story 0.53: Docker E2E Scenario Expansion

As a developer,
I want the Docker E2E test suite to cover all primary user workflows,
So that I can be confident the full application works end-to-end in a clean environment.

**Status:** Not Started | **Priority:** P2

**Acceptance Criteria:**
- **AC1:** Audit of existing Docker E2E scenarios documented
- **AC2:** Missing scenarios added for: task completion, blocking, daily planning, source connection
- **AC3:** Docker E2E tests pass locally via `docker compose -f docker-compose.test.yml up`
- **AC4:** Each new scenario has descriptive name and clear pass/fail criteria
- **AC5:** No flaky tests — deterministic timing with teatest `WaitFor`
- **AC6:** CI integration unchanged (push-only per Story 55.1)

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

### Story 1.3a (originally 1.4): Quick Search & Command Palette ✅

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

## Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging) ✅ COMPLETE

**Epic Goal:** Refactor core architecture, harden adapters, establish test infrastructure, and resolve technical debt from rapid Epic 1-3 implementation. This bridging epic prepares the codebase for Epic 4+ work by establishing the architectural foundations specified in Architecture v2.0.

**Prerequisites:** Epic 3 complete ✅
**Blocks:** Epic 4 (stories 3.5.5, 3.5.6), Epic 7 (stories 3.5.1, 3.5.2, 3.5.3), Epic 9 (story 3.5.7), Epic 11 (story 3.5.4)
**Origin:** Party mode bridging discussion (2026-03-02)
**Status:** COMPLETE — All 8 stories implemented and merged (PRs #90-#97).

### Story 3.5.1: Core Domain Extraction ✅

As a developer,
I want `internal/tasks` split into `internal/core` (domain logic) and separate adapter packages,
So that the architecture follows the five-layer design specified in Architecture v2.0 and enables the Plugin SDK (Epic 7).

**Status:** Done (PR #90)

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

### Story 3.5.2: TaskProvider Interface Hardening ✅

**Status:** Done (PR #91)

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

### Story 3.5.3: Config.yaml Schema & Migration Spike ✅

**Status:** Done (PR #92)

As a developer,
I want a spike on config.yaml schema design and migration path,
So that Epic 7's config-driven provider selection has a validated foundation.

**Acceptance Criteria:**

**Given** the current scattered configuration (hardcoded paths, text files)
**When** the spike is complete
**Then** `../../_bmad-output/planning-artifacts/config-schema.md` documents: proposed config.yaml schema, provider section design, migration path from current config
**And** spike verifies zero-friction upgrade: existing users without config.yaml default to current behavior (text file provider)
**And** sample config.yaml drafted with commented provider examples
**And** spike identifies any breaking changes and mitigation strategies

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 3.5.4: Apple Notes Adapter Hardening ✅

**Status:** Done (PR #93)

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

### Story 3.5.5: Baseline Regression Test Suite ✅

**Status:** Done (PR #94)

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

### Story 3.5.6: Session Metrics Reader Library ✅

**Status:** Done (PR #95)

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

### Story 3.5.7: Adapter Test Scaffolding & CI Coverage Floor ✅

**Status:** Done (PR #96)

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

### Story 3.5.8: Validation Gate Decision Documentation ✅

**Status:** Done (PR #97)

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

## Epic 4: Learning & Intelligent Door Selection ✅ COMPLETE

**Epic Goal:** Use historical session metrics (captured in Epic 1 Story 1.5) to analyze user patterns and adapt door selection to improve task engagement and completion rates.

**Prerequisites:** Epic 3 complete ✅, sufficient usage data collected
**FRs covered:** FR20, FR21
**Status:** COMPLETE — All 6 stories implemented and merged (PRs #40, #42-#45, #82).

### Story 4.1: Task Categorization & Tagging ✅

**Status:** Done (PR #40)

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

### Story 4.2: Session Metrics Pattern Analysis ✅

**Status:** Done (PR #43)

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

### Story 4.3: Mood-Aware Adaptive Door Selection ✅

**Status:** Done (PR #44)

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

### Story 4.4: Avoidance Detection & User Insights ✅

**Status:** Done (PR #45)

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

### Story 4.5: Goal Re-evaluation Prompts ✅

**Status:** Done (PR #42)

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

### Story 4.6: "Better Than Yesterday" Multi-Dimensional Tracking ✅

**Status:** Done (PR #82)

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

## Epic 6: Data Layer & Enrichment (Optional) ✅ COMPLETE

**Epic Goal:** Add enrichment storage layer for metadata that cannot live in source systems.

**Prerequisites:** Epic 4 complete, proven need for enrichment beyond what backends support
**FRs covered:** FR11
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #53, #56). Note: PR #53 was titled "Story 5.1" but implements Epic 6 Story 6.1 (SQLite Enrichment).

### Story 6.1: SQLite Enrichment Database Setup ✅

**Status:** Done (PR #53)

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

### Story 6.2: Cross-Reference Tracking ✅

**Status:** Done (PR #56)

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

## Epic 7: Plugin/Adapter SDK & Registry ✅ COMPLETE

**Epic Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide.

**Prerequisites:** Epic 2 ✅ (adapter pattern established)
**FRs covered:** FR31, FR32, FR33
**Status:** COMPLETE — All 3 stories implemented and merged (PRs #68, #70, #72).

### Story 7.1: Adapter Registry & Runtime Discovery ✅

**Status:** Done (PR #68)

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

### Story 7.2: Config-Driven Provider Selection ✅

**Status:** Done (PR #70)

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

### Story 7.3: Adapter Developer Guide & Contract Tests ✅

**Status:** Done (PR #72)

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

## Epic 8: Obsidian Integration (P0 - #2 Integration) ✅ COMPLETE

**Epic Goal:** Add Obsidian vault as second task storage backend. Local-first Markdown integration with bidirectional sync.

**Prerequisites:** Epic 7 ✅ (adapter SDK)
**FRs covered:** FR27, FR28, FR29, FR30
**Status:** COMPLETE — All 4 stories implemented and merged (PRs #73, #75, #77, #79).

### Story 8.1: Obsidian Vault Reader/Writer Adapter ✅

**Status:** Done (PR #73)

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

### Story 8.2: Obsidian Bidirectional Sync ✅

**Status:** Done (PR #75, combined with Story 8.3)

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

### Story 8.3: Obsidian Vault Configuration ✅

**Status:** Done (PR #75, combined with Story 8.2)

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

### Story 8.4: Obsidian Daily Note Integration ✅

**Status:** Done (PRs #77, #79)

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

## Epic 9: Testing Strategy & Quality Gates ✅ COMPLETE

**Epic Goal:** Establish comprehensive testing infrastructure ensuring reliability as the adapter ecosystem grows.

**Prerequisites:** Epic 2 ✅, Epic 7 ✅
**FRs covered:** FR49, FR50, FR51
**Status:** COMPLETE — All 5 stories implemented and merged (PRs #83, #89, #142, #103, #102).

### Story 9.1: Apple Notes Integration E2E Tests ✅

**Status:** Done (PR #83)

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

### Story 9.2: Contract Tests for Adapter Compliance ✅

**Status:** Done (PR #89)

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

### Story 9.3: Performance Benchmarks ✅

**Status:** Done (PR #142)

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

### Story 9.4: Functional E2E Tests ✅

**Status:** Done (PR #103)

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

### Story 9.5: CI Coverage Gates ✅

**Status:** Done (PR #102)

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

## Epic 10: First-Run Onboarding Experience ✅ COMPLETE

**Epic Goal:** Provide a guided welcome flow for new users.

**Prerequisites:** Epic 3 ✅
**FRs covered:** FR38, FR39
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #55, #59).

### Story 10.1: Welcome Flow & Three Doors Explanation ✅

**Status:** Done (PR #55)

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

### Story 10.2: Values/Goals Setup & Task Import ✅

**Status:** Done (PR #59)

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

## Epic 11: Sync Observability & Offline-First ✅ COMPLETE

**Epic Goal:** Ensure robust offline-first operation with sync visibility and conflict resolution.

**Prerequisites:** Epic 2 ✅
**FRs covered:** FR40, FR41, FR42, FR43
**Status:** COMPLETE — All 3 stories implemented and merged (PRs #62, #66, #85).

### Story 11.1: Offline-First Local Change Queue ✅

**Status:** Done (PR #62)

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

### Story 11.2: Sync Status Indicator ✅

**Status:** Done (PR #66)

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

### Story 11.3: Conflict Visualization & Sync Log ✅

**Status:** Done (PR #85)

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

## Epic 12: Calendar Awareness (Local-First, No OAuth) ✅ COMPLETE

**Epic Goal:** Add time-contextual door selection from local calendar sources only.

**Prerequisites:** Epic 4 ✅
**FRs covered:** FR44, FR45
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #65, #81).

### Story 12.1: Local Calendar Source Reader ✅

**Status:** Done (PR #65)

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

### Story 12.2: Time-Contextual Door Selection ✅

**Status:** Done (PR #81)

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

## Epic 13: Multi-Source Task Aggregation View ✅ COMPLETE

**Epic Goal:** Unified cross-provider task pool with dedup and source attribution.

**Prerequisites:** Epic 7 ✅, Epic 8 ✅
**FRs covered:** FR46, FR47, FR48
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #84, #143).

### Story 13.1: Cross-Provider Task Pool Aggregation ✅

**Status:** Done (PR #84)

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

### Story 13.2: Duplicate Detection & Source Attribution ✅

**Status:** Done (PR #143)

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

## Epic 14: LLM Task Decomposition & Agent Action Queue ✅ COMPLETE

**Epic Goal:** Enable LLM-powered task decomposition for coding agent pickup.

**Prerequisites:** Epic 3+ ✅
**FRs covered:** FR35, FR36, FR37
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #63, #87).

### Story 14.1: LLM Task Decomposition Spike ✅

**Status:** Done (PR #63)

As a developer,
I want to spike on LLM task decomposition feasibility,
So that we understand the approach before full implementation.

**Acceptance Criteria:**

**Given** a spike investigation
**When** completed
**Then** `../../_bmad-output/planning-artifacts/llm-decomposition.md` covers prompt engineering, output schema, git automation
**And** tests multiple providers (local: Ollama; cloud: Claude API)
**And** agent handoff protocol drafted
**And** recommendation: build vs wait, local vs cloud, effort estimate

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓ | See Appendix for full BDD criteria and pre-PR checklist.

### Story 14.2: Agent Action Queue Integration ✅

**Status:** Done (PR #87)

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

## Epic 15: Psychology Research & Validation ✅ COMPLETE

**Epic Goal:** Build evidence base for ThreeDoors design decisions.

**Prerequisites:** None (can run in parallel)
**FRs covered:** FR34
**Status:** COMPLETE — All 2 stories implemented and merged (PRs #54, #58).

### Story 15.1: Choice Architecture Literature Review ✅

**Status:** Done (PR #54)

As the product team,
I want a literature review on the Three Doors choice architecture,
So that design decisions are grounded in behavioral science.

**Acceptance Criteria:**

**Given** research task
**When** review completed
**Then** `../../_bmad-output/planning-artifacts/choice-architecture.md` covers choice overload, paradox of choice, decision fatigue
**And** specific evidence for why 3 options
**And** comparable systems analysis
**And** practical recommendations

**Quality Gate (AC-Q5):** All changed files directly related to this story's ACs | scope-checked ✓ | See Appendix for full BDD criteria.

### Story 15.2: Mood-Task Correlation & Procrastination Research ✅

**Status:** Done (PR #58)

As the product team,
I want research on mood-task correlation and procrastination interventions,
So that Epic 4's learning algorithm is evidence-informed.

**Acceptance Criteria:**

**Given** research task
**When** review completed
**Then** `../../_bmad-output/planning-artifacts/mood-correlation.md` and `../../_bmad-output/planning-artifacts/procrastination.md` produced
**And** evidence assessment for "progress over perfection"
**And** actionable recommendations for Epic 4
**And** bibliography with accessible references

**Quality Gate (AC-Q5):** All changed files directly related to this story's ACs | scope-checked ✓ | See Appendix for full BDD criteria.

---

## Epic 18: Docker E2E & Headless TUI Testing Infrastructure ✅ COMPLETE

**Epic Goal:** Establish reproducible, automated end-to-end testing for the TUI application using Docker containers for environment isolation and Bubbletea's `teatest` package for headless interaction testing — eliminating manual TUI testing as the sole E2E validation method.

**Prerequisites:** Epic 3 ✅, Epic 9 (Stories 9.4, 9.5)
**FRs covered:** FR49, FR51 (extends Epic 9's scope with concrete implementation approach)
**Origin:** Party mode testing infrastructure discussion (2026-03-02). Party mode consensus identified two critical gaps: (1) no reproducible test environment — tests depend on developer machine state, and (2) TUI testing is entirely manual — 10% of the test pyramid has zero automation.
**Status:** COMPLETE — All 5 stories implemented and merged (PRs #64, #86, #105, #104, #107).

**Why a separate epic from Epic 9:** Epic 9 defines *what* to test (Apple Notes E2E, contract tests, performance benchmarks, functional E2E, CI gates). This epic defines *how* to test the TUI layer specifically — the Docker infrastructure and headless testing tooling that Epic 9 Story 9.4 depends on but doesn't specify.

### Story 18.1: Headless TUI Test Harness with teatest ✅

**Status:** Done (PR #64)

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

### Story 18.2: Golden File Snapshot Tests for TUI Views ✅

**Status:** Done (PR #86)

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

### Story 18.3: Input Sequence Replay Tests for User Workflows ✅

**Status:** Done (PR #105)

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

### Story 18.4: Docker Test Environment for Reproducible E2E ✅

**Status:** Done (PR #104)

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

### Story 18.5: CI Integration for Docker E2E Tests ✅

**Status:** Done (PR #107)

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

## Epic 16: iPhone Mobile App (SwiftUI) — NOT STARTED

**Epic Goal:** Bring the Three Doors experience to iPhone with a native SwiftUI app that syncs tasks via Apple Notes.

**Prerequisites:** Epic 2 ✅
**Tech Stack:** Swift 5.9+, SwiftUI, iCloud Drive, Xcode 16+, iOS 17+
**Status:** NOT STARTED — 7 stories planned (16.1-16.7). See `docs/prd/epic-details.md` for full story details.

### Stories:
- **16.1:** SwiftUI Project Setup & CI
- **16.2:** Task Provider Protocol & Apple Notes Reader
- **16.3:** Three Doors Card Carousel
- **16.4:** Door Detail & Status Actions
- **16.5:** Session Metrics & iCloud Sync
- **16.6:** Swipe Gestures & Pull-to-Refresh
- **16.7:** Polish & TestFlight Distribution

---

## Epic 17: Door Theme System ✅ COMPLETE

**Epic Goal:** Replace uniform door appearance with visually distinct themed doors using ASCII/ANSI art frames, with user-selectable themes via onboarding, settings, and config.yaml.

**Prerequisites:** Epic 3 ✅, Epic 10 ✅
**FRs covered:** FR55-FR62
**Status:** COMPLETE — All 6 stories implemented and merged (PRs #119, #120, #121, #123, #124, #122).

### Story 17.1: Theme Types, Registry, and Classic Theme Wrapper ✅

**Status:** Done (PR #119)

### Story 17.2: Modern, Sci-Fi, and Shoji Theme Implementations ✅

**Status:** Done (PR #120)

### Story 17.3: DoorsView Integration — Load Theme from Config ✅

**Status:** Done (PR #121)

### Story 17.4: Theme Picker in Onboarding Flow ✅

**Status:** Done (PR #123)

### Story 17.5: Settings View — `:theme` Command with Preview ✅

**Status:** Done (PR #124)

### Story 17.6: Golden File Tests for All Themes ✅

**Status:** Done (PR #122)

---

## Epic 19: Jira Integration ✅ COMPLETE

**Epic Goal:** Integrate Jira as a task source, enabling developers to see Jira issues as ThreeDoors tasks.

**Prerequisites:** Epic 7 ✅, Epic 11 ✅, Epic 13 ✅
**FRs covered:** FR63-FR66
**Status:** COMPLETE — All 4 stories implemented and merged (PRs #132, #138, #150, #153).

### Story 19.1: Jira HTTP Client ✅

**Status:** Done (PR #132)

### Story 19.2: Jira Read-Only Provider ✅

**Status:** Done (PR #138)

### Story 19.3: Jira Bidirectional Sync ✅

**Status:** Done (PR #150)

### Story 19.4: Jira Config and Registration ✅

**Status:** Done (PR #153)

---

## Epic 20: Apple Reminders Integration ✅ COMPLETE

**Epic Goal:** Add Apple Reminders as a task source with full CRUD support.

**Prerequisites:** Epic 7 ✅
**FRs covered:** FR67-FR69
**Status:** COMPLETE — All 4 stories implemented and merged (PRs #137, #148, #155, #158).

### Story 20.1: Reminders JXA Scripts and CommandExecutor ✅

**Status:** Done (PR #137)

### Story 20.2: Reminders Read-Only Provider ✅

**Status:** Done (PR #148)

### Story 20.3: Reminders Write Support ✅

**Status:** Done (PR #155)

### Story 20.4: Reminders Config, Registration, and Health Check ✅

**Status:** Done (PR #158)

---

## Epic 21: Sync Protocol Hardening ✅ COMPLETE

**Epic Goal:** Harden sync architecture for reliable multi-provider operation with background scheduling, fault isolation, and cross-provider identity mapping.

**Prerequisites:** Epic 11 ✅, Epic 13 ✅
**FRs covered:** FR70-FR72
**Status:** COMPLETE — All 4 stories implemented and merged (PRs #139, #132, #151, #157).

### Story 21.1: Sync Scheduler with Per-Provider Loops ✅

**Status:** Done (PR #139)

### Story 21.2: Circuit Breaker per Provider ✅

**Status:** Done (PR #132)

### Story 21.3: Canonical ID Mapping (SourceRef) ✅

**Status:** Done (PR #151)

### Story 21.4: Sync Dashboard Enhancements ✅

**Status:** Done (PR #157)

---

## Epic 33: Seasonal Door Theme Variants

**Epic Goal:** Extend the Door Theme System (Epic 17) with time-based seasonal theme variants that auto-switch based on the current date, adding visual variety and delight while maintaining accessibility standards. Each seasonal theme is a standalone `DoorTheme` using only safe Unicode character ranges.

**Prerequisites:** Epic 17 ✅ (Door Theme System)
**FRs covered:** FR132-FR137
**NFRs covered:** NFR28-NFR30
**Status:** Not Started

### Story 33.1: Seasonal Theme Metadata Model and Date-Range Resolver

As a developer,
I want a pure-function seasonal resolver and metadata extensions on DoorTheme,
So that the system can determine which seasonal theme to apply based on the current date.

**Acceptance Criteria:**

**Given** the `DoorTheme` struct in `internal/tui/themes/theme.go`
**When** the seasonal metadata fields are added
**Then** the struct includes `Season string`, `SeasonStart MonthDay`, and `SeasonEnd MonthDay` fields
**And** a `MonthDay` struct is defined with `Month int` and `Day int` fields
**And** zero-value `Season` ("") indicates a non-seasonal theme (backward compatible)

**Given** a `ResolveSeason(now time.Time, ranges []SeasonRange) string` pure function in `internal/tui/themes/seasonal.go`
**When** called with a date falling within a season's date range
**Then** it returns the season name (e.g., "winter", "spring", "summer", "autumn")
**And** when no season matches, it returns an empty string

**Given** `DefaultSeasonRanges` using meteorological seasons
**When** the default ranges are defined
**Then** spring starts March 1, summer starts June 1, autumn starts September 1, winter starts December 1
**And** winter correctly wraps across year boundary (December 1 - February 28/29)

**Given** the `Registry` in `internal/tui/themes/registry.go`
**When** `GetBySeason(season string)` is called
**Then** it returns the first theme with a matching `Season` field, or `(nil, false)` if none match

**Given** table-driven tests in `internal/tui/themes/seasonal_test.go`
**When** `ResolveSeason` is tested
**Then** tests cover: spring start (March 1), spring end (May 31), summer start (June 1), autumn start (September 1), winter start (December 1), winter wrap (January 15), leap year (February 29), season boundary transitions
**And** `ResolveSeason` completes in under 1 microsecond per Go benchmark (NFR28)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- `ResolveSeason` has no I/O dependencies — pure date arithmetic
- Existing themes (classic, modern, scifi, shoji) remain unchanged — `Season` field is zero-value

---

### Story 33.2: Four Seasonal Theme Implementations

As a user,
I want visually distinct seasonal door themes for winter, spring, summer, and autumn,
So that the door presentation reflects the current season with appropriate visual patterns.

**Acceptance Criteria:**

**Given** the theme creation pattern established in `internal/tui/themes/modern.go`
**When** four seasonal themes are implemented
**Then** each theme is in its own file: `winter.go`, `spring.go`, `summer.go`, `autumn.go`
**And** each exports a constructor: `NewWinterTheme()`, `NewSpringTheme()`, `NewSummerTheme()`, `NewAutumnTheme()`
**And** each returns a `*DoorTheme` with populated `Season`, `SeasonStart`, and `SeasonEnd` fields

**Given** the Unicode character constraint (NFR17)
**When** seasonal themes render door frames
**Then** only characters from box-drawing (`U+2500–U+257F`), block elements (`U+2580–U+259F`), and geometric shapes (`U+25A0–U+25FF`) are used
**And** no emoji or Unicode symbols outside these ranges appear in any theme

**Given** the seasonal theme design patterns
**When** each theme renders content
**Then** winter uses crystalline angular frames and dense dot patterns
**And** spring uses flowing curved lines and light open patterns
**And** summer uses radiating lines and bold geometric shapes
**And** autumn uses layered block elements and angular textures
**And** each theme has a distinct visual identity distinguishable from all other themes

**Given** WCAG AA accessibility requirements (FR136)
**When** seasonal themes are rendered
**Then** all text content maintains minimum 4.5:1 contrast ratio against both dark (#000000) and light (#FFFFFF) terminal backgrounds
**And** contrast ratios are validated programmatically in `internal/tui/themes/accessibility_test.go` (NFR30)

**Given** the golden file testing pattern (NFR19, NFR29)
**When** golden file tests run for seasonal themes
**Then** 24 golden files exist: 4 seasons × 3 widths (minimum, 80-column, 120-column) × 2 states (selected, unselected)
**And** all golden files pass comparison

**Given** the `NewDefaultRegistry()` function
**When** the registry is created
**Then** all four seasonal themes are registered alongside existing themes (classic, modern, scifi, shoji)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- Each seasonal render function follows the same pattern as `modernRender` — pure function, `strings.Builder`, `fmt.Fprintf`
- Spring theme uses `╭╮╰╯` curves (Tier 2 risk per analyst review) — verify rendering across iTerm2, Terminal.app, Alacritty

---

### Story 33.3: Auto-Switch Integration in DoorsView and Config

As a user,
I want ThreeDoors to automatically switch to the appropriate seasonal theme based on today's date,
So that I see seasonally appropriate door styling without manual configuration.

**Acceptance Criteria:**

**Given** `seasonal_themes: true` in `~/.threedoors/config.yaml` (default when not specified)
**When** the DoorsView is constructed at app startup
**Then** `ResolveSeason(time.Now().UTC(), DefaultSeasonRanges)` is called
**And** if a matching seasonal theme exists in the registry, it is used instead of the user's configured base theme
**And** the resolved theme is stored in DoorsView for the session duration (no per-render rechecks)

**Given** `seasonal_themes: false` in `~/.threedoors/config.yaml`
**When** the DoorsView is constructed
**Then** no seasonal resolution occurs
**And** the user's configured base theme is used directly

**Given** `seasonal_themes` is not present in config (new or upgraded install)
**When** the config is loaded
**Then** seasonal themes default to enabled (`true`)

**Given** the terminal width is below the seasonal theme's declared `MinWidth`
**When** DoorsView renders doors
**Then** the system falls back to the user's configured base theme (consistent with FR61)

**Given** a planning session starts (FR133)
**When** the planning mode initializes
**Then** seasonal theme resolution is re-checked (handles overnight sessions crossing season boundaries)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- Config schema adds `seasonal_themes` boolean — backward compatible (missing = true)
- DoorsView seasonal resolution is at construction time, not per-render
- Integration test: mock time to verify correct seasonal theme selection

---

### Story 33.4: Seasonal Theme Picker and `:seasonal` Command

As a user,
I want to preview all seasonal themes and manually override the current season,
So that I can see what each seasonal theme looks like and choose my preferred seasonal style.

**Acceptance Criteria:**

**Given** the existing `:theme` command pattern (FR58, Story 17.5)
**When** a user enters `:seasonal` in the TUI command palette
**Then** a seasonal theme picker view opens showing all four seasonal themes in a horizontal preview grid
**And** each preview renders a sample door at standard width in the seasonal theme's style
**And** the currently active seasonal theme (if any) is highlighted

**Given** the seasonal theme picker is open
**When** the user selects a seasonal theme
**Then** the selected seasonal theme is applied immediately (no restart required)
**And** the theme change persists for the current session

**Given** the seasonal theme picker is open
**When** the user presses Esc
**Then** the picker closes without changing the active theme

**Given** the theme_picker.go already exists for `:theme` command
**When** the `:seasonal` picker is implemented
**Then** it reuses the existing theme picker component with a filter for themes where `Season != ""`
**And** the existing `:theme` command continues to show only non-seasonal themes

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- Reuse `ThemePicker` component filtered by `Season != ""`
- Manual season override is session-scoped (not persisted to config — auto-switch resumes on next launch)
- Add `:seasonal` to command palette registration in MainModel

---

### Epic 33 Story Dependencies

```
33.1 Seasonal Theme Metadata Model and Date-Range Resolver
  └── 33.2 Four Seasonal Theme Implementations (depends on 33.1)
  └── 33.3 Auto-Switch Integration in DoorsView and Config (depends on 33.1)
       └── 33.4 Seasonal Theme Picker and `:seasonal` Command (depends on 33.2, 33.3)
```

---

## Epic 35: Door Visual Appearance — Door-Like Proportions ✅ COMPLETE

**Epic Goal:** Redesign all door themes to visually read as actual doors rather than cards/panels, using portrait orientation, panel dividers, asymmetric handles, and threshold/floor lines. Addresses user feedback that "none of the door themes look like doors."

**Prerequisites:** Epic 17 ✅ (Door Theme System)
**FRs covered:** FR138-FR147
**Status:** All 7 stories complete (PRs #226, #229, #234, #236, #237, #238, #239)

### Story 35.1: Door Anatomy Model and Height-Aware Render Signature

As a developer,
I want a `DoorAnatomy` helper type and an updated `Render()` signature that accepts height,
So that all themes can calculate structural row positions for door-like rendering.

**Acceptance Criteria:**

**Given** the `DoorTheme` struct in `internal/tui/themes/theme.go`
**When** the height-aware changes are applied
**Then** the `Render` function signature is `func(content string, width int, height int, selected bool) string`
**And** a `MinHeight int` field is added to `DoorTheme`
**And** all existing theme constructors set `MinHeight` (Classic: 10, Modern: 12, Sci-Fi: 14, Shoji: 14)

**Given** a new `DoorAnatomy` struct in `internal/tui/themes/anatomy.go`
**When** `NewDoorAnatomy(height int)` is called
**Then** it returns structural row positions: `LintelRow` (0), `ContentStart` (2), `PanelDivider` (~45% height), `HandleRow` (~60% height), `ThresholdRow` (height-1)
**And** all positions are within bounds (0 to height-1)

**Given** the `DoorsView` in `internal/tui/doors_view.go`
**When** terminal height is available from `WindowSizeMsg`
**Then** DoorsView calculates door height as available vertical space minus header/footer chrome
**And** passes the calculated height to `theme.Render()`
**And** when door height < theme's `MinHeight`, falls back to compact rendering (existing card layout)

**Given** table-driven tests in `internal/tui/themes/anatomy_test.go`
**When** `NewDoorAnatomy` is tested
**Then** tests cover: minimum height (10), standard height (16), tall height (24), boundary conditions
**And** all row positions are monotonically increasing (lintel < content < divider < handle < threshold)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- All existing theme render functions must be updated to accept the new height parameter
- Initial implementation can ignore height and render at current fixed size — subsequent stories add height-aware rendering
- DoorsView currently only distributes horizontal space; this story adds vertical space management

---

### Story 35.2: Classic Theme — Door-Like Proportions

As a user,
I want the Classic theme to render doors with portrait orientation, panel dividers, handle, and threshold,
So that Classic doors visually read as actual doors.

**Acceptance Criteria:**

**Given** the Classic theme render function
**When** rendering a door with height >= MinHeight (10)
**Then** the door is rendered in portrait orientation (more rows than columns of content)
**And** a panel divider (`├─────────────┤`) is rendered at ~45% of door height
**And** a doorknob (`●`) is rendered on the right side at ~60% of door height
**And** a threshold/shadow line (`▔▔▔▔▔`) is rendered below the bottom border
**And** the door number is rendered in the lintel area (top border)

**Given** the Classic theme in compact mode (height < MinHeight)
**When** rendering a door
**Then** the existing card-style layout is used (backward compatible)

**Given** golden file tests for the Classic theme
**When** tests run at standard height (16 rows) and wide height (24 rows)
**Then** golden files match expected door-like output
**And** both selected and unselected states are tested

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 35.3: Modern Theme — Door-Like Proportions

As a user,
I want the Modern theme to render doors with minimalist door-like proportions,
So that Modern doors clearly read as sleek, minimal doors.

**Acceptance Criteria:**

**Given** the Modern theme render function
**When** rendering a door with height >= MinHeight (12)
**Then** the door uses portrait orientation with heavy top/bottom bars (`━━━`)
**And** a minimalist panel line (`─`) divides upper and lower panels at ~45% height
**And** a minimalist handle (`○`) is rendered on the right side at ~60% height
**And** the door number is centered in the top bar
**And** selected state uses heavy frame characters (`┃`) as currently implemented

**Given** golden file tests for the Modern theme
**When** tests run at minimum, standard, and wide heights
**Then** golden files match expected door-like output

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 35.4: Sci-Fi Theme — Door-Like Proportions

As a user,
I want the Sci-Fi theme to render doors as bulkhead/airlock panels with door-like proportions,
So that Sci-Fi doors clearly read as spaceship doors.

**Acceptance Criteria:**

**Given** the Sci-Fi theme render function
**When** rendering a door with height >= MinHeight (14)
**Then** the door uses portrait orientation with double-line outer frame (`╔═╗║╚═╝`)
**And** a bulkhead divider (`╞═════════╡`) separates upper and lower sections at ~45% height
**And** an access panel handle (`◈──┤`) is rendered on the right side at ~60% height
**And** the `[ACCESS]` label is rendered in the lower panel area
**And** a floor grating line (`▓▓▓`) is rendered below the bottom frame
**And** shade rails (`░` / `▓`) remain on sides

**Given** golden file tests for the Sci-Fi theme
**When** tests run at minimum, standard, and wide heights
**Then** golden files match expected door-like output

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 35.5: Shoji Theme — Door-Like Proportions

As a user,
I want the Shoji theme to render doors as sliding screen panels with door-like proportions,
So that Shoji doors clearly read as traditional Japanese sliding doors.

**Acceptance Criteria:**

**Given** the Shoji theme render function
**When** rendering a door with height >= MinHeight (14)
**Then** the door uses portrait orientation with lattice grid pattern
**And** lattice cross-bars create panel divisions at ~45% height and near the bottom
**And** a recessed sliding handle (`○`) is rendered center-right at ~60% height
**And** the door number is placed in the top rail
**And** selected state uses heavy grid characters (`╋━┃`) as currently implemented

**Given** golden file tests for the Shoji theme
**When** tests run at minimum, standard, and wide heights
**Then** golden files match expected door-like output

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 35.6: Golden File Test Regeneration and Accessibility Validation

As a QA engineer,
I want comprehensive golden file tests and accessibility validation for all door-like themes,
So that door proportions are regression-tested and accessible.

**Acceptance Criteria:**

**Given** all four themes with door-like proportions
**When** golden file tests are regenerated
**Then** golden files exist for each theme at 3 heights (minimum, 16-row, 24-row) × 3 widths (minimum, 80-col, 120-col) × 2 states (selected, unselected)
**And** all golden file comparisons pass

**Given** monochrome rendering mode
**When** all themes are rendered without color
**Then** door signifiers (panel divider, handle position, threshold) are still visually distinguishable using structural characters only

**Given** the compact mode fallback
**When** terminal height is below each theme's MinHeight
**Then** the system renders using the card-style layout
**And** all task content remains fully readable

**Given** screen reader compatibility
**When** door content is extracted as plain text
**Then** task text is accessible in reading order without interference from decorative frame characters

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 35.7: Shadow/Depth Effect for 3D Door Appearance

As a user,
I want doors to have subtle shadow/depth effects on the right and bottom edges,
So that doors appear to have dimension and stand out from the background.

**Acceptance Criteria:**

**Given** any door theme rendering a door
**When** the door is rendered at sufficient width (>= MinWidth + 2)
**Then** a shadow column using half-block characters (`▐` or `░`) is rendered on the right edge of the door
**And** a shadow row using half-block characters (`▄`) is rendered below the threshold line
**And** the shadow creates a visual impression of depth/3D

**Given** the selected door state
**When** a door is highlighted as selected
**Then** the shadow effect is enhanced (brighter or thicker shadow) to lift the selected door visually

**Given** terminal width insufficient for shadow (< MinWidth + 2)
**When** doors are rendered
**Then** shadow is omitted gracefully (no layout breakage)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Technical Notes:**
- Shadow characters are from block elements range (U+2580-U+259F) — safe Unicode
- Shadow adds 1-2 chars width and 1 row height overhead — account in width/height calculations

---

### Epic 35 Story Dependencies

```
35.1 Door Anatomy Model and Height-Aware Render Signature
  ├── 35.2 Classic Theme — Door-Like Proportions (depends on 35.1)
  ├── 35.3 Modern Theme — Door-Like Proportions (depends on 35.1)
  ├── 35.4 Sci-Fi Theme — Door-Like Proportions (depends on 35.1)
  └── 35.5 Shoji Theme — Door-Like Proportions (depends on 35.1)
       └── 35.6 Golden File Test Regeneration (depends on 35.2-35.5)
       └── 35.7 Shadow/Depth Effect (depends on 35.2-35.5)
```

---

## Epic 36: Door Selection Interaction Feedback

**Epic Goal:** Make door selection feel responsive and satisfying by enhancing visual feedback contrast, adding deselect toggle, and ensuring universal quit. Addresses GitHub Issue #219.

**Prerequisites:** None (complements Epic 35 but does not depend on it)
**FRs covered:** FR148-FR151
**Status:** COMPLETE (4/4 stories done)

### Story 36.1: Enhanced Door Selection Visual Feedback

As a user,
I want the selected door to be visually unmistakable through strong contrast with unselected doors,
So that every keypress produces a satisfying, confident "I picked this one" response.

**Acceptance Criteria:**

**Given** the doors view with three doors displayed and no door selected
**When** the user presses a selection key (a/left, w/up, d/right)
**Then** the selected door renders with bold text, bright foreground, and enhanced border
**And** unselected doors render with faint/dimmed text and subdued border color

**Given** no door is selected (selectedDoorIndex == -1)
**When** all three doors render
**Then** all doors use their normal (non-dimmed, non-emphasized) styling

**Given** theme-based rendering is active
**When** a door is selected
**Then** the contrast difference is apparent even in monochrome mode

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 36.2: Deselect Toggle — Press Same Key to Unselect

As a user,
I want to press the same selection key again to deselect a door,
So that I can explore options freely without feeling prematurely committed.

**Acceptance Criteria:**

**Given** door N is currently selected
**When** the user presses the same key that selected door N
**Then** selectedDoorIndex is set to -1 (no selection)
**And** all doors return to their neutral visual state

**Given** door N is currently selected
**When** the user presses a different selection key
**Then** selection switches to the new door (normal behavior)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 36.3: Universal Quit — 'q' Works From All Screens

As a user,
I want 'q' to quit the application from any non-text-input view,
So that I never feel trapped in a screen.

**Acceptance Criteria:**

**Given** the user is on any non-text-input view
**When** the user presses 'q'
**Then** the application exits cleanly

**Given** the user is in a text input mode (search, feedback, quick add)
**When** the user presses 'q'
**Then** 'q' is treated as text input, not as quit

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

---

### Story 36.4: Space/Enter Toggle — Close Door by Pressing Same Key

As a user,
I want to press space or enter again to close an open door (return from DetailView to DoorsView),
So that opening and closing a door uses the same natural gesture, like a real door.

**Acceptance Criteria:**

**Given** the user is in DetailView (viewing a task's details) and no text input is active
**When** the user presses space or enter
**Then** the view switches back to DoorsView with the door still selected

**Given** the user is in DetailView with a text input active (feedback form, snooze picker)
**When** the user presses space or enter
**Then** the keypress is handled by the text input (not intercepted as toggle)

**Given** the user presses Escape in DetailView
**Then** the view switches back to DoorsView (existing behavior preserved)

**Quality Gate:** AC-Q1 (formatting), AC-Q2 (lint), AC-Q3 (test coverage), AC-Q4 (rebase), AC-Q5 (scope)

**Decision:** D-137 — Space/Enter as toggle in DetailView

---

### Epic 36 Story Dependencies

```
36.1 Enhanced Door Selection Visual Feedback (independent)
36.2 Deselect Toggle (independent)
36.3 Universal Quit (independent)
36.4 Space/Enter Toggle (depends on 36.2 pattern, 39.6 spacebar alias)
```

Stories 36.1-36.3 are independent. Story 36.4 extends the toggle pattern from 36.2.

---

## Epic 37: Persistent BMAD Agent Infrastructure

**Epic Goal:** Enable autonomous project governance by adding persistent BMAD agents (project-watchdog, arch-watchdog) and cron jobs (SM sprint health, QA coverage audit) that automatically maintain story status, ROADMAP accuracy, architecture doc currency, and quality metrics.

**Prerequisites:** None (infrastructure epic, independent of all feature epics)
**FRs covered:** N/A (development infrastructure, not product features)
**Status:** COMPLETE

### Story 37.1: Agent Definitions — project-watchdog and arch-watchdog

As a project supervisor,
I want persistent project-watchdog and arch-watchdog agents with well-defined monitoring surfaces, authority boundaries, and restart behavior,
So that project governance happens automatically after every PR merge.

**Acceptance Criteria:**

**Given** the `agents/` directory
**When** implementation is complete
**Then** `agents/project-watchdog.md` and `agents/arch-watchdog.md` exist with monitoring surfaces, trigger models, authority boundaries, escalation rules, restart behavior, and correlation ID tracking

**Given** either agent re-processes a previously-seen PR
**When** the correlation ID matches an already-processed PR
**Then** no duplicate messages or file edits are produced (idempotency verified)

**Quality Gate:** AC-Q5 (scope)

---

### Story 37.2: Cron Configuration — SM Sprint Health and QA Coverage Audit

As a project supervisor,
I want automated sprint health checks every 4 hours and weekly QA coverage audits,
So that blocked stories, stale PRs, and coverage regressions are surfaced without manual intervention.

**Acceptance Criteria:**

**Given** the SM sprint health cron
**When** running every 4 hours
**Then** it queries stale PRs, blocked stories, and worker activity, reporting risks to supervisor

**Given** the QA coverage audit cron
**When** running weekly
**Then** it compares per-package coverage against a stored baseline at `docs/quality/coverage-baseline.json` and flags regressions >5 percentage points

**Quality Gate:** AC-Q5 (scope)

---

### Story 37.3: Agent Communication Architecture Documentation

As a developer or agent maintainer,
I want comprehensive architecture documentation for persistent agent communication patterns, authority boundaries, and anti-patterns,
So that agent behavior is predictable, debuggable, and extensible.

**Acceptance Criteria:**

**Given** the `docs/architecture/` directory
**When** implementation is complete
**Then** `docs/architecture/agent-governance.md` exists with: agent interaction architecture, communication protocol, authority boundaries, anti-patterns and safeguards, resource budget and scaling, lifecycle management
**And** `docs/architecture/index.md` references the new file

**Quality Gate:** AC-Q5 (scope)

---

### Story 37.4: Monitoring, Tuning, and Phase 1 Evaluation

As a project supervisor,
I want a 2-week evaluation framework with clear success metrics and tuning criteria,
So that I can objectively assess persistent agent value and adjust accordingly.

**Acceptance Criteria:**

**Given** the evaluation framework
**When** implementation is complete
**Then** `docs/operations/agent-evaluation.md` exists with success metrics, tuning guidelines, Phase 1 evaluation checklist, and escalation criteria

**Quality Gate:** AC-Q5 (scope)

---

### Epic 37 Story Dependencies

```
37.1 Agent Definitions (independent)
37.2 Cron Configuration (independent)
37.3 Agent Communication Architecture Documentation (depends on 37.1)
37.4 Monitoring, Tuning, and Phase 1 Evaluation (depends on 37.1, 37.2)
```

---

## Epic 38: Dual Homebrew Distribution

**Epic Goal:** Establish parallel Homebrew distribution channels — stable (`threedoors`) and alpha (`threedoors-a`) — with signing parity, publishing controls, verification, and retention management.

**Prerequisites:** GoReleaser release pipeline (complete), Apple Developer ID signing infrastructure (complete)
**Status:** In Progress (1/5 stories in review)

**Deliverables:**
- Alpha Homebrew formula (`threedoors-a.rb`) auto-updated on every push to main
- Publishing toggle (`vars.ALPHA_TAP_ENABLED`) for controlled activation
- Code signing and notarization for stable GoReleaser releases (signing parity)
- Alpha formula verification via tap CI monitoring
- Alpha release retention cleanup (keep last 30)

**Stories:**

### Story 38.1: Alpha Homebrew Formula (`threedoors-a`)
- **Status:** In Review (PR #273)
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **ACs:** Alpha binary naming (`threedoors-a-*`), formula in tap, CI auto-update, code signing, no conflicts with stable, channel identifier in `--version`, alpha release includes new binaries

### Story 38.2: Alpha Publishing Toggle
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** S (1-2 hours)
- **Depends on:** 38.1
- **ACs:** `vars.ALPHA_TAP_ENABLED` toggle gates formula push, alpha release always created, optional template refactor

### Story 38.3: Stable Release Signing & Notarization
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** L (5-8 hours)
- **Depends on:** None (independent)
- **ACs:** Post-GoReleaser signing job on macOS, darwin binary signing, notarization, re-upload signed archives, fail-open behavior, shared signing identity

### Story 38.4: Alpha Release Verification
- **Status:** Not Started
- **Priority:** P2
- **Estimate:** M (3-5 hours)
- **Depends on:** 38.1, 38.2
- **ACs:** Tap CI monitoring for alpha formula, issue creation on failure, graceful handling when toggle is off

### Story 38.5: Alpha Release Retention Cleanup
- **Status:** Not Started
- **Priority:** P2
- **Estimate:** S (1-2 hours)
- **Depends on:** None (independent)
- **ACs:** Keep last 30 alpha releases, delete older with `--cleanup-tag`, stable releases never affected, idempotent and safe

### Story 38.6: Fix Alpha Homebrew Formula Template DSL
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** S (1-2 hours)
- **Depends on:** 38.1
- **Issue:** [#296](https://github.com/arcaven/ThreeDoors/issues/296)
- **ACs:** Alpha formula template uses `if OS.mac? && Hardware::CPU.arm?` pattern, no `on_arm`/`on_intel`/`on_linux` with `url`/`sha256`, formula passes `brew audit --strict`, homebrew-tap CI green

---

### Epic 38 Story Dependencies

```
38.1 Alpha Homebrew Formula (independent, PR #273)
38.2 Alpha Publishing Toggle (depends on 38.1)
38.3 Stable Release Signing (independent)
38.4 Alpha Release Verification (depends on 38.1, 38.2)
38.5 Alpha Release Retention (independent)
38.6 Fix Alpha Formula Template DSL (depends on 38.1, fixes #296)
```

---

## Epic 39: Keybinding Display System

**Epic Goal:** Add toggleable keybinding discoverability to the ThreeDoors TUI — a concise context-sensitive bar at the bottom of every view showing available keys, and a full keybinding overlay (`?` key) as a comprehensive reference. Improves discoverability without adding decision complexity, aligning with SOUL.md's friction-reduction philosophy.

**Prerequisites:** None (all required infrastructure exists — Lipgloss, config.yaml persistence, MainModel composition, isTextInputActive() guard)
**Status:** In Progress (4/12)

**Deliverables:**
- Compile-time keybinding registry mapping each ViewMode to available key bindings with priority levels
- Concise bottom bar showing 5-6 priority keys per view with dim Lipgloss styling
- Full-screen keybinding overlay organized by category with scroll support
- `h` key toggles bar visibility (persisted to config.yaml), `?` key opens/closes overlay
- Terminal size adaptation: auto-hide bar < 10 lines, compact mode 10-15 lines, width truncation
- Context-sensitive bar content (changes per view mode, sub-mode aware)
- Inline key hints rendered directly on interactive elements (doorknob metaphor for doors)
- Auto-fade after N sessions (default 5) with graceful dimming transition
- `:hints` command for manual re-enable/disable

**Design References:**
- Party mode: `_bmad-output/planning-artifacts/keybinding-display-party-mode.md`
- UX review: `_bmad-output/planning-artifacts/keybinding-display-ux-review.md`
- Architecture: `_bmad-output/planning-artifacts/keybinding-display-architecture.md`

### Story 39.1: Keybinding Registry Model
- **Status:** Done (PR #305)
- **Priority:** P1
- **Estimate:** S (1-2 hours)
- **Depends on:** None
- **ACs:** KeyBinding/KeyBindingGroup types, per-view registry functions (all 18 ViewModes), barBindings() convenience function (priority-1 only, max 8 per view), allKeyBindingGroups() for overlay, comprehensive table-driven tests

### Story 39.2: Concise Keybinding Bar Component
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.1
- **ACs:** RenderKeybindingBar() function, context-sensitive content from registry, terminal width adaptation (4 breakpoints), terminal height adaptation (3 breakpoints), dim Lipgloss styling with separator line, golden file tests, unit tests

### Story 39.3: Full Keybinding Overlay
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.1
- **ACs:** RenderKeybindingOverlay() function, bordered box with categorized bindings, context highlighting (current view first), scroll support with j/k/arrows, fixed footer, golden file tests, unit tests

### Story 39.4: Toggle Behavior, Config Persistence, and MainModel Integration
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.2, 39.3
- **ACs:** MainModel fields and config initialization, `h` toggle with async config write, `?` overlay toggle, overlay key interception, View() composition with height adjustment, config.yaml persistence, integration tests, race detector pass

### Story 39.5: View-Specific Keybinding Completeness and Polish
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.4
- **ACs:** Full keybinding audit (every case handler registered), sub-mode awareness (confirm-delete, expand input, command mode), overlay includes `:` commands section, visual polish, comprehensive golden files for all major views, edge case tests

### Story 39.6: Spacebar as Enter Alias in Doors View
- **Status:** Done (PR #303)
- **Priority:** P1
- **Estimate:** XS (<1 hour)
- **Depends on:** None
- **ACs:** Spacebar opens selected door (Enter alias), no-op when no selection, help text updated, table-driven tests, race detector passes

### Story 39.7: Global `:` Command Mode
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** S (1-2 hours)
- **Depends on:** None
- **ACs:** `:` intercepted at MainModel level with `isTextInputActive()` guard (same as D-059/D-087 pattern), removed from `updateDoors()`, `previousView` tracks originating view, text input views unaffected, table-driven tests, race detector passes

### Story 39.8: Command Autocomplete/Completion
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** None (benefits from 39.7 but independent)
- **ACs:** Command registry with names and descriptions, dynamic prefix-match filtering as user types, inline suggestion rendering with descriptions, arrow/Tab navigation and completion, table-driven tests, race detector passes

### Story 39.9: Inline Hint Rendering Infrastructure
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** S (1-2 hours)
- **Depends on:** 39.1
- **ACs:** renderInlineHint() function, theme Render() signature extended with hint parameter, config model extension (show_inline_hints, session counter, fade threshold), `:hints` command registration, session counter logic with auto-fade, unit tests

### Story 39.10: Door View Inline Hints
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.9
- **ACs:** Door frame hints [a]/[w]/[d] via doorknob metaphor (Approach B), selection state awareness, additional action hints [s]/[n]/[enter], help text simplification when hints active, theme compatibility, golden file tests, race detector passes

### Story 39.11: Non-Door View Inline Hints
- **Status:** Not Started
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.9
- **ACs:** Detail view hints (esc/c/b/e/f), search view hints (enter/esc/arrows), mood view numbered labels, add task hints, consistent styling via renderInlineHint(), golden file tests, race detector passes

### Story 39.12: Auto-Fade After N Sessions
- **Status:** Cancelled (superseded by D-140 — auto-fade removed in favor of manual `h` toggle)
- **Priority:** ~~P2~~
- **Estimate:** ~~S (1-2 hours)~~
- **Depends on:** 39.9, 39.10
- **ACs:** ~~Graceful dimming at session N-1 (ANSI 240), auto-disable at session N with flash message, re-enable via :hints on resets counter, configurable threshold (default 5, 0 = never fade), unit tests~~ — Cancelled per course correction D-137/D-140

### Story 39.13: Unified Door Key Indicator Toggle
- **Status:** Done (PR #407)
- **Priority:** P1
- **Estimate:** M (3-5 hours)
- **Depends on:** 39.4, 39.9, 39.10
- **ACs:** Unify `h` toggle and `:hints` under single `show_key_hints` config; door key indicators `[a]`/`[w]`/`[d]` on doors controlled by `h`; remove bottom bar in doors view; keep bar for non-door views; config migration from old field names; remove auto-fade mechanism; `:hints` becomes alias; table-driven tests; race detector passes

---

### Epic 39 Story Dependencies

```
39.1 Keybinding Registry Model (independent)
39.2 Concise Bar Component (depends on 39.1)
39.3 Full Keybinding Overlay (depends on 39.1)
39.4 Toggle + Integration (depends on 39.2, 39.3)
39.5 Completeness + Polish (depends on 39.4)
39.6 Spacebar as Enter Alias (independent)
39.7 Global : Command Mode (independent)
39.8 Command Autocomplete/Completion (independent, benefits from 39.7)
39.9 Inline Hint Infrastructure (depends on 39.1)
39.10 Door View Inline Hints (depends on 39.9)
39.11 Non-Door View Inline Hints (depends on 39.9)
39.12 Auto-Fade After N Sessions — CANCELLED (D-140)
39.13 Unified Door Key Indicator Toggle (depends on 39.4, 39.9, 39.10)
```

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

---

## Epic 22: Self-Driving Development Pipeline ✅ COMPLETE

**Epic Goal:** Enable ThreeDoors tasks to directly trigger multiclaude worker agents, creating a closed loop where the app dispatches its own development work and tracks results (PRs, CI status) back in the TUI. This is the "meta" feature: ThreeDoors managing its own development.

**Prerequisites:** Epic 14 ✅ (LLM Decomposition — provides AgentService for optional story generation), multiclaude installed and configured
**FRs covered:** FR73, FR74, FR75, FR76, FR77, FR78, FR79, FR80
**NFRs covered:** NFR24, NFR25, NFR26, NFR27
**Origin:** Self-driving development pipeline research (2026-03-04). Research document at `../../_bmad-output/planning-artifacts/self-driving-development-pipeline.md`.
**Architecture:** Option B (TUI-Native Dispatch) — single-process, unified UX, leverages existing multiclaude CLI and Bubbletea patterns.
**Status:** COMPLETE — All 8 stories implemented and merged (PRs #149, #152, #163, #162, #161, #164, #159, #160).

**Key Design Decisions:**
- Dispatch state (`DevDispatch`) is orthogonal to task lifecycle status — a task can be `in-progress` AND dispatched
- File-based queue (`~/.threedoors/dev-queue.yaml`) — consistent with YAML data model, offline-capable, inspectable
- 30-second `tea.Tick` polling via `multiclaude repo history` — simple, reliable, matches Bubbletea patterns
- Feature gated behind `dev_dispatch_enabled: true` in config — disabled by default
- No auto-dispatch by default — user must explicitly approve each dispatch
- Max 2 concurrent workers — conservative default to prevent cost runaway

### Story 22.1: Dev Dispatch Data Model and Queue Persistence ✅

**Status:** Done (PR #149)

As a developer,
I want a `DevDispatch` struct on the `Task` type and a file-based dev queue,
So that dispatch state is tracked independently from task lifecycle and persists across TUI restarts.

**Acceptance Criteria:**

**Given** the need to track dev dispatch state orthogonal to task status
**When** the data model is created
**Then:**
- AC1: `DevDispatch` struct defined in `internal/dispatch/model.go` with fields: Queued (`bool`), QueuedAt (`*time.Time`), WorkerName (`string`), PRNumber (`int`), PRStatus (`string`), DispatchErr (`string`)
- AC2: `QueueItem` struct defined with fields: ID, TaskID, TaskText, Context, Status (pending/dispatched/completed/failed), Priority, Scope, AcceptanceCriteria, QueuedAt, DispatchedAt, CompletedAt, WorkerName, PRNumber, PRURL, Error
- AC3: `DevQueue` struct with `Load(path string) error`, `Save(path string) error`, `Add(item QueueItem) error`, `Get(id string) (QueueItem, error)`, `Update(id string, fn func(*QueueItem)) error`, `List() []QueueItem`
- AC4: Queue file location defaults to `~/.threedoors/dev-queue.yaml`
- AC5: Queue persistence uses atomic write pattern (write to `.tmp`, sync, rename)
- AC6: `Task` struct in `internal/core/task.go` extended with `DevDispatch *DevDispatch` field (pointer, omitempty)
- AC7: Unit tests for queue CRUD operations and atomic write safety

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.2: Dispatch Engine with multiclaude CLI Wrapper ✅

**Status:** Done (PR #152)

As a developer,
I want a dispatch engine that wraps the multiclaude CLI,
So that ThreeDoors can create workers, list workers, get history, and remove workers programmatically.

**Acceptance Criteria:**

**Given** the need to interact with multiclaude from within ThreeDoors
**When** the dispatch engine is implemented
**Then:**
- AC1: `Dispatcher` interface defined in `internal/dispatch/dispatcher.go` with methods: `CreateWorker(ctx, task string) (workerName string, err error)`, `ListWorkers(ctx) ([]WorkerInfo, error)`, `GetHistory(ctx, limit int) ([]HistoryEntry, error)`, `RemoveWorker(ctx, name string) error`
- AC2: `CLIDispatcher` concrete implementation wraps `os/exec` calls to `multiclaude` CLI
- AC3: `CommandRunner` interface for testability (mock subprocess execution)
- AC4: Task-to-worker translation builds rich prompt from task text, context, acceptance criteria, scope, and standard suffix (signing, fork workflow)
- AC5: `CheckAvailable(ctx) error` method validates `multiclaude` is on PATH
- AC6: Unit tests with mock `CommandRunner` for all dispatch operations
- AC7: Error wrapping follows `fmt.Errorf("dispatch %s: %w", op, err)` pattern

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.3: TUI Dispatch Key Binding and Confirmation Flow ✅

**Status:** Done (PR #163)

As a user,
I want to press 'x' in the task detail view or type `:dispatch` to dispatch a task to the dev queue,
So that I can trigger automated development work on a selected task.

**Acceptance Criteria:**

**Given** a task is selected in the detail view and dev dispatch is enabled
**When** the user presses 'x' or types `:dispatch`
**Then:**
- AC1: Confirmation dialog appears: "Dispatch '<task text>' to dev queue? [y/n]"
- AC2: On 'y', task is added to dev queue with status `pending` and `Task.DevDispatch.Queued` set to `true`
- AC3: On 'n', confirmation is dismissed with no side effects
- AC4: If task is already dispatched, show message "Task already dispatched" and do not re-enqueue
- AC5: If multiclaude is not available, 'x' key and `:dispatch` command are hidden/disabled with message "multiclaude not found — dev dispatch unavailable"
- AC6: If `dev_dispatch_enabled` is `false` in config, 'x' key and `:dispatch` are not registered
- AC7: `[DEV]` badge appears on dispatched tasks in the doors view

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.4: Dev Queue View (List, Approve, Kill) ✅

**Status:** Done (PR #162)

As a user,
I want a dev queue view where I can see pending dispatches, approve them, and kill running workers,
So that I maintain control over what gets dispatched and can stop runaway agents.

**Acceptance Criteria:**

**Given** the user opens the dev queue view via `:devqueue` command
**When** the view renders
**Then:**
- AC1: Queue items displayed as a list with columns: Status (icon), Task Text (truncated), Worker Name, PR #, Queued At
- AC2: 'y' key approves a pending item (triggers `multiclaude worker create`)
- AC3: 'n' key rejects a pending item (removes from queue)
- AC4: 'K' key kills a dispatched/running worker (`multiclaude worker rm`)
- AC5: 'j'/'k' or arrow keys navigate the list
- AC6: ESC returns to the doors view
- AC7: Status icons: ⏳ pending, ⚙️ dispatched, ✅ completed, ❌ failed
- AC8: View auto-refreshes on 30-second tick (same as worker status polling)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.5: Worker Status Polling and Task Update Loop ✅

**Status:** Done (PR #161)

As a user,
I want ThreeDoors to automatically check on worker status and update tasks with PR results,
So that I can see development progress without leaving the TUI.

**Acceptance Criteria:**

**Given** one or more queue items are in `dispatched` status
**When** the 30-second tick fires
**Then:**
- AC1: `tea.Tick` command fires every 30 seconds while any queue items are in `dispatched` status
- AC2: Tick runs `multiclaude repo history` via the dispatch engine and parses output
- AC3: Worker name matched to queue item; status updated (dispatched → completed/failed)
- AC4: PR number and URL extracted and set on queue item and `Task.DevDispatch`
- AC5: Task badge in doors view updates to show PR status (e.g., `[PR #134]`)
- AC6: Polling stops when no items are in `dispatched` status (no unnecessary ticks)
- AC7: Parse errors logged but do not crash the TUI — Bubbletea `Update()` must never panic

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.6: Auto-Generated Review and Follow-Up Tasks ✅

**Status:** Done (PR #164)

As a user,
I want ThreeDoors to automatically create review and follow-up tasks when workers produce results,
So that PR reviews and CI fixes appear naturally in my door rotation.

**Acceptance Criteria:**

**Given** a worker has completed and a PR has been created
**When** the polling loop detects the completion
**Then:**
- AC1: New task created: "Review PR #N: <original task text>" with status `todo`
- AC2: If CI fails on the PR, new task created: "Fix CI on PR #N: <failure summary>" with status `todo`
- AC3: Generated tasks appear in normal door rotation
- AC4: Generated tasks reference the original task ID in their context field
- AC5: No duplicate tasks generated — if "Review PR #N" already exists, skip creation
- AC6: Auto-generated tasks have `DevDispatch.PRNumber` pre-set for traceability

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.7: Optional Story File Generation via AgentService ✅

**Status:** Done (PR #159)

As a user,
I want to optionally generate story files before dispatching a task,
So that workers receive structured requirements following the project's story-driven development pattern.

**Acceptance Criteria:**

**Given** `require_story: true` is set in dev queue settings
**When** a task is dispatched
**Then:**
- AC1: `AgentService.DecomposeAndWrite()` is called to generate BMAD-style story files from the task
- AC2: Story files are committed to a branch before the worker is spawned
- AC3: Worker task description includes instructions to implement the generated stories
- AC4: If `require_story: false`, story generation is skipped and the worker receives the raw task description
- AC5: Story generation failure is non-fatal — logs error, proceeds with raw task dispatch, sets warning on queue item
- AC6: Configuration option `require_story` defaults to `false`

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 22.8: Safety Guardrails (Rate Limiting, Cost Caps, Audit Log) ✅

**Status:** Done (PR #160)

As a user,
I want safety guardrails preventing runaway agent spawning and providing an audit trail,
So that I can use the self-driving pipeline without risk of excessive cost or uncontrolled automation.

**Acceptance Criteria:**

**Given** the dispatch engine is operational
**When** guardrails are configured
**Then:**
- AC1: Max concurrent workers enforced (default 2) — dispatch refused with message if at capacity
- AC2: Manual approval gate by default (`auto_dispatch: false`) — pending items require explicit 'y' in dev queue view
- AC3: Minimum 5-minute cooldown between dispatches to the same task
- AC4: Daily dispatch limit enforced (default 10) — dispatch refused with message if exceeded
- AC5: Every dispatch, completion, and failure logged to `~/.threedoors/dev-dispatch.log` in JSONL format
- AC6: `:dispatch --dry-run` shows the full multiclaude command without executing
- AC7: Guardrail settings configurable in `~/.threedoors/config.yaml` under `dev_dispatch` section
- AC8: All guardrail violations produce user-visible messages in the TUI (not silent failures)

**Quality Gate (AC-Q1–Q8):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Epic 22 Story Dependencies

```
22.1 (Data Model) ──┬──> 22.2 (Dispatch Engine) ──┬──> 22.4 (Dev Queue View)
                    │                               ├──> 22.5 (Status Polling)
                    │                               ├──> 22.7 (Story Generation)
                    │                               └──> 22.8 (Safety Guardrails)
                    │
                    └──> 22.3 (TUI Dispatch Binding)

                         22.5 (Status Polling) ────> 22.6 (Auto-Generated Tasks)
```

### MVP Phasing

**MVP-1 (Stories 22.1, 22.2, 22.3):** Data model + dispatch engine + TUI binding. User can dispatch tasks from the TUI; approval and execution via manual script or dev queue view.

**MVP-2 (Stories 22.4, 22.5):** Dev queue view + polling. Full TUI-integrated dispatch with automatic status tracking.

---

## Epic 24: MCP/LLM Integration Server ✅ COMPLETE

**Epic Goal:** Expose ThreeDoors task management services to LLMs through the Model Context Protocol (MCP). LLMs can query tasks, propose enrichments (with user approval), mine productivity analytics, and traverse task relationship graphs across providers. Core design principle: LLMs propose, users approve — no direct task modification.

**Prerequisites:** Epic 13 ✅ (Multi-Source Aggregation), Epic 6 ✅ (Enrichment DB)
**FRs covered:** FR81, FR82, FR83, FR84, FR85, FR86, FR87, FR88
**Origin:** LLM Integration & MCP Server Research (2026-03-06). Research document at `../../_bmad-output/planning-artifacts/llm-integration-mcp.md`.
**Architecture:** Separate binary (`cmd/threedoors-mcp/`) sharing `internal/` packages. No new storage layer — reads same YAML, JSONL, and SQLite as TUI.
**Status:** All 8 stories complete (PRs #164-#196)

**Key Design Decisions:**
- MCP server is a separate binary from the TUI — independently deployable
- LLMs NEVER directly edit task data — all modifications flow through proposal/approval pattern
- No new storage layer — MCP server reads from the same files as the TUI
- Proposal store is append-only JSONL (`~/.threedoors/proposals.jsonl`)
- Server supports stdio (Claude Desktop) and SSE (remote) transports
- Security: rate limiting, audit logging with SHA-256 hash chain, input validation, read-only enforcement

### Story 24.1: MCP Server Binary & Transport Layer

**Status:** draft

As a developer,
I want a standalone MCP server binary that implements the MCP protocol over stdio and SSE transports,
So that LLM clients (Claude Desktop, Cursor, etc.) can connect to ThreeDoors and discover available capabilities.

**Acceptance Criteria:**
- **AC1:** `cmd/threedoors-mcp/main.go` entry point with `MCPServer` wrapping existing core components
- **AC2:** MCP JSON-RPC protocol handlers: `initialize`, `resources/list`, `tools/list`, `prompts/list`
- **AC3:** stdio transport (default) for Claude Desktop integration
- **AC4:** SSE transport (`--transport sse --port 8080`) for remote access
- **AC5:** `MCPMiddleware` type as `func(Handler) Handler` decorator pattern
- **AC6:** `Makefile` updated with `build-mcp` target
- **AC7:** Unit tests for protocol handshake and transport selection

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.2: Read-Only Task Resources & Query Tools

**Status:** draft

As an LLM client connected via MCP,
I want to read task data, query tasks with filters, and inspect provider health,
So that I can understand the user's task landscape and answer questions about their work.

**Acceptance Criteria:**
- **AC1:** MCP Resources: `threedoors://tasks`, `threedoors://tasks/{id}`, `threedoors://tasks/status/{status}`, `threedoors://tasks/provider/{name}`
- **AC2:** MCP Resources: `threedoors://providers`, `threedoors://session/current`, `threedoors://session/history`
- **AC3:** MCP Tool `query_tasks` with filters: status, type, effort, provider, text, dates, limit, sort
- **AC4:** MCP Tools: `get_task`, `list_providers`, `get_session`
- **AC5:** Response metadata on all queries: `total_count`, `returned_count`, `query_time_ms`, `providers_queried`, `data_freshness`
- **AC6:** `TaskQueryEngine` with text search, token overlap scoring, field weighting, recency boost
- **AC7:** Unit tests for all resources, tools, and query engine

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.3: Security Middleware — Rate Limiting & Audit Logging

**Status:** draft

As a system administrator,
I want the MCP server to enforce rate limits, log all requests, and validate inputs,
So that the server is protected against abuse and maintains a tamper-evident audit trail.

**Acceptance Criteria:**
- **AC1:** `RateLimiter`: 100 req/min global, 20 proposals/min, 60 queries/min, 5 pending proposals/task, 10-request burst
- **AC2:** `AuditLogger`: JSONL to `~/.threedoors/mcp-audit.jsonl` with SHA-256 hash chain
- **AC3:** `SchemaValidator`: UUID v4 task IDs, 500-char text limit, valid status/timestamps
- **AC4:** `ReadOnlyEnforcer`: blocks direct `SaveTask()` calls
- **AC5:** Middleware chain: ReadOnlyEnforcer → RateLimiter → AuditLogger → SchemaValidator → coreHandler
- **AC6:** Daily log rotation with 30-day retention
- **AC7:** Unit tests for each middleware and composed chain

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.4: Proposal Store & Controlled Enrichment API

**Status:** draft

As an LLM client,
I want to propose task enrichments through a controlled API,
So that I can suggest improvements without directly modifying task data.

**Acceptance Criteria:**
- **AC1:** `Proposal` struct with ID, Type, TaskID, BaseVersion, Payload, Status, Source, Rationale, timestamps
- **AC2:** 8 proposal types: enrich-metadata, add-subtasks, add-context, add-note, suggest-blocker, suggest-relationship, suggest-category, update-effort
- **AC3:** `ProposalStore` with append-only JSONL persistence
- **AC4:** Optimistic concurrency: BaseVersion vs current UpdatedAt — stale detection
- **AC5:** MCP Tools: `propose_enrichment`, `suggest_task`, `suggest_relationship`
- **AC6:** MCP Resource: `threedoors://proposals/pending`
- **AC7:** Deduplication, per-task caps (5), 7-day expiration
- **AC8:** `IntakeChannel` interface for extensible intake sources

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.5: TUI Proposal Review View

**Status:** draft

As a user,
I want to review, approve, and reject LLM-generated proposals from within the TUI,
So that I maintain full control over what changes are applied to my tasks.

**Acceptance Criteria:**
- **AC1:** Badge indicator on doors view: `[3 suggestions]`
- **AC2:** Review view via `S` key or `:suggestions` command — split pane layout
- **AC3:** Quick actions: Enter=approve, Backspace=reject, Tab=skip, Ctrl+A=approve all
- **AC4:** On approve: payload applied to task via `SaveTask()`, enrichment DB updated
- **AC5:** Stale proposals visually distinguished with tooltip
- **AC6:** Preview mode showing task diff before/after
- **AC7:** Batch grouping by task, j/k navigation, ESC to exit

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.6: Pattern Mining & Mood-Execution Analytics

**Status:** draft

As an LLM client,
I want to access productivity analytics including mood-execution correlations, streaks, and burnout risk,
So that I can provide data-driven productivity insights and coaching.

**Acceptance Criteria:**
- **AC1:** `PatternMiner` with methods: `MoodCorrelation`, `ProductivityProfile`, `StreakAnalysis`, `BurnoutRisk`, `WeeklySummary`
- **AC2:** MCP Resources: `threedoors://analytics/mood-correlation`, `/time-of-day`, `/streaks`, `/burnout-risk`, `/task-preferences`, `/weekly-summary`
- **AC3:** MCP Tools: `get_mood_correlation`, `get_productivity_profile`, `burnout_risk`, `get_completions`
- **AC4:** MCP Prompts: `daily_summary`, `weekly_retrospective` templates
- **AC5:** Burnout risk composite score (0-1) from 5+ signals, >0.7 = warning

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.7: Task Relationship Graph & Cross-Provider Linking

**Status:** draft

As an LLM client,
I want to traverse task relationship graphs and discover cross-provider dependencies,
So that I can answer questions about task dependencies across systems.

**Acceptance Criteria:**
- **AC1:** `TaskGraph` with nodes and edges; `EdgeType` constants: blocks, related-to, subtask-of, duplicate-of, sequential, cross-ref
- **AC2:** `RelationshipInferencer` with 6 strategies: text similarity, temporal, cross-ref, blocker chains, subtask patterns, duplicate detection
- **AC3:** MCP Tools: `walk_graph`, `find_paths`, `get_critical_path`, `get_orphans`, `get_clusters`
- **AC4:** `CrossProviderLinker` for cross-provider relationship discovery
- **AC5:** MCP Tools: `get_provider_overlap`, `get_unified_view`, `suggest_cross_links`
- **AC6:** MCP Resources: `threedoors://graph/dependencies`, `threedoors://graph/cross-provider`

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Story 24.8: MCP Prompt Templates & Advanced Interaction Patterns

**Status:** draft

As an LLM client,
I want pre-built prompt templates and advanced tools for prioritization, workload analysis, and what-if modeling,
So that I can provide high-quality coaching with consistent responses.

**Acceptance Criteria:**
- **AC1:** MCP Prompts: `blocked_tasks`, `task_deep_dive`, `weekly_retrospective`
- **AC2:** MCP Tool `prioritize_tasks` with multi-signal scoring (blocking, age, effort fit, mood fit, time-of-day, streak impact)
- **AC3:** MCP Tool `analyze_workload` — total tasks, estimated hours, overload risk, focus recommendations
- **AC4:** MCP Tool `focus_recommendation(mood, available_minutes)` — optimal task sequence
- **AC5:** MCP Tool `what_if(complete_task_ids)` — scenario modeling without mutation
- **AC6:** MCP Tool `context_switch_analysis` — switches/session, cost, optimal batching

**Quality Gate (QG1-QG6):** gofumpt ✓ | golangci-lint ✓ | tests pass ✓ | rebased ✓ | scope-checked ✓ | errors handled ✓

### Epic 24 Story Dependencies

```
24.1 (MCP Server) ──┬──> 24.2 (Read-Only Resources) ──┬──> 24.4 (Proposals) ──> 24.5 (TUI Review)
                     │                                  ├──> 24.6 (Analytics)
                     │                                  └──> 24.7 (Graph)
                     └──> 24.3 (Security Middleware)
                                                        24.6 + 24.7 ──> 24.8 (Advanced Interactions)
```

### MVP Phasing

**Phase 1 (Stories 24.1, 24.2, 24.3):** Read-only MCP server. Claude can see and query tasks. Immediate value — AI-assisted task understanding.

**Phase 2 (Stories 24.4, 24.5):** Proposals + enrichment. LLMs can suggest improvements. Users maintain full control via TUI review.

**Phase 3 (Story 24.6):** Analytics + pattern mining. LLMs provide data-driven productivity insights.

**Phase 4 (Story 24.7):** Relationship graphs + cross-provider linking. LLMs understand task dependencies and cross-system relationships.

**Phase 5 (Story 24.8):** Advanced interactions. LLMs become a personal productivity coach with prioritization, workload analysis, and what-if modeling.

**MVP-3 (Stories 22.6, 22.7, 22.8):** Auto-generated tasks + story generation + safety guardrails. Complete closed-loop self-driving pipeline.

## Epic 23: CLI Interface ✅ COMPLETE

**Epic Goal:** Provide a complete non-TUI CLI interface for ThreeDoors that serves both human power users (scriptable task management) and LLM agents (structured JSON output). The CLI shares `internal/core` with the TUI — no domain logic duplication. `threedoors` with no args launches the TUI (backward compatible); any subcommand routes to the Cobra-based CLI.

**Prerequisites:** None (core domain layer is already CLI-ready with JSON struct tags)
**Framework:** Cobra (`github.com/spf13/cobra`) for subcommand routing, shell completions, and help generation
**Origin:** CLI interface design research (`../../_bmad-output/planning-artifacts/cli-interface-design.md`)
**Architecture:** Layered CLI/TUI coexistence — `internal/cli/` imports `internal/core/`, never `internal/tui/`
**Status:** All 11 stories complete (PRs #161-#192, #225)

**Key Design Decisions:**
- Noun-verb command taxonomy: `threedoors task <verb>` (modeled after `gh` CLI)
- `--json` persistent flag switches all output from human-readable to structured JSON
- JSON envelope with `schema_version: 1` for forward compatibility
- Exit codes 0-5 for machine-parseable error handling
- ID prefix matching for human-friendly task references (like git short SHAs)
- Non-interactive by default — CLI prints and exits, `--interactive` opt-in
- `threedoors doors` is the signature CLI command (equivalent of TUI launch)

### Story 23.1: Cobra Scaffolding, Root Command, and Output Formatter

**Status:** Draft

As a developer,
I want a Cobra-based CLI scaffold with a root command, `--json` persistent flag, and a shared output formatter,
So that all subsequent CLI commands have a consistent foundation for routing, output formatting, and error handling.

**Acceptance Criteria:**

**Given** the need for a CLI interface alongside the existing TUI
**When** the scaffolding is implemented
**Then:**
- AC1: `internal/cli/root.go` defines the Cobra root command with `--json` persistent flag
- AC2: `internal/cli/output.go` defines an `OutputFormatter` supporting human-readable (tabwriter) and JSON modes
- AC3: JSON output uses envelope: `{"schema_version": 1, "command": "<cmd>", "data": ..., "metadata": {...}}`
- AC4: `cmd/threedoors/main.go` updated with subcommand detection — backward compatible TUI launch
- AC5: `go.mod` updated with Cobra dependency
- AC6: Exit code constants defined (0-5)
- AC7: Unit tests for output formatter

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.2: Task List and Task Show Commands with Prefix Matching

**Status:** Draft

As a CLI user,
I want to list tasks with filters and view task details by ID prefix,
So that I can browse and inspect my tasks without launching the TUI.

**Acceptance Criteria:**

**Given** the CLI scaffold from Story 23.1 is in place
**When** task list and show commands are implemented
**Then:**
- AC1: `threedoors task list` displays all active tasks in a table
- AC2: `--status`, `--type`, `--effort` filter flags, composable
- AC3: `threedoors task list --json` with metadata (total, filtered, filters)
- AC4: `threedoors task show <id>` with ID prefix matching
- AC5: `FindByPrefix()` method added to `TaskPool`
- AC6: Exit code 2 (not found), 5 (ambiguous prefix)

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.3: Task Add and Task Complete Commands

**Status:** Draft

As a CLI user,
I want to add new tasks and mark tasks complete from the command line,
So that I can manage my task lifecycle without the TUI.

**Acceptance Criteria:**

**Given** the CLI scaffold from Story 23.1 is in place
**When** task add and complete commands are implemented
**Then:**
- AC1: `threedoors task add "text"` with optional `--context`, `--type`, `--effort`
- AC2: `threedoors task complete <id>` with prefix matching
- AC3: Batch complete: `threedoors task complete <id1> <id2> <id3>`
- AC4: `--json` support for both commands
- AC5: Exit code 2 (not found), 3 (invalid transition)

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.4: Doors Command — CLI Three Doors Experience

**Status:** Draft

As a CLI user or LLM agent,
I want a `threedoors doors` command that presents three randomly selected tasks,
So that I can experience the core Three Doors mechanic without the TUI.

**Acceptance Criteria:**

**Given** the CLI scaffold from Story 23.1 is in place
**When** the doors command is implemented
**Then:**
- AC1: `threedoors doors` displays 3 randomly selected tasks (human-readable)
- AC2: `threedoors doors --json` with door numbers, task data, and metadata
- AC3: Selection uses existing `SelectDoors()` — no logic duplication
- AC4: `threedoors doors --pick N` selects door N and marks task in-progress
- AC5: Non-interactive by default

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.5: Health, Version Commands and Exit Code Enforcement

**Status:** Draft

As a CLI user or LLM agent,
I want `threedoors health` and `threedoors version` commands,
So that I can verify system status and integrate ThreeDoors into health-check scripts.

**Acceptance Criteria:**

**Given** the CLI scaffold from Story 23.1 is in place
**When** health and version commands are implemented
**Then:**
- AC1: `threedoors health` runs `HealthChecker.RunAll()` with table output
- AC2: `threedoors health --json` with overall, duration, checks array
- AC3: Exit code 4 if any health check fails
- AC4: `threedoors version` with version, commit, build date via ldflags
- AC5: Makefile updated for ldflags injection

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.6: Task Block, Unblock, and Status Commands

**Status:** Draft

As a CLI user,
I want to block, unblock, and change task status from the command line,
So that I can manage task state transitions without the TUI.

**Acceptance Criteria:**

**Given** the task list/show commands from Story 23.2 are in place
**When** status management commands are implemented
**Then:**
- AC1: `threedoors task block <id> --reason "..."` with prefix matching
- AC2: `threedoors task unblock <id>` transitions blocked -> todo
- AC3: `threedoors task status <id> <new-status>` for any valid transition
- AC4: Invalid transitions return exit code 3
- AC5: `--json` support for all commands

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.7: Task Edit, Delete, Note, and Search Commands

**Status:** Draft

As a CLI user,
I want to edit, delete, annotate, and search tasks from the command line,
So that I have full task management capability without the TUI.

**Acceptance Criteria:**

**Given** the task list/show commands from Story 23.2 are in place
**When** edit, delete, note, and search commands are implemented
**Then:**
- AC1: `threedoors task edit <id> --text/--context` with prefix matching
- AC2: `threedoors task delete <id>` with batch support
- AC3: `threedoors task note <id> "text"` adds a note
- AC4: `threedoors task search "query"` searches text and context
- AC5: `--json` support for all commands

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.8: Mood and Stats Commands

**Status:** Draft

As a CLI user,
I want to record mood and view productivity statistics from the command line,
So that I can track patterns without the TUI.

**Acceptance Criteria:**

**Given** the CLI scaffold from Story 23.1 is in place
**When** mood and stats commands are implemented
**Then:**
- AC1: `threedoors mood set <mood>` records mood via SessionTracker
- AC2: `threedoors mood history` shows mood entries
- AC3: `threedoors stats` with `--daily`, `--weekly`, `--patterns` flags
- AC4: `--json` support for all commands

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.9: Config Commands and Stdin/Pipe Support

**Status:** Draft

As a CLI user or LLM agent,
I want to view/modify config and pipe task text via stdin,
So that I can script task creation and configure ThreeDoors without editing files.

**Acceptance Criteria:**

**Given** the CLI scaffold and task add command are in place
**When** config commands and stdin support are implemented
**Then:**
- AC1: `threedoors config show/get/set` commands
- AC2: `echo "text" | threedoors task add` reads from stdin
- AC3: `--stdin` flag for multi-line input (one task per line)
- AC4: Config key validation with exit code 3 for unknown keys

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Story 23.10: Shell Completions and Interactive Doors Mode

**Status:** Draft

As a CLI power user,
I want shell completions and an interactive doors selection mode,
So that I can use the CLI efficiently with tab completion.

**Acceptance Criteria:**

**Given** all CLI commands from previous stories are in place
**When** shell completions and interactive mode are implemented
**Then:**
- AC1: `threedoors completion bash/zsh/fish` outputs completion scripts
- AC2: Completions cover all subcommands, flags, and enum values
- AC3: `threedoors doors --interactive` prompts for door selection
- AC4: Interactive mode auto-disabled when stdout is not a TTY

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

### Epic 23 Story Dependencies

```
23.1 (Scaffolding) ──┬──> 23.2 (List/Show) ──┬──> 23.6 (Block/Status)
                      │                        └──> 23.7 (Edit/Delete/Note/Search)
                      ├──> 23.3 (Add/Complete) ──> 23.9 (Config/Stdin)
                      ├──> 23.4 (Doors) ──────────> 23.10 (Completions/Interactive)
                      ├──> 23.5 (Health/Version)
                      └──> 23.8 (Mood/Stats)
```

### MVP Phasing

**Phase 1 — Minimum Viable CLI (Stories 23.1–23.5):** Cobra scaffold + output formatter + core commands (task list/show/add/complete, doors, health, version). Enables both human and LLM usage of ThreeDoors from the command line.

**Phase 2 — Extended CLI (Stories 23.6–23.9):** Full task lifecycle (block/unblock/status/edit/delete/note/search), mood tracking, stats, config management, and stdin/pipe support. Complete parity with TUI task operations.

**Phase 3 — Polish (Story 23.10):** Shell completions and interactive doors mode. Quality-of-life improvements for power users.

### Story 23.11: Fix Nil Pointer Panic on Missing Provider ✅

**Status:** Done (PR #225)

**GitHub Issue:** #218 (closed)

As a ThreeDoors user,
I want the CLI to return a clear error when no provider is available,
So that I don't experience a panic on first run.

**Bug fix:** `loadTaskPool()` in `doors.go` and MCP server init call `NewProviderFromConfig()` which can return nil. Neither checks for nil before dereferencing. Fix: add nil check and return descriptive error in both locations.

- **AC1:** `loadTaskPool()` returns error (not panic) when provider is nil
- **AC2:** Error message is actionable — indicates no provider available
- **AC3:** MCP server handles nil provider without panic
- **AC4:** All existing tests pass
- **AC5:** New tests cover nil provider scenario in `loadTaskPool()`
- **AC6:** New tests cover nil provider scenario in MCP server

**Quality Gate (AC-Q1–Q8):** gofumpt | golangci-lint | tests pass | rebased | scope-checked | errors handled

---

## Epic 25: Todoist Integration

**Priority:** P1 — High value, all infrastructure in place
**Status:** COMPLETE (4/4 stories)
**Dependencies:** Epic 7 (Adapter SDK) COMPLETE, Epic 13 (Multi-Source Aggregation) COMPLETE, Epic 21 (Sync Protocol Hardening) COMPLETE

### Epic Goal

Integrate Todoist as a task source adapter using the REST API v1, enabling ThreeDoors users to view and manage their Todoist tasks through the Three Doors decision-friction interface.

### Requirements Coverage

| Requirement | Story | Description |
|------------|-------|-------------|
| FR89 | 25.1, 25.2 | Todoist REST API v1 integration with field mapping |
| FR90 | 25.1 | API token auth with project/filter config |
| FR91 | 25.2 | Priority-to-Effort scale inversion mapping |
| FR92 | 25.3 | Bidirectional sync with WAL queuing |

### Stories

#### Story 25.1: Todoist HTTP Client & Auth Configuration

**Status:** Not Started | **Priority:** P1 | **Depends On:** Epic 7 (done)

As a developer, I want a thin HTTP client for the Todoist REST API v1, so that the TodoistProvider can read and complete tasks without a third-party SDK dependency.

**Acceptance Criteria:**
- AC1: `Client` struct with `NewClient(config AuthConfig) *Client`
- AC2: Bearer token authentication via personal API token
- AC3: `GetTasks(ctx, projectID, filter)` method for task retrieval
- AC4: `CloseTask(ctx, taskID)` method for task completion
- AC5: `GetProjects(ctx)` method for health check
- AC6: HTTP 429 handling with `Retry-After` header parsing
- AC7: Config struct with mutually exclusive `project_ids` and `filter` validation
- AC8: Unit tests with `httptest.NewServer`
- AC9: No third-party dependencies beyond stdlib

**File:** `docs/stories/25.1.story.md`

---

#### Story 25.2: Read-Only Todoist Adapter with Field Mapping

**Status:** Not Started | **Priority:** P1 | **Depends On:** 25.1

As a ThreeDoors user who uses Todoist, I want my Todoist tasks to appear as doors in the Three Doors TUI.

**Acceptance Criteria:**
- AC1: `TodoistProvider` implementing full `TaskProvider` interface
- AC2: Field mapping: content->Text, description->Context, is_completed->Status
- AC3: Priority-to-Effort inversion: 0->quick-win, 1->quick-win, 2->medium, 3->deep-work, 4->deep-work
- AC4: Deleted tasks (`is_deleted == true`) filtered from results
- AC5: Write methods return `core.ErrReadOnly`
- AC6: `HealthCheck()` verifies API connectivity via `GetProjects()`
- AC7: `[TDT]` source badge in TUI
- AC8: Local cache at `~/.threedoors/todoist-cache.yaml`
- AC9: Factory function for adapter registry registration

**File:** `docs/stories/25.2.story.md`

---

#### Story 25.3: Bidirectional Sync & WAL Integration

**Status:** Not Started | **Priority:** P1 | **Depends On:** 25.2

As a ThreeDoors user, I want completing a Todoist task in ThreeDoors to mark it complete in Todoist.

**Acceptance Criteria:**
- AC1: `MarkComplete(taskID)` calls `CloseTask` via Todoist API
- AC2: On API failure, change queued via `WALProvider`
- AC3: WAL replay on connectivity restoration
- AC4: Rate limit handling with exponential backoff
- AC5: Circuit breaker integration (5 failures / 2min window)
- AC6: Unit tests for success, WAL fallback, and rate limit flows

**File:** `docs/stories/25.3.story.md`

---

#### Story 25.4: Contract Tests & Integration Testing

**Status:** Not Started | **Priority:** P1 | **Depends On:** 25.2

As a developer, I want the Todoist adapter to pass the contract test suite with comprehensive edge case coverage.

**Acceptance Criteria:**
- AC1: `adapters.RunContractTests` passes with mock HTTP server
- AC2: Table-driven priority mapping tests (all 5 values: 0-4)
- AC3: Deleted task filtering tests
- AC4: Config validation tests (mutually exclusive options)
- AC5: Rate limit handling tests
- AC6: Empty response edge case test
- AC7: Special character handling tests
- AC8: Test coverage >= 80% for todoist package

**File:** `docs/stories/25.4.story.md`

---

### Implementation Notes

- **Architecture:** Follows the exact same adapter pattern as Jira (Epic 19) and Apple Reminders (Epic 20)
- **Package:** `internal/adapters/todoist/` — thin HTTP client, provider, field mapping, config
- **No SDK:** Build raw HTTP client against REST API v1 (go-todoist library targets deprecated v2)
- **Priority inversion:** Todoist 4 (critical) maps to highest effort; Todoist 0/1 maps to lowest
- **Read-only first:** Story 25.2 delivers 80% of user value; 25.3 completes the loop
- **Config:** Mutually exclusive `project_ids` and `filter` options for task scoping

---

## Epic 26: GitHub Issues Integration

**Priority:** P1 — High value, all infrastructure in place
**Status:** All 4 stories complete (PRs #201-#205)
**Dependencies:** Epic 7 (Adapter SDK) COMPLETE, Epic 13 (Multi-Source Aggregation) COMPLETE, Epic 21 (Sync Protocol Hardening) COMPLETE

### Epic Goal

Integrate GitHub Issues as a task source adapter using the official `go-github` SDK, enabling ThreeDoors users to view and manage their GitHub issues through the Three Doors decision-friction interface. Target audience overlap is maximum — developers already track work in GitHub Issues.

### Requirements Coverage

| Requirement | Story | Description |
|------------|-------|-------------|
| FR93 | 26.1, 26.2 | GitHub Issues integration with field mapping via go-github SDK |
| FR94 | 26.1 | PAT auth with repo list and assignee filter config |
| FR95 | 26.2 | Label-based priority/status mapping conventions |
| FR96 | 26.3 | Bidirectional sync with WAL queuing |

### Stories

#### Story 26.1: GitHub SDK Client & Auth Configuration ✅

**Status:** Done (PR #201) | **Priority:** P1 | **Depends On:** Epic 7 (done)

As a developer, I want a GitHub API client using the official go-github SDK, so that the GitHubProvider can read and close issues with proper authentication and rate limit handling.

**Acceptance Criteria:**
- AC1: `GitHubClient` struct in `internal/adapters/github/github_client.go` wrapping `go-github` SDK
- AC2: PAT authentication via `GITHUB_TOKEN` env var or config.yaml settings
- AC3: `ListIssues(ctx, repo, assignee) ([]GitHubIssue, error)` method using go-github SDK
- AC4: `CloseIssue(ctx, repo, issueNumber) error` method for issue completion
- AC5: `GetAuthenticatedUser(ctx) (string, error)` method for health check
- AC6: Rate limit handling wrapping go-github's native `*github.RateLimitError` into adapter `RateLimitError` pattern
- AC7: `GitHubConfig` struct with `Token`, `Repos`, `Assignee`, `PollInterval`, `PriorityLabels`, `InProgressLabel` fields
- AC8: Config validation: `repos` list required, `assignee` defaults to `@me`
- AC9: Unit tests using `httptest.NewServer` with canned GitHub API responses

**File:** `docs/stories/26.1.story.md`

---

#### Story 26.2: Read-Only GitHub Provider with Field Mapping ✅

**Status:** Done (PR #202) | **Priority:** P1 | **Depends On:** 26.1

As a ThreeDoors user who tracks work in GitHub Issues, I want my assigned issues to appear as doors in the Three Doors TUI.

**Acceptance Criteria:**
- AC1: `GitHubProvider` implementing full `TaskProvider` interface
- AC2: Field mapping: title->Text, body->Context, state->Status, labels->Tags
- AC3: Status mapping: `open`->`todo`, `closed`->`complete`, `in-progress` label->`in-progress`
- AC4: Label-to-Effort mapping: `priority:critical`->deep-work, `priority:high`->deep-work, `priority:medium`->medium, `priority:low`->quick-win
- AC5: Write methods return `core.ErrReadOnly`
- AC6: `HealthCheck()` verifies API connectivity via `GetAuthenticatedUser()`
- AC7: `[GH]` source badge in TUI
- AC8: Local cache at `~/.threedoors/github-cache.yaml` with configurable TTL
- AC9: Factory function for adapter registry registration
- AC10: Multi-repo aggregation — issues from all configured repos merged into single result

**File:** `docs/stories/26.2.story.md`

---

#### Story 26.3: Bidirectional Sync & WAL Integration ✅

**Status:** Done (PR #204) | **Priority:** P1 | **Depends On:** 26.2

As a ThreeDoors user, I want completing a GitHub issue in ThreeDoors to close it on GitHub.

**Acceptance Criteria:**
- AC1: `MarkComplete(taskID)` calls `CloseIssue` via GitHub API
- AC2: On API failure, change queued via `WALProvider`
- AC3: WAL replay on connectivity restoration
- AC4: Rate limit handling with exponential backoff
- AC5: Circuit breaker integration (5 failures / 2min window)
- AC6: Unit tests for success, WAL fallback, and rate limit flows

**File:** `docs/stories/26.3.story.md`

---

#### Story 26.4: Contract Tests & Integration Testing ✅

**Status:** Done (PR #205) | **Priority:** P1 | **Depends On:** 26.2

As a developer, I want the GitHub adapter to pass the contract test suite with comprehensive edge case coverage.

**Acceptance Criteria:**
- AC1: `adapters.RunContractTests` passes with mock HTTP server
- AC2: Table-driven label-to-priority mapping tests (all combinations)
- AC3: Status mapping tests (open, closed, in-progress label)
- AC4: Multi-repo aggregation tests
- AC5: Rate limit handling tests
- AC6: Empty repo (no issues) edge case test
- AC7: Special character handling tests (issue titles with unicode, markdown bodies)
- AC8: Assignee filtering tests
- AC9: Test coverage >= 80% for github package

**File:** `docs/stories/26.4.story.md`

---

### Implementation Notes

- **Architecture:** Follows the exact same adapter pattern as Jira (Epic 19), Apple Reminders (Epic 20), and Todoist (Epic 25)
- **Package:** `internal/adapters/github/` — SDK client wrapper, provider, field mapping, config
- **SDK:** Use official `google/go-github` (unlike Todoist which uses raw HTTP due to deprecated SDK)
- **Label conventions:** `priority:critical/high/medium/low` for effort mapping; `in-progress` for status enrichment
- **Read-only first:** Story 26.2 delivers 80% of user value; 26.3 completes the loop
- **Config:** Explicit repo list required; org-level queries deferred to future enhancement
- **Schedule:** Recommended after Epic 25 (Todoist) to leverage learnings

---

## Epic 27: Daily Planning Mode

**Priority:** P1
**Status:** COMPLETE
**Dependencies:** Epic 1 (session tracking) COMPLETE, Epic 3 (mood capture) COMPLETE, Epic 4 (task categorization) COMPLETE

### Epic Goal

Add a guided daily planning ritual that transforms ThreeDoors from a reactive task picker into a proactive morning engagement tool, driving long-term retention through structured planning sessions.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 27.1 | Planning Data Model & Focus Tag | Done (PR #323) | P1 | Epic 1 (done) |
| 27.2 | Review Incomplete Tasks Flow | Done (PR #339) | P1 | 27.1 |
| 27.3 | Focus Selection Flow | Done (PR #352) | P1 | 27.1 |
| 27.4 | Energy Level Matching & Time-of-Day Inference | Done (PR #354) | P1 | 27.1 |
| 27.5 | Planning Session Metrics & CLI/TUI Commands | Done (PR #360) | P1 | 27.1-27.4 |

**FRs covered:** FR97-FR103
**Research:** See `../../_bmad-output/planning-artifacts/ux-workflow-improvements-research.md`

---

## Epic 28: Snooze/Defer as First-Class Action

**Priority:** P1
**Status:** Not Started
**Dependencies:** None

### Epic Goal

Surface existing `StatusDeferred` as a first-class user action with date-based snooze and auto-return.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 28.1 | DeferUntil Field, Status Transitions, and Auto-Return Logic | Not Started | P1 | None |
| 28.2 | Snooze TUI View and Z-Key Binding | Not Started | P1 | 28.1 |
| 28.3 | Deferred List View and :deferred Command | Not Started | P1 | 28.1 |
| 28.4 | Session Metrics Logging for Snooze Events | Not Started | P1 | 28.1 |

---

## Epic 29: Task Dependencies & Blocked-Task Filtering

**Priority:** P1
**Status:** In Progress (3/4 stories done; 29.3 remaining)
**Dependencies:** None

### Epic Goal

Native dependency graph support. Blocks tasks with unmet dependencies from door selection.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 29.1 | DependsOn Field, DependencyResolver, and YAML Persistence | Done (PR #307) | P1 | None |
| 29.2 | Door Selection Filter and Auto-Unblock on Completion | Done (PR #319) | P1 | 29.1 |
| 29.3 | TUI Blocked-By Indicator and Dependency Management | In Review | P1 | 29.1 |
| 29.4 | Session Metrics Logging for Dependency Events | Done (PR #356) | P1 | 29.1 |

---

## Epic 30: Linear Integration

**Priority:** P2
**Status:** Not Started
**Dependencies:** Epic 7 (Adapter SDK) COMPLETE, Epic 13 (Multi-Source Aggregation) COMPLETE

### Epic Goal

Integrate Linear as a task source for engineering teams via the Linear GraphQL API, leveraging Linear's excellent task model alignment for high-fidelity task import.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 30.1 | Linear GraphQL Client & Auth Configuration | In Review | P2 | Epic 7 (done) |
| 30.2 | Read-Only Linear Provider with Field Mapping | Not Started | P2 | 30.1 |
| 30.3 | Bidirectional Sync & WAL Integration | Not Started | P2 | 30.2 |
| 30.4 | Contract Tests & Integration Testing | Not Started | P2 | 30.2 |

**FRs covered:** FR116-FR119
**Research:** See `../../_bmad-output/planning-artifacts/task-source-expansion-research.md`

---

## Epic 31: Expand/Fork Key Implementations

**Priority:** P2
**Status:** Not Started
**Dependencies:** None

### Epic Goal

Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 31.1 | Task Model ParentID Extension | Not Started | P2 | None |
| 31.2 | Enhanced Expand — Sequential Subtask Creation | Not Started | P2 | 31.1 |
| 31.3 | Subtask List Rendering in Detail View | Not Started | P2 | 31.1, 31.2 |
| 31.4 | Enhanced Fork — Variant Creation with ForkTask Factory | Not Started | P2 | None |
| 31.5 | Design Decision H9 Status Update | Not Started | P2 | 31.1-31.4 |

**FRs covered:** FR120-FR126

---

## Epic 32: Undo Task Completion

**Priority:** P1
**Status:** Complete
**Dependencies:** None

### Epic Goal

Allow reversing accidental task completion via `complete -> todo` transition. Validated pain point from Phase 1 gate.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 32.1 | Status Model — Complete-to-Todo Transition | Done (PR #306) | P1 | None |
| 32.2 | Session Metrics — Undo Complete Event Logging | Done (PR #322) | P1 | 32.1 |
| 32.3 | TUI & CLI Undo Experience | Done (PR #337) | P1 | 32.1, 32.2 |

**FRs covered:** FR127-FR131

---

## Epic 40: Beautiful Stats Display

**Priority:** P1 (Phase 1-2), P2 (Phase 3)
**Status:** COMPLETE (10/10 stories done)
**Dependencies:** None (all data infrastructure exists from Epics 1 and 4)

### Epic Goal

Transform the insights dashboard from plain text into a visually delightful, SOUL-aligned celebration of user activity. Three phases: visual polish (Lipgloss panels, gradient sparklines, fun facts), new visualizations (bar charts, heatmap, animated counters, hidden metrics), and thematic integration (theme-matched colors, milestone celebrations).

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 40.1 | Stats Dashboard Shell with Lipgloss Panels | Done (PR #343) | P1 | None |
| 40.2 | Gradient Sparkline with Color-Blind Safe Palette | Done (PR #366) | P1 | 40.1 |
| 40.3 | Fun Facts Engine | Done (PR #371) | P1 | 40.1 |
| 40.4 | Horizontal Bar Charts for Mood Correlation | Done (PR #362) | P1 | 40.1 |
| 40.5 | GitHub-Style Activity Heatmap | Done (PR #391) | P2 | 40.1, 40.8 |
| 40.6 | Surface Hidden Session Metrics | Done (PR #368) | P1 | 40.1 |
| 40.7 | Animated Counter Reveals | Done (PR #392) | P2 | 40.1 |
| 40.8 | Tab Navigation for Detail View | Done (PR #367) | P1 | 40.1 |
| 40.9 | Theme-Matched Stats Color Palettes | Done (PR #380) | P2 | 40.1, 40.2, Epic 17 |
| 40.10 | Milestone Celebrations | Done (PR #393) | P2 | 40.1 |

### Phase Breakdown

**Phase 1 — Visual Polish (Stories 40.1-40.3):** Lipgloss bordered panels, responsive 2-column layout, hero number, gradient sparklines with color-blind safe palette, fun facts engine with celebration-oriented session insights.

**Phase 2 — New Visualizations (Stories 40.4-40.8):** Horizontal bar charts for mood correlation, GitHub-style 8-week activity heatmap, surfaced hidden session metrics, animated counter reveals, Tab navigation for Overview/Detail views.

**Phase 3 — Thematic Integration (Stories 40.9-40.10):** Theme-matched stats color palettes extending ThemeColors, milestone celebrations with observation language (4 thresholds, no gamification).

### Design Decisions

- D-096: Custom Lipgloss rendering, no new dependencies in Phase 1
- D-097: Single view (Phase 1), Tab navigation (Phase 2)
- D-098: Phase 1 = 3 stories (dashboard shell, sparklines, fun facts)
- D-099: Fun facts content rules (observe, celebrate, frame gaps, no decline)
- D-100: Heatmap: 8 weeks, custom Unicode+color, Detail tab
- D-101: Animated counters: subtle 300-500ms, numbers only, once per entry
- D-102: Independent palette (Phase 1), theme coupling (Phase 3)
- D-103: Trophy room deferred; milestones limited to 4 observations
- D-104: Epic number 40

### Research

- Research: `_bmad-output/planning-artifacts/beautiful-stats-research.md`
- UX Review: `_bmad-output/planning-artifacts/beautiful-stats-ux-review.md`
- Party Mode: `_bmad-output/planning-artifacts/beautiful-stats-party-mode.md`
- Architecture: `_bmad-output/planning-artifacts/architecture-beautiful-stats.md`

---

## Epic 41: Charm Ecosystem Adoption & TUI Polish

**Priority:** P2
**Status:** Not Started
**Dependencies:** None hard. Stories 41.3-41.4 should be sequenced after Epic 39 keybinding overlay work to avoid merge conflicts.

### Epic Goal

Systematically adopt underutilized charmbracelet ecosystem components to reduce custom code maintenance, improve UX consistency, and deliver on SOUL.md's "physical objects" promise. Prioritizes replacements (viewport, layout) over additions (spinner, harmonica).

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 41.1 | Spinner Component for Async Provider Operations | Not Started | P2 | None |
| 41.2 | Lipgloss Layout Utilities Adoption | Not Started | P2 | None |
| 41.3 | Viewport Adoption for Help View | Not Started | P2 | None (sequence after Epic 39) |
| 41.4 | Viewport Adoption for Synclog and Keybinding Overlay | Not Started | P2 | 41.3 |
| 41.5 | Harmonica Door Transition Spike | Not Started | P2 | None |
| 41.6 | Adaptive Color Profile Support | Not Started | P2 | None |

### Story Details

#### Story 41.1: Spinner Component for Async Provider Operations

Add bubbles/spinner for sync and provider loading feedback. When Todoist sync, Apple Notes fetch, or other provider operations are in flight, show a spinner. Replace "no feedback" state with clear activity indication.

**AC:** Spinner visible during provider sync (>100ms), stops on completion/error, deterministic golden file testing, no spinner for instant operations.

#### Story 41.2: Lipgloss Layout Utilities Adoption

Adopt lipgloss.JoinVertical and Place for cleaner layout composition. Replace manual `\n` concatenation with JoinVertical, manual centering with Place. Pure refactor — no functional change.

**AC:** Manual vertical joins replaced, centered content uses Place, all golden tests pass, no behavioral change.

#### Story 41.3: Viewport Adoption for Help View

Replace custom page-based scrolling in help_view with bubbles/viewport. First of three viewport migrations. Adds mouse wheel scrolling and continuous (vs page-based) navigation.

**AC:** Viewport used for help content, mouse wheel works, j/k/up/down preserved, PgUp/PgDn supported, golden tests updated, no content regression.

#### Story 41.4: Viewport Adoption for Synclog and Keybinding Overlay

Complete viewport migration for synclog_view and keybinding_overlay. Create shared NewScrollableView() factory for ThreeDoors viewport defaults. After this, zero custom scroll implementations remain.

**AC:** Both views use viewport, shared factory function exists, mouse wheel works in both, consistent scroll behavior across all three migrated views, golden tests updated.

#### Story 41.5: Harmonica Door Transition Spike

Research spike: spring-physics door selection animation with harmonica. Proof of concept on door selection → detail view transition. Primary deliverable is documented testing pattern for frame-based animations, not the animation itself.

**AC:** harmonica in go.mod, POC transition works, testing pattern documented, golden file strategy documented, performance verified, go/no-go decision recorded.

#### Story 41.6: Adaptive Color Profile Support

Terminal-aware color degradation via lipgloss color profiles. Replace hardcoded 256-color values with adaptive definitions that work across TrueColor, ANSI256, ANSI, and ASCII terminals.

**AC:** Profile detected at startup, graceful degradation in 16-color terminals, theme colors adapt, no change on modern terminals, existing golden tests pass.

### Design Decisions

- D-128: Adopt bubbles/viewport to replace 3 custom scroll implementations
- D-129: Adopt bubbles/spinner for async operation feedback
- D-130: Adopt lipgloss.JoinVertical + Place for layout composition
- D-131: Harmonica door transitions via spike-first approach
- D-132: Reject bubbles/list for ThreeDoors selection UIs (3-door constraint intentional)
- D-133: Reject textarea, table, filepicker, timer, help, huh, wish (contradicts SOUL.md)
- D-134: Epic number 41

### Research

- Audit & Party Mode: `_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md`

---

## Epic 48: Door-Like Doors — Visual Door Metaphor Enhancement

**Priority:** P2
**Status:** Complete (4/4 done)
**Dependencies:** Epic 35 (Door Visual Appearance — complete), Epic 17 (Door Theme System — complete), Story 41.5 (Harmonica spike — complete, D-136 GO decision)

### Epic Goal

Transform rectangular card/panel doors into visually convincing doors by implementing 5 adopted proposals from the "doors more door-like" party mode research (5 rounds, 7 agents). The key insight: a door is defined by **asymmetry** (hinge vs opening side), **hardware** (a handle you can grab), **behavior** (it opens/closes), and **context** (it exists on a floor). Current rendering captures almost none of these.

The 5 adopted proposals collectively raise "doorness" from ~3.5/7 (Classic) to ~6.5/7 across all themes, at a total cost of ~200-250 LOC and 0-1 character of width.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 48.1 | Side-Mounted Handle + Hinge Marks | Done (PR #451) | P2 | Epic 35 (done), Epic 17 (done) |
| 48.2 | Continuous Threshold / Floor Line | Done (PR #483) | P2 | None |
| 48.3 | Crack of Light Effect on Selection | Done (PR #572) | P2 | 48.1 |
| 48.4 | Handle Turn Micro-Animation | Done (PR #588) | P2 | 48.1 |

**Dependency graph:**
```
Story 48.1 (Handle + Hinge) ──┬──→ Story 48.3 (Crack of Light)
                               └──→ Story 48.4 (Handle Turn)
Story 48.2 (Threshold) ──────────→ (independent)
```

Stories 48.1 & 48.2 can parallelize. Stories 48.3 & 48.4 can parallelize after 48.1.

### Story Details

#### Story 48.1: Side-Mounted Handle + Hinge Marks (Proposals B + G)

Move doorknob from centered inline content to right edge at HandleRow. Add double-weight box-drawing characters on left border for hinge visual asymmetry. Update all 4 themes with per-theme hinge/handle characters. Extend DoorAnatomy with HingeCol. Zero width cost.

**AC:** Handle at right edge in all themes, hinge marks (heavier left border) in all themes, DoorAnatomy has HingeCol, minimum width (15 chars) still works, golden files updated.

#### Story 48.2: Continuous Threshold / Floor Line (Proposal F)

Render continuous floor line after `JoinHorizontal()` in doors_view.go spanning full width. Uses `▔` or `─` character. Respects theme colors. ~15 LOC, very low risk.

**AC:** Threshold line spans full door row width, uses theme colors, appears below shadow layer, renders at various widths.

#### Story 48.3: Crack of Light Effect on Selection (Proposal C)

Highest-scoring proposal (15/15). On selection, replace right border chars with crack chars (╎) and append shade (░). Synchronize with spring animation emphasis. Reverse on deselect. 1 character width cost when selected.

**AC:** Crack effect on selection, reversed on deselect, synced with spring emphasis, minimum width works with crack, golden files updated.

#### Story 48.4: Handle Turn Micro-Animation (Proposal D)

4-frame handle character sequence synced with spring emphasis: ● (0.0) → ◐ (0.3) → ○ (0.6+). Reverse on deselect: ○ → ◑ → ●. ~20 LOC.

**AC:** Handle animates on selection, reverses on deselect, deterministic frame selection based on emphasis value, all themes support animated handle.

### Design Decisions

- D-141: Adopt 5 proposals (B, C, D, F, G) for door-like doors; reject 4 (A, E, H, I)
- X-080: Reject Nested Frame (A) — high width cost
- X-081: Reject Wall Context (E) — 4 chars/side too expensive
- X-082: Reject Door Swing (H) — high complexity, friction
- X-083: Reject Light Spill (I) — achievable more simply via Crack of Light

### Noted Opportunities (Out of Scope)

1. Door-Opening Transition Animation (perspective shrink on Enter → detail view; ~200+ LOC, separate spike)
2. Light Spill Enhancement (warm gradient layered on Crack of Light; future polish story)

---

## Epic 43: Connection Manager Infrastructure

**Priority:** P1
**Status:** COMPLETE (6/6 stories done)
**Dependencies:** Epic 7 (Adapter SDK — complete), Epic 11 (Sync Observability — complete)

### Epic Goal

Build the connection lifecycle layer for data source integrations. ThreeDoors has 8 working adapters but no infrastructure for users to set up, monitor, manage, and troubleshoot their data source connections. The ConnectionManager provides state machine lifecycle, secure credential storage, config schema for named multi-instance connections, CRUD operations, sync event logging, and migration of all existing adapters to the new pattern.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 43.1 | Connection State Machine and ConnectionManager Type | Done (PR #428) | P1 | None |
| 43.2 | Keyring Integration with Environment Variable Fallback | Done (PR #442) | P1 | None |
| 43.3 | Config Schema v3 Migration with Connections Support | Done (PR #467) | P1 | None |
| 43.4 | Connection CRUD Operations | Done (PR #526) | P1 | 43.1, 43.2, 43.3 |
| 43.5 | Sync Event Logging Infrastructure | Done (PR #439) | P1 | None |
| 43.6 | Migrate Existing Adapters to ConnectionManager Pattern | Done (PR #540) | P1 | 43.1-43.5 |

**Dependency graph:**
```
43.1 (State Machine) ──┐
43.2 (Keyring)      ───┼──→ 43.4 (CRUD) ──┐
43.3 (Config v3)    ───┘                    ├──→ 43.6 (Adapter Migration)
43.5 (Sync Log)     ───────────────────────┘
```

Stories 43.1, 43.2, 43.3, and 43.5 can parallelize. Story 43.4 depends on 43.1-43.3. Story 43.6 depends on all.

### Story Details

#### Story 43.1: Connection State Machine and ConnectionManager Type

ConnectionState enum with 7 states (Disconnected, Connecting, Connected, Syncing, Error, AuthExpired, Paused) and validated transition table. Connection struct with ULID ID, provider name, label, state, sync metadata. ConnectionManager with thread-safe CRUD.

**AC:** State transitions validated, invalid transitions return error, concurrent access safe with RWMutex, List() returns all connections with current state.

#### Story 43.2: Keyring Integration with Environment Variable Fallback

CredentialStore interface with priority chain: env vars → system keychain → encrypted file fallback. Uses `99designs/keyring`. Credentials never in config.yaml. Env var patterns: `THREEDOORS_<PROVIDER>_TOKEN`, `THREEDOORS_CONN_<LABEL_SLUG>_TOKEN`. Respects `GH_TOKEN`/`GITHUB_TOKEN`.

**AC:** Credentials stored in keychain, env vars override keychain, connection-specific env vars override provider-level, headless fallback works, credentials masked in display.

#### Story 43.3: Config Schema v3 Migration with Connections Support

Config schema v3 adds `connections:` array. Auto-migration from v2 single-provider format. Credentials excluded from config.yaml. Atomic save pattern.

**AC:** v2 configs auto-migrate to v3, multiple connections with same provider supported, no credentials in saved YAML, backward compatible.

#### Story 43.4: Connection CRUD Operations

Complete lifecycle operations: Add (with credential storage + config persistence), Remove (with keepTasks option + credential cleanup), Pause/Resume, TestConnection (health check), ForceSync.

**AC:** All CRUD operations validated, credentials cleaned up on remove, health check returns API/token/rate-limit status.

#### Story 43.5: Sync Event Logging Infrastructure

JSONL sync log per connection. SyncEvent types: sync_complete, sync_error, conflict. Rolling retention of last 1000 events. Queryable by connection, time, severity.

**AC:** Events logged per sync cycle, errors and conflicts captured, reader supports filtering and limit, rolling retention works.

#### Story 43.6: Migrate Existing Adapters to ConnectionManager Pattern

Bridge all 8 existing adapters to ConnectionManager. Wrap sync cycles to emit SyncEvents. Integrate pause/resume with sync scheduler. Preserve backward compatibility for legacy single-provider configs.

**AC:** All adapters work through ConnectionManager, legacy configs work, health checks mapped, pause/resume stops/starts sync.

### Design Decisions

- D-147: Use `99designs/keyring` for credential storage (over `zalando/go-keyring`)
- D-149: Compiled-in providers (not pluggable) — Registry+Factory pattern
- D-152: Named connections with ULID IDs for multi-instance support

### Research

- Full lifecycle research: `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`

---

## Epic 44: Sources TUI

**Priority:** P1
**Status:** In Progress (6/7 done)
**Dependencies:** Epic 43 (Connection Manager Infrastructure)

### Epic Goal

TUI interfaces for data source management: a 4-step setup wizard (`:connect`), sources dashboard (`:sources`), source detail view with health checks, sync log view, status bar health alerts, disconnection flow with task preservation, and re-authentication flow.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 44.1 | Setup Wizard with huh Forms | Done (PR #574) | P1 | Epic 43 |
| 44.2 | Sources Dashboard View | Done (PR #553) | P1 | Epic 43 |
| 44.3 | Source Detail View | Done (PR #563) | P1 | 44.2 |
| 44.4 | Sync Log View | Done (PR #582) | P1 | 43.5 |
| 44.5 | Status Bar Integration for Connection Health Alerts | Done (PR #562) | P1 | Epic 43 |
| 44.6 | Disconnection Flow with Task Preservation Options | Done (PR #581) | P1 | 44.2 |
| 44.7 | Re-Authentication Flow | Not Started | P1 | 44.3, Epic 46 |

### Story Details

#### Story 44.1: Setup Wizard with huh Forms

4-step wizard using `charmbracelet/huh`: (1) Provider selection, (2) Provider-specific config (API token, OAuth, or local path), (3) Sync configuration (mode, filters, poll interval), (4) Test connection and confirm. `:connect` command triggers wizard.

**AC:** All 4 steps render, provider-adaptive forms work, Esc cancels without changes, connection test runs before confirmation.

#### Story 44.2: Sources Dashboard View

`:sources` command opens list view with status indicators (●/○/⚠/✗), labels, sync times, task counts. Keybindings: a (add), d (disconnect), p (pause/resume), r (re-sync), t (test), Enter (detail), Esc (back).

**AC:** All connections listed with status indicators, keybindings dispatch correct actions, empty state handled.

#### Story 44.3: Source Detail View

Full metadata display: status, sync time, task counts, provider settings, health checks (✓/✗ for API, token, rate limit, cache). Actions: edit, re-auth, pause, disconnect, view log.

**AC:** All metadata displayed, health checks shown, keybindings work.

#### Story 44.4: Sync Log View

Scrollable sync event history per connection using bubbles/viewport. Events show timestamp, status indicator (✓/⚠/✗), description.

**AC:** Events displayed in reverse chronological order, viewport scrolls, empty state shows message.

#### Story 44.5: Status Bar Integration for Connection Health Alerts

Non-intrusive alert in doors view when connections need attention (auth expired, persistent errors). Only most critical alert shown with `:sources` hint.

**AC:** No alert when healthy, yellow warning for auth expired, error count for multiple issues.

#### Story 44.6: Disconnection Flow with Task Preservation Options

Confirmation dialog: "Keep tasks locally" vs "Remove synced tasks". Uses huh forms. Cleans up credentials.

**AC:** Confirmation required, keep/remove options work, Esc cancels, credentials deleted.

#### Story 44.7: Re-Authentication Flow

Re-auth for expired tokens without disconnect/reconnect. API token: masked input for new token. OAuth: triggers device code flow (Epic 46). On success, connection transitions back to Connected.

**AC:** API token re-entry works, connection test verifies new token, failure keeps AuthExpired state.

### Design Decisions

- D-150: Use `charmbracelet/huh` for setup wizard forms

### Research

- Full UX research: `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`

---

## Epic 45: Sources CLI

**Priority:** P1
**Status:** Complete (5/5 done)
**Dependencies:** Epic 43 (Connection Manager Infrastructure), Epic 23 (CLI Interface — complete)

### Epic Goal

Non-interactive CLI commands for data source management. Supports both human power users (table output) and automation (JSON output via `--json` flag). Consistent with existing CLI patterns from Epic 23.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 45.1 | `threedoors connect` Command (Non-Interactive) | Done (PR #573) | P1 | Epic 43 |
| 45.2 | `threedoors sources` List/Status/Test Commands | Done (PR #550) | P1 | Epic 43 |
| 45.3 | `threedoors sources` Management Commands | Done (PR #587) | P1 | Epic 43 |
| 45.4 | `threedoors sources log` Command | Done (PR #565) | P1 | 43.5 |
| 45.5 | JSON Output Support for All Sources Commands | Done (PR #589) | P1 | 45.1-45.4 |

### Story Details

#### Story 45.1: `threedoors connect` Command (Non-Interactive)

`threedoors connect <provider>` with flags: `--label`, `--token`, provider-specific flags (`--repos`, `--server`, `--lists`, `--path`). No flags → launches interactive wizard.

**AC:** Each provider connectable via flags, connection test runs, error for missing flags, no-flag mode launches wizard.

#### Story 45.2: `threedoors sources` List/Status/Test Commands

`sources` (table list), `sources status <name>` (detailed), `sources test <name>` (health check with ✓/✗). All support `--json`.

**AC:** Table output formatted, detailed status includes health checks, test exit code reflects health, JSON output valid.

#### Story 45.3: `threedoors sources` Management Commands

`sources pause/resume/sync/reauth/edit/disconnect <name>`. Disconnect has `--keep-tasks` flag (interactive prompt without it).

**AC:** Each command validates state, disconnect prompts interactively without flag.

#### Story 45.4: `threedoors sources log` Command

`sources log <name>` with `--last N` and `--errors` flags. Shows recent sync events.

**AC:** Events displayed with timestamps, `--last` limits count, `--errors` filters, JSON output works.

#### Story 45.5: JSON Output Support for All Sources Commands

Consistent JSON serialization across all sources subcommands. Error envelope: `{"error":"message"}`.

**AC:** All commands produce valid JSON with `--json`, warnings to stderr, error envelope consistent.

### Research

- CLI design: `_bmad-output/planning-artifacts/data-source-setup-ux-research.md` (Section 4)

---

## Epic 46: OAuth Device Code Flow

**Priority:** P2
**Status:** Not Started
**Dependencies:** None (consumed by Epics 44/45 for OAuth-supporting providers)

### Epic Goal

Generic, reusable OAuth device code flow client (RFC 8628) for browser-based authentication. Provider-specific integrations for GitHub and Linear. Silent token refresh with explicit re-auth on expiry.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 46.1 | Generic Device Code Flow Client | Not Started | P2 | None |
| 46.2 | GitHub OAuth Integration | Not Started | P2 | 46.1 |
| 46.3 | Linear OAuth Integration | Not Started | P2 | 46.1 |
| 46.4 | Token Refresh Lifecycle | Not Started | P2 | 46.1 |

### Story Details

#### Story 46.1: Generic Device Code Flow Client

Provider-agnostic device code flow: request device code, display user code + URL, open browser, poll for token with backoff, exchange for access/refresh tokens. Uses `context.Context` for cancellation.

**AC:** Successful flow returns tokens, timeout handled, slow_down backoff works, cancellation stops polling.

#### Story 46.2: GitHub OAuth Integration

Wire device code client to GitHub endpoints. Scope: `repo`. Fallback: `GH_TOKEN`/`GITHUB_TOKEN` env vars, PAT entry.

**AC:** Device code flow works with GitHub, env var fallback offered, PAT fallback for Enterprise.

#### Story 46.3: Linear OAuth Integration

API key as primary auth (simpler). OAuth 2.0 authorization code as secondary if available. Validates via Linear viewer endpoint.

**AC:** API key auth works, connection test verifies via viewer endpoint.

#### Story 46.4: Token Refresh Lifecycle

Pre-emptive refresh when token within 5 min of expiry. Refresh failure → AuthExpired state. 401 detection for API key connections.

**AC:** Silent refresh before expiry, AuthExpired on refresh failure, 401 triggers AuthExpired, events logged.

### Design Decisions

- D-148: OAuth device code flow (not callback server) — no port conflicts, works in SSH/containers

### Research

- OAuth patterns: `_bmad-output/planning-artifacts/data-source-setup-ux-research.md` (Section 5.5)

---

## Epic 47: Sync Lifecycle & Advanced Features

**Priority:** P2
**Status:** Not Started
**Dependencies:** Epic 43 (Connection Manager Infrastructure), Epic 44 (Sources TUI)

### Epic Goal

Advanced sync features: field-level conflict resolution with logging, orphaned task handling (mark not auto-delete), auto-detection of installed tools in setup wizard, and proactive connection health notifications.

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 47.1 | Conflict Resolution Strategy with Logging | Not Started | P2 | Epic 43 |
| 47.2 | Orphaned Task Handling | Not Started | P2 | 47.1 |
| 47.3 | Auto-Detection of Existing Tools in Setup Wizard | Not Started | P2 | 44.1 |
| 47.4 | Proactive Connection Health Notifications | Not Started | P2 | 44.5 |

### Story Details

#### Story 47.1: Conflict Resolution Strategy with Logging

Last-writer-wins with field-level strategy: remote-wins for metadata (status, priority), local-wins for ThreeDoors fields (effort category, door assignment). All conflicts logged. Never auto-delete.

**AC:** Metadata conflicts resolved with remote-wins, ThreeDoors fields local-wins, conflicts logged with both values, remote deletions mark tasks as orphaned.

#### Story 47.2: Orphaned Task Handling

Tasks deleted remotely are flagged as orphaned (not auto-deleted). Excluded from door selection. `:orphaned` command lets users keep (convert to local) or delete.

**AC:** Orphaned tasks flagged, excluded from doors, keep/delete actions work.

#### Story 47.3: Auto-Detection of Existing Tools in Setup Wizard

Detect `gh` CLI, `TODOIST_API_TOKEN`, `.obsidian/` directories, Jira configs. Detected tools appear at top of provider list with badges and pre-filled settings.

**AC:** Detected tools shown at top with badges, pre-fill works, no detections shows default order.

#### Story 47.4: Proactive Connection Health Notifications

Predictive warnings: token expiring within 7 days, rate limit >80%, 3+ consecutive sync errors. Non-intrusive status bar alerts.

**AC:** Token expiry warning, rate limit warning, error streak warning, no warnings when healthy.

### Design Decisions

- D-151: Last-writer-wins conflict resolution with field-level strategy

### Research

- Sync patterns: `_bmad-output/planning-artifacts/data-source-setup-ux-research.md` (Section 5.3)

---

## Epic 42: Application Security Hardening

**Priority:** P1
**Status:** Not Started
**Dependencies:** None

### Epic Goal

Remediate all actionable findings from the application security audit — standardize file permissions, add symlink validation, enforce input size limits, protect credentials, and harden CI supply chain.

### Stories

**Stories:** 42.1-42.5 (5 stories)

#### Story 42.1: File Permission Standardization (0o700/0o600)

Standardize all file and directory permissions across the codebase. Change `~/.threedoors/` from `0o755` to `0o700`, all data files from `0o644` to `0o600`, and temp files from `os.Create()` to `os.OpenFile(..., 0o600)`. Add startup migration to fix existing permissive directories. Addresses HIGH-1, MEDIUM-3, LOW-2.

**AC:** Directory created with 0o700, all data files 0o600, temp files 0o600, existing directories migrated on startup, consistent across all subsystems.

#### Story 42.2: Symlink Validation for File Operations

Add symlink detection before all file read/write operations. Verify `~/.threedoors/` is not a symlink on startup via `os.Lstat()`. Check file paths before atomic writes. Optionally check directory ownership. Create reusable `ValidatePath()` helper modeled on Obsidian adapter's `sanitizeDailyNotePath()`. Addresses HIGH-2.

**AC:** Startup rejects symlinked data directory, writes reject symlinked target files, ownership check warns on mismatch, reusable validation helper exists.

#### Story 42.3: Input Size Limits for YAML and JSONL Readers

Add file size validation before YAML reads (10MB limit) and explicit scanner buffer limits to all JSONL readers. Create `ReadFileWithLimit()` helper. Set scanner buffer to 1MB max (matching MCP transport). Addresses MEDIUM-1, MEDIUM-2, LOW-3.

**AC:** Oversized YAML files rejected with clear error, all JSONL scanners have explicit buffer limits, oversized lines logged and skipped gracefully.

#### Story 42.4: Credential Protection in Config Files

Warn on startup if config.yaml is world-readable and contains non-empty token fields. Ensure all credential struct fields use `yaml:"-"` to prevent accidental serialization. Improve sample config documentation with env var recommendations. Addresses MEDIUM-4.

**AC:** Warning on startup for exposed credentials, yaml:"-" on all token fields, sample config recommends env vars for all credentials.

#### Story 42.5: CI Supply Chain Hardening

Pin third-party GitHub Actions (golangci-lint-action, paths-filter, goreleaser-action) to SHA hashes with version comments. Add `govulncheck ./...` as a CI quality gate step. Verify Dependabot compatibility with SHA-pinned actions.

**AC:** Third-party actions SHA-pinned, first-party actions remain tag-pinned, govulncheck runs in CI and blocks on findings, Dependabot can still propose updates.

### Design Decisions

- D-153: New Epic 42 for application security hardening (5 stories from security audit)

### Research

- Security Audit: `_bmad-output/planning-artifacts/security-audit-application.md`

---

## Epic 49: ThreeDoors Doctor — Self-Diagnosis Command

**Priority:** P1
**Status:** Not Started
**Dependencies:** Epic 23 (CLI Interface — complete)

### Epic Goal

Comprehensive self-diagnosis command (`threedoors doctor`) with flutter-style category-based output, conservative auto-repair via `--fix`, and channel-aware version checking. Supersedes existing `health` command (`internal/cli/health.go`).

### Stories

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 49.1 | Doctor Command Skeleton & Health Alias | Done (PR #444) | P1 | None |
| 49.2 | Environment Checks | Done (PR #473) | P1 | 49.1 |
| 49.3 | Task Data Integrity Checks | Done (PR #471) | P1 | 49.1 |
| 49.4 | Provider Health Checks | Done (PR #475) | P1 | 49.1 |
| 49.5 | Session & Analytics Checks | Done (PR #474) | P1 | 49.1 |
| 49.6 | Sync & Offline Queue Checks | Done (PR #472) | P1 | 49.1 |
| 49.7 | Enrichment Database Checks | Done (PR #470) | P1 | 49.1 |
| 49.8 | Auto-Repair (`--fix` flag) | Done (PR #529) | P1 | 49.2-49.7 |
| 49.9 | Channel-Aware Version Checking | Done (PR #476) | P1 | 49.1 |
| 49.10 | Verbose Mode, Category Filter & Polish | Done (PR #530) | P1 | 49.2-49.9 |

**Dependency graph:**
```
49.1 (skeleton) → 49.2-49.7 (check categories, parallelizable)
49.2-49.7 → 49.8 (--fix needs checks)
49.1 → 49.9 (version check, independent)
49.2-49.9 → 49.10 (polish)
```

### Story Details

#### Story 49.1: Doctor Command Skeleton & Health Alias

Create `internal/cli/doctor.go` with cobra command, `internal/core/doctor.go` with `DoctorChecker` struct and result types. Implement category-based output with flutter-style icons (`[✓]`/`[!]`/`[✗]`/`[i]`/`[ ]`). Register `doctor` as primary, `health` as alias. Support `--json` output. Implement Environment category only.

**AC:** `threedoors doctor` runs and shows Environment category. `threedoors health` is an alias. `--json` produces valid JSON with envelope pattern.

#### Story 49.2: Environment Checks

Expand environment checks: config directory existence/permissions, config file YAML validation + schema version, terminal capability detection (size, color profile), Go runtime version.

**AC:** Environment category shows config dir, config file, terminal info. Missing dir = FAIL. Bad schema = FAIL. Narrow terminal = WARN. ASCII color = INFO.

#### Story 49.3: Task Data Integrity Checks

Task file existence and YAML validity. Per-task `Validate()`. Duplicate ID detection. Dependency reference validation. Blocker/CompletedAt consistency. Legacy migration detection (`tasks.txt`, `source_provider`).

**AC:** Task Data shows task count by status. Duplicate IDs = FAIL. Dangling deps = WARN. Inconsistent fields = WARN. Legacy files = WARN.

#### Story 49.4: Provider Health Checks

Check each configured provider for health. Integrate existing `HealthChecker` methods. Multi-provider iteration with timeout protection.

**AC:** Each provider checked and reported. No provider = FAIL. One failing = category WARN. Obsidian missing vault = FAIL.

#### Story 49.5: Session & Analytics Checks

Validate `sessions.jsonl` line-by-line. Check `patterns.json` integrity. Report session statistics.

**AC:** Corrupt JSONL lines reported with line numbers. Corrupt patterns.json = WARN. No sessions = INFO. Missing fields = WARN.

#### Story 49.6: Sync & Offline Queue Checks

Validate sync state staleness, WAL queue health, orphaned temp files.

**AC:** Stale sync (>24h) = WARN. Stuck ops (retries ≥ 10) = WARN. Excessive backlog = WARN. Orphaned .tmp files = WARN. No sync history = INFO.

#### Story 49.7: Enrichment Database Checks

SQLite integrity check, schema version verification, existence check.

**AC:** Missing DB = WARN. Failed integrity = FAIL. Schema mismatch = WARN. Healthy = OK with version.

#### Story 49.8: Auto-Repair (`--fix` flag)

Conservative auto-repair for safe, reversible issues only. Never auto-modify user data.

**AC:** Fix orphaned .tmp, corrupt patterns.json, missing config.yaml, stale version cache, incorrect permissions, legacy migration. Never fix corrupt tasks/sessions/DB.

#### Story 49.9: Channel-Aware Version Checking

GitHub Releases API with 24h cache. Channel-aware comparison. Opt-out controls. CI auto-disable.

**AC:** Cached check, channel-aware updates, cross-channel for alpha, dev build skip, env var opt-out, CI detection.

#### Story 49.10: Verbose Mode, Category Filter & Polish

`--verbose`, `--category`, `--skip-version` flags. Colored icons. Summary line. Exit codes for scripting.

**AC:** Verbose shows detail. Category filter works. Colors respect --no-color. Exit codes: 0/1/2.

### Design Decisions

- D-154: Command name `doctor`; supersedes existing `health` command
- D-155: Flutter-style icon output as default format
- D-156: Conservative auto-fix with `--fix` flag
- D-157: 24-hour cached version check (gh CLI pattern)
- D-158: Channel-aware version: show within channel + cross-channel if higher
- X-094: Rejected interactive repair wizard
- X-095: Rejected doctor as part of every command startup
- X-096: Rejected telemetry/crash reporting in doctor
- X-097: Rejected `health` and `doctor` coexisting as separate commands
- X-098: Rejected command name `check` or `diagnose`

### Research

- Doctor Research: `_bmad-output/planning-artifacts/threedoors-doctor-research.md`

---

## Epic 50: In-App Bug Reporting

**Epic Goal:** Add a `:bug` command for frictionless in-app bug reporting with navigation breadcrumb trail, automatic environment context, mandatory preview, and tiered submission methods.

**Prerequisites:** None (standalone feature)
**Status:** In Progress (0/3 done; 50.1 In Review)
**Priority:** P2

### Overview

When ThreeDoors users encounter a bug, reporting friction is high: leave the TUI, open browser, navigate to GitHub Issues, manually describe environment, reconstruct steps from memory. By step 3, most users give up. The `:bug` command makes bug reporting a natural part of the TUI conversation, aligned with SOUL.md's "friend helping you" philosophy.

The implementation follows strict privacy principles: a ring buffer captures navigation breadcrumbs (view transitions, non-text keys, command names) in memory only. Text input (`tea.KeyRunes`) is never captured — the privacy firewall is at the capture level, not the report level. Users see a mandatory preview of exactly what will be sent before any data leaves their machine.

### Stories

#### Story 50.1: Breadcrumb Tracking System

**Status:** In Review
**Priority:** P2
**Depends On:** None
**Effort:** Small

Ring buffer breadcrumb system tracking the last 50 user navigation actions in memory. Captures view transitions, non-text key events, and command names (arguments stripped). Text input (`tea.KeyRunes`) is never recorded — this is the privacy firewall. Integrated at top of `MainModel.Update()`.

**AC:** Ring buffer records view transitions with timestamps. Non-text keys recorded (`key:Enter`). Text input never captured. Commands recorded without args (`cmd:stats`). Buffer wraps at 50 entries. `Format()` returns chronological human-readable output.

#### Story 50.2: Bug Report View & Environment Collection

**Status:** Not Started
**Priority:** P2
**Depends On:** 50.1
**Effort:** Medium

New `ViewBugReport` mode with text description input, automatic environment data collection (version, OS, terminal, theme, task count — strict allowlist), breadcrumb trail integration, and mandatory preview screen. Wired via `:bug` command in `search_view.executeCommand()`.

**AC:** `:bug` opens bug report view. Environment shows only allowlisted data. Blocklist disclaimer displayed. Enter shows full markdown preview. Esc cancels and returns to previous view. No task content, file paths, or personal data in output.

#### Story 50.3: Submission Methods (Browser, API, File)

**Status:** Not Started
**Priority:** P2
**Depends On:** 50.2
**Effort:** Medium

Three tiered submission methods from the preview screen: (1) Browser URL — opens GitHub issue creation with pre-filled title/body via URL query params, zero auth needed; (2) GitHub API — direct submission if `GITHUB_TOKEN` configured; (3) Local file — saves to `~/.threedoors/bug-reports/` as GitHub-flavored markdown. Error cascading between methods.

**AC:** `[b]` opens browser with pre-filled issue URL. `[s]` (if token) creates issue via API, shows URL. `[f]` saves to `~/.threedoors/bug-reports/bug-<timestamp>.md`. Browser failure offers clipboard fallback. API failure offers browser/file alternatives. Success returns to previous view.

### Design Decisions

- D-112: Browser URL as primary bug report submission (zero-auth)
- D-113: Ring buffer breadcrumbs (50 entries, count-bounded)
- D-114: Allowlist-only privacy for bug reports (capture-level filtering)
- D-115: Mandatory preview before bug report submission
- D-116: Bug report target repo hardcoded to arcaven/ThreeDoors
- X-059: Rejected OAuth device flow for bug report auth
- X-060: Rejected gh CLI for bug report submission
- X-061: Rejected time-bounded breadcrumb buffer
- X-062: Rejected blocklist approach for bug report privacy
- X-063: Rejected configurable target repo for bug reports

### Research

- Research: `_bmad-output/planning-artifacts/in-app-bug-reporting-research.md`
- Party Mode: `_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md`

---

## Epic 51: SLAES — Self-Learning Agentic Engineering System (P1)

**Goal:** Build a continuous improvement meta-system with a persistent `retrospector` agent that monitors PR merges, detects process waste, audits doc consistency, analyzes CI/conflict patterns, and files improvement recommendations to BOARD.md.

**Prerequisites:** Epic 37 (Persistent BMAD Agents — complete)

**Status:** In Progress (5/11 stories done; 5 In Review)

**Phasing:**
- Phase 0 (Bootstrap): Stories 51.1-51.2 — Agent definition rewrites
- Phase 1 (MVP): Stories 51.3-51.6 — Core monitoring and recommendations
- Phase 2 (After 2-week validation): Stories 51.7-51.10 — Advanced analysis and PR creation

### Stories

#### Story 51.1: Retrospector Agent Definition (Responsibility+WHY Format)

Create `agents/retrospector.md` — the SLAES primary agent — using responsibility+WHY format from day one. Persistent agent with 15-minute polling, Level 2 authority (read + BOARD.md for MVP), dual-loop architecture, 5 Watchmen safeguards, and context exhaustion mitigation.

**AC:** Definition uses responsibility+WHY format (no procedural instructions). Incident-hardened guardrails reference INC-001/002/003. Includes authority table (CAN/CANNOT), operational mode rotation, self-restart trigger (20 PRs or 8 hours), consumer model for project-watchdog/arch-watchdog interaction.

#### Story 51.2: Rewrite Operational Agent Definitions (Responsibility+WHY Format)

Rewrite 5 operational agent definitions (merge-queue, pr-shepherd, worker, project-watchdog, supervisor) in responsibility+WHY format with incident-hardened guardrails. Phase 0 bootstrap task.

**AC:** All 5 definitions rewritten with WHY rationale and incident citations. merge-queue owns merge integrity. pr-shepherd has INC-001 isolation guardrail. worker has INC-002 no-manual-sync guardrail. project-watchdog has INC-003 mutex guardrail. supervisor has subagent abuse guardrail. All agents restarted and verified operational.

#### Story 51.3: JSONL Findings Log & Per-Merge Lightweight Retro

Implement the data collection foundation: every merged PR generates a JSONL entry with AC match checking, CI first-pass rate, conflict data, and rebase count. Rolling retention with archival.

**AC:** Per-PR JSONL entries appended on merge detection. Fields: pr, story, ac_match (full/partial/none/no-story), ci_first_pass, conflicts, rebase_count, timestamp. Rolling retention at 1000 entries with monthly archive. Idempotent processing with last_processed_pr marker.

#### Story 51.4: Saga Detection (Dispatch Waste Alerting)

Detect when 2+ workers are dispatched for the same fix within 4 hours. Alert supervisor with failure chain analysis and recommended action.

**AC:** Same-branch dispatch detection within 4-hour window. Supervisor alert with branch name, worker names, failure chain. Escalation trap pattern detection (fix-A-break-B chain). JSONL logging with type "saga_detected". Recurrence tracking with BOARD.md recommendation after 3+ occurrences.

#### Story 51.5: Doc Consistency Audit (Periodic Cross-Check)

Periodically cross-check epic-list, epics-and-stories, ROADMAP, and story files for drift. Runs as a rotating deep analysis mode every 4 hours.

**AC:** Detects status mismatches across all 4 planning doc layers. Detects orphaned stories and ghost entries. Messages supervisor only on inconsistencies (no noise on clean runs). JSONL logging with type "doc_inconsistency" or "doc_audit_clean". Runs in 4-hour rotation with other deep analysis modes.

#### Story 51.6: BOARD.md Recommendation Pipeline

Unified output layer: aggregate findings into BOARD.md recommendations with confidence scoring (High/Medium/Low), evidence trails, and kill switch safeguard.

**AC:** Recommendations filed in BOARD.md "Pending Recommendations" section with auto-assigned P-NNN ID. Confidence scoring: High (5+ data points), Medium (2-4), Low (1). Kill switch: 3 consecutive rejections → read-only mode. Rate-limited to 3 recommendations per batch cycle.

#### Story 51.7: Merge Conflict Rate Analysis (Phase 2)

Analyze merge conflict patterns: hot file detection, epic collision zones, parallel dispatch safety scoring, rebase churn analysis.

**AC:** Hot file detection (3+ concurrent PRs). Epic collision zone identification. Rebase churn flagging (3+ rebases). BOARD.md recommendations with specific file paths and sequencing suggestions. Depends on Phase 1 validation.

#### Story 51.8: CI Failure Rate Analysis & Coding Standard Proposals (Phase 2)

Classify CI failures by taxonomy (race condition, lint, flakiness, build, coverage) and trace to fixable spec-chain layer. Propose CLAUDE.md rule changes and coding standard updates.

**AC:** Failure taxonomy classification. Spec-chain layer tracing (Code → Story → PRD → Architecture → CLAUDE.md). Pattern detection across 3+ PRs triggers BOARD.md recommendation. Unclassified failures flagged for human review. Depends on Phase 1 validation.

#### Story 51.9: Research Lifecycle Tracking (Phase 2)

Track research artifacts through lifecycle: active → formalized → stale → abandoned. Alert on unformalized research older than 2 weeks.

**AC:** Scan `_bmad-output/planning-artifacts/`. Classify by lifecycle state. Flag stale research (>2 weeks, no stories). Report total/active/formalized/stale counts. Depends on Phase 1 validation.

#### Story 51.10: PR Creation Authority & Trend Reporting (Phase 2)

Expand retrospector authority to create PRs proposing improvements. Generate weekly trend reports with project health metrics.

**AC:** PR creation on `slaes/` branches for high-confidence recommendations pending >48 hours. Self-modification safeguard (never touch own definition). Weekly trend report with CI first-pass rate, rebase count, saga count, recommendation acceptance rate. Metric regression detection. Depends on all Phase 1 + Phase 2 stories.

#### Story 51.11: Retrospector Autonomy Fixes — Agent Definition Rewrite

Rewrite `agents/retrospector.md` to eliminate language that causes Claude to seek human confirmation. Five targeted changes: add imperative "Your Rhythm" polling loop with bash commands, reframe "Periodic Human Review" as passive/async, change "Kill Switch" to self-monitor BOARD.md, move communication rules higher in doc, add anti-prompting guardrail. Text-only changes — no code modifications.

**AC:** Imperative polling loop with explicit bash commands (matching arch-watchdog/envoy/project-watchdog pattern). "Periodic Human Review" reframed as passive — not agent's responsibility. "Kill Switch" detects rejections via BOARD.md state, not interactive feedback. Communication section positioned before Watchmen Safeguards. Anti-prompting guardrail in Incident-Hardened Guardrails section. All 5 Watchmen safeguards preserved (reframed, not removed). Retrospector operates autonomously after restart.

### Design Decisions

- D-1: Single agent (not 3 separate agents) — avoids pushing agent count to 8
- D-2: System name SLAES, agent name `retrospector`
- D-3: Persistent agent with 15-minute polling (not 5-minute, not cron)
- D-4: Level 2 authority (read + propose, not read-only or auto-apply)
- D-5: Consumer model — consumes project-watchdog/arch-watchdog outputs, never duplicates
- D-6: Dual-loop architecture (spec-chain quality + operational efficiency)
- D-7: Per-PR lightweight data collection + periodic batch analysis
- D-8: 5 Watchmen safeguards (no self-mod, audit trail, confidence scoring, human review, kill switch)
- D-9: Operational mode rotation for context management
- D-10: Prevention over detection — responsibility+WHY definition methodology
- X-A1: Rejected three focused agents (agent count → 8)
- X-A2: Rejected two agents (duplicate analytical pipeline)
- X-A3: Rejected `kaizen-agent` name (potentially pretentious)
- X-A9: Rejected ephemeral/cron-based (loses state, saga detection needs continuity)
- X-A10: Rejected 5-minute polling (unnecessary, budget impact)
- X-A12: Rejected read-only recommend-only (advisory doesn't work — registry precedent)
- X-A13: Rejected auto-apply (INC-001 precedent)

### Research

- SLAES Party Mode: `_bmad-output/planning-artifacts/agentic-engineering-agent-party-mode.md`
- Subagent Abuse Investigation: `_bmad-output/planning-artifacts/subagent-abuse-investigation.md`

---

## Epic 54: Gemini Research Supervisor — Deep Research Agent Infrastructure (Rearchitected) (P2)

**Goal:** Deploy a persistent research-supervisor agent that wraps the official Gemini CLI (`@google/gemini-cli`) with OAuth authentication, providing web-grounded research with context packaging, result shielding, and dual-tier budget management (Pro + Flash).

**Prerequisites:** Epic 37 (Persistent BMAD Agents — complete), Node.js/npm, Google Account

**Dependencies:** None (agent infrastructure, not application code)

**Status:** In Progress (2/5 stories done — PRs #537, #538)

**Rearchitecture Note:** This epic was originally designed around `24601/agent-deep-research` (Python, paid API key). Rearchitected per user request to use the official Gemini CLI with OAuth (free tier). See D-164 (supersedes D-154).

### Story 54.1: Research-Supervisor Agent Definition (Gemini CLI + OAuth)

**As** the supervisor agent,
**I want** a persistent research-supervisor agent that invokes the Gemini CLI via OAuth for web-grounded research, with a clear definition file, message-check loop, request protocol, and authority matrix,
**So that** any agent can request research via messaging and receive structured findings without blocking.

**Acceptance Criteria:**
- Agent definition at `agents/research-supervisor.md` follows Responsibility+WHY format (Story 51.2)
- 5-minute message-check loop documented: check messages → dispatch queued via `gemini -p` → process results
- Synchronous execution model: `gemini -p "<query>" --output-format json` (no async polling)
- Model selection: `gemini-2.5-pro` for deep/standard, `gemini-2.5-flash` for quick
- Request protocol: `RESEARCH priority=<high|normal|low> depth=<quick|standard|deep> [context=<bundles>]: <question>`
- Authority matrix: CAN (receive, queue, dispatch, store, summarize, deliver) / CANNOT (decide, execute, create stories, modify code) / ESCALATE (budget exhausted, OAuth failure, repeated failures)
- Communication protocols: RESEARCH-RESULT, RESEARCH-ERROR, RESEARCH-BUDGET message formats

**Priority:** P2 | **Depends On:** None

### Story 54.2: Gemini CLI Installation, OAuth Setup & Wrapper Script

**As** the research-supervisor agent,
**I want** the Gemini CLI installed via npm, authenticated via OAuth, and wrapped in a research-oriented shell script,
**So that** I can dispatch research queries with `scripts/gemini-research.sh` and receive structured JSON responses.

**Acceptance Criteria:**
- Gemini CLI installable via `npm install -g @google/gemini-cli` (documented)
- OAuth flow: `gemini` → browser sign-in → token cached and auto-refreshed (documented)
- Verification: `gemini -p "Hello" --output-format json` returns valid JSON
- Wrapper script at `scripts/gemini-research.sh` with `--depth` (model selection) and `--query` parameters
- `GEMINI.md` research system prompt in project root (automatic context for all queries)
- `_bmad-output/research-reports/` directory created with `.gitkeep`
- Report directories gitignored; `budget.json` tracked
- No `_tools/` directory, no Python, no API key

**Priority:** P2 | **Depends On:** None

### Story 54.3: Context Packaging & Prompt Engineering (Gemini CLI)

**As** the research-supervisor agent,
**I want** pre-defined context bundles and a prompt template that tailor each query via `--include-directories` and stdin piping,
**So that** Gemini receives focused, project-specific grounding that improves result relevance within token budgets.

**Acceptance Criteria:**
- 8 context bundles documented: core, architecture, prd, stories, decisions, code-sample, tui, tasks
- Keyword-to-bundle auto-detection mapping
- 60KB context budget with priority shedding order (drop code-sample → truncate stories → drop prd)
- Standard prompt template with: project context, grounding instructions (use GoogleSearch), question, output requirements
- `--include-directories` for directory-aligned bundles; stdin piping for assembled bundles
- `GEMINI.md` and `GEMINI_SYSTEM_MD` interaction documented

**Priority:** P2 | **Depends On:** 54.1

### Story 54.4: Result Shielding & Artifact Storage (Gemini CLI JSON)

**As** the supervisor (or any requesting agent),
**I want** research results delivered as concise executive summaries with full reports on disk,
**So that** context windows are not overwhelmed by lengthy research output.

**Acceptance Criteria:**
- Three-layer summary architecture: executive summary (≤500 words) → detailed report → raw JSON
- Per-query directory: `YYYYMMDD-HHMMSS-<slug>/` with report.md, executive-summary.md, response.json, request.json, context-bundle.md
- JSON parsing: `.response` field extracted from `gemini -p --output-format json` output
- Only executive summary sent via messaging; full report path included
- RESEARCH-RESULT message format documented
- Gating flow documented: research → summary → human/supervisor decision → optional action
- No autonomous action restriction in CANNOT section

**Priority:** P2 | **Depends On:** 54.1, 54.2

### Story 54.5: Rate Limiting, Budget Management & Model Selection

**As** the supervisor,
**I want** dual-tier daily query count tracking (Pro + Flash), priority queuing, model selection, and deduplication,
**So that** the free-tier limits (50 Pro/day, 1,000 Flash/day) are used efficiently.

**Acceptance Criteria:**
- `budget.json` schema: date, pro_limit/used/remaining, flash_limit/used/remaining, query history
- Daily reset at midnight UTC
- Reserve pool: 5 Pro queries held back after 6pm UTC for high-priority
- Priority queue: high=immediate, normal=FIFO, low=budget-permitting
- Model selection: quick→Flash, standard/deep→Pro, with Pro→Flash fallback on exhaustion
- Batch optimization: 3+ related queries combined with merged prompt
- Deduplication: 7-day lookback against existing `request.json` files
- 2-minute cooldown for Pro dispatches (60 RPM for Flash)
- 80% Pro budget warning threshold
- RESEARCH-BUDGET escalation message for exhausted budget
- 429 rate-limit fallback: Pro → Flash downgrade with logging

**Priority:** P2 | **Depends On:** 54.1, 54.2

### Dependency Graph

```
54.1 (Agent Definition)      ─┬──▶ 54.3 (Context Packaging)
54.2 (CLI + OAuth + Wrapper)  ─┤
                                ├──▶ 54.4 (Result Shielding) ← depends on 54.1 + 54.2
                                └──▶ 54.5 (Rate Limiting)    ← depends on 54.1 + 54.2
```

Stories 54.1 and 54.2 can parallelize. Stories 54.3, 54.4, and 54.5 can parallelize after 54.1+54.2 complete.

### Decisions

- D-164: Gemini CLI + OAuth as research execution layer (supersedes D-154: agent-deep-research)
- Rejected: keep agent-deep-research (paid API key), deep-research extension (paid key), direct API (manual polling), Python SDK (retains Python dep)

### Research

- Rearchitecture Research: `_bmad-output/planning-artifacts/gemini-cli-oauth-research.md`
- Original Design (superseded): `_bmad-output/planning-artifacts/gemini-research-supervisor-design.md`

---

## Epic 55: CI Optimization Phase 1

**Goal:** Reduce PR CI wall clock time from 3m33s to ~2m08s through CI configuration changes only — no test code modifications. Fix Docker E2E redundancy, add benchmark path filtering, improve local dev speed.

**Prerequisites:** None
**Status:** Complete (3/3 done)

### Story 55.1: Docker E2E Push-Only and Lint Version Fix

**As** a contributor,
**I want** the Docker E2E job to run only on push-to-main (not on PRs),
**So that** PR CI cost is reduced without losing defense-in-depth on the main branch.

**Acceptance Criteria:**
- Docker E2E job (`test-docker-e2e`) does NOT execute on pull requests
- Docker E2E job continues to execute on push-to-main
- `Dockerfile.test` `GOLANGCI_LINT_VERSION` updated from `v2.1.6` to `v2.10.1` (matching CI)
- Required checks (Quality Gate, Benchmarks) unaffected

**Technical Notes:**
- Change `test-docker-e2e` `if` to `github.event_name == 'push'`
- Remove `needs: changes` (no path filtering needed when always running on push)
- Fix lint version in `Dockerfile.test` line 9

**Priority:** P1 | **Depends On:** None

### Story 55.2: Benchmark Path Filtering

**As** a contributor,
**I want** benchmarks to only run on PRs that touch performance-relevant code,
**So that** non-performance PRs complete faster (3m33s → 2m08s).

**Acceptance Criteria:**
- `changes` job outputs a `perf` filter matching `internal/core/**`, `internal/adapters/textfile/**`, `go.mod`
- Benchmarks skip on PRs where `perf` is false
- Benchmarks always run on push-to-main (safety net)
- PR critical path becomes Quality Gate (~2m08s) for non-perf PRs

**Technical Notes:**
- Add `perf` output to `changes` job with dorny/paths-filter
- Update `benchmarks` `if` to: `github.event_name == 'push' || (code == 'true' && perf == 'true')`

**Priority:** P1 | **Depends On:** None

### Story 55.3: Local Dev Acceleration (make test-fast + CI Cache)

**As** a developer,
**I want** a `make test-fast` target that runs tests in short mode,
**So that** I can get rapid local feedback (~10s) without running the full test suite (~33s).

**Acceptance Criteria:**
- `make test-fast` runs `go test -short ./...`
- `test-fast` listed in `.PHONY` declaration
- All `setup-go@v6` steps in CI have `cache-dependency-path: go.sum`
- Existing `make test` behavior unchanged

**Technical Notes:**
- Add `test-fast` target to Makefile
- Add `cache-dependency-path: go.sum` to 3 `setup-go` steps in CI

**Priority:** P1 | **Depends On:** None

### Dependency Graph

```
55.1 (Docker E2E Push-Only)    ──▶ Independent
55.2 (Benchmark Path Filter)   ──▶ Independent
55.3 (Local Dev Acceleration)  ──▶ Independent
```

All three stories are fully independent and can be implemented in parallel.

### Decisions

- D-166: CI Optimization Phase 1 scope (Docker E2E push-only, benchmark path filtering, make test-fast)
- Rejected: remove Docker E2E entirely (defense-in-depth value), incremental linting (can miss cross-file issues), benchmarks only on push (delays detection), split Quality Gate (marginal gain), `-short` in CI (risks missing bugs)

### Research

- Full research: `_bmad-output/planning-artifacts/ci-test-optimization/` (5 party mode sessions)
- Synthesis: `_bmad-output/planning-artifacts/ci-test-optimization/05-synthesis-optimization-roadmap.md`

## Epic 56: Door Visual Redesign — Three-Layer Depth System (P1)

**Goal:** Transform door rendering from imperceptible wireframe shadows into solid, 3D-feeling surfaces using a three-layer approach: background fill for visual mass, bevel lighting for raised-surface perception, and gradient shadow for spatial depth.

**Prerequisites:** Epic 48 (Door-Like Doors — complete), Epic 17 (Door Theme System — complete)

**Status:** Not Started (0/5 stories)

### Story 56.1: ThemeColors Extension + Background Fill
- **As a** user, **I want** door interiors to have a solid background color instead of being transparent wireframes, **so that** doors look like physical surfaces with visual mass rather than flat line drawings.
- **Scope:** Extend `ThemeColors` struct with 5 new depth color fields (`FillLower`, `Highlight`, `ShadowEdge`, `ShadowNear`, `ShadowFar`). Add background fill to all 8 themes' `Render()` functions using `lipgloss.Style.Background(Fill)` for interior rows. ~80 LOC, zero width cost.
- **Depends on:** None (foundational story)
- **AC:** ThemeColors has 5 new fields; all 8 themes render interior bg; zero width cost; golden files updated
- **Story file:** `docs/stories/56.1.story.md`

### Story 56.2: Bevel Lighting
- **As a** user, **I want** door borders to have a "raised surface" bevel effect with lighter top/left edges and darker bottom/right edges, **so that** doors appear as 3D raised surfaces rather than flat outlines.
- **Scope:** Top/left borders use `Highlight` color, bottom/right use `ShadowEdge`. Classic "raised button" GUI effect. All 8 themes. ~120 LOC, zero width cost.
- **Depends on:** None (can parallelize with 56.1)
- **AC:** Bevel colors on all 8 themes; Selected overrides bevel; zero width cost; golden files updated
- **Story file:** `docs/stories/56.2.story.md`

### Story 56.3: Shadow Overhaul
- **As a** user, **I want** doors to cast a visible gradient shadow instead of the current imperceptible 1-character shadow, **so that** doors appear to "lift off" the terminal background with proper spatial depth.
- **Scope:** Refactor `ApplyShadow()` from post-processor into per-theme `Render()`. Width-adaptive 0-3 column gradient shadow. Integrate with crack-of-light. ~150 LOC, +1-2 chars width (adaptive).
- **Depends on:** Story 56.1 (needs ShadowNear/ShadowFar fields)
- **AC:** Gradient shadow at 2+ cols; width-adaptive; crack-of-light integration; golden files updated
- **Story file:** `docs/stories/56.3.story.md`

### Story 56.4: Panel Zone Shading
- **As a** user, **I want** the upper and lower door panels to have differentiated background colors, **so that** the two-panel door structure is visually distinct and adds depth.
- **Scope:** Upper panel uses `Fill`, lower panel uses `FillLower` (darker). Switchover at `DoorAnatomy.PanelDividerRow`. ~60 LOC, zero width cost.
- **Depends on:** Story 56.1 (needs FillLower field and bg fill infrastructure)
- **AC:** Upper/lower panels visually differentiated; divider row belongs to lower; golden files updated
- **Story file:** `docs/stories/56.4.story.md`

### Story 56.5: Width-Adaptive Shadow Tuning
- **As a** user, **I want** the shadow gradient to be fine-tuned across terminal widths and all 8 themes, **so that** shadows look correct at every width.
- **Scope:** Tuning and golden file pass — validate shadow contrast ratios, add width-tier tests (narrow/medium/wide), adjust colors if needed. ~40 LOC.
- **Depends on:** Story 56.3 (shadow overhaul must be complete)
- **AC:** Golden files at 3 width tiers per theme; ShadowNear ≥ 4:1 contrast; selected gets wider shadow
- **Story file:** `docs/stories/56.5.story.md`

### Dependency Graph
```
Story 56.1 (ThemeColors + Bg Fill) ──┬──→ Story 56.3 (Shadow Overhaul) ──→ Story 56.5 (Tuning)
                                     └──→ Story 56.4 (Panel Zones)
Story 56.2 (Bevel) ─────────────────────→ (independent)
```

Stories 56.1 & 56.2 can parallelize. Stories 56.3 & 56.4 can parallelize after 56.1.

### Decisions

- D-173: Three-layer depth system (background fill + bevel lighting + shadow gradient)
- Rejected: X-109 Full Corridor (width cost too high), X-110 Adaptive Depth/terminal detection (over-engineering), X-111 Interior Texture (hurts readability), X-112 Braille Patterns (accessibility concerns)

### Research

- Full research: `_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md` (5-round party mode, 6 agents)

---

## Epic 57: LLM CLI Services (P1)

**Goal:** Enable ThreeDoors to invoke LLM CLI tools (Claude CLI, Gemini CLI, Ollama CLI) as subprocess-based service providers for intelligent task operations: extraction from natural language, enrichment, and breakdown. ThreeDoors as CLIENT calling LLMs (Direction 1), complementing Epic 24's MCP server (Direction 2, where ThreeDoors is SERVER).

**Prerequisites:** Epic 14 (LLM Decomposition — complete), Epic 23 (CLI Interface — complete)

**Status:** Not Started (0/8 stories done)

**Two Directions of Integration:**
- Direction 1 (THIS EPIC): ThreeDoors → LLM CLIs. ThreeDoors invokes claude/gemini/ollama CLIs via `os/exec` for task extraction, enrichment, breakdown. ThreeDoors is the CLIENT.
- Direction 2 (Epic 24 — complete): LLM Agents → ThreeDoors MCP Server. Claude Desktop, Cursor, etc. connect via MCP. ThreeDoors is the SERVER.
- These compose: MCP tools become thin wrappers around the LLM Service Layer.

### Story 57.1: CLIProvider + CLISpec + CommandRunner Abstraction ⬜

**As** a developer,
**I want** a generic CLI-based LLM backend that implements the existing `LLMBackend` interface via subprocess execution,
**So that** ThreeDoors can invoke any LLM CLI tool through a uniform abstraction.

**Acceptance Criteria:**
- `CLISpec` struct defined with: Name, Command, BaseArgs, SystemPrompt (ArgTemplate), OutputFormat (ArgTemplate), InputMethod (stdin/arg/file), Timeout, ResponseParser
- `CLIProvider` implements `LLMBackend` interface (`Name()`, `Complete()`, `Available()`)
- Pre-built `CLISpec` factories: `ClaudeCLISpec()`, `GeminiCLISpec()`, `OllamaCLISpec(model)`, `CustomCLISpec(cmd, args)`
- `CommandRunner` with stdin support (`RunWithStdin`)
- Timeout enforcement via context cancellation
- Error handling: non-zero exit includes stderr, empty response returns `ErrEmptyResponse`

**Priority:** P0 | **Depends On:** None

### Story 57.2: Auto-Discovery and Fallback Chain ⬜

**As** a user,
**I want** ThreeDoors to automatically detect which LLM CLI tools are installed and use the best available one,
**So that** LLM features work with zero configuration.

**Acceptance Criteria:**
- `DiscoverBackend(cfg)` checks CLI tools in priority order (claude → gemini → ollama → HTTP backends)
- User-configured backend takes priority over auto-discovery
- Graceful degradation when no backends available (`ErrBackendUnavailable`)
- Discovery result logged at INFO level for debugging

**Priority:** P0 | **Depends On:** 57.1

### Story 57.3: TaskExtractor Service + Extraction Prompt ⬜

**As** a user,
**I want** to extract actionable tasks from unstructured text (meeting notes, Obsidian pages, transcripts, clipboard),
**So that** tasks hiding in prose are surfaced into my ThreeDoors pool.

**Acceptance Criteria:**
- `TaskExtractor` service with `ExtractFromText(ctx, text) ([]ExtractedTask, error)`
- `ExtractedTask` struct: Text, Effort (1-5), Tags, Source, Confidence
- Source helpers: `ExtractFromFile`, `ExtractFromClipboard` (pbpaste)
- 32KB input size limit with clear error message
- JSON-only LLM output format with retry on malformed response
- Deduplication check against existing task pool

**Priority:** P0 | **Depends On:** 57.1

### Story 57.4: Extraction TUI — `:extract` Command + Review Screen ⬜

**As** a TUI user,
**I want** to type `:extract` to extract tasks from text, files, or clipboard and review them before importing,
**So that** I can convert unstructured text into actionable tasks without leaving the TUI.

**Acceptance Criteria:**
- `:extract` command registered in command palette
- Source picker: [f]ile, [c]lipboard, [p]aste
- Loading spinner with Esc cancel
- Review screen: task list with Space toggle, A/N select all/none, E inline edit, Enter import
- Flash message after import: "Imported N tasks from [source]"
- Latency target: <5s for short text

**Priority:** P0 | **Depends On:** 57.3

### Story 57.5: Extraction CLI — `threedoors extract` ⬜

**As** a CLI user,
**I want** to run `threedoors extract --file notes.txt` or pipe text via stdin,
**So that** I can batch-import tasks from scripts and non-interactive contexts.

**Acceptance Criteria:**
- `--file`, `--clipboard`, stdin pipe support
- Human-readable output with confirmation prompt
- `--json` flag for scripted output
- `--yes` flag for auto-import without confirmation

**Priority:** P0 | **Depends On:** 57.3

### Story 57.6: TaskEnricher Service + Enrichment TUI ⬜

**As** a user viewing a vague task,
**I want** to press E (or type `:enrich`) to have an LLM add context, tags, and effort,
**So that** sparse tasks become actionable.

**Acceptance Criteria:**
- `TaskEnricher` service with `Enrich(ctx, task) (*EnrichedTask, error)`
- Before/after diff display in TUI detail view
- Accept/edit/discard actions
- Latency target: <3s

**Priority:** P1 | **Depends On:** 57.1

### Story 57.7: TaskBreakdown Service — Extend Epic 14 Decomposer ⬜

**As** a user facing an overwhelming task,
**I want** to press B (or type `:breakdown`) to decompose it into subtasks,
**So that** large tasks become approachable.

**Acceptance Criteria:**
- Existing `LLMTaskDecomposer` works with CLI backends (interface compatibility)
- `:breakdown` command in detail view with loading spinner
- Subtask review screen with toggle/import
- Latency target: <8s

**Priority:** P1 | **Depends On:** 57.1

### Story 57.8: `threedoors llm status` Command ⬜

**As** a user,
**I want** to run `threedoors llm status` to see which LLM backend is active,
**So that** I can debug LLM feature issues.

**Acceptance Criteria:**
- CLI: shows active backend, command path, availability, fallbacks, service readiness
- TUI: `:llm-status` command shows same info
- `--json` flag for scripted output
- Helpful message when no backends available

**Priority:** P1 | **Depends On:** 57.1, 57.2

## Epic 58: Supervisor Shift Handover — Context-Aware Supervisor Rotation (P2)

**Goal:** Detect supervisor context window degradation via daemon monitoring, serialize operational state, and transfer control to a fresh supervisor instance — all while workers continue uninterrupted.

**Prerequisites:** None (multiclaude daemon infrastructure already exists)

**Status:** Not Started (0/7 stories)

**Phasing:**
- Phase 1 (MVP): Stories 58.1-58.4 — Shift clock, rolling snapshot, handover orchestrator, supervisor startup
- Phase 2 (Hardening): Stories 58.5-58.7 — Emergency protocol, audit trail, manual trigger

### Story 58.1: Shift Clock — Transcript Monitoring in Daemon Refresh Loop

**As a** multiclaude daemon, **I want to** monitor the supervisor's JSONL transcript for context window utilization signals, **so that** I can detect when the supervisor is approaching degradation and trigger a handover.

**Acceptance Criteria:**
- Locates supervisor's active JSONL transcript file
- Collects three metrics: file size (bytes), compression event count, assistant message count
- Classifies session as Green/Yellow/Red zone using three-tier thresholds
- Enforces 30-minute time floor (no premature handover)
- Anti-oscillation: 30-minute minimum between handovers
- Natural seam detection: waits for task boundary (no active dispatches, no pending acks)
- Writes signal file when handover conditions met

**Technical Notes:**
- Shell script in daemon's existing 5-minute refresh loop
- No Go code changes needed for v1
- Thresholds: Yellow at compression count >= 3 or JSONL > 5MB; Red at >= 6 or > 10MB

**Priority:** P2 | **Depends On:** None

### Story 58.2: Rolling State Snapshot Generator

**As a** multiclaude daemon, **I want to** maintain a rolling snapshot of system state from external sources, **so that** incoming supervisors have immediate operational context without relying on a potentially degraded outgoing supervisor.

**Acceptance Criteria:**
- Produces YAML file at `~/.multiclaude/handover/<repo>/shift-state.yaml`
- Contains worker state (name, task, branch, PR, dispatch time)
- Contains persistent agent status (from tmux session list)
- Contains open PR list (number, title, CI status)
- Schema versioned (`version: 1`) with UTC ISO-8601 timestamps
- Warns if file exceeds 10KB
- Atomic writes (write to .tmp, rename)
- Preserves existing supervisor delta sections when updating observable state

**Technical Notes:**
- Shell script in daemon's existing refresh loop
- Data sources: `multiclaude worker list`, `gh pr list`, `tmux list-windows`

**Priority:** P2 | **Depends On:** None

### Story 58.3: Handover Orchestrator — Daemon Coordination Logic

**As a** multiclaude daemon, **I want to** orchestrate the full handover sequence from outgoing to incoming supervisor, **so that** authority transfers cleanly without message loss, split-brain, or worker disruption.

**Acceptance Criteria:**
- Detects handover signal file and removes it to prevent re-triggering
- Sends HANDOVER_REQUESTED to outgoing supervisor
- Waits up to 120s for HANDOVER_COMPLETE (stubs emergency path for 58.5)
- Spawns incoming supervisor with state file reference
- Waits up to 180s for incoming READY signal
- Kills outgoing supervisor after incoming confirms ready
- At no point: zero supervisors watching, or two supervisors dispatching
- Records handover metadata (timing, type, anomalies)

**Technical Notes:**
- Shell script extending daemon's existing bash logic
- Authority transfer implicit via tmux window replacement
- File-based messaging survives supervisor lifecycle transitions

**Priority:** P2 | **Depends On:** 58.1, 58.2

### Story 58.4: Supervisor Startup with State File

**As an** incoming supervisor agent, **I want to** read a shift-state.yaml file on startup and assume operational control, **so that** I can seamlessly continue managing workers and agents without losing context.

**Acceptance Criteria:**
- Detects SHIFT_HANDOVER task and reads shift-state.yaml
- Knows each active worker's name, task, branch, and last status
- Pings each active worker for status verification
- Adopts priorities from state file
- Tracks unresolved pending decisions
- Processes all unacknowledged messages before accepting new work
- Signals READY to daemon
- Operates correctly with daemon-only snapshot (no supervisor delta) in emergency mode

**Technical Notes:**
- Modifies supervisor agent definition (`agents/supervisor.md`)
- Conditional startup branch: SHIFT_HANDOVER vs normal startup
- Worker pings are non-blocking

**Priority:** P2 | **Depends On:** 58.2, 58.3

### Story 58.5: Emergency Handover Protocol

**As a** multiclaude daemon, **I want to** handle the case where the outgoing supervisor is unresponsive during handover, **so that** the system never gets permanently stuck due to a broken supervisor.

**Acceptance Criteria:**
- Emergency protocol activates after 120s timeout
- Force-kills outgoing supervisor (tmux session and Claude process)
- Spawns incoming with emergency flag
- Incoming does full worker audit (pings ALL workers, checks ALL PRs, reconciles messages)
- Reports discrepancies to user
- Retries spawn once after 30s on failure; alerts user on second failure

**Technical Notes:**
- Extends orchestrator from Story 58.3 (replaces timeout stub)
- Updates supervisor definition for emergency mode detection

**Priority:** P2 | **Depends On:** 58.3, 58.4

### Story 58.6: Handover History & Audit Trail

**As a** multiclaude operator, **I want** handover events archived with timing metrics and anomaly tracking, **so that** I can debug handover issues, tune thresholds, and monitor system health.

**Acceptance Criteria:**
- State file archived to `~/.multiclaude/handover/<repo>/history/` with ISO timestamp suffix
- JSONL event log appended after each handover (timestamp, names, type, metrics, duration, anomalies)
- History directory capped at 50 files (rolling window)
- Summary command displays recent handover events
- Alerts on high handover frequency (>3/hour)

**Technical Notes:**
- Archive directory and JSONL log managed by daemon
- Summary could be `multiclaude supervisor history` subcommand

**Priority:** P2 | **Depends On:** 58.3

### Story 58.7: Manual Handover Trigger & User Notification

**As a** multiclaude user, **I want to** manually trigger a supervisor handover and be notified when automatic handovers occur, **so that** I can intervene when I notice degradation the shift clock missed, and stay informed about system state changes.

**Acceptance Criteria:**
- `multiclaude supervisor handover` command triggers handover via signal file
- Anti-oscillation warning with manual override (`--force` flag)
- Notification on handover start (automatic or manual) with trigger metrics
- Notification on handover completion with duration and mode
- `multiclaude status` includes shift clock data (zone, JSONL size, compression count, session duration)

**Technical Notes:**
- New subcommand: `multiclaude supervisor handover`
- Extends `multiclaude status` output

**Priority:** P2 | **Depends On:** 58.1, 58.3

### Dependency Graph

```
57.1 (CLIProvider + CLISpec)
 ├──▶ 57.2 (Auto-Discovery)
 │     └──▶ 57.8 (llm status)
 ├──▶ 57.3 (TaskExtractor)
 │     ├──▶ 57.4 (Extraction TUI)
 │     └──▶ 57.5 (Extraction CLI)
 ├──▶ 57.6 (TaskEnricher)
 └──▶ 57.7 (TaskBreakdown)
```

57.1 is the foundation. After 57.1, stories 57.2-57.7 can parallelize (except 57.4/57.5 depend on 57.3). 57.8 depends on 57.1 and 57.2.

### Decisions

- S2-D1: Extend `LLMBackend` with CLI implementations (Option A). Rejected: new CLIService interface (Option B), capability negotiation (Option C)
- S2-D2: Two-layer architecture — Services (what) + Backends (how)
- S2-D4: Auto-discovery via `exec.LookPath` with fallback chain
- S5-D1: Declarative CLISpec struct. Rejected: per-provider classes with duplicated exec logic
- S1-D7: Privacy-tiered model (local default, cloud opt-in per SOUL.md)
- S1-D8: All LLM services user-initiated (no automatic/ambient processing)
- S3-D1: All sources reduce to `extractFromText(text)` — simpler architecture
- S3-D2: User review required before import — LLMs hallucinate
- S3-D5: 32KB input size limit for MVP — YAGNI on chunking
- S4-D1: Explicit commands for MVP; contextual suggestions P1; ambient P2
- S5-D6: Streaming, conversation, tool use deferred — P0 is request-response only

### Research

- Full research: `_bmad-output/planning-artifacts/llm-services-architecture/` (5 party mode sessions)
- Synthesis: `_bmad-output/planning-artifacts/llm-services-architecture/synthesis.md`

### Dependency Graph

```
58.1 (Shift Clock)        ──▶ Independent (daemon monitoring)
58.2 (Rolling Snapshot)   ──▶ Independent (daemon state collection)
58.3 (Orchestrator)       ──▶ Depends on 58.1, 58.2
58.4 (Supervisor Startup) ──▶ Depends on 58.2, 58.3
58.5 (Emergency Protocol) ──▶ Depends on 58.3, 58.4
58.6 (History & Audit)    ──▶ Depends on 58.3
58.7 (Manual Trigger)     ──▶ Depends on 58.1, 58.3

Phase 1 (MVP):        58.1 + 58.2 can parallelize → 58.3 → 58.4
Phase 2 (Hardening):  58.5, 58.6, 58.7 can parallelize after Phase 1
```

### Decisions

- D-168: External daemon monitoring (not supervisor self-reporting)
- D-169: Cold start (no hot/warm standby)
- D-170: Daemon-maintained rolling snapshot
- D-171: Hybrid shift clock (time floor + usage ceiling)
- D-172: Role-based agent addressing

### Research

- Full research: `_bmad-output/planning-artifacts/supervisor-shift-handover/` (5 party mode sessions)
- Synthesis: `_bmad-output/planning-artifacts/supervisor-shift-handover/synthesis-supervisor-shift-handover.md`

## Epic 59: Full-Terminal Vertical Layout (P1)

**Goal:** Transform ThreeDoors from a content-driven partial-terminal app into a full-terminal experience using AltScreen, a fixed-header/flex-content/fixed-footer layout engine, capped door height with perceptual centering, and graceful degradation across terminal sizes.

**Prerequisites:** None (foundational layout work)

**Status:** Not Started (0/2 stories)

**Phasing:**
- Story A (MVP): AltScreen, layout engine, door height cap, vertical centering, help dynamic height, SetHeight propagation
- Story B (Follow-up): Header/footer extraction from DoorsView, graceful degradation breakpoints, all secondary views fill terminal height

### Story 59.1: AltScreen + Layout Engine + Door Height Cap

**As a** user, **I want** ThreeDoors to fill my entire terminal with a clean layout, **so that** the app feels like a focused, deliberate experience that owns its space.

**Acceptance Criteria:**
- Program uses `tea.WithAltScreen()` — terminal content preserved/restored on exit
- `layoutFull()` function pads output to exactly terminal height
- Door height: `min(max(10, available * 0.5), 25)` — capped at 25 lines
- 40/60 top/bottom padding split for perceptual centering
- Help view uses terminal height instead of hardcoded 20 lines
- `SetHeight()` propagated to all views that currently lack it
- All tests pass including `go test -race ./internal/tui/...`

**Technical Notes:**
- One-line change for AltScreen at `cmd/threedoors/main.go:173`
- Layout engine in `MainModel.View()` — reference: `keybinding_overlay.go` already does full-height
- Door height change in `doors_view.go:268-273`
- Help view in `help_view.go` — replace `helpPageSize = 20` with dynamic
- Height propagation in `main_model.go:253-308`

**Priority:** P1 | **Depends On:** None

### Story 59.2: Header/Footer Extraction + Graceful Degradation + Secondary Views Fill Height

**As a** user, **I want** all views to fill my terminal height and the app to degrade gracefully on small terminals, **so that** every screen feels consistent and works well at any size.

**Acceptance Criteria:**
- `DoorsView.RenderDoors()` returns only doors; `RenderStatusSection()` returns status indicators
- Header (greeting, time context) and footer (help text, footer message) rendered by MainModel
- Breakpoint degradation: <10 minimal, 10-15 compact, 16-24 standard, 25-40 comfortable, 40+ spacious
- All secondary views (stats, search, sync log, themes, etc.) fill available terminal height
- Degradation is invisible — no error messages or warnings
- All tests pass including `go test -race ./internal/tui/...`

**Technical Notes:**
- DoorsView refactor: split View() into RenderDoors() and RenderStatusSection()
- Breakpoint function: `layoutBreakpoint(height int)` returns tier
- Secondary views use their height parameter for content sizing

**Priority:** P1 | **Depends On:** 59.1

### Dependency Graph

```
59.1 (AltScreen + Layout Engine + Door Cap)
 └──▶ 59.2 (Header/Footer Extraction + Degradation + All Views)
```

59.1 is foundation. 59.2 depends on 59.1. Sequential implementation required.

### Decisions

- D-114: Use `tea.WithAltScreen()` for full-terminal ownership
- D-115: Fixed header + Flex middle + Fixed footer layout model
- D-116: Door height capped at 25 lines: `min(max(10, available * 0.5), 25)`
- D-117: 40/60 top/bottom padding split for vertical centering
- D-118: Help/stats/search views use full available terminal height
- D-119: Breakpoint-based graceful degradation for small terminals
- D-120: Two-story implementation (MVP + refactor)
- D-121: Layout engine is prerequisite for Story 39.2 (keybinding bar)
- X-059: Stay in normal scrollback buffer (rejected)
- X-060: Unlimited door height growth (rejected)
- X-065: Single large story (rejected — too much scope risk)

### Research

- Full research: `_bmad-output/planning-artifacts/full-terminal-layout-research.md`
- Party mode: `_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md`
