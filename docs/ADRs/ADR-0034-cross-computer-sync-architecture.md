# ADR-0034: Cross-Computer Sync Architecture

- **Status:** Accepted
- **Date:** 2026-03-13
- **Decision Makers:** Architecture research spike (Story 64.1)
- **Related PRs:** (this PR)
- **Related ADRs:** ADR-0012 (Property-Level Conflict Resolution), ADR-0013 (Offline-First WAL), ADR-0011 (Sync Scheduler)

## Context

ThreeDoors users want to use the app across multiple computers (e.g., work laptop and personal desktop) and see the same task state everywhere. The app currently has robust single-machine sync infrastructure — WAL, Connection Manager, three-way sync engine, circuit breaker — designed for syncing with external providers (Jira, Obsidian, etc.). Cross-computer sync is a fundamentally different problem: there is no single authoritative source, both devices are equal peers, and changes can occur simultaneously on disconnected machines.

This ADR documents the architectural decisions for the cross-computer sync feature (Epic 64), covering transport, conflict resolution, device identity, sync scope, and security.

## Key Decisions

### 1. Transport: Git-Based Sync via Shared Bare Repository

**Decision:** Use Git as the sync transport layer. Each device clones a shared bare repository (hosted on any Git server — GitHub, GitLab, self-hosted, or local network), and syncs by committing local changes and pulling remote changes.

**How it works:**
1. User creates or designates a Git remote (any Git hosting) during sync setup
2. ThreeDoors initializes a Git repo inside `~/.threedoors/sync/` (separate from the task data directory)
3. On sync trigger: stage changed files → commit → pull with rebase → push
4. Sync triggers are debounced (30s after last change) + manual via `threedoors sync`
5. The `SyncTransport` interface allows alternative transports in the future

**Why Git:**
- ThreeDoors already stores data as files (YAML, JSONL) — Git is a natural fit
- Built-in versioning, diffing, and conflict detection
- Works offline natively — commits queue locally, push when connected
- Users choose their own hosting — no vendor lock-in, no ThreeDoors-operated server
- Existing tooling ecosystem (SSH keys, GPG signing, access control)
- Append-only JSONL files (session logs) merge cleanly via union strategy

### 2. Conflict Resolution: Property-Level LWW with Device Vector Clocks

**Decision:** Extend the existing property-level last-writer-wins strategy (ADR-0012) with per-device vector clocks for causal ordering across machines.

**How it works:**
1. Each `FieldVersion` in the task's `FieldVersions` map gains a `DeviceID` field
2. A vector clock (`map[DeviceID]uint64`) on each task tracks the logical timestamp per device
3. Non-overlapping changes (different fields modified on different devices) merge automatically — both edits preserved
4. Overlapping changes (same field modified on both devices) use timestamp comparison with device ID as tiebreaker
5. The rejected version is logged to `~/.threedoors/sync/conflicts.jsonl` for recovery
6. Manual override available via `threedoors sync resolve` for unresolved conflicts

**Why this approach:**
- ADR-0012 already established property-level resolution — this extends rather than replaces it
- Vector clocks provide happens-before ordering without a central clock
- LWW with vector clocks is proven in distributed systems (Dynamo, Riak)
- Rejected versions are preserved — no permanent data loss

### 3. Device Identity: UUID Seeded from Machine ID + Install Path

**Decision:** Each ThreeDoors installation generates a deterministic UUID derived from the machine's hardware ID and the install/data path.

**How it works:**
1. On first run, read `/etc/machine-id` (Linux) or `IOPlatformUUID` (macOS) + hash of `~/.threedoors/` path
2. Generate UUID v5 (namespace UUID + machine-id + path) for deterministic stability
3. Persist in `~/.threedoors/device.yaml` with metadata: ID, hostname, first-seen, last-sync
4. If machine-id is unavailable, fall back to UUID v4 (random) — still persisted, so stable across restarts
5. Device discovery happens through the shared Git repo — each device registers itself by writing to `devices/` directory

**Why this approach:**
- Deterministic: survives app reinstalls on the same machine without orphaning the device
- No central server needed for device registration — piggybacks on Git repo
- Human-readable names (hostname default) for user-facing display
- Collision risk is negligible with UUID v5 + machine-specific inputs

### 4. Sync Scope

| Data | Syncs? | Strategy | Rationale |
|------|--------|----------|-----------|
| `tasks.yaml` | **Yes** | Property-level merge with vector clocks | Core data — must be consistent across devices |
| `config.yaml` (theme, values, show_key_hints) | **Yes** | Last-write-wins (whole file) | Rarely changes; no field-level conflicts expected |
| `config.yaml` (provider connections, credentials) | **No** | Excluded | Credentials are machine-specific; provider configs may reference local paths |
| `sessions.jsonl` | **Yes** | Append-only union merge | Sessions are immutable; each device's sessions are unique by session_id |
| `sync-state/` | **No** | Local only | WAL queue and sync status are per-device state |
| `device.yaml` | **No** (but registered in sync repo) | Device writes own entry to `devices/` in sync repo | Identity is local; registration is shared |

### 5. Security Model

**Encryption in transit:** Delegated to Git transport — SSH (recommended) or HTTPS provide encryption. ThreeDoors does not implement its own transport encryption.

**Encryption at rest:** The sync repository can be encrypted using git-crypt or similar tooling. ThreeDoors documents this as a recommended setup step but does not enforce it — the user controls their Git hosting security.

**Authentication:** Git SSH keys or HTTPS credentials. ThreeDoors stores only the Git remote URL — authentication is handled by the user's existing Git credential manager.

