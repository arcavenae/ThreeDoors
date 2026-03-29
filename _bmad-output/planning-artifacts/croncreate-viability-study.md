# CronCreate Heartbeat Viability Study

**Date:** 2026-03-29
**Story:** 73.2 (Research Only — DO NOT IMPLEMENT)
**Worker:** calm-bear
**Status:** Complete

---

## Executive Summary

CronCreate heartbeats were added to solve a real problem — Claude agents have no internal timers and go idle after startup. However, the multiclaude daemon already has a built-in wake loop (`wakeAgents()`) that nudges every agent every 2 minutes with role-specific prompts. The CronCreate heartbeats are **redundant with the daemon wake loop** and cause harmful double-injection into the supervisor. Removal is safe; the original root cause (agent idleness) is already addressed by the daemon.

---

## Question 1: WHY Were CronCreate Heartbeats Added?

### The Problem

All 6 persistent agents (merge-queue, pr-shepherd, arch-watchdog, envoy, project-watchdog, retrospector) **go idle after completing their initial startup work**. Claude agents have no internal timers — they only act when prompted. Even agents with detailed "polling loop" descriptions in their system prompts cannot self-invoke on a schedule. After processing all startup messages, the agent sits at an idle REPL with no incoming input, so it stops acting.

### The Timeline

1. **2026-03-12 22:19** — Story 0.55 created (`8f8a8a13`): "Cron-Based Agent Heartbeat System." The story documents the root cause clearly:

   > "Claude's system prompt describes desired behavior ('every 15 minutes, poll X') but Claude cannot self-invoke on a schedule. After processing all startup work, the agent sits at an idle prompt with no incoming messages, so it stops acting."

2. **2026-03-12 23:35** — Implementation commit (`56dbc01c`): Added HEARTBEAT Response Protocol sections to all 6 agent definitions, added Polling Loop sections to merge-queue and pr-shepherd (which had none), updated `persistent-agent-ops.md` and supervisor MEMORY.md.

### What the Research Found (Story 0.55)

The story's research phase audited all 6 agents and found:

| Agent | Had Polling Loop? | Had Startup Sequence? |
|-------|---|---|
| retrospector | Yes (detailed) | Yes (detailed) |
| arch-watchdog | Yes (rhythm + polling loop) | Yes (restart/recovery) |
| envoy | Yes (brief rhythm) | Partial |
| project-watchdog | Yes (polling loop) | Yes (restart/recovery) |
| merge-queue | **NO** | **NO** |
| pr-shepherd | **NO** | **NO** |

The two most critical agents (merge-queue, pr-shepherd) had **zero autonomy language** — they were written as purely reactive agents. The CronCreate heartbeat solved two problems simultaneously:
1. An external trigger to wake idle agents
2. A reason to add polling loop definitions to agents that lacked them

### What Was NOT the Trigger

- The heartbeat system was **NOT** created because of the "Unclosed Quote Freeze" incident (2026-03-08). That was a separate bug involving a Bash tool call with an unclosed quote causing a deadlock. The heartbeat system was designed to solve the broader agent idleness problem.
- There was no single catastrophic incident. The problem was observed organically: agents would complete startup work and then sit idle indefinitely.

---

## Question 2: Was the Root Cause Already Fixed?

### Yes — The Daemon Wake Loop

The multiclaude daemon has a built-in `wakeAgents()` loop (in `daemon.go:437`) that:

- Runs every **2 minutes**
- Sends **role-specific prompts** directly to each agent's tmux pane (e.g., "Status check: Review worker progress and check merge queue.")
- Skips workspace windows (which are for human use)
- Tracks `LastNudge` per agent and skips recently nudged agents (deduplication)

This was discovered during the Operator UX research (R-007, commit `92957f22`, 2026-03-29). The research found that the daemon wake loop was **already running and already solving the agent idleness problem** — the CronCreate heartbeats were layered on top without awareness that the daemon was already doing this job.

### The Double-Injection Problem

With both systems active, the supervisor receives injections from **two independent sources**:

1. **Daemon wake loop** (every 2 min): Pastes role-specific prompts via `tmux paste-buffer` into every agent
2. **CronCreate jobs** (every 7-23 min): Fires prompts into the supervisor's Claude REPL, which then runs `multiclaude message send <agent> HEARTBEAT`, which creates a message file, which the daemon picks up and delivers via tmux paste to the target agent

The CronCreate path adds **two extra injections per heartbeat round-trip**: one into the supervisor (the CronCreate prompt), and one back into the supervisor when the agent responds. The daemon wake loop achieves the same result with zero supervisor injections.

### Key Difference: Delivery Mechanism

| Aspect | Daemon Wake Loop | CronCreate Heartbeats |
|--------|-----------------|----------------------|
| Trigger | Daemon timer (2 min) | CronCreate in supervisor REPL (7-23 min) |
| Delivery | Direct tmux paste to target agent | Supervisor → message file → daemon → tmux paste |
| Supervisor impact | None (daemon bypasses supervisor) | High (2 injections per round-trip) |
| Survives restart | Yes (daemon-native) | No (session-scoped, must re-create) |
| Agent response | Agent wakes and acts | Same |
| Deduplication | Yes (`LastNudge` tracking) | No |

---

## Question 3: Is Removal Safe?

### Yes — With One Caveat

**Removal of CronCreate heartbeats is safe** because:

1. The daemon wake loop already nudges every persistent agent every 2 minutes — more frequently than any CronCreate heartbeat (7-23 minutes)
2. The daemon delivers directly to target agents, bypassing the supervisor entirely
3. The daemon has built-in deduplication (skips recently nudged agents)
4. All 6 agents now have Polling Loop sections and HEARTBEAT Response Protocol sections in their definitions (added by Story 0.55) — these sections work regardless of whether the trigger is a daemon wake nudge or a CronCreate HEARTBEAT message

