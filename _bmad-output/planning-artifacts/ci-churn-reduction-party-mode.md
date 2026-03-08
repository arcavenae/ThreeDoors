# CI Churn Reduction — Party Mode Consensus

> **Date:** 2026-03-08
> **Participants:** Winston (Architect), Amelia (Dev), Quinn (QA), John (PM)
> **Topic:** Review CI churn reduction research and reach consensus on recommendations

## Problem Summary

O(n²) CI waste from multi-agent parallel PR workflow. 7 open PRs generating 40 CI runs/hour, 60% of which are churn reruns. 33% of PRs are docs-only running full Go test suite unnecessarily.

## Adopted Approach (with rationale)

### Step 1: Path Filtering via `paths-ignore` (Do Today)

**What:** Add `paths-ignore` to CI workflow trigger to skip docs-only PRs.

**Rationale:** Zero risk, 15-minute change, eliminates ~33% of all CI runs. Docs PRs cannot affect Go code. Use `paths-ignore` rather than `paths` because it's safer — we won't accidentally miss new file types that matter.

**Implementation:**
```yaml
on:
  pull_request:
    branches: [main]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'LICENSE'
      - '.gitignore'
      - 'ROADMAP.md'
      - 'CHANGELOG.md'
      - 'SOUL.md'
      - '_bmad/**'
      - '_bmad-output/**'
  push:
    branches: [main]
```

### Step 2: Main CI Failure Circuit Breaker (Do Today)

**What:** Add logic to merge-queue agent to pause merging if push-to-main CI fails.

**Rationale:** This is the prerequisite safety net for Step 3. If we relax the up-to-date requirement, we need assurance that post-merge CI failures are caught immediately and don't cascade. Without this, multiple PRs could merge onto a broken main before anyone notices.

### Step 3: Relax Up-to-Date Requirement (Do After Steps 1-2)

**What:** Modify merge-queue agent to stop requiring PRs be up-to-date with main before merging.

**Rationale:** This is the highest-leverage change (~70-80% CI run reduction). ThreeDoors' architecture makes this safe:
- Go tests are hermetic (no shared state, no integration tests between providers)
- No database or external service dependencies
- Interface-based provider pattern means changes to one provider don't affect others
- Git merge conflicts catch most conflicting file edits
- Push-to-main CI trigger provides a post-merge safety net (made reliable by Step 2)

### Metric to Track

**CI runs per merged PR.** Current: ~5-10. Target: ~1.5 (one on branch, one on main push).

## Rejected Options (with reasons)

### GitHub Native Merge Queue (Option 3) — REJECTED for now

**Reason:** Too much operational complexity for uncertain marginal gain. Requires:
- Reworking the multiclaude merge-queue agent to use `gh pr merge --merge-queue`
- Updating CI workflow to trigger on `merge_group` events
- Configuring merge group sizes and timeouts
- Testing the entire pipeline end-to-end

The quick wins (path filtering + relaxed up-to-date) should provide 80%+ reduction. If insufficient after a week of measurement, revisit merge queue.

### CI Caching Improvements (Option 5) — REJECTED

**Reason:** Marginal gains (~10-20% per-run time reduction). Doesn't address the core O(n²) run count problem. Go builds are already fast (~3.5 min). Existing caching (Go modules, Docker buildx, golangci-lint) is already configured.

### Conditional Job Execution via dorny/paths-filter (Option 4) — REJECTED

**Reason:** Adds workflow complexity (extra `changes` job, conditional `if` clauses on every job). The simpler `paths-ignore` at the workflow trigger level achieves similar results for our PR mix (33% docs-only). If we needed finer-grained control (e.g., skip Docker E2E but keep quality-gate for Go-only changes), this would be worth revisiting, but the current approach is sufficient.

## Deferred Options

### Make Benchmarks Non-Blocking (Option 6) — ADOPTED but DEFERRED

**Reason:** Good idea in principle — benchmarks are informational, not correctness checks. However, Quinn (QA) recommends keeping all three checks required initially. Don't remove two safety layers (up-to-date + non-blocking checks) simultaneously. Measure the impact of Tier 1 changes first, then consider relaxing individual check requirements.

**Quinn's specific concern:** Docker E2E tests validate the actual built container and golden file output. Making these non-blocking while also relaxing up-to-date removes two independent safety layers at once. Keep Docker E2E required.

## Key Discussion Points

1. **Winston (Architect):** Favored layered approach — combine Options 1+2 for defense-in-depth. Emphasized that the push-to-main CI trigger is the safety net, but a safety net only works if someone catches you. Made the circuit breaker (Step 2) a prerequisite for Step 3.

2. **Amelia (Dev):** Confirmed no GitHub-native branch protection configured — merge-queue agent enforces rules programmatically. This gives us full control. Recommended `paths-ignore` over `paths` for safer, easier maintenance.

3. **Quinn (QA):** Flagged the quality risk of relaxing multiple safety layers simultaneously. Insisted on keeping Docker E2E required and implementing the main-CI circuit breaker before relaxing up-to-date. Wants a week of data before considering further changes.

4. **John (PM):** Emphasized ROI — 210 wasted job-minutes per hour during active development. Questioned why path filtering wasn't done from day one. Formalized the three-step sequenced rollout with measurement after each step.

## Consensus

Unanimous agreement on the three-step sequenced approach. All agents agreed that GitHub Merge Queue should be deferred pending measurement of quick wins. The key insight: implement changes in risk-ascending order, with each step de-risking the next.
