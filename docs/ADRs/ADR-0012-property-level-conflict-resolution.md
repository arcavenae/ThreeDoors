# ADR-0012: Property-Level Conflict Resolution

- **Status:** Accepted
- **Date:** 2026-01-20
- **Decision Makers:** Design decision C3
- **Related PRs:** #151 (Story 21.3), #157 (Story 21.4)
- **Related ADRs:** ADR-0010 (Incremental Task Model), ADR-0013 (Offline-First)

## Context

When tasks are modified in multiple providers simultaneously, conflicts arise. The resolution strategy affects data integrity and user experience.

## Considered Options

1. **Task-level last-write-wins (LWW)** — Entire task replaced by most recent write
2. **Property-level LWW** — Each field tracks its own `(value, updatedAt, actor)`
3. **Task-level LWW now, property-level later** — Start simple, migrate when needed

## Decision

**Property-level LWW** (Option B). Each field tracks its own version information via `FieldVersions` on the Task struct.

The initial recommendation was Option C (start simple), but the decision was to go directly to property-level to avoid a painful future migration once data existed.

## Rationale

- Prevents data loss from concurrent edits to different fields
- Example: editing title on one device while editing notes on another — both edits preserved
- Avoids a future migration that would be harder once real user data exists
- FieldVersions map (`map[string]FieldVersion`) is extensible to new fields

## Consequences

### Positive
- No data loss from concurrent edits to different fields
- Clean foundation for multi-source sync
- Field-level history enables future "show changes" features

### Negative
- ~3x storage per task due to per-field version metadata
- More complex merge logic compared to simple LWW
- All adapters must correctly populate FieldVersions on writes
