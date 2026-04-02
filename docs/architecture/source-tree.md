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
│   │   ├── command_dispatch.go      # Command dispatch helpers extracted from main_model (Epic 69)
│   │   ├── view_auxiliary_controller.go  # Auxiliary view message handling (Epic 69)
│   │   ├── view_sources_controller.go    # Source view resize/update handling (Epic 69)
│   │   ├── view_navigation.go       # View navigation and transition logic (Epic 69)
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
│   │   ├── jira/                    # Jira adapter
│   │   │   └── ...                 # Jira REST API client & provider
│   │   ├── linear/                  # Linear adapter (Epic 30, 46)
│   │   │   ├── client.go           # Linear GraphQL API client
│   │   │   ├── config.go           # Linear adapter configuration
│   │   │   ├── provider.go         # LinearProvider — read-only TaskProvider (Story 30.2)
│   │   │   ├── types.go            # GraphQL query/response types
│   │   │   └── oauth.go            # Linear OAuth integration
│   │   ├── clickup/                 # ClickUp adapter (Epic 63)
│   │   │   ├── clickup_client.go   # ClickUp REST API v2 client
│   │   │   └── config.go           # ClickUp auth & workspace configuration
│   │   ├── reminders/               # Apple Reminders adapter
│   │   │   └── ...                 # Reminders EventKit bridge
│   │   ├── todoist/                 # Todoist adapter
│   │   │   └── ...                 # Todoist REST API client & provider
│   │   └── obsidian/                # Obsidian vault adapter (Epic 8)
│   │       ├── adapter.go          # ObsidianAdapter
│   │       ├── markdown.go         # Markdown task parser
│   │       └── daily_notes.go      # Daily note integration
│   │
│   ├── sync/                         # Sync Engine (Phase 3, Epic 11/64)
│   │   ├── transport.go             # SyncTransport interface, SyncOp, Changeset
│   │   ├── git_transport.go         # GitSyncTransport — Git-based SyncTransport impl
│   │   ├── git_executor.go          # GitExecutor interface, ExecGitExecutor
│   │   ├── connection.go            # GitSyncConnection — Connection Manager adapter
│   │   ├── offline.go               # OfflineManager — offline queue & reconciliation (Epic 64)
│   │   ├── debounce.go              # Debouncer — coalesces rapid sync events
│   │   └── errors.go                # Sentinel errors (ErrGitNotFound, ErrRebaseConflict, etc.)
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
│   ├── calendar/                      # Calendar Awareness (extracted from intelligence/)
│   │   ├── calendar.go              # Calendar integration core
│   │   ├── config.go                # Calendar configuration
│   │   ├── applescript_reader.go    # macOS Calendar.app reader
│   │   └── caldav_cache_reader.go   # CalDAV cache reader
│   │
│   ├── ci/                            # CI Validation
│   │   └── ci_validation_test.go    # CI pipeline validation tests
│   │
│   ├── dispatch/                      # CLI Dispatch & Audit
│   │   ├── cli_dispatcher.go        # CLI command dispatcher
│   │   ├── command_runner.go        # Command execution runner
│   │   └── audit.go                 # Dispatch audit logging
│   │
│   ├── dist/                          # Distribution & Packaging
│   │   ├── dist.go                  # Distribution pipeline
│   │   ├── codesign.go              # macOS code signing
│   │   ├── notarize.go              # macOS notarization
│   │   ├── pkg_builder.go           # Package builder
│   │   └── version.go               # Version embedding
│   │
│   ├── docaudit/                      # Documentation Audit
│   │   ├── audit.go                 # Documentation audit engine
│   │   ├── loader.go                # Document loader
│   │   └── parse_epic_list.go       # Epic list parser
│   │
│   ├── enrichment/                    # Task Enrichment
│   │   ├── enrichment.go            # Enrichment engine
│   │   └── cross_references.go      # Cross-reference analysis
│   │
│   ├── mcp/                           # MCP Server Tools (Epic 53+)
│   │   ├── protocol.go              # MCP protocol implementation
│   │   ├── middleware.go             # Request middleware
│   │   ├── analytics.go             # Analytics tools
│   │   ├── advanced_tools.go        # Advanced MCP tools
│   │   ├── graph.go                 # Task graph tools
│   │   ├── intake.go                # Task intake tools
│   │   ├── prompts.go               # MCP prompt definitions
│   │   ├── proposal.go              # Task proposal model
│   │   ├── proposal_tools.go        # Proposal MCP tools
│   │   └── proposal_store.go        # Proposal persistence
│   │
│   ├── quota/                          # Quota Monitoring & Usage Tracking (Epic 76)
│   │   ├── types.go                 # Domain types: TokenCount, Interaction, SessionUsage, WindowUsage, UsageSnapshot, PlanBudget
│   │   ├── parser.go                # JSONL token usage parser (Claude session logs)
│   │   ├── discover.go              # Session file discovery (finds JSONL logs)
│   │   ├── aggregate.go             # Rolling-window token aggregation
│   │   ├── attribution.go           # Per-agent usage attribution
│   │   ├── config.go                # Threshold configuration
│   │   ├── threshold.go             # Warning threshold engine (4-tier)
│   │   ├── notify.go                # Threshold notification system
│   │   ├── snapshot.go              # Usage snapshot for /stats integration
│   │   └── testdata/                # Test fixtures (JSONL samples)
│   │
│   ├── retrospector/                  # Retrospective Analysis (Epic 62)
│   │   ├── finding.go               # Finding model
│   │   ├── findings.go              # Findings collection/persistence
│   │   ├── ac_match.go              # Acceptance criteria matching
│   │   ├── board.go                 # Decision board analysis
│   │   ├── ci_analysis.go           # CI failure analysis
│   │   ├── ci_taxonomy.go           # CI failure taxonomy
│   │   ├── confidence.go            # Confidence scoring
│   │   └── conflict_analysis.go     # Merge conflict analysis
│   │
│   ├── slaes/                         # Saga/Pattern Detection
│   │   └── saga/                    # Saga pattern detector
│   │       ├── detector.go          # Saga detection engine
│   │       ├── dispatch.go          # Saga dispatch logic
│   │       ├── findings.go          # Saga findings model
│   │       └── recurrence.go        # Recurrence pattern detection
│   │
│   └── testkit/                       # Test Utilities
│       ├── assertions.go            # Custom test assertions
│       ├── factories.go             # Test data factories
│       └── provider.go              # Mock TaskProvider for tests
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

1. **`internal/core/`** replaces `internal/tasks/` as the primary domain package (tasks/ kept for Phase 1 compatibility). Sync engine and aggregator live in core/.
2. **`internal/adapters/`** each adapter in its own sub-package for isolation
3. **`internal/intelligence/`** optional features, no core dependencies
4. **`internal/mcp/`** MCP tool definitions; `internal/mcpbridge/` for the bridge server
5. **`internal/retrospector/`** retrospective analysis (Epic 62), standalone package
6. **`internal/dist/`** distribution/packaging pipeline, standalone from core
7. **`internal/testkit/`** shared test utilities (factories, assertions, mock providers)
8. **`internal/quota/`** quota monitoring and usage tracking (Epic 76), standalone operational package — parses Claude JSONL logs, aggregates token usage, provides threshold warnings
9. **Dependency direction:** TUI → Core → Adapters (never reverse)

---
