# ADR-0019: Docker E2E and Headless TUI Testing

- **Status:** Accepted
- **Date:** 2026-02-15
- **Decision Makers:** Architecture review
- **Related PRs:** #60, #64 (Story 18.1), #86 (Story 18.2), #104 (Story 18.4), #105 (Story 18.3), #107 (Story 18.5)

## Context

TUI applications are notoriously difficult to test. Standard unit tests can verify logic but not visual rendering, keyboard interaction sequences, or terminal-specific behavior. ThreeDoors needed comprehensive testing beyond unit tests.

## Decision

Implement a **three-tier TUI testing strategy**:

1. **Headless TUI tests** (teatest) — Test Bubbletea models without a real terminal
2. **Golden file snapshots** — Capture expected TUI output and compare against future renders
3. **Docker E2E tests** — Full application tests in reproducible containerized environments

## Rationale

- `teatest` enables testing Bubbletea models programmatically with simulated input
- Golden files catch visual regressions (layout changes, styling breaks)
- Docker containers provide reproducible terminal environments (consistent `TERM`, dimensions, locale)
- Three tiers catch different categories of bugs

## Implementation

- **teatest harness** (Story 18.1) — Programmatic TUI interaction testing
- **Golden file tests** (Story 18.2) — `testdata/*.golden` files for all themes and views
- **Input sequence replay** (Story 18.3) — Test multi-step user workflows
- **Docker environment** (Story 18.4) — `Dockerfile.test` with controlled terminal settings
- **CI integration** (Story 18.5) — Docker tests run in GitHub Actions pipeline

## Consequences

### Positive
- Visual regressions caught automatically
- Theme changes validated across all views
- User workflow sequences tested end-to-end
- Reproducible test environment eliminates "works on my machine" issues

### Negative
- Golden file updates required when intentional visual changes are made
- Docker tests are slower than unit tests (~30s vs ~2s)
- Container image must be rebuilt when Go version or dependencies change
- Maintaining golden files adds review burden on visual PRs
