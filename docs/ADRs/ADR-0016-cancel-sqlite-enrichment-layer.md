# ADR-0016: Cancel SQLite Enrichment Layer

- **Status:** Accepted
- **Date:** 2026-01-20
- **Decision Makers:** Design decision H1
- **Related PRs:** #53 (Story 5.1 — initial setup, later superseded)
- **Related ADRs:** ADR-0003 (YAML Persistence)

## Context

Epic 6 proposed a SQLite enrichment database for storing cross-references, metadata, and query-optimized task data. The question was whether this complexity was justified.

## Decision

**Cancel** the SQLite enrichment layer. File-based storage with in-memory indexing is sufficient.

## Rationale

- YAGNI — file-based storage handles current and foreseeable scale
- Personal task management rarely exceeds thousands of tasks
- In-memory loading of YAML files is fast enough (<100ms for local adapters)
- SQLite would add CGO dependency (or pure Go SQLite with performance trade-offs)
- Cross-reference tracking implemented directly in task metadata instead

## Consequences

### Positive
- No database dependency — simpler deployment and testing
- No schema migration tooling for a separate database
- Fewer failure modes — no database corruption scenarios
- Binary remains pure Go — no CGO required

### Negative
- No SQL query capabilities for complex cross-reference analysis
- All tasks must fit in memory (acceptable for personal use)
- If scale requirements change dramatically, this decision may need revisiting
