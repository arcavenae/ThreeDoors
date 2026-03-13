# PRD Process Bypass Investigation & Remediation Plan

**Date:** 2026-03-13
**Author:** gentle-owl (worker, research task)
**Severity:** High — systemic process failure affecting planning doc integrity

---

## Executive Summary

Since early March 2026, `/plan-work` and `/course-correct` workflows have been creating epics and stories (Epics 22-65+) without updating the core PRD documents (`requirements.md`, `product-scope.md`, `epic-details.md`). While the skill definitions technically *instruct* workers to update these docs, the mechanism is broken at multiple levels. The result: **47+ epics exist in `epic-list.md` with no corresponding PRD coverage**, and the PRD has effectively become a fossil from Phase 1-3 planning.

---

## 1. Forensic Audit

### PRD Update Timeline

| Document | Last Meaningful PRD Update | Current State |
|---|---|---|
| `requirements.md` | 2026-03-08 (door selection feedback, nil pointer fix) | Covers through FR131 + NFR-DX6. No requirements for Epics 36-65. |
| `product-scope.md` | 2026-03-08 (same batch) | Covers through Phase 5+ (Autonomous Governance). Missing ~30 epics' scope definitions. |
| `epic-details.md` | 2026-03-08 (consolidation PR #324) | Contains only **18 epic detail entries** (Epics 1-21). **47 epics have no detail entry.** |
| `epic-list.md` | 2026-03-13 (governance syncs) | **Current — 65+ epics listed.** This is the only planning doc that stayed up to date. |
| `epics-and-stories.md` | 2026-03-13 (governance syncs) | **Current — story lists maintained.** |

### Epics Created Without PRD Updates

Every epic below was added to `epic-list.md` and `epics-and-stories.md` but has **no corresponding entries** in `requirements.md`, `product-scope.md`, or `epic-details.md`:

| Epic | Title | Created ~Date | Sprint Change Proposal? | PRD Updated? |
|---|---|---|---|---|
| 22 | Self-Driving Development Pipeline | 2026-03-03 | No | No |
| 23 | CLI Interface | 2026-03-03 | No | No |
| 24 | MCP/LLM Integration Server | 2026-03-03 | No | No |
| 25 | Todoist Integration | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07-todoist.md) | No |
| 26 | GitHub Issues Integration | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07-github-issues.md) | No |
| 27 | Daily Planning Mode | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07.md) | Partial — requirements.md has Phase 6+ section for this |
| 28 | Snooze/Defer | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07-snooze-defer.md) | Partial — product-scope.md Phase 3.5 covers this |
| 29 | Task Dependencies | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07-task-dependencies.md) | Partial — product-scope.md Phase 3.5+ covers this |
| 30 | Linear Integration | 2026-03-07 | Yes (sprint-change-proposal-2026-03-07-linear-integration.md) | Partial — product-scope.md Phase 5 mentions it |
| 31 | Expand/Fork Key Implementations | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08-expand-fork.md) | Partial — product-scope.md Phase 3.5+ covers it |
| 32 | Undo Task Completion | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08-undo-task-completion.md) | Partial — requirements.md Phase 6+ covers it |
| 33 | Seasonal Door Theme Variants | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08.md) | No |
| 34 | SOUL.md + Custom Dev Skills | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08-soul-skills.md) | Partial — requirements.md Developer Experience section |
| 35 | Door Visual Appearance | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08 via BMAD) | No |
| 36 | Door Selection Feedback | 2026-03-08 | Via course correction | No |
| 37 | Persistent BMAD Agents | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08-persistent-agents.md) | Partial — product-scope.md Phase 5+ |
| 38 | Dual Homebrew Distribution | ~2026-03-09 | Via course correction | No |
| 39 | Keybinding Display System | ~2026-03-09 | No formal proposal found | No |
| 40 | Beautiful Stats Display | 2026-03-08 | Yes (sprint-change-proposal-2026-03-08-beautiful-stats.md) | No |
| 41 | Charm Ecosystem Adoption | ~2026-03-09 | No formal proposal found | No |
| 42 | Application Security Hardening | ~2026-03-11 | No formal proposal found | No |
| 43 | Connection Manager Infrastructure | ~2026-03-11 | No formal proposal found | No |
| 44 | Sources TUI | ~2026-03-11 | No formal proposal found | No |
| 45 | Sources CLI | ~2026-03-11 | No formal proposal found | No |
| 46 | OAuth Device Code Flow | ~2026-03-11 | No formal proposal found | No |
| 47 | Sync Lifecycle & Advanced Features | ~2026-03-11 | No formal proposal found | No |
| 48 | Door-Like Doors Visual Enhancement | ~2026-03-11 | No formal proposal found | No |
| 49 | ThreeDoors Doctor | ~2026-03-11 | No formal proposal found | No |
| 50 | In-App Bug Reporting | ~2026-03-11 | No formal proposal found | No |
| 51 | SLAES | ~2026-03-11 | No formal proposal found | No |
| 52 | Envoy Three-Layer Firewall | ~2026-03-11 | No formal proposal found | No |
| 53 | Remote Collaboration | ~2026-03-11 | No formal proposal found | No |
| 54 | Gemini Research Supervisor | ~2026-03-11 | No formal proposal found | No |
| 55 | CI Optimization Phase 1 | 2026-03-11 | Yes (embedded in planning PR) | No |
| 56 | Door Visual Redesign | 2026-03-11 | Yes (embedded in planning PR) | No |
| 57 | LLM CLI Services | 2026-03-11 | Yes (embedded in planning PR) | No |
| 58 | Supervisor Shift Handover | 2026-03-11 | Yes (embedded in planning PR) | No |
| 59 | Full-Terminal Vertical Layout | 2026-03-11 | Yes (embedded in planning PR) | No |
| 60 | README Overhaul | 2026-03-11 | Yes (embedded in planning PR) | No |
| 61 | GitHub Pages User Guide | 2026-03-11 | Yes (embedded in planning PR) | No |
| 62 | Retrospector Reliability | 2026-03-12 | Yes (sprint-change-proposal-2026-03-12-retrospector-reliability.md) | No |
| 63 | ClickUp Integration | 2026-03-13 | Yes (sprint-change-proposal-2026-03-13-prd-coverage-gaps.md) | No |
| 64 | Cross-Computer Sync | 2026-03-13 | Yes (same as above) | No |
| 65 | CLI Test Coverage Hardening | 2026-03-13 | Yes (embedded in planning PR) | No |

