# Party Mode Session 1: Test Suite Profiling and Bottleneck Identification

**Participants:** Architect, TEA/QA, Dev, PM, SM
**Date:** 2026-03-11
**Topic:** Profile the ThreeDoors test suite, identify slow tests and bottlenecks

---

## Baseline Metrics

### Package-Level Execution Times (cumulative, `go test ./... -count=1`)

| Package | Time (s) | % of Total | Notes |
|---------|----------|------------|-------|
| `internal/tui` | 33.02 | 34.3% | **#1 bottleneck** — E2E teatest tests |
| `internal/core/connection/oauth` | 8.29 | 8.6% | Sleep-based polling tests |
| `internal/core/connection` | 7.99 | 8.3% | TestRollingRetention: 8.14s |
| `internal/adapters/textfile` | 5.89 | 6.1% | NFR13 benchmark-style tests |
| `internal/core` | 4.20 | 4.4% | Core domain logic |
| `internal/adapters/linear` | 3.74 | 3.9% | Network timeout test |
| `internal/adapters/applenotes` | 3.51 | 3.6% | |
| `internal/cli` | 3.02 | 3.1% | |
| `internal/adapters/obsidian` | 2.66 | 2.8% | |
| Other (21 packages) | 23.83 | 24.8% | All < 2.2s each |
| **Total** | **96.15** | **100%** | **30 packages, 1167 top-level tests** |

### Slowest Individual Tests

| Test | Time (s) | Package | Root Cause |
|------|----------|---------|------------|
| TestWorkflow_MultipleRerolls | 9.01 | tui | Multiple teatest cycles + sleep waits |
| TestRollingRetention | 8.14 | connection | time.Sleep-based retention window simulation |
| TestPollForToken_SlowDown | 7.00 | oauth | Simulated 7s slow-down backoff |
| TestAdapterReadWriteNFR13 | 3.43 | textfile | Large dataset (500 tasks) write benchmark |
| TestE2E_MoodTracking_AllOptions | 3.32 | tui | 6 subtests × 0.55s each (sequential) |
| TestWorkflow_RerollDoors | 3.00 | tui | Multiple reroll cycles |
| TestE2E_FullSession_MultipleActions | 2.05 | tui | Multi-step interaction sequence |
| TestLinearClientNetworkTimeout | 2.00 | linear | 2s timeout wait |
| TestPollForToken_Timeout | 2.00 | oauth | 2s timeout wait |
| TestPollForToken_GitHubStyleOKWithError | 2.00 | oauth | 2s simulated delay |

### Test Infrastructure Stats

- **Total test files:** 287 (290 `*_test.go` files)
- **Total top-level tests:** 1,167
- **`t.Parallel()` usage:** 2,720 calls — **extensively parallelized already**
- **`TestMain` usage:** 1 file only (`proposals_view_test.go`)
- **`time.Sleep` in tests:** 58+ calls in TUI e2e_test.go alone

### CI Pipeline Timing (actual code PR run)

| Job | Duration | Runs In Parallel? |
|-----|----------|-------------------|
| Detect Changes | 4s | First (gate) |
| Quality Gate | **2m08s** | Yes (with below) |
| Docker E2E Tests | **2m51s** | Yes |
| Performance Benchmarks | **3m33s** | Yes |
| Build Binaries | 1m11s | After Quality Gate |
| Sign & Notarize | 2m21s | After Build |
| Create Release | 45s | After Sign |

**PR wall clock time: ~3m37s** (limited by Benchmarks, the slowest parallel job)
**Push-to-main wall clock: ~6m40s** (includes post-merge release pipeline)

---

## Bottleneck Analysis

### Architect's Assessment

The test suite's 96s cumulative time isn't extreme for a project of this size. The real concern is the CI wall clock — 3.5 minutes for PRs. The bottleneck hierarchy:

1. **TUI E2E tests (33s)**: Sleep-based timing synchronization. Each `time.Sleep(200ms)` is a hard floor that can't be parallelized away within a single test. With 58+ sleep calls averaging 150ms, that's ~8.7s of pure waiting in e2e_test.go alone.

2. **OAuth polling tests (8.3s)**: Realistic polling simulation with real `time.Sleep`. TestPollForToken_SlowDown alone sleeps 7s.

3. **Connection retention test (8.1s)**: TestRollingRetention uses real wall-clock timing for retention window testing.

4. **Benchmark-in-test pattern (3.4s)**: NFR13 validation tests that run large-scale operations belong in the benchmark job, not quality-gate.

### TEA/QA Assessment

The test parallelism is excellent — 2,720 `t.Parallel()` calls across 287 files. However, parallelism only helps at the package level (tests within a package share a single binary). The TUI package is the long pole precisely because its E2E tests are inherently sequential (each simulates a user interaction timeline).

The Docker E2E job is **fully redundant** — it runs the exact same `go test ./...` as quality-gate, just inside a container. Unless we're catching environment-specific bugs, this duplicates ~2 minutes of CI time.

### Dev Assessment

Key code-level observations:
1. The `time.Sleep` pattern in TUI tests is fragile and slow. Teatest v2 supports `WaitFor` helpers that poll for conditions instead of sleeping.
2. OAuth polling tests use real sleeps to simulate server timing. These could use a clock abstraction.
3. The `TestRollingRetention` test actually needs real time passage for its rolling window logic — but 8 seconds is excessive for what's being validated.

### PM Assessment

3.5 minutes per PR is acceptable for a small team. The real cost is developer flow disruption. Priorities should be:
1. Things that are free (no code changes): CI config optimizations
2. Things that improve reliability AND speed: replacing sleeps with condition-based waits
3. Things that require careful trade-offs: splitting test suites, conditional execution

### SM Assessment

Current velocity impact: developers wait ~3.5 minutes per push. With TDD discipline requiring red-green cycles, the CI wait is multiplied. Sprint impact is real but not critical. Focus on quick wins first.

---

## Consensus: Top Bottlenecks to Address

1. **Docker E2E redundancy** — Free win, ~2.5 min saved on CI runners (not wall clock since parallel, but reduces resource usage and flaky surface area)
2. **TUI sleep-based E2E tests** — Highest impact on quality-gate duration
3. **OAuth sleep-based polling tests** — 7-8s of avoidable wall-clock time
4. **Connection rolling retention test** — 8s for a test that could use a mock clock
5. **Benchmark-in-test (NFR13)** — Already runs in benchmark job, may be duplicated in quality-gate
