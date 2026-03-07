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

**Implementation Status:** Epics 0-15, 3.5, 17-22 are COMPLETE. Epics 16, 23, 24, and 25 are NOT STARTED. 164 merged PRs total. Last audit: 2026-03-07.

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
| (cross-cutting) | Epic 0 ✅ | Infrastructure & Process Backfill (COMPLETE) |
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
| FR81-FR88 | Epic 24 | MCP/LLM Integration Server (NOT STARTED) |
| FR89-FR92 | Epic 25 | Todoist Integration (NOT STARTED) |
| FR93-FR96 | Epic 26 | GitHub Issues Integration (NOT STARTED) |

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

### Epic 24: MCP/LLM Integration Server — NOT STARTED
Expose ThreeDoors task management to LLMs via Model Context Protocol. Read-only queries, controlled enrichment proposals, analytics mining, and relationship graphs.
**FRs covered:** FR81-FR88
**Prerequisites:** Epic 13 ✅ (Multi-Source Aggregation), Epic 6 ✅ (Enrichment DB)
**Status:** Not Started. 8 stories planned (24.1-24.8). Research at `docs/research/llm-integration-mcp.md`.

### Epic 25: Todoist Integration — NOT STARTED
Todoist as a task source with thin HTTP client against REST API v1, read-only adapter, bidirectional sync.
**FRs covered:** FR89-FR92
**Prerequisites:** Epic 7 ✅ (Adapter SDK), Epic 13 ✅ (Multi-Source Aggregation), Epic 21 ✅ (Sync Protocol Hardening)
**Status:** Not Started. 4 stories planned (25.1-25.4). Research at `docs/research/task-source-expansion-research.md`.

### Epic 26: GitHub Issues Integration — NOT STARTED
GitHub Issues as a task source for developer workflows using the official go-github SDK. Label-based priority/status conventions.
**FRs covered:** FR93-FR96
**Prerequisites:** Epic 7 ✅ (Adapter SDK), Epic 13 ✅ (Multi-Source Aggregation), Epic 21 ✅ (Sync Protocol Hardening)
**Status:** Not Started. 4 stories planned (26.1-26.4). Research at `docs/research/task-source-expansion-research.md`.

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
**Then** `docs/spikes/config-schema.md` documents: proposed config.yaml schema, provider section design, migration path from current config
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
**Then** `docs/spikes/llm-decomposition.md` covers prompt engineering, output schema, git automation
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
**Then** `docs/research/choice-architecture.md` covers choice overload, paradox of choice, decision fatigue
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
**Then** `docs/research/mood-correlation.md` and `docs/research/procrastination.md` produced
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
**Origin:** Self-driving development pipeline research (2026-03-04). Research document at `docs/research/self-driving-development-pipeline.md`.
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

## Epic 24: MCP/LLM Integration Server

**Epic Goal:** Expose ThreeDoors task management services to LLMs through the Model Context Protocol (MCP). LLMs can query tasks, propose enrichments (with user approval), mine productivity analytics, and traverse task relationship graphs across providers. Core design principle: LLMs propose, users approve — no direct task modification.

**Prerequisites:** Epic 13 ✅ (Multi-Source Aggregation), Epic 6 ✅ (Enrichment DB)
**FRs covered:** FR81, FR82, FR83, FR84, FR85, FR86, FR87, FR88
**Origin:** LLM Integration & MCP Server Research (2026-03-06). Research document at `docs/research/llm-integration-mcp.md`.
**Architecture:** Separate binary (`cmd/threedoors-mcp/`) sharing `internal/` packages. No new storage layer — reads same YAML, JSONL, and SQLite as TUI.
**Status:** Not Started

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

## Epic 23: CLI Interface

**Epic Goal:** Provide a complete non-TUI CLI interface for ThreeDoors that serves both human power users (scriptable task management) and LLM agents (structured JSON output). The CLI shares `internal/core` with the TUI — no domain logic duplication. `threedoors` with no args launches the TUI (backward compatible); any subcommand routes to the Cobra-based CLI.

**Prerequisites:** None (core domain layer is already CLI-ready with JSON struct tags)
**Framework:** Cobra (`github.com/spf13/cobra`) for subcommand routing, shell completions, and help generation
**Origin:** CLI interface design research (`docs/research/cli-interface-design.md`)
**Architecture:** Layered CLI/TUI coexistence — `internal/cli/` imports `internal/core/`, never `internal/tui/`
**Status:** Not Started

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

---

## Epic 25: Todoist Integration

**Priority:** P1 — High value, all infrastructure in place
**Status:** Not Started (0/4 stories)
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
**Status:** Not Started (0/4 stories)
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

#### Story 26.1: GitHub SDK Client & Auth Configuration

**Status:** Not Started | **Priority:** P1 | **Depends On:** Epic 7 (done)

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

#### Story 26.2: Read-Only GitHub Provider with Field Mapping

**Status:** Not Started | **Priority:** P1 | **Depends On:** 26.1

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

#### Story 26.3: Bidirectional Sync & WAL Integration

**Status:** Not Started | **Priority:** P1 | **Depends On:** 26.2

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

#### Story 26.4: Contract Tests & Integration Testing

**Status:** Not Started | **Priority:** P1 | **Depends On:** 26.2

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