**Summary:** 47+ epics created since early March. Of those:
- ~18 had formal sprint change proposals
- ~15 had no formal proposal at all (bulk-created in batch planning sessions)
- **0 updated `epic-details.md`** (which only covers Epics 1-21)
- **0 updated `requirements.md`** with new functional requirements
- **~6 had partial pre-existing coverage in `product-scope.md`** (features that were mentioned as future scope before being formalized as epics)

---

## 2. Root Cause Analysis

### The Three-Layer Failure

The PRD bypass is not a single bug — it's a systemic failure at three layers:

#### Layer 1: Skill Definitions Say the Right Thing But Can't Enforce It

**`/plan-work` (`.claude/commands/plan-work.md`):**
- Phase 2 says: "Launch `/bmad-bmm-agent-pm`" and "run `/bmad-bmm-correct-course`"
- Phase 5 says: "Update `docs/prd/epics-and-stories.md`, `docs/prd/epic-list.md`, `ROADMAP.md`"
- **Critical gap:** Phase 5 lists only `epics-and-stories.md`, `epic-list.md`, and `ROADMAP.md`. It does NOT mention `requirements.md`, `product-scope.md`, or `epic-details.md`.

**`/course-correct` (`.claude/commands/course-correct.md`):**
- Step 5 says: "Add stories to appropriate epic in `docs/prd/epics-and-stories.md`" and "Update `ROADMAP.md`"
- **Critical gap:** Step 5 does NOT mention `requirements.md`, `product-scope.md`, or `epic-details.md`.

**The skill definitions themselves are incomplete.** They update the "index docs" (epic-list, epics-and-stories, ROADMAP) but skip the "content docs" (requirements, product-scope, epic-details).

#### Layer 2: BMAD Agent Invocation Is Aspirational, Not Actual

The `/plan-work` skill says to invoke `/bmad-bmm-agent-pm` and run course correction. But:

1. **Slash command chaining doesn't work reliably.** When a worker runs `/plan-work`, it's already executing a slash command. Invoking `/bmad-bmm-agent-pm` inside that context requires the Claude instance to load a second agent persona mid-workflow. In practice, workers skip this or do a lightweight version.

2. **No enforcement mechanism.** There's no validation gate that checks "did the PM agent actually review and amend the PRD?" before the PR is created. Workers can (and do) proceed straight to story creation without Phase 2.

3. **Worker context window pressure.** Loading the PM agent, running course correction, loading party mode, loading the architect — each requires reading and processing large definition files. Workers running in isolated worktrees with finite context often shortcut these phases.

#### Layer 3: Governance Sync Masks the Problem

