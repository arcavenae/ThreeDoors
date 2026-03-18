# Party Mode: Operationalize GitHub Label Usage Across Agents

**Date:** 2026-03-18
**Participants:** PM, Architect, SM
**Input:** Label Usage Gap Analysis (PR #806), Sprint Change Proposal (2026-03-18)
**Topic:** How to close the gap between label infrastructure and actual label usage

---

## Discussion Summary

### Round 1: Problem Validation

**PM:** The gap analysis is thorough. Zero PR labels across 800+ PRs is a clear systemic gap, not a bug. The root cause — "labels designed but not wired into agent workflows" — mirrors the PRD content doc gap we found (D-179): infrastructure exists but operational instructions are missing.

**Architect:** This is an operational wiring problem, not an architectural one. No new infrastructure needed. The label taxonomy (D-107), authority matrix (label-authority.md), and agent definitions are all correct — they just need explicit "when you do X, also apply label Y" instructions.

**SM:** Four stories is right. Story 72.1 (merge-queue) has the highest impact — it fills the entire PR labeling gap. Story 72.2 (envoy) fixes issue reliability. Stories 72.3 and 72.4 are lower-risk cleanup.

### Round 2: Approach Validation

**PM:** Option A (merge-queue owns PR labeling) is correct. The merge-queue's polling loop already examines every PR. Title-prefix inference (`feat:` → `type.feature`) is deterministic and doesn't require AI reasoning — it's a mechanical step in the validation checklist.

**Architect:** Agreed. No architectural concerns. The title-prefix mapping is a convention already enforced by CLAUDE.md commit message format. The `scope.in-scope` inference from story references is similarly mechanical.

**SM:** One concern: what happens when a PR title doesn't match any prefix? The merge-queue should skip type labeling rather than guess. Missing labels are better than wrong labels.

**PM:** Good point. The instruction should say "if missing, infer from title prefix" with a clear mapping, and do nothing if no match.

### Round 3: PRD Changes Review

**PM:** The proposed PRD changes are appropriate:
- FR-GOV1-3 cover the three new capabilities (PR labeling, issue lifecycle, mutual exclusivity)
- NFR-GOV1 captures the startup catch-up requirement
- Phase 6 placement is correct — this is governance/operations

**Architect:** No architectural docs need updating. This is purely operational.

**SM:** Story acceptance criteria should include verification that labels are actually applied — not just that the instructions exist in the agent definition.

---

## Decisions

### Adopted

| ID | Decision | Rationale |
|----|----------|-----------|
| D-184 | Merge-queue owns routine PR labeling via title-prefix inference | Natural extension of merge validation; deterministic mapping; no role confusion |
| D-185 | Envoy adds startup catch-up scan for unlabeled issues | Fills the gap when envoy is down; retroactive labeling within one polling cycle |

### Rejected

| ID | Option | Why Rejected |
|----|--------|--------------|
| X-124 | PR-shepherd applies PR labels | Blurs pr-shepherd's focused "branch health" role |
| X-125 | Dedicated label-enforcement agent | Over-engineering; 5-10 lines in existing agents suffice |

---

## Recommendations Applied

1. **Skip type label when no title prefix matches** — missing is better than wrong
2. **Story 72.2 includes mutual exclusivity enforcement** — explicit remove-before-add in envoy instructions
3. **Story 72.4 is P2** — retroactive cleanup is lower priority than preventing future gaps
4. **Verification criteria added** — each story's ACs include checking that labels are actually present after the workflow runs

---

## Artifact Location

- Sprint change proposal: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-03-18-label-operationalization.md`
- Gap analysis: `_bmad-output/planning-artifacts/label-usage-gap-analysis.md`
- This document: `_bmad-output/planning-artifacts/label-operationalization-party-mode.md`
