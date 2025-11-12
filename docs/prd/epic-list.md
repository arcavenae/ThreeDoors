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

## Phase 3: Future Expansion (12+ months out)

**Epic 6+: Additional Integrations** (Jira, Linear, Google Calendar, Slack, etc.)
**Epic 7+: Cross-Computer Sync** (Implement alternative to monolithic SQLite on cloud storage)
**Epic 8+: LLM Integration** (Task breakdown assistance, assumption challenging, dependency collapse)
**Epic 9+: Advanced Features** (Voice interface, mobile app, web interface, trading mechanic, gamification)

**Guiding Principle:** Each epic must deliver tangible user value and be informed by real usage patterns from previous phases. No speculation-driven development.

---
