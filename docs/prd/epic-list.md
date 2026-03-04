# Epic List

## Phase 1: Technical Demo & Validation ✅ COMPLETE

**Epic 1: Three Doors Technical Demo** ✅
- **Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists
- **Status:** COMPLETE — All stories implemented and merged (PRs #2, #4, #5, #7, #8, #13, #16, #18)
- **Stories:** 1.1 (Project Setup), 1.2 (Display Three Doors), 1.3 (Door Selection & Status Management), 1.3a (Quick Search & Command Palette), 1.5 (Session Metrics Tracking), 1.6 (Essential Polish), 1.7 (CI/CD Pipeline)
- **Tech Stack:** Go 1.25.4+, Bubbletea/Lipgloss, local YAML files, JSONL metrics
- **Result:** Concept validated through daily use; proceed to Full MVP

---

## Phase 2: Post-Validation Roadmap ✅ COMPLETE (Epics 2-3), IN PROGRESS (Epics 3.5-6)

**Epic 2: Foundation & Apple Notes Integration** ✅
- **Goal:** Replace text file backend with Apple Notes integration, enabling mobile task editing while maintaining Three Doors UX
- **Status:** COMPLETE — All 6 stories implemented and merged (PRs #15, #17, #19, #20, #21, #22)
- **Deliverables:**
  - Refactor to adapter pattern (text file + Apple Notes backends)
  - Bidirectional sync with Apple Notes
  - Health check command for Notes connectivity
- **FRs covered:** FR2, FR4, FR5, FR12, FR15

**Epic 3: Enhanced Interaction & Task Context** ✅
- **Goal:** Add task capture, values/goals display, and basic feedback mechanisms to improve task management workflow
- **Status:** COMPLETE — All 7 stories implemented and merged (PRs #23-#31)
- **Deliverables:**
  - Quick add mode for task capture
  - Extended capture with "why" context
  - Values/goals setup and persistent display
  - Door feedback options (Blocked, Not now, Needs breakdown)
  - Daily completion tracking, improvement prompt, enhanced navigation
- **FRs covered:** FR3, FR6, FR7, FR8, FR9, FR10, FR16, FR18, FR19

**Epic 3.5: Platform Readiness & Technical Debt Resolution (Bridging)** 🆕
- **Goal:** Refactor core architecture, harden adapters, establish test infrastructure, and resolve tech debt from rapid Epic 1-3 implementation to prepare for Epic 4+ work
- **Prerequisites:** Epic 3 complete ✅
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
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Blocks:** Epic 4 (partially), Epic 7, Epic 8, Epic 9, Epic 11
- **Origin:** Party mode bridging discussion (2026-03-02)

**Epic 4: Learning & Intelligent Door Selection**
- **Goal:** Use historical session metrics (captured in Epic 1 Story 1.5) to analyze user patterns and adapt door selection to improve task engagement and completion rates
- **Prerequisites:** Epic 3 complete ✅, Epic 3.5 stories 3.5.5/3.5.6 complete, sufficient usage data
- **Data Foundation:** Epic 1 Story 1.5 captures door position selections, task bypasses, status changes, and mood/emotional context—essential for pattern analysis
- **Deliverables:**
  - Task categorization (type, effort level, context)
  - Pattern recognition (which task types are selected vs bypassed)
  - Mood correlation analysis (emotional states → task selection/avoidance patterns)
  - Avoidance detection (tasks repeatedly shown but never selected)
  - Adaptive selection based on current mood state and historical patterns
  - User insights ("When stressed, you avoid complex tasks")
  - Goal re-evaluation prompts when persistent avoidance detected
  - "Better than yesterday" multi-dimensional tracking
- **Stories:** 4.1-4.6 (6 stories)
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week
- **FRs covered:** FR20, FR21
- **Risk:** Algorithm complexity; may need to simplify learning approach

**Epic 5: macOS Distribution & Packaging** ✅
- **Goal:** Provide a trusted, seamless installation experience on macOS by signing, notarizing, and packaging the binary so Gatekeeper does not quarantine it
- **Status:** COMPLETE — Story 5.1 consolidated all deliverables (PR #30)
- **Independence:** This epic is independent of the story pipeline
- **FRs covered:** FR22-FR26

**Epic 6: Data Layer & Enrichment (Optional)**
- **Goal:** Add enrichment storage layer for metadata that cannot live in source systems
- **Prerequisites:** Epic 4 complete; proven need for enrichment beyond what backends support
- **Deliverables:**
  - SQLite enrichment database
  - Cross-reference tracking (tasks across multiple systems)
  - Metadata not supported by Apple Notes (categories, learning patterns, etc.)
- **Stories:** 6.1-6.2 (2 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR11
- **Risk:** May be YAGNI; consider deferring indefinitely if not clearly needed

---

## Phase 3: Platform Expansion & Intelligence (Post-MVP)

**Epic 7: Plugin/Adapter SDK & Registry**
- **Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide. Unblocks all future integrations.
- **Prerequisites:** Epic 2 ✅, Epic 3.5 (stories 3.5.1, 3.5.2, 3.5.3)
- **Deliverables:**
  - Adapter registry with runtime discovery and loading
  - Config-driven provider selection via `~/.threedoors/config.yaml`
  - Adapter developer guide and interface specification
  - Contract test suite for adapter compliance validation
- **Stories:** 7.1-7.3 (3 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR31, FR32, FR33
- **Risk:** Over-engineering the plugin system; keep minimal until 3+ adapters exist

**Epic 8: Obsidian Integration (P0 - #2 Integration)**
- **Goal:** Add Obsidian vault as second task storage backend after Apple Notes. Local-first Markdown integration with bidirectional sync.
- **Prerequisites:** Epic 7 (adapter SDK)
- **Deliverables:**
  - Obsidian vault reader/writer adapter
  - Bidirectional sync with external vault changes
  - Vault configuration (path, folder, file naming) via config.yaml
  - Daily note integration for task read/write
- **Stories:** 8.1-8.4 (4 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR27, FR28, FR29, FR30
- **Risk:** Obsidian file format edge cases; daily note plugin variations

**Epic 9: Testing Strategy & Quality Gates**
- **Goal:** Establish comprehensive testing infrastructure with integration, contract, performance, and E2E tests
- **Prerequisites:** Epic 2 ✅, Epic 7, Epic 3.5 (story 3.5.7)
- **Deliverables:**
  - Apple Notes integration E2E tests
  - Contract tests for adapter compliance
  - Performance benchmarks (<100ms NFR validation)
  - Functional E2E tests for full user workflows
  - CI coverage gates preventing regression
- **Stories:** 9.1-9.5 (5 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR49, FR50, FR51
- **Risk:** Test infrastructure overhead; keep pragmatic

**Epic 10: First-Run Onboarding Experience** 🔄 IN PROGRESS
- **Goal:** Provide a guided welcome flow for new users to set up values/goals, understand Three Doors, learn key bindings, and optionally import existing tasks
- **Prerequisites:** Epic 3 ✅
- **Status:** IN PROGRESS — Stories 10.1 and 10.2 both in progress
- **Deliverables:**
  - Welcome flow with Three Doors concept explanation
  - Values/goals setup wizard
  - Key bindings walkthrough
  - Import from existing task sources
- **Stories:** 10.1-10.2 (2 stories)
- **Estimated Effort:** 1-2 weeks at 2-4 hrs/week
- **FRs covered:** FR38, FR39
- **Risk:** Over-designing onboarding for a CLI tool; keep lightweight

**Epic 11: Sync Observability & Offline-First**
- **Goal:** Ensure robust offline-first operation with local queue, sync status visibility, conflict visualization, and sync debugging
- **Prerequisites:** Epic 2 ✅, Epic 3.5 (story 3.5.4)
- **Deliverables:**
  - Offline-first local change queue with replay
  - Sync status indicator in TUI per provider
  - Conflict visualization and resolution UI
  - Sync log for debugging
- **Stories:** 11.1-11.3 (3 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR40, FR41, FR42, FR43
- **Risk:** Conflict resolution complexity; start with last-write-wins, iterate

**Epic 12: Calendar Awareness (Local-First, No OAuth)**
- **Goal:** Add time-contextual door selection by reading local calendar sources. No OAuth, no cloud APIs.
- **Prerequisites:** Epic 4 (intelligent door selection to integrate with)
- **Deliverables:**
  - macOS Calendar.app reader via AppleScript
  - .ics file parser
  - CalDAV cache reader
  - Time-contextual door selection based on available time blocks
- **Stories:** 12.1-12.2 (2 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR44, FR45
- **Risk:** AppleScript reliability; calendar format edge cases

**Epic 13: Multi-Source Task Aggregation View**
- **Goal:** Unified cross-provider task pool with dedup detection and source attribution in the TUI
- **Prerequisites:** Epic 7 (multiple providers configured), Epic 8 or additional adapters
- **Deliverables:**
  - Cross-provider task pool aggregation
  - Duplicate detection across providers
  - Source attribution display in TUI
- **Stories:** 13.1-13.2 (2 stories)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR46, FR47, FR48
- **Risk:** Dedup heuristics may produce false positives; provide manual override

---

## Phase 4: Future Expansion (12+ months out)

**Epic 14: LLM Task Decomposition & Agent Action Queue**
- **Goal:** Enable LLM-powered task breakdown where selected tasks are decomposed into stories/specs output to git repos for coding agent pickup
- **Prerequisites:** Epic 3+ ✅
- **Deliverables:**
  - Spike: prompt engineering, output schema, git automation, agent handoff
  - LLM-generated BMAD-style stories/specs
  - Git repo structure output for Claude Code / multiclaude pickup
  - Configurable LLM backend (local vs cloud)
- **Stories:** 14.1-14.2 (2 stories)
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week (spike-driven)
- **FRs covered:** FR35, FR36, FR37
- **Risk:** High uncertainty; spike-first approach essential. Output quality depends on prompt engineering.

**Epic 15: Psychology Research & Validation**
- **Goal:** Build evidence base for ThreeDoors design decisions through literature review and validation studies
- **Prerequisites:** None (can run in parallel with development)
- **Deliverables:**
  - Literature review: choice architecture (why 3 doors?)
  - Mood-task correlation validation study
  - Procrastination intervention research summary
  - Evidence for "progress over perfection" as motivational framework
  - Findings feed into Epic 4 learning algorithm refinement
- **Stories:** 15.1-15.2 (2 stories)
- **Estimated Effort:** Ongoing research track (2-4 hrs/week)
- **FRs covered:** FR34
- **Risk:** Academic research may not yield actionable insights; focus on practical findings

**Epic 16: iPhone Mobile App (SwiftUI)** 🆕
- **Goal:** Bring the Three Doors experience to iPhone with a native SwiftUI app that shares tasks via Apple Notes and syncs seamlessly with the desktop TUI
- **Prerequisites:** Epic 2 ✅ (Apple Notes integration), Epic 3.5 (platform readiness for shared specs)
- **Deliverables:**
  - Native SwiftUI iPhone app with swipeable Three Doors card carousel
  - Apple Notes integration via Swift (reading tasks from same note as TUI)
  - Task completion and status changes from mobile
  - Session metrics collection compatible with desktop JSONL format
  - iCloud Drive sync for config and metrics
  - TestFlight distribution (App Store submission in Phase 2)
- **Stories:** 16.1 (SwiftUI Project Setup & CI), 16.2 (Task Provider Protocol & Apple Notes Reader), 16.3 (Three Doors Card Carousel), 16.4 (Door Detail & Status Actions), 16.5 (Session Metrics & iCloud Sync), 16.6 (Swipe Gestures & Pull-to-Refresh), 16.7 (Polish & TestFlight Distribution)
- **Estimated Effort:** 6-8 weeks at 4-6 hrs/week
- **Tech Stack:** Swift 5.9+, SwiftUI, CloudKit/iCloud Drive, Xcode 16+, iOS 17+ target
- **Risk:** Apple Notes API access from Swift may differ from osascript approach; App Store review timeline uncertainty
- **Origin:** Party mode mobile app discussion (2026-03-02)
- **Research:** See `docs/research/mobile-app-research.md` for full analysis

**Epic 18: Docker E2E & Headless TUI Testing Infrastructure** 🆕
- **Goal:** Establish reproducible, automated E2E testing using Docker containers and Bubbletea's `teatest` package for headless TUI interaction testing — replacing manual testing as the sole E2E validation method
- **Prerequisites:** Epic 3 ✅, Epic 9 (Stories 9.4, 9.5)
- **Deliverables:**
  - Headless TUI test harness using `teatest` (pseudo-TTY, programmatic key input, model assertions)
  - Golden file snapshot tests for TUI visual regression detection
  - Input sequence replay tests for complete user workflow validation
  - Docker-based reproducible test environment (`Dockerfile.test` + `docker-compose.test.yml`)
  - CI integration running Docker E2E tests on every PR
- **Stories:** 18.1 (Headless Harness), 18.2 (Golden Files), 18.3 (Workflow Replay), 18.4 (Docker Environment), 18.5 (CI Integration)
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **FRs covered:** FR52, FR53, FR54
- **Risk:** teatest is experimental; API may change. Docker adds CI time; mitigate with layer caching.
- **Origin:** Party mode testing infrastructure discussion (2026-03-02)

**Epic 19+: Additional Integrations** (Jira, Linear, Google Calendar, Slack, etc.)
**Epic 20+: Cross-Computer Sync** (Implement alternative to monolithic SQLite on cloud storage)
**Epic 21+: Advanced Features** (Voice interface, web interface, Apple Watch, iPad, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---

## Story Count Summary

| Epic | Stories | Status |
|------|---------|--------|
| Epic 1: Technical Demo | 7 | ✅ Complete |
| Epic 2: Apple Notes Integration | 6 | ✅ Complete |
| Epic 3: Enhanced Interaction | 7 | ✅ Complete |
| Epic 3.5: Platform Readiness (Bridging) | 8 | 🆕 Not Started |
| Epic 4: Learning & Door Selection | 6 | Not Started |
| Epic 5: macOS Distribution | 1 | ✅ Complete |
| Epic 6: Data Layer (Optional) | 2 | Not Started |
| Epic 7: Plugin/Adapter SDK | 3 | Not Started |
| Epic 8: Obsidian Integration | 4 | Not Started |
| Epic 9: Testing Strategy | 5 | Not Started |
| Epic 10: Onboarding | 2 | 🔄 In Progress |
| Epic 11: Sync Observability | 3 | Not Started |
| Epic 12: Calendar Awareness | 2 | Not Started |
| Epic 13: Multi-Source Aggregation | 2 | Not Started |
| Epic 14: LLM Decomposition | 2 | Not Started |
| Epic 15: Psychology Research | 2 | Not Started |
| Epic 16: iPhone Mobile App | 7 | 🆕 Not Started |
| Epic 18: Docker E2E & Headless TUI Testing | 5 | 🆕 Not Started |
| **Total** | **74** | **21 complete, 53 remaining** |

---
