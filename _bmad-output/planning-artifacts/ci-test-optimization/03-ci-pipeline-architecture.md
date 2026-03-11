# Party Mode Session 3: CI Pipeline Architecture Improvements

**Participants:** Architect, TEA/QA, Dev, PM, SM
**Date:** 2026-03-11
**Topic:** Optimize CI pipeline structure without compromising integrity

---

## Current CI Architecture

```
                    ┌──────────────┐
                    │   Trigger    │
                    │  (PR/push)   │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │   Detect     │
                    │   Changes    │  ~4s
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
      ┌───────▼──────┐ ┌──▼──────────┐ ┌▼────────────┐
      │ Quality Gate │ │ Docker E2E  │ │ Benchmarks  │
      │   2m08s      │ │   2m51s     │ │   3m33s     │
      └───────┬──────┘ └─────────────┘ └─────────────┘
              │ (push only)
      ┌───────▼──────┐
      │Build Binaries│  1m11s
      └───────┬──────┘
      ┌───────▼──────┐
      │Sign/Notarize │  2m21s  (conditional)
      └───────┬──────┘
      ┌───────▼──────┐
      │   Release    │  45s
      └──────────────┘
```

**PR critical path:** max(Quality Gate, Docker E2E, Benchmarks) = **3m33s** (Benchmarks)
**Push critical path:** 3m33s + Build + Sign + Release = **~7m48s**

---

## Optimization Opportunities

### 1. Eliminate Docker E2E Redundancy

**Status:** HIGHEST PRIORITY — eliminates redundant CI work

**Problem:** The Docker E2E job runs `go test ./... -v -count=1 -timeout 5m` inside a Docker container. The Quality Gate job runs `go test ./... -v -count=1 -race -coverprofile=coverage.out -covermode=atomic`. These overlap almost completely.

**What Docker E2E adds beyond Quality Gate:**
- Tests run in a Linux container (validates cross-platform compat)
- Tests run without race detector (different code paths for `-race` builds)
- Tests run with CGO_ENABLED=0 (pure Go build)

**What Docker E2E lacks vs Quality Gate:**
- No race detection
- No coverage reporting
- No linting

**Proposed change:** Docker E2E should only run tests that specifically need a Docker environment (e.g., testing the Docker build itself, or testing with specific OS behaviors). Currently it's a full test suite duplicate.

**Options:**

**Option A: Remove Docker E2E entirely** (recommended for PRs)
- Quality Gate already validates all tests with race detection
- Docker E2E only catches cross-platform issues, which are extremely rare for a TUI app
- Keep Docker E2E for push-to-main only (defense in depth without blocking PRs)

**Option B: Docker E2E runs only golden file tests**
- Subset the Docker E2E to only validate golden file output matches
- This is the primary cross-platform concern (terminal rendering differences)

**Option C: Docker E2E runs on a schedule instead of per-PR**
- Weekly cron job catches environment regressions without per-PR cost

**Consensus: Option A** — Move Docker E2E to push-only. PRs get Quality Gate + Benchmarks.

**Expected impact:** No wall-clock change for PRs (jobs are parallel), but reduces CI runner cost by ~2.5 minutes per PR. More importantly, eliminates a source of flaky failures that block PRs.

### 2. Conditional Benchmark Execution

**Status:** MEDIUM PRIORITY — reduces PR critical path

**Problem:** Benchmarks are the longest PR job at 3m33s. They run on every code-change PR, even when the changed code can't affect performance (e.g., adding a new adapter, changing TUI rendering).

**Proposed change:** Add path filtering for benchmarks:

```yaml
benchmarks:
  name: Performance Benchmarks
  needs: changes
  if: |
    github.event_name == 'push' ||
    (needs.changes.outputs.code == 'true' && needs.changes.outputs.perf == 'true')
  # ...

# In changes job, add:
perf:
  - 'internal/core/**'
  - 'internal/adapters/textfile/**'
  - 'go.mod'
```

**Expected impact:** PRs that don't touch core/textfile skip benchmarks, reducing PR wall clock from 3m33s → 2m08s (Quality Gate becomes the bottleneck).

**Risk:** A change outside core/textfile could regress performance. Mitigated by always running benchmarks on push-to-main.

