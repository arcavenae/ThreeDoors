# Knowledge Decisions Board

> The living dashboard for all ThreeDoors project decisions. See [README.md](README.md) for how this board works.

---

## Open Questions

| ID | Question | Date | Owner | Context |
|----|----------|------|-------|---------|
| Q-001 | Should Jira adapter use story points or priority for effort mapping? | 2026-03-03 | — | [Jira Research](../../_bmad-output/planning-artifacts/jira-integration-research.md) — **Resolved:** Support story points as a preference, but require the custom field involved to be specified — it's not built in. |
| Q-002 | Should Jira adapter support multi-project JQL or explicit project keys? | 2026-03-03 | — | [Jira Research](../../_bmad-output/planning-artifacts/jira-integration-research.md) — **Resolved:** Support multi-project JQL in Jira adapter. |
| Q-003 | Should project-watchdog batch governance sync PRs instead of one-per-story? | 2026-03-09 | PM | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |
| Q-004 | Should workers stop updating planning docs (ROADMAP.md, epic-list.md, epics-and-stories.md) and leave that exclusively to project-watchdog? | 2026-03-09 | PM | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |

## Active Research

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| R-001 | State of Testing Audit — comprehensive test health assessment with gap analysis and prioritized recommendations | 2026-03-09 | TEA Agent | [Report](../../_bmad-output/planning-artifacts/state-of-testing-report.md) |

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-006 | In-app bug reporting via `:bug` command — browser URL primary, PAT upgrade, file fallback | 2026-03-09 | Party mode (4 rounds: PM, Architect, UX, Dev) | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md), [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) | Epic/story creation |
| P-001 | Migrate from Makefile to Justfile | 2026-03-04 | Research spike | [Analysis](../../_bmad-output/planning-artifacts/makefile-vs-justfile-analysis.md) | Owner sign-off |
| P-002 | Envoy three-layer firewall implementation | 2026-03-08 | Party mode (8 sessions) | [Artifact](../../_bmad-output/planning-artifacts/envoy-scope-and-firewall-design.md) | Story creation |
| P-003 | GitHub issue labeling taxonomy and triage flow | 2026-03-08 | Party mode (5 sessions) | [Artifact](../../_bmad-output/planning-artifacts/issue-labeling-and-triage-strategy.md) | Story creation |
| P-004 | Update pr-shepherd definition to remove fork references | 2026-03-08 | Investigation | [Research](../../_bmad-output/planning-artifacts/persistent-agent-communication-research.md) | **Approved** — update then run `/sync-enhancements` after merge |
| P-005 | Scoped label taxonomy: 27 labels with `.` separator, migration plan | 2026-03-08 | Party mode (3 rounds) + research spike | [Party Mode](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md), [Research](../../_bmad-output/planning-artifacts/scoped-labels-research.md) | Story creation for migration |

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
| D-029 | Knowledge Decisions Board | 2026-03-08 | Lifecycle-aware kanban; single file dashboard; zero infrastructure | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
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
| D-048 | JXA via osascript for Apple Reminders adapter | 2026-03-03 | Mirrors Apple Notes adapter pattern; native JSON output avoids brittle parsing | [Research](../../_bmad-output/planning-artifacts/apple-reminders-integration-research.md) |
| D-049 | JQL-based Jira integration via REST API v3 | 2026-03-03 | Standard API; go-jira library available; WALProvider for offline support | [Research](../../_bmad-output/planning-artifacts/jira-integration-research.md) |
| D-050 | Linear integration as read-only GraphQL adapter | 2026-03-08 | Read-only first per multi-provider strategy; GraphQL is Linear's primary API | [Artifact](../../_bmad-output/planning-artifacts/epic-30-linear-integration.md) |
| D-051 | GitHub Issues as task source via gh CLI / PAT | 2026-03-07 | gh CLI already available; PAT via env var or config.yaml | [Artifact](../../_bmad-output/planning-artifacts/party-mode-github-issues-2026-03-07.md) |
| D-052 | Three-door count validated by choice architecture research | 2026-03-02 | Hick's Law, jam study, working memory limits all support 3 as optimal | [Research](../../_bmad-output/planning-artifacts/choice-architecture-research.md) |
| D-053 | Standardized Decisions Summary table in artifacts | 2026-03-08 | Makes decision extraction mechanical; consistent format across all artifacts | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
| D-054 | DRY spec cleanup for story files | 2026-03-08 | Remove duplicated content from story specs; specs reference architecture docs | [Artifact](../../_bmad-output/planning-artifacts/34.4-party-mode-dry-cleanup.md) |
| D-055 | CI churn reduction | 2026-03-08 | Relax up-to-date rule + path filtering; defer merge queue; 70-80% CI reduction | [ADR-0030](../ADRs/ADR-0030-ci-churn-reduction.md) |
| D-056 | Alpha binary named `threedoors-a` (not `threedoors`) | 2026-03-08 | Prevents Homebrew conflicts; allows simultaneous install; clear channel identity | [Artifact](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-party-mode.md) |
| D-057 | Alpha formula `threedoors-a.rb` in same tap | 2026-03-08 | Single tap; consistent UX; no `conflicts_with` needed | [Research](../../_bmad-output/planning-artifacts/dual-homebrew-distribution-research.md) |
| D-058 | Manual planning doc reconciliation over automation | 2026-03-08 | Automation rejected — drift is infrequent, docs are heterogeneous, CLAUDE.md reminder sufficient | [Artifact](../../_bmad-output/planning-artifacts/planning-docs-reconciliation-triage-party-mode.md) |
| D-059 | Universal quit via MainModel-level 'q' interception | 2026-03-08 | Centralizes quit logic; views don't need individual 'q' handlers; `isTextInputActive()` guards text input views | [Story 36.3](../stories/36.3.story.md) |
| D-060 | Content pre-styling for door selection contrast | 2026-03-08 | Style content before theme Render(); avoids modifying each theme; uses Bold/Faint + DoubleBorder for structural emphasis | [Story 36.1](../../docs/stories/36.1.story.md) |
| D-061 | Replace softprops/action-gh-release with gh CLI | 2026-03-08 | Eliminates third-party supply chain risk; gh CLI is GitHub-maintained and pre-installed on runners | [Story 0.31](../stories/0.31.story.md) |
| D-062 | Protected GitHub environment for release secrets | 2026-03-08 | Scopes signing/deployment secrets to release jobs only; requires manual environment setup by repo owner | [Story 0.31](../stories/0.31.story.md) |
| D-063 | Post-GoReleaser signing job for stable releases | 2026-03-09 | Reuses proven alpha signing pipeline; doesn't replace GoReleaser; fail-open (unsigned is current baseline) | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-064 | Single Apple Developer ID for both channels | 2026-03-09 | Certificate is per-team; bundle ID differentiates; no need for separate certs | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-065 | `vars.ALPHA_TAP_ENABLED` toggle, default OFF | 2026-03-09 | Matches `SIGNING_ENABLED` pattern; conscious activation; prevents premature formula pushes | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-066 | Two separate formulas (no shared template) | 2026-03-09 | GoReleaser controls stable formula; different structures; copying > dependency | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-067 | Alpha release verification via tap CI monitoring | 2026-03-09 | Mirrors stable release-verify.yml pattern; lightweight; tap CI is primary gate | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-068 | Keep last 30 alpha releases, delete older | 2026-03-09 | Prevents release pollution; formula always points to latest before cleanup; 2-6 days of history | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| D-069 | Rename Phase 5+ to Phase 4.5 (Active Governance) | 2026-03-08 | Better reflects evolutionary nature; not a new phase but an extension of existing governance | [ADR-0029](../ADRs/ADR-0029-governance-phase-renaming.md) |
| D-070 | Three-tier decision system for party mode | 2026-03-08 | Micro (single agent), Standard (party mode), Strategic (owner); scales governance to decision weight | [ADR-0030](../ADRs/ADR-0030-decision-tiers-for-party-mode.md) |
| D-071 | Biweekly sprint cadence for agent team | 2026-03-08 | Balances planning overhead with delivery velocity for AI agent workflow | [ADR-0031](../ADRs/ADR-0031-biweekly-sprint-cadence.md) |
| D-072 | BMAD files as primary tracker with ROADMAP sync | 2026-03-08 | Story files are source of truth; ROADMAP.md synced periodically; no external tracker needed | [ADR-0032](../ADRs/ADR-0032-work-tracking-bmad-files-with-roadmap-sync.md) |
| D-073 | CLI -> MCP -> iPhone implementation priority | 2026-03-07 | CLI most valuable for core persona; MCP enables AI workflows; iPhone has no validated demand | [Research](../../_bmad-output/planning-artifacts/next-phase-prioritization-research.md) |
| D-074 | MIT license for ThreeDoors | 2026-03-08 | Charm ecosystem alignment; maximum adapter freedom; zero maintenance; Homebrew compatible | [Research](../../_bmad-output/planning-artifacts/license-selection-research.md) |
| D-075 | Focus state via session-scoped +focus tags (Epic 27) | 2026-03-07 | Reuses existing tag infrastructure; no new Task model field needed | [Architecture](../../_bmad-output/planning-artifacts/architecture-daily-planning-mode.md) |
| D-076 | Energy level inferred from time-of-day as default (Epic 27) | 2026-03-07 | Reduces friction; user can override; morning=high, afternoon=medium, evening=low | [Architecture](../../_bmad-output/planning-artifacts/architecture-daily-planning-mode.md) |
| D-077 | Soft progress indicator for planning mode (Epic 27) | 2026-03-07 | Step counter + elapsed time; no hard timer pressure; aligns with anti-anxiety philosophy | [Architecture](../../_bmad-output/planning-artifacts/architecture-daily-planning-mode.md) |
| D-078 | Seasonal themes as standalone DoorTheme instances (Epic 33) | 2026-03-08 | Replacement model, not overlay; each season is self-contained with own render function | [Architecture](../../_bmad-output/planning-artifacts/architecture-seasonal-themes.md) |
| D-079 | SeasonalResolver as pure function (Epic 33) | 2026-03-08 | No interface or struct needed; date-in theme-name-out; trivially testable | [Architecture](../../_bmad-output/planning-artifacts/architecture-seasonal-themes.md) |
| D-080 | Single story (0.31) for combined CI/security hardening | 2026-03-08 | All four issues target same CI workflow; combined diff is small; four stories would be ceremony overhead | [Artifact](../../_bmad-output/planning-artifacts/ci-security-hardening-triage-party-mode.md) |
| D-081 | SOUL.md + CLAUDE.md restructuring for AI agent alignment | 2026-03-02 | Codifies project philosophy; enables consistent agent decisions; DRY story specs | [Research](../../_bmad-output/planning-artifacts/ai-tooling-findings-research.md) |
| D-082 | Fork as variant creation with ForkTask factory (Epic 31) | 2026-03-08 | Preserves text/context/effort/tags; resets status/timestamps; adds cross-reference via enrichment DB | [Artifact](../../_bmad-output/planning-artifacts/party-mode-expand-fork-2026-03-08.md) |
| D-083 | No property inheritance for subtasks (Epic 31) | 2026-03-08 | Each subtask is own unit of work; inheriting effort would quintuple estimates; misleading | [Artifact](../../_bmad-output/planning-artifacts/party-mode-expand-fork-2026-03-08.md) |
| D-084 | No auto-completion of parent on subtask completion (Epic 31) | 2026-03-08 | Show completion ratio instead; parent excluded from door rotation when it has children | [Artifact](../../_bmad-output/planning-artifacts/party-mode-expand-fork-2026-03-08.md) |
| D-085 | Sequential expand mode for subtask creation (Epic 31) | 2026-03-08 | Stay in expand input after Enter; show running count; only Esc exits; reduces friction | [Artifact](../../_bmad-output/planning-artifacts/party-mode-expand-fork-2026-03-08.md) |
| D-086 | Dedicated ViewHelp view mode for :help display | 2026-03-08 | Consistent with all existing informational views; FlashMsg is fundamentally wrong for help content | [Artifact](../../_bmad-output/planning-artifacts/help-display-redesign.md) |
| D-087 | ? global keybinding opens help from any view | 2026-03-08 | TUI standard (vim, less, lazygit); solves discoverability; no existing conflicts | [Artifact](../../_bmad-output/planning-artifacts/help-display-redesign.md) |
| D-088 | Bar rendered by MainModel, views unaware (Epic 39) | 2026-03-08 | Clean separation; no per-view changes needed; MainModel adjusts height | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-architecture.md) |
| D-089 | Overlay as boolean flag, not new ViewMode (Epic 39) | 2026-03-08 | Preserves user context; overlay is ephemeral; no ViewMode transition needed | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-party-mode.md) |
| D-090 | Compile-time keybinding registry, not config-driven (Epic 39) | 2026-03-08 | All bindings known at compile time; YAGNI for runtime registration | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-architecture.md) |
| D-091 | `h` toggles bar, `?` toggles overlay (Epic 39) | 2026-03-08 | Separates "persistent reference" (h) from "help me now" (?); both universally understood | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-party-mode.md) |
| D-092 | Bar defaults ON for new users (Epic 39) | 2026-03-08 | Progressive disclosure; bridges onboarding-to-mastery gap; power users press h to hide | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-ux-review.md) |
| D-093 | Bar is theme-independent, dim styling only (Epic 39) | 2026-03-08 | Bar is chrome, not content; themed bars would require updating every theme for non-core feature | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-ux-review.md) |
| D-094 | Auto-hide bar below 10 lines terminal height (Epic 39) | 2026-03-08 | Doors must have priority for screen space; bar is helpful but not essential | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-ux-review.md) |
| D-095 | Fix alpha formula template with `if OS.mac? && Hardware::CPU.arm?` pattern | 2026-03-09 | Matches stable formula fix; addresses root cause in ci.yml not homebrew-tap; minimal change | [Artifact](../../_bmad-output/planning-artifacts/issue-296-alpha-formula-ci-triage.md) |
| D-096 | Custom Lipgloss rendering for stats (no new dependencies in Phase 1) (Epic 40) | 2026-03-08 | Sparklines/bars trivial to build in-tree; ntcharts evaluated for heatmap only in Phase 2 | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-097 | Single view for Phase 1 stats, Tab navigation for Phase 2 (Epic 40) | 2026-03-08 | Fits 24 rows at 80 cols; Tab is simplest interaction; no routing changes needed | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-098 | Phase 1 scope: 3 stories (dashboard shell, sparklines, fun facts) (Epic 40) | 2026-03-08 | Clean dependencies; parallelizable (40.2 and 40.3 independent after 40.1) | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-099 | Fun facts content rules: observe, celebrate, frame gaps, no decline (Epic 40) | 2026-03-08 | SOUL.md alignment; testable via banned-words QA test; daily-seeded rotation | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-100 | Activity heatmap: 8 weeks, custom Unicode+color, Detail tab (Epic 40) | 2026-03-08 | 8 weeks fits 80 cols; custom for consistency; Detail tab keeps main view clean | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-101 | Animated counters: subtle 300-500ms, numbers only, once per entry (Epic 40) | 2026-03-08 | "Button feel" polish; not distracting; low CPU cost; ~16 frames at 30ms tick | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-102 | Independent stats palette (Phase 1), theme coupling (Phase 3) (Epic 40) | 2026-03-08 | Reduces Phase 1 scope; theme extension is cross-cutting change | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-103 | Trophy room deferred; milestones limited to 4 observation-language thresholds (Epic 40) | 2026-03-08 | Trophy room: high complexity, gamification risk; milestones: SOUL.md boundary | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-104 | Beautiful Stats Display assigned Epic 40 | 2026-03-08 | Originally Epic 39 but renumbered to avoid collision with Keybinding Display (Epic 39) | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| D-105 | Spacebar as Enter alias in doors view (Epic 39) | 2026-03-08 | 11-agent unanimous consensus; largest key = most common action; zero new state; consistent with onboarding spacebar behavior | [Artifact](../../_bmad-output/planning-artifacts/spacebar-action-debate.md) |
| D-106 | `.` as label scope separator | 2026-03-08 | Universal namespace separator; clean on GitHub; avoids `:` ambiguity | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-107 | 9 scopes, 27 total labels (trimmed from initial 35) | 2026-03-08 | Each survived consumer challenge; 6 net new vs current 21; cuts: type.ux, resolution.fixed, process.party-mode, 3 status labels, 3 agent labels | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-108 | `status.do-not-merge` label for merge-queue hard stop | 2026-03-08 | Merge-queue needs explicit stop signal beyond draft PRs | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-109 | Agent labels limited to `agent.envoy` + `agent.worker` only | 2026-03-08 | Only two agents need GitHub visibility labels; others coordinate via story files/messages | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-110 | Label migration via rename-first strategy | 2026-03-08 | Preserves label-issue associations during transition | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-111 | Implementation as separate story from research PR | 2026-03-08 | Unanimous; research PR should not apply changes | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| D-112 | Epic number reservation registry in BOARD.md | 2026-03-09 | Prevents parallel workers from claiming same epic number; lightweight alternative to persistent PM agent | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |
| D-113 | No housekeeping epic for governance syncs | 2026-03-09 | Governance syncs are routine doc maintenance, not feature work; adding an epic creates overhead for non-deliverable work | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |
| D-114 | Use `tea.WithAltScreen()` for full-terminal ownership | 2026-03-09 | Standard TUI pattern (lazygit, k9s, soft-serve); clean terminal lifecycle; task picker doesn't need scrollback | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-115 | Fixed header + Flex middle + Fixed footer layout model | 2026-03-09 | Proven Bubbletea pattern; clean separation; enables keybinding bar footer slot (D-088) | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-116 | Door height capped at 25 lines: `min(max(10, available * 0.5), 25)` | 2026-03-09 | Door metaphor requires proportions that feel like doors (D-031); prevents skyscraper doors on tall terminals | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-117 | 40/60 top/bottom padding split for vertical centering | 2026-03-09 | Perceptual centering — content slightly above mathematical center feels natural; standard OS dialog placement | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-118 | Help/stats/search views use full available terminal height | 2026-03-09 | Hardcoded `helpPageSize = 20` is a bug; wastes space on tall terminals | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-119 | Breakpoint-based graceful degradation for small terminals | 2026-03-09 | Invisible to users; progressive collapse without error messages; respects D-094 bar hiding | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-120 | Two-story implementation (MVP layout + follow-up refactor) | 2026-03-09 | MVP delivers 80% value; refactor separates header/footer from DoorsView cleanly | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-121 | Layout engine is prerequisite for Story 39.2 (keybinding bar) | 2026-03-09 | Layout engine provides footer slot; building bar first would require rework | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| D-122 | Global `:` command mode via MainModel-level interception | 2026-03-09 | Follows D-059/D-087 pattern; `isTextInputActive()` guard prevents conflicts; one change location; zero per-view modifications | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| D-123 | Custom lightweight command completion (not bubbles/list) | 2026-03-09 | Only 16 commands; prefix match sufficient; bubbles/list is heavyweight with fuzzy matching wrong for command completion; no new dependency | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| D-124 | Inline suggestion rendering for command autocomplete | 2026-03-09 | Consistent with SearchView pattern; push content down; no overlay infrastructure exists | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| D-125 | Inline hints as frame decoration (doorknob metaphor), separate from bar toggle | 2026-03-09 | Hints are onboarding scaffolding (auto-fade), bar is reference tool (manual toggle); Approach B preserves door metaphor; no runtime toggle key — `:hints` command only; avoids 39.4 `h` toggle collision | [Artifact](../../_bmad-output/planning-artifacts/default-tooltips-mode-party-mode.md) |
| D-126 | Session-based auto-fade for inline hints (default 5 sessions) | 2026-03-09 | Simpler than per-key tracking; 90% as effective; graceful dim at N-1 then disable at N; `:hints on` re-enables | [Artifact](../../_bmad-output/planning-artifacts/default-tooltips-mode-party-mode.md) |
| D-127 | Inline tooltips as Epic 39 stories 39.9-39.12, not a new epic | 2026-03-09 | Same keybinding registry data source (39.1); same config infrastructure; same toggle ecosystem; new epic would fragment discoverability | [Artifact](../../_bmad-output/planning-artifacts/default-tooltips-mode-party-mode.md) |
| D-128 | Scope 'q' quit to doors view only; sub-views treat 'q' as go-back (overrides Story 36.3) | 2026-03-09 | Matches TUI conventions (vim, lazygit, htop); 'q' = "close current thing"; quit at root, back in sub-views; SOUL.md "work with human nature" | [Triage](../../_bmad-output/planning-artifacts/issue-330-dashboard-q-triage.md) |