The project-watchdog agent actively syncs `epic-list.md` and `epics-and-stories.md` through "governance sync" PRs. This creates the illusion that planning docs are maintained, because the *index docs* are always current. But project-watchdog's mandate doesn't include `requirements.md`, `product-scope.md`, or `epic-details.md` — those are PM territory. Since the PM is never invoked, they rot.

### The "Built But Not Wired" Pattern

This is structurally identical to the connect wizard gap (Story 45.6): the BMAD PM agent has full PRD amendment capabilities (`[EP] Edit PRD`, `[CC] Course Correction`), but the skills that should invoke those capabilities don't wire to them correctly. The infrastructure exists but the integration is missing.

Specifically:
- The BMAD correct-course workflow (`_bmad/bmm/workflows/4-implementation/correct-course/instructions.md`) has comprehensive PRD analysis (Section 3: Artifact Conflict and Impact Analysis, checklist item 3.1 explicitly checks PRD conflicts)
- The PM agent has an `[EP] Edit PRD` menu option
- But `/plan-work` Phase 2 and `/course-correct` Step 5 don't call these with enough specificity, and workers don't follow through

### Contributing Factor: Batch Epic Creation

Around March 11, a large batch of epics (42-61) were created in rapid succession without individual sprint change proposals. This appears to have been a bulk planning session that bypassed the `/plan-work` pipeline entirely, going straight to epic/story creation without any PRD review.

---

## 3. Remediation Paths

### 3a. Fix the Pipeline

**Changes needed to prevent future bypass:**

1. **Update `/plan-work` Phase 5** to explicitly require updates to ALL PRD docs:
   ```
   - Update `docs/prd/requirements.md` with new functional/non-functional requirements
   - Update `docs/prd/product-scope.md` with scope additions for the appropriate phase
   - Update `docs/prd/epic-details.md` with detailed epic breakdown
   - Update `docs/prd/epics-and-stories.md` with epic and story outlines
   - Update `docs/prd/epic-list.md` with epic entry
   - Update `ROADMAP.md` with roadmap entry
   ```

2. **Update `/course-correct` Step 5** with the same explicit list.

3. **Add a validation gate to Phase 8 (PR creation) of `/plan-work`:**
   ```
   Before creating the PR, verify:
   - [ ] `requirements.md` has FR/NFR entries for every new capability
   - [ ] `product-scope.md` has the feature in the correct phase section
   - [ ] `epic-details.md` has a detailed breakdown for the new epic
   If any are missing, HALT and complete them before proceeding.
   ```

4. **Add the same gate to `/course-correct` Step 6 (Report).**

5. **Update project-watchdog's mandate** to flag when `epic-list.md` has epics that don't appear in `epic-details.md`.

**Effort:** Small — 4 file edits to skill definitions, 1 to watchdog config. Could be a single story.

### 3b. Forensic Reconstruction

**Scope of retroactive PRD update work:**

The gap covers **47 epics** (22-65+, minus the ~6 with partial coverage). For each:

- `requirements.md`: Needs new FR/NFR entries. Estimate: 2-4 requirements per epic × 47 = ~100-190 new requirement entries.
- `product-scope.md`: Needs scope sections. Estimate: A paragraph per epic, organized by phase. ~47 paragraph additions.
- `epic-details.md`: Needs full detail sections. Estimate: 50-100 lines per epic × 47 = ~2,350-4,700 lines. This is the largest effort.

**Batching strategy:**

1. **Tier 1 — Completed epics (priority):** Epics already COMPLETE should be reconstructed first since they represent delivered functionality that the PRD should document. Estimated ~20 completed epics.

2. **Tier 2 — In-progress epics:** Epics with active stories being implemented. Estimated ~10.

3. **Tier 3 — Not-started epics:** Future work. These can be done as part of normal planning when the epic comes up for implementation. Estimated ~17.

**Approach options:**

- **Option A: Dedicated reconstruction sprint.** Spawn 3-5 parallel workers, each handling 10 epics. Use existing story files, sprint change proposals, and party mode artifacts as source material. Estimated 3-5 worker-hours.

- **Option B: Incremental reconstruction.** Add a step to `/implement-story` that checks if the epic's PRD entries exist before starting work. If not, create them as part of the implementation workflow. This is slower but self-healing.

- **Option C: PM-led batch review.** Run `/bmad-bmm-agent-pm` in a dedicated session to review all 47 gaps at once, producing a single massive PRD update PR. Highest quality but highest supervisor attention cost.

**Recommendation:** Option A for Tier 1 (completed epics — we have all the information), Option B for Tiers 2-3 (self-healing going forward).

### 3c. PRD Drift Detection

