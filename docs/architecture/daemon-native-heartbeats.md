# Daemon-Native Heartbeats Design

**Story:** 73.6 — Daemon-Native Heartbeats
**Date:** 2026-03-29
**Status:** Design Complete
**Decisions:** Q-C-011 (drop CronCreate), Q-C-015 (configurable per-agent intervals)
**Depends on:** Story 73.2 (CronCreate heartbeats removed)

---

## Problem Statement

After Story 73.2 removed CronCreate heartbeats, the system relies solely on the daemon's 2-minute wake loop (`wakeAgents()`, `daemon.go:437`). This loop keeps agents alive but has three gaps:

1. **One-size-fits-all cadence.** merge-queue needs 5-minute polling; retrospector needs 15-minute cycles. The 2-minute interval wastes context on agents that don't need frequent nudges and may be too slow for time-sensitive agents during high-throughput periods.

2. **No workflow triggers.** The wake loop sends generic "check your work" prompts. It cannot send workflow-specific messages like `SYNC_OPERATIONAL_DATA` — the only non-heartbeat CronCreate job that existed.

3. **No activity awareness.** The `LastNudge` deduplication prevents back-to-back nudges within the same cycle, but doesn't detect whether an agent is actively working. An agent mid-task gets interrupted by a redundant nudge that consumes context.

This design extends the daemon wake loop into a configurable, activity-aware heartbeat system that replaces all CronCreate scheduling needs.

---

## Design Overview

### Architecture

