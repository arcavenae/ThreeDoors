# ADR-0028: CI Quality Gates and Testing Strategy

- **Status:** Accepted
- **Date:** 2025-11-09 (initial), evolved through 2026-03-08
- **Decision Makers:** Design decisions H10, M14, L8
- **Related PRs:** #8, #9, #10, #41, #89, #96, #102, #106, #107, #133, #142
- **Related ADRs:** ADR-0019 (Docker E2E Testing)

## Context

As a project built primarily by AI agents, strict quality gates are essential to prevent regression and maintain code quality. The CI pipeline must catch issues that individual workers might miss.

## Decision

Implement **multi-layer CI quality gates** in GitHub Actions:

### Gate 1: Formatting and Linting
- `gofumpt` — stricter than `gofmt`, zero-tolerance
- `golangci-lint` — zero warnings policy
- No `//nolint` directives without justifying comments

### Gate 2: Testing
- `go test ./... -v` — all tests must pass
- `go test -race ./...` — race detector enabled
- Table-driven tests required for functions with >2 cases

### Gate 3: Coverage
- 70% global minimum (Decision M14)
- 0% regression tolerance — coverage must not decrease
- Coverage report uploaded to CI artifacts

### Gate 4: Performance
- Benchmark local adapter operations against <100ms threshold (Decision H10)
- Network adapters have separate SLA: <2s for API calls
- Performance benchmarks run in CI (Story 9.3)

### Gate 5: E2E
- Docker-based E2E tests for reproducible TUI testing
- Golden file validation for all themes
- Input sequence replay tests for user workflows

### No Emergency Overrides (Decision L8)
Quality gates cannot be bypassed. If a fix is critical, it can still pass lint and tests.

## Consequences

### Positive
- Consistent quality regardless of which AI agent wrote the code
- Regressions caught before merge
- Performance degradation detected early
- Visual regressions caught by golden files

### Negative
- CI pipeline takes 3-5 minutes per run
- Golden file updates required for intentional visual changes
- Coverage threshold can block legitimate changes that reduce covered code
