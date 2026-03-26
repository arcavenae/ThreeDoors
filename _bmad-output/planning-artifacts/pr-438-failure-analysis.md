# PR #438 Failure Analysis — Story 42.5 CI Supply Chain Hardening

**PR:** https://github.com/arcavenae/ThreeDoors/pull/438
**Branch:** `work/kind-rabbit`
**Date:** 2026-03-10
**Analyst:** lively-eagle (worker)

## Summary

PR #438 has TWO failing CI checks, both caused by a **pre-existing race condition** in `internal/core/config_paths_test.go` — NOT by any change in PR #438 itself.

## Failing Checks

| Check | Status | Root Cause |
|-------|--------|------------|
| Quality Gate | FAILURE | Race condition in `SetHomeDir()` tests |
| Performance Benchmarks | SUCCESS (in run 22930053037) | N/A — actually passing |
| Docker E2E Tests | FAILURE | Same race condition (Docker runs full test suite) |

**Note:** The task description said "Quality Gate AND Performance Benchmarks" fail. In the most recent completed run (22930053037), Performance Benchmarks actually PASSED. Quality Gate and Docker E2E Tests failed.

## Root Cause: Race Condition in `SetHomeDir()` (Pre-existing Bug)

### The Bug

`internal/core/config_paths.go:15-19` has a package-level `var testHomeDir string` with an unsynchronized `SetHomeDir()` setter:

```go
var testHomeDir string

func SetHomeDir(dir string) {
    testHomeDir = dir  // Line 19 — unsynchronized write to package-level var
}
```

### Why It Races

Three tests in `config_paths_test.go` all use `t.Parallel()` and call `SetHomeDir()`:
- `TestEnsureConfigDir_CreatesWithRestrictivePermissions` (line 10-12)
- `TestEnsureConfigDir_MigratesPermissiveDirectory` (line 33-37)
- `TestEnsureConfigDir_NoChangeWhenAlreadyRestrictive` (line 76-80)

Each test:
1. Calls `SetHomeDir(tmpDir)` at start
2. Registers `t.Cleanup(func() { SetHomeDir("") })` to reset

When run in parallel, goroutines write to `testHomeDir` concurrently → DATA RACE.

### Race Detector Output (from CI)

```
WARNING: DATA RACE
Write at 0x000000df5360 by goroutine 96:
  github.com/arcavenae/ThreeDoors/internal/core.SetHomeDir()
      config_paths.go:19 +0xb2
  TestEnsureConfigDir_MigratesPermissiveDirectory()
      config_paths_test.go:36 +0x55

Previous write at 0x000000df5360 by goroutine 97:
  github.com/arcavenae/ThreeDoors/internal/core.SetHomeDir()
      config_paths.go:19 +0x78
  TestEnsureConfigDir_NoChangeWhenAlreadyRestrictive.func1()
      config_paths_test.go:80 +0x12
```

### Origin

This code was introduced in **Story 42.1 — File Permission Standardization (PR #437)**, merged recently. The race was latent in the test design. Main CI passed because the race detector is non-deterministic — it sometimes catches races and sometimes doesn't.

## Classification

**Pre-existing code bug (fixable)** — NOT caused by PR #438's changes.

PR #438 only modifies:
- `.github/workflows/ci.yml` — SHA-pinning third-party actions + adding govulncheck step
- `.github/workflows/release.yml` — SHA-pinning goreleaser action
- `docs/stories/42.5.story.md` — Status update

None of these changes touch Go source code or tests.

## Fix

The race in `config_paths_test.go` needs to be fixed. Two approaches:

### Option A: Remove `t.Parallel()` from these three tests (simplest)
These tests share mutable package-level state (`testHomeDir`). They cannot safely run in parallel. Remove `t.Parallel()` from all three.

### Option B: Refactor to accept homeDir as parameter (better design)
Change `EnsureConfigDir()` to accept a homeDir parameter (or use a struct with methods) so tests don't need shared mutable state. This is more invasive but eliminates the design problem.

### Recommended: Option A for immediate fix
Remove `t.Parallel()` from the three `TestEnsureConfigDir_*` tests. This is the minimal fix that resolves the race without refactoring the production API. A separate story could address the design if desired.

## Recommendation

1. **Fix the race condition** in a separate PR (it's a bug in Story 42.1, not Story 42.5)
2. **Rebase PR #438** after the race fix merges
3. **Retry CI** — PR #438's changes are correct and should pass once the race is fixed

Alternatively, the race fix could be included directly in PR #438 as a drive-by fix, since it's blocking this PR's CI. But the cleaner approach is a separate fix PR.

## Current Run Status

Run 22930276093 is currently IN_PROGRESS — may produce same results depending on race detector timing.

## govulncheck Status

The govulncheck step (new in this PR) was NOT reached in the failed run because the test step failed first. Its functionality remains unverified.
