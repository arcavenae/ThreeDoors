# CI Churn Reduction Research

> **Date:** 2026-03-08
> **Author:** cool-platypus (research worker)
> **Problem:** O(n²) CI waste from multi-agent parallel PR workflow

## Problem Statement

Our multiclaude workflow produces 5-7 PRs in parallel. Branch protection requires PRs to be up-to-date with main AND CI green before merging. Every PR merge invalidates all other open PRs, forcing `update-branch` + CI rerun. With N open PRs, each merge triggers N-1 reruns, producing O(n²) total CI runs.

**Observed impact (2026-03-08 snapshot):**
- 7 open PRs at time of analysis
- 40 CI runs examined in ~1 hour window
- Multiple branches show 2-3 runs each (brave-lion: 4 runs, proud-squirrel: 3 runs, bright-badger: 2 runs)
- Each CI run takes ~3.5 minutes (PR) or ~6.5 minutes (push to main)
- Estimated: 60-70% of CI runs are "churn" reruns from branch updates, not actual code changes

## Current State Analysis

### CI Pipeline (`.github/workflows/ci.yml`)

**Triggers:** `pull_request` to main, `push` to main

**Jobs (PR triggers):**

| Job | What It Does | Duration | Required? |
|-----|-------------|----------|-----------|
| `quality-gate` | Format check (gofumpt), go vet, golangci-lint, tests w/ race+coverage, coverage floor (75%), coverage comment, build validation | ~3.5 min | Yes (core) |
| `benchmarks` | Performance benchmarks, NFR13 validation | ~3.5 min | No dependency |
| `test-docker-e2e` | Docker build + E2E tests + golden file diff | ~3.5 min | No dependency |

**Jobs (push to main only):**

| Job | What It Does | Duration | Required? |
|-----|-------------|----------|-----------|
| `build-binaries` | Cross-platform binary builds | ~2 min | Needs quality-gate |
| `sign-and-notarize` | Apple code signing + notarization | ~5 min | Needs build-binaries, conditional |
| `release` | GitHub release creation | ~1 min | Needs sign-and-notarize |
| `update-homebrew` | Homebrew tap update | ~1 min | Needs release, conditional |

**Key observation:** All three PR jobs (`quality-gate`, `benchmarks`, `test-docker-e2e`) run in parallel on every PR event. There is no path filtering — docs-only PRs run the full suite including Docker E2E tests and benchmarks.

### Branch Protection

The branch protection API returns 404. No rulesets configured either (empty array). This suggests:
- Branch protection may be configured through GitHub UI settings not exposed via classic API
- OR the multiclaude merge-queue agent enforces merge rules programmatically (checking CI status + up-to-date before merging)

**Implication:** If protection is enforced by our merge-queue agent rather than GitHub native branch protection, we have more flexibility to change the rules without GitHub plan constraints.

### PR Pattern Analysis (Last 30 Merged PRs)

| Category | Count | % | Avg Additions |
|----------|-------|---|--------------|
| Docs-only (`docs:`) | 10 | 33% | ~400 lines |
| Feature PRs (`feat:`) | 17 | 57% | ~800 lines |
| Fix PRs (`fix:`) | 3 | 10% | ~200 lines |

**Key finding:** One-third of all PRs are docs-only. These PRs cannot affect Go code, tests, or build output. Running the full CI suite (Go tests, benchmarks, Docker E2E) on docs-only PRs is pure waste.

### CI Run Analysis (Last 40 Runs, ~1 Hour Window)

- **Total runs:** 40
- **PR-triggered:** 33 (82.5%)
- **Push-to-main:** 7 (17.5%)
- **Unique PR branches with multiple runs:** brave-lion (4), proud-squirrel (3), bright-badger (2), wise-rabbit (3), brave-rabbit (3)
- **Estimated churn runs:** ~20 of 33 PR runs (60%) were likely reruns from branch updates

**Cost estimate:** At ~3.5 min per PR run × 3 parallel jobs = ~10.5 job-minutes per run. 20 churn runs = ~210 wasted job-minutes per hour during active development.

## Options Evaluated

### Option 1: Path-Based CI Triggers (Quick Win)

**Description:** Add `paths` filter to CI workflow so docs-only changes skip Go-related jobs.

**Implementation:**
```yaml
on:
  pull_request:
    branches: [main]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - 'Makefile'
      - 'Dockerfile*'
      - 'docker-compose*'
      - '.github/workflows/**'
      - '.golangci.yml'
```

**Pros:**
- Eliminates ~33% of CI runs entirely (docs-only PRs)
- Zero risk — docs PRs genuinely cannot affect Go code
- 5-minute implementation
- GitHub-native, no tooling changes

**Cons:**
- Doesn't address churn between code PRs
- If branch protection requires the `quality-gate` check to pass, skipped workflows will block merge (need `paths-ignore` or a separate docs-only workflow that always passes, or reconfigure required checks)

**Risk:** Low. But must handle the "required check not present" problem — see implementation steps.

**Impact:** ~30% reduction in total CI runs.

