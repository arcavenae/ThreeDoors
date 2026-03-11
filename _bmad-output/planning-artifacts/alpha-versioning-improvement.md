# Alpha Versioning Improvement — Planning Artifact

**Date:** 2026-03-10
**Status:** Decision made — ready for implementation

---

## Problem Statement

Alpha release versions have random ordering because the build identifier after `YYYYMMDD` is a git SHA prefix, which doesn't sort chronologically.

**Current scheme** (`.github/workflows/ci.yml` line 235):
```
VERSION="0.1.0-alpha.$(date -u +%Y%m%d).${GITHUB_SHA::7}"
TAG="alpha-$(date -u +%Y%m%d)-${GITHUB_SHA::7}"
```

**Example of the problem:** Two merges on 2026-03-10 produce:
- `0.1.0-alpha.20260310.abc1234` (merged at 09:00)
- `0.1.0-alpha.20260310.def5678` (merged at 14:30)

Per SemVer 2.0 §11.4, `abc1234` and `def5678` are alphanumeric identifiers compared lexically (ASCII order). `abc1234 < def5678` happens to be correct here, but SHA ordering is random — it has no relationship to chronological order.

This affects:
1. `brew upgrade` — Homebrew's version comparison may pick the wrong "latest"
2. GitHub release list sorting
3. Story 49.9 (channel-aware version checking) — comparing alpha versions requires reliable ordering
4. User confusion when `threedoors --version` shows a "newer" version that's actually older

## Options Evaluated

### Option A: HHMMSS Timecode (ADOPTED)

Insert UTC time-of-day after the date, before the SHA:

```
VERSION="0.1.0-alpha.$(date -u +%Y%m%d).$(date -u +%H%M%S).${GITHUB_SHA::7}"
TAG="alpha-$(date -u +%Y%m%d)-$(date -u +%H%M%S)-${GITHUB_SHA::7}"
```

**Result:** `0.1.0-alpha.20260310.143022.abc1234`

**SemVer analysis:**
- Identifiers: `alpha` (string/lexical), `20260310` (numeric), `143022` (numeric), `abc1234` (string/lexical)
- Date + time are both numeric → SemVer sorts them numerically → correct chronological order
- SHA becomes the 4th identifier — only consulted if two builds happen in the same second (impossible with CI)
- SHA remains for traceability (link back to commit)

**Homebrew analysis:**
- Homebrew `Version` class splits on `.` and `-`, compares segments left to right
- Adding a segment is fully compatible — more segments = more precision, not a breaking change
- Numeric segments compare numerically

**Pros:**
- One-line change to CI workflow
- Fully SemVer 2.0 compliant
- Sorts correctly everywhere (SemVer, Homebrew, lexical string sort, GitHub UI)
- No external state needed (no "last version" tracking)
- SHA preserved for traceability
- Backward compatible — new versions always sort after old ones (more identifiers = higher precedence when all prior identifiers match per SemVer §11.4.4)

**Cons:**
- Slightly longer version strings (6 more characters)
- Two `date -u` calls in the workflow (negligible)

### Option B: Auto-Incrementing Semantic Version (REJECTED)

Bump `0.1.0` → `0.2.0` → `0.3.0` based on some heuristic (conventional commits, PR labels, manual).

**Why rejected:**
1. **State tracking required** — Need to query GitHub API for last release tag during CI, adding a network dependency and failure mode to the build
2. **Complexity** — Conventional commit parsing is a whole subsystem (`standard-version`, `semantic-release`, etc.) — massive scope creep for alpha builds
3. **PR label approach** — Requires humans to remember labels; errors result in wrong version bumps
4. **Manual bumps** — Defeats the purpose of automation
5. **Premature** — Alpha is pre-release by definition; semantic versioning of pre-releases is overkill when the base version (`0.1.0`) isn't changing yet
6. **Doesn't solve the sort problem** — Even with `0.2.0-alpha`, same-day builds of the same version still need sub-ordering

### Option C: Unix Timestamp Instead of Date+Time (REJECTED)

Use epoch seconds: `0.1.0-alpha.1741609822.abc1234`

**Why rejected:**
1. **Human-unreadable** — `1741609822` means nothing at a glance; `20260310.143022` is immediately parseable
2. **Loses date grouping** — Can't quickly see "all builds from March 10" in a release list
3. **SemVer numeric overflow** — Some tools may truncate large numeric identifiers (epoch seconds are 10 digits)

### Option D: Zero-Padded Build Counter Per Day (REJECTED)

Track a per-day build counter: `0.1.0-alpha.20260310.001`

**Why rejected:**
1. **State tracking** — Need to persist and increment a counter somewhere (GitHub API query, file in repo, etc.)
2. **Race condition** — Two concurrent CI runs could collide on the counter
3. **More complex than HHMMSS** for no benefit — time is already a natural, collision-free counter

## Decision

**Adopt Option A: HHMMSS timecode insertion.** Decision D-163.

Single-line CI change, zero new dependencies, fully SemVer compliant, sorts correctly everywhere. The SHA suffix remains for commit traceability but no longer determines sort order.

**Auto-incrementing semantic version explicitly rejected** — premature for alpha builds, adds state tracking complexity, and doesn't even solve the sub-ordering problem within a version.

## Impact Assessment

### Files to Change
1. `.github/workflows/ci.yml` — lines 235-236 (version generation)
2. `internal/dist/version_test.go` — if any tests assert version format
3. `internal/cli/version_test.go` — if any tests assert version format

### Downstream Effects
- **Story 49.9 (channel-aware version checking):** Unblocked — alpha versions now sort reliably
- **Homebrew formula:** No changes needed — formula uses `version "${VERSION}"` which accepts any version string
- **GoReleaser:** Not affected — only used for tagged stable releases, not alpha builds
- **GitHub release cleanup:** The `tail -n +31` cleanup step works on creation order (not version order), so unaffected
- **Existing releases:** Old-format releases sort before new-format releases with the same date (fewer identifiers = lower precedence per SemVer §11.4.4). This is correct behavior.
