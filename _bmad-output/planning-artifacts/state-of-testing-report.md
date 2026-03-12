# State of Testing Report — ThreeDoors

> **Status: Not Yet Planned**
> Report merged (PR #349) but no epic or stories created from its recommendations yet.

**Date:** 2026-03-09
**Auditor:** TEA Agent (Test Engineering & Architecture)
**Method:** Automated analysis of all test files, coverage reports, race detector, and manual review of test patterns

---

## Executive Summary

ThreeDoors has a **strong testing foundation** with 198 test files containing 2,287 test functions across 24 packages. All tests pass, including the race detector. Test-to-source line ratio is 1.73:1 (61,136 test lines vs 35,246 source lines), indicating significant test investment. Coverage ranges from 0% (entry points) to 94% (reminders adapter), with most packages above 80%.

**Overall health: GOOD with specific gaps to address.**

---

## 1. Coverage Metrics

### Per-Package Coverage

| Package | Coverage | Assessment |
|---------|----------|------------|
| `cmd/threedoors` | 0.0% | Entry point — acceptable (thin main) |
| `cmd/threedoors-mcp` | 0.0% | Entry point — acceptable (thin main) |
| `internal/adapters` (root) | 76.1% | Contract suite — good |
| `internal/adapters/applenotes` | 86.1% | Strong |
| `internal/adapters/github` | 85.3% | Strong |
| `internal/adapters/jira` | 85.6% | Strong |
| `internal/adapters/obsidian` | 82.7% | Good |
| `internal/adapters/reminders` | 94.0% | Excellent |
| `internal/adapters/textfile` | 81.1% | Good |
| `internal/adapters/todoist` | 81.9% | Good |
| `internal/calendar` | 88.7% | Strong |
| `internal/ci` | [no statements] | Validation-only — N/A |
| `internal/cli` | **34.8%** | **CRITICAL GAP** |
| `internal/core` | 88.1% | Strong |
| `internal/core/metrics` | 91.2% | Excellent |
| `internal/dispatch` | 86.7% | Strong |
| `internal/dist` | 88.5% | Strong |
| `internal/enrichment` | 80.4% | Good |
| `internal/intelligence` | 93.1% | Excellent |
| `internal/intelligence/llm` | 89.6% | Strong |
| `internal/mcp` | 82.8% | Good |
| `internal/testkit` | 83.3% | Good |
| `internal/tui` | 80.8% | Good |
| `internal/tui/themes` | 91.5% | Excellent |

### Coverage Tiers

- **Excellent (≥90%):** 5 packages — reminders, core/metrics, intelligence, tui/themes, dist (partially)
- **Strong (80-89%):** 12 packages — most adapters, core, dispatch, calendar, enrichment, llm, tui
- **Good (70-79%):** 1 package — adapters root
- **Critical (<50%):** 1 package — `internal/cli` at 34.8%

---

## 2. Test Inventory

### By Type

| Type | Count | Notes |
|------|-------|-------|
| Unit tests | ~2,200 | Vast majority of test functions |
| Contract tests | 15 | Adapter contract compliance |
| E2E tests | 47 | TUI e2e via teatest + adapter e2e |
| Benchmark tests | 14 | Performance-critical paths |
| Golden file tests | 24 | TUI visual regression |
| Skipped tests | 9 | Scaffolds (3), platform-specific (4), env-conditional (2) |

### By Package (test function count, top packages)

| Package | Test Files | Approx Functions |
|---------|-----------|-----------------|
| `internal/core` | 40+ | ~600+ |
| `internal/tui` | 35+ | ~500+ |
| `internal/adapters/*` | 30+ | ~400+ |
| `internal/cli` | 11 | ~100+ |
| `internal/mcp` | 10 | ~120+ |
| `internal/dispatch` | 8 | ~80+ |
| Others | ~60 | ~400+ |

---

## 3. Quality Assessment

### Standards Compliance

| Standard | Adoption | Assessment |
|----------|----------|------------|
| Table-driven tests | 122 files (62%) | **Good** — widely adopted |
| `t.Parallel()` | 129 files (65%) | **Good** — extensive parallel test use |
| `t.Helper()` | 46 files (23%) | **Fair** — could be higher in helper-heavy packages |
| `t.Cleanup()` | 38 files (19%) | **Fair** — growing adoption |
| stdlib `testing` (no testify) | 100% | **Excellent** — zero testify imports |
| Golden file tests | 13 files | **Good** — all theme/TUI visual tests covered |
| Benchmarks | 14 functions | **Adequate** — covers critical paths |

### Race Detector

All 24 packages pass `go test -race ./...` cleanly. No data races detected. This is mandatory per CLAUDE.md for TUI/CLI packages and is being followed.

### Flaky Tests

No flaky tests detected in this run. One test (`obsidian_watcher_test.go`) is skipped on CI due to filesystem event timing unreliability — this is a known and correctly handled limitation.

---

## 4. Gap Analysis

### CRITICAL: `internal/cli` — 34.8% Coverage

This is the most significant gap. The CLI package has 0% coverage on:

- `bootstrap()` — application bootstrap/initialization
- `runConfigGet()`, `runConfigSet()`, `runConfigShow()` — config subcommands
- `loadTaskPool()` — core task loading path
- `runHealth()` — health check command
- `runMoodSet()`, `runMoodHistory()` — mood subcommands
- `Execute()`, `KnownSubcommands()` — root command functions
- `isJSONOutput()`, `isTerminal()` — utility functions
- `NewDoorsCmd()` — only 16.7% covered (main user-facing command)

**Risk: HIGH** — The CLI is the primary user entry point. Untested paths can break silently.

### IMPORTANT: Missing Contract Tests

Contract test suite covers 5 of 7 adapters:
- ✅ TextFile (full contract via `RunContractTests`)
- ✅ Apple Notes (contract subset — position-based IDs)
- ✅ Obsidian (contract tests)
- ✅ Reminders (contract tests)
- ✅ WAL provider (contract tests)
- ❌ **GitHub** — has 41 provider tests but no formal contract suite
- ❌ **Jira** — has provider tests but no formal contract suite
- ❌ **Todoist** — no contract tests at all (flagged previously)

Three scaffolded contract tests are skipped (calendar, remote, composite) — these are for future adapters per Epic 9.

**Risk: MEDIUM** — GitHub/Jira/Todoist adapters have individual tests but lack formal contract compliance verification. The Todoist gap is most concerning as it's actively used for bidirectional sync (Story 25.3).

### MODERATE: Entry Point Coverage

Both `cmd/threedoors` and `cmd/threedoors-mcp` show 0% coverage. They have tests (quit key, nil provider panic), but the main functions themselves are thin wrappers and typical to leave untested. Low risk.

### MODERATE: `t.Helper()` Adoption

Only 23% of test files use `t.Helper()`. Many test helper functions (especially in adapters and TUI packages) would benefit from adding `t.Helper()` to improve failure reporting.

### LOW: Benchmark Coverage

14 benchmark functions exist across:
- `internal/core` — task pool, sync engine, door selector
- `internal/adapters/textfile` — provider benchmarks

Missing benchmarks for:
- MCP server request handling
- TUI view rendering (important for responsiveness)
- Adapter load/save operations (GitHub, Jira, Todoist API simulation)

---

## 5. Testing Infrastructure

### Strengths

1. **Three-tier TUI testing** (ADR-0019) is fully implemented:
   - Headless teatest harness (576-line `e2e_test.go`)
   - Golden file snapshots in `testdata/` directories
   - Docker E2E framework (`Dockerfile.test`, `docker-compose.test.yml`)

2. **Contract test framework** (`internal/adapters/contract.go`) provides reusable test suite that any `TaskProvider` can run against

3. **Test factories** (`internal/testkit/`) provide consistent test data generation

4. **Testdata directories** properly organized in 6 locations

5. **CI pipeline** runs tests with race detector

### Weaknesses

1. **No integration tests** for multi-adapter scenarios (e.g., sync between Todoist and local, conflict resolution across real adapters)

2. **No load/stress tests** for concurrent multi-provider sync

3. **Docker E2E tests** exist but coverage of scenarios is unclear — should verify all user workflows are represented

---

## 6. Prioritized Recommendations

### P0 — Critical (address in next sprint)

1. **CLI package coverage** — Increase `internal/cli` from 34.8% to ≥70%
   - Priority functions: `NewDoorsCmd`, `loadTaskPool`, `bootstrap`, `Execute`
   - Config commands (`runConfigShow/Get/Set`)
   - Mood and health commands
   - Estimated: 1 story, ~400 lines of tests

2. **Todoist contract tests** — Add formal contract test suite
   - Todoist is the only actively-syncing adapter without contract tests
   - Especially critical given bidirectional sync (Story 25.3)
   - Note: Stories 25.4 and 30.4 exist in ROADMAP as "Not Started" for contract tests
   - Estimated: Part of Story 25.4

### P1 — Important (next 2 sprints)

3. **GitHub adapter contract tests** — Add formal `RunContractTests` integration
   - Has 41 tests but not verified against shared contract
   - Estimated: 0.5 story

4. **Jira adapter contract tests** — Add formal `RunContractTests` integration
   - Same situation as GitHub
   - Estimated: 0.5 story

5. **`t.Helper()` audit** — Add `t.Helper()` to all test helper functions
   - Low effort, high payoff for debugging test failures
   - Estimated: 1 PR, no story needed

### P2 — Nice to Have (backlog)

6. **TUI view benchmarks** — Add benchmarks for `View()` rendering in complex views
   - Ensures TUI responsiveness doesn't degrade
   - Estimated: 0.5 story

7. **Multi-adapter integration tests** — Test sync conflict resolution with real (simulated) adapter pairs
   - Currently only unit-tested
   - Estimated: 1 story

8. **Docker E2E scenario expansion** — Audit and expand Docker E2E test scenarios
   - Verify all primary user workflows are covered
   - Estimated: 1 story

---

## 7. Story Outlines for Critical Gaps

### Story: CLI Test Coverage Hardening

**Epic:** Testing / Quality
**Priority:** P0
**Acceptance Criteria:**
- `internal/cli` coverage ≥ 70%
- All config subcommands (`show`, `get`, `set`) have unit tests
- `NewDoorsCmd` path tested including `loadTaskPool` integration
- `bootstrap()` tested with mock dependencies
- Health, mood, and stats commands have table-driven tests
- All tests pass with `-race`

### Story: Todoist Contract Test Suite (Story 25.4)

**Epic:** 25 — Todoist Integration
**Priority:** P0
**Acceptance Criteria:**
- Todoist adapter runs against `RunContractTests` contract suite
- Bidirectional sync paths have contract-level validation
- Mock HTTP client for API simulation (no real API calls in tests)
- Field mapping edge cases covered (priority, due dates, labels)

### Story: GitHub/Jira Contract Test Alignment (Story 30.4)

**Epic:** 30 — GitHub/Jira Integration
**Priority:** P1
**Acceptance Criteria:**
- GitHub adapter runs against `RunContractTests`
- Jira adapter runs against `RunContractTests`
- Read-only contract behaviors properly handled
- Mock HTTP clients for both adapters

---

## 8. Summary Statistics

| Metric | Value |
|--------|-------|
| Total test files | 198 |
| Total test functions | 2,287 |
| Total test lines | 61,136 |
| Source lines | 35,246 |
| Test:Source ratio | 1.73:1 |
| Packages with tests | 24/24 (100%) |
| Packages ≥80% coverage | 17/24 (71%) |
| Packages <50% coverage | 1/24 (4%) — `internal/cli` |
| Race detector | ✅ All clean |
| Flaky tests | 0 detected |
| Skipped tests | 9 (all justified) |
| Contract test coverage | 5/7 adapters |
| Testify usage | 0 (compliant) |

---

*Report generated by TEA agent audit. Next review recommended after Stories 25.4 and 30.4 are implemented.*
