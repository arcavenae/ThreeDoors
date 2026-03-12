# Source Tree

```
ThreeDoors/
├── cmd/
│   └── threedoors/
│       └── main.go                    # Application entry point, Bubbletea initialization
│
├── internal/                          # Private application code
│   ├── tui/                          # TUI Layer - Bubbletea components
│   │   ├── main_model.go            # Root Bubbletea model, view routing
│   │   ├── doors_view.go            # Three Doors display component
│   │   ├── task_detail_view.go      # Task detail and options component
│   │   ├── status_menu.go           # Status update menu subcomponent
│   │   ├── notes_input.go           # Notes text input subcomponent
│   │   ├── blocker_input.go         # Blocker input subcomponent
│   │   ├── styles.go                # Lipgloss style definitions
│   │   └── messages.go              # Bubbletea message types
│   │
│   └── tasks/                        # Domain Layer - Business logic
│       ├── task.go                  # Task model, methods, validation
│       ├── task_status.go           # TaskStatus enum, constants
│       ├── task_pool.go             # TaskPool collection manager
│       ├── door_selection.go        # DoorSelection model, algorithm
│       ├── door_selector.go         # Door selection logic
│       ├── status_manager.go        # Status transition validator
│       ├── file_manager.go          # YAML I/O, atomic writes
│       └── config.go                # Configuration model, defaults
│
├── docs/                             # Documentation
│   ├── prd.md                       # Product Requirements Document
│   ├── architecture.md              # This architecture document
│   └── stories/                     # Story breakdowns (from PRD)
│
├── .bmad-core/                       # BMAD methodology artifacts
│   ├── core-config.yaml
│   ├── agents/
│   ├── tasks/
│   ├── templates/
│   └── data/
│
├── .github/                          # GitHub configuration (Epic 2+)
│   └── workflows/                   # CI/CD pipelines (deferred)
│
├── bin/                              # Build output (gitignored)
│   └── threedoors                   # Compiled binary
│
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
├── Makefile                          # Build automation
├── .gitignore                        # Git ignore rules
└── README.md                         # Quick start guide

User Data Directory (created at runtime):
~/.threedoors/
├── tasks.yaml                        # Active tasks with metadata
└── completed.txt                     # Completed task log
```

**Key Organization Principles:**

1. **`cmd/` for entry points:** Single main.go bootstraps the application
2. **`internal/` for private code:** Cannot be imported by external projects
3. **`internal/tui/` for presentation:** All Bubbletea UI components
4. **`internal/tasks/` for domain:** Business logic, no UI dependencies
5. **Flat package structure:** No deep nesting (2 levels max)
6. **Clear separation:** TUI layer imports tasks, never vice versa

---

## Post-MVP Source Tree (Phase 2–3)

As the architecture evolves, the source tree expands to accommodate adapters, sync engine, intelligence layer, and multi-source aggregation.

