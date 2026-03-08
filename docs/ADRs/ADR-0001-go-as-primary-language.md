# ADR-0001: Go as Primary Language

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #6, #8

## Context

ThreeDoors needed a language for building a terminal-based task management application. Requirements included:
- Single binary distribution (no runtime dependencies)
- Fast compilation for rapid iteration
- Excellent CLI/TUI ecosystem
- Cross-compilation for macOS (primary) and potential Linux/Windows support
- Strong concurrency primitives for future sync engine work

## Decision

Use **Go** (1.25.4+) as the sole implementation language.

## Consequences

### Positive
- Single binary output simplifies distribution and installation
- Bubbletea/Lipgloss/Bubbles ecosystem provides best-in-class TUI support
- Fast compilation enables rapid development cycles
- Strong stdlib reduces external dependency count
- Goroutines and channels natural fit for sync engine concurrency
- `go test` provides built-in testing without framework dependencies

### Negative
- No generics initially (available since Go 1.18, used sparingly)
- Verbose error handling compared to exception-based languages
- Limited GUI options if desktop app is ever needed
- No sum types — status state machine uses string constants with validation

### Constraints
- Formatter: `gofumpt` (stricter than `gofmt`)
- Linter: `golangci-lint` with zero-warning policy
- Testing: stdlib `testing` package only — no testify
- Build: `make` for automation
