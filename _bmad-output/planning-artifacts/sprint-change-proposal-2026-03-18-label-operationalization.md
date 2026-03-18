# Sprint Change Proposal: Operationalize GitHub Label Usage Across Agents

**Date:** 2026-03-18
**Author:** fancy-dolphin (worker agent, /plan-work pipeline)
**Source:** Label Usage Gap Analysis (PR #806, `_bmad-output/planning-artifacts/label-usage-gap-analysis.md`)
**Proposed Epic:** 72 — Operationalize GitHub Label Usage

---

## Problem Statement

ThreeDoors has excellent label infrastructure (27 well-designed scoped labels, complete authority matrix, detailed agent definitions) but **zero PRs have ever had labels applied** and **issue labeling is inconsistent**. The root cause is not missing documentation — it's that no agent definition includes explicit workflow instructions to apply labels as part of its routine operations.

### Impact Analysis

- **PR discoverability:** With 800+ merged PRs and zero labels, filtering PRs by type/scope is impossible
- **Issue reliability:** Issues created during envoy downtime never get retroactively labeled
- **Mutual exclusivity violations:** Existing issues have conflicting labels (e.g., both `triage.in-progress` and `triage.complete`)
- **Missing label:** `resolution.wontfix` doesn't exist on GitHub despite being in the taxonomy
- **Dashboard usability:** Label-based GitHub dashboards and queries are unreliable

### Scope

This is entirely agent definition and operational documentation changes — **no application code**. Changes affect:
- `agents/merge-queue.md` — add PR labeling section
- `agents/envoy.md` — add startup catch-up, context warning, mutual exclusivity enforcement
- Supervisor memory — add label discipline instructions
- `docs/operations/label-authority.md` — update merge-queue entry
- GitHub label API — create missing label, fix existing violations

---

## Proposed Approach (Adopted)

### Option A: Merge-queue owns routine PR labeling (ADOPTED)

Merge-queue already reads every open PR during its polling loop. Adding label inference from PR title prefix is a natural, coherent extension of its validation role.

**Rationale:**
- Merge-queue is the natural owner because it already validates PR metadata (scope, CI, reviews)
- Adding label classification is a small extension with no role confusion
- Labels applied before merge ensures all merged PRs are classified
- Inference from title prefix is deterministic and doesn't require AI reasoning

### Option B: PR-shepherd applies PR labels (REJECTED)

PR-shepherd also reads every PR. However, pr-shepherd's responsibility is "branch health" not "PR metadata" — adding labeling would blur its focused role. See X-124.

### Option C: Create a dedicated label-enforcement agent (REJECTED)

A new persistent agent solely for label management. Over-engineering for what amounts to 5-10 lines of instructions added to existing agents. See X-125.

---

## Stories (4 total)

| Story | Title | Priority | Scope |
|-------|-------|----------|-------|
| 72.1 | Merge-Queue PR Labeling | P1 | Add PR labeling section to merge-queue agent definition |
| 72.2 | Envoy Label Resilience | P1 | Startup catch-up, context exhaustion warning, mutual exclusivity enforcement |
| 72.3 | Supervisor Label Discipline & Missing Label | P1 | Supervisor instructions + create resolution.wontfix on GitHub |
| 72.4 | Retroactive Label Cleanup | P2 | One-time fix of unlabeled/mislabeled issues |

---

## PRD Changes Required

### `docs/prd/requirements.md`
- Add FR-GOV1: Agents shall apply appropriate GitHub labels to PRs during merge validation
- Add FR-GOV2: Agents shall apply triage labels to issues on detection and maintain label state through the triage lifecycle
- Add FR-GOV3: Agents shall enforce mutual exclusivity for scoped labels (remove old before applying new)
- Add NFR-GOV1: Unlabeled issues from agent downtime shall be retroactively labeled within one polling cycle of agent restart

### `docs/prd/product-scope.md`
- Add to Phase 6 (Developer Experience & Governance) under "Autonomous Project Governance": GitHub label operationalization across agent workflows

### `docs/prd/epic-details.md`
- Add Epic 72 detailed breakdown with all 4 stories and acceptance criteria

---

## Risk Assessment

- **Low risk:** All changes are text/documentation — no application code, no build changes
- **No merge conflict risk:** Agent definition files rarely have parallel edits
- **Rollback:** Simple git revert if label automation causes issues
- **Dependency:** None — all stories are independent except 72.4 depends on 72.3 (needs resolution.wontfix to exist)
