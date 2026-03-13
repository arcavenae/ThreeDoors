# Cross-Computer Sync — Architecture Research Spike (Story 64.1)

**Date:** 2026-03-13
**Type:** Architecture Research / Party Mode Deliberation
**ADR:** ADR-0034

## Deliberation Summary

This spike evaluated four approaches for cross-computer task synchronization in ThreeDoors. The evaluation considered transport mechanisms, conflict resolution strategies, device identity, sync scope, and security — grounded in ThreeDoors' existing sync infrastructure (WAL, Connection Manager, three-way sync engine, circuit breaker).

## Adopted Approach: Git-Based Sync with Property-Level LWW + Vector Clocks

### Transport: Git Shared Bare Repository
ThreeDoors will sync via a Git repository that users provide (GitHub, GitLab, self-hosted, or even a bare repo on a shared drive). The sync cycle is: stage → commit → pull --rebase → push, triggered by a 30s debounce after changes or manually via `threedoors sync`.

**Rationale:** ThreeDoors stores data as files (YAML, JSONL). Git is the native tool for syncing files with built-in versioning, conflict detection, offline queuing (unpushed commits), and an ecosystem of hosting options. It avoids cloud vendor lock-in and respects the local-first philosophy.

### Conflict Resolution: Extended ADR-0012 with Device Vector Clocks
The existing property-level LWW from ADR-0012 is extended with per-device vector clocks. Each field version tracks which device made the change. Non-overlapping edits merge automatically; overlapping edits use timestamp + device-ID tiebreaker. Rejected versions are logged for recovery.

**Rationale:** Builds on proven infrastructure rather than introducing CRDTs or a new resolution model. Vector clocks provide causal ordering without a central server.

### Device Identity: UUID v5 from Machine ID + Install Path
Deterministic, stable across reinstalls, no central registry needed. Devices discover each other through the shared Git repo.

### Sync Scope
- **Sync:** tasks.yaml (property-level merge), config.yaml preferences (LWW), sessions.jsonl (append-union)
- **Don't sync:** Provider credentials, sync-state/ (per-device), WAL queue

### Security
Delegated to Git transport (SSH/HTTPS). Optional git-crypt for at-rest encryption.

## Rejected Approaches

### Approach B: Cloud Intermediary (S3/GCS) — REJECTED
- **What:** Store task snapshots as versioned objects in cloud storage
- **Why rejected:** Adds cloud dependency and recurring cost, violating local-first philosophy. No built-in conflict detection — would require reimplementing what Git provides. Authentication complexity (IAM, access keys) is worse than SSH keys. Delta sync requires extra work; full snapshots are wasteful.
- **When to reconsider:** If ThreeDoors adds a hosted/SaaS tier where the operator controls infrastructure.

### Approach C: Peer-to-Peer (mDNS + Direct TCP) — REJECTED
- **What:** Discover devices via mDNS on LAN, sync directly via TCP
- **Why rejected:** Only works reliably on the same network. The primary use case (work laptop ↔ home desktop) requires NAT traversal, which is effectively building a relay server. Devices must be online simultaneously — no offline queue. Implementation complexity is 3-5x higher than Git for worse reliability. Firewall issues in corporate environments.
- **When to reconsider:** Never as primary transport. Could be added as optional LAN accelerator if Git latency becomes a user complaint.

### Approach D: Hybrid Git + LAN Discovery — REJECTED (DEFERRED)
- **What:** Use Git for cross-network sync, add mDNS+TCP for fast same-network sync
- **Why rejected:** Doubles implementation surface for marginal benefit. 30s debounced Git sync is fast enough. Two change sources create ordering complexity. YAGNI.
- **When to reconsider:** Post-64.6, if user feedback indicates latency is a problem for same-network devices.

### Conflict Resolution Alternative: Full CRDTs — REJECTED
- **What:** Implement operation-based CRDTs for all task fields
- **Why rejected:** Massive implementation complexity for a personal task manager. CRDTs are designed for systems with thousands of concurrent writers — ThreeDoors has 2-5 devices for one user. The existing property-level LWW with vector clocks provides equivalent correctness for this scale with 10% of the complexity. CRDT debugging is notoriously difficult.
- **When to reconsider:** If ThreeDoors adds real-time collaborative editing (multiple users editing the same task simultaneously).

### Conflict Resolution Alternative: Manual-Only Resolution — REJECTED
- **What:** Surface all conflicts to the user for manual resolution
- **Why rejected:** Terrible UX. Most cross-device "conflicts" are non-overlapping (edited title on laptop, added note on desktop) and should merge automatically. Manual resolution creates friction that discourages syncing. The property-level approach auto-resolves 90%+ of conflicts.
- **When to reconsider:** Never as the default. Manual override is available for edge cases via `threedoors sync resolve`.

## Key Design Decisions

| Decision | Choice | Key Tradeoff |
|----------|--------|-------------|
| Transport | Git bare repo | Setup friction vs. no cloud dependency |
| Conflict resolution | Property-level LWW + vector clocks | Complexity vs. data preservation |
| Device identity | UUID v5 (machine-id + path) | Determinism vs. portability |
| Session sync | Append-only union | Storage growth vs. complete history |
| Config sync | LWW (exclude credentials) | Convenience vs. security |
| Security | Delegate to Git (SSH/HTTPS) | Simplicity vs. app-level encryption |

## Infrastructure Reuse Map

| Existing Component | Reuse in Cross-Computer Sync |
|-------------------|------------------------------|
| WAL (ADR-0013) | Unchanged — handles provider sync. Git has its own queue (unpushed commits) |
| Connection Manager | New `GitSyncConnection` type reuses state machine, health, circuit breaker |
| Sync Scheduler | New trigger: debounced file-watch for Git commit cycle |
| Three-Way Sync Engine | Extended for device-to-device merges |
| Circuit Breaker | Reused for Git remote failures |
| FieldVersions (ADR-0012) | Extended with DeviceID field |

## Open Questions for Implementation

1. Custom Git merge driver vs. application-level merge for `tasks.yaml`
2. Optimal debounce interval (30s default — configurable?)
3. Git repo size management cadence (`git gc` trigger)
4. Whether provider connection configs (sans credentials) should sync
5. Multi-user sync as future extension (per-user branches)

## Story Refinements

Based on this spike, stories 64.2-64.6 have been refined with concrete acceptance criteria replacing provisional placeholders. Key changes:
- **64.2:** Device identity uses UUID v5, registration via Git repo `devices/` directory
- **64.3:** Transport is Git-based, implements `SyncTransport` interface for pluggability
- **64.4:** Conflict resolution extends ADR-0012 with vector clocks, adds Git merge driver
- **64.5:** Offline queue leverages Git's unpushed commit queue, not a separate mechanism
- **64.6:** E2E tests use loopback Git repos with `t.TempDir()` for device simulation