**Access control:** The Git remote's access control governs who can sync. For single-user multi-device sync, a private repo suffices. For shared sync (future), Git branch permissions can provide per-user isolation.

### 6. Relationship to Existing Infrastructure

| Component | Current Role | Cross-Computer Sync Role |
|-----------|-------------|--------------------------|
| **WAL (Epic 21)** | Queues writes to unavailable providers | Unchanged — WAL handles provider sync failures. Cross-computer sync has its own queue (Git staging area + unpushed commits) |
| **Connection Manager (Epic 43)** | Manages provider lifecycle and state | New `GitSyncConnection` connection type — reuses state machine, health checking, circuit breaker |
| **Sync Scheduler (Epic 47)** | Polls providers on adaptive intervals | New sync trigger: debounced file-watch (30s) for Git commit+push cycle |
| **Three-Way Sync Engine** | Detects changes between local, remote, and snapshot | Extended to handle device-to-device merges — the sync repo becomes another "remote" |
| **Circuit Breaker** | Protects against cascading provider failures | Reused for Git remote failures — trips on repeated push/pull failures |
| **FieldVersions (ADR-0012)** | Tracks per-field version for provider conflict resolution | Extended with `DeviceID` for cross-machine attribution |

## Evaluated Approaches

### Approach A: Git-Based Sync (CHOSEN)

**Pros:**
- Natural fit for file-based data (YAML, JSONL)
- Built-in versioning, history, and conflict detection
- Users control hosting — no vendor dependency
- Offline-first by design (local commits queue)
- Existing ecosystem (SSH, GPG, branch protection)
- Append-only files merge well

**Cons:**
- Requires user to set up a Git remote (friction)
- Git merge conflicts on YAML require custom merge driver
- Binary data (if added later) would bloat repo
- Git history grows unbounded (needs periodic gc/repack)

**Mitigations:**
- Sync setup wizard (Story 64.2) minimizes friction — guides user through remote creation
- Custom YAML merge driver handles task-level merges (Story 64.4)
- Periodic `git gc` integrated into sync cycle
- `.gitattributes` configures merge strategies per file type

### Approach B: Cloud Intermediary (S3/GCS)

**Pros:**
- Simple read/write model — no merge complexity at transport layer
- Serverless — no infrastructure to maintain
- Object versioning provides history
- Works from any network (no SSH port restrictions)

**Cons:**
- Requires cloud account and billing (recurring cost)
- Vendor lock-in (S3 vs GCS vs Azure Blob)
- No built-in conflict detection — must implement entirely in ThreeDoors
- Upload/download of full snapshots is wasteful (no delta sync without extra work)
- Authentication complexity (IAM roles, access keys, credential rotation)

**Rejected because:** The cloud dependency violates ThreeDoors' local-first philosophy. Users shouldn't need a cloud account to sync between two machines on the same network. The lack of built-in conflict detection means reimplementing what Git provides for free.

### Approach C: Peer-to-Peer (mDNS + Direct TCP)

**Pros:**
- No server needed — zero external dependencies
- Works on LAN without internet
- Low latency for same-network devices
- Maximum privacy — data never leaves local network

**Cons:**
- NAT traversal is extremely complex for cross-network sync (STUN/TURN, hole punching)
- Device discovery fails across networks (mDNS is LAN-only)
- No offline queue — devices must be online simultaneously
- Requires implementing transport security from scratch (TLS, certificate management)
- Firewall issues in corporate environments
- No sync history — if both devices are offline, changes are lost until both are online

**Rejected because:** P2P only works reliably on the same network. Cross-network sync (the primary use case — work laptop ↔ home desktop) requires NAT traversal infrastructure that is effectively building a server. The simultaneous-online requirement contradicts offline-first design. Implementation complexity is 3-5x higher than git-based for a worse user experience.

### Approach D: Hybrid Git + LAN Discovery (Evaluated but deferred)

**Pros:**
- Git for cross-network (reliable, async), mDNS+TCP for same-network (fast, real-time)
- Best of both worlds — low latency on LAN, reliable across networks

**Cons:**
- Double the implementation and testing surface
- Conflict resolution must handle two different change sources
- LAN sync may conflict with queued Git sync — ordering becomes complex
- Premature optimization — the 30s debounced Git sync is fast enough for most users

**Rejected (for now) because:** YAGNI. Git-based sync with 30s debounce covers 95% of use cases. LAN discovery can be added as a Story 64.7+ enhancement if users report latency issues. Adding it now doubles implementation scope for marginal benefit.

## Open Questions

1. **Git merge driver:** Should ThreeDoors register a custom Git merge driver for `tasks.yaml`, or handle conflicts entirely in application code post-merge?
2. **Sync frequency:** Is 30s debounce the right default? Should it be configurable?
3. **Repo size management:** When should `git gc` run? After N commits? On a timer?
4. **Multi-user sync:** This ADR covers single-user multi-device. Multi-user task sharing is a separate feature that could build on this foundation (per-user branches, merge-on-sync).
5. **Provider credential sync:** Should provider connection configs (minus credentials) sync so that both devices auto-discover the same Jira/Obsidian sources?

## Consequences

### Positive
- Leverages existing sync infrastructure (Connection Manager, circuit breaker, scheduler)
- Extends existing conflict resolution (ADR-0012) rather than replacing it
- No cloud dependency — users control their own data and hosting
- Offline-first design preserved — Git commits queue locally
- Future-proof — `SyncTransport` interface allows swapping transport without touching sync logic

### Negative
- Git setup is a one-time friction point (mitigated by setup wizard)
- Custom YAML merge driver adds maintenance burden
- Git repo grows over time (mitigated by periodic gc)
- Users must understand basic Git concepts (remote, SSH keys) for initial setup
