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

**Epic 9: Testing Strategy & Quality Gates** COMPLETE
- **Goal:** Establish comprehensive testing infrastructure with integration, contract, performance, and E2E tests
- **Prerequisites:** Epic 2, Epic 7
- **Status:** COMPLETE -- All 5 stories implemented and merged (PRs #83, #89, #142, #103, #102).
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

**Epic 13: Multi-Source Task Aggregation View** COMPLETE
- **Goal:** Unified cross-provider task pool with dedup detection and source attribution in the TUI
- **Prerequisites:** Epic 7, Epic 8
- **Status:** COMPLETE -- All 2 stories implemented and merged (PRs #84, #143).
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
- **Research:** See `../../_bmad-output/planning-artifacts/mobile-app-research.md` for full analysis

**Epic 17: Door Theme System** COMPLETE
- **Goal:** Replace the uniform rounded-border door appearance with visually distinct themed doors using ASCII/ANSI art frames, with user-selectable themes via onboarding, settings view, and config.yaml
- **Prerequisites:** Epic 3 (enhanced interaction), Epic 10 (onboarding -- for theme picker integration, can proceed independently)
- **Status:** COMPLETE -- All 6 stories implemented and merged (PRs #119-#124)
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
- **Research:** See `../../_bmad-output/planning-artifacts/door-themes-research.md`, `../../_bmad-output/planning-artifacts/door-themes-analyst-review.md`, `../../_bmad-output/planning-artifacts/door-themes-party-mode.md`

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

**Epic 19: Jira Integration** COMPLETE
- **Goal:** Integrate Jira as a task source via read-only adapter (Phase 1) and bidirectional sync (Phase 2), enabling developers to see their Jira issues as ThreeDoors tasks
- **Prerequisites:** Epic 7 (adapter SDK), Epic 11 (sync observability), Epic 13 (multi-source aggregation)
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #132, #138, #150, #153)
- **Deliverables:**
  - Thin Jira REST API v3 HTTP client (auth, search, pagination, rate limits)
  - JiraProvider implementing TaskProvider (JQL search, field mapping)
  - Bidirectional sync via transitions API + WAL queuing
  - Configurable status/priority mapping and JQL in config.yaml
- **Stories:** 19.1 (HTTP Client), 19.2 (Read-Only Provider), 19.3 (Bidirectional Sync), 19.4 (Config & Registration)
- **FRs covered:** FR63, FR64, FR65, FR66
- **Research:** See `../../_bmad-output/planning-artifacts/jira-integration-research.md`, `../../_bmad-output/planning-artifacts/task-sync-analyst-brief.md`

**Epic 20: Apple Reminders Integration** COMPLETE
- **Goal:** Add Apple Reminders as a task source with full CRUD support, leveraging structured data model (persistent IDs, native priority/due dates) for a higher-quality integration than Apple Notes
- **Prerequisites:** Epic 7 (adapter SDK), macOS only
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #137, #148, #155, #158)
- **Deliverables:**
  - JXA scripts for reading, creating, updating, completing, and deleting reminders
  - RemindersProvider implementing TaskProvider with CommandExecutor pattern
  - Field mapping (priority -> effort, completion -> status, list -> source)
  - Configurable list filtering in config.yaml
- **Stories:** 20.1 (JXA Scripts & CommandExecutor), 20.2 (Read-Only Provider), 20.3 (Write Support), 20.4 (Config, Registration & Health Check)
- **FRs covered:** FR67, FR68, FR69
- **Research:** See `../../_bmad-output/planning-artifacts/apple-reminders-integration-research.md`, `../../_bmad-output/planning-artifacts/task-sync-analyst-brief.md`

**Epic 21: Sync Protocol Hardening** COMPLETE
- **Goal:** Harden the sync architecture for reliable multi-provider operation with background scheduling, fault isolation, and cross-provider identity mapping
- **Prerequisites:** Epic 11 (sync observability), Epic 13 (multi-source aggregation)
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #139, #132, #151, #157)
- **Deliverables:**
  - Sync scheduler with per-provider independent loops and adaptive intervals
  - Circuit breaker per provider (Closed/Open/Half-Open states)
  - Canonical ID mapping via SourceRef for cross-provider deduplication
  - Enhanced sync status dashboard with staleness indicators
- **Stories:** 21.1 (Sync Scheduler), 21.2 (Circuit Breaker), 21.3 (Canonical ID Mapping), 21.4 (Dashboard Enhancements)
- **FRs covered:** FR70, FR71, FR72
- **Research:** See `../../_bmad-output/planning-artifacts/sync-architecture-scaling-research.md`, `../../_bmad-output/planning-artifacts/task-sync-analyst-brief.md`

**Epic 22: Self-Driving Development Pipeline** COMPLETE
- **Goal:** Enable ThreeDoors tasks to directly trigger multiclaude worker agents, creating a closed loop where the app dispatches its own development work and tracks results (PRs, CI status) back in the TUI
- **Prerequisites:** Epic 14 (LLM Decomposition -- provides AgentService for optional story generation), multiclaude installed and configured
- **Status:** COMPLETE -- All 8 stories implemented and merged (PRs #149, #152, #163, #162, #161, #164, #159, #160)
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
- **FRs covered:** FR73, FR74, FR75, FR76, FR77, FR78, FR79, FR80
- **NFRs covered:** NFR24, NFR25, NFR26, NFR27
- **Research:** See `../../_bmad-output/planning-artifacts/self-driving-development-pipeline.md`

**Epic 23: CLI Interface** COMPLETE
- **Goal:** Provide a complete non-TUI CLI interface for ThreeDoors that serves both human power users (scriptable task management) and LLM agents (structured JSON output)
- **Prerequisites:** None (core domain layer is already CLI-ready with JSON struct tags)
- **Status:** COMPLETE -- All 11 stories implemented and merged (PRs #170-#195, #225)
- **Deliverables:**
  - Cobra-based CLI scaffold with `--json` persistent flag and output formatter
  - Task CRUD commands (`task list`, `task show`, `task add`, `task update`, `task complete`)
  - Door selection commands (`doors`, `doors pick`, `doors roll`)
  - Session and analytics commands
  - Shell completions (bash, zsh, fish, PowerShell)
- **Stories:** 23.1-23.11 (11 stories)
- **FRs covered:** CLI interface requirements

**Epic 24: MCP/LLM Integration Server** COMPLETE
- **Goal:** Expose ThreeDoors functionality as an MCP (Model Context Protocol) server enabling LLM agents to interact with tasks programmatically
- **Prerequisites:** Epic 23 (CLI Interface)
- **Status:** COMPLETE -- All 8 stories implemented and merged (PRs #177-#197)
- **Deliverables:**
  - MCP server with tool definitions for task management
  - Structured JSON responses for LLM consumption
  - Resource endpoints for task context
  - Integration with existing TaskProvider infrastructure
- **Stories:** 24.1-24.8 (8 stories)

**Epic 25: Todoist Integration** COMPLETE
- **Goal:** Integrate Todoist as a task source via REST API v1 with thin HTTP client, read-only then bidirectional sync
- **Prerequisites:** Epic 7 (Adapter SDK -- complete)
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #308, #321, plus Stories 25.3 & 25.4)
- **Stories:** 25.1-25.4 (4 stories)

**Epic 26: GitHub Issues Integration** COMPLETE
- **Goal:** Integrate GitHub Issues as a task source, enabling developers to see their GitHub issues as ThreeDoors tasks
- **Prerequisites:** Epic 7 (Adapter SDK -- complete)
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #201-#205)
- **Stories:** 26.1-26.4 (4 stories)

**Epic 27: Daily Planning Mode** (P1)
- **Goal:** Add a guided daily planning ritual that transforms ThreeDoors from a reactive task picker into a proactive morning engagement tool, driving long-term retention through structured planning sessions
- **Prerequisites:** Epic 1 (session tracking), Epic 3 (mood capture, values/goals flow patterns), Epic 4 (task categorization)
- **Status:** COMPLETE
- **Deliverables:**
  - Planning data model with session-scoped `+focus` tag and energy level constants
  - Review incomplete tasks flow (continue/defer/drop quick triage)
  - Focus selection flow (pick 3-5 tasks from pool, filtered by energy)
  - Energy level matching with time-of-day inference and user override
  - Planning session metrics (JSONL `planning_session` event type)
  - Focus-aware door selection scoring boost
  - CLI `threedoors plan` subcommand and TUI `:plan` command
- **Stories:** 27.1-27.5 (5 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR97, FR98, FR99, FR100, FR101, FR102, FR103
- **Research:** See `../../_bmad-output/planning-artifacts/ux-workflow-improvements-research.md` (Improvement #2: Daily Planning Mode)

**Epic 28: Snooze/Defer as First-Class Action** (P1)
- **Goal:** Surface existing `StatusDeferred` as a first-class user action with date-based snooze and auto-return
- **Prerequisites:** None
- **Status:** Not Started
- **Stories:** 28.1-28.4 (4 stories)

**Epic 29: Task Dependencies & Blocked-Task Filtering** (P1)
- **Goal:** Native dependency graph support. Blocks tasks with unmet dependencies from door selection
- **Prerequisites:** None
- **Status:** In Progress (3/4 stories done; 29.3 remaining)
- **Stories:** 29.1-29.4 (4 stories)

**Epic 30: Linear Integration** (P2)
- **Goal:** Integrate Linear as a task source for engineering teams via the Linear GraphQL API, leveraging Linear's excellent task model alignment (rich workflow states, priority, estimates, labels, due dates) for high-fidelity task import
- **Prerequisites:** Epic 7 (Adapter SDK — complete), Epic 13 (Multi-Source Aggregation — complete)
- **Status:** Not Started
- **Deliverables:**
  - Linear GraphQL client with typed queries, cursor-based pagination, and API key auth
  - Read-only LinearProvider with full field mapping (status, priority, effort, labels, due dates)
  - Bidirectional sync: complete tasks via GraphQL mutation, WAL offline queuing
  - Contract tests and integration tests with mocked GraphQL server
- **Stories:** 30.1-30.4 (4 stories)
- **Estimated Effort:** 4-5 days
- **FRs covered:** FR116, FR117, FR118, FR119
- **Research:** See `../../_bmad-output/planning-artifacts/task-source-expansion-research.md` (Linear section)

**Epic 31: Expand/Fork Key Implementations** (P2)
- **Goal:** Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9
- **Prerequisites:** None
- **Status:** Not Started
- **Stories:** 31.1-31.5 (5 stories)

**Epic 32: Undo Task Completion** (P1)
- **Goal:** Allow reversing accidental task completion via `complete -> todo` transition
- **Prerequisites:** None
- **Status:** Complete
- **Stories:** 32.1-32.3 (3 stories)

**Epic 33: Seasonal Door Theme Variants** (P2)
- **Goal:** Time-based seasonal theme variants that auto-switch based on current date, extending Epic 17's theme infrastructure
- **Prerequisites:** Epic 17 (Door Theme System -- complete)
- **Status:** Not Started
- **Stories:** 33.1-33.4 (4 stories)

**Epic 34: SOUL.md + Custom Development Skills** COMPLETE
- **Goal:** Create SOUL.md project philosophy document and 4 custom Claude Code slash commands (/pre-pr, /validate-adapter, /check-patterns, /new-story) to improve AI agent alignment and developer workflow
- **Prerequisites:** None (CLAUDE.md already exists)
- **Status:** COMPLETE -- All 4 stories implemented and merged (PRs #222, #224, #228, #230)
- **Deliverables:**
  - SOUL.md at project root — project philosophy, design principles, behavioral guidelines for AI agents
  - `/pre-pr` slash command — 8-step pre-PR validation automation
  - `/validate-adapter` slash command — TaskProvider compliance checking
  - `/check-patterns` slash command — design pattern violation scanning
  - `/new-story` slash command — story template generator referencing CLAUDE.md
- **Stories:** 34.1-34.4 (4 stories)
- **NFRs covered:** NFR-DX1, NFR-DX2, NFR-DX3, NFR-DX4, NFR-DX5, NFR-DX6
- **Research:** See `../../_bmad-output/planning-artifacts/ai-tooling-findings.md`

**Epic 35: Door Visual Appearance — Door-Like Proportions** COMPLETE
- **Goal:** Redesign all door themes to visually read as actual doors using portrait orientation, panel dividers, asymmetric handles, and threshold/floor lines
- **Prerequisites:** Epic 17 ✅ (Door Theme System)
- **Status:** COMPLETE -- All 7 stories implemented and merged (PRs #226, #229, #234, #236, #237, #238, #239)
- **Deliverables:**
  - DoorAnatomy helper type and height-aware Render signature
  - All 4 themes updated with door-like proportions (portrait orientation, panel dividers, handles, thresholds)
  - Compact mode fallback for short terminals
  - Shadow/depth effects for 3D appearance
  - Golden file test regeneration and accessibility validation
- **Stories:** 35.1-35.7 (7 stories)
- **FRs covered:** FR138-FR147
- **Research:** See `_bmad-output/planning-artifacts/door-appearance-party-mode.md`

**Epic 36: Door Selection Interaction Feedback** (P1)
- **Goal:** Make door selection feel responsive and satisfying by enhancing visual feedback contrast, adding deselect toggle, and ensuring universal quit. Addresses GitHub Issue #219.
- **Prerequisites:** None (complements Epic 35 but does not depend on it)
- **Status:** COMPLETE (4/4 stories done)
- **Stories:** 36.1-36.4 (4 stories)
- **FRs covered:** FR148-FR151

**Epic 37: Persistent BMAD Agent Infrastructure**
- **Goal:** Enable autonomous project governance by adding persistent BMAD agents and cron jobs that maintain story status, ROADMAP accuracy, architecture doc currency, and quality metrics
- **Prerequisites:** None
- **Status:** COMPLETE -- All 4 stories implemented (PR #271, PR #280, PR #279, PR #281)
- **Deliverables:**
  - Agent definition files for project-watchdog and arch-watchdog (`agents/`)
  - Cron configuration for SM sprint health (4h) and QA coverage audit (weekly) (`docs/quality/cron-setup.md`)
  - Agent communication architecture documentation (`_bmad-output/planning-artifacts/architecture-persistent-agent-infrastructure.md`)
  - Monitoring, tuning, and Phase 1 evaluation framework (`docs/operations/agent-evaluation.md`)
- **Stories:** 37.1-37.4 (4 stories)

**Epic 38: Dual Homebrew Distribution** (P1)
- **Goal:** Establish parallel Homebrew distribution channels (stable + alpha) with signing parity, publishing controls, verification, and retention management
- **Prerequisites:** GoReleaser release pipeline (complete), Apple Developer ID signing infrastructure (complete)
- **Status:** In Progress (2/6 done)
- **Deliverables:**
  - Alpha Homebrew formula (`threedoors-a.rb`) auto-updated on every push to main
  - Publishing toggle (`vars.ALPHA_TAP_ENABLED`) for controlled activation
  - Code signing and notarization for stable GoReleaser releases
  - Alpha formula verification via tap CI monitoring
  - Alpha release retention cleanup (keep last 30)
- **Stories:** 38.1-38.6 (6 stories)
- **Research:** See `../../_bmad-output/planning-artifacts/dual-homebrew-distribution-research.md`, `_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md`

**Epic 39: Keybinding Display System** (P1)
- **Goal:** Add toggleable keybinding discoverability to the TUI: a concise context-sensitive bar at the bottom of every view, a full keybinding overlay accessible via `?` key, and default-on inline key hints rendered directly on interactive elements. Door key indicators unified under `h` toggle (D-137).
- **Prerequisites:** None (all required infrastructure exists)
- **Status:** COMPLETE (12/13 done, 1 cancelled)
- **Deliverables:**
  - Compile-time keybinding registry mapping each ViewMode to its available key bindings
  - Concise bottom bar showing 5-6 priority keys per view, with Lipgloss dim styling (non-door views)
  - Full-screen keybinding overlay (`?` key) organized by category with scroll support
  - `h` key toggles key hints: door key indicators in doors view, bar in non-door views (D-137, D-138)
  - Terminal size adaptation (auto-hide bar on small terminals, compact mode, width truncation)
  - Context-sensitive bar content (changes per view mode)
  - Inline key hints `[a]`/`[w]`/`[d]` on door frames (doorknob metaphor), controlled by unified `h` toggle
  - ~~Auto-fade after N sessions~~ — Cancelled (D-140); manual toggle only
  - `:hints` command as alias for `h` toggle
- **Stories:** 39.1-39.13 (13 stories, 1 cancelled)
- **Research:** See `_bmad-output/planning-artifacts/keybinding-display-party-mode.md`, `_bmad-output/planning-artifacts/keybinding-display-ux-review.md`, `_bmad-output/planning-artifacts/keybinding-display-architecture.md`, `_bmad-output/planning-artifacts/default-tooltips-mode-party-mode.md`, `_bmad-output/planning-artifacts/door-key-indicators-course-correction.md`

**Epic 40: Beautiful Stats Display** (P1)
- **Goal:** Transform the insights dashboard from plain text into a visually delightful, SOUL-aligned celebration of user activity using Lipgloss styled panels, gradient sparklines, bar charts, fun facts, heatmaps, and milestone celebrations
- **Prerequisites:** None (all data infrastructure exists from Epics 1 and 4)
- **Status:** COMPLETE (10/10 stories done)
- **Deliverables:**
  - Phase 1: Stats dashboard shell with Lipgloss bordered panels and responsive layout
  - Phase 1: Gradient sparklines with color-blind safe palette (blue-teal-yellow)
  - Phase 1: Fun facts engine with celebration-oriented session insights
  - Phase 2: Horizontal bar charts for mood correlation
  - Phase 2: GitHub-style activity heatmap (8-week range)
  - Phase 2: Surface hidden session metrics (duration, fastest start, peak hour, etc.)
  - Phase 2: Animated counter reveals on view entry
  - Phase 2: Tab navigation for Overview/Detail views
  - Phase 3: Theme-matched stats color palettes
  - Phase 3: Milestone celebrations (4 thresholds, observation language only)
- **Stories:** 40.1-40.10 (10 stories)
- **Estimated Effort:** 17-23 hours across 3 phases
- **Research:** See `_bmad-output/planning-artifacts/beautiful-stats-research.md`

**Epic 41: Charm Ecosystem Adoption & TUI Polish** (P2)
- **Goal:** Systematically adopt underutilized charmbracelet ecosystem components to reduce custom code maintenance, improve UX consistency, and deliver on SOUL.md's "physical objects" promise
- **Prerequisites:** None hard. Stories 41.3-41.4 should follow Epic 39 keybinding overlay work to avoid conflicts.
- **Status:** Not Started
- **Deliverables:**
  - bubbles/spinner for async provider operation feedback
  - lipgloss.JoinVertical + Place for cleaner layout composition
  - bubbles/viewport replacing 3 custom scroll implementations (help, synclog, keybinding overlay)
  - harmonica spring-physics door transition spike (proof of concept + testing pattern)
  - Adaptive color profile support for terminal-aware color degradation
- **Stories:** 41.1-41.6 (6 stories)
- **Estimated Effort:** 10-15 days
- **Research:** See `_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md`
- **Decisions:** D-128 (viewport), D-129 (spinner), D-130 (layout), D-131 (harmonica spike), D-132 (reject list), D-133 (reject textarea/table/etc.), D-134 (epic number 41)

**Epic 49: ThreeDoors Doctor — Self-Diagnosis Command** (P1)
- **Goal:** Comprehensive self-diagnosis command (`threedoors doctor`) with flutter-style category-based output, conservative auto-repair, and channel-aware version checking. Supersedes existing `health` command.
- **Prerequisites:** Epic 23 (CLI Interface — complete)
- **Status:** Not Started
- **Deliverables:**
  - Doctor command skeleton with DoctorChecker framework and `health` alias
  - 6 check categories: Environment, Task Data, Providers, Sessions, Sync, Database
  - Channel-aware version checking with 24h cache (gh CLI pattern)
  - Conservative auto-repair via `--fix` flag (safe/reversible ops only)
  - Verbose mode (`-v`), category filter (`--category`), JSON output (`--json`)
  - flutter-style icons: `[✓]` pass, `[!]` warn, `[✗]` fail, `[i]` info, `[ ]` skip
- **Stories:** 49.1-49.10 (10 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Research:** See `_bmad-output/planning-artifacts/threedoors-doctor-research.md`
- **Decisions:** D-154 (doctor supersedes health), D-155 (flutter-style icons), D-156 (conservative auto-fix), D-157 (24h cached version check), D-158 (channel-aware version)

**Epic 42: Application Security Hardening** (P1)
- **Goal:** Remediate all actionable findings from the application security audit — standardize file permissions, add symlink validation, enforce input size limits, protect credentials, and harden CI supply chain
- **Prerequisites:** None
- **Status:** Not Started
- **Deliverables:**
  - Standardize all file permissions to 0o700 (dirs) / 0o600 (files) with startup migration for existing installs
  - Symlink validation via `os.Lstat()` on startup and before file writes
  - File size limits before YAML reads, explicit scanner buffer limits on all JSONL readers
  - Credential exposure warning on startup, `yaml:"-"` on all token fields
  - SHA-pinned third-party GitHub Actions, govulncheck in CI quality gate
- **Stories:** 42.1-42.5 (5 stories)
- **Research:** See `_bmad-output/planning-artifacts/security-audit-application.md`
- **Decisions:** D-153 (epic creation and story grouping rationale)

**Epic 48: Door-Like Doors — Visual Door Metaphor Enhancement** (P2)
- **Goal:** Transform rectangular card/panel doors into visually convincing doors using side-mounted handles, hinge marks, threshold lines, crack-of-light selection feedback, and handle turn micro-animations
- **Prerequisites:** Epic 35 (Door Visual Appearance — complete), Epic 17 (Door Theme System — complete)
- **Status:** Not Started
- **Deliverables:**
  - Side-mounted handle at right edge + hinge marks on left edge (asymmetry)
  - Continuous threshold/floor line grounding all doors in shared space
  - Crack of light effect on door selection (door becomes "ajar")
  - Handle turn micro-animation synced with spring physics
- **Stories:** 48.1-48.4 (4 stories)
- **Research:** See `_bmad-output/planning-artifacts/doors-more-doorlike-research.md` (5-round party mode, 7 agents)
- **Decisions:** D-141 (adopt 5 proposals), X-080 through X-083 (4 rejected alternatives)

**Epic 43: Connection Manager Infrastructure** (P1)
- **Goal:** Build the connection lifecycle layer for data source integrations: state machine, credential storage (system keychain via 99designs/keyring), config schema v3 (named connections with ULID IDs), CRUD operations, sync event logging, and migration of existing 8 adapters to the new pattern
- **Prerequisites:** Epic 7 (Adapter SDK — complete), Epic 11 (Sync Observability — complete)
- **Status:** Not Started
- **Deliverables:**
  - ConnectionManager type with 7-state machine (Disconnected, Connecting, Connected, Syncing, Error, AuthExpired, Paused)
  - CredentialStore with keychain + env var + encrypted file fallback chain
  - Config schema v3 with `connections:` array, auto-migration from v2
  - Connection CRUD: add, remove, pause, resume, test, force-sync
  - JSONL sync event logging per connection with rolling retention
  - All existing adapters wrapped in ConnectionManager pattern
- **Stories:** 43.1-43.6 (6 stories)
- **Research:** See `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`
- **Decisions:** D-147 (keyring), D-149 (compiled-in providers), D-152 (named connections with ULIDs)

**Epic 44: Sources TUI** (P1)
- **Goal:** TUI interfaces for data source management: setup wizard (`:connect` command using charmbracelet/huh), sources dashboard (`:sources`), source detail view with health checks, sync log view, status bar health alerts, disconnection flow with task preservation, and re-authentication flow
- **Prerequisites:** Epic 43 (Connection Manager Infrastructure)
- **Status:** Not Started
- **Deliverables:**
  - 4-step setup wizard with provider-adaptive forms (API token, OAuth device code, local path)
  - Sources dashboard with status indicators (connected/paused/error/auth expired)
  - Source detail view with health checks and sync statistics
  - Sync log view with scrollable event history
  - Status bar alerts for connections needing attention
  - Disconnect and re-auth flows
- **Stories:** 44.1-44.7 (7 stories)
- **Research:** See `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`
- **Decisions:** D-150 (charmbracelet/huh for wizard)

**Epic 45: Sources CLI** (P1)
- **Goal:** Non-interactive CLI commands for data source management: `threedoors connect <provider>` with flags, `threedoors sources` (list/status/test/manage/log), and consistent `--json` output for scripting and CI/automation
- **Prerequisites:** Epic 43 (Connection Manager Infrastructure), Epic 23 (CLI Interface — complete)
- **Status:** Not Started
- **Deliverables:**
  - `threedoors connect <provider>` with per-provider flag sets
  - `threedoors sources` list, status, test, pause/resume/sync/disconnect commands
  - `threedoors sources log` with filtering
  - JSON output for all commands
- **Stories:** 45.1-45.5 (5 stories)
- **Research:** See `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`

**Epic 46: OAuth Device Code Flow** (P2)
- **Goal:** Generic OAuth device code flow client (RFC 8628) for browser-based authentication, with provider-specific integrations for GitHub and Linear, and silent token refresh lifecycle
- **Prerequisites:** None (consumed by Epics 44/45)
- **Status:** Not Started
- **Deliverables:**
  - Reusable device code flow client (request code, display URL, poll for token)
  - GitHub OAuth integration (device code + PAT fallback)
  - Linear OAuth/API key integration
  - Silent token refresh with explicit re-auth on expiry
- **Stories:** 46.1-46.4 (4 stories)
- **Research:** See `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`
- **Decisions:** D-148 (device code flow over callback server)

**Epic 47: Sync Lifecycle & Advanced Features** (P2)
- **Goal:** Advanced sync features: conflict resolution (last-writer-wins with field-level strategy), orphaned task handling (mark not delete), auto-detection of installed tools in setup wizard, and proactive connection health notifications
- **Prerequisites:** Epic 43, Epic 44
- **Status:** Not Started
- **Deliverables:**
  - ConflictResolver with remote-wins for metadata, local-wins for ThreeDoors fields
  - Orphaned task marking and management UI
  - Installed tool detection (gh CLI, Todoist config, Obsidian vaults)
  - Predictive warnings (token expiry, rate limits, error streaks)
- **Stories:** 47.1-47.4 (4 stories)
- **Research:** See `_bmad-output/planning-artifacts/data-source-setup-ux-research.md`
- **Decisions:** D-151 (conflict resolution strategy)

**Epic 50: In-App Bug Reporting** (P2)
- **Goal:** Add a `:bug` command for frictionless in-app bug reporting with navigation breadcrumb trail, automatic environment context, mandatory preview, and tiered submission (browser URL, GitHub API, local file)
- **Prerequisites:** None (standalone feature)
- **Status:** Not Started
- **Deliverables:**
  - Ring buffer breadcrumb tracking (50 entries, view transitions + non-text keys, privacy-safe)
  - Bug report view with text description input, environment summary, and mandatory preview
  - Three submission paths: browser URL (zero-auth), GitHub API (PAT upgrade), local file (offline)
  - Strict privacy allowlist at capture level — task content, search queries, and personal data never collected
- **Stories:** 50.1-50.3 (3 stories)
- **Research:** See `../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md`, `../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md`
- **Decisions:** D-112 (browser URL primary), D-113 (ring buffer), D-114 (allowlist privacy), D-115 (mandatory preview), D-116 (hardcoded target repo)

**Epic 51: SLAES — Self-Learning Agentic Engineering System** (P1)
- **Goal:** Build a continuous improvement meta-system with a persistent `retrospector` agent that monitors PR merges, detects process waste (saga patterns), audits doc consistency, analyzes CI/conflict patterns, and files improvement recommendations to BOARD.md. Dual-loop architecture: spec-chain quality and operational efficiency.
- **Prerequisites:** Epic 37 (Persistent BMAD Agents — complete)
- **Status:** Not Started
- **Deliverables:**
  - Phase 0: Retrospector agent definition in responsibility+WHY format; rewrite 5 operational agent definitions with incident-hardened guardrails
  - Phase 1 (MVP): JSONL findings log with per-merge lightweight retro, saga detection (2+ workers on same fix), doc consistency audit (periodic cross-check of planning docs), BOARD.md recommendation pipeline with confidence scoring
  - Phase 2: Merge conflict rate analysis (hot files, parallelization safety), CI failure taxonomy and spec-chain tracing, research lifecycle tracking, PR creation authority, weekly trend reporting
  - 5 Watchmen safeguards: no self-modification, audit trail, confidence scoring, periodic human review, kill switch (3 rejections → read-only)
- **Stories:** 51.1-51.10 (10 stories)
- **Research:** See `_bmad-output/planning-artifacts/agentic-engineering-agent-party-mode.md`, `_bmad-output/planning-artifacts/subagent-abuse-investigation.md`
- **Decisions:** D-1 (single agent), D-2 (SLAES/retrospector naming), D-3 (persistent 15-min polling), D-4 (Level 2 authority), D-5 (consumer model), D-6 (dual-loop), D-7 (per-PR + batch), D-8 (5 Watchmen safeguards), D-9 (mode rotation), D-10 (responsibility+WHY definitions)

**Epic 52+: Advanced Features** (Voice interface, web interface, Apple Watch, iPad, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---

## Story Count Summary

| Epic | Stories | Status |
|------|---------|--------|
| Epic 0: Infrastructure & Process (Backfill) | 14 | Partial (10/14) |
| Epic 1: Technical Demo | 7 | Complete |
| Epic 2: Apple Notes Integration | 6 | Complete |
| Epic 3: Enhanced Interaction | 7 | Complete |
| Epic 3.5: Platform Readiness (Bridging) | 8 | Complete |
| Epic 4: Learning & Door Selection | 6 | Complete |
| Epic 5: macOS Distribution | 1 | Complete |
| Epic 6: Data Layer (Optional) | 2 | Complete |
| Epic 7: Plugin/Adapter SDK | 3 | Complete |
| Epic 8: Obsidian Integration | 4 | Complete |
| Epic 9: Testing Strategy | 5 | Complete |
| Epic 10: Onboarding | 2 | Complete |
| Epic 11: Sync Observability | 3 | Complete |
| Epic 12: Calendar Awareness | 2 | Complete |
| Epic 13: Multi-Source Aggregation | 2 | Complete |
| Epic 14: LLM Decomposition | 2 | Complete |
| Epic 15: Psychology Research | 2 | Complete |
| Epic 16: iPhone Mobile App | 7 | Icebox |
| Epic 17: Door Theme System | 6 | Complete |
| Epic 18: Docker E2E & Headless TUI Testing | 5 | Complete |
| Epic 19: Jira Integration | 4 | Complete |
| Epic 20: Apple Reminders Integration | 4 | Complete |
| Epic 21: Sync Protocol Hardening | 4 | Complete |
| Epic 22: Self-Driving Dev Pipeline | 8 | Complete |
| Epic 23: CLI Interface | 11 | Complete |
| Epic 24: MCP/LLM Integration | 8 | Complete |
| Epic 25: Todoist Integration | 4 | Complete |
| Epic 26: GitHub Issues Integration | 4 | Complete |
| Epic 27: Daily Planning Mode | 5 | Complete |
| Epic 28: Snooze/Defer | 4 | Not Started |
| Epic 29: Task Dependencies | 4 | In Progress (3/4) |
| Epic 30: Linear Integration | 4 | Not Started |
| Epic 31: Expand/Fork Key | 5 | Not Started |
| Epic 32: Undo Task Completion | 3 | Complete |
| Epic 33: Seasonal Theme Variants | 4 | Not Started |
| Epic 34: SOUL.md + Custom Dev Skills | 4 | Complete |
| Epic 35: Door Visual Appearance | 7 | Complete |
| Epic 36: Door Selection Feedback | 4 | Complete |
| Epic 37: Persistent BMAD Agents | 4 | Complete |
| Epic 38: Dual Homebrew Distribution | 6 | In Progress (2/6) |
| Epic 39: Keybinding Display System | 13 | COMPLETE (12/13, 1 cancelled) |
| Epic 40: Beautiful Stats Display | 10 | Complete |
| Epic 41: Charm Ecosystem Adoption | 6 | Not Started |
| Epic 42: Application Security Hardening | 5 | Not Started |
| Epic 43: Connection Manager Infrastructure | 6 | Not Started |
| Epic 44: Sources TUI | 7 | Not Started |
| Epic 45: Sources CLI | 5 | Not Started |
| Epic 46: OAuth Device Code Flow | 4 | Not Started |
| Epic 47: Sync Lifecycle & Advanced Features | 4 | Not Started |
| Epic 48: Door-Like Doors | 4 | Not Started |
| Epic 49: ThreeDoors Doctor | 10 | Not Started |
| Epic 50: In-App Bug Reporting | 3 | Not Started |
| Epic 51: SLAES | 10 | Not Started |
| **Total** | **274** | **146 complete, 4 epics in progress, 128 not started** |
---
