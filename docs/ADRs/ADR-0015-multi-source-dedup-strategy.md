# ADR-0015: Multi-Source Dedup Strategy

- **Status:** Accepted
- **Date:** 2026-02-01
- **Decision Makers:** Design decision C6
- **Related PRs:** #84 (Story 13.1), #109 (Story 13.2), #143 (Story 13.2 integration)
- **Related ADRs:** ADR-0006 (TaskProvider Interface)

## Context

When the same logical task exists in multiple providers (e.g., a Jira ticket and an Obsidian note referencing the same work), duplicates appear in the unified task pool. False-positive dedup (merging unrelated tasks) is destructive and worse than showing duplicates.

## Considered Options

1. **Manual linking only** — User explicitly links tasks across providers
2. **Auto-detect with confirmation** — Fuzzy match on title/description, prompt user
3. **Eager auto-merge** — Automatically merge above a similarity threshold
4. **SourceRef-based linking** — Tasks linked via explicit cross-references in metadata

## Decision

Use **both**:
- **Auto-detect with confirmation** (Option B) for user-facing dedup — fuzzy matching surfaces candidates, user confirms
- **SourceRef-based linking** (Option D) for programmatic dedup — adapters populate cross-references in task metadata

## Rationale

- Auto-detect catches duplicates users might miss
- User confirmation prevents false-positive merges
- SourceRef linking provides reliable programmatic cross-referencing
- Two strategies complement each other: fuzzy matching for discovery, SourceRef for certainty

## Implementation

- Similarity detection uses Levenshtein distance on task titles
- Source attribution badges in TUI show which provider each task came from
- Cross-reference tracking (Story 6.2) stores explicit links between tasks
- Duplicate detection UI surfaces candidates with confidence scores

## Consequences

### Positive
- No false-positive merges — user always confirms
- SourceRef links are reliable and bidirectional
- Cross-provider task relationships are visible in the TUI

### Negative
- User confirmation adds friction for legitimate duplicates
- Fuzzy matching has compute cost proportional to task count squared
- SourceRef requires adapter support — not all providers can store cross-references