```
┌─────────────────────────────────────────────────┐
│                multiclaude daemon                │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │         Heartbeat Scheduler              │    │
│  │                                          │    │
│  │  Per-agent timers:                       │    │
│  │    merge-queue:      every 5m            │    │
│  │    pr-shepherd:      every 5m            │    │
│  │    envoy:            every 10m           │    │
│  │    project-watchdog:  every 15m          │    │
│  │    arch-watchdog:     every 20m          │    │
│  │    retrospector:      every 15m          │    │
│  │                                          │    │
│  │  Workflow triggers:                      │    │
│  │    SYNC_OPERATIONAL_DATA → pw every 3h   │    │
│  │                                          │    │
│  │  Activity check before each delivery:    │    │
│  │    → Skip if agent active (recent I/O)   │    │
│  │    → Deliver if agent idle               │    │
│  └──────────────────────────────────────────┘    │
│                    │                             │
│                    ▼                             │
│  ┌──────────────────────────────────────────┐    │
│  │         Delivery Layer                   │    │
│  │                                          │    │
│  │  Option A: tmux paste-buffer (current)   │    │
│  │  Option B: message queue file            │    │
│  │  Recommendation: Keep tmux paste-buffer  │    │
│  │  for heartbeats; use message queue for   │    │
│  │  workflow triggers (SYNC, QUOTA_CHECK)   │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scheduling location | Daemon-native (not CronCreate, not agent-internal) | Survives restarts, no session scoping, single source of truth |
| Interval model | Per-agent configurable with sensible defaults | Different agents have different workload patterns |
| Delivery mechanism | Tmux paste for heartbeats, message queue for workflow triggers | Heartbeats are lightweight nudges; workflow triggers need reliable delivery and acknowledgment |
| Activity detection | Tmux pane activity timestamp + message ack recency | Available without agent cooperation; no agent code changes needed |
| Staggering strategy | Offset from daemon start time, not wall-clock cron expressions | Simpler implementation; avoids thundering-herd at :00/:30 |

---

## AC1: Per-Agent Interval Configuration

### Recommended Intervals by Agent Type

Intervals are based on each agent's workload characteristics observed over 300+ merged PRs:

| Agent | Interval | Rationale |
|-------|----------|-----------|
| merge-queue | 5 min | Time-sensitive: PR merges block other PRs. Must poll frequently. |
| pr-shepherd | 5 min | Time-sensitive: Branch updates and rebases unblock merge-queue. |
| envoy | 10 min | Medium: Issue triage is important but not latency-critical. |
| project-watchdog | 15 min | Batch-oriented: Doc sync and watchdog checks are naturally batched. |
| arch-watchdog | 20 min | Low-frequency: Architecture drift detection is a slow-changing signal. |
| retrospector | 15 min | Batch-oriented: PR analysis works on accumulated data, not real-time. |

### Why Not Keep 2-Minute Universal?

The current 2-minute interval causes:
- **Context waste:** retrospector receives 30 nudges/hour but only needs 4. Each nudge adds ~200 tokens of prompt + response overhead. Over a 6-hour session, that's 3,600 wasted context tokens — significant for an agent that already processes large PR diffs.
- **Noise in transcripts:** Operators reviewing agent logs see mostly heartbeat acknowledgments rather than substantive work.
- **No differentiation:** The daemon can't distinguish "check for PRs to merge" from "run a 15-minute analysis cycle."

### Why Not Prime-Number Intervals?

The CronCreate system used prime intervals (7, 11, 13, 23 min) to prevent heartbeat storms — multiple crons firing simultaneously. This is unnecessary with daemon-native scheduling because:
- The daemon controls all timers and can stagger them by design
- Per-agent timers start from daemon boot time with automatic offsets
- There is no wall-clock alignment that would cause simultaneous firing

---

## AC2: Agent Activity Detection

### Detection Signals

The daemon should check two signals before delivering a heartbeat:

**Signal 1: Tmux pane activity timestamp**

```go
// tmux tracks when each pane last received output
// Available via: tmux display -p -t <target> '#{pane_last_activity}'
lastActivity := getTmuxPaneActivity(agentPane)
idleDuration := time.Now().UTC().Sub(lastActivity)
```

If the pane has produced output within a configurable threshold (default: 60 seconds), the agent is actively working. Skip the heartbeat.

**Signal 2: Last message acknowledgment**

```go
// The daemon already tracks message delivery times
// Check if the agent has ack'd its most recent message
lastAck := getLastMessageAck(agentName)
pendingMessages := countPendingMessages(agentName)
```

If the agent has unacknowledged messages older than 2x the heartbeat interval, the agent may be stuck — deliver the heartbeat anyway (it may unstick a confused agent).

### Decision Matrix

| Pane Active? | Messages Pending? | Action |
|-------------|-------------------|--------|
| Yes (< 60s) | No | **Skip** — agent is working |
| Yes (< 60s) | Yes (< 2x interval old) | **Skip** — agent is working, will get to messages |
| Yes (< 60s) | Yes (> 2x interval old) | **Deliver** — agent may be stuck in a loop |
| No (> 60s) | No | **Deliver** — agent is idle, needs nudge |
| No (> 60s) | Yes | **Deliver** — agent is idle with pending work |

### Why Not More Sophisticated Detection?

Options considered and rejected:

- **Parse tmux pane content for tool calls:** Fragile, requires understanding Claude's output format, breaks on model updates.
- **Agent self-reports activity via file/message:** Requires agent code changes, adds complexity, agents can't reliably track their own idle state.
- **CPU/process monitoring:** Claude runs inside tmux via the `claude` CLI — process-level metrics don't distinguish "thinking" from "idle."

The tmux activity timestamp is simple, reliable, requires no agent cooperation, and is already available via tmux's built-in tracking.

---

## AC3: Configuration Schema

### Location

Per-repo configuration in `~/.multiclaude/repos/<repo-name>/config.yaml` (extends existing `multiclaude config` structure).

### Schema

```yaml
# ~/.multiclaude/repos/<repo-name>/config.yaml
heartbeats:
  # Enable/disable the heartbeat system (default: true)
  enabled: true

  # Default interval for agents not listed below (default: 5m)
  default_interval: 5m

  # Activity detection: skip heartbeat if agent pane was active
  # within this window (default: 60s)
  activity_threshold: 60s

  # Per-agent interval overrides
  agents:
    merge-queue: 5m
    pr-shepherd: 5m
    envoy: 10m
    project-watchdog: 15m
    arch-watchdog: 20m
    retrospector: 15m

  # Workflow triggers (non-heartbeat periodic messages)
  triggers:
    - name: SYNC_OPERATIONAL_DATA
      target: project-watchdog
      interval: 3h
      message: "SYNC_OPERATIONAL_DATA"
      # Delivery via message queue (not tmux paste) for reliability
      delivery: message

    # Future: Story 76.6 QUOTA_CHECK
    # - name: QUOTA_CHECK
    #   target: supervisor
    #   interval: 5m
    #   message: "QUOTA_CHECK"
    #   delivery: message
