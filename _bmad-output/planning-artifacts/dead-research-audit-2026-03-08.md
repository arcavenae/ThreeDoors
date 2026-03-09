# Dead Research Audit — 2026-03-08

**Auditor:** agent worker (dead-research-audit task)
**Scope:** All files in `_bmad-output/planning-artifacts/`, `docs/research/`, `docs/ADRs/`, cross-referenced against `docs/decisions/BOARD.md`

---

## Summary

| Category | Total | Tracked in BOARD.md | Untracked |
|----------|-------|---------------------|-----------|
| Planning artifacts (committed) | 59 | 24 | 35 |
| Planning artifacts (uncommitted) | 3 | 0 | 3 |
| Research docs | 31 | 6 | 25 |
| ADRs | 34 | 29 | 5 |
| **Total** | **127** | **59** | **68** |

Not all 68 untracked items require BOARD.md entries. Many are supporting artifacts (sprint change proposals, epic breakdowns, story-level party modes) that don't contain standalone decisions. The audit below classifies each and recommends action.

---

## Category 1: ADRs Missing from BOARD.md (5)

These are accepted architectural decisions with no BOARD.md entry. All should be added to the Decided section.

| ADR | Decision | Action |
|-----|----------|--------|
| ADR-0029 | Governance phase renaming (Phase 5+ to 4.5) | Add to Decided |
| ADR-0030 (decision-tiers) | Three-tier decision system for party mode | Add to Decided |
| ADR-0031 | Biweekly sprint cadence | Add to Decided |
| ADR-0032 | BMAD epic/story files as primary tracker with ROADMAP sync | Add to Decided |
| ADR-0033 | Project-level work tracking deferred | Add to Decided |

---

## Category 2: Research Docs with Decisions Not in BOARD.md (14)

These research docs reached conclusions that influenced project direction but aren't tracked on the board.

### Produced completed epics (decisions already executed, need retroactive BOARD entries)

| Research Doc | Outcome | Recommendation |
|-------------|---------|----------------|
| `cli-interface-design.md` | Epic 23 (CLI) — COMPLETE | Add to Decided: Cobra CLI alongside TUI (already D-022, but research doc not linked) |
| `llm-integration-mcp.md` | Epic 24 (MCP) — COMPLETE | Add to Decided: MCP server design (already D-021, but research doc not linked) |
| `door-themes-research.md` | Epic 17 — COMPLETE | Add to Decided: theme system design (already D-020, but research doc not linked) |
| `door-themes-analyst-review.md` | Fed into Epic 17 | No action — supporting doc for D-020 |
| `door-themes-party-mode.md` | Fed into Epic 17 | No action — supporting doc for D-020 |
| `code-signing-findings.md` | Epic 5 / Story 0.31 — COMPLETE | Add to Decided: CI signing infrastructure (already D-018, but research not linked) |
| `mood-correlation.md` | Epic 15 — COMPLETE | No action — research validated existing design, no standalone decision |
| `procrastination.md` | Epic 15 — COMPLETE | No action — research validated three-door count (already D-052) |
| `sync-architecture-scaling-research.md` | Epic 21 — COMPLETE | No action — produced ADR-0011 through ADR-0015 (already tracked) |
| `task-sync-analyst-brief.md` | Summary doc | No action — analyst brief summarizing other research |

### Produced decisions that need BOARD.md entries

| Research Doc | Key Decision/Recommendation | Status | Recommendation |
|-------------|---------------------------|--------|----------------|
| `next-phase-prioritization.md` | CLI -> MCP -> iPhone priority ordering | Executed (Epics 23, 24 done; 16 iceboxed) | Add to Decided |
| `license-selection-research.md` | MIT license for ThreeDoors | Executed (LICENSE file exists) | Add to Decided |
| `ai-tooling-findings.md` | SOUL.md + CLAUDE.md restructuring + custom skills | Executed (Epic 34 done) | Add to Decided |
| `multiclaude-auto-execution-research.md` | Shell script MVP for story dispatch | Executed (Epic 22 done) | Already D-026, no action |

