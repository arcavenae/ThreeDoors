# Epic List

## Phase 1: Technical Demo & Validation COMPLETE

**Epic 1: Three Doors Technical Demo** COMPLETE
- **Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists
- **Status:** COMPLETE -- All stories implemented and merged (PRs #2, #4, #5, #7, #8, #13, #16, #18)
- **Stories:** 1.1 (Project Setup), 1.2 (Display Three Doors), 1.3 (Door Selection & Status Management), 1.3a (Quick Search & Command Palette), 1.5 (Session Metrics Tracking), 1.6 (Essential Polish), 1.7 (CI/CD Pipeline)
- **Tech Stack:** Go 1.25.4+, Bubbletea/Lipgloss, local YAML files, JSONL metrics
- **Result:** Concept validated through daily use; proceed to Full MVP

---

## Phase 2: Post-Validation Roadmap COMPLETE (Epics 2-3, 3.5, 5), IN PROGRESS (Epic 4, 6)

**Epic 2: Foundation & Apple Notes Integration** COMPLETE
- **Goal:** Replace text file backend with Apple Notes integration, enabling mobile task editing while maintaining Three Doors UX
- **Status:** COMPLETE -- All 6 stories implemented and merged (PRs #15, #17, #19, #20, #21, #22)
- **Deliverables:**
  - Refactor to adapter pattern (text file + Apple Notes backends)
  - Bidirectional sync with Apple Notes
  - Health check command for Notes connectivity
- **FRs covered:** FR2, FR4, FR5, FR12, FR15

**Epic 3: Enhanced Interaction & Task Context** COMPLETE
- **Goal:** Add task capture, values/goals display, and basic feedback mechanisms to improve task management workflow
- **Status:** COMPLETE -- All 7 stories implemented and merged (PRs #23-#31)
- **Deliverables:**
  - Quick add mode for task capture
  - Extended capture with "why" context
  - Values/goals setup and persistent display
  - Door feedback options (Blocked, Not now, Needs breakdown)
  - Daily completion tracking, improvement prompt, enhanced navigation
- **FRs covered:** FR3, FR6, FR7, FR8, FR9, FR10, FR16, FR18, FR19

**Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging)** COMPLETE
- **Goal:** Refactor core architecture, harden adapters, establish test infrastructure, and resolve tech debt from rapid Epic 1-3 implementation to prepare for Epic 4+ work
- **Prerequisites:** Epic 3 complete
- **Status:** COMPLETE -- All 8 stories complete (PRs #90-#97)
- **Deliverables:**
  - Core domain extraction (split internal/tasks into internal/core + adapter packages)
  - TaskProvider interface hardening (formalize Watch, HealthCheck, ChangeEvent)
  - Config.yaml schema & migration spike
  - Apple Notes adapter hardening (timeouts, retries, error categorization)
  - Baseline regression test suite for door selection and task management
  - Session metrics reader library for Epic 4
  - Adapter test scaffolding & CI coverage floor
  - Validation gate decision documentation
- **Stories:** 3.5.1-3.5.8 (8 stories)
- **Origin:** Party mode bridging discussion (2026-03-02)

**Epic 4: Learning & Intelligent Door Selection** COMPLETE
- **Goal:** Use historical session metrics (captured in Epic 1 Story 1.5) to analyze user patterns and adapt door selection to improve task engagement and completion rates
- **Prerequisites:** Epic 3 complete, Epic 3.5 stories 3.5.5/3.5.6 complete
- **Status:** COMPLETE -- All 6 stories complete (PRs #40, #42-#45, #82)
- **Data Foundation:** Epic 1 Story 1.5 captures door position selections, task bypasses, status changes, and mood/emotional context -- essential for pattern analysis
- **Deliverables:**
  - Task categorization (type, effort level, context)
  - Pattern recognition (which task types are selected vs bypassed)
  - Mood correlation analysis (emotional states -> task selection/avoidance patterns)
  - Avoidance detection (tasks repeatedly shown but never selected)
  - Adaptive selection based on current mood state and historical patterns
  - User insights ("When stressed, you avoid complex tasks")
  - Goal re-evaluation prompts when persistent avoidance detected
  - "Better than yesterday" multi-dimensional tracking
- **Stories:** 4.1-4.6 (6 stories)
- **FRs covered:** FR20, FR21

**Epic 5: macOS Distribution & Packaging** COMPLETE
- **Goal:** Provide a trusted, seamless installation experience on macOS by signing, notarizing, and packaging the binary so Gatekeeper does not quarantine it
- **Status:** COMPLETE -- Story 5.1 consolidated all deliverables (PR #30)
- **Independence:** This epic is independent of the story pipeline
- **FRs covered:** FR22-FR26

**Epic 6: Data Layer & Enrichment (Optional)** COMPLETE
- **Goal:** Add enrichment storage layer for metadata that cannot live in source systems
- **Status:** COMPLETE -- All 2 stories complete (PRs #53, #56)
- **Deliverables:**
  - SQLite enrichment database
  - Cross-reference tracking (tasks across multiple systems)
  - Metadata not supported by Apple Notes (categories, learning patterns, etc.)
- **Stories:** 6.1-6.2 (2 stories)
- **FRs covered:** FR11

---

## Phase 3: Platform Expansion & Intelligence (Post-MVP)

**Epic 7: Plugin/Adapter SDK & Registry** COMPLETE
- **Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide. Unblocks all future integrations.
- **Prerequisites:** Epic 2
- **Status:** COMPLETE -- All 3 stories complete (PRs #68, #70, #72)
- **Deliverables:**
  - Adapter registry with runtime discovery and loading
  - Config-driven provider selection via `~/.threedoors/config.yaml`
  - Adapter developer guide and interface specification
  - Contract test suite for adapter compliance validation
- **Stories:** 7.1-7.3 (3 stories)
- **FRs covered:** FR31, FR32, FR33

**Epic 8: Obsidian Integration (P0 - #2 Integration)** COMPLETE
- **Goal:** Add Obsidian vault as second task storage backend after Apple Notes. Local-first Markdown integration with bidirectional sync.
- **Prerequisites:** Epic 7
- **Status:** COMPLETE -- All 4 stories complete (PRs #73, #75, #77, #79)
- **Deliverables:**
  - Obsidian vault reader/writer adapter
  - Bidirectional sync with external vault changes
  - Vault configuration (path, folder, file naming) via config.yaml
  - Daily note integration for task read/write
- **Stories:** 8.1-8.4 (4 stories)
- **FRs covered:** FR27, FR28, FR29, FR30

**Epic 9: Testing Strategy & Quality Gates** PARTIAL
- **Goal:** Establish comprehensive testing infrastructure with integration, contract, performance, and E2E tests
- **Prerequisites:** Epic 2, Epic 7
- **Status:** PARTIAL -- 2/5 stories complete (PRs #83, #89). Stories 9.3-9.5 pending.
- **Deliverables:**
  - Apple Notes integration E2E tests
  - Contract tests for adapter compliance
  - Performance benchmarks (<100ms NFR validation)
  - Functional E2E tests for full user workflows
  - CI coverage gates preventing regression
- **Stories:** 9.1-9.5 (5 stories)
- **FRs covered:** FR49, FR50, FR51

**Epic 10: First-Run Onboarding Experience** COMPLETE
- **Goal:** Provide a guided welcome flow for new users to set up values/goals, understand Three Doors, learn key bindings, and optionally import existing tasks
- **Prerequisites:** Epic 3
- **Status:** COMPLETE -- All 2 stories complete (PRs #55, #59)
- **Deliverables:**
  - Welcome flow with Three Doors concept explanation
  - Values/goals setup wizard
  - Key bindings walkthrough
  - Import from existing task sources
- **Stories:** 10.1-10.2 (2 stories)
- **FRs covered:** FR38, FR39

**Epic 11: Sync Observability & Offline-First** COMPLETE
- **Goal:** Ensure robust offline-first operation with local queue, sync status visibility, conflict visualization, and sync debugging
- **Prerequisites:** Epic 2
- **Status:** COMPLETE -- All 3 stories complete (PRs #62, #66, #85)
- **Deliverables:**
  - Offline-first local change queue with replay
  - Sync status indicator in TUI per provider
  - Conflict visualization and resolution UI
  - Sync log for debugging
- **Stories:** 11.1-11.3 (3 stories)
- **FRs covered:** FR40, FR41, FR42, FR43

**Epic 12: Calendar Awareness (Local-First, No OAuth)** COMPLETE
- **Goal:** Add time-contextual door selection by reading local calendar sources. No OAuth, no cloud APIs.
- **Prerequisites:** Epic 4
- **Status:** COMPLETE -- All 2 stories complete (PRs #65, #81)
- **Deliverables:**
  - macOS Calendar.app reader via AppleScript
  - .ics file parser
  - CalDAV cache reader
  - Time-contextual door selection based on available time blocks
- **Stories:** 12.1-12.2 (2 stories)
- **FRs covered:** FR44, FR45

**Epic 13: Multi-Source Task Aggregation View** PARTIAL
- **Goal:** Unified cross-provider task pool with dedup detection and source attribution in the TUI
- **Prerequisites:** Epic 7, Epic 8
- **Status:** PARTIAL -- 1/2 stories complete (PR #84). Story 13.2 pending.
- **Deliverables:**
  - Cross-provider task pool aggregation
  - Duplicate detection across providers
  - Source attribution display in TUI
- **Stories:** 13.1-13.2 (2 stories)
- **FRs covered:** FR46, FR47, FR48

---

## Phase 4: Future Expansion

**Epic 14: LLM Task Decomposition & Agent Action Queue** COMPLETE
- **Goal:** Enable LLM-powered task breakdown where selected tasks are decomposed into stories/specs output to git repos for coding agent pickup
- **Prerequisites:** Epic 3+
- **Status:** COMPLETE -- All 2 stories complete (PRs #63, #87)
- **Deliverables:**
  - Spike: prompt engineering, output schema, git automation, agent handoff
  - LLM-generated BMAD-style stories/specs
  - Git repo structure output for Claude Code / multiclaude pickup
  - Configurable LLM backend (local vs cloud)
- **Stories:** 14.1-14.2 (2 stories)
- **FRs covered:** FR35, FR36, FR37

**Epic 15: Psychology Research & Validation** COMPLETE
- **Goal:** Build evidence base for ThreeDoors design decisions through literature review and validation studies
- **Prerequisites:** None (can run in parallel with development)
- **Status:** COMPLETE -- All 2 stories complete (PRs #60, #61)
- **Deliverables:**
  - Literature review: choice architecture (why 3 doors?)
  - Mood-task correlation validation study
  - Procrastination intervention research summary
  - Evidence for "progress over perfection" as motivational framework
  - Findings feed into Epic 4 learning algorithm refinement
- **Stories:** 15.1-15.2 (2 stories)
- **FRs covered:** FR34

**Epic 16: iPhone Mobile App (SwiftUI)**
- **Goal:** Bring the Three Doors experience to iPhone with a native SwiftUI app that shares tasks via Apple Notes and syncs seamlessly with the desktop TUI
- **Prerequisites:** Epic 2 (Apple Notes integration), Epic 3.5 (platform readiness for shared specs)
- **Status:** Not Started
- **Deliverables:**
  - Native SwiftUI iPhone app with swipeable Three Doors card carousel
  - Apple Notes integration via Swift (reading tasks from same note as TUI)
  - Task completion and status changes from mobile
  - Session metrics collection compatible with desktop JSONL format
  - iCloud Drive sync for config and metrics
  - TestFlight distribution (App Store submission in Phase 2)
- **Stories:** 16.1-16.7 (7 stories)
- **Estimated Effort:** 6-8 weeks at 4-6 hrs/week
- **Tech Stack:** Swift 5.9+, SwiftUI, CloudKit/iCloud Drive, Xcode 16+, iOS 17+ target
- **FRs covered:** (mobile-specific, not yet in PRD FRs)
- **Research:** See `docs/research/mobile-app-research.md` for full analysis

**Epic 17: Door Theme System** IN PROGRESS
- **Goal:** Replace the uniform rounded-border door appearance with visually distinct themed doors using ASCII/ANSI art frames, with user-selectable themes via onboarding, settings view, and config.yaml
- **Prerequisites:** Epic 3 (enhanced interaction), Epic 10 (onboarding -- for theme picker integration, can proceed independently)
- **Status:** IN PROGRESS -- Stories 17.1-17.6 being implemented
- **Deliverables:**
  - DoorTheme type, ThemeColors, and theme registry (`internal/tui/themes/`)
  - Classic theme wrapper (preserves current Lipgloss border rendering)
  - Three new themes: Modern/Minimalist, Sci-Fi/Spaceship, Japanese Shoji
  - DoorsView integration -- load theme from config, apply in View()
  - Theme picker in first-run onboarding flow (horizontal preview, arrow key browsing)
  - Settings view -- `:theme` command with live preview
  - Config.yaml persistence for theme selection
  - Width-aware fallback to Classic theme at narrow terminal widths
  - Golden file tests for all themes at multiple widths and selection states
- **Stories:** 17.1 (Theme Types & Registry), 17.2 (Theme Implementations), 17.3 (DoorsView Integration), 17.4 (Onboarding Theme Picker), 17.5 (Settings Theme Command), 17.6 (Golden File Tests)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR55, FR56, FR57, FR58, FR59, FR60, FR61, FR62
- **NFRs covered:** NFR17, NFR18, NFR19
- **Research:** See `docs/research/door-themes-research.md`, `docs/research/door-themes-analyst-review.md`, `docs/research/door-themes-party-mode.md`

**Epic 18: Docker E2E & Headless TUI Testing Infrastructure** COMPLETE
- **Goal:** Establish reproducible, automated E2E testing using Docker containers and Bubbletea's `teatest` package for headless TUI interaction testing
- **Prerequisites:** Epic 3, Epic 9 (Stories 9.4, 9.5)
- **Status:** COMPLETE -- All stories implemented and merged
- **Deliverables:**
  - Headless TUI test harness using `teatest` (pseudo-TTY, programmatic key input, model assertions)
  - Golden file snapshot tests for TUI visual regression detection
  - Input sequence replay tests for complete user workflow validation
  - Docker-based reproducible test environment (`Dockerfile.test` + `docker-compose.test.yml`)
  - CI integration running Docker E2E tests on every PR
- **Stories:** 18.1 (Headless Harness), 18.2 (Golden Files), 18.3 (Workflow Replay), 18.4 (Docker Environment), 18.5 (CI Integration)
- **FRs covered:** FR52, FR53, FR54

**Epic 19: Jira Integration**
- **Goal:** Integrate Jira as a task source via read-only adapter (Phase 1) and bidirectional sync (Phase 2), enabling developers to see their Jira issues as ThreeDoors tasks
- **Prerequisites:** Epic 7 (adapter SDK), Epic 11 (sync observability), Epic 13 (multi-source aggregation)
- **Status:** Not Started
- **Deliverables:**
  - Thin Jira REST API v3 HTTP client (auth, search, pagination, rate limits)
  - JiraProvider implementing TaskProvider (JQL search, field mapping)
  - Bidirectional sync via transitions API + WAL queuing
  - Configurable status/priority mapping and JQL in config.yaml
- **Stories:** 19.1 (HTTP Client), 19.2 (Read-Only Provider), 19.3 (Bidirectional Sync), 19.4 (Config & Registration)
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week
- **FRs covered:** FR63, FR64, FR65, FR66
- **Research:** See `docs/research/jira-integration-research.md`, `docs/research/task-sync-analyst-brief.md`

**Epic 20: Apple Reminders Integration**
- **Goal:** Add Apple Reminders as a task source with full CRUD support, leveraging structured data model (persistent IDs, native priority/due dates) for a higher-quality integration than Apple Notes
- **Prerequisites:** Epic 7 (adapter SDK), macOS only
- **Status:** Not Started
- **Deliverables:**
  - JXA scripts for reading, creating, updating, completing, and deleting reminders
  - RemindersProvider implementing TaskProvider with CommandExecutor pattern
  - Field mapping (priority -> effort, completion -> status, list -> source)
  - Configurable list filtering in config.yaml
- **Stories:** 20.1 (JXA Scripts & CommandExecutor), 20.2 (Read-Only Provider), 20.3 (Write Support), 20.4 (Config, Registration & Health Check)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR67, FR68, FR69
- **Research:** See `docs/research/apple-reminders-integration-research.md`, `docs/research/task-sync-analyst-brief.md`

**Epic 21: Sync Protocol Hardening**
- **Goal:** Harden the sync architecture for reliable multi-provider operation with background scheduling, fault isolation, and cross-provider identity mapping
- **Prerequisites:** Epic 11 (sync observability), Epic 13 (multi-source aggregation)
- **Status:** Not Started
- **Deliverables:**
  - Sync scheduler with per-provider independent loops and adaptive intervals
  - Circuit breaker per provider (Closed/Open/Half-Open states)
  - Canonical ID mapping via SourceRef for cross-provider deduplication
  - Enhanced sync status dashboard with staleness indicators
- **Stories:** 21.1 (Sync Scheduler), 21.2 (Circuit Breaker), 21.3 (Canonical ID Mapping), 21.4 (Dashboard Enhancements)
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week
- **FRs covered:** FR70, FR71, FR72
- **Research:** See `docs/research/sync-architecture-scaling-research.md`, `docs/research/task-sync-analyst-brief.md`

**Epic 22: Self-Driving Development Pipeline**
- **Goal:** Enable ThreeDoors tasks to directly trigger multiclaude worker agents, creating a closed loop where the app dispatches its own development work and tracks results (PRs, CI status) back in the TUI
- **Prerequisites:** Epic 14 (LLM Decomposition -- provides AgentService for optional story generation), multiclaude installed and configured
- **Status:** Not Started
- **Deliverables:**
  - DevDispatch data model and file-based queue persistence (`~/.threedoors/dev-queue.yaml`)
  - Dispatch engine wrapping multiclaude CLI (`CreateWorker`, `ListWorkers`, `GetHistory`, `RemoveWorker`)
  - TUI dispatch key binding ('x' in detail view) and `:dispatch` command
  - Dev queue view (list, approve, kill queue items)
  - Worker status polling via `tea.Tick` (30-second intervals)
  - Auto-generated follow-up tasks (review PRs, fix CI, address comments)
  - Optional story file generation via existing `AgentService`
  - Safety guardrails (max concurrent workers, approval gate, rate limiting, audit log)
- **Stories:** 22.1-22.8 (8 stories)
- **Estimated Effort:** 4-6 weeks at 2-4 hrs/week
- **FRs covered:** FR73, FR74, FR75, FR76, FR77, FR78, FR79, FR80
- **NFRs covered:** NFR24, NFR25, NFR26, NFR27
- **Research:** See `docs/research/self-driving-development-pipeline.md`

**Epic 23+: Additional Integrations** (Todoist, Linear, GitHub Issues, ClickUp, etc.)
**Epic 24+: Cross-Computer Sync** (Implement alternative to monolithic SQLite on cloud storage)
**Epic 25+: Advanced Features** (Voice interface, web interface, Apple Watch, iPad, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---

## Story Count Summary

| Epic | Stories | Status |
|------|---------|--------|
| Epic 0: Infrastructure & Process (Backfill) | 19 | Complete |
| Epic 1: Technical Demo | 7 | Complete |
| Epic 2: Apple Notes Integration | 6 | Complete |
| Epic 3: Enhanced Interaction | 7 | Complete |
| Epic 3.5: Platform Readiness (Bridging) | 8 | Complete |
| Epic 4: Learning & Door Selection | 6 | Complete |
| Epic 5: macOS Distribution | 1 | Complete |
| Epic 6: Data Layer (Optional) | 2 | Complete |
| Epic 7: Plugin/Adapter SDK | 3 | Complete |
| Epic 8: Obsidian Integration | 4 | Complete |
| Epic 9: Testing Strategy | 5 | Partial (2/5) |
| Epic 10: Onboarding | 2 | Complete |
| Epic 11: Sync Observability | 3 | Complete |
| Epic 12: Calendar Awareness | 2 | Complete |
| Epic 13: Multi-Source Aggregation | 2 | Partial (1/2) |
| Epic 14: LLM Decomposition | 2 | Complete |
| Epic 15: Psychology Research | 2 | Complete |
| Epic 16: iPhone Mobile App | 7 | Not Started |
| Epic 17: Door Theme System | 6 | In Progress |
| Epic 18: Docker E2E & Headless TUI Testing | 5 | Complete |
| Epic 19: Jira Integration | 4 | Not Started |
| Epic 20: Apple Reminders Integration | 4 | Not Started |
| Epic 21: Sync Protocol Hardening | 4 | Not Started |
| Epic 22: Self-Driving Dev Pipeline | 8 | Not Started |
| **Total** | **119** | **81 complete, 3 partial, 35 remaining** |

---