**Rejected alternative:** Running benchmarks only on push-to-main. This delays performance regression detection — the Brownian ratchet demands catching issues at PR time when possible.

**Consensus: Implement with path filtering.** Push-to-main always runs them (safety net).

### 3. Split Quality Gate into Parallel Steps

**Status:** LOW PRIORITY — marginal gain for increased complexity

**Current Quality Gate steps (sequential):**
1. Checkout + setup-go: ~20s
2. Install gofumpt: ~5s
3. Check formatting: ~2s
4. go vet: ~5s
5. golangci-lint: ~25s
6. go test with race + coverage: ~90s
7. Coverage summary: ~2s
8. Coverage floor check: ~1s
9. Coverage comment: ~3s
10. Build validation: ~5s

**Opportunity:** Split into two parallel jobs:
- **Lint job:** Steps 1-5 (~55s)
- **Test job:** Steps 1, 6-10 (~120s)

**Expected impact:** Quality Gate from ~2m08s → ~2m00s. Marginal because test step dominates and both need checkout+setup.

**Rejected:** The go setup cache means step 1 is fast when cached. Splitting adds job orchestration overhead and complicates the required-checks configuration. Not worth it.

### 4. Incremental Linting (`--new-from-rev`)

**Status:** REJECTED

**Concept:** Only lint changed files: `golangci-lint run --new-from-rev=origin/main`

**Why rejected:**
- A change in one file can cause linting issues in another (e.g., changing an interface)
- Brownian ratchet philosophy: if it was clean before, it should stay clean
- Risk of accumulating hidden lint debt
- The full lint is only ~25s — not worth the risk

### 5. Smarter Go Module Caching

**Status:** LOW EFFORT — verify current setup

**Question:** Does `actions/setup-go@v6` cache both `~/go/pkg/mod` and `~/.cache/go-build`?

**Answer:** `setup-go@v6` caches module downloads by default but the build cache behavior depends on configuration. Adding explicit build cache can save 10-20s on recompilation:

```yaml
- uses: actions/setup-go@v6
  with:
    go-version-file: 'go.mod'
    cache-dependency-path: go.sum
```

The `cache-dependency-path` triggers the built-in cache. Verify this is set.

### 6. Parallel NFR13 Validation

**Status:** LOW PRIORITY

**Problem:** NFR13 validation runs in both Quality Gate (as part of `go test ./...`) and Benchmarks (as `go test -run='NFR13'`). This is deliberate — Quality Gate validates correctness, Benchmarks validate timing.

**Opportunity:** The NFR13 test in Quality Gate doesn't need to run the large-scale (500 task) variant. Guard it with `-short`:

```go
func TestAdapterReadWriteNFR13(t *testing.T) {
    // ... small tests always run ...
    t.Run("write/large_(500_tasks)", func(t *testing.T) {
        if testing.Short() {
            t.Skip("large-scale NFR13 test skipped in short mode")
        }
        // ...
    })
}
```

Then in Quality Gate: `go test ./... -short -race -coverprofile=...`

**Rejected:** Using `-short` in CI risks missing bugs. The 1.7s for the large variant isn't worth the risk.

---

## Proposed CI Architecture (After Optimizations)

```
                    ┌──────────────┐
                    │   Trigger    │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │   Detect     │
                    │   Changes    │  ~4s
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │ (if perf paths changed)
      ┌───────▼──────┐    │     ┌──────▼──────┐
      │ Quality Gate │    │     │ Benchmarks  │
      │   ~1m30s*    │    │     │   3m33s     │
      └───────┬──────┘    │     └─────────────┘
              │           │
              │    ┌──────▼──────┐
              │    │ Docker E2E  │  (push-only)
              │    │   2m51s     │
              │    └─────────────┘
              │ (push only)
      ┌───────▼──────┐
      │Build Binaries│
      └──────────────┘
      ... (release pipeline unchanged)
```

*Quality Gate reduced from 2m08s → ~1m30s after TUI sleep optimization (Session 2).

**PR critical path (non-perf changes):** ~1m30s (down from 3m33s — **57% reduction**)
**PR critical path (perf changes):** ~3m33s (unchanged, benchmarks dominate)
**Push critical path:** Unchanged (all jobs run)
