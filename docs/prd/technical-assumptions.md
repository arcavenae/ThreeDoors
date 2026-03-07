# Technical Assumptions

## Technical Demo Phase Architecture

**Decision:** Minimal monolithic application with simple text file I/O

**Rationale:**
- **Speed to validation**: Build and test in 4-8 hours
- **Simple is fast**: No database, no complex integrations, no abstractions until needed
- **Easy external task population**: Text files can be edited with any editor, populated from scripts, etc.
- **Prove the concept first**: Validate Three Doors UX before investing in infrastructure
- **Low risk**: Can throw away and rebuild if concept fails validation

**Tech Demo Structure:**
```
ThreeDoors/
├── cmd/
│   └── threedoors/        # Main application (single file initially)
├── internal/
│   ├── tui/              # Bubbletea Three Doors interface
│   └── tasks/            # Simple file I/O (read tasks.txt, write completed.txt)
├── docs/                  # Documentation (including this PRD)
├── .bmad-core/           # BMAD methodology artifacts
├── Makefile              # Simple build: build, run, clean
└── README.md             # Quick start guide
```

**Data Files (created at runtime in `~/.threedoors/`):**
```
~/.threedoors/
├── tasks.txt             # One task per line (user can edit directly)
├── completed.txt         # Completed tasks with timestamps
└── config.txt            # Optional: Simple key=value config (if needed)
```

---

## Full MVP Architecture (Post-Validation - Deferred)

**Structure evolves to:**
```
ThreeDoors/
├── cmd/                    # CLI entry points
│   └── threedoors/        # Main application
├── internal/              # Private application code
│   ├── core/             # Core domain logic
│   ├── tui/              # Bubbletea interface components
│   ├── integrations/     # Adapter implementations
│   │   ├── textfile/    # Text file backend (from Tech Demo)
│   │   └── applenotes/  # Apple Notes integration
│   ├── enrichment/       # Local enrichment storage
│   └── learning/         # Door selection & pattern tracking
├── pkg/                   # Public, reusable packages (if any)
├── docs/                  # Documentation (including this PRD)
├── .bmad-core/           # BMAD methodology artifacts
└── Makefile              # Build automation
```

## Service Architecture

**Technical Demo Phase:**

**Decision:** Single-layer CLI/TUI application with direct file I/O

**Rationale:**
- **No abstractions yet**: Build for one thing (text files), refactor when adding second thing
- **Validate UX first**: Door selection algorithm is the innovation, not the data layer
- **Fast iteration**: Change anything without navigating architecture layers

**Demo Architecture:**
- **TUI Layer (Bubbletea)** - Three Doors interface, keyboard handling, rendering
- **Direct File I/O** - Read tasks.txt, write completed.txt, no abstraction layer
- **Simple Door Selection** - Random selection of 3 tasks from available pool (no learning/categorization yet)

---

**Full MVP Phase (Post-Validation - Deferred):**

**Decision:** Layered architecture with pluggable integration adapters

**Architecture Layers:**
1. **TUI Layer (Bubbletea)** - User interaction, rendering, keyboard handling
2. **Core Domain Logic** - Task management, door selection algorithm, progress tracking
3. **Integration Adapters** - Abstract interface with concrete implementations (text file, Apple Notes, others later)
4. **Enrichment Storage** - Metadata, cross-references, learning patterns not stored in source systems
5. **Configuration & State** - User preferences, values/goals, application state

**Key Architectural Principles:**
- Core domain logic has NO dependencies on specific integrations (dependency inversion)
- Integrations implement common `TaskProvider` interface
- Enrichment layer wraps tasks from any source with additional metadata
- TUI layer depends only on core domain, not specific integrations

## Testing Requirements

**Technical Demo Phase:**

**Decision:** Manual testing only - validate UX through real use

**Demo Testing Approach:**
- **No automated tests for Tech Demo** - premature given throwaway prototype nature
- **Manual testing** via daily use for 1 week
- **Success measurement**: Does Three Doors feel better than a list? Yes/No decision point
- **Quality gate**: If it crashes or feels bad to use, iterate or abandon concept

**Rationale:**
- 4-8 hours to build entire demo - testing infrastructure would consume half that time
- Real usage is the test: if developer won't use it daily, concept fails regardless of test coverage
- Can add tests when/if proceeding to Full MVP

---

**Full MVP Phase (Post-Validation - Deferred):**

**Testing Scope:**
- **Unit tests** for core domain logic (door selection algorithm, categorization, progress tracking)
- **Integration tests** for backend adapters (text file, Apple Notes)
- **Manual testing** for TUI interactions (Bubbletea testing framework is immature)

**Test Coverage Goals:**
- Core domain logic: 70%+ coverage (pragmatic, not perfectionist)
- Integration adapters: Critical paths covered (read, write, sync scenarios)
- TUI layer: Manual testing via developer use

**Testing Strategy:**
- Table-driven tests (idiomatic Go pattern)
- Test fixtures for data structures
- Mock `TaskProvider` interface for testing core logic without real integrations
- CI/CD runs tests on every commit (GitHub Actions)

**Deferred for Post-MVP:**
- End-to-end testing framework
- Property-based testing for door selection algorithm
- Performance/load testing

## Additional Technical Assumptions and Requests

**Technical Demo Phase Assumptions:**

**Text File Format:**
- **Simple line-delimited format**: One task per line in `tasks.txt`
- **Completed format**: `[timestamp] task description` in `completed.txt`
- **No metadata yet**: Task is just text; no categories, priorities, or context for Tech Demo
- **Easy population**: User can edit files with any text editor, generate from scripts, copy-paste, etc.

