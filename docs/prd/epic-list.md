# Epic List

## Phase 1: Technical Demo & Validation (Immediate - Week 1)

**Epic 1: Three Doors Technical Demo**
- **Goal:** Build and validate the Three Doors interface with minimal viable functionality to prove the UX concept reduces friction compared to traditional task lists
- **Timeline:** 1 week (4-8 hours development time - optimized sequence)
- **Deliverables:** Working CLI/TUI showing Three Doors, reading from text file, door refresh, task selection with expanded detail view, search/command palette, mood tracking, marking tasks complete, comprehensive session metrics
- **Stories:** 1.1 (Project Setup), 1.2 (Display Three Doors), 1.3 (Door Selection & Status Management), 1.3a (Quick Search & Command Palette), 1.5 (Session Metrics Tracking), 1.6 (Essential Polish)
- **Success Criteria:**
  - Developer uses tool daily for 1 week
  - Three Doors selection feels meaningfully different from scrolling a list
  - Session metrics provide objective data for validation decision
  - Decision point reached: proceed to Full MVP or pivot/abandon
- **Tech Stack:** Go 1.25.4+, Bubbletea/Lipgloss, local text files, JSONL metrics
- **Risk:** UX concept might not feel better than simple list; easy to pivot if fails
- **Optimization:** Reordered stories to validate refresh UX before completion; merged/simplified non-essential features; added search/command palette and mood tracking for richer validation data

---

## Phase 2: Post-Validation Roadmap (Conditional on Phase 1 Success)

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
- **Goal:** Use historical session metrics (captured in Epic 1 Story 1.5) to analyze user patterns and adapt door selection to improve task engagement and completion rates
- **Prerequisites:** Epic 3 complete (enough usage data to learn from)
- **Data Foundation:** Epic 1 Story 1.5 captures door position selections, task bypasses, status changes, and mood/emotional context—essential for pattern analysis
- **Deliverables:**
  - Pattern recognition (which task types are selected vs bypassed)
  - Mood correlation analysis (emotional states → task selection/avoidance patterns)
  - Avoidance detection (tasks repeatedly shown but never selected)
  - Status pattern analysis (task types that get blocked/procrastinated, correlated with mood)
  - Adaptive selection based on current mood state and historical patterns
  - User insights ("When stressed, you avoid complex tasks")
  - Goal re-evaluation prompts when persistent avoidance + mood patterns detected
  - Encouragement system with mood-aware messaging
  - Task categorization (type, effort level, context)
  - "Better than yesterday" multi-dimensional tracking
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week
- **Risk:** Algorithm complexity; may need to simplify learning approach

**Epic 5: macOS Distribution & Packaging**
- **Goal:** Provide a trusted, seamless installation experience on macOS by signing, notarizing, and packaging the binary so Gatekeeper does not quarantine it
- **Prerequisites:** None (independent of feature epics; can be implemented at any time)
- **Deliverables:**
  - Code signing with Apple Developer certificate in CI
  - Notarization with Apple's notarization service
  - Homebrew tap formula (`brew install arcaven/tap/threedoors`)
  - DMG or pkg installer as alternative to Homebrew
  - Automated release pipeline for signed/notarized binaries
- **Estimated Effort:** 1-2 weeks at 2-4 hrs/week
- **Risk:** Requires active Apple Developer Program membership ($99/year); certificate management adds CI complexity
- **Independence:** This epic is independent of the story pipeline and can be merged at any time

**Epic 6: Data Layer & Enrichment (Optional)**
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

## Phase 3: Platform Expansion & Intelligence (Post-MVP)

**Epic 7: Plugin/Adapter SDK & Registry**
- **Goal:** Formalize the adapter pattern into a plugin SDK with registry, config-driven provider selection, and developer guide. Unblocks all future integrations.
- **Prerequisites:** Epic 2 (adapter pattern established)
- **Deliverables:**
  - Adapter registry with runtime discovery and loading
  - Config-driven provider selection via `~/.threedoors/config.yaml`
  - Adapter developer guide and interface specification
  - Contract test suite for adapter compliance validation
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Over-engineering the plugin system; keep minimal until 3+ adapters exist

