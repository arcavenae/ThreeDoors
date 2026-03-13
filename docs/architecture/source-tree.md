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
│   ├── threedoors/
│   │   └── main.go                    # Application entry point
│   └── multiclaude-mcp-bridge/
│       └── main.go                    # MCP bridge server entry point (Epic 53)
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
│   │   ├── breakdown_view.go       # Task breakdown TUI view (Epic 57)
│   │   ├── extract_view.go         # :extract command — review screen for extracted tasks (Epic 57)
│   │   ├── llm_status.go           # LLM backend status TUI view (Epic 57)
│   │   ├── detail_view.go          # Task detail enrichment view
│   │   ├── feedback_view.go        # User feedback collection view
│   │   ├── deferred_list_view.go   # Deferred tasks list view
│   │   ├── devqueue_view.go        # Developer queue view
│   │   ├── health_view.go          # System health status view
│   │   ├── insights_view.go        # Analytics insights view
│   │   ├── mood_view.go            # Mood tracking view
│   │   ├── next_steps_view.go      # Next steps suggestions view
│   │   ├── orphaned_view.go       # Orphaned task handling view (Epic 47)
│   │   ├── reauth_dialog.go      # Re-authentication dialog for expired connections (Epic 44)
│   │   ├── proposals_view.go       # Task proposals view
│   │   ├── planning_view.go        # Planning mode view
│   │   ├── planning_select.go      # Planning task selection
│   │   ├── planning_review.go      # Planning review screen
│   │   ├── planning_confirm.go     # Planning confirmation screen
│   │   ├── snooze_view.go          # Task snooze view
│   │   ├── tag_view.go             # Tag management view
│   │   ├── values_view.go          # Values alignment view
│   │   ├── help_view.go            # Help overlay view
│   │   ├── theme_picker.go         # Theme selection picker
│   │   ├── avoidance_prompt_view.go # Avoidance pattern prompt
│   │   ├── add_task_view.go        # Inline task creation view
│   │   ├── sync_status_view.go     # Sync status display
│   │   ├── synclog_view.go         # Sync log viewer
│   │   ├── keybindings.go          # Keybinding definitions
│   │   ├── keybinding_bar.go       # Keybinding hint bar (Epic 39)
│   │   ├── keybinding_overlay.go   # Full keybinding overlay (Epic 39)
│   │   ├── inline_hints.go         # Inline contextual hints
│   │   ├── breadcrumb.go           # Navigation breadcrumb
│   │   ├── scrollable_view.go      # Scrollable content wrapper
│   │   ├── spinner.go              # Loading spinner component
│   │   ├── animation.go            # Animation utilities
│   │   ├── layout.go               # AltScreen layout engine, door height cap (Epic 59)
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
│   │   ├── exitcodes.go             # Standardized exit codes
│   │   ├── extract.go               # `threedoors extract` task extraction from text (Epic 57)
│   │   ├── llm.go                   # `threedoors llm` command group (Epic 57)
│   │   ├── llm_status.go            # `threedoors llm status` backend discovery (Epic 57)
│   │   ├── interactive.go           # Interactive mode helpers
│   │   ├── completion.go            # Shell completion generation
│   │   ├── flag_completions.go      # Flag-level shell completions
│   │   ├── doc_audit.go             # `threedoors doc-audit` documentation checker
│   │   └── task_status_cmds.go      # Task status shortcut commands
│   │
│   ├── core/                         # Core Domain (Phase 2+)
│   │   ├── task.go                  # Extended Task model (source, tags, duration)
│   │   ├── task_status.go           # TaskStatus enum, constants
│   │   ├── task_pool.go             # Unified TaskPool (multi-source)
│   │   ├── door_selection.go        # DoorSelection model
│   │   ├── door_selector.go         # Intelligent door selection (learning + calendar)
│   │   ├── doctor.go                # System diagnostics checker (Epic 49)
│   │   ├── doctor_database.go       # Database diagnostics
│   │   ├── doctor_session.go        # Session diagnostics
│   │   ├── doctor_sync.go           # Sync diagnostics
│   │   ├── doctor_task_data.go      # Task data diagnostics
│   │   ├── sync_engine.go           # Core sync engine orchestrator
│   │   ├── sync_log.go              # Sync event logging
│   │   ├── sync_scheduler.go        # Sync scheduling logic
│   │   ├── sync_state.go            # Sync state persistence
│   │   ├── sync_status.go           # Sync status reporting
│   │   ├── conflict_resolver.go     # Sync conflict resolution
│   │   ├── field_conflict_resolver.go # Field-level conflict resolution (Epic 50)
│   │   ├── aggregator.go            # Multi-source task aggregation
│   │   ├── dedup_store.go           # Deduplication store
│   │   ├── duplicate_detector.go    # Duplicate task detection
│   │   ├── provider.go              # TaskProvider interface
│   │   ├── provider_config.go       # Provider configuration model
│   │   ├── provider_factory.go      # Provider factory functions
│   │   ├── registry.go              # Provider registry
│   │   ├── fallback_provider.go     # Fallback provider chain
│   │   ├── wal_provider.go          # Write-ahead log provider wrapper
│   │   ├── write_queue.go           # Batched write queue
│   │   ├── circuit_breaker.go       # Circuit breaker for external calls
│   │   ├── health_checker.go        # Provider health checking
│   │   ├── credential_check.go      # Credential validation
│   │   ├── config_paths.go          # Configuration path resolution
│   │   ├── file_limits.go           # File size/count limits
│   │   ├── path_validation.go       # Path validation utilities
│   │   ├── version_check.go         # Version compatibility checking
│   │   ├── onboarding.go            # First-run onboarding logic
│   │   ├── session_tracker.go       # Session tracking/analytics
│   │   ├── metrics_writer.go        # Metrics JSONL writer
│   │   ├── completion_counter.go    # Task completion tracking
│   │   ├── milestones.go            # Milestone celebrations (Epic 40)
│   │   ├── pattern_analyzer.go      # User pattern analysis
│   │   ├── greeting_insights.go     # Greeting-time insights
│   │   ├── insights_formatter.go    # Insights display formatting
│   │   ├── planning_metrics.go      # Planning mode metrics
│   │   ├── fun_facts.go             # Fun facts for idle screens
│   │   ├── energy.go                # Energy level tracking
│   │   ├── focus.go                 # Focus mode logic
│   │   ├── mood_selector.go         # Mood selection logic
│   │   ├── time_context.go          # Time-of-day context
│   │   ├── defer_return.go          # Deferred task return logic
│   │   ├── dependency.go            # Task dependency management
│   │   ├── tag_parser.go            # Tag parsing utilities
│   │   ├── task_categorization.go   # Task auto-categorization
│   │   ├── task_importer.go         # Bulk task import
│   │   ├── inline_hints_config.go   # Inline hints configuration
│   │   ├── theme_config.go          # Theme configuration
│   │   └── values_config.go         # Values alignment configuration
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
│   │       ├── token_refresh.go    # OAuth token refresh logic
│   │       ├── detect.go           # Provider detection utilities
│   │       └── oauth/               # OAuth device code flow
│   │           ├── devicecode.go    # Device code grant implementation
│   │           └── browser.go       # Cross-platform browser launcher
│   │
│   ├── mcpbridge/                    # MCP Bridge Server (Epic 53)
│   │   ├── server.go               # MCP server implementation
│   │   ├── tools.go                # MCP tool definitions and handlers
│   │   └── runner.go               # Bridge runner/lifecycle
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
│   │   ├── github/                  # GitHub Issues adapter (Epic 46, 50)
│   │   │   ├── config.go           # GitHub adapter configuration
│   │   │   ├── github_client.go    # GitHub API client (issues, bug reports)
│   │   │   ├── github_provider.go  # GitHub TaskProvider implementation
│   │   │   └── oauth.go            # GitHub OAuth device code flow integration
│   │   ├── linear/                  # Linear adapter (Epic 46)
│   │   │   ├── client.go           # Linear GraphQL API client
│   │   │   ├── config.go           # Linear adapter configuration
│   │   │   ├── types.go            # GraphQL query/response types
│   │   │   └── oauth.go            # Linear OAuth integration
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
│   │   ├── agent_service.go         # AgentService — auto-discovery + fallback chain (Epic 57)
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
│   │   │   ├── discovery.go        # Auto-discovery of installed CLI backends (Epic 57)
│   │   │   ├── claude.go           # Claude CLI backend implementation
│   │   │   ├── ollama.go           # Ollama backend implementation
│   │   │   ├── story_spec.go       # Story specification model for LLM output
│   │   │   └── git_writer.go       # Git repo story writer
│   │   ├── services/                # Intelligence services (Epic 57)
│   │   │   ├── extractor.go        # TaskExtractor — LLM-based task extraction from text
│   │   │   ├── breakdown.go        # TaskBreakdown — LLM-based task decomposition
│   │   │   ├── enricher.go         # TaskEnricher — LLM-based task detail enrichment
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
