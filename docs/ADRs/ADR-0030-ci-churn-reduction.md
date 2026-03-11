# ADR-0030: CI Churn Reduction — Branch Protection & Merge Queue Optimization

- **Status:** Accepted
- **Date:** 2026-03-08
- **Decision Makers:** Supervisor, cool-platypus (research)
- **Related Story:** 0.20
- **Related Research:** [CI Churn Reduction Research](../research/ci-churn-reduction-research.md)
- **Related ADRs:** ADR-0028 (CI Quality Gates)

## Context

Our multiclaude workflow produces 5-7 PRs in parallel. With branch protection requiring PRs to be up-to-date with main AND CI green before merging, every PR merge invalidates all other open PRs, forcing `update-branch` + CI rerun. With N open PRs, each merge triggers N-1 reruns, producing O(n^2) total CI runs.

**Observed impact (2026-03-08):**
- 40 CI runs in a 1-hour window, ~60% were churn reruns
- ~210 wasted job-minutes per hour during active development
- 33% of PRs are docs-only but run the full Go test/benchmark/Docker E2E suite

## Decision

### Phase 1: Relax "Require Up-to-Date" Enforcement

The merge-queue agent no longer requires PRs to be rebased on main before merging. The pr-shepherd agent no longer proactively rebases PRs.

**Rationale:**
- Go tests are fully isolated — no cross-PR integration dependencies
- Most PRs touch completely different files (docs vs. different feature areas)
- Git merge conflicts catch most conflicting edits
- The `push` trigger on main re-runs CI, catching any post-merge breakage
- Impact: ~70-80% reduction in CI runs

**Safety net:** If push-to-main CI fails after a merge, the merge-queue agent enters emergency mode and halts further merges until fixed. **Circuit breaker is now operational** (Story 0.36) — the merge-queue agent prompt includes an explicit post-merge CI check workflow with polling, emergency mode entry/exit, and `broke-main` labeling.

**Admin action required:** If GitHub branch protection has "Require branches to be up to date before merging" enabled, the repo admin should disable it. This is a repo settings change, not a code change.

### Phase 2: Path-Based CI Triggers for Docs-Only PRs

CI workflow updated with `dorny/paths-filter` to conditionally skip Go-related jobs when only documentation files change. A lightweight `docs-pass` job provides a passing status for docs-only PRs.

**Path filters:**
- Go code: `**.go`, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`
- Docker: `Dockerfile*`, `docker-compose*`
- CI config: `.github/workflows/**`

Docs-only PRs (matching none of the above) skip `quality-gate`, `benchmarks`, and `test-docker-e2e` entirely. The `docs-pass` job runs in ~10 seconds and provides a green check.

**Impact:** ~33% of PRs skip the full CI suite entirely.

### Phase 3: GitHub Native Merge Queue — Deferred

**Decision:** Defer GitHub merge queue adoption.

**Rationale:**
- Relaxing the up-to-date requirement (Phase 1) already eliminates the O(n^2) cascade
- Path filtering (Phase 2) eliminates waste on docs-only PRs
- GitHub merge queue adds latency (PRs wait in queue) without proportional benefit now that the cascade is eliminated
- The multiclaude merge-queue agent already provides sequenced merge logic
- Re-evaluate if post-merge CI failures become frequent (indicating the relaxed rule is causing integration issues)

**Re-entry gate:** If main branch CI fails more than 3 times in a week due to incompatible PR merges, reconsider enabling GitHub native merge queue.

## Alternatives Considered

1. **GitHub Native Merge Queue** — Deferred (adds latency, marginal benefit after Phase 1)
2. **Conditional job execution with dorny/paths-filter for code PRs** — Could skip benchmarks for non-benchmark-affecting changes, but adds complexity for marginal gain
3. **CI caching improvements** — Already well-cached; marginal per-run savings don't address run count
4. **Make benchmarks/Docker E2E non-blocking** — Considered but kept as-is; these catch real regressions

## Consequences

- PRs merge faster — no waiting for rebase + CI rerun cycles
- Docs-only PRs complete in seconds instead of ~3.5 minutes
- Theoretical risk: two PRs could pass CI independently but break when merged together. Mitigated by push-to-main CI and merge-queue emergency mode.
- Agent prompts updated to reflect new workflow (no mandatory rebase, no update-branch calls)
