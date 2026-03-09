# Sync Architecture Scaling Research

> How ThreeDoors' sync architecture should evolve to support many simultaneous task sources reliably.

## Current Architecture Summary

ThreeDoors has a well-structured sync foundation built across Epics 11 and 13:

| Component | File | Purpose |
|-----------|------|---------|
| SyncEngine | `internal/core/sync_engine.go` | Three-way sync with change detection & conflict resolution |
| SyncState | `internal/core/sync_state.go` | Persists snapshots for three-way diff |
| WALProvider | `internal/core/wal_provider.go` | Offline queue with replay and exponential backoff |
| SyncLog | `internal/core/sync_log.go` | JSONL audit trail with rotation |
| SyncStatusTracker | `internal/core/sync_status.go` | Per-provider sync phase tracking for TUI |
| ConflictResolver | `internal/core/conflict_resolver.go` | Interactive conflict resolution (local/remote/both) |
| MultiSourceAggregator | `internal/core/aggregator.go` | Merges tasks from multiple TaskProviders |
| FallbackProvider | `internal/core/fallback_provider.go` | Graceful degradation chain |
| AdapterRegistry | `internal/core/registry.go` | Runtime adapter registration and factory pattern |

**Current adapters:** TextFile (YAML), Apple Notes (AppleScript), Obsidian (vault files).

**Current conflict strategy:** Task-level last-write-wins by `UpdatedAt` timestamp. Remote wins on ties.

**Current scheduling:** No dedicated scheduler. Sync runs on LoadTasks() calls. WAL replay triggers opportunistically when a provider becomes available.

---

## Scaling Challenges

### 1. Sync Frequency Management

**Problem:** With N providers, naive polling creates N concurrent sync loops. Each provider has different latency, rate limits, and reliability characteristics. A 30-second poll interval across 10 providers means 20 API/IPC calls per minute — wasteful when most providers have no changes.

**Current gap:** No sync scheduler exists. Sync is triggered by user actions (LoadTasks), not by a background cadence. This works for 1-3 providers but breaks down at scale — the user must interact with the app to discover remote changes.

### 2. Conflict Resolution at Scale

**Problem:** Task-level last-write-wins loses data. If a user changes the title in Apple Notes and the status in Obsidian, the current engine picks one version and discards the other entirely.

**Current gap:** The `SyncEngine.ResolveConflicts()` method compares `UpdatedAt` at the task level. There is no field-level tracking, so partial merges are impossible.

### 3. Cross-Provider Deduplication

**Problem:** The same task may appear in multiple providers (e.g., a task in both Apple Notes and Obsidian). Current dedup uses heuristics (text similarity), which produces false positives at scale and has no way to "link" deduplicated tasks permanently.

**Current gap:** The `MultiSourceAggregator` tracks task origin via `SourceProvider` field and `taskOrigins` map, but there's no canonical identity mapping. A task created locally and synced to two providers gets two different IDs with no link between them.

### 4. Eventual Consistency Guarantees

**Problem:** With multiple providers syncing at different cadences, the local TaskPool may reflect a mix of states — fresh from one provider, stale from another. Users see inconsistent data without knowing which provider is current.

**Current gap:** `SyncStatusTracker` shows per-provider phase (synced/syncing/pending/error) and `LastSyncTime`, but there's no concept of "staleness" — a task synced 30 minutes ago looks the same as one synced 5 seconds ago.

### 5. Partial Failure Handling

**Problem:** When one provider is down, the system must continue operating with the others while clearly communicating degraded state.

**Current strength:** The `MultiSourceAggregator.LoadTasks()` already isolates provider failures — it returns all successful results even if some providers fail, only returning `ErrAllProvidersFailed` if every provider is down. The `FallbackProvider` adds a second layer of graceful degradation.

**Remaining gap:** Write operations don't have the same resilience. `SaveTask()` routes to the originating provider — if that provider is down, the WAL catches it, but there's no automatic fallback to a writable provider.

---

## Patterns from Multi-Source Tools

### Automation Platforms (Zapier, Make.com)

**Hybrid push/poll with overlap windows.** Make.com uses webhooks as primary, with a 2-5 minute polling fallback that runs concurrently. The poller catches any events the webhook missed. Idempotent upserts ensure applying the same event twice is safe.

**Per-integration circuit breaking.** A failing integration enters error state independently — other integrations continue. Backoff prevents hammering a dead service.

**Rate limiting via token buckets.** Each integration has its own rate limiter. Free plans poll every 15 minutes; paid plans as fast as 1 minute.

