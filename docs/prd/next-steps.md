---
title: Next Steps
section: Roadmap & Priorities
lastUpdated: '2026-03-15'
---

# Next Steps

## Current Focus

**Objective:** All primary epics (1-67, 69-70) are complete. Focus is on remaining planned epics and future-phase work.

**Ready to Start:**
- **Epic 68: BOARD.md Redesign** (P2) -- Split BOARD.md into focused active dashboard and complete decision archive, extract Epic Number Registry to its own file, fix duplicate IDs, update agent definitions.

---

## Next Epics in Priority Order

**Not-started epics from the current backlog:**

1. **Epic 68: BOARD.md Redesign** (P2) -- Decision management infrastructure improvement. No prerequisites.
2. **Epic 33: Seasonal Door Theme Variants** (P2) -- Time-based seasonal auto-switching themes. Prerequisites met (Epic 17 complete).
3. **Epic 16: iPhone Mobile App** (P3) -- SwiftUI app with Apple Notes sync, Three Doors card carousel, TestFlight distribution. Prerequisites met (Epic 2, 3.5 complete).
4. **Epic 22: Self-Driving Development Pipeline** (P3) -- Dispatch multiclaude workers from TUI. Prerequisites met (Epic 14 complete).

**Future-phase epics (prerequisites may need validation):**

5. **Epic 4: Mood-Aware Adaptive Door Selection** -- Learning algorithm for context-aware task presentation.
6. **Epic 8: Web-Based Configuration Dashboard** -- Browser-based settings and analytics.
7. **Epic 15: Psychology Research Integration** -- Evidence-based task selection refinement.

---

## Decision Points

**Epic 68 scope:** BOARD.md is 400+ lines. Determine whether to split in a single story or phased approach.

**Phase 4+ prioritization:** With 67+ epics complete, evaluate which future-phase epics deliver the most user value. Mobile (Epic 16), mood-awareness (Epic 4), and web dashboard (Epic 8) are all candidates.

**Integration maintenance:** 8 data source adapters are live. Monitor for API changes, deprecations, or new provider requests from users.

---

## Completed Milestones

- **Phase 1** (Technical Demo): COMPLETE -- Concept validated through daily use
- **Phase 2** (Post-Validation): COMPLETE -- Apple Notes, enhanced interaction, platform readiness, learning, macOS distribution, data layer all shipped
- **Phase 3** (Platform Expansion): COMPLETE -- Plugin SDK, Obsidian, onboarding, sync, calendar, LLM decomposition, psychology research, Docker E2E all shipped
- **Phase 4** (Data Source Ecosystem): COMPLETE -- Connection manager, sources TUI/CLI, OAuth, Todoist, Linear, GitHub, Jira, ClickUp integrations all shipped
- **Phase 5** (Polish & Infrastructure): COMPLETE -- Door visual redesign, LLM CLI services, full-terminal layout, CI optimization, README overhaul, GitHub Pages, security hardening, doctor command, bug reporting, SLAES, retrospector, cross-computer sync all shipped
- **Phase 6** (TUI Decomposition & History): COMPLETE -- MainModel decomposition (Epic 69), completion history & progress view (Epic 70)

**Total: 67+ epics complete, 764+ PRs merged, 353 stories across 70 epics.**

---

*This PRD embodies "progress over perfection" -- comprehensive enough to guide development, flexible enough to adapt based on learnings, and structured to prevent premature investment in unvalidated concepts.*

---
