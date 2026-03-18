# Label Usage Gap Analysis — Why Labels Aren't Being Applied

**Date:** 2026-03-18
**Researcher:** kind-rabbit (worker agent)
**Purpose:** Identify why GitHub labels are not being consistently applied to issues and PRs despite comprehensive label infrastructure

---

## Executive Summary

ThreeDoors has **excellent label infrastructure** — 27 well-designed scoped labels, a complete authority matrix (`docs/operations/label-authority.md`), detailed agent definitions with label responsibilities, and thorough research/party mode artifacts. Despite all this, **zero PRs have ever had labels applied**, and **recent issues are inconsistently labeled**. The root cause is not missing documentation or labels — it's that **no agent definition includes explicit instructions to apply labels as part of its workflow**, and PR labeling was never assigned to any agent.

---

## 1. Current State of Labels

### 1.1 Labels That Exist (27 total, as designed)

All 27 labels from the party mode taxonomy (D-107) are correctly created on GitHub with proper colors and descriptions. The `dependencies` label (managed by Renovate) also exists, making 28 total.

**Notable:** `resolution.wontfix` is **missing** from GitHub despite being in the 27-label taxonomy. Only 27 labels exist but the taxonomy specifies 27 including `resolution.wontfix`. Current count is 27 = 26 scoped + `dependencies`. The label was either never created or was deleted.

### 1.2 Actual Label Usage on Issues

| Issue | Created | Labels Applied | Notes |
|-------|---------|---------------|-------|
| #803 | 2026-03-18 | **None** | Most recent issue — no envoy triage |
| #770 | 2026-03-15 | type.bug, triage.in-progress, priority.p1, agent.envoy | Correctly labeled by envoy |
| #725 | 2026-03-13 | type.bug, triage.in-progress, priority.p1, agent.envoy | Correctly labeled by envoy |
| #592 | 2026-03-12 | type.bug, triage.complete, priority.p1, agent.envoy | Correctly labeled by envoy |
| #466 | 2026-03-11 | **None** | Created same day as label migration — may have been missed |
| #414 | 2026-03-09 | **None** | Pre-migration |
| #408 | 2026-03-09 | **None** | Pre-migration |
| #396 | 2026-03-09 | **None** | Pre-migration |
| #334 | pre-migration | type.bug, triage.complete, priority.p1 | Labels correctly migrated (rename preserved associations) |
| #330 | pre-migration | type.bug, triage.in-progress, triage.complete, priority.p1 | Has TWO triage labels (violates mutual exclusivity) |
| #296 | pre-migration | type.bug, triage.complete, priority.p1, type.infra, scope.in-scope, process.fast-track | Well-labeled, but has TWO type labels (violates mutual exclusivity) |

**Pattern:** Envoy labels issues when it's running and responsive to HEARTBEATs. Unlabeled issues (#803, #466) were likely created during periods when envoy was down, restarting, or had exhausted its context window. Older unlabeled issues (#414, #408, #396) predate the label migration or were created during agent downtime.

### 1.3 Actual Label Usage on PRs

**Zero PRs have ever had labels applied.** Checked the 20 most recently merged PRs — all have empty label arrays. This includes PRs that touched workflows (should have gotten `status.needs-human`), PRs that were out of scope (should have gotten `scope.out-of-scope`), and routine PRs (could have gotten `type.docs`, `type.feature`, etc.).

---

## 2. Who SHOULD Be Applying Labels

### 2.1 Issue Labels — Envoy (Primary), Supervisor (Override)

Per `agents/envoy.md` and `docs/operations/label-authority.md`:

| Responsibility | Agent | What They Should Apply |
|---------------|-------|----------------------|
| **Initial triage** | Envoy | `triage.new` + `type.*` on detection |
| **Screening progress** | Envoy | `triage.in-progress` (replacing `triage.new`) |
| **Classification** | Envoy | `priority.*`, `agent.envoy` |
| **Triage completion** | Envoy | `triage.complete`, `scope.*` |
| **Fast-track** | Envoy | `process.fast-track` |
| **Staleness** | Envoy | `status.stale` |
| **Scope decisions** | Supervisor | `scope.in-scope` / `scope.out-of-scope` |
| **Worker assignment** | Supervisor | `agent.worker` |
| **Resolution** | Envoy (propose) + Supervisor (confirm) | `resolution.duplicate` / `resolution.wontfix` |