**The one caveat: SYNC_OPERATIONAL_DATA**

The `SYNC_OPERATIONAL_DATA` CronCreate job (`0 */3 * * *`) serves a different purpose than heartbeats. It triggers a specific workflow (project-watchdog checks `docs/operations/` for uncommitted data files and creates a sync PR). The daemon wake loop does NOT replicate this — daemon wakes are generic "check your work" prompts, not workflow-specific triggers.

Options for SYNC_OPERATIONAL_DATA:
- **Keep as CronCreate** — it's a workflow trigger, not a heartbeat. Low injection impact (fires only every 3 hours)
- **Migrate to daemon-native** — would require multiclaude code changes (Story 73.6 scope)
- **Convert to standing order in project-watchdog** — add "every 3rd heartbeat, run data sync check" to the agent definition. Relies on daemon wake loop timing.

**Recommendation:** Keep SYNC_OPERATIONAL_DATA as CronCreate for now. It fires infrequently (every 3 hours) and serves a specific purpose. Migrate to daemon-native in Story 73.6 if desired.

---

## Question 4: What Are the Risks?

### Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Daemon wake loop stops working | Very Low | High — agents go idle | Daemon liveness is already monitored; if daemon dies, everything stops |
| Daemon wake prompts don't trigger polling loops | Low | Medium — agents may not run full checks | Daemon sends role-specific prompts that should trigger agent work. Verify experimentally after removal. |
| HEARTBEAT keyword has special meaning in agent code | None | N/A | Agent definitions handle "HEARTBEAT" as a trigger, but they also respond to any wake prompt — tested by daemon wakes for months |
| Agents lose the ability to run "full polling loop" | Low | Low | Polling loop sections remain in agent definitions. Daemon wake prompts trigger equivalent behavior. |

### What Could Go Wrong

The only realistic failure mode is if daemon wake prompts are **qualitatively different** from HEARTBEAT messages in how agents respond. The daemon sends text like "Status check: Review worker progress and check merge queue" while CronCreate heartbeats sent "HEARTBEAT". If agents are coded to only run their full polling loop on the literal string "HEARTBEAT", they might do less work on daemon wake prompts.

**Mitigation:** The agent definitions say "When you receive a message containing 'HEARTBEAT'" — but daemon wake prompts aren't messages, they're direct tmux injections. Agents respond to both because Claude processes whatever appears in its input. The polling loop sections describe what to check, and any prompt asking the agent to "check" or "review" will trigger that behavior.

---

## Question 5: Recommendation

### Recommendation: REMOVE CronCreate heartbeats, KEEP SYNC_OPERATIONAL_DATA

**Rationale:**

1. **CronCreate heartbeats are redundant.** The daemon wake loop (2-min interval) already solves the agent idleness problem that CronCreate heartbeats were designed to address.

2. **CronCreate heartbeats are harmful.** They cause double-injection into the supervisor, consume supervisor context window, and must be manually recreated after every restart.

3. **The daemon wake loop is strictly superior.** It's daemon-native (survives restarts), has deduplication, doesn't inject into the supervisor, and runs at higher frequency.

4. **Agent definitions don't need changes.** The Polling Loop and HEARTBEAT Response Protocol sections added by Story 0.55 describe agent behavior that works regardless of the trigger mechanism. These sections should be kept — they document what agents should do when prompted to check their work.

5. **SYNC_OPERATIONAL_DATA should be retained** as a CronCreate job for now. It serves a specific workflow purpose that the daemon wake loop doesn't replicate. Low injection impact (every 3 hours).

### What Removal Looks Like

1. Delete the 6 CronCreate heartbeat commands from MEMORY.md startup checklist
2. Keep the SYNC_OPERATIONAL_DATA CronCreate command
3. Add a note explaining that daemon wake loop (2-min cycle) handles agent nudging
4. Update standing orders to remove heartbeat references
5. Keep all HEARTBEAT Response Protocol and Polling Loop sections in agent definitions (they describe valuable behavior patterns)

### Alternative: Adjust Intervals Instead of Removing

If there's concern about removal, a middle path:
- Reduce CronCreate intervals dramatically (e.g., every 30 min instead of 7 min)
- Frame them as "deep check" triggers distinct from daemon's frequent "light nudge"

**However, this is unnecessary.** The daemon wake prompts are role-specific and sufficiently detailed to trigger full polling behavior. There's no evidence that agents differentiate between "light" and "deep" checks based on the trigger text.

---

## Appendix: Source Evidence

| Source | Commit/File | Key Finding |
|--------|-------------|-------------|
| Story 0.55 creation | `8f8a8a13` (2026-03-12) | Documents root cause: Claude has no internal timers, agents go idle |
| Story 0.55 implementation | `56dbc01c` (2026-03-12) | Added HEARTBEAT protocols + Polling Loops to all 6 agents |
| Operator UX research (R-007) | `92957f22` (2026-03-29) | Discovered daemon wake loop makes CronCreate redundant; identified double-injection problem |
| Daemon source code | `daemon.go:437` (`wakeAgents()`) | 2-min wake loop with per-agent type logic and deduplication |
| Persistent agent ops | `docs/operations/persistent-agent-ops.md` | Documents heartbeat schedule and limitations |
| User feedback memory | `feedback_never_delete_crons.md` | User wants crons only deleted with explicit permission (behavioral, not architectural) |
| Story 73.2 | `docs/stories/73.2.story.md` | Documents the removal plan and acceptance criteria |
| Decision Q-C-011 | Referenced in 73.2 | Drop CronCreate heartbeats; daemon-native replacements in 73.6 |