```

### Validation Rules

- `default_interval` must be >= 1m and <= 60m
- Per-agent intervals must be >= 1m and <= 120m
- `activity_threshold` must be >= 10s and <= 300s (5 min)
- Trigger intervals must be >= 5m (no sub-5-minute workflow triggers)
- Agent names must match known agents in the repo's state file
- Unknown agent names in the `agents` map should warn, not error (agents may not be spawned yet)

### CLI Integration

```bash
# View current heartbeat config
multiclaude config heartbeats

# Set default interval
multiclaude config heartbeats --default-interval 5m

# Set per-agent interval
multiclaude config heartbeats --agent merge-queue --interval 3m

# Add a workflow trigger
multiclaude config heartbeats --add-trigger SYNC_OPERATIONAL_DATA \
  --target project-watchdog --interval 3h

# Disable heartbeats entirely
multiclaude config heartbeats --enabled false
```

---

## AC4: Migration Path from CronCreate

### Phase Timeline

```
Phase 1 (Story 73.2) ─── COMPLETE ──────────────────────────────────────
  ✓ Removed 6 CronCreate heartbeat jobs from startup checklist
  ✓ Daemon 2-min wake loop is the sole heartbeat mechanism
  ✓ SYNC_OPERATIONAL_DATA CronCreate retained (but fragile)
  ✓ Agent definitions retain Polling Loop and HEARTBEAT sections

Phase 2 (Story 73.6) ─── THIS DESIGN ───────────────────────────────────
  → Design daemon-native per-agent intervals (this document)
  → Design activity detection
  → Design SYNC_OPERATIONAL_DATA as daemon trigger
  → Write multiclaude enhancement spec

Phase 3 (Implementation) ─── FUTURE multiclaude PR ─────────────────────
  → Implement heartbeat scheduler in daemon
  → Implement activity detection
  → Implement config schema and CLI
  → Implement workflow triggers
  → Remove SYNC_OPERATIONAL_DATA CronCreate from startup checklist
  → Update persistent-agent-ops.md

Phase 4 (Story 73.4 Integration) ─── FUTURE ────────────────────────────
  → Session handoff state includes heartbeat/response metrics
  → Daemon persists heartbeat history for handoff context
  → Agent state directory tracks heartbeat response patterns
```

### Gap Analysis: Phase 1 → Phase 3

During the gap between CronCreate removal and daemon-native implementation:

| Capability | Status | Impact | Mitigation |
|-----------|--------|--------|------------|
| Agent wake nudges | **Working** — daemon 2-min loop | None | N/A |
| Per-agent intervals | **Missing** — all agents get 2-min | Low — slightly wasteful for slow agents | Accept until Phase 3 |
| Activity detection | **Partial** — `LastNudge` dedup only | Low — some redundant nudges | Accept until Phase 3 |
| SYNC_OPERATIONAL_DATA | **BROKEN** — CronCreate removed, no replacement | **High** — operational data not syncing to git | **Interim fix required** (see below) |

### Interim Fix: SYNC_OPERATIONAL_DATA (Phase 1→3 Bridge)

The SYNC_OPERATIONAL_DATA pipeline is dead right now. Two interim options until Phase 3:

**Option A (Recommended): Add to project-watchdog's wake response**

Add a standing order to `agents/project-watchdog.md`:
> On every wake nudge, check if 3+ hours have elapsed since the last SYNC_OPERATIONAL_DATA run. If so, execute the sync workflow. Track the last run timestamp in the checkpoint file.

This is imperfect — Claude can't reliably track elapsed time. But the daemon wake prompt could include a UTC timestamp, giving project-watchdog a clock reference:

```
// In wakeAgents(), append timestamp to the nudge:
"Status check: [2026-03-29T14:30:00Z] Review operational data freshness."
```

**Option B: Supervisor re-creates CronCreate on startup**

Keep the `CronCreate("0 */3 * * *", ...)` in the startup checklist. This works but reintroduces the session-scoping fragility that 73.2 aimed to eliminate.

**Recommendation:** Option A for the interim. It's imperfect but doesn't regress on CronCreate removal. Phase 3 replaces it with proper daemon-native triggers.

---

## AC5: SYNC_OPERATIONAL_DATA Replacement

### Current Pipeline

```
CronCreate (every 3h)
  → supervisor runs: multiclaude message send project-watchdog SYNC_OPERATIONAL_DATA
    → project-watchdog checks docs/operations/ for changes
      → if changes: create data-sync branch, commit, push, create PR
        → merge-queue merges the PR