**Assessment: Partially working.** When envoy is running, it does apply labels during triage (#770, #725, #592). The gap is:
- Envoy doesn't always apply `triage.new` first (jumps to `triage.in-progress`)
- Envoy never applies `scope.*` labels (despite being documented as proposing them)
- Supervisor never applies `agent.worker` when dispatching workers
- No agent applies `resolution.*` when closing issues
- Issues created during envoy downtime never get retroactively labeled

### 2.2 PR Labels — **Nobody Is Assigned**

This is the **primary gap**. The label authority matrix and agent definitions specify:

- **Merge-queue** can apply `status.needs-human`, `scope.out-of-scope`, `status.do-not-merge`, `broke-main` — but only in specific exceptional scenarios (workflow files, scope violations, emergency mode). It has **no routine PR labeling responsibility**.
- **PR-shepherd** can apply `status.stale` — but only for PRs with no activity past staleness threshold. No routine labeling.
- **Envoy** — issues only. No PR responsibility.
- **Supervisor** — no routine PR labeling defined.
- **Workers** — not label-aware per design.

**No agent is responsible for applying `type.*` or `priority.*` labels to PRs.** The entire label taxonomy was designed around issues. PR labeling was briefly mentioned in the research (Section 5.5 of `scoped-labels-research.md` mentions "status labels for PRs and issues") but the operational docs (`label-authority.md`) describe issue workflows exclusively.

---

## 3. Root Cause Analysis

### 3.1 PR Labels: Never Designed Into Any Workflow

The scoped labels research and party mode focused on **issue triage**. The triage flow diagram in `label-authority.md` shows an issue-centric pipeline. No equivalent PR workflow exists. The party mode's "consumer challenge" (Round 2) tested each label against "who reads it?" — but the answers were always about issue consumers, not PR consumers.

The only PR-specific label usage defined is reactive/exceptional:
- `status.needs-human` — merge-queue applies when PR touches `.github/workflows/`
- `scope.out-of-scope` — merge-queue applies when scope is questionable
- `status.do-not-merge` — supervisor/human applies as a merge block
- `status.stale` — pr-shepherd applies after 7+ days inactivity
- `broke-main` — merge-queue creates ad-hoc after CI break

None of these are routine — they're all edge-case signals.

### 3.2 Issue Labels: Envoy Downtime Creates Gaps

Envoy is an ephemeral agent that exhausts its context window after ~12 hours or ~20+ triage cycles (documented in envoy.md's context exhaustion risk note... actually missing from envoy.md but present in merge-queue.md and pr-shepherd.md). When envoy is down:
- New issues go unlabeled
- No catch-up mechanism exists to label issues missed during downtime
- The HEARTBEAT cron fires but goes unanswered if envoy has crashed

### 3.3 Mutual Exclusivity Not Enforced

Issue #330 has both `triage.in-progress` and `triage.complete`. Issue #296 has both `type.bug` and `type.infra`. The convention enforcement protocol in `label-authority.md` (remove old label before applying new) is documented but not mechanically enforced. Agents sometimes add without removing.

### 3.4 Missing Label: `resolution.wontfix`

The taxonomy specifies 27 labels but `resolution.wontfix` doesn't exist on GitHub. It was likely missed during the Story 0.44 migration. This means the merge-queue's documented behavior of labeling superseded PRs with `resolution.wontfix` silently fails.

---

## 4. Recommendations

### 4.1 PR Label Responsibilities — New Agent Instructions Needed

**Who should label PRs?** Two options:

**Option A: Merge-queue applies PR labels during validation (Recommended)**

Merge-queue already reads every open PR during its polling loop. Adding label application is a natural extension:

```
During merge validation, before merging:
1. If PR has no type.* label: infer from PR title/branch name and apply
   - "feat:" or "feature" → type.feature
   - "fix:" or "bug" → type.bug
   - "docs:" → type.docs
   - "chore:" or "refactor:" → type.infra
2. If PR title contains a story reference (Story X.Y): apply scope.in-scope
3. Apply agent.worker if the PR was created by a worker branch (work/*)
```

**Option B: PR-shepherd applies PR labels during polling**

PR-shepherd also reads every open PR. It could classify PRs during its staleness/conflict checks. However, pr-shepherd's responsibility is "branch health" not "PR metadata" — this would blur its role.

**Recommendation: Option A.** Merge-queue is the natural owner because it already validates PR metadata (scope, CI, reviews) before merging. Adding label classification is a small, coherent extension.

### 4.2 Envoy Improvements for Issue Labels

Add to `agents/envoy.md`:

1. **Startup catch-up:** On startup/restart, scan all open issues for missing labels. Apply `triage.new` + `type.*` to any unlabeled issue.

2. **Context exhaustion warning:** Add the same context exhaustion risk note that merge-queue and pr-shepherd have: "After ~12 hours or ~20+ triage cycles, context fills and the agent silently stops responding."

3. **Mutual exclusivity enforcement:** Before applying any label in an exclusive scope, explicitly remove existing labels in that scope:
   ```bash
   # Before applying triage.in-progress:
   gh issue edit <N> --remove-label triage.new
   gh issue edit <N> --add-label triage.in-progress
   ```

### 4.3 Missing Label Creation

Create `resolution.wontfix`:
```bash
gh label create "resolution.wontfix" --color cfd3d7 --description "Will not be addressed — see comment for reason"
```

### 4.4 Supervisor Label Discipline

Add to `agents/supervisor.md`:

1. When dispatching a worker for an issue, apply `agent.worker` to the issue
2. When making scope decisions, apply `scope.in-scope` or `scope.out-of-scope`
3. When confirming resolution, apply `resolution.duplicate` or `resolution.wontfix`

### 4.5 Retroactive Label Cleanup

Create a one-time story to:
1. Label all currently unlabeled open issues (#803, #466, #414, #408, #396, #278, #218)
2. Fix mutual exclusivity violations (#330: remove `triage.in-progress`; #296: remove `type.bug` since `type.infra` is more specific)
3. Create `resolution.wontfix` label

---

## 5. Concrete Agent Definition Changes

### 5.1 `agents/merge-queue.md` — Add PR Label Section

Add after the "Merge Validation Checklist" section:

```markdown
### PR Labeling (Applied During Validation)

Before merging, ensure the PR has appropriate labels:

1. **Type label** — if missing, infer from PR title prefix:
   - `feat:` / `feature` → `type.feature`
   - `fix:` → `type.bug`
   - `docs:` → `type.docs`
   - `chore:` / `refactor:` / `ci:` → `type.infra`
   - `test:` → `type.infra`

2. **Scope label** — if PR title references a Story (e.g., "Story X.Y"):
   - Apply `scope.in-scope`

3. **Agent label** — if PR branch starts with `work/`:
   - Apply `agent.worker`

Apply labels via:
```bash
gh pr edit <number> --add-label <label>
```
```

### 5.2 `agents/envoy.md` — Add Startup Catch-Up and Context Warning

Add to "Your rhythm" section:
```markdown
**On startup/restart (catch-up):**
- Scan all open issues: `gh issue list --state open --json number,labels`
- For issues with zero labels: apply `triage.new` + classify with `type.*`
- This catches issues created during previous envoy downtime
```

Add at bottom:
```markdown
## Context Exhaustion Risk

After ~12 hours or ~20+ triage cycles, context fills and the agent silently stops responding. See [persistent-agent-ops.md](../docs/operations/persistent-agent-ops.md). The supervisor should restart this agent proactively every 4-6 hours.
```

### 5.3 `agents/supervisor.md` — Add Label Discipline

Add to operational notes:
```markdown
### Label Application on Dispatch

When dispatching a worker for an issue:
1. Apply `agent.worker` to the issue: `gh issue edit <N> --add-label agent.worker`
2. When making scope decisions, apply `scope.in-scope` or `scope.out-of-scope`
3. When confirming closure, apply appropriate `resolution.*` label
```

### 5.4 `agents/pr-shepherd.md` — No Changes Needed

PR-shepherd's label responsibilities (setting `status.stale`) are already documented and narrow. No changes recommended.

### 5.5 `agents/project-watchdog.md` — No Changes Needed

Project-watchdog doesn't interact with labels. No changes recommended.

---

## 6. Summary: Issue Labels vs PR Labels

| Dimension | Issue Labels | PR Labels |
|-----------|-------------|-----------|
| **Designed?** | Yes — comprehensive triage flow | No — only exceptional cases |
| **Documented?** | Yes — label-authority.md, envoy.md | Partially — merge-queue.md mentions a few |
| **Working?** | Partially — when envoy is up | Not at all — zero PRs labeled |
| **Primary gap** | Envoy downtime, missing catch-up | No agent assigned routine labeling |
| **Fix complexity** | Low — add catch-up to envoy startup | Low — add labeling step to merge-queue |
| **Agent responsible** | Envoy (triage), Supervisor (scope/assignment) | Merge-queue (recommended) |

---

## 7. Implementation Priority

1. **Create `resolution.wontfix` label** — trivial, should be done immediately
2. **Add PR labeling to merge-queue.md** — highest impact, fills the entire PR gap
3. **Add startup catch-up to envoy.md** — fills the issue gap during envoy downtime
4. **Add context exhaustion warning to envoy.md** — helps supervisor proactively restart
5. **Add label discipline to supervisor.md** — fills `agent.worker` and `scope.*` gaps
6. **One-time retroactive label cleanup** — fix existing unlabeled/mislabeled issues
7. **Mutual exclusivity enforcement** — add explicit remove-before-add to envoy's labeling logic

---

## Sources

- `docs/operations/label-authority.md` — authoritative label authority matrix
- `agents/envoy.md` — envoy agent definition with label authority table
- `agents/merge-queue.md` — merge-queue agent definition
- `agents/pr-shepherd.md` — pr-shepherd agent definition
- `docs/stories/0.44.story.md` — label migration story (Done, PR #513)
- `docs/stories/0.45.story.md` — agent definition updates for labels (Done, PR #520)
- `docs/stories/0.46.story.md` — label authority documentation (Done, PR #519)
- `_bmad-output/planning-artifacts/scoped-labels-party-mode.md` — party mode deliberation
- `_bmad-output/planning-artifacts/scoped-labels-research.md` — initial research
- `docs/decisions/ARCHIVE.md` — decisions D-106 through D-111
- GitHub API: `gh label list`, `gh pr list`, `gh issue list`