**Applicability:** ThreeDoors already has `Watch() <-chan ChangeEvent` on the TaskProvider interface for push-style updates (used by Obsidian's fsnotify watcher). The hybrid pattern — Watch() as primary, polling as fallback — maps directly.

### Calendar Apps (Fantastical, BusyCal)

**Per-account independent sync loops.** Each account (Google, iCloud, Exchange) runs its own sync goroutine. Failures in one account don't affect others.

**Protocol-aware scheduling.** CalDAV uses polling (1-5 min), Exchange uses streaming notifications, Google Calendar uses webhook channels (7-day TTL). Each sync loop adapts to its protocol's push capabilities.

**Stable identity + sequence counter = trivial conflict resolution.** CalDAV's RFC 5545 `UID` property and `SEQUENCE` counter mean conflicts are detected by comparing sequence numbers. Higher number wins.

**Applicability:** The per-account sync loop pattern maps to ThreeDoors' per-provider model. The sequence counter pattern could replace timestamp-based conflict detection for providers that support it.

### Task Managers (Todoist, Things, TickTick)

**Opaque sync tokens for incremental sync.** Todoist's Sync API returns a `sync_token` encoding the server's change log position. Subsequent syncs send this token and receive only deltas. No clock skew issues.

**Temp ID mapping for optimistic creation.** Client creates a task with a local UUID, sends it as `temp_id`. Server responds with `temp_id_mapping: {"local-uuid": "server-id"}`. Client replaces the temp ID.

**Adaptive sync frequency.** TickTick expands sync intervals when battery is low or network is metered; contracts when plugged in on WiFi.

**Applicability:** The sync token pattern would improve efficiency for providers with server-side change logs. The temp ID mapping pattern solves the identity mapping problem directly.

### CRDT-Based Tools (Figma, Linear)

**Property-level last-write-wins, not full CRDTs.** Figma tracks `(value, updatedAt, actorID)` per property per object. Two users changing different properties of the same object never conflict. Same property: later timestamp wins; equal timestamps: lexicographic actorID tiebreak for determinism.

**CRDTs only for rich text.** Linear uses CRDTs (Yjs) only for issue descriptions (collaborative text editing). Everything else uses property-level OCC. This hybrid avoids CRDT complexity where it's not needed.

**Applicability:** ThreeDoors tasks have discrete fields (text, status, tags). Property-level LWW is the right model. Full CRDTs are unnecessary — ThreeDoors doesn't support collaborative real-time editing.

### Event-Driven Architectures (Kafka patterns)

**Bulkhead isolation.** Each producer is independent. A failing producer doesn't affect others. This maps to per-provider circuit breakers.

**Dead letter queues.** Messages that fail N times are moved to a separate queue for inspection rather than blocking the pipeline. ThreeDoors' WAL with max retries (10) and entry eviction already implements this.

**Backpressure via token buckets.** Rate limiting at the producer side prevents overwhelming downstream systems.

---

## Proposed Architecture Improvements

### 1. Sync Scheduler with Per-Provider Cadence

Replace the current on-demand sync with a background scheduler that manages independent sync loops per provider.

```
┌─────────────────────────────────────────────┐
│                SyncScheduler                │
│                                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ TextFile │ │  Apple   │ │ Obsidian │   │
│  │  Loop    │ │  Notes   │ │  Loop    │   │
│  │          │ │  Loop    │ │          │   │
│  │ fsnotify │ │ poll 60s │ │ fsnotify │   │
│  │ primary  │ │ adaptive │ │ primary  │   │
│  └──────────┘ └──────────┘ └──────────┘   │
│                                             │
│  Each loop owns:                            │
│  - CircuitBreaker (health tracking)         │
│  - AdaptiveInterval (backoff on failure)    │
│  - RateLimiter (token bucket)               │
│  - SyncState (per-provider snapshots)       │
└─────────────────────────────────────────────┘
```

**Design:**

```go
type SyncScheduler struct {
    loops   map[string]*ProviderLoop
    results chan SyncResult  // fan-in from all loops
    stop    chan struct{}
}

type ProviderLoop struct {
    provider TaskProvider
    circuit  *CircuitBreaker
    interval *AdaptiveInterval
    limiter  *TokenBucket
    state    *SyncState
}

type AdaptiveInterval struct {
    current  time.Duration
    min      time.Duration // e.g., 30s
    max      time.Duration // e.g., 30m
    multiplier float64     // e.g., 2.0
    jitter     float64     // e.g., 0.2 (±20%)
}
```

**Behavior:**
- Each loop runs as an independent goroutine
- Providers with `Watch()` support use the channel as primary trigger; polling runs concurrently as a fallback (overlap window pattern from Make.com)
- On sync success: reset interval to minimum
- On sync failure: multiply interval, up to maximum
- Jitter (±20%) prevents thundering herd when multiple providers recover simultaneously
- Circuit breaker trips after 5 consecutive failures; probes every `max` interval

**Per-provider defaults:**

| Provider | Primary | Poll Interval | Min | Max |
|----------|---------|---------------|-----|-----|
| TextFile | fsnotify | fallback only | 30s | 5m |
| Apple Notes | polling | 60s | 30s | 30m |
| Obsidian | fsnotify | fallback only | 30s | 5m |
| Future HTTP APIs | webhook | fallback poll | 60s | 30m |

### 2. Task Identity Across Providers (Canonical ID Mapping)

Introduce a `SourceRef` that permanently links a task's internal ID to its provider-native ID.

```go
type SourceRef struct {
    Provider string // adapter name: "applenotes", "obsidian", etc.
    NativeID string // ID as the source system knows it
}

// Task gains a stable identity layer
type Task struct {
    ID         string      // internal UUID, stable across syncs
    SourceRefs []SourceRef // one per provider that knows this task
    // ... existing fields ...
}
```

**Identity resolution flow:**

```
1. Provider loads task with native ID "abc123"
2. Lookup: is there an existing Task with SourceRef{provider, "abc123"}?
   → Yes: update that Task (preserving internal ID)
   → No: check dedup heuristics (normalized text match within 24h window)
     → Match found: add SourceRef to existing Task (link across providers)
     → No match: create new Task with new internal UUID + SourceRef
```

**Benefits:**
- Write routing uses `SourceRefs` instead of single `SourceProvider` field — a task synced to multiple providers can be updated in all of them
- Dedup becomes permanent: once two tasks are linked, the link persists via SourceRefs
- Temp ID mapping (Todoist pattern) works naturally: local creation assigns UUID, provider sync adds SourceRef with the provider's native ID

**Migration:** The existing `SourceProvider string` field maps to `SourceRefs[0]`. Backward-compatible: if `SourceRefs` is empty, fall back to `SourceProvider`.

### 3. Property-Level Conflict Resolution

Replace task-level last-write-wins with field-level merging.

```go
type FieldVersion struct {
    Value     any
    UpdatedAt time.Time
    Actor     string // provider name or "local"
}

// Per-field tracking on Task
type TaskFields struct {
    Text   FieldVersion
    Status FieldVersion
    Tags   FieldVersion
    // ... one per mutable field
}
```

**Merge algorithm:**

```
For each field in {text, status, tags, ...}:
  if only local changed  → keep local
  if only remote changed → keep remote
  if both changed:
    if local.value == remote.value → no conflict (convergent edit)
    if remote.UpdatedAt > local.UpdatedAt → remote wins
    if equal timestamps → lexicographic actor tiebreak
    log the override for user notification
```

**Impact:**
- Title changed in Apple Notes + status changed in Obsidian → both changes preserved (no data loss)
- Eliminates the most common "silent data loss" scenario in multi-source sync
- The existing `ConflictResolver` UI can show field-level diffs instead of whole-task diffs

**Complexity tradeoff:** This adds per-field timestamp tracking to the Task model and SyncState snapshots. The storage overhead is modest (a few extra timestamps per task), but the serialization format changes. This is a breaking change to the sync state file format — requires migration.

### 4. Circuit Breaker per Provider

Wrap each provider in a circuit breaker that tracks health state and prevents cascading failures.

```
States:  Closed (healthy) → Open (tripped) → Half-Open (probing) → Closed

Closed:
  - Forward all requests to provider
  - Count consecutive failures
  - After 5 failures within 2 minutes → transition to Open

Open:
  - Return cached last-known state for reads
  - Queue writes to WAL
  - After timeout (starts at 30s, doubles each cycle, max 30m) → transition to Half-Open

Half-Open:
  - Allow one probe request
  - Success → Closed (reset failure count)
  - Failure → Open (double timeout)
```

**Integration with existing components:**
- `SyncStatusTracker` reads circuit state to display per-provider icons:
  - Closed → `✓` synced
  - Open → `✗` error (with last error message)
  - Half-Open → `↻` probing
- `WALProvider` already handles write queueing — circuit breaker adds the "don't even try" fast-fail layer on top
- `MultiSourceAggregator.LoadTasks()` uses circuit state to skip providers in Open state (return cached tasks instead of calling LoadTasks on a known-dead provider)

### 5. Webhook/Push Support vs. Polling

The `Watch() <-chan ChangeEvent` method on `TaskProvider` already supports push-style updates. To scale this:

**For file-backed providers (TextFile, Obsidian):**
- `fsnotify` file watching is the most efficient approach — zero polling, instant detection
- Already partially implemented in ObsidianAdapter
- Should be the default for any provider backed by local files

**For network/IPC providers (Apple Notes, future HTTP APIs):**
- Apple Notes: AppleScript polling is the only option (no push API). Adaptive interval with backoff.
- Future CalDAV providers: support server-push (SUBSCRIBE/NOTIFY) where available, poll as fallback
- Future HTTP API providers: webhook endpoint registration where the service supports it

**Hybrid pattern (from Make.com):**
```
Primary trigger:  Watch() channel (push/fsnotify)
Fallback:         Polling at adaptive interval (runs concurrently)
Dedup:            Idempotent sync — applying the same change twice is a no-op
                  (DetectChanges returns empty changeset if nothing actually changed)
```

The existing `SyncEngine.DetectChanges()` is already idempotent — if remote state matches the last sync snapshot, it returns empty changesets. This means the overlap window pattern works without modification to the sync engine.

### 6. Sync Health Dashboard Enhancements

Extend `SyncStatusTracker` to provide richer observability:

```
Current:   ✓ textfile synced | ✗ applenotes error | ↻ obsidian syncing

Enhanced:  ✓ textfile synced 5s ago
           ✗ applenotes error (circuit open, retry in 2m)
           ↻ obsidian syncing...
           ⏳ WAL pending (3 items, oldest 15m)

           Last 24h: 47 syncs, 2 conflicts resolved, 1 error
```

**New fields on `ProviderSyncStatus`:**

```go
type ProviderSyncStatus struct {
    Name          string
    Phase         SyncPhase
    LastSyncTime  time.Time
    PendingCount  int
    ErrorMsg      string
    // New fields:
    CircuitState  CircuitState
    RetryIn       time.Duration   // time until next probe (when circuit is open)
    StaleSince    time.Duration   // how long since last successful sync
    SyncCount24h  int             // successful syncs in last 24 hours
    ErrorCount24h int             // errors in last 24 hours
}
```

**Staleness indicator:** If `StaleSince` exceeds a threshold (e.g., 5 minutes for file-backed, 15 minutes for network), the TUI could dim or annotate tasks from that provider to signal that the data may be outdated.

---

## Implementation Priority

Ranked by impact-to-effort ratio:

| Priority | Improvement | Impact | Effort | Rationale |
|----------|------------|--------|--------|-----------|
| 1 | Sync Scheduler | High | Medium | Prerequisite for everything else. Without background sync, scaling providers is manual. |
| 2 | Circuit Breaker | High | Low | Small, self-contained component. Immediately improves resilience and status reporting. |
| 3 | Canonical ID Mapping | High | Medium | Solves dedup permanently. Required before adding more providers. |
| 4 | Property-Level Conflicts | Medium | Medium | Reduces data loss. Breaking change to sync state format requires migration. |
| 5 | Dashboard Enhancements | Medium | Low | Mostly UI work on existing SyncStatusTracker. High user-facing value. |
| 6 | Webhook/Push Support | Low (now) | Low | Infrastructure exists via Watch(). Becomes high priority when HTTP-based providers are added. |

---

## Risks and Considerations

**Clock skew.** Property-level LWW depends on accurate timestamps. Across providers, clock sources differ (local filesystem vs. Apple Notes server vs. future cloud APIs). Mitigation: use the local receipt time as the authoritative timestamp, not the remote-reported time. This sacrifices accuracy for consistency — all timestamps come from the same clock.

**Migration complexity.** Adding `SourceRefs` and per-field timestamps changes the Task model and SyncState format. Existing users' sync state files need migration. Mitigation: version the sync state format (`schema_version` field, already present in config); write a migration function that runs on first load of old-format state.

**Memory usage at scale.** With N providers × M tasks, the aggregator holds N×M task references (worst case, if every provider has every task). For ThreeDoors' scale (hundreds of tasks, <10 providers), this is negligible. For larger scales, the aggregator should deduplicate in memory via the canonical ID map.

**Testing complexity.** Per-provider sync loops with circuit breakers and adaptive intervals are inherently concurrent. Testing requires careful use of fake clocks, mock providers, and deterministic scheduling. The existing test patterns (table-driven tests, no testify) scale well to this, but integration tests will need a `TestSyncScheduler` with injectable time sources.

---

## References

- Make.com scheduling patterns: hybrid push/poll with overlap windows
- Todoist Sync API v9: opaque sync tokens, temp_id_mapping
- Figma multiplayer: property-level last-write-wins with actor tiebreak
- Linear sync engine: CRDTs for rich text, OCC for structured fields
- CalDAV RFC 5545: stable UID + SEQUENCE counter for conflict detection
- Circuit breaker pattern (Microsoft Azure Architecture Center)
- Kafka backpressure: token bucket rate limiting, dead letter queues
- Fantastical: per-account independent sync loops with protocol-aware scheduling
