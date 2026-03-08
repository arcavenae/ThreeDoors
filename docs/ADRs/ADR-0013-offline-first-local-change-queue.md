# ADR-0013: Offline-First with Local Change Queue (WAL)

- **Status:** Accepted
- **Date:** 2026-02-01
- **Decision Makers:** Architecture review
- **Related PRs:** #62 (Story 11.1), #66 (Story 11.2), #85 (Story 11.3)
- **Related ADRs:** ADR-0011 (Sync Scheduler), ADR-0012 (Conflict Resolution)

## Context

ThreeDoors integrates with external task sources (Apple Notes, Obsidian, Jira, GitHub Issues) that may be temporarily unavailable. The application must remain fully functional offline.

## Decision

Implement an **offline-first architecture** with a Write-Ahead Log (WAL):

1. All task operations succeed locally immediately
2. Changes are recorded in a local change queue (WAL)
3. The sync scheduler replays queued changes when providers become available
4. Conflict resolution handles divergent states on reconnection

## Rationale

- Core UX must never be blocked by network availability
- Personal task management is a high-frequency, low-latency activity
- External providers (Jira API, GitHub API) have unpredictable availability
- WAL pattern is proven in databases and messaging systems

## Implementation

- WAL stored as append-only JSONL file
- Each entry records: operation type, task ID, field changes, timestamp, provider
- Replay is idempotent — safe to replay the same change multiple times
- Sync status indicator (Story 11.2) shows pending/synced/error state in TUI
- Conflict visualization (Story 11.3) surfaces merge conflicts to the user

## Consequences

### Positive
- App works identically online and offline
- No data loss from connectivity issues
- Sync errors are surfaced, not hidden
- Users can review pending changes before sync

### Negative
- WAL file grows until changes are synced and acknowledged
- Conflict resolution requires user interaction in some cases
- Stale data possible during extended offline periods
