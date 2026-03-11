# Party Mode Session 5: Failure Modes and Recovery

**Date:** 2026-03-11
**Topic:** Supervisor Shift Handover — What If Handover Fails Mid-Transfer?
**Participants:** Murat (TEA), Winston (Architect), John (PM), Amelia (Dev), Dr. Quinn (Creative Problem Solver)

---

## Problem Statement

Handover is a critical operation. If it fails, the system could lose state, leave workers unmonitored, or enter a degraded state. We need a complete failure taxonomy and recovery strategies.

## Failure Taxonomy

### Category A: Detection Failures

| ID | Failure | Description | Severity |
|---|---|---|---|
| A1 | False negative | Supervisor is degraded but shift clock doesn't trigger | HIGH |
| A2 | False positive | Shift clock triggers too early, losing context unnecessarily | LOW |
| A3 | Monitor crash | Daemon's monitoring loop itself fails | HIGH |

### Category B: Handover Failures

| ID | Failure | Description | Severity |
|---|---|---|---|
| B1 | Outgoing unresponsive | Supervisor can't write its delta (frozen, stuck, exhausted) | CRITICAL |
| B2 | State file corruption | Handover YAML is malformed or incomplete | MEDIUM |
| B3 | Incoming fails to start | Claude API error, tmux issue, spawn failure | HIGH |
| B4 | Message loss in transition | Worker message arrives between outgoing death and incoming readiness | LOW (mitigated by design) |

### Category C: Post-Handover Failures

| ID | Failure | Description | Severity |
|---|---|---|---|
| C1 | State misinterpretation | Incoming reads handover file but draws wrong conclusions | MEDIUM |
| C2 | Worker routing failure | Worker sends messages to dead supervisor instance | MEDIUM |
| C3 | Persistent agent confusion | merge-queue/pr-shepherd report to wrong supervisor | MEDIUM |
| C4 | Rapid re-handover | Incoming fills context quickly, triggers another handover within minutes | HIGH |

## Recovery Strategies

### B1: Emergency Handover Protocol (CRITICAL)

The nightmare scenario: supervisor so degraded it can't even cooperate with handover.

**Protocol:**
1. Daemon detects supervisor hasn't responded to handover request in 120 seconds
2. Daemon forcibly kills the outgoing supervisor process
3. Daemon spawns incoming supervisor with emergency flag: "Previous supervisor was unresponsive. Using last daemon snapshot. Verify all worker states manually."
4. Incoming supervisor does full worker audit: ping every worker, check every PR, reconcile message queue
5. Incoming supervisor reports any discrepancies to user

**Key principle:** Never let a broken supervisor block the system. If it can't hand over gracefully, hand over ungracefully. Some context loss is better than a permanently stuck system.

### B4: Message Loss — Non-Issue by Design

The multiclaude message system is file-based. Messages persist on disk regardless of supervisor state. The incoming supervisor runs `multiclaude message list` as part of startup. **No messages can be lost** as long as the filesystem is intact.

### C2/C3: Role-Based Addressing

Workers and persistent agents should address messages to the "supervisor" **role**, not a specific instance name. The daemon maintains the role → instance mapping. When workers call `multiclaude message send supervisor "status update"`, the daemon routes to whoever is currently authoritative.

This is already how multiclaude works. No worker or agent code changes needed — just ensure the daemon updates the role mapping at handover time.

### C4: Anti-Oscillation Measures

- **Minimum handover interval: 30 minutes** — Daemon refuses to trigger handover if the last one was less than 30 minutes ago
- **State file size limit: 10KB** — Alert if handover state exceeds this; investigate why
- **Compact state design** — The proposed schema is ~2-3KB of YAML, negligible in context terms

## Recovery Test Matrix

| Failure | Detection Method | Recovery Action | Verification |
|---|---|---|---|
| A1: False negative | Manual quality check by user | User triggers manual handover | Compare pre/post decision quality |
| A2: False positive | Handover count exceeds threshold/hour | Increase shift clock thresholds | Monitor handover frequency |
| A3: Monitor crash | systemd/launchd watchdog on daemon | Auto-restart daemon | Daemon health endpoint |
| B1: Unresponsive supervisor | 120s timeout on handover request | Force-kill, use daemon snapshot | Worker audit by incoming |
| B2: State file corruption | YAML parse error on read | Fall back to daemon-only snapshot | Schema validation before write |
| B3: Incoming fails to start | Spawn command returns error | Retry once, alert user on 2nd failure | Check tmux session exists |
| B4: Message loss | N/A — file-based, can't happen | N/A | Assert zero unacked messages post-handover |
| C1: Misinterpretation | Hard to detect automatically | Worker pings catch factual errors | Cross-reference state with worker reports |
| C2: Worker routes dead | Message to non-existent agent errors | Role-based addressing | Test message routing post-handover |
| C3: Agent confusion | Same as C2 | Same — role-based addressing | Verify agents can reach new supervisor |
| C4: Rapid re-handover | Second handover within 30 min | Minimum interval, compact state | Monitor time-between-handovers |

## Future Evolution: Continuous Journaling (v2)

**Proposed by Dr. Quinn (Creative Problem Solver):**

Instead of a point-in-time snapshot, the supervisor continuously writes a machine-readable journal of decisions and observations during normal operation. At handover, the incoming supervisor reads only the entries since the last handover.

**Benefits over snapshot approach:**
- Eliminates B1 (unresponsive outgoing) — journal was written during normal operation
- Reduces C1 (misinterpretation) — journal shows reasoning chain, not just conclusions
- Makes C4 (rapid re-handover) less painful — each handover consumes only recent entries

**Why deferred to v2:**
- More complex to implement than snapshot + delta
- More context-heavy for incoming supervisor to parse
- Snapshot approach is sufficient for MVP
- Can be added as an upgrade if handover quality proves insufficient

## Architectural Principles

1. **Never let a broken supervisor block the system** — ungraceful handover > no handover
2. **Role-based addressing** — decouple agent identity from supervisor instance
3. **File-based messaging** — messages survive supervisor lifecycle transitions by design
4. **Minimum handover interval** — prevent oscillation/thrashing
5. **Daemon as authority** — the daemon decides handover timing, not the supervisor
6. **Graceful degradation** — if delta is unavailable, daemon snapshot alone is 90% sufficient