```
ThreeDoors/
├── cmd/
│   └── threedoors/
│       └── main.go                    # Application entry point
│
├── internal/                          # Private application code
│   ├── tui/                          # TUI Layer - Bubbletea components
│   │   ├── main_model.go            # Root model, view routing
│   │   ├── doors_view.go            # Three Doors display + source badges
│   │   ├── task_detail_view.go      # Task detail and options
│   │   ├── status_menu.go           # Status update menu
│   │   ├── notes_input.go           # Notes text input
│   │   ├── blocker_input.go         # Blocker input
│   │   ├── onboarding_view.go       # First-run onboarding wizard (Epic 10)
│   │   ├── sync_status_bar.go       # Per-provider sync status (Epic 11)
│   │   ├── conflict_view.go         # Sync conflict visualization (Epic 11)
│   │   ├── source_badge.go          # Provider attribution badges (Epic 13)
│   │   ├── decompose_view.go        # LLM decomposition results (Epic 14)
│   │   ├── sources_view.go          # Sources dashboard — connection list/status (Epic 44)
│   │   ├── source_detail_view.go   # Source detail: status, sync log, actions (Epic 44)
│   │   ├── connect_wizard.go        # Setup wizard using huh forms (Epic 44)
│   │   ├── disconnect_dialog.go    # Disconnect confirmation with task preservation (Epic 44)
│   │   ├── synclog_detail_view.go  # Sync event log viewer per connection (Epic 44)
│   │   ├── import_view.go          # :import command — bulk task import from text
│   │   ├── bug_report.go           # Bug report environment collection (Epic 50)
│   │   ├── bug_report_view.go      # Bug report TUI view (Epic 50)
│   │   ├── styles.go                # Lipgloss style definitions
│   │   └── messages.go              # Bubbletea message types
│   │
│   ├── tui/themes/                   # Door Themes — visual theming engine (Epic 48)
│   │   ├── theme.go                 # Theme interface, DoorTheme struct
│   │   ├── registry.go              # Theme registry, lookup by name
│   │   ├── anatomy.go               # Door anatomy: frame, panel, handle definitions
│   │   ├── crack.go                 # Crack of light selection effect
│   │   ├── shadow.go                # Door shadow rendering
│   │   ├── seasonal.go              # Seasonal theme auto-selection
│   │   ├── classic.go               # Classic theme
│   │   ├── modern.go                # Modern theme
│   │   ├── scifi.go                 # Sci-fi theme
│   │   ├── shoji.go                 # Shoji (Japanese) theme
│   │   ├── autumn.go                # Autumn seasonal theme
│   │   ├── winter.go                # Winter seasonal theme
│   │   ├── spring.go                # Spring seasonal theme
│   │   └── summer.go                # Summer seasonal theme
│   │
│   ├── cli/                          # CLI Layer - Cobra commands (Epic 45+)
│   │   ├── root.go                  # Root cobra command, global flags
│   │   ├── bootstrap.go             # CLI bootstrap: config loading, provider init
│   │   ├── doctor.go                # `threedoors doctor` diagnostics (Epic 49)
│   │   ├── sources.go               # `threedoors sources` list/status/test/log
│   │   ├── sources_manage.go        # `threedoors sources` pause/resume/sync/disconnect/reauth/edit
│   │   ├── sources_log.go           # `threedoors sources log` subcommand
│   │   ├── connect.go               # `threedoors connect` provider setup
│   │   ├── task.go                  # `threedoors task` CRUD commands
│   │   ├── doors.go                 # `threedoors doors` display command
│   │   ├── stats.go                 # `threedoors stats` analytics display
│   │   ├── config.go                # `threedoors config` management
│   │   ├── health.go                # `threedoors health` checks
│   │   ├── plan.go                  # `threedoors plan` planning mode
│   │   ├── mood.go                  # `threedoors mood` tracking
│   │   ├── version.go               # `threedoors version` info
│   │   ├── output.go                # Shared output formatting (JSON/text)
│   │   └── exitcodes.go             # Standardized exit codes
│   │
│   ├── core/                         # Core Domain (Phase 2+)
│   │   ├── task.go                  # Extended Task model (source, tags, duration)
│   │   ├── task_status.go           # TaskStatus enum, constants
│   │   ├── task_pool.go             # Unified TaskPool (multi-source)
│   │   ├── door_selector.go         # Intelligent door selection (learning + calendar)
│   │   ├── status_manager.go        # Status transition validator
│   │   ├── enrichment_store.go      # SQLite enrichment DB (Epic 6)
│   │   ├── doctor.go                # System diagnostics checker (Epic 49)
│   │   ├── config.go                # Configuration model, config.yaml loader
│   │   └── connection/              # Connection lifecycle management (Epic 43)
│   │       ├── connection.go        # Connection model (ULID ID, state, settings)
│   │       ├── state.go             # ConnectionState enum, state machine transitions
│   │       ├── manager.go           # ConnectionManager — thread-safe CRUD + state transitions
│   │       ├── service.go           # ConnectionService — orchestrates CRUD with rollback
│   │       ├── credential.go        # CredentialStore interface + Env/Chain implementations
│   │       ├── credential_ring.go   # OS keyring CredentialStore (keychain/keyring)
│   │       ├── health.go            # HealthChecker, Syncer interfaces + HealthCheckResult
│   │       ├── bridge.go            # ProviderBridge — adapts TaskProvider to HealthChecker/Syncer
│   │       ├── conn_scheduler.go    # ConnAwareScheduler — state-aware polling per connection
│   │       ├── resolve.go           # ResolveFromConfig — wires config → manager → providers
│   │       ├── form_spec.go         # FormSpec/FormField — provider config form definitions
│   │       ├── sync_event.go        # SyncEventLog — per-connection JSONL audit log
│   │       ├── health_warnings.go   # Proactive health notifications (Epic 47)
│   │       └── oauth/               # OAuth device code flow
│   │           ├── devicecode.go    # Device code grant implementation
│   │           └── browser.go       # Cross-platform browser launcher
│   │
│   ├── tasks/                        # Domain Layer (Phase 1 - Tech Demo)
│   │   ├── task.go                  # Task model (Tech Demo)
│   │   ├── task_status.go           # TaskStatus enum
│   │   ├── task_pool.go             # TaskPool collection manager
│   │   ├── door_selection.go        # DoorSelection model
│   │   ├── door_selector.go         # Door selection logic
│   │   ├── status_manager.go        # Status transition validator
│   │   ├── file_manager.go          # YAML I/O, atomic writes
│   │   └── config.go                # Configuration model
│   │
│   ├── adapters/                     # Adapter Layer (Phase 2+)
│   │   ├── registry.go              # AdapterRegistry - provider discovery/loading
│   │   ├── provider.go              # TaskProvider interface definition
│   │   ├── textfile/                # Text file adapter
│   │   │   └── adapter.go          # TextFileAdapter (evolved from FileManager)
│   │   ├── applenotes/              # Apple Notes adapter (Epic 2)
│   │   │   ├── adapter.go          # AppleNotesAdapter
│   │   │   └── applescript.go      # AppleScript bridge helpers
│   │   ├── github/                  # GitHub Issues adapter (Epic 46)
│   │   │   └── oauth.go            # GitHub OAuth device code flow integration
│   │   └── obsidian/                # Obsidian vault adapter (Epic 8)
│   │       ├── adapter.go          # ObsidianAdapter
│   │       ├── markdown.go         # Markdown task parser
│   │       └── daily_notes.go      # Daily note integration
│   │
│   ├── sync/                         # Sync Engine (Phase 3, Epic 11)
│   │   ├── engine.go                # SyncEngine orchestrator
│   │   ├── queue.go                 # OfflineQueue (JSONL)
│   │   ├── conflict.go             # ConflictResolver
│   │   └── log.go                   # SyncLog (rotating)
│   │
│   ├── intelligence/                 # Intelligence Layer (Phase 3-4)
│   │   ├── calendar/                # Calendar awareness (Epic 12)
│   │   │   ├── reader.go           # CalendarReader interface
│   │   │   ├── applescript.go      # macOS Calendar.app reader
│   │   │   ├── ics.go              # .ics file parser
│   │   │   └── caldav.go           # CalDAV cache reader
│   │   ├── llm/                     # LLM decomposition (Epic 14, 57)
│   │   │   ├── decomposer.go       # LLMTaskDecomposer
│   │   │   ├── backend.go          # LLMBackend interface, CLIConfig, NewCLIBackend factory
│   │   │   ├── cli_provider.go     # CLIProvider — subprocess-based LLMBackend (Epic 57)
│   │   │   ├── cli_spec.go         # CLISpec — command/args/parsing for CLI backends
│   │   │   ├── cli_specs.go        # Built-in specs: Claude CLI, Gemini CLI, Ollama CLI
│   │   │   ├── runner.go           # CLIRunner interface, ExecRunner implementation
│   │   │   ├── local.go            # Ollama/llama.cpp client
│   │   │   ├── cloud.go            # Anthropic/OpenAI client
│   │   │   └── git_output.go       # Git repo story writer
│   │   ├── services/                # Intelligence services (Epic 57)
│   │   │   ├── extractor.go        # TaskExtractor — LLM-based task extraction from text
│   │   │   └── prompts.go          # Extraction prompt templates
│   │   └── learning/                # Learning engine (Epic 4, enhanced Epic 12)
│   │       ├── engine.go            # Pattern analysis
│   │       └── patterns.go          # User pattern models
│   │
│   └── aggregator/                   # Multi-Source Aggregation (Phase 3, Epic 13)
│       ├── aggregator.go            # MultiSourceAggregator
│       └── dedup.go                 # DuplicateDetector
│
├── docs/                             # Documentation
│   ├── prd/                         # Product Requirements Document (sharded)
│   ├── architecture/                # Architecture documentation (sharded)
│   └── stories/                     # Story breakdowns
│
├── .bmad-core/                       # BMAD methodology artifacts
├── .github/
│   └── workflows/
│       └── ci.yml                   # CI/CD pipeline (quality gates + alpha release)
│
├── bin/                              # Build output (gitignored)
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
├── Makefile                          # Build automation
├── .gitignore                        # Git ignore rules
└── README.md                         # Quick start guide

User Data Directory (created at runtime):
~/.threedoors/
├── config.yaml                       # User configuration (Phase 2+)
├── tasks.yaml                        # Active tasks with metadata
├── completed.txt                     # Completed task log
├── metrics.jsonl                     # Session metrics (Phase 1+)
├── enrichment.db                     # SQLite enrichment (Phase 2+)
└── sync-state/                       # Sync engine state (Phase 3+)
    ├── queue.jsonl                   # Offline change queue
    └── sync.log                      # Sync debug log
```

**Post-MVP Organization Principles:**

1. **`internal/core/`** replaces `internal/tasks/` as the primary domain package (tasks/ kept for Phase 1 compatibility)
2. **`internal/adapters/`** each adapter in its own sub-package for isolation
3. **`internal/sync/`** self-contained sync engine, no TUI dependencies
4. **`internal/intelligence/`** optional features, no core dependencies
5. **`internal/aggregator/`** bridges adapters and core via unified pool
6. **Dependency direction:** TUI → Core → Adapters (never reverse)

---
