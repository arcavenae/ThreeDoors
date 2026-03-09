# Planning Artifact Sprawl Audit & Consolidation

**Date:** 2026-03-09
**Author:** zealous-platypus (worker agent)
**Scope:** All research, analysis, and planning artifacts across the ThreeDoors codebase

---

## Problem Statement

Planning artifacts were scattered across 10+ locations, making findability poor, SWEEP automation harder, and new agents unsure where to put outputs. The `docs/decisions/BOARD.md` + `SWEEP.md` system works well for indexing decisions, but the artifacts themselves lacked a canonical home.

---

## Pre-Consolidation Inventory

| Location | File Count | Content Types |
|----------|-----------|---------------|
| `_bmad-output/planning-artifacts/` | 74 | Party mode outputs, architecture docs, sprint change proposals, epic breakdowns, audits, triage notes |
| `_bmad-output/implementation-artifacts/` | 19 | Story-level implementation specs |
| `docs/research/` | 31 | Research reports, draft configs, spike output |
| `docs/analysis/` | 4 | Signing investigation, CI audit, PR gap analysis |
| `docs/spikes/` | 2 | Config schema, LLM decomposition |
| `docs/spike-reports/` | 1 | Apple Notes integration |
| `docs/course-corrections/` | 1 | Theme rendering fixes |
| `docs/` root | 8 | Scattered research/analysis artifacts |
| `scripts/spike/` | 11 | Executable PoC code (kept in place) |
| **Total moved** | **47** | |

### docs/ Root Files Moved

| File | Original Commit | New Name |
|------|----------------|----------|
| `design-decisions-needed.md` | `b1034b6` | `design-decisions-needed-analysis.md` |
| `brainstorming-session-results.md` | `ca90fc4` | `brainstorming-session-results-party-mode.md` |
| `comprehensive-analysis.md` | `5bd695d` | `comprehensive-analysis.md` |
| `source-tree-analysis.md` | `5bd695d` | `source-tree-analysis.md` |
| `DELIVERABLES-SUMMARY.md` | `b4489c7` | `deliverables-summary-analysis.md` |
| `existing-documentation-inventory.md` | `5bd695d` | `existing-documentation-inventory-analysis.md` |
| `sprint-status-report.md` | `daa839e` | `sprint-status-report-2026-03-02.md` |
| `project-scan-report.json` | `5bd695d` | `project-scan-report.json` |

### docs/ Root Files Kept (Governance/Project Docs)

| File | Reason |
|------|--------|
| `validation-decision-rubric.md` | Project governance — validation framework |
| `validation-gate-results.md` | Project governance — actual validation outcomes |
| `user-guide.md` | End-user documentation |
| `development-guide.md` | Developer onboarding |
| `envoy-operations.md` | Operational reference |
| `adapter-developer-guide.md` | Developer reference |
| `branch-protection.md` | Operational reference |
| `issue-tracker.md` | Active state file |
| `bmm-workflow-status.yaml` | Active state file |

---

## Party Mode Consensus

**Participants:** PM (John), Analyst (Mary), Architect (Winston)

### Decisions

| # | Decision | Adopted | Rejected Alternatives |
|---|----------|---------|----------------------|
| 1 | **Taxonomy**: Three categories — Planning (deliberation), Research (investigation), Implementation (execution). Sub-types via naming convention, not directories. | Three flat categories | Per-document-type directories (creates dead-end folders with 1-2 files each) |
| 2 | **Directory structure**: Flat `_bmad-output/planning-artifacts/` absorbs all scattered dirs | Single flat directory | Sub-directories within planning-artifacts (recreates sprawl); keeping docs/research as parallel canonical location (splits artifacts across two trees) |
| 3 | **docs/ root**: Move research/analysis artifacts to planning-artifacts; keep governance docs in docs/ | Clear governance/research split | Moving everything (governance docs belong with the project); keeping all in docs/ root (sprawl continues) |
| 4 | **scripts/spike/**: Executable PoC code stays in scripts/spike/; prose reports move to planning-artifacts | Code/prose separation | Moving code to docs (code and prose don't mix) |
| 5 | **Prevention**: CLAUDE.md rule + agent definition updates + SWEEP enforcement + worker dispatch checklist | Multi-layer prevention | Mandatory metadata headers only (over-engineering; made recommended instead) |
| 6 | **Naming convention**: `{topic}-{type}.md` with controlled type vocabulary; date-prefix only for temporal artifacts | Topic-first naming | Date-first naming (hurts topic discovery); free-form types (prevents automation) |

### Controlled Type Vocabulary

Artifact filenames should end with one of these suffixes:
`-research`, `-party-mode`, `-architecture`, `-analysis`, `-spike`, `-sprint-change`, `-epic-breakdown`, `-triage`, `-course-correction`, `-audit`, `-ux-review`

---

## Directories Eliminated

| Directory | Files | Destination |
|-----------|-------|-------------|
| `docs/research/` | 31 | `_bmad-output/planning-artifacts/` |
| `docs/analysis/` | 4 | `_bmad-output/planning-artifacts/` |
| `docs/spikes/` | 2 | `_bmad-output/planning-artifacts/` |
| `docs/spike-reports/` | 1 | `_bmad-output/planning-artifacts/` |
| `docs/course-corrections/` | 1 | `_bmad-output/planning-artifacts/` |

---

## Cross-Reference Updates

Updated all actionable links in:
- `docs/decisions/BOARD.md` — All `../research/` links → `../../_bmad-output/planning-artifacts/`
- `docs/decisions/SWEEP.md` — Scan target updated; historical note added
- `docs/decisions/README.md` — Research doc location updated
- `README.md` — Research link updated
- `agents/project-watchdog.md` — Monthly sweep path updated
- `docs/index.md` — Moved file links updated
- Story files in `docs/stories/` — Research references updated
- PRD files in `docs/prd/` — Research references updated
- ADR files in `docs/ADRs/` — Research references updated
- Cross-references within `_bmad-output/planning-artifacts/` — Sibling references simplified
- Cross-references in `_bmad-output/implementation-artifacts/` — Updated

**Preserved as-is:** Historical/narrative references in audit reports, decision research examples, and story descriptions that describe the state at time of writing.

---

## Prevention Recommendations

### 1. CLAUDE.md Rule (Applied)
Added artifact placement rule to CLAUDE.md specifying `_bmad-output/planning-artifacts/` as the canonical destination.

### 2. Agent Definition Updates (Applied)
Updated `agents/project-watchdog.md` sweep path.

### 3. Remaining Work (Not in Scope)
- Worker dispatch template in supervisor MEMORY.md should specify artifact output directory
- Consider adding metadata headers to artifacts (recommended, not mandatory)
- Future SWEEP runs should validate naming convention compliance

---

## Opportunity Notes (Not Implemented)

- Some planning artifact filenames don't follow the `{topic}-{type}.md` convention (e.g., `epics.md`, `goreleaser-draft.yml`). A bulk rename could improve consistency but has high cross-reference churn.
- The `_bmad-output/implementation-artifacts/` directory could potentially merge into planning-artifacts, but its files serve a distinct purpose (story-level specs vs. research/deliberation) and the split is working.
- `scripts/spike/` has its own README that could be enriched with links to the corresponding prose reports now in planning-artifacts.