### Option 2: Relax "Require Up-to-Date" Rule (High Impact)

**Description:** Allow merging PRs that pass CI on their branch tip, even if main has advanced since.

**Analysis for ThreeDoors specifically:**
- Go tests are fully isolated (no integration tests between PRs)
- No shared database or external service dependencies
- Task providers are interface-based — changes to one don't affect others
- Most PRs touch completely different files (docs vs different feature areas)
- The `push` trigger on main re-runs CI anyway, catching any merge breakage after the fact

**Pros:**
- Eliminates O(n²) churn entirely — each PR runs CI only when its code changes
- 80-90% reduction in total CI runs during parallel development
- Immediate effect, no CI pipeline changes needed

**Cons:**
- Theoretical risk of merge-time breakage (two PRs modify the same file differently)
- If breakage occurs, it's caught on the push-to-main run, but main is temporarily broken
- Requires cultural trust that post-merge CI catches issues

**Risk:** Low-medium for this project. The main risk scenario is two PRs editing the same Go file with conflicting changes. But:
1. Git merge conflicts would catch most of these
2. The push-to-main CI run catches the rest
3. In practice, our PRs rarely touch the same files (docs vs different features)

**Impact:** ~70-80% reduction in CI runs.

### Option 3: GitHub Native Merge Queue (Medium Impact)

**Description:** GitHub's built-in merge queue batches PRs and tests them together, preventing the "one merge invalidates all" cascade.

**How it works:**
- PRs enter a queue when approved
- GitHub creates a temporary merge branch combining queued PRs
- CI runs once on the combined branch
- If CI passes, all queued PRs merge together
- If CI fails, GitHub bisects to find the failing PR

**Pros:**
- Batches CI runs — 5 PRs might need only 1-2 CI runs instead of 5
- Prevents merge-time breakage (tests the combined state)
- GitHub-native, well-supported

**Cons:**
- Requires GitHub Team plan or higher for private repos (this repo is public, so available on Free plan)
- Adds latency — PRs wait in queue instead of merging immediately
- Configuration complexity (merge group sizes, timeout settings)
- Our multiclaude merge-queue agent would need adaptation to use `gh pr merge --merge-queue` instead of direct merge
- Merge queue creates its own `merge_group` events — CI workflow needs updating to trigger on these

**Risk:** Medium. Requires changes to both CI workflow and merge-queue agent.

**Impact:** ~60-70% reduction in CI runs, with guaranteed merge-time correctness.

### Option 4: Conditional Job Execution (Medium Impact)

**Description:** Use `dorny/paths-filter` or similar to conditionally skip irrelevant jobs within a single workflow run.

**Implementation:**
```yaml
jobs:
  changes:
    runs-on: ubuntu-latest
    outputs:
      go: ${{ steps.filter.outputs.go }}
      docker: ${{ steps.filter.outputs.docker }}
    steps:
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            go:
              - '**.go'
              - 'go.mod'
              - 'go.sum'
            docker:
              - 'Dockerfile*'
              - 'docker-compose*'

  quality-gate:
    needs: changes
    if: needs.changes.outputs.go == 'true'
    # ... existing steps

  benchmarks:
    needs: changes
    if: needs.changes.outputs.go == 'true'
    # ... existing steps

  test-docker-e2e:
    needs: changes
    if: needs.changes.outputs.docker == 'true' || needs.changes.outputs.go == 'true'
    # ... existing steps
```

**Pros:**
- Fine-grained control — skip benchmarks for non-Go changes, skip Docker E2E when only Go logic changes
- Still triggers the workflow (so required checks show up), but skips expensive jobs
- Can combine with Option 1 for layered optimization

**Cons:**
- Adds a `changes` job (~10 seconds) to every run
- Required checks that are skipped show as "Expected — Waiting for status to be reported" — may need status check configuration changes
- More complex workflow to maintain

**Risk:** Low-medium. The main risk is required check handling.

**Impact:** ~30-50% reduction in CI job-minutes (runs still trigger but skip expensive jobs).

### Option 5: CI Caching Improvements (Low Impact)

**Description:** Improve caching to reduce per-run cost even if run count stays the same.

**Current state:**
- Go module cache: Handled by `actions/setup-go` (already cached)
- Docker buildx cache: Already configured with local cache
- golangci-lint: Already cached by the action

**Potential improvements:**
- Cache `go build` output between runs
- Use GitHub-hosted larger runners for faster execution

**Pros:**
- Reduces wall-clock time per run
- Complements other strategies

**Cons:**
- Marginal gains — Go builds are already fast (~3.5 min total)
- Doesn't address the run count problem

**Risk:** Very low.

**Impact:** ~10-20% reduction in per-run time. Doesn't address the core O(n²) problem.

### Option 6: Separate Required vs Optional Checks (Quick Win)

**Description:** Make `benchmarks` and `test-docker-e2e` non-blocking for merge. Only `quality-gate` blocks.

**Pros:**
- Reduces merge friction — PRs can merge faster
- Benchmarks and Docker E2E still run but don't block
- Benchmark regressions caught on main push

