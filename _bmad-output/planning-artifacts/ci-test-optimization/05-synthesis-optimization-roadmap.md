# Party Mode Session 5 (Synthesis): Prioritized CI/Test Optimization Roadmap

**Participants:** Architect, TEA/QA, Dev, PM, SM
**Date:** 2026-03-11
**Topic:** Synthesize all findings into a prioritized optimization roadmap

---

## Executive Summary

ThreeDoors' CI pipeline runs at **3m33s per PR** and **~7m48s per push-to-main**. The test suite (30 packages, 1,167 tests, 287 test files) is already well-optimized with extensive use of `t.Parallel()` (2,720 calls). The primary bottlenecks are:

1. **Sleep-based test synchronization** in TUI E2E and OAuth tests (~30s wasted on `time.Sleep`)
2. **Redundant Docker E2E job** duplicating Quality Gate's test execution
3. **Benchmarks running on every PR** regardless of whether performance-sensitive code changed

With the proposed optimizations, **PR CI time can be reduced from 3m33s to ~1m30s** (58% reduction) for non-performance PRs, while maintaining or improving test reliability.

---

## Optimization Roadmap

### Phase 1: Quick Wins (No Code Changes) — Expected: 1-2 PRs

| # | Optimization | Impact | Risk | Effort |
|---|-------------|--------|------|--------|
| 1.1 | **Move Docker E2E to push-only** | Reduces runner cost ~2.5min/PR, eliminates flaky surface area | Very low — still runs on push | ~10 lines of CI yaml |
| 1.2 | **Add path filtering for benchmarks** | PR wall clock: 3m33s → 2m08s for non-perf PRs | Low — push-to-main still runs all | ~15 lines of CI yaml |
| 1.3 | **Fix golangci-lint version skew** in Dockerfile.test (v2.1.6 → v2.10.1) | Correctness fix | None | 1 line |
| 1.4 | **Verify Go build cache** in `setup-go@v6` | ~10-20s per CI job | None | Check/add `cache-dependency-path` |
| 1.5 | **Add `make test-fast`** for local dev (`-short` mode) | Better local DX | None | ~3 lines in Makefile |

**Phase 1 expected PR wall clock:** 3m33s → **2m08s** (40% reduction)

### Phase 2: TUI Test Acceleration — Expected: 2-3 PRs

| # | Optimization | Impact | Risk | Effort |
|---|-------------|--------|------|--------|
| 2.1 | **Replace `time.Sleep` with `teatest.WaitFor`** in TUI E2E tests | TUI package: 33s → ~12s | Medium — requires careful per-test migration | ~58 sleep sites |
| 2.2 | **Add `-short` guards** to slow TUI workflow tests (>2s) | Quality Gate faster with `-short` | Low | ~5 tests to guard |
| 2.3 | **Verify race detector** passes after each WaitFor migration | Ensures no new races | None | Part of migration |

**Phase 2 expected Quality Gate:** 2m08s → **~1m30s** (30% further reduction)

### Phase 3: Time-Dependent Test Refactoring — Expected: 2-3 PRs

| # | Optimization | Impact | Risk | Effort |
|---|-------------|--------|------|--------|
| 3.1 | **Introduce clock interface** for OAuth polling tests | OAuth: 8.3s → ~0.5s | Medium — changes production code interface | New interface + mock |
| 3.2 | **Mock clock for TestRollingRetention** | Connection: 8s → ~0.5s | Medium | Clock injection |
| 3.3 | **Reduce Linear network timeout test** to use faster mock | Linear: 2s → ~0.1s | Low | Mock the timeout |

**Phase 3 expected total test time:** 96s → **~70s cumulative**

### Phase 4: Nice-to-Haves — Opportunistic