### Research with unacted-upon recommendations

| Research Doc | Key Recommendation | Status | Recommendation |
|-------------|-------------------|--------|----------------|
| `ux-workflow-improvements-research.md` | Daily planning mode, snooze/defer, undo completion | Epics 27/28/32 exist but NOT STARTED | No action — already tracked as epics |
| `task-source-expansion-research.md` | Todoist, Jira, Reminders, GitHub Issues integrations | Epics 19/20/25/26 exist; 25 NOT STARTED | No action — already tracked |
| `persistent-agent-architecture-research.md` | PM + Architect as persistent agents | Epic 37 COMPLETE | No action — already D-044/D-045 |
| `persistent-agent-communication-investigation.md` | Agent restart protocol, OAuth workflow scope | Partially acted on | Add to Pending Recommendations: pr-shepherd definition update |
| `cross-repo-ci-strategy.md` | gh run list polling for cross-repo awareness | Decision captured as D-037 but research doc not linked | Link research doc to D-037 |
| `mobile-app-research.md` | SwiftUI + React Native evaluation | Epic 16 ICEBOXED | Already captured as X-006 |
| `ci-churn-reduction-research.md` | Path filtering + concurrency limits | Story 0.20 NOT STARTED | Already captured as D-036/D-055 |

---

## Category 3: Planning Artifacts Not in BOARD.md (35 committed + 3 uncommitted)

### Architecture docs with untracked decisions (3) — need BOARD.md entries

| Artifact | Untracked Decisions | Recommendation |
|----------|-------------------|----------------|
| `architecture-daily-planning-mode.md` | Focus via session-scoped tags; energy from time-of-day; soft progress indicator | Add to Decided (3 entries) |
| `architecture-seasonal-themes.md` | Seasonal themes as standalone instances; pure function resolver; auto-switch at init; config extension | Add to Decided (6 entries) |
| `architecture-persistent-agent-infrastructure.md` | Agent lifecycle, communication protocol, resource management | Already partially captured (D-044, D-045); link artifact |

### Party mode artifacts with untracked decisions (5)

| Artifact | Untracked Decisions | Recommendation |
|----------|-------------------|----------------|
| `party-mode-expand-fork-2026-03-08.md` | Fork semantics (variant creation); no property inheritance; no auto-completion; sequential expand mode | Add to Decided (4 entries) |
| `license-selection-party-mode.md` | MIT license adopted | Add to Decided |
| `ci-security-hardening-triage-party-mode.md` | Single story (0.31) for combined CI hardening | Already executed (Story 0.31 done) — add to Decided |
| `decision-management-party-mode.md` | Three-component decision system (board + format + hygiene) | Already captured as D-029; link artifact |
| `persistent-agents-course-correction-party-mode.md` | Confirmed 2 persistent + 2 cron agents | Already captured as D-044/D-045; link artifact |

### Epic breakdown docs (7) — no action needed

These are story breakdown docs, not decision artifacts. They operationalize decisions already tracked:
- `epic-28-snooze-defer.md`, `epic-29-task-dependencies.md`, `epic-30-linear-integration.md`
- `epic-31-expand-fork.md`, `epic-32-undo-task-completion.md`, `epic-34-soul-skills.md`
- `epics.md` (master epic breakdown)

### Story-level party mode artifacts (8) — no action needed

These are implementation-level party modes for specific stories. Decisions within them are scoped to their stories:
- `34.3-party-mode-dev-readiness.md`, `34.3-party-mode-test-readiness.md`, `34.4-party-mode-dry-cleanup.md`
- `35.2-classic-door-proportions.md`, `35.3-modern-door-proportions.md`, `35.4-scifi-door-proportions.md`
- `35.5-shoji-door-proportions.md`, `35.6-golden-file-accessibility.md`, `35.7-shadow-depth-effect.md`

