# Supervisor Shift Handover: Design Synthesis

**Date:** 2026-03-11
**Research Method:** Five BMAD party mode sessions with PM, Architect, TEA/QA, Dev, SM, Tech Writer, and Creative Problem Solver personas
**Status:** Research complete — ready for implementation planning

---

## Executive Summary

As multiclaude supervisor context windows fill, performance degrades. This document proposes a **shift handover system** that detects context degradation, serializes operational state, and transfers control to a fresh supervisor instance — all while workers continue uninterrupted.

The design follows three core principles:
1. **External monitoring** — The daemon detects degradation, not the supervisor itself
2. **Cold start with rolling snapshot** — No standby supervisor; the daemon maintains system state continuously
3. **Graceful degradation** — If the outgoing supervisor can't cooperate, handover proceeds anyway

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                  multiclaude daemon                   │
│                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ Refresh Loop  │  │ Shift Clock  │  │  Handover  │ │
│  │  (5 min)      │  │ (monitors    │  │  Manager   │ │
│  │               │  │  transcript) │  │            │ │
│  └──────┬───────┘  └──────┬───────┘  └─────┬──────┘ │
│         │                  │                 │        │
│         ▼                  ▼                 ▼        │
│  ┌──────────────────────────────────────────────┐    │
│  │        Rolling State Snapshot (YAML)          │    │
│  │   workers, PRs, agents, messages — updated    │    │
│  │   every 5 min from external sources           │    │
│  └──────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
         │                    │
         ▼                    ▼
┌─────────────┐      ┌─────────────┐
│  Outgoing    │      │  Incoming    │
│  Supervisor  │─────▶│  Supervisor  │
│  (degraded)  │state │  (fresh)     │
│              │delta │              │
└─────────────┘      └──────┬──────┘
                             │
                    ┌────────┼────────┐
                    ▼        ▼        ▼
              ┌────────┐┌────────┐┌────────┐
              │Worker 1││Worker 2││Worker 3│
              └────────┘└────────┘└────────┘
              (continue uninterrupted)
```

---

## Component Design

### 1. Shift Clock (Session 1)

**Implementation:** Shell script in daemon's 5-minute refresh loop.

**Metrics monitored:**
- JSONL transcript file size
- Context compression event count (grep for compression markers in JSONL)
- Assistant message count

**Threshold model (three-tier):**

| Zone | Trigger | Action |
|---|---|---|
| Green (0-60%) | Below all thresholds | Normal operation |
| Yellow (60-80%) | Compression count ≥ 3, or JSONL > 5MB | Begin rolling snapshot updates, log advisory |
| Red (80%+) | Compression count ≥ 6, or JSONL > 10MB | Trigger handover at next natural seam |

**Clock type:** Hybrid
- **Time floor:** No handover before 30 minutes (prevents thrashing)
- **Usage ceiling:** Force handover at N compression events regardless of time
- **Minimum interval:** 30 minutes between handovers (anti-oscillation)

**Natural seam detection:** Handover waits for: no `multiclaude work` commands in last 60 seconds, no pending message acknowledgments, no active tool calls.

### 2. Rolling State Snapshot (Sessions 3 & 4)

**Maintained by:** Daemon refresh loop (not the supervisor)

**Sources:** External commands only — `multiclaude worker list`, `multiclaude message list`, `gh pr list`, `tmux list-windows`

**Format:** YAML file at `~/.multiclaude/handover/<repo>/shift-state.yaml`

**Schema:** See Session 4 for complete schema. Three sections:
1. **Observable state** — Workers, persistent agents, open PRs (daemon-maintained)
2. **Supervisor delta** — Pending decisions, priorities, blockers, warnings (written by outgoing supervisor)
3. **Operational notes** — Known limitations, warnings

**Size constraint:** Maximum 10KB. Alert if exceeded.

### 3. Handover Protocol (Session 2)

**Normal handover (5 steps):**

```
1. Daemon triggers handover
   └─▶ Sends "HANDOVER_REQUESTED" to outgoing supervisor

2. Outgoing supervisor enters handover mode
   ├─▶ Stops dispatching new workers
   ├─▶ Waits for natural seam
   ├─▶ Writes delta to shift-state.yaml
   ├─▶ Acknowledges all pending messages
   └─▶ Signals "HANDOVER_COMPLETE"

3. Daemon spawns incoming supervisor
   └─▶ Task: "Read shift-state.yaml and assume control"

4. Incoming supervisor startup sequence
   ├─▶ Reads shift-state.yaml
   ├─▶ Reads MEMORY.md
   ├─▶ Reads ROADMAP.md
   ├─▶ Runs multiclaude message list
   ├─▶ Pings each active worker for status
   └─▶ Signals "READY"

5. Daemon kills outgoing supervisor
   └─▶ Archives shift-state.yaml to history/
