# PR-Story Gap Analysis Report

**Analysis Date:** 2026-03-03
**Scope:** All 77 PRs (#1–#77) in arcavenae/ThreeDoors
**Method:** `gh pr view` on each PR, cross-referenced against `docs/prd/epics-and-stories.md`

---

## Executive Summary

Of 77 PRs, **73 were merged** and **4 were closed without merge** (duplicates). Of the 73 merged PRs:

| Category | Count | % of Merged |
|----------|-------|-------------|
| Story-backed | 43 | 59% |
| Infrastructure (no story) | 9 | 12% |
| Docs/Research (no story) | 12 | 16% |
| Fix-up (no story) | 8 | 11% |
| Duplicate/Closed | 4 | — |

**Key finding:** 29 merged PRs (40%) shipped without a backing story. This represents significant untracked work — CI fixes, documentation updates, quality standards, tooling, and research that was done ad-hoc.

---

## Section 1: Infrastructure PRs with No Story

These PRs performed necessary work that had no story, ACs, or tracked scope.

| PR | Title | What It Did |
|----|-------|-------------|
| #1 | chore: Upgrade BMAD method framework | Added BMAD slash commands, agent definitions, project docs framework |
| #12 | feat: create GitHub Release with binaries on merge to main | Extended CI to create prerelease GitHub Releases with binaries |
| #32 | docs: add Pre-PR Submission Checklist to all story files | Researched 31 PRs for delay patterns, added checklists to 11 story files |
| #33 | docs: expand PR submission standards across all project documentation | Added NFR-CQ1 through NFR-CQ5, checklists to 27 stories, architecture doc |
| #36 | docs: PRD validation - add missing BMAD sections | 13-step BMAD validation; added executive-summary, user-journeys, product-scope |
| #37 | docs: sprint status audit, story validation, and status fixes | Audited 15 epics against 34 PRs, fixed 10 story files with stale metadata |
| #48 | feat: add /implement-story reusable workflow command | Created custom slash command codifying 8-phase implementation workflow |
| #50 | feat: Add comprehensive CLAUDE.md with Go quality rules | 10 idiomatic Go rules, error handling, testing standards, TUI rules |
| #52 | docs: add Quality Gate ACs to all unimplemented stories | Added AC-Q1–AC-Q8 quality gates to 41 unimplemented stories |

---

## Section 2: Fix-up PRs (Should Have Been Prevented by Story ACs)

These PRs fixed issues that escaped from previous PRs — each represents a missed AC or quality gate.

| PR | Title | Root Cause | Should Have Been Caught By |
|----|-------|-----------|---------------------------|
| #7 | test: Add comprehensive TUI test suite for Story 1.3 | 76 tests missing from PR #5 | Story 1.3 should have had AC: "≥70% test coverage for new code" |
| #9 | feat: add test coverage reporting to CI pipeline | Coverage reporting missing from CI | Story 1.7 should have had AC: "CI reports test coverage" |
| #10 | fix: apply gofumpt formatting to detail_view_test.go | Single formatting violation | Story 1.3 should have had AC: "make fmt produces no changes" |
| #61 | fix: align CI secret names and document signing setup | 3 secret name mismatches in CI | Story 5.1 should have had AC: "CI secrets match workflow variable names" |
| #67 | fix: bump notarization timeout to 30 minutes | Notarization timed out at 15 min | Story 5.1 should have had AC: "notarization timeout ≥30 min for new accounts" |
| #71 | docs: restore door emojis to README | Emojis stripped during PR #69 README update | README update story should have had AC: "existing formatting preserved" |
| #74 | test: Story 8.1 AC-Q6 input sanitization tests | Quality gate tests missing from Story 8.1 PR | Story 8.1 AC-Q6 existed but wasn't verified before merge |
| #76 | fix: increase notarization timeout to 1 hour | 30-min timeout still too short | Should have been research-informed from the start (Apple docs say up to 1 hr) |

**Pattern:** 3 of 8 fix-up PRs are CI/notarization related (PRs #61, #67, #76), suggesting Story 5.1 ACs were insufficient for CI configuration.

---

## Section 3: Docs/Research PRs with No Story

These PRs produced significant documentation and research artifacts without story-level tracking.

| PR | Title | Artifact Produced |
|----|-------|------------------|
| #6 | feat: Complete epics and stories breakdown for all phases | 26 FRs, 19 NFRs, stories for 5 epics |
| #11 | docs: add install and usage documentation to README | Installation, usage, keybinding docs |
| #26 | feat: Add macOS distribution & packaging to PRD (Epic 5) | Epic 5 with FR22–FR26 and Stories 5.1–5.3 |
| #34 | feat: integrate 9 party mode recommendations into PRD | FR27–FR51, NFR13–NFR16, Epics 7–15 |
| #35 | docs: AI tooling research — CLAUDE.md, SOUL.md, skills | Proposed CLAUDE.md, SOUL.md, 4 custom skills |
| #38 | docs: architecture v2.0 | 5-layer post-MVP architecture, 10 files (+1,312 lines) |
| #39 | docs: regenerate epics from PRD v2.0 + bridging Epic 3.5 | Epic 3.5 (8 stories), detailed Epic 4 (6 stories) |
| #46 | docs: code signing research findings | CI signing analysis and findings |
| #47 | feat: Add Epic 16 - iPhone Mobile App (SwiftUI) | Epic 16 with 7 stories, mobile research |
| #51 | docs: PR analysis-derived quality gates, NFRs | 8 quality ACs (AC-Q1–Q8), 15 NFRs |
| #60 | feat: Epic 18 - Docker E2E & Headless TUI Testing | Epic 18 with 5 stories, FR52–FR54 |
| #69 | docs: update README with all features since PR #11 | Comprehensive README overhaul |

---

## Section 4: Duplicate/Closed PRs

| PR | Title | Superseded By | Cause |
|----|-------|---------------|-------|
| #3 | feat: Story 1.3 - Door Selection & Task Status Management | PR #5 | Rebase required; new PR created instead of force-push |
| #14 | feat: Implement Story 1.3a - Quick Search & Command Palette | PR #13 | Alternative implementation; parallel work on same story |
| #49 | feat: Story 4.4 - Adaptive Door Selection Algorithm | PR #45 | Concurrent worker PRs for same story |
| #57 | feat: Story 6.2 - Cross-Reference Tracking | PR #56 | Concurrent worker PRs for same story |

**Pattern:** PRs #49 and #57 suggest multi-agent coordination gaps — two workers attempted the same story simultaneously.

---

## Section 5: Stale Story Statuses (Implemented but Marked Pending)

These stories are marked as "Pending" or have no status indicator in `epics-and-stories.md` but have been implemented via merged PRs. This is a documentation drift problem.

| Story | PR | Epic Status in Doc |
|-------|----|--------------------|
| Story 4.1: Task Categorization & Tagging | #40 | Epic 4: "NOT STARTED" |
| Story 4.2: Session Metrics Pattern Analysis | #43 | Epic 4: "NOT STARTED" |
| Story 4.3: Mood-Aware Adaptive Door Selection | #44 | Epic 4: "NOT STARTED" |
| Story 4.4: Avoidance Detection & User Insights | #45 | Epic 4: "NOT STARTED" |
| Story 4.5: User Insights Dashboard | #42 | Epic 4: "NOT STARTED" |
| Story 6.1: SQLite Enrichment Database Setup | #53 | Epic 6: "NOT STARTED" |
| Story 6.2: Cross-Reference Tracking | #56 | Epic 6: "NOT STARTED" |
| Story 7.1: Adapter Registry & Runtime Discovery | #68 | Epic 7: "NOT STARTED" |
| Story 7.2: Config-Driven Provider Selection | #70 | Epic 7: "NOT STARTED" |
| Story 7.3: Adapter Developer Guide & Contract Tests | #72 | Epic 7: "NOT STARTED" |
| Story 8.1: Obsidian Vault Reader/Writer Adapter | #73 | Epic 8: "PARTIALLY COMPLETE" |
| Story 8.2 & 8.3: Obsidian Sync & Config | #75 | Epic 8: "PARTIALLY COMPLETE" |
| Story 8.4: Obsidian Daily Note Integration | #77 | Epic 8: "PARTIALLY COMPLETE" |
| Story 10.1: Welcome Flow | #55 | Epic 10: "NOT STARTED" |
| Story 10.2: Values/Goals Setup & Task Import | #59 | Epic 10: "NOT STARTED" |
| Story 11.1: Offline-First Local Change Queue | #62 | Epic 11: "NOT STARTED" |
| Story 11.2: Sync Status Indicator | #66 | Epic 11: "NOT STARTED" |
| Story 12.1: Local Calendar Source Reader | #65 | Epic 12: "NOT STARTED" |
| Story 14.1: LLM Task Decomposition Spike | #63 | Epic 14: "NOT STARTED" |
| Story 15.1: Choice Architecture Literature Review | #54 | Epic 15: "NOT STARTED" |
| Story 15.2: Mood-Task Correlation Research | #58 | Epic 15: "NOT STARTED" |
| Story 18.1: Headless TUI Test Harness | #64 | Epic 18: "NOT STARTED" |

**22 stories** are implemented but not marked as done. Epics 4, 6, 7, 8, 10 are fully or mostly complete but listed as "NOT STARTED."

---

## Section 6: Phantom Completions Check

**Result: No phantom completions found.**

Every story marked as "Done" in the document has a corresponding merged PR:

- Epic 1 (7 stories): All verified against PRs #2, #4, #5/#7, #13, #16, #18, #8
- Epic 2 (6 stories): All verified against PRs #20, #22, #17, #21, #15, #19
- Epic 3 (7 stories): All verified against PRs #23, #24, #25, #27, #28, #29, #31
- Epic 5 (1 story): Verified against PR #30

---

## Section 7: Naming Mismatches

| Issue | Detail |
|-------|--------|
| Story 1.3a vs 1.4 | Epics doc uses "Story 1.3a" but PR #13 title says "Story 1.4" |
| Story 1.8 missing | PR #41 implements "Story 1.8 - CI Process Validation & Fixes" but no Story 1.8 exists in epics-and-stories.md |
| Story 4.5 mismatch | Epics doc: "Goal Re-evaluation Prompts" — PR #42: "User Insights Dashboard" — different scope |
| Story 5.1 collision | PR #30 (macOS distribution) and PR #53 (SQLite enrichment) both claim "Story 5.1" — different epics |

---

## Section 8: Truly Pending Stories (No PR)

These stories remain genuinely unimplemented:

| Story | Title |
|-------|-------|
| Stories 3.5.1–3.5.8 | All 8 Platform Readiness bridging stories |
| Story 4.6 | "Better Than Yesterday" Multi-Dimensional Tracking |
| Stories 9.1–9.5 | All 5 Testing Strategy stories |
| Story 11.3 | Conflict Visualization & Sync Log |
| Story 12.2 | Time-Contextual Door Selection |
| Story 14.2 | Agent Action Queue Integration |
| Stories 18.2–18.5 | Golden File Snapshots, Input Replay, Docker E2E, CI Integration |

**Total: 23 stories** remain genuinely pending across 5 epics.

---

## Section 9: Recommendations

1. **Add backfill stories** — Create Epic 0 to retroactively cover the 29 unstory'd PRs (see companion update to epics-and-stories.md)
2. **Update stale statuses** — Mark 22 implemented stories as Done with their PR numbers
3. **Resolve naming conflicts** — Standardize Story 1.3a/1.4, add Story 1.8, disambiguate Story 5.1 collision
4. **Strengthen CI/distribution ACs** — Story 5.1 spawned 3 fix-up PRs; future distribution stories need operational ACs
5. **Multi-agent coordination** — PRs #49 and #57 indicate workers need story assignment locks to prevent duplicate work
6. **Quality gate enforcement** — PR #7 (76 missing tests) and PR #74 (missing AC-Q6 tests) suggest ACs aren't verified pre-merge
