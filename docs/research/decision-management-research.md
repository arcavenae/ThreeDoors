# Decision Capture and Management Strategy — Research

**Date:** 2026-03-08
**Status:** Research Complete — Recommendations Ready
**Author:** Research Worker (eager-deer)

---

## 1. Problem Statement

ThreeDoors generates decisions across multiple workflows — party modes, triage pipelines, BMAD sessions, research spikes, and ad-hoc work. These decisions live scattered across:

| Location | Count | Content Type |
|----------|-------|-------------|
| `docs/ADRs/` | 28 | Formal architecture decision records |
| `docs/research/` | 21 | Research spikes and analysis |
| `_bmad-output/planning-artifacts/` | 33 | Party mode results, architecture reviews |
| `docs/stories/` | 135 | Story files with embedded decisions |
| PR descriptions/comments | 210+ | Inline decisions, trade-off discussions |
| Supervisor memory | 1 | Standing orders, process decisions |

**Consequences of the current state:**
- Completed research has no decision artifact (e.g., OAuth/go-gh research exists only in conversation history)
- Rejected options get re-proposed because no one recorded WHY they were rejected
- Workers lack context on prior decisions, leading to re-litigation
- No single place to see the "state of knowledge" — what's decided, what's pending, what's being researched

## 2. Current State Analysis

### What Works Well

**ADRs (docs/ADRs/):** The 28 existing ADRs follow a solid format — Status, Date, Context, Decision, Consequences. They have a README.md index organized by category. ADR format is well-suited for permanent architectural records.

**Party Mode Artifacts:** Some (like `issue-218-party-mode-consensus.md`) include a "Key Decisions" summary table at the bottom. This is the ideal extractable format — but it's not consistent across all artifacts.

**Research Docs:** Well-written but inconsistent. Some end with recommendations, others just present findings without a clear "so what?" conclusion.

### What's Missing

1. **Decision lifecycle tracking** — No way to see that a question became research, research produced a recommendation, and the recommendation is awaiting a decision
2. **Rejected options registry** — Party mode artifacts save adopted AND rejected approaches (per standing order), but there's no consolidated view
3. **Pending decisions surface** — Research that completes without a decision just sits there with no follow-up mechanism
4. **Cross-reference system** — No IDs or links connecting decisions across documents

## 3. Approaches Evaluated

### 3.1 ADRs as Single Canonical Format

**Concept:** Expand ADR usage to cover ALL decisions, not just architectural ones.

**Pros:** Single format, well-understood, existing tooling and index.

**Cons:** ADRs are heavyweight (~40 lines each). Creating an ADR for "we chose method X over Y in story 17.3" is overkill. Would create ADR sprawl — hundreds of ADRs diluting the value of the collection. ADRs are also retrospective (documenting what was decided) rather than tracking the lifecycle (question → research → decision).

**Verdict:** Keep ADRs for significant architectural decisions. Not suitable as the universal format.

### 3.2 Decision Log / Register

**Concept:** A structured table (spreadsheet-like) tracking all decisions with columns for ID, date, status, description, rationale, links.

**Pros:** Compact, scannable, complete view of all decisions.

**Cons:** A flat table doesn't capture the lifecycle. A decision in "Active Research" and a "Decided" item have very different needs.

**Verdict:** Useful as a component, but needs lifecycle awareness.

### 3.3 RFC (Request for Comments) Process

**Concept:** Formal proposal documents circulated for review before decisions are made.

**Pros:** Good for large organizations with async human reviewers.

**Cons:** Our "reviewers" are AI agents who respond in seconds, not days. The ceremony of an RFC process (draft → review period → approval) doesn't match our velocity. We already have party mode for multi-agent discussion.

**Verdict:** Too heavyweight. Party mode already serves the deliberation function.

### 3.4 Knowledge Kanban Board

**Concept:** A single markdown file with kanban-style columns tracking the lifecycle of decisions from question to resolution. Columns: Open Questions → Active Research → Pending Recommendations → Decided → Rejected.

**Pros:** Lifecycle-aware. Single file = single read for full state. Markdown = git-tracked, AI-readable, human-reviewable. Items move between columns as they progress. Surfaces pending items (research done, no decision yet).

**Cons:** Requires discipline to keep updated. Can get long over time (mitigated by archiving old entries).

**Verdict:** Best fit for our multi-agent context. Recommended as the primary component.

### 3.5 Tag-based System

**Concept:** Embed decision tags in story files and artifacts; sweep into a generated index.

**Pros:** Decisions stay close to their context.

**Cons:** Requires retrofitting 135+ story files. Distributes source of truth. Generated index is a view, not a source — creates sync issues.

**Verdict:** High retrofit cost, distributed truth. Rejected.

### 3.6 Automated ADR Generation

**Concept:** An agent periodically scans artifacts and auto-creates ADRs.

**Pros:** Reduces manual work.

**Cons:** ADR creation requires judgment about significance. Auto-generated ADRs would be low quality and create noise.

