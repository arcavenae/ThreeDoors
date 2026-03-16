# Knowledge Decisions Board — Guide

## Purpose

The Knowledge Decisions Board (`BOARD.md`) is the central dashboard tracking the lifecycle of all project decisions in ThreeDoors. It provides a single place to see what's decided, what's being researched, what's pending, and what was rejected.

**ADRs remain the permanent archive** for significant architectural decisions. The board is the living dashboard that ties everything together — ADRs, research, party mode outcomes, and ad-hoc decisions.

## Board Structure

The board is an action-oriented dashboard showing only items needing attention. Historical entries age into the archive.

```
Needs Decision ──────────→ Recently Decided ──→ (30 days) ──→ ARCHIVE.md
Under Investigation ─────↗ Recently Rejected ──→ (30 days) ──→ ARCHIVE.md
```

**Supporting files:**
- **[ARCHIVE.md](ARCHIVE.md)** — Full decision history (D-001+, X-001+, S-001+)
- **[EPIC_REGISTRY.md](EPIC_REGISTRY.md)** — Epic number allocation registry (formerly tracked on BOARD.md)

### Section Definitions

| Section | Purpose | What Goes Here |
|---------|---------|----------------|
| **Needs Decision** | Human input required | Recommendations from completed research awaiting owner sign-off |
| **Under Investigation** | Topics currently being researched | Research spikes, technical evaluations, analysis in progress |
| **Recently Decided** | Decisions made in the last 30 days | Accepted decisions with rationale and links to supporting artifacts |
| **Recently Rejected** | Options rejected in the last 30 days | Alternatives considered and rejected, with documented reasoning |

Entries older than 30 days in Recently Decided or Recently Rejected are moved to ARCHIVE.md during the hygiene sweep.

## ID Scheme

Each board entry has a unique ID based on its section:

| Prefix | Section | Example |
|--------|---------|---------|
| `P-NNN` | Needs Decision | P-001 |
| `R-NNN` | Under Investigation | R-001 |
| `D-NNN` | Recently Decided / Archive | D-001 |
| `X-NNN` | Recently Rejected / Archive | X-001 |
| `S-NNN` | Superseded (Archive only) | S-001 |

IDs are sequential within each prefix. When an item moves between sections (e.g., an investigation produces a recommendation), it gets a new ID in the destination section. The old ID can be noted in the description for traceability.

**Note:** The epic number registry is now maintained separately in [EPIC_REGISTRY.md](EPIC_REGISTRY.md).

## Lifecycle Flow

A typical decision lifecycle:

1. **Investigation opened** (R-NNN) — A question or topic enters Under Investigation
2. **Recommendation made** (P-NNN) — Research produces a recommendation, moves to Needs Decision
3. **Decision finalized** (D-NNN) — Recommendation accepted, moves to Recently Decided
   - OR **Rejected** (X-NNN) — Option explicitly rejected, moves to Recently Rejected
4. **Archived** — After 30 days in Recently Decided/Rejected, entry moves to ARCHIVE.md

Not all items follow the full lifecycle. Many decisions go directly to Recently Decided (e.g., ADRs created during implementation).

## Relationship to ADRs

| Aspect | ADRs (`docs/ADRs/`) | Decisions Board (`docs/decisions/BOARD.md`) |
|--------|---------------------|---------------------------------------------|
| **Scope** | Significant architectural decisions | All decisions (architectural + tactical) |
| **Format** | Full document with context, options, consequences | Single-row table entry with link |
| **When created** | For decisions with broad impact | For any decision worth remembering |
| **Permanence** | Permanent archive | Living dashboard |

A significant decision typically has both:
- A **D-NNN** entry on the board (for discoverability)
- An **ADR** document (for full context and rationale)

## How to Propose Entries

1. **Open a PR** that modifies `BOARD.md`
2. **Add your entry** to the appropriate column
3. **Use the next available ID** in that column's sequence
4. **Include a link** to supporting documentation (ADR, research doc, artifact, PR)
5. **PR description** should explain why the entry is being added

## Artifact Format Requirements

All party mode artifacts (`_bmad-output/planning-artifacts/`) **must** include a standardized **Decisions Summary** table at the end of the document. This ensures decisions are extractable to the board mechanically.

### Canonical Format

```markdown
## Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| Use approach X | Adopted | Reason for adopting | Approach Y (reason), Approach Z (reason) |
| Skip feature F | Rejected | Not needed because... | — |
```

### Column Definitions

| Column | Content |
|--------|---------|
| **Decision** | Short description of what was decided |
| **Status** | `Adopted` or `Rejected` |
| **Rationale** | Why this option was chosen/rejected |
| **Alternatives Rejected** | Other options considered and why they were not chosen; use `—` if none |

### Rules

1. **Placement:** The Decisions Summary table must be the last major section in the artifact
2. **Completeness:** Every decision made during the discussion must appear in the table
3. **Both sides:** Include both adopted decisions AND explicitly rejected alternatives
4. **Linkable:** Decisions in the table should be extractable to `BOARD.md` entries without needing to read the full artifact
5. **Existing artifacts:** The following exemplars demonstrate the standardized format:
   - `issue-218-party-mode-consensus.md` (consensus/triage type)
   - `architecture-snooze-defer.md` (architecture type)
   - `door-appearance-party-mode.md` (design consensus type)

### Research Docs

Research documents (`_bmad-output/planning-artifacts/*-research.md`) should include a **Recommendations** section at the end with clear, actionable conclusions. Use the same Decisions Summary table format when the research produces concrete decisions.

## Board Hygiene Sweep

The board stays current through a periodic hygiene sweep process. The sweep scans for unindexed artifacts, orphaned research, stale entries, and ADR candidates — then reports findings to the supervisor for action. It never auto-creates entries.

See [`SWEEP.md`](SWEEP.md) for the full process definition, report format, and running instructions.

**Schedule:** Weekly or after major milestones. Any agent can run it.

## Founding Decision

The board itself was created based on research into decision management approaches. See:
- Research: [`_bmad-output/planning-artifacts/decision-management-research.md`](../../_bmad-output/planning-artifacts/decision-management-research.md)
- Board entry: D-029
