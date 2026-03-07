# Goals and Background Context

## Goals

**Technical Demo & Validation Phase (Pre-MVP):**
- Validate the Three Doors UX concept in 1 week (4-8 hours of development)
- Prove the core hypothesis: "Presenting three diverse tasks is better than presenting a list"
- Build working TUI with Bubbletea to demonstrate feasibility
- Use simple local text file for rapid task population and testing
- Gather real usage feedback before investing in complex integrations

**Full MVP Goals (Post-Validation):**
- Master BMAD methodology through authentic, real-world application
- Create a todo app that reduces friction and actually helps with organization
- Build a personal achievement partner that works with human psychology, not against it
- Enable seamless cross-context navigation across multiple platforms and tools
- Capture the full story (what AND why) to improve stakeholder communication
- Achieve measurably better personal organization than current scattered approach
- Demonstrate progress-over-perfection philosophy in both product design and development process

## Background Context

Traditional todo apps work well for already-organized people, but they're fundamentally rudimentary tools that haven't evolved alongside modern technology capabilities. While they help those who are naturally organized stay organized, they offer little support for adapting to the dynamic reality of modern life—where the same person occupies multiple roles (employee, parent, partner, learner), experiences varying moods and energy states, and faces constantly shifting priorities.

ThreeDoors recognizes that as technology has advanced, we can offer substantially more support. We can organize our organization tools themselves, bringing together tasks scattered across multiple systems. More importantly, we can adapt technology support dynamically: responding to the user's current context, role, mood, and circumstances, re-routing based on changing conditions and priorities. This PRD defines the full product vision: a CLI/TUI application that integrates with Apple Notes, Obsidian, and other task sources, uses mood-aware adaptive door selection, and embodies a "progress over perfection" philosophy while serving as a practical demonstration of the BMAD methodology. The Technical Demo (Phase 1) has been validated through daily use, confirming the Three Doors concept reduces friction compared to traditional task lists.

## Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| 2025-11-07 | 1.0 | Initial PRD creation from project brief | John (PM Agent) |
| 2025-11-07 | 1.1 | Pivoted to Technical Demo & Validation approach (Option C): Simplified to text file storage, 1-week validation timeline, deferred Apple Notes and learning features to post-validation phases | John (PM Agent) |
| 2025-11-08 | 1.3 | Incorporated user feedback: application name change to "ThreeDoors", enhanced UX with new key bindings (arrow keys), dynamic door sizing, no initial selection, removal of "Door X" labels, and introduction of new task management key bindings (c, b, i, e, f, p) for future implementation. | Bob (SM Agent) |
| 2026-03-02 | 1.4 | Course correction: Added macOS Distribution & Packaging requirements (FR22-FR26), Epic 5 with Stories 5.1-5.3 for code signing, notarization, Homebrew tap, and pkg installer. Addresses Gatekeeper quarantine friction for unsigned binaries. Updated NFR3 to require signed+notarized binaries. Renumbered Data Layer epic from 5→6 and Phase 3 epics accordingly. | PM Agent |
| 2026-03-02 | 1.5 | PRD validation: Added missing BMAD core sections (Executive Summary, Product Scope, User Journeys). Fixed measurability issues in TD8, FR16, success criteria. Removed implementation leakage from FR11, FR26, FR44. Updated index to include new sections. | Validation Agent |
| 2026-03-03 | 1.6 | PRD validation follow-up: Removed "shall allow users to" filler from TD4, FR3, FR5, FR18, FR57. Added measurable verification methods to NFR3, NFR7, NFR9, NFR11. Added Journey 9 (Door Theme Customization) for FR55-FR62. Updated Product Scope with Epic 3.5, Docker E2E, iPhone app. Updated Epic 17 to IN PROGRESS, Epic 18 to COMPLETE. | Worker Agent |
| 2026-03-06 | 1.7 | PRD re-validation: Synchronized epic-list.md status with epics-and-stories.md (101 merged PRs). Added Success Criteria section. Added missing FR52-FR54 (Docker E2E testing). Fixed duplicate Phase 4 heading in product-scope.md. Aligned phase numbering across requirements.md. Updated stale next-steps.md and background context. Fixed deprecated io/ioutil reference. Updated story count to 119. | PM Agent |

---
