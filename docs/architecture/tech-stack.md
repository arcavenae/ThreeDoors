# Tech Stack

## Cloud Infrastructure

**Provider:** None (Local-only application)

**Key Services:** N/A - No cloud services in Technical Demo phase

**Deployment Regions:** N/A - Runs locally on developer's macOS machine

**Future Consideration (Epic 2+):** If persistence/sync needs arise, consider iCloud for Apple Notes integration or local SQLite for enrichment data.

## Technology Stack Table

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| **Language** | Go | 1.25.4 | Primary development language | Excellent CLI tooling, fast compilation, simple deployment |
| **Runtime** | Go Runtime | 1.25.4 | Execution environment | Bundled with language; single binary output |
| **TUI Framework** | Bubbletea | 1.2.4 | Terminal UI framework | Elm architecture for reactive UIs, perfect for multi-view navigation |
| **TUI Styling** | Lipgloss | 1.0.0 | Terminal styling and layout | ANSI color support, box rendering for doors and detail views |
| **TUI Components** | Bubbles | 0.20.0 | Pre-built TUI components | Text input for notes, list selection for status menu |
| **Terminal Utilities** | golang.org/x/term | 0.26.0 | Terminal size detection & control | Responsive layout across terminal sizes |
| **YAML Parser** | gopkg.in/yaml.v3 | 3.0.1 | Structured data parsing | Human-readable task storage with metadata (status, notes, timestamps) |
| **UUID Generator** | github.com/google/uuid | 1.6.0 | Unique task IDs | UUID v4 generation for task identity |
| **Formatting** | gofumpt | 0.7.0 | Code formatting | Enforces consistent style |
| **Linting** | golangci-lint | 2.10.1 | Static analysis | Catches bugs early |
| **Build System** | Make | System default (macOS) | Build automation | `build`, `run`, `clean`, `lint`, `fmt` targets |
| **Dependency Mgmt** | Go Modules | 1.25.4 | Package management | `go.mod` versioning |
| **Testing** | Go testing package | 1.25.4 (stdlib) | Unit testing | Built-in; no external framework |
| **Storage** | YAML Files | N/A | Task persistence with metadata | `~/.threedoors/tasks.yaml` with status, notes, timestamps |
| **Platform** | macOS | 14+ (Sonoma+) | Target OS | Developer's primary platform |
| **Terminal** | iTerm2 / Terminal.app | Latest | Terminal emulator | 256-color support |
| **CI/CD** | GitHub Actions | N/A | Continuous integration & alpha release | Public runners, native Go support, quality gates |
| **Version Control** | Git | 2.40+ | Source control | github.com/arcaven/ThreeDoors.git |

## Post-MVP Technology Additions (Phase 2–3)

| Category | Technology | Version | Purpose | Rationale | Phase |
|----------|------------|---------|---------|-----------|-------|
| **Enrichment DB** | SQLite (via modernc.org/sqlite) | Latest | Local enrichment storage | Pure Go, no CGO, cross-references and metadata | 2 (Epic 6) |
| **Filesystem Watch** | fsnotify | 1.7+ | Detect external file changes | Obsidian vault watching, adapter change detection | 3 (Epic 8) |
| **Config Format** | YAML (config.yaml) | N/A | User configuration | Provider selection, vault paths, LLM config | 2+ |
| **AppleScript Bridge** | os/exec (stdlib) | N/A | Apple Notes integration | Invoke AppleScript for Notes read/write | 2 (Epic 2) |
| **Apple Notes DB** | database/sql (stdlib) | N/A | Direct SQLite read from NoteStore | Optional read-only path for Apple Notes | 2 (Epic 2) |
| **Markdown Parser** | goldmark or yuin/goldmark | 1.7+ | Parse Obsidian Markdown | Extract tasks from checkbox syntax | 3 (Epic 8) |
| **HTTP Client** | net/http (stdlib) | N/A | LLM API calls | Anthropic/OpenAI API, local Ollama | 4 (Epic 14) |
| **Calendar (AppleScript)** | os/exec (stdlib) | N/A | macOS Calendar.app reader | Local-first, no OAuth | 3 (Epic 12) |
| **Calendar (.ics)** | emersion/go-ical | Latest | Parse .ics calendar files | Standard iCalendar format support | 3 (Epic 12) |
| **Git Operations** | go-git or os/exec | Latest | LLM output to git repos | Write story specs for coding agents | 4 (Epic 14) |
| **Text Similarity** | agnivade/levenshtein | Latest | Duplicate detection | Cross-provider task dedup heuristics | 3 (Epic 13) |
| **Contract Testing** | Go testing (stdlib) | N/A | Adapter compliance validation | Verify TaskProvider implementations | 3 (Epic 9) |
| **TUI Forms** | charmbracelet/huh | 1.0.0 | Interactive form wizards | Setup wizard for data source connections (Charm ecosystem) | 2 (Epic 44) |

**Technology Selection Principles (Post-MVP):**

1. **Pure Go preferred:** Avoid CGO dependencies for easy cross-compilation (hence `modernc.org/sqlite` over `mattn/go-sqlite3`)
2. **Stdlib first:** Use standard library for HTTP, exec, SQL before adding external deps
3. **Local-first:** No cloud service dependencies at runtime; LLM backends are opt-in
4. **No OAuth:** Calendar integration uses only local sources (AppleScript, .ics files, CalDAV cache)
5. **Minimal new deps:** Each addition must justify itself; resist framework creep

## Makefile Targets

```makefile
.PHONY: build run clean fmt lint test

build:
	go build -o bin/threedoors cmd/threedoors/main.go

run: build
	./bin/threedoors

clean:
	rm -rf bin/

fmt:
	gofumpt -l -w .

lint:
	golangci-lint run ./...

test:
	go test -v ./...
```

---
