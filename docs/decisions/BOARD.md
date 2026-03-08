# Knowledge Decisions Board

> The living dashboard for all ThreeDoors project decisions. See [README.md](README.md) for how this board works.

---

## Open Questions

| ID | Question | Date | Owner | Context |
|----|----------|------|-------|---------|
| | *No open questions* | | | |

## Active Research

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| | *No active research* | | | |

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| | *No pending recommendations* | | | | |

## Decided

| ID | Decision | Date | Rationale | Link |
|----|----------|------|-----------|------|
| D-001 | Go as primary language | 2025-11-07 | Single binary, fast compilation, excellent TUI ecosystem | [ADR-0001](../ADRs/ADR-0001-go-as-primary-language.md) |
| D-002 | Bubbletea TUI framework | 2025-11-07 | Elm Architecture MVU pattern, clean state management, async via tea.Cmd | [ADR-0002](../ADRs/ADR-0002-bubbletea-tui-framework.md) |
| D-003 | YAML for task persistence | 2025-11-07 | Human-readable, version-control friendly, no database dependency | [ADR-0003](../ADRs/ADR-0003-yaml-task-persistence.md) |
| D-004 | Monolithic CLI architecture | 2025-11-07 | Local-first, fast startup, simple distribution, all state in ~/.threedoors/ | [ADR-0004](../ADRs/ADR-0004-monolithic-cli-architecture.md) |
| D-005 | Layered architecture evolution | 2025-11-07 | Incremental 2-to-5-layer growth as epics demand new layers | [ADR-0005](../ADRs/ADR-0005-layered-architecture-evolution.md) |
| D-006 | TaskProvider interface pattern | 2025-11-10 | All storage backends implement one interface; ErrReadOnly for unsupported ops | [ADR-0006](../ADRs/ADR-0006-taskprovider-interface-pattern.md) |
| D-007 | Compile-time adapter registration | 2026-01-15 | Adapters as Go packages imported in main.go; simplest viable approach | [ADR-0007](../ADRs/ADR-0007-compile-time-adapter-registration.md) |
| D-008 | Atomic file writes | 2025-11-07 | Write to .tmp, fsync, rename; prevents corruption on crash | [ADR-0008](../ADRs/ADR-0008-atomic-file-writes.md) |
| D-009 | Task status state machine | 2025-11-07 | Five-state machine with validated transitions; supports blocking and review | [ADR-0009](../ADRs/ADR-0009-task-status-state-machine.md) |
| D-010 | Incremental task model extension | 2026-01-20 | Add struct fields as epics need them; each epic owns its migration | [ADR-0010](../ADRs/ADR-0010-incremental-task-model-extension.md) |
| D-011 | Sync scheduler via tea.Cmd | 2026-01-20 | Dispatch sync as Bubbletea commands; testable and composable | [ADR-0011](../ADRs/ADR-0011-sync-scheduler-via-tea-cmd.md) |
| D-012 | Property-level conflict resolution | 2026-01-20 | Per-field versioning via FieldVersions map; prevents cross-field data loss | [ADR-0012](../ADRs/ADR-0012-property-level-conflict-resolution.md) |
| D-013 | Offline-first with local change queue | 2026-02-01 | Write-ahead log records changes locally; sync replays when available | [ADR-0013](../ADRs/ADR-0013-offline-first-local-change-queue.md) |
| D-014 | Auto-migrate schema on load | 2026-02-15 | Detect schema version, transparently convert; zero user friction | [ADR-0014](../ADRs/ADR-0014-auto-migrate-schema-on-load.md) |
| D-015 | Multi-source dedup strategy | 2026-02-01 | Fuzzy matching + SourceRef linking; prevents false-positive merges | [ADR-0015](../ADRs/ADR-0015-multi-source-dedup-strategy.md) |
| D-016 | Cancel SQLite enrichment layer | 2026-01-20 | File-based storage sufficient; avoids CGO dependency and complexity | [ADR-0016](../ADRs/ADR-0016-cancel-sqlite-enrichment-layer.md) |
| D-017 | Local-first calendar integration | 2026-02-01 | AppleScript + .ics parsing; no OAuth/cloud APIs; privacy-preserving | [ADR-0017](../ADRs/ADR-0017-local-first-calendar-integration.md) |
| D-018 | macOS code signing and notarization | 2025-11-09 | CI-based via GitHub Actions; multiple distribution formats | [ADR-0018](../ADRs/ADR-0018-macos-code-signing-notarization.md) |
| D-019 | Docker E2E and headless TUI testing | 2026-02-15 | Three-tier: headless teatest, golden snapshots, Docker E2E | [ADR-0019](../ADRs/ADR-0019-docker-e2e-testing.md) |
| D-020 | Door theme system | 2026-02-20 | Registry-based with DoorTheme interface; Classic, Modern, Sci-Fi, Shoji | [ADR-0020](../ADRs/ADR-0020-door-theme-system.md) |
| D-021 | MCP server integration | 2026-03-01 | ThreeDoors as MCP server; stdio and SSE transports; AI agents manage tasks | [ADR-0021](../ADRs/ADR-0021-mcp-server-integration.md) |
| D-022 | CLI interface with Cobra | 2026-03-01 | Cobra-based CLI alongside TUI; --json output for scripting | [ADR-0022](../ADRs/ADR-0022-cli-interface-with-cobra.md) |
| D-023 | iPhone app deferred | 2026-03-07 | No validated demand; focus on core macOS persona; revisit if 5+ requests | [ADR-0023](../ADRs/ADR-0023-iphone-app-deferred.md) |
| D-024 | JSONL session metrics format | 2025-11-08 | Append-only JSONL; fast, safe, streamable; no schema migration needed | [ADR-0024](../ADRs/ADR-0024-jsonl-session-metrics.md) |
| D-025 | Story-driven development | 2026-01-15 | Mandatory story files; acceptance criteria as binary pass/fail; full traceability | [ADR-0025](../ADRs/ADR-0025-story-driven-development.md) |
| D-026 | Self-driving development pipeline | 2026-03-01 | Shell script MVP for story dispatch to multiclaude workers; safety guardrails | [ADR-0026](../ADRs/ADR-0026-self-driving-development-pipeline.md) |
| D-027 | Multi-provider integration strategy | 2026-02-01 | Three-phase per provider: read-only → bidirectional → advanced; contract tests | [ADR-0027](../ADRs/ADR-0027-multi-provider-integration-strategy.md) |
| D-028 | CI quality gates and testing strategy | 2025-11-09 | Multi-layer gates; 70% coverage minimum; no bypass allowed | [ADR-0028](../ADRs/ADR-0028-ci-quality-gates.md) |
| D-029 | Knowledge Decisions Board | 2026-03-08 | Lifecycle-aware kanban; single file dashboard; zero infrastructure | [Research](../research/decision-management-research.md) |

## Rejected

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-001 | SQLite enrichment layer | 2026-01-20 | Overhead exceeded benefit; file-based storage sufficient | [ADR-0016](../ADRs/ADR-0016-cancel-sqlite-enrichment-layer.md) |
| X-002 | ADRs as single canonical decision format | 2026-03-08 | Too heavyweight for micro-decisions; would create ADR sprawl | [Research](../research/decision-management-research.md) |
| X-003 | RFC process for decisions | 2026-03-08 | Too heavyweight; party mode already handles deliberation | [Research](../research/decision-management-research.md) |
| X-004 | Tag-based decision system | 2026-03-08 | High retrofit cost (135+ files); distributes source of truth | [Research](../research/decision-management-research.md) |
| X-005 | Automated ADR generation | 2026-03-08 | ADR creation requires significance judgment; auto-generated would be noisy | [Research](../research/decision-management-research.md) |
| X-006 | iPhone native app (Epic 16) | 2026-03-07 | No validated demand; deferred indefinitely | [ADR-0023](../ADRs/ADR-0023-iphone-app-deferred.md) |

## Superseded

| ID | Original Decision | Date | Superseded By | Link |
|----|-------------------|------|---------------|------|
| | *No superseded decisions* | | | |
