# Party Mode Consensus: Decision Capture and Management Strategy

**Date:** 2026-03-08
**Participants:** John (PM), Winston (Architect), Amelia (Dev), Quinn (QA), Bob (SM), Murat (Test Architect), Mary (Analyst), Paige (Tech Writer)
**Topic:** How to better capture, consolidate, and track decisions across ThreeDoors

## Verdict

**Three-component system:** Knowledge Board + Standardized Artifact Format + Periodic Board Hygiene Sweep. ADRs remain the permanent archive for significant architectural decisions; the board is the living dashboard that ties everything together.

## Problem Statement

Decisions are scattered across 5+ locations (ADRs, research docs, party mode artifacts, story files, PR comments, memory files). Completed research has no artifact trail. Rejected options get re-proposed. Workers lack context on prior decisions.

## Adopted Approach

### 1. Knowledge Board (`docs/decisions/BOARD.md`)

A single markdown file with kanban-style columns as tables:

| Column | Purpose | Example |
|--------|---------|---------|
| Open Questions | Unanswered questions needing research | "Should we use go-gh for GitHub API?" |
| Active Research | Research in progress with owner/link | "OAuth flow — see docs/research/..." |
| Pending Recommendation | Research done, awaiting decision | "go-gh complete, recommends X — needs sign-off" |
| Decided | Final decisions with rationale link | "ADR-0021: MCP Server" |
| Rejected | Explicitly rejected with WHY | "SQLite enrichment — overhead, see ADR-0016" |

Each entry has: ID, Title, Date, Owner/Source, Status, Links to supporting docs.

**Ownership:** Supervisor owns the board. Any agent can propose entries via PR. Items move between columns as they progress through the lifecycle.

**Discoverability:** Add to CLAUDE.md: "Check `docs/decisions/BOARD.md` before implementing for relevant prior decisions, rejected options, and active research."

### 2. Standardized Artifact Format

All party mode artifacts must end with a **Decisions Summary** table:

```markdown
## Decisions Summary
| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
```

All research docs must end with a **Recommendations** section stating what was found and what action is recommended.

This makes extraction to the board mechanical — an agent reads the summary, creates board entries.

### 3. Board Hygiene Sweep (Periodic Agent Task)

A periodic sweep (weekly or after major milestones) that:
- Scans `_bmad-output/planning-artifacts/` for files not referenced in BOARD.md
- Scans `docs/research/` for research without a corresponding board entry
- Flags stale entries (e.g., "Active Research" older than 2 weeks)
- Reports findings to supervisor — does NOT auto-create ADRs

Can be run as a `/loop` skill or manual task.

### Decision Supersession

When a new decision overrides an old one:
- Old entry moves to "Superseded" with a forward-reference to the new decision
- Same pattern ADRs already use (Status: Superseded by ADR-XXXX)

### ADR Relationship

ADRs remain the permanent archive for **significant architectural decisions**. The board tracks everything — questions, research, micro-decisions, ADR-level decisions alike. ADRs are created when a decision is significant enough to warrant permanent documentation. The sweep agent flags ADR candidates but does not auto-create them.

## Rejected Approaches

### Full RFC (Request for Comments) Process
**Why rejected:** Too heavyweight for our team size and velocity. RFCs assume multiple human reviewers with async comments over days. Our AI agents decide in minutes. The ceremony-to-value ratio is wrong.

### Tag-based System (decisions tagged in stories, swept into index)
**Why rejected:** Requires modifying 135+ existing story files to add tags. Distributes the source of truth across too many files. Retrofit cost is high. Centralized board is simpler and more discoverable.

### Database-backed Decision Register (SQLite, API)
**Why rejected:** Adds infrastructure for a problem solvable with a markdown file. We already rejected SQLite enrichment (ADR-0016). Same principle: don't add tooling when a flat file works.

### Separate Decision Tool/App
**Why rejected:** Against project philosophy (SOUL.md) — keep things simple, don't add external tools. Markdown in git IS the tool. Git provides history, diffing, and collaboration for free.

### ADRs as Single Canonical Format for ALL Decisions
**Why rejected:** ADRs are too formal for micro-decisions ("should we use method X or Y?"). Would create ADR sprawl and dilute the value of the ADR archive. Reserve ADRs for significant architectural choices; use lightweight board entries for everything else.

### Automated ADR Generation
**Why rejected:** ADRs require architectural judgment about significance, scope, and consequences. An agent can draft them, but deciding what deserves an ADR requires supervisor-level judgment. The sweep agent should flag candidates, not auto-create.

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Single BOARD.md file, not distributed tags | Centralized = discoverable, grep-friendly, one read for full state |
| Markdown tables, not database | Git-tracked, AI-readable, human-reviewable, zero infrastructure |
| Three tiers (Board + ADR + Artifacts), not one format | Different weights for different decisions; prevents ADR sprawl |
| Supervisor owns board, agents propose entries | Quality control via PR review; prevents board pollution |
| Sweep agent reports, doesn't auto-create | Decision-making about decisions requires judgment |
| Add CLAUDE.md instruction to check board | Minimal change, maximum discoverability for all workers |

## Implementation Notes

- Create `docs/decisions/BOARD.md` with initial columns and seed entries from existing ADRs
- Update CLAUDE.md with board-checking instruction
- Update party mode exit workflow to include Decisions Summary table
- Create board hygiene sweep task (future story)
- Backfill: extract decisions from existing party mode artifacts (future story)
