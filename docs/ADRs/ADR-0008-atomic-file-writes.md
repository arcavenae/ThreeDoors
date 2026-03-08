# ADR-0008: Atomic File Writes for Persistence

- **Status:** Accepted
- **Date:** 2025-11-07
- **Decision Makers:** Project founder
- **Related PRs:** #8, #91 (Story 3.5.2)
- **Related ADRs:** ADR-0003 (YAML Task Persistence)

## Context

ThreeDoors persists task state to YAML files. File corruption during writes (power loss, crash, concurrent access) would cause data loss. The application needs a safe write pattern.

## Decision

All file persistence uses the **atomic write pattern**: write to a `.tmp` file, `fsync`, then rename to the target path.

```
1. Write data to path.tmp
2. fsync(path.tmp)  — flush to disk
3. rename(path.tmp, path)  — atomic on POSIX filesystems
```

## Rationale

- POSIX `rename()` is atomic — readers always see either the old or new complete file
- `fsync` ensures data reaches disk before rename
- No file locking required — eliminates deadlock risks
- Pattern is well-established in databases and configuration management tools

## Consequences

### Positive
- Zero risk of partial writes or corrupted task files
- No file locking infrastructure needed
- Works correctly on all POSIX filesystems (HFS+, APFS)
- Simple to implement and test

### Negative
- Requires temporary file space (negligible for task files)
- Rename across filesystem boundaries would fail (not applicable — same directory)
- Slightly slower than direct write due to fsync (imperceptible for small files)
