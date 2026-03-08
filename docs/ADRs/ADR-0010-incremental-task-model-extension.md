# ADR-0010: Incremental Task Model Extension

- **Status:** Accepted
- **Date:** 2026-01-20
- **Decision Makers:** Design decision C1
- **Related PRs:** #91 (Story 3.5.2), #139 (Story 21.1), #151 (Story 21.3)
- **Related ADRs:** ADR-0003 (YAML Persistence), ADR-0014 (Auto-Migrate Schema)

## Context

Multiple epics (12, 13, 19, 20, 21) required new fields on the Task struct: DueDate, Priority, SourceRefs, FieldVersions. Three strategies were considered for evolving the task model.

## Considered Options

1. **Encode in Context field** — Stuff DueDate/Priority into the existing `Context` string
2. **Incremental struct extension** — Add fields one at a time as each epic needs them
3. **All-at-once model v2** — Single migration adding all planned fields

## Decision

**Incremental struct extension** (Option B). Each epic adds only the fields it needs, with its own migration if required.

## Rationale

- Aligns with YAGNI principle — don't add fields until an epic needs them
- Each migration is small and well-tested
- Avoids large blast radius of an all-at-once schema change
- Context field encoding (Option A) would be fragile and limit querying
- Multiple small migrations are easier to debug than one large one

## Fields Added Incrementally

| Field | Added By | Epic | Purpose |
|-------|----------|------|---------|
| DueDate | Story 12.1 | 12 | Calendar awareness |
| Priority | Story 19.2 | 19 | Jira field mapping |
| SourceRefs | Story 21.3 | 21 | Multi-source tracking |
| FieldVersions | Story 21.3 | 21 | Property-level conflict resolution |

## Consequences

### Positive
- Each epic owns its migration — clear responsibility
- Small migrations are easy to review and test
- Backward-compatible reads (new fields have zero values)
- No need to predict future field requirements

### Negative
- Multiple schema bumps over time (v1 → v2 → v3...)
- Each migration needs its own test coverage
- Auto-migration code accumulates (see ADR-0014)
