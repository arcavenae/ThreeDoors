# Party Mode Session 2: Go-Specific Test Optimization Techniques

**Participants:** Architect, TEA/QA, Dev, PM, SM
**Date:** 2026-03-11
**Topic:** Go test optimization strategies applicable to ThreeDoors

---

## Current State Assessment

ThreeDoors already applies many Go test best practices:
- **t.Parallel()**: 2,720 uses across 287 test files — near-maximum intra-package parallelism
- **Table-driven tests**: Extensively used per CLAUDE.md mandate
- **t.Cleanup()**: Used instead of defer
- **No testify dependency**: Stdlib testing only (good for compile speed)
- **Test fixtures in testdata/**: Standard layout

What follows are the gaps and opportunities.

---

## Optimization Techniques

### 1. Replace `time.Sleep` with Condition-Based Waits

**Status:** HIGH IMPACT — directly addresses the #1 bottleneck

**Problem:** TUI E2E tests use `time.Sleep(100-200ms)` between each simulated keystroke. With 58+ calls, this adds ~8.7s of pure waiting to the TUI package alone.

**Solution:** Bubbletea's teatest package supports `WaitFor(condition, timeout)` which polls for a condition at configurable intervals (default 10ms). Replace:

```go
// BEFORE: Fixed 200ms sleep regardless of actual render time
tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
time.Sleep(200 * time.Millisecond)

// AFTER: Wait for actual state change (typically resolves in 10-50ms)
tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
    return strings.Contains(string(bts), "Door 1 selected")
}, teatest.WithDuration(2*time.Second))
```

**Expected impact:** TUI E2E tests from ~33s → ~10-15s (50-70% reduction)

**Risk:** WaitFor requires knowing what output to expect. Some tests check intermediate states that may flash briefly. Need careful migration — test by test, not batch.

**Rejected alternative:** Reducing sleep durations (e.g., 50ms). This trades speed for flakiness — violates the Brownian ratchet.

### 2. Mock Clock for Time-Dependent Tests

**Status:** MEDIUM IMPACT — addresses oauth and connection bottlenecks

**Problem:** OAuth polling tests and TestRollingRetention use real `time.Sleep` to simulate time passage.

**Solution:** Inject a clock interface:

```go
type Clock interface {
    Now() time.Time
    Sleep(d time.Duration)
    After(d time.Duration) <-chan time.Time
}

// Production: uses real time
type realClock struct{}

// Test: advances instantly
type mockClock struct {
    mu  sync.Mutex
    now time.Time
}
```

**Expected impact:** OAuth tests from ~8.3s → ~0.5s. Connection tests from ~8s → ~0.5s. Total: ~15s saved.

**Risk:** Adds an interface to production code. Must be opt-in (default to real clock).

**Rejected alternative:** `go-clock` library. A little copying is better than a little dependency (Go proverb). The interface is 3 methods.

### 3. Package-Level Parallelism (`-p` flag)

**Status:** ALREADY OPTIMIZED — Go defaults to GOMAXPROCS packages in parallel

**Current behavior:** Go already runs packages in parallel (default `-p` = GOMAXPROCS). On CI (2-core ubuntu-latest), this means 2 packages compile/test simultaneously. On local dev (8-core M1), 8 packages at once.

**Opportunity:** The 96s cumulative time runs in ~33s wall clock locally because packages overlap. The long pole is the TUI package at 33s — no amount of package parallelism helps when one package dominates.

**Conclusion:** No action needed on `-p` flag. Focus on reducing the TUI package duration.

### 4. Build and Test Caching

**Status:** PARTIALLY LEVERAGED

**Current state:**
- Go's build cache is enabled by default (`GOCACHE`)
- CI uses `actions/setup-go@v6` which caches the module download but NOT the build cache
- `-count=1` (used in CI) bypasses test result caching — this is **correct** for CI (ensures fresh runs)

**Opportunity:** Cache the Go build cache in CI to speed up compilation:

```yaml
- uses: actions/cache@v5
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

**Expected impact:** Saves ~15-30s on compilation across jobs. `setup-go@v6` may already handle module cache, but build cache is often missed.

**Note:** `setup-go@v6` has built-in caching. Verify it caches both module and build artifacts.

### 5. Test Binary Caching

**Status:** NOT USED — moderate opportunity

**Problem:** Each CI job that runs `go test` rebuilds all test binaries from scratch.

**Solution:** Pre-build test binaries and reuse:

```bash
# Build all test binaries
go test -c ./internal/tui -o /tmp/test-bins/tui.test
go test -c ./internal/core -o /tmp/test-bins/core.test

# Run from pre-built binary
/tmp/test-bins/tui.test -test.v -test.count=1 -test.race
```

**Expected impact:** Minimal if build cache is warm. More useful for splitting tests across jobs.

**Rejected:** Complexity doesn't justify the gain for a 30-package project.

### 6. `-short` Mode for Fast Feedback

**Status:** NOT USED — good opportunity for local dev

**Problem:** Developers running `make test` locally wait for the full suite including slow E2E and timeout tests.

**Solution:** Add `testing.Short()` guards to expensive tests:

```go
func TestRollingRetention(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow retention test in short mode")
    }
    // ... 8 second test
}
```

Add to Makefile:
```makefile
test-fast:  ## Quick test pass (skips slow tests)
    go test ./... -short -count=1

test:       ## Full test suite
    go test ./... -v -count=1
```

**Expected impact:** Local dev feedback loop from ~33s → ~10s. CI still runs full suite.

**Consensus:** Good ergonomic improvement. Threshold: skip tests that take >2s individually.

### 7. Subtests with Shared Setup

**Status:** PARTIALLY USED

**Observation:** Some test files create expensive fixtures (temp dirs, large YAML files) per test function. Where multiple tests share setup, `t.Run` subtests with a shared parent save setup time.

**Example:** `TestAdapterReadWriteNFR13` creates 500-task files for each sub-case. A single fixture shared across subtests could save ~1s.

### 8. Build Tags for Test Categories

**Status:** NOT USED — useful for CI splitting

**Concept:** Tag slow tests so they can be included/excluded:

```go
//go:build slow
package tui

func TestWorkflow_MultipleRerolls(t *testing.T) { ... }
```

```bash
# Fast suite (excludes slow)
go test ./... -count=1

# Full suite (includes slow)
go test -tags=slow ./... -count=1
```

**Rejected by consensus:** Adds complexity. The `-short` flag achieves the same goal without build tags. Build tags are better suited for platform-specific tests, not speed tiers.

---

## Prioritized Technique Summary

| # | Technique | Impact | Effort | Recommend? |
|---|-----------|--------|--------|------------|
| 1 | Replace time.Sleep with WaitFor in TUI tests | HIGH (~18s saved) | MEDIUM | **YES** |
| 2 | Mock clock for oauth/connection tests | MEDIUM (~15s saved) | MEDIUM | **YES** |
| 3 | `-short` mode for local dev | MEDIUM (local DX) | LOW | **YES** |
| 4 | Verify/improve Go build caching in CI | LOW-MEDIUM (~15-30s) | LOW | **YES** |
| 5 | Package parallelism (`-p`) | NONE (already optimal) | N/A | No |
| 6 | Test binary caching | LOW | HIGH | No |
| 7 | Shared test fixtures via subtests | LOW (~1-2s) | LOW | Maybe |
| 8 | Build tags for test categories | LOW | MEDIUM | No |