**Epic 8: Obsidian Integration (P0 - #2 Integration)**
- **Goal:** Add Obsidian vault as second task storage backend after Apple Notes. Local-first Markdown integration with bidirectional sync.
- **Prerequisites:** Epic 7 (adapter SDK), Epic 2 (adapter pattern)
- **Deliverables:**
  - Obsidian vault reader/writer adapter
  - Bidirectional sync with external vault changes
  - Vault configuration (path, folder, file naming) via config.yaml
  - Daily note integration for task read/write
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Obsidian file format edge cases; daily note plugin variations

**Epic 9: Testing Strategy & Quality Gates**
- **Goal:** Establish comprehensive testing infrastructure with integration, contract, performance, and E2E tests
- **Prerequisites:** Epic 2 (adapters to test), Epic 7 (contract test framework)
- **Deliverables:**
  - Apple Notes integration E2E tests
  - Contract tests for adapter compliance
  - Performance benchmarks (<100ms NFR validation)
  - Functional E2E tests for full user workflows
  - CI coverage gates preventing regression
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Test infrastructure overhead; keep pragmatic

**Epic 10: First-Run Onboarding Experience**
- **Goal:** Provide a guided welcome flow for new users to set up values/goals, understand Three Doors, learn key bindings, and optionally import existing tasks
- **Prerequisites:** Epic 3 (values/goals features exist to configure)
- **Deliverables:**
  - Welcome flow with Three Doors concept explanation
  - Values/goals setup wizard
  - Key bindings walkthrough
  - Import from existing task sources
- **Estimated Effort:** 1-2 weeks at 2-4 hrs/week
- **Risk:** Over-designing onboarding for a CLI tool; keep lightweight

**Epic 11: Sync Observability & Offline-First**
- **Goal:** Ensure robust offline-first operation with local queue, sync status visibility, conflict visualization, and sync debugging
- **Prerequisites:** Epic 2 (sync infrastructure exists)
- **Deliverables:**
  - Offline-first local change queue with replay
  - Sync status indicator in TUI per provider
  - Conflict visualization and resolution UI
  - Sync log for debugging
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Conflict resolution complexity; start with last-write-wins, iterate

**Epic 12: Calendar Awareness (Local-First, No OAuth)**
- **Goal:** Add time-contextual door selection by reading local calendar sources. No OAuth, no cloud APIs.
- **Prerequisites:** Epic 4 (intelligent door selection to integrate with)
- **Deliverables:**
  - macOS Calendar.app reader via AppleScript
  - .ics file parser
  - CalDAV cache reader
  - Time-contextual door selection based on available time blocks
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** AppleScript reliability; calendar format edge cases

**Epic 13: Multi-Source Task Aggregation View**
- **Goal:** Unified cross-provider task pool with dedup detection and source attribution in the TUI
- **Prerequisites:** Epic 7 (multiple providers configured), Epic 8 or additional adapters
- **Deliverables:**
  - Cross-provider task pool aggregation
  - Duplicate detection across providers
  - Source attribution display in TUI
- **Estimated Effort:** 2-3 weeks at 2-4 hrs/week
- **Risk:** Dedup heuristics may produce false positives; provide manual override

---

## Phase 4: Future Expansion (12+ months out)

**Epic 14: LLM Task Decomposition & Agent Action Queue**
- **Goal:** Enable LLM-powered task breakdown where selected tasks are decomposed into stories/specs output to git repos for coding agent pickup
- **Prerequisites:** Epic 3+ (task management mature enough to decompose from)
- **Deliverables:**
  - Spike: prompt engineering, output schema, git automation, agent handoff
  - LLM-generated BMAD-style stories/specs
  - Git repo structure output for Claude Code / multiclaude pickup
  - Configurable LLM backend (local vs cloud)
- **Estimated Effort:** 3-4 weeks at 2-4 hrs/week (spike-driven)
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
- **Estimated Effort:** Ongoing research track (2-4 hrs/week)
- **Risk:** Academic research may not yield actionable insights; focus on practical findings

**Epic 16+: Additional Integrations** (Jira, Linear, Google Calendar, Slack, etc.)
**Epic 17+: Cross-Computer Sync** (Implement alternative to monolithic SQLite on cloud storage)
**Epic 18+: Advanced Features** (Voice interface, mobile app, web interface, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---