```

### Daemon-Native Replacement

```
Daemon heartbeat scheduler (trigger: SYNC_OPERATIONAL_DATA, every 3h)
  → daemon writes message to project-watchdog's queue
    → daemon delivers message via tmux paste-buffer (next route cycle)
      → project-watchdog runs sync workflow (unchanged)
        → merge-queue merges the PR (unchanged)
```

### Key Differences

| Aspect | CronCreate | Daemon-Native |
|--------|-----------|---------------|
| Trigger source | Supervisor REPL (session-scoped) | Daemon timer (persistent) |
| Survives restart | No — must re-create on every supervisor restart | Yes — config file is persistent |
| Supervisor involvement | Yes — CronCreate fires into supervisor, supervisor runs command | None — daemon delivers directly |
| Delivery | Indirect: supervisor → message file → daemon → agent | Direct: daemon → message file → daemon → agent |
| Configuration | Hardcoded in startup checklist | YAML config, editable via CLI |

### Workflow Triggers as a General Pattern

SYNC_OPERATIONAL_DATA is the first workflow trigger, but the pattern generalizes:

```yaml
triggers:
  - name: SYNC_OPERATIONAL_DATA
    target: project-watchdog
    interval: 3h
    message: "SYNC_OPERATIONAL_DATA"
    delivery: message

  # Future: quota monitoring (Story 76.6)
  - name: QUOTA_CHECK
    target: supervisor
    interval: 5m
    message: "QUOTA_CHECK"
    delivery: message

  # Future: BMAD PM sprint audit
  - name: SPRINT_AUDIT
    target: supervisor
    interval: 30m
    message: "/bmad-bmm-sprint-status"
    delivery: message
```

Each trigger is:
- **Named** for logging and debugging
- **Targeted** to a specific agent
- **Interval-based** with daemon-managed timing
- **Message-delivered** for reliable queuing (not tmux paste for workflow triggers)

### Activity Detection for Triggers

Workflow triggers should NOT be skipped based on activity detection. Unlike heartbeats (which are redundant if the agent is already working), workflow triggers initiate specific work that the agent wouldn't do otherwise. The SYNC_OPERATIONAL_DATA trigger should fire regardless of whether project-watchdog is currently active.

Configuration supports this:

```yaml
triggers:
  - name: SYNC_OPERATIONAL_DATA
    target: project-watchdog
    interval: 3h
    message: "SYNC_OPERATIONAL_DATA"
    delivery: message
    skip_if_active: false  # Always deliver workflow triggers