### Sprint change proposals (12) — no action needed

Sprint change proposals are process documents, not decision artifacts. They propose scope changes that either got adopted (and produced epics/stories) or didn't:
- `sprint-change-proposal-2026-03-02.md`
- `sprint-change-proposal-2026-03-07.md` and 5 topic-specific variants
- `sprint-change-proposal-2026-03-08.md` and 5 topic-specific variants

### Uncommitted artifacts (3) — need triage

These exist in the main worktree but have not been committed:

| Artifact | Content | Recommendation |
|----------|---------|----------------|
| `envoy-scope-and-firewall-design.md` | Envoy agent boundaries, three-layer firewall | Add to Pending Recommendations: envoy firewall implementation |
| `issue-labeling-and-triage-strategy.md` | GitHub issue labeling taxonomy and triage flow | Add to Pending Recommendations: issue labeling implementation |
| `multiclaude-customizations-audit.md` | Inventory of ThreeDoors-specific multiclaude customizations | No BOARD action — informational audit |

---

## Category 4: Open Questions Review

### Q-001: Jira story points vs priority for effort mapping
**Status:** Still open. Epic 19 (Jira) is complete per ROADMAP. Check if this was resolved during implementation.
**Finding:** Epic 19 stories are marked complete. This question was likely resolved during implementation but BOARD.md was not updated.
**Recommendation:** Verify implementation and either move to Decided or close as moot (Jira adapter is done).

### Q-002: Jira multi-project JQL vs explicit project keys
**Status:** Same situation as Q-001 — Epic 19 is complete.
**Recommendation:** Same as Q-001.

---

## Category 5: Pending Recommendations Review

### P-001: Migrate from Makefile to Justfile
**Status:** Still pending. No story created. Makefile is still in use.
**Recommendation:** Keep as Pending Recommendation. Low priority — current Makefile works fine.

---

## BOARD.md Updates Made

The following entries were added to BOARD.md as part of this audit:

### New Decided entries
- D-069: Governance phase renaming (Phase 5+ to 4.5) — ADR-0029
- D-070: Three-tier decision system for party mode — ADR-0030
- D-071: Biweekly sprint cadence — ADR-0031
- D-072: BMAD files as primary tracker with ROADMAP sync — ADR-0032
- D-073: CLI -> MCP -> iPhone priority ordering
- D-074: MIT license for ThreeDoors
- D-075: Focus state via session-scoped tags (Epic 27)
- D-076: Energy level from time-of-day default (Epic 27)
- D-077: Soft progress indicator for planning mode (Epic 27)
- D-078: Seasonal themes as standalone DoorTheme instances (Epic 33)
- D-079: SeasonalResolver as pure function (Epic 33)
- D-080: Single story approach for CI/security hardening (Story 0.31)
- D-081: SOUL.md + CLAUDE.md restructuring for AI agent alignment (Epic 34)
- D-082: Fork as variant creation with ForkTask factory (Epic 31)
- D-083: No property inheritance for subtasks (Epic 31)
- D-084: No auto-completion of parent on subtask completion (Epic 31)
- D-085: Sequential expand mode for subtask creation (Epic 31)

### New Rejected entries
- X-031: Project-level work tracking (deferred per ADR-0033)

### New Pending Recommendations
- P-002: Envoy three-layer firewall implementation
- P-003: GitHub issue labeling taxonomy implementation
- P-004: pr-shepherd definition update (remove fork references)

### Open Questions updated
- Q-001 and Q-002 flagged for review (Jira epic is complete — questions may be resolved)

---

## Artifacts That Are NOT Dead Research

The following untracked items are intentionally not flagged because they are supporting documents, not standalone decision artifacts:
- 7 epic breakdown docs (operationalize existing decisions)
- 8 story-level party modes (scoped to individual stories)
- 12 sprint change proposals (process documents)
- 3 research summary/analyst briefs (consolidate other research)
- 2 research docs that validated existing designs without new decisions