**Cons:**
- Docker E2E failures could reach main
- Benchmark regressions not caught pre-merge

**Risk:** Low. Benchmarks are informational. Docker E2E tests overlap with the quality-gate unit tests.

**Impact:** Reduces merge latency, not run count. But complements other strategies.

## Recommendation (Ranked by Impact ÷ Effort)

### Tier 1: Implement Immediately (This Week)

1. **Relax "require up-to-date" in merge-queue agent** — If our merge-queue agent enforces the "up-to-date" requirement programmatically, modify it to skip that check. The push-to-main CI trigger provides a safety net. **Impact: ~70-80% reduction. Effort: ~30 min.**

2. **Add path filtering for docs-only PRs** — Add `paths` or `paths-ignore` to the CI trigger so docs-only PRs skip the entire CI suite. Add a lightweight `docs-check` job that always passes for docs PRs. **Impact: ~30% reduction. Effort: ~15 min.**

### Tier 2: Implement Soon (Next Sprint)

3. **GitHub Native Merge Queue** — Enable merge queue in repo settings and update the merge-queue agent to use `gh pr merge --merge-queue`. Update CI workflow to also trigger on `merge_group` events. **Impact: ~60-70% reduction with correctness guarantees. Effort: ~2-3 hours.**

4. **Conditional job execution** — Use `dorny/paths-filter` to skip benchmarks and Docker E2E for changes that don't affect them. **Impact: ~30-50% reduction in job-minutes. Effort: ~1 hour.**

### Tier 3: Nice to Have

5. **Make benchmarks and Docker E2E non-blocking** — Keep them running but don't require them for merge.

6. **CI caching improvements** — Marginal gains, lowest priority.

## Implementation Steps for Top Recommendation

### Step 1: Relax Up-to-Date Requirement

The merge-queue agent likely checks if a PR's branch is up-to-date with main before merging. To change this:

1. Find the merge-queue agent logic that enforces the up-to-date check
2. Remove or make configurable the `update-branch` call before merge
3. Rely on the push-to-main CI trigger as the safety net
4. If main CI fails after a merge, the merge-queue agent should alert and optionally revert

### Step 2: Add Path Filtering to CI

```yaml
on:
  pull_request:
    branches: [main]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - 'Makefile'
      - 'Dockerfile*'
      - 'docker-compose*'
      - '.github/workflows/**'
      - '.golangci.yml'
  push:
    branches: [main]
```

**Caveat:** If GitHub branch protection (or the merge-queue agent) requires the `quality-gate` check to be present on all PRs, we need a workaround:

**Option A:** Use `paths-ignore` instead of `paths` and list doc extensions:
```yaml
paths-ignore:
  - '**.md'
  - 'docs/**'
  - 'LICENSE'
  - '.gitignore'
```

**Option B:** Add a separate lightweight job that runs for docs-only PRs:
```yaml
  docs-pass:
    name: Quality Gate  # Same name as the real check
    if: # only runs when no Go files changed
    runs-on: ubuntu-latest
    steps:
      - run: echo "Docs-only PR, no code checks needed"
```

Option B is cleaner but requires the `dorny/paths-filter` approach to determine when to use it.

### Step 3: Enable GitHub Merge Queue (Tier 2)

1. Go to repo Settings → General → Pull Requests → Enable "Merge queue"
2. Configure merge group size (suggest: 3-5 PRs per batch)
3. Update CI workflow to trigger on `merge_group`:
   ```yaml
   on:
     pull_request:
       branches: [main]
     merge_group:
     push:
       branches: [main]
   ```
4. Update multiclaude merge-queue agent to use `gh pr merge --merge-queue` instead of direct merge
5. Test with a batch of docs PRs first

## Quick Wins Available Today

1. **Path filtering** — 15-minute CI workflow change, blocks docs PRs from running Go tests
2. **Relax up-to-date** — Modify merge-queue agent behavior (if it's our custom enforcement)
3. **Make benchmarks non-blocking** — Just remove from required checks (if applicable)

## Appendix: CI Run Data

### Sample Run Timing (2026-03-08, ~15:45-16:00 UTC)

```
Branch              Runs  Status     Duration
work/witty-hawk       2   in_progress  ~1.5 min
work/proud-squirrel   3   success      ~3.5 min each
work/brave-rabbit     3   success      ~3.5 min each
work/proud-lion       2   success      ~3.5 min each
work/bright-badger    2   success      ~3.5 min each
work/brave-lion       4   success      ~3.5-5 min each
work/wise-rabbit      3   success      ~3.5 min each
main (push)           7   success      ~6.5 min each
```

Total estimated CI minutes consumed in this 1-hour window: ~200 job-minutes.
Estimated waste from churn reruns: ~120 job-minutes (60%).

### PR Type Distribution (Last 30 Merged)

- 33% docs-only (would be eliminated by path filtering)
- 57% feature PRs (benefit from relaxed up-to-date or merge queue)
- 10% fix PRs (small, fast to test)