```

---

## Delivery Mechanism Details

### Heartbeats: Tmux Paste-Buffer (Keep Current)

The current `tmux paste-buffer` delivery works well for heartbeats:
- Low overhead: no file I/O, no message routing
- Immediate: appears in agent's input stream instantly
- Lightweight: heartbeats are short strings
- No acknowledgment needed: heartbeats are best-effort nudges

### Workflow Triggers: Message Queue (New)

Workflow triggers should use the message queue (`~/.multiclaude/messages/<repo>/<agent>/msg-*.json`):
- **Reliable:** Message files persist until acknowledged
- **Traceable:** Each trigger creates a message with ID, timestamp, and body
- **Retryable:** If the agent doesn't ack within a timeout, the daemon can redeliver
- **Auditable:** Message history shows when triggers fired and when agents responded

### Hybrid Delivery Logic

```go
func (d *Daemon) deliverHeartbeat(agent *AgentState, heartbeatType string) {
    switch heartbeatType {
    case "heartbeat":
        // Direct tmux paste — lightweight, best-effort
        d.tmux.SendKeysLiteralWithEnter(agent.Pane, agent.WakePrompt)
    case "trigger":
        // Message queue — reliable, tracked
        d.messages.Send(Message{
            From:    "daemon",
            To:      agent.Name,
            Body:    agent.TriggerMessage,
            Type:    "workflow_trigger",
        })
    }
}
```

---

## Staggering Strategy

### Problem

If all agent heartbeats fire at daemon startup, they create a burst of activity. With 6 agents at various intervals, alignment points occur where multiple heartbeats coincide.

### Solution: Boot-Time Offset

Each agent's first heartbeat is offset from daemon start by a fraction of its interval:

```go
func calculateOffset(agentIndex int, totalAgents int, interval time.Duration) time.Duration {
    // Spread agents evenly across the first interval period
    fraction := float64(agentIndex) / float64(totalAgents)
    return time.Duration(fraction * float64(interval))
}

// Example with 6 agents, daemon starts at T=0:
// merge-queue (5m):      first at T+0:00, then every 5m
// pr-shepherd (5m):      first at T+0:50, then every 5m
// envoy (10m):           first at T+1:40, then every 10m
// project-watchdog (15m): first at T+2:30, then every 15m
// retrospector (15m):    first at T+3:20, then every 15m
// arch-watchdog (20m):   first at T+4:10, then every 20m
```

This is simpler than prime-number cron expressions and achieves the same goal: no thundering herd.

---

## Observability

### Daemon Log Output

Each heartbeat decision should be logged:

```
[heartbeat] merge-queue: delivered (idle 312s, interval 300s)
[heartbeat] pr-shepherd: skipped (active 23s ago, interval 300s)
[heartbeat] retrospector: delivered (idle 1847s, interval 900s)
[trigger] SYNC_OPERATIONAL_DATA → project-watchdog: delivered (interval 10800s)
```

### Metrics (Future)

If/when the daemon gains metrics export (Story 76.x), heartbeat data is valuable:

- `heartbeat_delivered_total{agent}` — counter
- `heartbeat_skipped_total{agent,reason}` — counter (reasons: active, recently_nudged)
- `heartbeat_response_time_seconds{agent}` — histogram (time from delivery to next pane activity)
- `trigger_delivered_total{trigger_name}` — counter
- `trigger_ack_time_seconds{trigger_name}` — histogram

---

## Rejected Alternatives

| Alternative | Why Rejected |
|-------------|-------------|
| Keep CronCreate for workflow triggers | Session-scoped, doesn't survive restarts, injects into supervisor — the exact problems 73.2 solved for heartbeats |
| Agent-internal timers (counter-based) | Claude cannot reliably maintain counters across context compressions; resets unpredictably |
| Filesystem watching (inotify/fswatch) | Requires new daemon capability; overkill for 1-2 workflow triggers |
| Wall-clock cron expressions in daemon | Over-complex for the scheduling needs; boot-time offsets are simpler |
| Prime-number intervals | Unnecessary with daemon-controlled staggering; adds cognitive overhead for operators |
| Single delivery mechanism for all | Heartbeats and triggers have different reliability requirements; hybrid is appropriate |

---

## References

- Research: `_bmad-output/planning-artifacts/croncreate-audit-research.md`
- Research: `_bmad-output/planning-artifacts/retrospector-persistence-research.md`
- Research: `_bmad-output/planning-artifacts/croncreate-viability-study.md`
- Research: `_bmad-output/planning-artifacts/multiclaude-operator-ux-research.md`
- Story 73.2: CronCreate heartbeat removal
- Story 73.4: Session Handoff Protocol (future integration point)
- Story 76.6: QUOTA_CHECK cron (future consumer of trigger system)
- `docs/operations/persistent-agent-ops.md`: Current operational docs (to be updated in Phase 3)
