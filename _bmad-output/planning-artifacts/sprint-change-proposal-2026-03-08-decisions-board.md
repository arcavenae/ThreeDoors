# Sprint Change Proposal: Knowledge Decisions Board

**Date:** 2026-03-08
**Trigger:** Research complete (PR #235) — decision management strategy ready for implementation
**Scope:** Minor — add 3 stories to Epic 0 (Infrastructure)
**Change Type:** Direct Adjustment (add stories to existing epic)

## 1. Issue Summary

Decisions across ThreeDoors are scattered across 5+ locations (ADRs, research docs, party mode artifacts, story files, PR comments). Completed research has no decision artifact trail. Rejected options get re-proposed because no one recorded why they were rejected. Workers lack context on prior decisions.

Research (PR #235, `decision-management-research.md`) evaluated 6 approaches and party mode consensus (`_bmad-output/planning-artifacts/decision-management-party-mode.md`) adopted a three-component system:

1. **Knowledge Board** (`docs/decisions/BOARD.md`) — kanban-style columns tracking decision lifecycle
2. **Standardized Artifact Endings** — all party mode artifacts get a Decisions Summary table
3. **Board Hygiene Sweep** — periodic scan for unindexed decisions

## 2. Impact Analysis

### Epic Impact
- **Epic 0 (Infrastructure):** Add 3 new stories (0.25, 0.26, 0.27). Epic moves from 19/20 to 19/23 (with 0.20 still not started).
- **All other epics:** No impact. This is orthogonal process infrastructure.

### Artifact Impact
- **CLAUDE.md:** Add one instruction line to check BOARD.md before implementing
- **PRD:** No changes needed
- **Architecture:** No changes — this is docs/process infrastructure, not code architecture
- **UI/UX:** No impact

### Technical Impact
- **No code changes.** All deliverables are markdown documentation files.
- **No CI/CD impact.** No tests to add for documentation.

## 3. Recommended Approach

**Direct Adjustment** — add 3 stories to Epic 0:

| Story | Title | Effort | Risk |
|-------|-------|--------|------|
| 0.25 | Knowledge Decisions Board — Create & Seed | Low | Low |
| 0.26 | Artifact Format Standardization & Backfill | Medium | Low |
| 0.27 | Board Hygiene Sweep Process | Low | Low |

**Rationale:** Research is complete, consensus is clear, deliverables are well-defined documentation. No architectural risk. Stories are independent of all other active epics.

**Alternatives considered:**
- Single large story: Rejected — too much scope for one PR, harder to review
- Five granular stories: Rejected — unnecessary ceremony for docs-only work

## 4. Detailed Change Proposals

### Story 0.25: Knowledge Decisions Board — Create & Seed

**Deliverables:**
- Create `docs/decisions/BOARD.md` with kanban columns (Open Questions, Active Research, Pending Recommendations, Decided, Rejected, Superseded)
- Seed "Decided" column from all 28 existing ADRs
- Add CLAUDE.md instruction: "Before implementing, check `docs/decisions/BOARD.md` for relevant prior decisions"
- Create `docs/decisions/README.md` explaining the board system

### Story 0.26: Artifact Format Standardization & Backfill

**Deliverables:**
- Define standardized Decisions Summary table format for party mode artifacts
- Backfill board entries from existing ~46 party mode artifacts (extract decisions into board)
- Backfill board entries from research docs that contain recommendations
- Update standing orders for party mode artifact format

### Story 0.27: Board Hygiene Sweep Process

**Deliverables:**
- Define sweep process: what to scan, what to flag, frequency
- Create sweep task definition (runnable as BMAD task or `/loop`)
- Document sweep outputs and escalation paths
- Add sweep schedule to standing orders

## 5. Implementation Handoff

**Scope classification:** Minor — direct implementation by workers via `/implement-story`

**Handoff:**
- Stories 0.25-0.27 can be implemented sequentially by workers
- No PO/SM coordination needed — these are process docs
- No architect involvement needed — no code architecture impact
- Supervisor updates ROADMAP.md Epic 0 count after merge

**Success criteria:**
- BOARD.md exists with all 28 ADR entries seeded
- Workers can find prior decisions by reading BOARD.md
- Party mode artifacts have standardized endings
- Sweep process is defined and documented

## Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| 3 stories under Epic 0 | Adopted | Right granularity for docs-only work | Single story (too big), 5 stories (too granular) |
| Sequential implementation | Adopted | 0.26 depends on 0.25 (board must exist for backfill) | Parallel (dependency conflict) |
| No code changes | Adopted | Process infrastructure is docs-only | Adding tooling (against project philosophy) |
