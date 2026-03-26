# Contributing to ThreeDoors

Thank you for your interest in contributing to ThreeDoors! This guide will help you get started.

## Development Model

ThreeDoors uses a **solo maintainer + AI agent team** development model. The human maintainer (arcaven) directs a team of AI agents that handle most implementation work. All PRs are reviewed by the maintainer before merge.

You don't need to use AI agents to contribute — just follow this guide and submit a PR like any other open-source project.

**All work is story-driven.** Every code change should have a corresponding story file in `docs/stories/`. If you're working on something new, create a story file as part of your PR. See existing story files for the format.

## Prerequisites

- **Go 1.25.4+** — [install](https://go.dev/doc/install)
- **make** — for build commands
- **gofumpt** — `go install mvdan.cc/gofumpt@latest` (stricter than gofmt)
- **golangci-lint** — [install](https://golangci-lint.run/welcome/install/)

## Getting Started

```bash
git clone https://github.com/arcavenae/ThreeDoors.git
cd ThreeDoors
make build    # Build the binary
make test     # Run tests
make lint     # Run linter (must pass with zero warnings)
make fmt      # Format code with gofumpt
```

## How to Contribute

### Reporting Bugs

Open a [Bug Report](https://github.com/arcavenae/ThreeDoors/issues/new?template=bug-report.yml) and include:

- ThreeDoors version (`threedoors --version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior

### Suggesting Features

Open a [Feature Request](https://github.com/arcavenae/ThreeDoors/issues/new?template=feature-request.yml).

**Please read [SOUL.md](SOUL.md) first.** ThreeDoors is opinionated by design — "three doors, not three hundred." Features that add complexity need strong justification. Features that conflict with the project philosophy will be declined.

### Submitting Code

1. Fork the repo and create a feature branch (`git checkout -b feat/your-feature`)
2. Ensure a story file exists in `docs/stories/` (or create one as part of your PR)
3. Write tests — table-driven, using stdlib `testing` only (no testify)
4. Run the full quality gate:
   ```bash
   make fmt
   make lint
   make test
   ```
5. Run the race detector for TUI or CLI changes:
   ```bash
   go test -race ./internal/tui/... ./internal/cli/...
   ```
6. Create a PR using the PR template
7. Update your story file status to `Done (PR #NNN)`

### Commit Message Format

```
feat|fix|docs|refactor: description (Story X.Y)
```

Examples:
- `feat: add keyboard shortcut display (Story 39.1)`
- `fix: prevent crash on empty task file (Story 12.3)`
- `docs: update installation guide (Story 0.43)`

## Code Standards

ThreeDoors follows strict Go coding standards. Key rules:

- **Formatting:** gofumpt (not gofmt) — run `make fmt`
- **Linting:** golangci-lint with zero warnings — run `make lint`
- **Testing:** Table-driven tests using stdlib `testing`. Use `t.Helper()`, `t.Cleanup()`, and `t.Parallel()` where appropriate.
- **Errors:** Always handle errors. Wrap with `%w` for context. Define sentinel errors as package-level vars.
- **Context:** `context.Context` is always the first parameter.
- **Timestamps:** Always use `time.Now().UTC()`.
- **No `init()` functions** — pass dependencies explicitly.
- **No panics** in user-facing code — Bubbletea `Update()` and `View()` must never panic.

See the "Go Quality Rules" section in [CLAUDE.md](CLAUDE.md) for the complete list.

## What NOT to Contribute

- Features that conflict with [SOUL.md](SOUL.md) philosophy (local-first, privacy-always, minimal complexity)
- Heavy dependencies — prefer the standard library
- Telemetry, analytics, or phone-home features
- Testify or other test framework dependencies
- Cloud sync or account features (unless the user explicitly opts in)

## Architecture Overview

```
cmd/threedoors/       # Entry point (main.go)
internal/tasks/       # Task domain: models, providers, persistence, analytics
internal/tui/         # Bubbletea views and UI components
docs/                 # Architecture docs, stories, PRD
```

The key abstraction is the `TaskProvider` interface (`internal/tasks/provider.go`). New storage backends are added by implementing this interface.

- **TUI framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)
- **Data format:** YAML task files, JSONL session logs
- **All TUI output** goes through Bubbletea `View()` methods — never `fmt.Println`
- **File persistence** uses atomic writes (write to `.tmp`, sync, rename)

## License

ThreeDoors is [MIT licensed](LICENSE). By contributing, you agree that your contributions will be licensed under the same license.