```

**Emergency handover (outgoing unresponsive):**

```
1. Daemon sends "HANDOVER_REQUESTED"
2. 120-second timeout expires with no response
3. Daemon force-kills outgoing supervisor
4. Daemon spawns incoming with emergency flag
5. Incoming does full worker audit (all pings, all PRs, all messages)
6. Incoming reports discrepancies to user
```

**Authority transfer:** At no point are zero supervisors watching workers. At no point are two supervisors dispatching work. The daemon is the mutex.

### 4. Standby Model (Session 3)

**Decision: Cold start.** No pre-spawned standby supervisor.

| Option | Verdict | Reason |
|---|---|---|
| Hot standby | REJECTED | Doubles API cost, causes split-brain, shadow fills its own context |
| Warm standby | REJECTED | Claude can't hot-swap prompts; "upgrade" requires kill+respawn anyway |
| Cold start | ADOPTED | 60-90s startup is acceptable; simplest; no wasted tokens |

### 5. Failure Recovery (Session 5)

**Critical failure modes and mitigations:**

| Failure | Mitigation |
|---|---|
| Supervisor unresponsive at handover | Emergency protocol: force-kill after 120s, use daemon snapshot |
| Message loss during transition | Non-issue: file-based messaging survives supervisor lifecycle |
| Rapid re-handover (oscillation) | 30-minute minimum interval + compact state file |
| Worker routes to dead supervisor | Role-based addressing: "supervisor" is a role, not an instance |
| Monitor crash | systemd/launchd watchdog restarts daemon |

---

## Implementation Roadmap

### Phase 1: MVP (v1)

**Effort:** ~2-3 stories

1. **Shift clock script** — Add transcript monitoring to daemon refresh loop. Shell script that checks JSONL size, compression count, message count. Emits metrics and writes signal file when thresholds crossed.

2. **Rolling snapshot generator** — Shell script that runs `multiclaude worker list`, `gh pr list`, etc. and writes `shift-state.yaml`. Integrated into daemon refresh loop.

3. **Handover orchestrator** — Daemon logic to: send handover message to outgoing supervisor, wait for delta or timeout, spawn incoming supervisor with state file, kill outgoing after incoming is ready.

4. **Supervisor startup with state file** — Modify supervisor agent definition to check for `shift-state.yaml` on startup and read it as initial briefing.

### Phase 2: Hardening (v2)

5. **Emergency handover protocol** — Timeout-based force-kill and emergency flag for incoming supervisor.

6. **Handover history and audit** — Archive state files, track handover frequency, alert on anomalies.

7. **Continuous journaling** — Supervisor writes structured journal entries during normal operation. Incoming reads journal instead of/in addition to snapshot. (Only if v1 handover quality proves insufficient.)

### Phase 3: Optimization (v3)

8. **Adaptive thresholds** — Shift clock learns from handover history to optimize trigger points.

9. **User-configurable shift policy** — Allow users to set their own thresholds, minimum intervals, and preferences.

---

## Decisions Register

| # | Decision | Adopted | Rejected | Session |
|---|---|---|---|---|
| 1 | Detection mechanism | External daemon monitoring | Supervisor self-reporting | 1 |
| 2 | Metrics | Compression count + JSONL size + message count | Response quality analysis | 1 |
| 3 | Threshold model | Three-tier (green/yellow/red) | Binary (go/no-go) | 1 |
| 4 | Clock type | Hybrid (time floor + usage ceiling) | Pure time or pure usage | 1 |
| 5 | Handover initiation | Daemon decides, supervisor participates | Supervisor self-initiates | 1 |
| 6 | Handover timing | Natural seam (task boundary) | Immediate on threshold | 1 |
| 7 | State transfer | State file + worker pings | Message replay | 2 |
| 8 | Authority transfer | Phased (outgoing → transition → incoming) | Instant cutover | 2 |
| 9 | Standby model | Cold start (60-90s) | Hot standby, warm standby | 3 |
| 10 | Snapshot maintenance | Daemon-maintained rolling snapshot | Supervisor writes everything at handover | 3 |
| 11 | Decision capture | Externalize immediately during operation | Dump at handover time | 4 |
| 12 | State file format | Versioned YAML, max 10KB | JSON, free-form text | 4 |
| 13 | Unresponsive recovery | Force-kill after 120s timeout | Wait indefinitely | 5 |
| 14 | Anti-oscillation | 30-minute minimum interval | No minimum | 5 |
| 15 | Agent addressing | Role-based ("supervisor") | Instance-based ("gentle-hawk") | 5 |

---

## Open Questions

1. **Calibration:** What are the actual JSONL sizes and compression counts at which quality degrades? Needs empirical measurement.
2. **Persistent agent awareness:** Should persistent agents (merge-queue, pr-shepherd) be notified of supervisor handover? They're independent, but awareness might help.
3. **User notification:** Should the user be notified when a handover occurs? Probably yes — a brief notification in the tmux status bar or via message.
4. **Multi-repo:** If multiclaude manages multiple repos, does each get its own shift clock? Probably yes.
5. **Manual trigger:** Should the user be able to force a handover? (e.g., `multiclaude supervisor handover`) Almost certainly yes.

---

## Appendix: Party Mode Sessions

1. [Session 1: Context Window Detection and Shift Clock Design](./session-1-context-detection-shift-clock.md)
2. [Session 2: Worker Handover Protocol](./session-2-worker-handover-protocol.md)
3. [Session 3: Standby Supervisor Feasibility](./session-3-standby-feasibility.md)
4. [Session 4: State Serialization](./session-4-state-serialization.md)
5. [Session 5: Failure Modes and Recovery](./session-5-failure-modes-recovery.md)
