# PRD Coverage Gap Analysis

**Date:** 2026-03-13
**Analyst:** clever-tiger (worker)
**Method:** Cross-referenced all PRD component files (requirements.md, product-scope.md, executive-summary.md, user-journeys.md, user-interface-design-goals.md, epic-details.md) against epics-and-stories.md, epic-list.md, ROADMAP.md, and docs/stories/*.story.md

## Summary

The PRD is **well-covered** by existing epics. Of ~150+ functional requirements (TD1-TD9, FR2-FR151, NFR series), all but a small number have corresponding epics with stories. The project has 62+ epics, 680+ merged PRs, and comprehensive story files.

**3 actionable gaps found. 2 PRD inconsistencies noted.**

---

## Gap 1: ClickUp Integration — No Epic

**PRD Reference:** product-scope.md Phase 5 ("Additional integrations (GitHub Issues, ClickUp)")
**Phase 4 out-of-scope note:** "Linear, GitHub Issues, ClickUp integrations (deferred to Phase 5+)"

**Current State:** GitHub Issues integration is Epic 26 (COMPLETE). Linear integration is Epic 30 (0/4, Not Started). No epic exists for ClickUp.

**Assessment:** Low priority. Phase 5 is "12+ months out." ClickUp would follow the same adapter pattern as Jira (Epic 19), Todoist (Epic 25), GitHub Issues (Epic 26), and Linear (Epic 30). The Connection Manager Infrastructure (Epic 43) and Sources TUI/CLI (Epics 44-45) provide the framework.

**Recommendation:** Create epic with 4 stories following the established integration pattern:
1. ClickUp REST API Client & Auth Configuration
2. Read-Only ClickUp Provider with Field Mapping
3. Bidirectional Sync & WAL Integration
4. Contract Tests & Integration Testing

**Priority:** P2 — Not urgent. Can be planned now and executed when Phase 5 begins.

---

## Gap 2: Cross-Computer Sync — No Epic

**PRD Reference:** product-scope.md Phase 5 ("Cross-computer sync" listed as in-scope)
**Deferred in:** Phase 1 (out of scope), Phase 2 (out of scope), Phase 3 (not mentioned), Phase 4 (out of scope)
**Technical context:** technical-assumptions.md states "Cross-computer sync is deferred post-MVP; single-computer local storage is sufficient"

**Current State:** No epic, no stories, no architecture research. Epic 53 (Remote Collaboration) covers SSH access to multiclaude agents, NOT task data sync across machines.

**Assessment:** Significant architecture effort. Would need:
- Sync protocol design (CRDTs? Last-writer-wins across machines?)
- Transport mechanism (cloud storage, peer-to-peer, git-based?)
- Conflict resolution when same task modified on two machines
- Identity/device management

**Recommendation:** Create epic with research spike + implementation stories. The existing sync infrastructure (WAL, SourceRef, circuit breaker from Epics 21, 43, 47) provides a foundation but cross-machine sync is architecturally distinct from cross-provider sync.

**Priority:** P2 — Phase 5 scope. Needs architecture research before stories can be fully specified.

---

## Gap 3: FR25 — DMG/pkg Installer (Story 5.3 never created)

**PRD Reference:** requirements.md FR25 ("The system shall provide a DMG or pkg installer as an alternative installation method")
**Epic Details Reference:** epic-details.md defines Story 5.3 (DMG/pkg Installer) with acceptance criteria
**Change Log:** Version 1.4 explicitly added "FR22-FR26, Epic 5 with Stories 5.1-5.3"

**Current State:** Epic 5 was simplified to 1 story (5.1 — Homebrew). Story 5.3 was defined in epic-details.md but never created as a story file and never implemented. Epic 38 (Dual Homebrew Distribution) covers cask + formula but NOT pkg/DMG.

**Assessment:** The Homebrew distribution (stable + alpha) covers the primary use case. DMG/pkg is an alternative for users who prefer graphical installation. With Homebrew being the standard Go CLI distribution method, this is lower priority than originally planned.

**Recommendation:** Create story file for 5.3 or re-scope as a new story under Epic 38 (or a new Infrastructure story 0.xx). Single story:
- CI generates signed .pkg installer via pkgbuild/productbuild
- .pkg is notarized with Apple
- .pkg published as GitHub Release asset alongside binaries

**Priority:** P2 — Nice to have. Homebrew covers most users.

---

## PRD Inconsistency 1: FR9 — Session Improvement Prompt (Deliberately Reversed)

**PRD Reference:** requirements.md FR9 ("The system shall prompt the user once per session with: 'What's one thing you could improve...'")

**Current State:** This was implemented as Story 3.6 (Session Reflection / ImprovementView) but was **deliberately removed** by Story 0.50 (PR #556, P0) because it violated SOUL.md — the quit intercept frustrated users who expected immediate exit. Decision D-165 documents the reversal.

**Recommendation:** FR9 should be removed from the PRD requirements or marked as "Reversed (D-165)" to reflect the current product direction. This is NOT a gap to fill — it's a PRD-code divergence where the code is correct.

---

## PRD Inconsistency 2: FR7 — Choose-Your-Own-Adventure Navigation

**PRD Reference:** requirements.md FR7 ("The system shall provide a 'choose-your-own-adventure' interactive navigation flow that presents options rather than demands")

**Current State:** This is more of a design principle than a testable requirement. The Three Doors interface itself embodies this principle (presenting options, not demands). The command palette, door selection, and detail view all offer choices. However, there's no explicit "adventure-style" navigation flow implemented as a discrete feature.

**Recommendation:** Either remove FR7 as a standalone requirement (it's a UX philosophy already embedded in the product) or clarify it into testable acceptance criteria if the intent was something more specific.

---

## Coverage Matrix — All FRs Accounted For

| FR Range | Feature | Epic | Status |
|----------|---------|------|--------|
| TD1-TD9 | Technical Demo | Epic 1 | COMPLETE |
| FR2-FR15 | Apple Notes, Health Check | Epics 2, 3, 4 | COMPLETE |
| FR16-FR21 | Quick Add, Learning | Epics 3, 4 | COMPLETE |
| FR22-FR24, FR26 | Signing, Homebrew, Release | Epics 5, 38 | COMPLETE |
| **FR25** | **DMG/pkg Installer** | **None** | **GAP** |
| FR27-FR30 | Obsidian | Epic 8 | COMPLETE |
| FR31-FR33 | Plugin SDK | Epic 7 | COMPLETE |
| FR34 | Psychology Research | Epic 15 | COMPLETE |
| FR35-FR37 | LLM Decomposition | Epic 14 | COMPLETE |
| FR38-FR39 | Onboarding | Epic 10 | COMPLETE |
| FR40-FR43 | Sync Observability | Epic 11 | COMPLETE |
| FR44-FR45 | Calendar | Epic 12 | COMPLETE |
| FR46-FR48 | Multi-Source | Epic 13 | COMPLETE |
| FR49-FR51 | Testing | Epic 9 | COMPLETE |
| FR52-FR54 | Docker E2E | Epic 18 | COMPLETE |
| FR55-FR62 | Door Themes | Epic 17 | COMPLETE |
| FR63-FR66 | Jira | Epic 19 | COMPLETE |
| FR67-FR69 | Apple Reminders | Epic 20 | COMPLETE |
| FR70-FR72 | Sync Hardening | Epic 21 | COMPLETE |
| FR73-FR80 | Self-Driving Dev | Epic 22 | COMPLETE |
| FR89-FR92 | Todoist | Epic 25 | COMPLETE |
| FR93-FR96 | GitHub Issues | Epic 26 | COMPLETE |
| FR97-FR103 | Daily Planning | Epic 27 | COMPLETE |
| FR104-FR109 | Snooze/Defer | Epic 28 | COMPLETE |
| FR110-FR115 | Dependencies | Epic 29 | 3/4 done |
| FR116-FR119 | Linear | Epic 30 | 0/4 Not Started |
| FR120-FR126 | Expand/Fork | Epic 31 | 0/5 Not Started |
| FR127-FR131 | Undo Completion | Epic 32 | COMPLETE |
| FR132-FR137 | Seasonal Themes | Epic 33 | COMPLETE |
| FR138-FR147 | Door Proportions | Epic 35 | COMPLETE |
| FR148-FR151 | Selection Feedback | Epic 36 | COMPLETE |
| NFR-DX1-DX6 | Dev Experience | Epic 34 | COMPLETE |

**PRD Features with no FR but in product-scope:**
| Feature | Epic | Status |
|---------|------|--------|
| ClickUp Integration | **None** | **GAP** |
| Cross-computer Sync | **None** | **GAP** |
| iPhone Mobile App | Epic 16 | ICEBOX |

---

## Next Steps

1. **Supervisor action needed:** Request epic numbers from project-watchdog for:
   - ClickUp Integration (new epic, ~4 stories)
   - Cross-computer Sync (new epic, ~5-6 stories, needs research spike)
   - FR25 DMG/pkg Installer (either new story 5.3 or new infra story 0.xx)

2. **PRD maintenance:** Update requirements.md to mark FR9 as "Reversed (D-165)" and clarify FR7.

3. **PM planning:** Once epic numbers assigned, create story files and update planning docs.