**Should there be a watchdog for this?**

Yes. Two options:

1. **Extend project-watchdog:** Add a periodic check (daily or on PR merge) that compares epic counts across `epic-list.md`, `requirements.md`, `product-scope.md`, and `epic-details.md`. If any epic exists in `epic-list.md` but not in the others, file a GitHub issue.

2. **New `/reconcile-prd` skill:** A slash command that scans all PRD docs for consistency and reports gaps. Can be run manually or via cron.

**Recommendation:** Both. project-watchdog catches drift automatically; `/reconcile-prd` provides on-demand verification.

### 3d. Additional Recommendations

1. **Deprecate or restructure `epic-details.md`:** With 65+ epics, maintaining detailed breakdowns in a single file is unsustainable. Consider either:
   - Splitting into per-epic detail files (like story files)
   - Accepting that story files ARE the detail docs and removing the duplication
   - Keeping `epic-details.md` as a summary (1-2 paragraphs per epic) rather than full breakdowns

2. **Align doc authority:** Currently, `epic-list.md` and `epics-and-stories.md` are maintained by project-watchdog, while `requirements.md`/`product-scope.md`/`epic-details.md` are PM territory. This split means automated governance keeps indexes current while manual-only docs rot. Either automate PRD doc maintenance or accept they'll be permanently stale.

3. **Story template PRD link:** Add a field to the story template that records which requirements (FR/NFR) the story implements. This creates traceability and makes reconstruction easier.

---

## 4. Lessons Learned

### Process Pattern: "Built But Not Wired" (Skill Level)

This is the third instance of the "built but not wired" pattern in this project:

| Instance | What Existed | What Was Missing | Impact |
|---|---|---|---|
| **Connect Wizard** (Story 45.6) | Full TUI wizard in `internal/tui/` | Not wired into CLI entry point | Users couldn't access the wizard |
| **BMAD PM Agent** (This investigation) | Full PRD amendment workflow | Not wired into `/plan-work` and `/course-correct` effectively | 47 epics with no PRD coverage |
| **Correct-Course BMAD Workflow** | Comprehensive checklist with PRD conflict analysis (Section 3) | `/course-correct` skill bypasses the BMAD workflow, uses its own lightweight steps | PRD analysis never runs |

**Root pattern:** When a sophisticated capability exists in one layer (BMAD agent, TUI component) but the integration layer (skill definition, CLI entry point) provides a simpler shortcut that skips the capability.

### Why This Went Undetected

1. **The index docs masked the gap.** `epic-list.md` and `epics-and-stories.md` were always current, so the planning system *appeared* healthy.
2. **No consumer of PRD content docs.** No agent, skill, or workflow reads `requirements.md` or `epic-details.md` as input to daily operations. They're reference docs that nobody references.
3. **Rapid epic creation velocity.** 47 epics in ~11 days made individual PRD updates feel like overhead, leading to systematic skipping.
4. **Worker isolation.** Workers in isolated worktrees don't see the full PRD state. They create their story files and epics in local scope without awareness of the broader PRD consistency picture.

### Preventive Principle

> **Every process step that modifies shared state must have a corresponding validation gate that checks the modification was made.**

Sprint change proposals are created → validated by party mode → approved.
Stories are created → validated by story template → approved.
PRD amendments are... never validated. That's the gap.

---

## 5. Recommended Next Steps

1. **Immediate (Story):** Fix `/plan-work` and `/course-correct` skill definitions to include all 6 PRD docs
2. **Short-term (Epic):** Forensic reconstruction of PRD for completed epics (Tier 1)
3. **Medium-term (Story):** Add PRD drift detection to project-watchdog
4. **Long-term (Decision):** Decide whether `epic-details.md` should be deprecated in favor of story files as the detail layer

---

## Appendix: File Locations

- `/plan-work` skill: `.claude/commands/plan-work.md`
- `/course-correct` skill: `.claude/commands/course-correct.md`
- BMAD correct-course workflow: `_bmad/bmm/workflows/4-implementation/correct-course/instructions.md`
- BMAD PM agent: `_bmad/bmm/agents/pm.md`
- PRD requirements: `docs/prd/requirements.md` (587 lines, last meaningful update 2026-03-08)
- PRD product scope: `docs/prd/product-scope.md` (195 lines, last meaningful update 2026-03-08)
- PRD epic details: `docs/prd/epic-details.md` (1617 lines, 18 epics — last updated 2026-03-08)
- PRD epic list: `docs/prd/epic-list.md` (current, 65+ epics)
- PRD epics and stories: `docs/prd/epics-and-stories.md` (current)