| # | Optimization | Impact | Risk | Effort |
|---|-------------|--------|------|--------|
| 4.1 | **Slim Docker image** (alpine or distroless) | ~30% faster Docker build | Low | Dockerfile change |
| 4.2 | **Strip tools from Dockerfile.test** | ~15s faster build | None | Remove 2 installs |
| 4.3 | **Shared test fixtures** in NFR13 tests | ~1-2s saved | Low | Refactor subtests |
| 4.4 | **TestMain for expensive one-time setup** | Variable | Low | Per-package eval |

---

## Impact Summary

| Metric | Current | After Phase 1 | After Phase 2 | After Phase 3 |
|--------|---------|---------------|---------------|---------------|
| **PR wall clock (non-perf)** | 3m33s | **2m08s** | **~1m30s** | ~1m20s |
| **PR wall clock (perf)** | 3m33s | 3m33s | 3m33s | 3m33s |
| **Quality Gate** | 2m08s | 2m08s | **~1m30s** | ~1m15s |
| **Cumulative test time** | 96s | 96s | ~75s | ~70s |
| **CI runner minutes/PR** | ~9min* | ~5.5min | ~4min | ~3.5min |
| **Local `make test`** | ~33s | ~33s | ~15s | ~12s |
| **Local `make test-fast`** | N/A | **~10s** | ~8s | ~6s |

*CI runner minutes = sum of all parallel job durations

---

## Constraints Verified

All optimizations maintain:
- **Race detector**: Always runs in Quality Gate for TUI/CLI packages (MANDATORY per CLAUDE.md)
- **Coverage floor**: 75% threshold unchanged
- **TDD discipline**: No tests removed, no assertions weakened
- **Brownian ratchet**: CI passing = shippable; no CI weakening
- **Full test suite on push-to-main**: All jobs run on merge (defense in depth)
- **Benchmark validation**: NFR13 <100ms threshold always checked

---

## Rejected Approaches (with rationale)

| Approach | Why Rejected |
|----------|-------------|
| Incremental linting (`--new-from-rev`) | Can miss cross-file issues; 25s full lint is acceptable |
| Build tags for test categories | `-short` flag achieves same goal with less complexity |
| Test binary caching | Negligible benefit with warm build cache |
| Removing Docker E2E entirely | Defense-in-depth value on push-to-main justifies keeping |
| Running tests in CI with `-short` | Risks missing bugs; `-short` is for local dev only |
| Splitting Quality Gate into parallel jobs | Marginal gain (~8s) for added orchestration complexity |
| Pre-built Docker images for tests | Complexity not justified for project size |
| Running benchmarks only on push-to-main | Delays performance regression detection too much |

---

## Implementation Notes

### Story Candidates

Each phase maps to 1-2 stories:

- **Phase 1:** "Optimize CI pipeline configuration" — CI yaml changes + Makefile
- **Phase 2:** "Replace sleep-based TUI E2E synchronization with WaitFor" — test code changes
- **Phase 3:** "Introduce clock abstraction for time-dependent tests" — interface + test changes

### Risk Mitigation

- **Phase 2 is highest risk**: Each `time.Sleep` → `WaitFor` migration must be validated individually with race detector. Budget for debugging 2-3 subtle races.
- **Phase 3 changes production code**: The clock interface must be opt-in (default to real clock) and must not change any public API signatures.

### Measurement Plan

Before each phase, record:
1. CI wall clock time (from GitHub Actions run page)
2. Quality Gate duration (from job timing)
3. Local `make test` duration
4. Per-package test times (`go test -json` analysis)

After each phase, compare. If any metric regresses, investigate before proceeding.

---

## Final Consensus

**Phase 1 is a no-brainer** — pure CI config changes with zero code risk. Implement immediately.

**Phase 2 is the highest-value code change** — the TUI sleep pattern is both the biggest time sink and a reliability risk (sleeps can be too short on slow CI runners). WaitFor is strictly better.

**Phase 3 is a design decision** — introducing a clock interface is a small but permanent addition to the production codebase. Worth it for the 15s savings and improved test isolation, but should go through story planning.

**Phase 4 is opportunistic** — do these when touching the relevant files for other reasons.
