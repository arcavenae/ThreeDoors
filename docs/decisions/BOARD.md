# Knowledge Decisions Board

> Action-oriented dashboard for ThreeDoors project decisions. Shows only items needing attention.
>
> **Full history:** [ARCHIVE.md](ARCHIVE.md) | **Epic Number Registry:** [EPIC_REGISTRY.md](EPIC_REGISTRY.md)

---

## Needs Decision (Human Input Required)

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-001 | Migrate from Makefile to Justfile | 2026-03-04 | Research spike | [Analysis](../../_bmad-output/planning-artifacts/makefile-vs-justfile-analysis.md) | Owner sign-off — *Implementation completed via Story 0.59 (PR #768)* |

## Under Investigation

| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| R-002 | PRD Post-Reconstruction Quality Audit — 3 HIGH, 4 MEDIUM issues | 2026-03-15 | PM Validation | [Report](../../_bmad-output/planning-artifacts/prd-validation-report-2026-03-15.md) — **Open.** HIGH: stale next-steps.md, chaotic phase numbering in product-scope.md, missing v2.0 change log entry. MEDIUM: stale BOARD.md epic registry, incomplete user journeys, no YAML frontmatter, stale checklist-results-report.md. |

## Recently Decided (Last 30 Days)

> Entries from ~March 2026. For the complete decided history (D-001–D-182), see [ARCHIVE.md](ARCHIVE.md).

| ID | Decision | Date | Rationale | Link |
|----|----------|------|-----------|------|
| D-160 | Shell script for CI metrics, not GitHub Action (Story 0.37) | 2026-03-10 | No external deps beyond `gh`; immediately runnable by retrospector agent | [Research](../../_bmad-output/planning-artifacts/ci-churn-reduction-research.md) |
| D-161 | Project-watchdog batches governance sync PRs (resolves Q-003) | 2026-03-10 | One-per-story creates PR fatigue; batching reduces churn and conflicts | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |
| D-162 | Workers update ONLY story files; project-watchdog owns all planning docs (resolves Q-004) | 2026-03-10 | Eliminates concurrent edit conflicts between workers and project-watchdog | [Investigation](../../_bmad-output/planning-artifacts/epic-39-governance-sync-investigation.md) |
| D-163 | HHMMSS timecode for alpha version sorting (Story 0.47) | 2026-03-10 | UTC HHMMSS between date and SHA for chronological SemVer sorting | [Artifact](../../_bmad-output/planning-artifacts/alpha-versioning-improvement.md) |
| D-164 | Gemini CLI + OAuth as research execution layer (Epic 54 rearchitecture) | 2026-03-11 | Eliminates Python dependency, paid API key; OAuth is free tier | [Research](../../_bmad-output/planning-artifacts/gemini-cli-oauth-research.md) |
| D-164b | ClickUp bidirectional sync follows Todoist pattern (Story 63.3) | 2026-03-13 | Reuse CircuitBreaker + WALProvider infrastructure | Story 63.3 |
| D-165 | Remove quit intercept (Session Reflection) — reverses Story 3.6 | 2026-03-11 | Violates SOUL.md: adds friction on quit; three principles violated | [Proposal](../../_bmad-output/planning-artifacts/sprint-change-proposal-2026-03-11-remove-quit-intercept.md) |
| D-166 | CI Optimization Phase 1: Docker E2E push-only, benchmark path filtering | 2026-03-11 | Pure CI config changes; PR wall clock 3m33s→2m08s | [Synthesis](../../_bmad-output/planning-artifacts/ci-test-optimization/05-synthesis-optimization-roadmap.md) |
| D-167 | LLM CLI Services: two-layer Services+Backends, CLISpec, auto-discovery (Epic 57) | 2026-03-11 | `Complete(ctx, prompt)` sufficient for P0; CLISpec enables 5-min provider addition | [Synthesis](../../_bmad-output/planning-artifacts/llm-services-architecture/synthesis.md) |
| D-168 | External daemon monitoring for supervisor context degradation | 2026-03-11 | Degrading supervisor can't self-assess; daemon monitors transcript size | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/synthesis-supervisor-shift-handover.md) |
| D-169 | Cold start supervisor replacement (no hot/warm standby) | 2026-03-11 | Hot doubles cost; warm has unsolvable upgrade problem; 60-90s gap acceptable | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-3-standby-feasibility.md) |
| D-170 | Daemon-maintained rolling snapshot (not supervisor-serialized) | 2026-03-11 | Minimizes load on degraded supervisor; daemon collects from external sources | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-4-state-serialization.md) |
| D-171 | Hybrid shift clock: time floor + usage ceiling | 2026-03-11 | 30-min floor + compression ceiling; three-tier thresholds; anti-oscillation guard | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-1-context-detection-shift-clock.md) |
| D-172 | Role-based agent addressing for supervisor handover | 2026-03-11 | Workers address "supervisor" role, not instance name; daemon maps role→instance | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-5-failure-modes-recovery.md) |
| D-173 | Three-layer depth system for door visual redesign (Epic 56) | 2026-03-11 | Background fill + bevel lighting + gradient shadow address all three failure modes | [Research](../../_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md) |
| D-174 | MkDocs + Material for GitHub Pages user guide (Epic 61) | 2026-03-11 | Markdown-native; GoReleaser precedent; built-in search | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| D-175 | Supervisor handover startup: state file + worker pings + READY signal | 2026-03-12 | Speed (instant orientation) + accuracy (fresh confirmation) | [Schema](../architecture/shift-handover-state-schema.md) |
| D-176 | File-based fallback inbox for agent messaging (Epic 62) | 2026-03-12 | Workaround for multiclaude identity registration bug; durable JSONL inbox | [Proposal](../../_bmad-output/planning-artifacts/sprint-change-proposal-2026-03-12-retrospector-reliability.md) |
| D-177 | Recommendation queue file instead of direct BOARD.md writes (Epic 62) | 2026-03-12 | Separates detection from persistence; respects authority model | [Party Mode](../../_bmad-output/planning-artifacts/retrospector-reliability-party-mode.md) |
| D-178 | JSON checkpoint for retrospector state persistence (Epic 62) | 2026-03-12 | Context exhaustion after ~45 PRs loses state; periodic checkpoint + restore | [Party Mode](../../_bmad-output/planning-artifacts/retrospector-reliability-party-mode.md) |
| D-179 | PRD coverage gap formalization — Epics 63, 64, Story 5.3 | 2026-03-13 | Three PRD features lacked corresponding epics/stories; all P2/Phase 5 | [Analysis](../../_bmad-output/planning-artifacts/prd-coverage-gap-analysis.md) |
| D-180 | Standalone tea.Program wrapping ConnectWizard for CLI (Story 45.6) | 2026-03-13 | Wizard already implements tea.Model; thin wrapper avoids logic duplication | [Research](../../_bmad-output/planning-artifacts/connect-wizard-gap-research.md) |
| D-181 | Move adapter registration before CLI/TUI routing branch (Epic 66) | 2026-03-13 | Registration only ran in TUI path; CLI had empty registry | [Audit](../../_bmad-output/planning-artifacts/unwired-features-audit.md) |
| D-182 | BOARD.md redesign: action-oriented sections, two-file split with ARCHIVE.md + EPIC_REGISTRY.md (Epic 68) | 2026-03-15 | Reduces dashboard noise; 150+ historical entries moved to archive; epic registry separated for cleaner allocation | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |

## Recently Rejected (Last 30 Days)

> Entries from ~March 10–15, 2026. For the complete rejected history (X-001–X-123), see [ARCHIVE.md](ARCHIVE.md).

| ID | Option | Date | Why Rejected | Link |
|----|--------|------|--------------|------|
| X-094 | Interactive repair wizard for doctor command (Epic 49) | 2026-03-10 | Too complex for v1; doctor should be non-interactive | [Research](../../_bmad-output/planning-artifacts/threedoors-doctor-research.md) |
| X-095 | Doctor as part of every command startup (Epic 49) | 2026-03-10 | Too slow; doctor is explicit command | [Research](../../_bmad-output/planning-artifacts/threedoors-doctor-research.md) |
| X-096 | Telemetry/crash reporting in doctor (Epic 49) | 2026-03-10 | Out of scope; doctor is local-only diagnostics | [Research](../../_bmad-output/planning-artifacts/threedoors-doctor-research.md) |
| X-097 | `health` and `doctor` coexist as separate commands (Epic 49) | 2026-03-10 | User confusion; health is new enough to absorb into doctor | [Research](../../_bmad-output/planning-artifacts/threedoors-doctor-research.md) |
| X-098 | Command name `check` or `diagnose` instead of `doctor` (Epic 49) | 2026-03-10 | `check` too generic; `diagnose` too verbose; `doctor` is established pattern | [Research](../../_bmad-output/planning-artifacts/threedoors-doctor-research.md) |
| X-099 | Network-native multiclaude (TCP/WebSocket/NATS transport) | 2026-03-10 | Requires significant core refactoring; premature before Tier 1-2 validation | [Investigation](../../_bmad-output/planning-artifacts/remote-collaboration-investigation.md) |
| X-100 | NATS/Redis as inter-agent message broker | 2026-03-10 | Adds external dependency; SSH achieves same result at current scale | [Problem Solving](../../_bmad-output/planning-artifacts/remote-collab-creative-problem-solving.md) |
| X-101 | Direct daemon socket exposure over TCP | 2026-03-10 | No auth layer; race conditions with daemon state | [Feasibility](../../_bmad-output/planning-artifacts/remote-collab-analyst-feasibility.md) |
| X-102 | File sync (rsync/Syncthing) for message directories | 2026-03-10 | Covers messages only; sync conflicts if both sides write | [Feasibility](../../_bmad-output/planning-artifacts/remote-collab-analyst-feasibility.md) |
| X-103 | Typing into tmux windows for agent communication | 2026-03-10 | Corrupts conversation state; unpredictable and dangerous | [Investigation](../../_bmad-output/planning-artifacts/remote-collaboration-investigation.md) |
| X-104 | Hot standby supervisor (Epic 56) | 2026-03-11 | Doubles API costs; split-brain risk | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-3-standby-feasibility.md) |
| X-105 | Warm standby supervisor (Epic 56) | 2026-03-11 | Claude can't hot-swap system prompts; kill+respawn anyway | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-3-standby-feasibility.md) |
| X-106 | Supervisor self-reporting for context degradation (Epic 56) | 2026-03-11 | Degrading agent can't trust own assessment | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-1-context-detection-shift-clock.md) |
| X-107 | Message replay for handover context (Epic 56) | 2026-03-11 | High token cost; lacks full context | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-2-worker-handover-protocol.md) |
| X-108 | Instant authority cutover at handover (Epic 56) | 2026-03-11 | Risks message loss and split-brain during transition | [Research](../../_bmad-output/planning-artifacts/supervisor-shift-handover/session-2-worker-handover-protocol.md) |
| X-109 | Full Corridor wall context for doors (Epic 56) | 2026-03-11 | Width cost of 6+ chars too expensive at min 15-char doors | [Research](../../_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md) |
| X-110 | Adaptive Depth terminal detection (Epic 56) | 2026-03-11 | Over-engineering; CompleteColor already handles fallbacks | [Research](../../_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md) |
| X-111 | Interior texture shade chars as wood grain (Epic 56) | 2026-03-11 | Competes with content readability; background color achieves mass | [Research](../../_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md) |
| X-112 | Braille patterns for subpixel door detail (Epic 56) | 2026-03-11 | Compatibility and accessibility concerns | [Research](../../_bmad-output/planning-artifacts/door-visual-redesign/party-mode-door-redesign.md) |
| X-113 | Hugo for docs site (Epic 61) | 2026-03-11 | Docsy theme more complex; GoReleaser chose MkDocs Material | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| X-114 | Jekyll for docs site (Epic 61) | 2026-03-11 | Slow builds, dated, not designed for technical docs | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| X-115 | Docusaurus for docs site (Epic 61) | 2026-03-11 | React/Node toolchain overkill for CLI tool docs | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| X-116 | GitHub Wiki for docs (Epic 61) | 2026-03-11 | Not version-controlled with repo; limited formatting | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| X-117 | Mix user-facing docs into existing `docs/` directory (Epic 61) | 2026-03-11 | 262 story files would create confusion and bloat nav | [Planning](../../_bmad-output/planning-artifacts/gh-pages-user-guide-plan.md) |
| X-118 | Standalone interactive prompts (huh forms) in CLI for connect wizard (Story 45.6) | 2026-03-13 | Duplicates logic in connect_wizard.go; double maintenance burden | [Research](../../_bmad-output/planning-artifacts/connect-wizard-gap-research.md) |
| X-119 | Status tags on single BOARD.md file (Epic 68) | 2026-03-15 | Doesn't reduce file size; tags add maintenance without structural improvement | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |
| X-120 | Database-backed board with YAML/JSON + generated markdown (Epic 68) | 2026-03-15 | Over-engineering; violates SOUL.md; D-029 chose zero infrastructure | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |
| X-121 | Per-epic decision files (Epic 68) | 2026-03-15 | Scatters decisions; cross-cutting decisions hard to find; 65+ small files | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |
| X-122 | Aggressive archival — delete entries older than 90 days (Epic 68) | 2026-03-15 | Git history is worst way to find past decisions; archive preserves searchability | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |
| X-123 | Three-tier system: Active + Reference + Archive (Epic 68) | 2026-03-15 | "Still relevant" distinction is subjective; reference tier would re-bloat | [Research](../../_bmad-output/planning-artifacts/board-redesign-research.md) |
