# Knowledge Decisions Board

> The living dashboard for all ThreeDoors project decisions. See [README.md](README.md) for how this board works.

---

## Open Questions

| ID | Question | Date | Owner | Context |
|----|----------|------|-------|---------|
| Q-001 | Should Jira adapter use story points or priority for effort mapping? | 2026-03-03 | — | [Jira Research](../research/jira-integration-research.md) |
| Q-002 | Should Jira adapter support multi-project JQL or explicit project keys? | 2026-03-03 | — | [Jira Research](../research/jira-integration-research.md) |

## Active Research

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| | *No active research* | | | |

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-001 | Migrate from Makefile to Justfile | 2026-03-04 | Research spike | [Analysis](../research/makefile-vs-justfile-analysis.md) | Owner sign-off |

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
| D-030 | Nil check guard for nil provider (Issue #218) | 2026-03-08 | Minimal fix for P0 crash; matches existing bootstrap.go pattern | [Artifact](../../_bmad-output/planning-artifacts/issue-218-party-mode-consensus.md) |
| D-031 | Portrait door proportions with door anatomy model | 2026-03-08 | Door metaphor fidelity; taller-than-wide is #1 "door" signifier | [Artifact](../../_bmad-output/planning-artifacts/door-appearance-party-mode.md) |
| D-032 | DoorTheme.Render() extended with height parameter | 2026-03-08 | Enables proportional door rendering; DoorsView calculates available height | [Architecture](../../_bmad-output/planning-artifacts/door-appearance-architecture.md) |
| D-033 | DeferUntil as nullable *time.Time for snooze | 2026-03-07 | Zero-value safe; nil = not deferred; matches existing CompletedAt pattern | [Architecture](../../_bmad-output/planning-artifacts/architecture-snooze-defer.md) |
| D-034 | Z key binding for snooze action | 2026-03-07 | "Zzz" mnemonic; S already bound; available in DoorsView and DetailView | [Architecture](../../_bmad-output/planning-artifacts/architecture-snooze-defer.md) |
| D-035 | Dual auto-return for snoozed tasks (startup + tea.Tick) | 2026-03-07 | Covers both app-reopen and always-running scenarios; 1-min interval sufficient | [Architecture](../../_bmad-output/planning-artifacts/architecture-snooze-defer.md) |
| D-036 | Three-step sequenced CI churn reduction | 2026-03-08 | Path filtering → concurrency limits → caching; risk-ascending order | [Artifact](../../_bmad-output/planning-artifacts/ci-churn-reduction-party-mode.md) |
| D-037 | gh run list polling for cross-repo CI awareness | 2026-03-08 | Simple, well-understood; sufficient for infrequent release cadence | [Artifact](../../_bmad-output/planning-artifacts/cross-repo-ci-awareness-party-mode.md) |
| D-038 | Weekly grouped dependency PRs with security as individual | 2026-03-08 | Reduces PR noise; security gets immediate individual attention | [Artifact](../../_bmad-output/planning-artifacts/dependency-management-party-mode.md) |
| D-039 | OSV vulnerability scanning from day one | 2026-03-08 | Critical security baseline; low effort to enable | [Artifact](../../_bmad-output/planning-artifacts/dependency-management-party-mode.md) |
| D-040 | Custom Homebrew tap before homebrew-core | 2026-03-08 | Provides real user value now; need users to get stars, not vice versa | [Artifact](../../_bmad-output/planning-artifacts/homebrew-distribution-party-mode.md) |
| D-041 | Dependency-based door pool filtering | 2026-03-07 | Tasks with unmet dependencies excluded from doors; cleaner task selection | [Architecture](../../_bmad-output/planning-artifacts/architecture-task-dependencies.md) |
| D-042 | Undo task completion via complete→todo transition | 2026-03-08 | Minimal scope; completed.txt remains immutable audit trail | [Architecture](../../_bmad-output/planning-artifacts/architecture-undo-task-completion.md) |
| D-043 | Parent-child tasks via parent_id field (expand/fork) | 2026-03-08 | Additive field with omitempty; no schema migration needed | [Architecture](../../_bmad-output/planning-artifacts/architecture-expand-fork.md) |
| D-044 | PM + Architect as persistent agents | 2026-03-08 | Continuous monitoring needed for PR quality and architecture drift | [Artifact](../../_bmad-output/planning-artifacts/persistent-agent-architecture-round1-role-evaluation.md) |
| D-045 | SM as 4-hour cron, QA as weekly cron | 2026-03-08 | Sprint health and quality checks don't need continuous monitoring | [Artifact](../../_bmad-output/planning-artifacts/persistent-agent-architecture-round3-mvp.md) |
| D-046 | Self-directed poll-based envoy patrol rhythm | 2026-03-08 | Envoy monitors autonomously; no external trigger needed | [Artifact](../../_bmad-output/planning-artifacts/envoy-rules-of-behavior-party-mode.md) |
| D-047 | Slash commands as project skills (Epic 34) | 2026-03-08 | Codifies tribal knowledge; DRY specs as living documentation | [Architecture](../../_bmad-output/planning-artifacts/architecture-soul-skills.md) |
| D-048 | JXA via osascript for Apple Reminders adapter | 2026-03-03 | Mirrors Apple Notes adapter pattern; native JSON output avoids brittle parsing | [Research](../research/apple-reminders-integration-research.md) |
| D-049 | JQL-based Jira integration via REST API v3 | 2026-03-03 | Standard API; go-jira library available; WALProvider for offline support | [Research](../research/jira-integration-research.md) |
| D-050 | Linear integration as read-only GraphQL adapter | 2026-03-08 | Read-only first per multi-provider strategy; GraphQL is Linear's primary API | [Artifact](../../_bmad-output/planning-artifacts/epic-30-linear-integration.md) |
| D-051 | GitHub Issues as task source via gh CLI / PAT | 2026-03-07 | gh CLI already available; PAT via env var or config.yaml | [Artifact](../../_bmad-output/planning-artifacts/party-mode-github-issues-2026-03-07.md) |
| D-052 | Three-door count validated by choice architecture research | 2026-03-02 | Hick's Law, jam study, working memory limits all support 3 as optimal | [Research](../research/choice-architecture.md) |
| D-053 | Standardized Decisions Summary table in artifacts | 2026-03-08 | Makes decision extraction mechanical; consistent format across all artifacts | [Research](../research/decision-management-research.md) |
| D-054 | DRY spec cleanup for story files | 2026-03-08 | Remove duplicated content from story specs; specs reference architecture docs | [Artifact](../../_bmad-output/planning-artifacts/34.4-party-mode-dry-cleanup.md) |
| D-055 | CI churn reduction | 2026-03-08 | Relax up-to-date rule + path filtering; defer merge queue; 70-80% CI reduction | [ADR-0030](../ADRs/ADR-0030-ci-churn-reduction.md) |
| D-056 | Alpha binary named `threedoors-a` (not `threedoors`) | 2026-03-08 | Prevents Homebrew conflicts; allows simultaneous install; clear channel identity | [Artifact](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-party-mode.md) |
| D-057 | Alpha formula `threedoors-a.rb` in same tap | 2026-03-08 | Single tap; consistent UX; no `conflicts_with` needed | [Research](../research/dual-homebrew-distribution-research.md) |
| D-058 | Manual planning doc reconciliation over automation | 2026-03-08 | Automation rejected — drift is infrequent, docs are heterogeneous, CLAUDE.md reminder sufficient | [Artifact](../../_bmad-output/planning-artifacts/planning-docs-reconciliation-triage-party-mode.md) |
| D-059 | Universal quit via MainModel-level 'q' interception | 2026-03-08 | Centralizes quit logic; views don't need individual 'q' handlers; `isTextInputActive()` guards text input views | [Story 36.3](../stories/36.3.story.md) |
| D-060 | Content pre-styling for door selection contrast | 2026-03-08 | Style content before theme Render(); avoids modifying each theme; uses Bold/Faint + DoubleBorder for structural emphasis | [Story 36.1](../../docs/stories/36.1.story.md) |

## Rejected

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-001 | SQLite enrichment layer | 2026-01-20 | Overhead exceeded benefit; file-based storage sufficient | [ADR-0016](../ADRs/ADR-0016-cancel-sqlite-enrichment-layer.md) |
| X-002 | ADRs as single canonical decision format | 2026-03-08 | Too heavyweight for micro-decisions; would create ADR sprawl | [Research](../research/decision-management-research.md) |
| X-003 | RFC process for decisions | 2026-03-08 | Too heavyweight; party mode already handles deliberation | [Research](../research/decision-management-research.md) |
| X-004 | Tag-based decision system | 2026-03-08 | High retrofit cost (135+ files); distributes source of truth | [Research](../research/decision-management-research.md) |
| X-005 | Automated ADR generation | 2026-03-08 | ADR creation requires significance judgment; auto-generated would be noisy | [Research](../research/decision-management-research.md) |
| X-006 | iPhone native app (Epic 16) | 2026-03-07 | No validated demand; deferred indefinitely | [ADR-0023](../ADRs/ADR-0023-iphone-app-deferred.md) |
| X-007 | NewProviderFromConfig() signature refactor | 2026-03-08 | Too large for bug fix scope; separate story needed | [Artifact](../../_bmad-output/planning-artifacts/issue-218-party-mode-consensus.md) |
| X-008 | Door opening/closing animations | 2026-03-08 | P2 scope; portrait proportions are the prerequisite | [Artifact](../../_bmad-output/planning-artifacts/door-appearance-party-mode.md) |
| X-009 | Perspective/vanishing point hallway effect | 2026-03-08 | Too complex for initial door redesign; deferred to future epic | [Artifact](../../_bmad-output/planning-artifacts/door-appearance-party-mode.md) |
| X-010 | GitHub Merge Queue for CI | 2026-03-08 | Deferred pending measurement of path filtering + concurrency quick wins | [Artifact](../../_bmad-output/planning-artifacts/ci-churn-reduction-party-mode.md) |
| X-011 | repository_dispatch for cross-repo CI monitoring | 2026-03-08 | More complex than polling; requires PAT with cross-repo scope | [Artifact](../../_bmad-output/planning-artifacts/cross-repo-ci-awareness-party-mode.md) |
| X-012 | GitHub webhooks for CI notification | 2026-03-08 | Requires webhook server infrastructure; massive overengineering for the need | [Artifact](../../_bmad-output/planning-artifacts/cross-repo-ci-awareness-party-mode.md) |
| X-013 | Wait for homebrew-core before any Homebrew distribution | 2026-03-08 | Circular — need distribution to get users to get stars; custom tap first | [Artifact](../../_bmad-output/planning-artifacts/homebrew-distribution-party-mode.md) |
| X-014 | Homebrew Cask instead of Formula | 2026-03-08 | Casks are for GUI apps; ThreeDoors is CLI/TUI — formula is correct | [Artifact](../../_bmad-output/planning-artifacts/homebrew-distribution-party-mode.md) |
| X-015 | General undo system / undo stack | 2026-03-08 | Out of scope; only task completion undo is validated user need | [Architecture](../../_bmad-output/planning-artifacts/architecture-undo-task-completion.md) |
| X-016 | Time-limited undo with countdown UI | 2026-03-08 | Adds unnecessary complexity; simple status transition suffices | [Artifact](../../_bmad-output/planning-artifacts/party-mode-undo-task-completion-2026-03-08.md) |
| X-017 | Tech Writer as persistent agent | 2026-03-08 | Doc drift happens over weeks; weekly cron audit achieves same result | [Artifact](../../_bmad-output/planning-artifacts/persistent-agent-architecture-round1-role-evaluation.md) |
| X-018 | UX Designer as persistent agent | 2026-03-08 | Zero monitoring surface for CLI/TUI; UX decisions made during story planning | [Artifact](../../_bmad-output/planning-artifacts/persistent-agent-architecture-round1-role-evaluation.md) |
| X-019 | Dense agent mesh (every agent talks to every other) | 2026-03-08 | Combinatorial explosion; hub-and-spoke (PM as hub) is simpler and sufficient | [Artifact](../../_bmad-output/planning-artifacts/persistent-agent-architecture-round2-collaboration.md) |
| X-020 | SLSA verification for dependencies | 2026-03-08 | Deferred to future story; OSV scanning is sufficient baseline | [Artifact](../../_bmad-output/planning-artifacts/dependency-management-party-mode.md) |
| X-021 | Same binary name with `conflicts_with` for alpha | 2026-03-08 | Prevents simultaneous install; poor UX for users wanting both channels | [Artifact](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-party-mode.md) |
| X-022 | `threedoors@alpha` formula naming | 2026-03-08 | `@` is for versioned formulae, not rolling channels; triggers keg_only expectations | [Artifact](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-party-mode.md) |
| X-023 | Separate tap for alpha (`homebrew-tap-alpha`) | 2026-03-08 | Unnecessary complexity; single tap with two formulae is standard | [Artifact](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-party-mode.md) |
| X-024 | Per-view 'q' handler for universal quit | 2026-03-08 | Would require modifying 10+ view files; higher maintenance burden; centralized interception is simpler | [Story 36.3](../stories/36.3.story.md) |
| X-025 | Per-theme Render() modification for selection contrast | 2026-03-08 | Content pre-styling at DoorsView level is simpler and requires no theme changes | [Story 36.1](../../docs/stories/36.1.story.md) |

## Superseded

| ID | Original Decision | Date | Superseded By | Link |
|----|-------------------|------|---------------|------|
| | *No superseded decisions* | | | |