**Door Selection Algorithm (Tech Demo):**
- **Random selection**: Pick 3 random tasks from available pool
- **Simple diversity**: Ensure no duplicates in the three doors
- **No intelligence yet**: No learning, no categorization, no context awareness
- **Validation goal**: Prove that having 3 options reduces friction vs. scrolling a full list

**File I/O:**
- **Go standard library**: Use `os`, `bufio`, `io` - no external dependencies for file operations
- **Error handling**: Create files with defaults if missing; graceful degradation if corrupted
- **Concurrency**: Not a concern for single-user local files

---

**Full MVP Phase Assumptions (Post-Validation - Deferred):**

**Apple Notes Integration:**
- **Options Identified (2025):**
  1. **DarwinKit (github.com/progrium/darwinkit)** - Native macOS API access from Go; requires translating Objective-C patterns; full API control but higher complexity
  2. **Direct SQLite Database Access** - Apple Notes stores data in `~/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite`; note content is gzipped protocol buffers in `ZICNOTEDATA.ZDATA` column; read-only safe, write risks corruption
  3. **AppleScript Bridge** - Use `os/exec` to invoke AppleScript; simpler than native APIs; proven approach (see `sballin/alfred-search-notes-app`)
  4. **Existing MCP Server** - `mcp-apple-notes` server exists for Apple Notes integration; could potentially leverage this instead of building from scratch
- **Assumption:** Multiple viable paths exist; choice depends on read-only vs. read-write needs, complexity tolerance, and reliability requirements (WILL REQUIRE VALIDATION when implementing Phase 2)
- **Spike Required:** Evaluate options before implementing Apple Notes integration
- **Preferred Exploration Order:** Start with Option 4 (MCP server) or Option 2 (SQLite read-only), fall back to Option 3 (AppleScript) if bidirectional sync required, reserve Option 1 (DarwinKit) for complex scenarios

**Cloud Storage for Cross-Computer Sync (DEFERRED - Not MVP):**
- **Status:** Cross-computer sync is deferred post-MVP; single-computer local storage is sufficient for initial development and use
- **Future Exploration:** When implementing sync, explore alternatives to monolithic SQLite file:
  - Individual JSON/YAML files per task or per day (more granular, better suited for file-based cloud sync)
  - Conflict-free Replicated Data Types (CRDTs) for eventual consistency
  - Event sourcing with append-only logs
  - Cloud-native solutions (S3, Firebase, etc.) if local-first constraint relaxes
- **Awareness:** Monolithic SQLite on cloud storage (iCloud/Google Drive) is known problematic—corruption risk, locking issues, slow sync
- **MVP Decision:** Store enrichment data locally only; revisit sync architecture when/if multi-computer use becomes actual need

**Go Language & Ecosystem (Tech Demo):**
- **Language:** Go 1.25.4+ (current stable as of November 2025)
- **Formatting:** `gofumpt` (run before commits)
- **Linting:** Skip for Tech Demo (adds no validation value at this stage)
- **Dependency Management:** Go modules
- **TUI Framework:** Bubbletea + Lipgloss (styling) - minimal Bubbles components, only if needed

**Data Storage (Tech Demo):**
- **Storage:** Plain text files in `~/.threedoors/`
- **No database**: Not needed for line-delimited text
- **No configuration file initially**: Hardcode paths, add config only if needed

**Build & Development (Tech Demo):**
- **Build System:** Minimal Makefile
  ```makefile
  build:
      go build -o bin/threedoors cmd/threedoors/main.go

  run: build
      ./bin/threedoors

  clean:
      rm -rf bin/
  ```
- **Development Workflow:** Direct iteration on macOS
- **No CI/CD for Tech Demo**: Overkill for validation prototype

**Performance Expectations (Tech Demo):**
- **File I/O**: <10ms to read tasks.txt (even with 100+ tasks)
- **Door selection**: <1ms for random selection from array
- **TUI rendering**: Bubbletea handles 60fps, not a concern
- **Startup time**: <100ms total from launch to Three Doors display

**Security & Privacy (Tech Demo):**
- **Local files only**: No network, no external services
- **No logging**: Not even metadata for Tech Demo
- **File permissions**: Standard user file permissions on `~/.threedoors/`

---

**Full MVP Phase (Post-Validation - Deferred):**

**Go Language & Ecosystem:**
- **Language:** Go 1.25.4+
- **Formatting:** `gofumpt`
- **Linting:** `golangci-lint` with standard rule set
- **Dependency Management:** Go modules
- **TUI Framework:** Bubbletea + Lipgloss + Bubbles

**Data Storage:**
- **Primary:** Apple Notes (user-facing tasks) or text file backend
- **Enrichment:** SQLite for metadata (door feedback, blockers, categorization, learning patterns)
- **Configuration:** YAML or TOML for user preferences, values/goals
- **Location:** `~/.config/threedoors/` (XDG Base Directory spec on Linux, macOS equivalent)

**Build & Development:**
- **Build System:** Makefile with full targets (build, test, lint, install)
- **CI/CD:** GitHub Actions running tests on every commit
- **Development Workflow:** Direct iteration on macOS

**Performance Expectations:**
- Door selection algorithm: <100ms to choose 3 tasks from up to 1000 total tasks
- Backend sync: <2 seconds for typical data set
- TUI rendering: 60fps equivalent for smooth interaction

**Deferred Technical Decisions (Post-MVP):**
- Cross-computer sync architecture (see deferred section above)
- LLM provider integration architecture (local vs. cloud, which providers)
- Additional integration adapters (Jira, Linear, Google Calendar, etc.)
- Remote access agent for Geodesic environments
- Vector database for semantic task search
- Voice interface integration

---