**Verdict:** Automate discovery (flagging undocumented decisions), not creation.

## 4. Recommendation: Three-Component System

Based on party mode consensus (see `_bmad-output/planning-artifacts/decision-management-party-mode.md`), we recommend:

### Component 1: Knowledge Board (`docs/decisions/BOARD.md`)

A single markdown file with kanban-style columns:

```markdown
## Open Questions
| ID | Question | Date | Owner | Context |
|----|----------|------|-------|---------|
| Q-001 | Should we use go-gh for GitHub API access? | 2026-03-05 | — | Came up in CLI work |

## Active Research
| ID | Topic | Date | Owner | Link |
|----|-------|------|-------|------|
| R-001 | OAuth flow for GitHub integration | 2026-03-06 | worker | docs/research/... |

## Pending Recommendations
| ID | Recommendation | Date | Source | Link | Awaiting |
|----|---------------|------|--------|------|----------|
| P-001 | Use go-gh library for GitHub API | 2026-03-07 | research spike | docs/research/... | Owner sign-off |

## Decided
| ID | Decision | Date | Rationale | Link |
|----|----------|------|-----------|------|
| D-001 | Go as primary language | 2025-11-07 | Single binary, TUI ecosystem | ADR-0001 |

## Rejected
| ID | Option | Date | Why Rejected | Link |
|----|--------|------|-------------|------|
| X-001 | SQLite enrichment layer | 2026-01-20 | Overhead exceeded benefit | ADR-0016 |
```

**Key properties:**
- Items move between columns as they progress (Q → R → P → D or X)
- Each entry has a unique ID for cross-referencing
- Links point to supporting documents (ADRs, research, artifacts)
- Owned by supervisor, updated by any agent via PR
- One read gives full state of all knowledge in the project

### Component 2: Standardized Artifact Endings

All party mode artifacts must include a **Decisions Summary** table:

```markdown
## Decisions Summary
| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| ... | Adopted/Rejected | ... | ... |
```

All research docs must include a **Recommendations** section.

This makes extraction to the board mechanical and consistent.

### Component 3: Board Hygiene Sweep

A periodic task (weekly or milestone-triggered) that:

1. Scans `_bmad-output/planning-artifacts/` for files not referenced in BOARD.md
2. Scans `docs/research/` for research without a corresponding board entry
3. Flags stale entries (Active Research > 2 weeks without update)
4. Identifies ADR candidates from significant decided items
5. Reports findings to supervisor — does NOT auto-create anything

### Integration with Existing Systems

| System | Role | Relationship to Board |
|--------|------|----------------------|
| ADRs (`docs/ADRs/`) | Permanent archive of significant architectural decisions | Board "Decided" entries link to ADRs when applicable |
| Research docs (`docs/research/`) | Detailed findings and analysis | Board "Active Research" / "Pending Recommendation" entries link here |
| Party mode artifacts | Discussion records with decisions | Board entries created from Decisions Summary tables |
| Story files | Implementation context | Stories reference board IDs for decision context |
| CLAUDE.md | Worker instructions | Add instruction to check BOARD.md before implementing |

### Decision Supersession

When a new decision overrides an old one:
- Old board entry moves to a "Superseded" section
- Forward-reference to the new decision added
- Matches the pattern ADRs already use (Status: Superseded by ADR-XXXX)

### Discoverability

Add one line to CLAUDE.md:

> **Before implementing, check `docs/decisions/BOARD.md`** for relevant prior decisions, rejected options, and active research.

Every worker reads CLAUDE.md. Every worker checks the board. No additional tooling required.

## 5. Implementation Path

### Immediate (this research)
- [x] Research approaches and reach consensus
- [x] Document findings and recommendations
- [x] Save party mode artifact

### Next Steps (separate stories)
1. Create `docs/decisions/BOARD.md` with initial structure and seed entries from existing ADRs
2. Update CLAUDE.md with board-checking instruction
3. Backfill: extract decisions from existing party mode artifacts into board entries
4. Create board hygiene sweep task definition
5. Update party mode workflow to require Decisions Summary table in artifacts

## 6. Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| Knowledge Board as central dashboard | Adopted | Lifecycle-aware, single file, markdown = zero infrastructure | Database register, tag-based system, separate tool |
| Three tiers: Board + ADR + Artifacts | Adopted | Different weights for different decision types | ADRs as single format (too heavyweight for micro-decisions) |
| Supervisor owns board, agents propose | Adopted | Quality control via PR review | Fully automated (no judgment gate) |
| Sweep agent reports, doesn't auto-create | Adopted | ADR creation requires significance judgment | Automated ADR generation |
| Standardized artifact endings | Adopted | Makes extraction mechanical | Free-form artifacts (current state) |
| CLAUDE.md instruction for discoverability | Adopted | Minimal change, maximum reach | Separate tooling, search systems |
| RFC process | Rejected | Too heavyweight; party mode already handles deliberation | — |
| Tag-based system | Rejected | High retrofit cost, distributed truth | — |
