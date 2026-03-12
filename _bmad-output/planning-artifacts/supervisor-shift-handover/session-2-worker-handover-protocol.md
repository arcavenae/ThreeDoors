# Party Mode Session 2: Worker Handover Protocol

**Date:** 2026-03-11
**Topic:** Supervisor Shift Handover — How to Transfer Ownership of Running Workers
**Participants:** Winston (Architect), John (PM), Murat (TEA), Amelia (Dev), Bob (SM)

---

## Problem Statement

When a supervisor hands off to a fresh instance, running workers (tmux sessions with their own Claude instances) must continue uninterrupted. The incoming supervisor needs to know what each worker is doing, what they've reported, and what decisions are pending.

## Key Insight: Workers Are Independent

Workers in multiclaude are **not owned by the supervisor process** — they're independent tmux sessions with their own Claude instances. The supervisor's relationship to workers is purely informational: it dispatched them, it receives their messages, it tracks their status.

"Transfer ownership" = transfer the supervisor's **mental model** of what workers are doing.

## Approaches Considered

### 1. State File Approach (ADOPTED — combined with #3)
Outgoing supervisor writes a structured handover document listing all active workers, their tasks, last known status, and pending messages. Incoming supervisor reads it on startup.

**Pros:** Immediate context, structured, deterministic
**Cons:** Point-in-time snapshot, may be stale by the time incoming reads it

### 2. Message Replay Approach (REJECTED)
Incoming supervisor replays the message history from `multiclaude message list` to reconstruct context.

**Pros:** Uses existing infrastructure
**Cons:** Messages lack full context (a completion message doesn't explain what the task was), high token cost to replay everything, may not capture verbal context from the supervisor's conversation

### 3. Worker Self-Report Approach (ADOPTED — combined with #1)
Incoming supervisor sends each active worker a "status ping" and they report back what they're doing.

**Pros:** Fresh, verified information directly from workers
**Cons:** Workers may be mid-task and slow to respond, adds latency to handover

### Combined Approach (ADOPTED)
State file for immediate context, worker pings for verification. State file gets the incoming supervisor oriented immediately; worker pings confirm/update that picture.

## The Split-Brain Problem

During handover, which supervisor do workers report to? If both are running simultaneously, a worker might send a completion message that the outgoing supervisor sees but the incoming one doesn't.

### Authority Transfer Protocol

- **Phase 1: Outgoing is authoritative.** It writes the handover state. Still processing messages.
- **Phase 2: Transition window (30-60 seconds).** No new dispatches from either supervisor. Workers continue working but hold non-urgent messages.
- **Phase 3: Incoming is authoritative.** It reads handover state, pings workers, processes queue.

The multiclaude messaging system uses files on disk (`~/.multiclaude/messages/`), not in-memory queues. Messages persist regardless of supervisor lifecycle. The incoming supervisor runs `multiclaude message list` and processes everything.

**Key invariant:** At no point are zero supervisors watching the workers. The outgoing stays alive until the incoming is ready. Conversely, at no point should TWO supervisors be dispatching work.

The daemon is the mutex for authority.

## Handover State File Format

```yaml
# ~/.multiclaude/handover/ThreeDoors/shift-state.yaml
timestamp: "2026-03-11T14:30:00Z"
outgoing_supervisor: "gentle-hawk"
active_workers:
  - name: "bold-eagle"
    task: "Implement story 42.3"
    story_file: "docs/stories/42.3.story.md"
    branch: "work/bold-eagle"
    status: "implementing"
    last_message: "Running tests, 3/5 passing"
    dispatched_at: "2026-03-11T13:45:00Z"
  - name: "swift-fox"
    task: "Fix CI lint failures on PR #567"
    branch: "work/swift-fox"
    status: "debugging"
    last_message: "Found the issue, pushing fix"
    dispatched_at: "2026-03-11T14:10:00Z"
persistent_agents:
  - name: "merge-queue"
    status: "active"
    pending_prs: ["#565", "#566"]
  - name: "pr-shepherd"
    status: "active"
    rebasing: "#564"
pending_decisions:
  - "Worker bold-eagle asked about scope ambiguity in 42.3 AC #3 — not yet resolved"
open_issues_being_triaged:
  - issue: "#89"
    stage: "PM examination"
priorities:
  - "Story 42.3 is blocking Epic 42 completion"
  - "PR #565 has been open 2 days — needs merge-queue attention"
  - "Issue #89 triage should complete before EOD"
memory_file: "MEMORY.md"
```

## Shift Handover Protocol (v1)

1. **Daemon triggers handover** (from Session 1's shift clock)
2. **Outgoing supervisor enters handover mode:**
   - Stops dispatching new workers
   - Waits for natural seam (no active tool calls)
   - Writes `shift-state.yaml` with workers, agents, decisions, priorities
   - Acknowledges all pending messages
   - Sends `multiclaude message send supervisor "HANDOVER_COMPLETE"`
3. **Daemon spawns incoming supervisor** with task: "Read shift-state.yaml and assume control"
4. **Incoming supervisor startup sequence:**
   - Reads `shift-state.yaml`
   - Reads `MEMORY.md`
   - Reads `ROADMAP.md` for scope context
   - Runs `multiclaude message list` for messages that arrived during transition
   - Pings each active worker: "Status check — new supervisor online"
   - Processes worker responses, updating its mental model
5. **Daemon kills outgoing supervisor** after incoming confirms ready
6. **Incoming supervisor resumes normal operation**

## Critical Test Scenarios (TEA)

1. **Worker completes during transition window** — Does the completion message survive? (Yes — file-based message queue persists)
2. **Worker sends error during handover** — Must not be lost in the "nobody's watching" window
3. **Persistent agent sends merge notification** — Which supervisor gets it? (Whichever is authoritative per daemon's pointer)

**Post-handover assertion:** `multiclaude message list` shows zero unacknowledged messages. If any exist, incoming supervisor must process them before accepting new work.

## Sprint Context Considerations (SM)

The handover state must include **priorities** — not just what's running, but what matters most. The outgoing supervisor has accumulated understanding of sprint priorities, which stories are more urgent, which PRs are blocking others. This context lives in memory, not in the worker list.

Incoming supervisor must read:
- `shift-state.yaml` (immediate operational state)
- `MEMORY.md` (persistent cross-session context)
- Sprint plan / `ROADMAP.md` (scope and priority context)
