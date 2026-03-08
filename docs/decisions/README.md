# Knowledge Decisions Board — Guide

## Purpose

The Knowledge Decisions Board (`BOARD.md`) is the central dashboard tracking the lifecycle of all project decisions in ThreeDoors. It provides a single place to see what's decided, what's being researched, what's pending, and what was rejected.

**ADRs remain the permanent archive** for significant architectural decisions. The board is the living dashboard that ties everything together — ADRs, research, party mode outcomes, and ad-hoc decisions.

## Board Structure

The board uses kanban-style columns reflecting the decision lifecycle:

```
Open Questions → Active Research → Pending Recommendations → Decided
                                                           → Rejected
                                                           → Superseded
```

### Column Definitions

| Column | Purpose | What Goes Here |
|--------|---------|----------------|
| **Open Questions** | Unanswered questions needing investigation | Questions raised during implementation, triage, or review that don't have answers yet |
| **Active Research** | Topics currently being investigated | Research spikes, technical evaluations, competitive analysis in progress |
| **Pending Recommendations** | Research complete, awaiting decision | Recommendations from completed research that need owner sign-off |
| **Decided** | Finalized decisions | All accepted decisions — links to ADRs, research, or artifacts that document the rationale |
| **Rejected** | Options explicitly rejected | Alternatives considered and rejected, with documented reasoning to prevent re-proposal |
| **Superseded** | Decisions replaced by newer ones | Old decisions overridden by new ones, with forward-references to the replacement |

## ID Scheme

Each board entry has a unique ID based on its column:

| Prefix | Column | Example |
|--------|--------|---------|
| `Q-NNN` | Open Questions | Q-001 |
| `R-NNN` | Active Research | R-001 |
| `P-NNN` | Pending Recommendations | P-001 |
| `D-NNN` | Decided | D-001 |
| `X-NNN` | Rejected | X-001 |
| `S-NNN` | Superseded | S-001 |

IDs are sequential within each prefix. When an item moves between columns (e.g., a question becomes active research), it gets a new ID in the destination column. The old ID can be noted in the description for traceability.

## Lifecycle Flow

A typical decision lifecycle:

1. **Question raised** (Q-NNN) — Someone identifies an unanswered question
2. **Research started** (R-NNN) — Investigation begins, question moves to Active Research
3. **Recommendation made** (P-NNN) — Research produces a recommendation, awaiting sign-off
4. **Decision finalized** (D-NNN) — Recommendation accepted, moves to Decided
   - OR **Rejected** (X-NNN) — Option explicitly rejected with documented reasoning

Not all items follow the full lifecycle. Many decisions go directly to Decided (e.g., ADRs created during implementation).

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

## Founding Decision

The board itself was created based on research into decision management approaches. See:
- Research: [`docs/research/decision-management-research.md`](../research/decision-management-research.md)
- Board entry: D-029