## Rejected

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-001 | SQLite enrichment layer | 2026-01-20 | Overhead exceeded benefit; file-based storage sufficient | [ADR-0016](../ADRs/ADR-0016-cancel-sqlite-enrichment-layer.md) |
| X-002 | ADRs as single canonical decision format | 2026-03-08 | Too heavyweight for micro-decisions; would create ADR sprawl | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
| X-003 | RFC process for decisions | 2026-03-08 | Too heavyweight; party mode already handles deliberation | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
| X-004 | Tag-based decision system | 2026-03-08 | High retrofit cost (135+ files); distributes source of truth | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
| X-005 | Automated ADR generation | 2026-03-08 | ADR creation requires significance judgment; auto-generated would be noisy | [Research](../../_bmad-output/planning-artifacts/decision-management-research.md) |
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
| X-026 | Move GoReleaser to macOS runner for stable signing | 2026-03-09 | Doesn't add signing by itself; loses native Linux builds | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| X-027 | Replace GoReleaser with custom pipeline | 2026-03-09 | Throws away changelog, archiving, formula push, checksums; massive scope increase | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| X-028 | Shared formula template for stable and alpha | 2026-03-09 | GoReleaser controls stable formula; different structures make sharing impractical | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| X-029 | Per-push `brew install` verification for alpha | 2026-03-09 | Expensive (macOS runner + Homebrew install time); tap CI provides equivalent coverage | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| X-030 | Default `ALPHA_TAP_ENABLED` to ON | 2026-03-09 | First push would attempt formula push before infrastructure is validated | [Artifact](../../_bmad-output/planning-artifacts/homebrew-dual-publish-course-correction.md) |
| X-031 | Project-level work tracking across repos | 2026-03-08 | Deferred — single-repo tracking sufficient for now; complexity not justified | [ADR-0033](../ADRs/ADR-0033-project-level-tracking-deferred.md) |
| X-032 | Custom keybinding remapping (Epic 39) | 2026-03-08 | Out of scope; adds config complexity for unvalidated need; YAGNI | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-party-mode.md) |
| X-033 | Animated bar show/hide transitions (Epic 39) | 2026-03-08 | No animation system exists; Lipgloss is static styling; instant response is more "deliberate" per SOUL.md | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-party-mode.md) |
| X-034 | Themed keybinding bar per door theme (Epic 39) | 2026-03-08 | Bar is infrastructure chrome, not content; would require modifying all themes for non-core feature | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-ux-review.md) |
| X-035 | Single `?` key for both bar toggle and overlay (Epic 39) | 2026-03-08 | Conflates "persistent reference" with "help me now"; separate keys (h/?) are clearer | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-party-mode.md) |
| X-036 | Overlay as new ViewMode (Epic 39) | 2026-03-08 | ViewMode transition clears previous view; overlay is ephemeral context-preserving layer | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-architecture.md) |
| X-037 | Runtime keybinding registration / config-driven bindings (Epic 39) | 2026-03-08 | All bindings are known at compile time; runtime flexibility adds complexity for zero current benefit | [Artifact](../../_bmad-output/planning-artifacts/keybinding-display-architecture.md) |
| X-038 | Shared formula template for stable and alpha (Issue #296) | 2026-03-09 | GoReleaser controls stable; prior decision X-028; YAGNI for two instances | [Artifact](../../_bmad-output/planning-artifacts/issue-296-alpha-formula-ci-triage.md) |
| X-039 | Fix alpha formula directly in homebrew-tap (Issue #296) | 2026-03-09 | Would be overwritten on next CI push; root cause is in ThreeDoors ci.yml | [Artifact](../../_bmad-output/planning-artifacts/issue-296-alpha-formula-ci-triage.md) |
| X-040 | ntcharts as primary charting library for stats (Epic 40) | 2026-03-08 | Adds dependency when Lipgloss can build sparklines/bars in-tree; evaluate for heatmap only | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-041 | Trophy Room with door-sliding animation (Epic 40) | 2026-03-08 | High complexity, uncertain payoff, "trophy" has gamification connotations, stats should be accessible not hidden | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-042 | Gamified milestones ("Century Club unlocked!") (Epic 40) | 2026-03-08 | SOUL.md: "no gamification, no guilt"; achievement language creates extrinsic motivation | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-043 | Burnout indicators in TUI (Epic 40) | 2026-03-08 | Too judgmental for user-facing display; keep in MCP for AI agent consumption only | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-044 | Productivity scores/grades in stats (Epic 40) | 2026-03-08 | SOUL.md: "not a productivity report"; assigning scores implies judgment | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-045 | Braille pattern characters for high-res charts (Epic 40) | 2026-03-08 | Inconsistent rendering across terminals and fonts; Unicode blocks are universally safe | [Artifact](../../_bmad-output/planning-artifacts/beautiful-stats-party-mode.md) |
| X-046 | Spacebar as context-sensitive select+confirm | 2026-03-08 | Adds state-dependent behavior diverging from Enter; creates door-1 bias when no selection; ambiguity generates bug reports | [Artifact](../../_bmad-output/planning-artifacts/spacebar-action-debate.md) |
| X-047 | Spacebar as quick-complete from doors view | 2026-03-08 | Destructive action behind most easily-hit key; accidental completions; requires undo infrastructure that doesn't exist | [Artifact](../../_bmad-output/planning-artifacts/spacebar-action-debate.md) |
| X-048 | Spacebar as knock/peek tooltip | 2026-03-08 | Adds step between user and task; new rendering mode needed; scope disproportionate to value | [Artifact](../../_bmad-output/planning-artifacts/spacebar-action-debate.md) |
| X-049 | Spacebar as door cycle (tab equivalent) | 2026-03-08 | Adds parallel selection system alongside a/w/d; new cycle state tracking; confusing with existing positional keys | [Artifact](../../_bmad-output/planning-artifacts/spacebar-action-debate.md) |
| D-112 | Browser URL as primary bug report submission (zero-auth) | 2026-03-09 | Zero-config; uses existing browser session; URL query params for pre-filled issue | [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) |
| D-113 | Ring buffer breadcrumbs (50 entries, count-bounded) | 2026-03-09 | Fixed memory; sufficient context for bug reports; no persistence without user action | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md) |
| D-114 | Allowlist-only privacy for bug reports (capture-level filtering) | 2026-03-09 | Defense in depth — can't leak what was never captured; tea.KeyRunes never recorded | [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) |
| D-115 | Mandatory preview before bug report submission | 2026-03-09 | SOUL.md trust alignment; user sees exactly what will be sent; friend-helping-friend feel | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md) |
| D-116 | Bug report target repo hardcoded to arcaven/ThreeDoors | 2026-03-09 | Single-product reporter; YAGNI for configuration; forks can change in code | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md) |
| D-117 | Sort search results in filterTasks(), not in GetAllTasks() | 2026-03-09 | Consumer-side sort has zero blast radius; other callers don't need stable order; rejected: GetAllTasks sort (unnecessary cost), cached order (premature optimization), ordered data structure (over-engineered) | [Triage](../../_bmad-output/planning-artifacts/issue-334-search-jumping-triage.md) |
| D-128 | Adopt bubbles/viewport to replace 3 custom scroll implementations | 2026-03-09 | Standardizes behavior, adds mouse wheel, reduces maintenance (3 impls → 1 dep); rejected: keep custom (higher maintenance) | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-129 | Adopt bubbles/spinner for async operation feedback | 2026-03-09 | Trivial effort, eliminates "is it frozen?" uncertainty; deterministic testing | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-130 | Adopt lipgloss.JoinVertical + Place for layout composition | 2026-03-09 | Layout QoL, eliminates manual padding math and `\n` concatenation | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-131 | Harmonica door transitions via spike-first approach | 2026-03-09 | Spring physics delivers SOUL.md "physical objects" promise; testing risk requires validation first; extends prior deferral X-008 | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-132 | Reject bubbles/list for ThreeDoors selection UIs | 2026-03-09 | 3-door constraint is intentional; generic list fights core design philosophy | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-133 | Reject bubbles/textarea, table, filepicker, timer, help, huh, wish | 2026-03-09 | Each solves problems ThreeDoors doesn't have or contradicts SOUL.md values | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| D-134 | New Epic 41: Charm Ecosystem Adoption & TUI Polish | 2026-03-09 | 5-7 cohesive stories; too many for infra backlog, too few for multiple epics; P2 priority | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |

