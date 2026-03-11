# Party Mode Session 3: Standby Supervisor Feasibility

**Date:** 2026-03-11
**Topic:** Supervisor Shift Handover — Hot Standby vs Warm Standby vs Cold Start
**Participants:** Winston (Architect), John (PM), Murat (TEA), Amelia (Dev)

---

## Problem Statement

Should we maintain a standby supervisor ready to take over, or start fresh when handover is needed? What's the trade-off between readiness and cost/complexity?

## Options Evaluated

### 1. Hot Standby (REJECTED)

A second supervisor runs continuously, shadowing the primary. Reads the same messages, maintains parallel state.

**Why rejected:**
- **Doubles Claude API costs** for zero additional throughput during normal operation
- Creates true **split-brain risk** — two agents interpreting the same messages
- The shadow supervisor fills its own context window just by watching
- **Architectural nightmare** — two supervisors parsing the same message stream = chaos

### 2. Warm Standby (REJECTED)

Spawned in advance when shift clock hits yellow zone. Reads MEMORY.md and ROADMAP.md to warm up, doesn't process messages until handover.

**Why rejected:**
- **The upgrade problem kills it** — Claude can't hot-swap its system prompt. You can't tell a running Claude instance "now you're the active supervisor." You'd have to kill and respawn it anyway, making it effectively a cold start with extra steps.
- Pre-reading files is wasted if the instance gets killed and restarted
- Adds ~200 lines of bash + new message protocol for marginal benefit (~30-40 seconds saved)
- Managing two supervisor lifecycles adds complexity with little payoff

**Cost analysis (PM):** Bounded at 2-3 minutes of Claude time for pre-reading, but the upgrade problem means those tokens are wasted.

### 3. Cold Start (ADOPTED)

No pre-spawning. When handover triggers, daemon spawns a fresh supervisor from scratch. It reads the handover state file and bootstraps.

**Why adopted:**
- Simplest implementation (~50 lines of bash)
- Total cold start time: **~60-90 seconds** (acceptable gap)
  - `multiclaude agents spawn`: ~5-10 seconds
  - Claude reading CLAUDE.md + system prompt: ~10-15 seconds
  - Reading MEMORY.md: ~5 seconds
  - Reading ROADMAP.md: ~5 seconds
  - Reading shift-state.yaml: ~3 seconds
  - Worker pings and responses: ~30-60 seconds
- No wasted API tokens
- No split-brain risk
- No lifecycle management complexity

## Key Optimization: Daemon-Maintained Rolling Snapshot

Instead of relying on the (potentially degraded) outgoing supervisor to serialize all state at handover time, the **daemon maintains a rolling snapshot** of system state:

- Updated every 5 minutes as part of the daemon refresh loop
- Pulls from external sources: `multiclaude worker list`, `multiclaude message list`, tmux session list
- Does NOT require supervisor cooperation for the base snapshot

**At handover time:** Daemon asks the outgoing supervisor "anything to add that only you know?" — pending decisions, priorities, context not in the message system. Supervisor adds its delta and exits.

**Benefits:**
- Minimizes cognitive load on a degraded supervisor — it only contributes what it uniquely knows
- Snapshot can be validated against ground truth (worker list, message list, tmux sessions) as a periodic health check
- If snapshot drifts from reality, that's caught before handover, not during

## Decisions Summary

| Decision | Adopted | Rejected | Rationale |
|---|---|---|---|
| Standby model | Cold start | Hot standby, Warm standby | Hot doubles cost; Warm has unsolvable upgrade problem |
| State preparation | Daemon-maintained rolling snapshot | Supervisor serializes at handover | Degraded supervisor is unreliable for complex serialization |
| Supervisor delta | Outgoing adds only unique context | Outgoing writes everything | Minimizes load on degraded instance |
| Transition gap | 60-90 seconds acceptable | Zero-downtime requirement | Workers are independent; they continue during gap |
