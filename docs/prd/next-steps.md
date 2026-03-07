# Next Steps

## Current Focus: Complete In-Progress Epics

**Objective:** Finish partially complete work before starting new epics.

**In Progress:**
- **Epic 17: Door Theme System** -- Stories 17.1-17.6 being implemented. Adds user-selectable themed door frames with ASCII/ANSI art.

**Partially Complete (need remaining stories):**
- **Epic 9: Testing Strategy & Quality Gates** -- 2/5 stories done (PRs #83, #89). Stories 9.3-9.5 pending: performance benchmarks, functional E2E tests, CI coverage gates.
- **Epic 13: Multi-Source Task Aggregation** -- 1/2 stories done (PR #84). Story 13.2 (duplicate detection) pending.

---

## Next Epics in Priority Order

**After current in-progress work:**

1. **Epic 19: Jira Integration** -- Read-only adapter with JQL search, then bidirectional sync. Prerequisites met (Epic 7, 11, 13).
2. **Epic 20: Apple Reminders Integration** -- Full CRUD via JXA scripts. Prerequisites met (Epic 7).
3. **Epic 21: Sync Protocol Hardening** -- Per-provider sync scheduler, circuit breaker, canonical ID mapping. Prerequisites met (Epic 11, 13).

**Longer-term (prerequisites may not yet be met):**

4. **Epic 16: iPhone Mobile App** -- SwiftUI app with Apple Notes sync, Three Doors card carousel, TestFlight distribution.
5. **Epic 22: Self-Driving Development Pipeline** -- Dispatch multiclaude workers from TUI. Depends on Epic 14 (complete).

---

## Decision Points

**Epic 9 scope review:** Assess whether remaining stories (9.3-9.5) are still needed given Epic 18 (Docker E2E testing) is complete. May overlap.

**Epic 16 timing:** iPhone app depends on stable Apple Notes integration. Evaluate whether current integration is mature enough before starting.

**Epic 22 feasibility:** Self-driving pipeline depends on multiclaude stability. Evaluate CLI availability before committing resources.

---

## Completed Milestones

- Phase 1 (Technical Demo): COMPLETE -- Concept validated through daily use
- Phase 2 (Post-Validation): COMPLETE -- Apple Notes, enhanced interaction, platform readiness, learning, macOS distribution, data layer all shipped
- Phase 3 (Platform Expansion): MOSTLY COMPLETE -- Plugin SDK, Obsidian, onboarding, sync, calendar, LLM decomposition, psychology research, Docker E2E all shipped

---

*This PRD embodies "progress over perfection" -- comprehensive enough to guide development, flexible enough to adapt based on learnings, and structured to prevent premature investment in unvalidated concepts.*

---