## Rejected

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-059 | OAuth device flow for bug report auth | 2026-03-09 | Too complex for bug reports; browser URL achieves zero-auth goal without token management | [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) |
| X-060 | gh CLI for bug report submission | 2026-03-09 | External dependency; SOUL.md discourages requiring external tools | [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) |
| X-061 | Time-bounded breadcrumb buffer | 2026-03-09 | Variable memory; hard to reason about edge cases (idle vs active); count-bounded is simpler | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md) |
| X-062 | Blocklist approach for bug report privacy | 2026-03-09 | Blocklists risk leaks from new data types; allowlist at capture level is defense in depth | [Research](../../_bmad-output/planning-artifacts/in-app-bug-reporting-research.md) |
| X-063 | Configurable target repo for bug reports | 2026-03-09 | YAGNI; ThreeDoors reports to ThreeDoors; adds config complexity for no validated need | [Party Mode](../../_bmad-output/planning-artifacts/in-app-bug-reporting-party-mode.md) |

| X-050 | `epic.N` per-epic labels | 2026-03-08 | GitHub milestones serve this purpose; 40+ labels is bloat | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-051 | `sprint.*` labels | 2026-03-08 | No fixed sprints in ThreeDoors workflow | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-052 | `effort.*` Fibonacci labels | 2026-03-08 | Effort tracked in story files, not issues | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-053 | Full 5-agent `agent.*` label set | 2026-03-08 | Agent workloads visible through other signals; only envoy + worker need labels | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-054 | `status.in-review`, `status.approved`, `status.changes-requested` labels | 2026-03-08 | GitHub review states queryable via API; labels duplicate native state | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-055 | `status.merge-ready` and `status.ci-failing` labels | 2026-03-08 | Merge-queue evaluates composite conditions directly; labels would need sync bot | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-056 | `::` (GitLab-style) as label separator | 2026-03-08 | Looks unusual on GitHub; `.` is more universal | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-057 | `/` as label scope separator | 2026-03-08 | Conflicts with path references | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-058 | `process.party-mode` label | 2026-03-08 | Party mode is a process step, not an issue state; covered by `scope.needs-decision` | [Artifact](../../_bmad-output/planning-artifacts/scoped-labels-party-mode.md) |
| X-071 | Adopt bubbles/list for custom cursor-based UIs | 2026-03-09 | Fights 3-door constraint; generic list adds selection modes ThreeDoors doesn't want; prior X-068 rejected for commands | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-072 | Adopt bubbles/textarea for multi-line input | 2026-03-09 | Contradicts friction-reduction philosophy; single-step captures are intentional | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-073 | Adopt bubbles/table for data display | 2026-03-09 | No tabular data in ThreeDoors; app shows 3 tasks, not spreadsheets | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-074 | Adopt bubbles/help instead of custom help system | 2026-03-09 | Keybinding registry integration (D-090) too deep; bubbles help too simple for multi-category view-aware system | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-075 | Adopt huh for form building | 2026-03-09 | Over-engineered for at-most 2-step input flows; textinput sufficient | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-076 | Adopt wish for SSH/remote terminal access | 2026-03-09 | SOUL.md: "Local-First, Privacy-Always"; no remote terminal mode needed | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-077 | Adopt bubbles/filepicker for file selection | 2026-03-09 | Path input in onboarding is fine as textinput; filepicker adds UI complexity for rare operation | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-078 | Adopt bubbles/timer/stopwatch for session display | 2026-03-09 | Session tracking is backend JSONL, not displayed; visible timers create pressure (anti-SOUL.md) | [Audit](../../_bmad-output/planning-artifacts/bubbletea-feature-audit-party-mode.md) |
| X-059 | Stay in normal scrollback buffer (no AltScreen) | 2026-03-09 | Scrollback not useful for task picker; every serious TUI uses AltScreen; dead space below app is accidental | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-060 | Unlimited door height growth (no cap) | 2026-03-09 | Produces absurd 60+ line "skyscraper" doors on tall terminals; violates door metaphor proportions (D-031) | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-061 | Fixed door height regardless of terminal size | 2026-03-09 | Ignores terminal size entirely; wastes space on normal terminals; proportional with cap is better | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-062 | Proportional zones layout (all regions scale) | 2026-03-09 | Over-complicated for the need; header/footer are fixed-size content; flex middle is sufficient | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-063 | Fill space below doors with new content (mini-stats, context) | 2026-03-09 | SOUL.md: "Show less"; filling space ≠ filling with stuff; whitespace IS the design | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-064 | "Terminal too small" warning messages | 2026-03-09 | Hostile UX; degradation should be invisible; users don't need to know about layout math | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-065 | Single large story for full layout implementation | 2026-03-09 | Too much scope risk; MVP (AltScreen + layout + door cap) delivers 80% value alone | [Artifact](../../_bmad-output/planning-artifacts/full-terminal-layout-party-mode.md) |
| X-066 | Per-view `:` handler for command mode | 2026-03-09 | Would require modifying 17 view Update() methods; higher maintenance burden; same approach rejected for `q` in X-024 | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| X-067 | Selective `:` command mode (add to "a few more" views) | 2026-03-09 | Half-measures create inconsistent behavior; users would need to memorize which views support `:`; global is simpler | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| X-068 | `bubbles/list` for command autocomplete | 2026-03-09 | Heavyweight for 16 items; fuzzy matching wrong for command completion (users expect prefix matching from vim muscle memory) | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| X-069 | Overlay-style dropdown for command suggestions | 2026-03-09 | No overlay rendering system exists; all existing patterns use inline rendering; would require Z-ordering infrastructure | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| X-070 | Argument-level completion for commands (MVP) | 2026-03-09 | Only 2 commands benefit (`:insights mood\|avoidance`, `:goals edit`); disproportionate complexity for MVP | [Artifact](../../_bmad-output/planning-artifacts/global-command-mode-analysis.md) |
| X-071 | Exemption list for 'q' quit (Option A for #330) | 2026-03-09 | Growing maintenance list; wrong default (quit is dangerous, should require opt-in not opt-out) | [Triage](../../_bmad-output/planning-artifacts/issue-330-dashboard-q-triage.md) |
| X-072 | Sub-view consumes 'q' before universal handler (Option B for #330) | 2026-03-09 | Architecturally impossible — universal handler at line 910 fires before view delegation at line 921 | [Triage](../../_bmad-output/planning-artifacts/issue-330-dashboard-q-triage.md) |
| X-073 | Different key for universal quit (Option D for #330) | 2026-03-09 | 'q' is THE standard TUI quit key; problem is scope not key choice; changing it violates muscle memory | [Triage](../../_bmad-output/planning-artifacts/issue-330-dashboard-q-triage.md) |

## Epic Number Registry

> **Purpose:** Prevents epic number collisions when multiple agents work in parallel. Check this table before assigning a new epic number. PM is the authority for allocating numbers (per standing orders). See D-112.

| Epic | Feature | Allocated | Status |
|------|---------|-----------|--------|
| 41 | Charm Ecosystem Adoption & TUI Polish | 2026-03-09 | Proposed (pending PM approval) |
| 42 | *(next available)* | — | — |

**Rules:**
1. Before creating a new epic, check this table for the next available number
2. Reserve the number here FIRST, before creating story files or updating ROADMAP.md
3. Only PM (or supervisor acting as PM) may allocate epic numbers
4. Completed epics (0-38) are not listed here — see ROADMAP.md Completed Epics table
5. Active epics (39-40) are already allocated — do not reuse

## Superseded

| ID | Original Decision | Date | Superseded By | Link |
|----|-------------------|------|---------------|------|
| | *No superseded decisions* | | | |
