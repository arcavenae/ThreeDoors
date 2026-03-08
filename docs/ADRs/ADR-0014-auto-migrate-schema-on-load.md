# ADR-0014: Auto-Migrate Schema on Load

- **Status:** Accepted
- **Date:** 2026-02-15
- **Decision Makers:** Design decision C7
- **Related PRs:** #151 (Story 21.3)
- **Related ADRs:** ADR-0010 (Incremental Task Model), ADR-0003 (YAML Persistence)

## Context

Schema evolution (e.g., `SourceProvider` string → `[]SourceRef` array) creates backward-compatibility challenges. Existing task files must be readable after upgrades.

## Considered Options

1. **Auto-migrate on load** — Detect schema version, convert old format, write new format
2. **Migration CLI command** — `threedoors migrate` one-time conversion
3. **Dual-read support** — Read both old and new formats, write new only

## Decision

**Auto-migrate on load** (Option A). When loading task files, detect the schema version and transparently convert old formats to the current version.

## Rationale

- Least user friction — no manual migration step required
- Users who update the binary get automatic migration on next run
- Migration code is co-located with the data loading logic
- Tested via unit tests with fixture files in old formats

## Consequences

### Positive
- Seamless upgrades — users never run a migration command
- Old task files from backups are automatically compatible
- Migration tested as part of normal load path

### Negative
- Migration code accumulates permanently in the codebase
- Migration must handle every historical schema version
- First load after upgrade is slightly slower (migration + write-back)
- Hard to roll back if migration has bugs (mitigated by atomic writes preserving old file until rename)
