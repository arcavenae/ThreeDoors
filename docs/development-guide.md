# ThreeDoors Development Guide

## Prerequisites

| Requirement | Version | Installation |
|---|---|---|
| Go | 1.25.4+ | `asdf install golang 1.25.4` or [golang.org](https://golang.org/doc/install) |
| Git | 2.40+ | System package manager |
| Make | System default | Pre-installed on macOS |
| gofumpt | 0.7.0 | `go install mvdan.cc/gofumpt@latest` |
| golangci-lint | 1.61.0 | `brew install golangci-lint` or [docs](https://golangci-lint.run/usage/install/) |

## Environment Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/arcaven/ThreeDoors.git
   cd ThreeDoors
   ```

2. Verify Go version:
   ```bash
   go version  # Should show go1.25.4 or higher
   ```

3. No environment variables or `.env` files required — this is a local-only application.

4. Application data directory: `~/.threedoors/` (created automatically on first run)

## Build & Run Commands

| Command | Description |
|---|---|
| `make build` | Compile to `bin/threedoors` |
| `make run` | Build and run |
| `make test` | Run all tests (`go test -v ./...`) |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofumpt |
| `make clean` | Remove build artifacts (`bin/`) |

## Coding Standards

### Formatting & Linting
- **Always** run `gofumpt -l -w .` before committing
- **Always** ensure `golangci-lint run ./...` passes with zero warnings
- Import ordering: stdlib > external > internal (auto-handled by gofumpt)

### Naming Conventions
- Packages: lowercase single word (`tui`, `tasks`)
- Files: lowercase snake_case (`task_pool.go`, `doors_view.go`)
- Exported types/functions: PascalCase (`TaskPool`, `NewTaskPool`)
- Private types/functions: camelCase (`validateTask`, `renderDoor`)

### Critical Rules
1. No `fmt.Println` in TUI code — output via Bubbletea `View()` only
2. All file writes use atomic write pattern (write .tmp, sync, rename)
3. Always validate status transitions via `StatusManager` before applying
4. Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
5. No panics in user-facing code
6. Task IDs are immutable after creation
7. Timestamps always in UTC (`time.Now().UTC()`)
8. YAML field tags must match schema exactly

## Testing Approach

### Strategy
- **Pragmatic testing** — focus on domain logic, minimal TUI testing
- **Table-driven tests** for multiple scenarios
- **No mocking frameworks** — use interfaces and simple stubs

### Coverage Goals
| Package | Target |
|---|---|
| `internal/tasks/*` | 70%+ |
| `internal/tui/*` | 20%+ |
| Overall | 50%+ |

### Test Pyramid
- 70% Unit tests (fast, isolated)
- 20% Integration tests (component interactions, file I/O via `t.TempDir()`)
- 10% Manual testing (end-to-end TUI workflows)

### Running Tests
```bash
go test ./...           # All tests
go test ./internal/tasks/...  # Domain tests only
go test -cover ./...    # With coverage report
go test -v ./...        # Verbose output
```

## Deployment

- **Strategy:** Direct binary distribution (no cloud infrastructure)
- **Build:** `make build` compiles to `bin/threedoors`
- **Install:** Copy binary to `/usr/local/bin/` or run via `make run`
- **CI/CD:** None configured for Technical Demo phase (deferred to Epic 2)
- **Future:** Homebrew tap planned (`brew install arcaven/tap/threedoors`)

## Data Compatibility

- YAML schema must remain backward compatible
- Forward migrations add fields with defaults
- Never break existing `tasks.yaml` format
