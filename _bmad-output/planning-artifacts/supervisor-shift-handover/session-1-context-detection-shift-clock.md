# Party Mode Session 1: Context Window Detection and Shift Clock Design

**Date:** 2026-03-11
**Topic:** Supervisor Shift Handover — Context Window Detection and Shift Clock Design
**Participants:** Winston (Architect), John (PM), Murat (TEA), Amelia (Dev)

---

## Problem Statement

multiclaude runs a supervisor agent that coordinates workers, persistent agents, and project state. As conversations grow, the context window fills and performance degrades. We need a mechanism to detect when context is shrinking and trigger a handover to a fresh supervisor.

## Key Questions Explored

1. How do we detect context window utilization?
2. What thresholds should trigger handover?
3. Should the shift clock be time-based, usage-based, or hybrid?
4. How does Claude Code's automatic context compression interact with this?
5. What's the minimum viable shift clock implementation?
6. Should the supervisor self-report its health, or should an external monitor watch it?
7. Trade-offs between proactive handover vs reactive

## Discussion

### Observable Signals (Architect)

Claude Code doesn't expose a "tokens used" counter to the running agent. Observable signals:

1. **Context compression events** — When Claude Code compresses prior messages, the agent sees a system notification. Can be intercepted via hooks or by monitoring the conversation transcript JSONL file. Each compression event is a strong signal that the window is filling.
2. **Message count / conversation length** — Crude but reliable. Count tool calls or user turns. After N interactions, assume degradation is approaching.
3. **Response quality degradation** — This is the *symptom*, not the signal. By the time you detect it, you're already past optimal. **Rejected as primary trigger.**
4. **External monitoring** — A daemon process watching the JSONL transcript file size, or counting messages in it. Most architecturally clean approach because it decouples detection from the agent being monitored.

**Recommendation:** External daemon monitoring of the transcript JSONL, with compression events as a secondary confirmation signal. The supervisor shouldn't self-diagnose — a degrading supervisor is the worst entity to trust for self-assessment.

### Threshold Tiers (PM)

The trust problem: why trust a degrading supervisor to accurately report its own degradation? That's like asking someone with a concussion to evaluate their own cognitive function.

Three threshold tiers:

- **Green zone (0-60% utilization)**: Normal operation. No action needed.
- **Yellow zone (60-80%)**: Advisory. Incoming supervisor should start warming up. Active workers should be catalogued.
- **Red zone (80%+)**: Mandatory handover. The current supervisor completes no new dispatches and begins state serialization.

**Cost asymmetry analysis:** Premature handover = mild disruption, fresh context, maybe loses some nuance. Late handover = degraded decisions, missed context, potential worker mismanagement. The asymmetry clearly favors erring on the side of early handover. Set thresholds conservatively.

### Risk Analysis (TEA)

Highest-risk failure modes:

1. **False negative** — the supervisor thinks it's fine but it's actually degraded
2. **Oscillation** — the shift clock triggers, handover starts, but the new supervisor fills quickly because the handover itself is context-heavy

Requirements for testing:

1. **Measurable proxy metrics** — Compression event count, JSONL file size. Can write assertions against these.
2. **Calibration data** — Empirically determine "at what compression count does quality actually degrade?" Requires controlled experiments: run supervisors through identical workloads, measure decision quality at various context levels.
3. **The monitor itself must be non-Claude** — A shell script, a cron job, a daemon hook. NOT another LLM agent. The monitor must be deterministic.

The multiclaude daemon already runs a refresh loop every 5 minutes. Adding transcript monitoring to that loop is the minimum viable approach. Calculate: JSONL size, message count, compression event count. Emit as metrics. Let a threshold trigger the handover.

### Implementation Approach (Dev)

The JSONL transcript lives at `~/.claude/projects/<project-hash>/<session-id>.jsonl`. Monitoring:

```bash
# Pseudocode for shift clock check
jsonl_size=$(stat -f%z "$transcript_path")
compression_count=$(grep -c '"type":"context_compression"' "$transcript_path")
message_count=$(grep -c '"role":"assistant"' "$transcript_path")
```

Could be a new subcommand: `multiclaude supervisor health` that returns these metrics. The daemon's 5-minute refresh loop calls it. When thresholds are crossed, writes a signal file or sends a message via `multiclaude message send supervisor "SHIFT_HANDOVER_ADVISORY"`.

The shift clock should be **hybrid**: time-based floor (no handover before 30 min to avoid thrashing) + usage-based ceiling (force handover at N compression events regardless of time). Time-only misses burst usage. Usage-only might trigger on a 5-minute sprint.

**Minimum viable: a shell script in the daemon loop. No Go code changes needed for v1.**

### Architectural Decisions

**Handover initiation:** External daemon initiates, not the supervisor. The outgoing supervisor participates in handover (serializing state) but doesn't decide when. Principle: don't let the monitored entity control its own monitoring.

**Task boundaries (PM):** Don't trigger handover mid-dispatch. Wait for natural seams: after a worker completes, after a PR merges, after a status check. The daemon sets a "handover pending" flag, and the supervisor checks it at natural break points. Graceful handover without mid-operation disruption.

**Safe handover point definition (TEA):** No `multiclaude work` commands in the last 60 seconds, no pending message acknowledgments, no active tool calls. This is a testable condition.

## Decisions Summary

| Decision | Adopted | Rejected | Rationale |
|---|---|---|---|
| Detection mechanism | External daemon monitoring JSONL transcript | Self-reporting by supervisor | Degrading agent can't trust its own assessment |
| Metrics | Compression count + JSONL size + message count | Response quality analysis | Quality is a lagging indicator; by the time it's detectable, damage is done |
| Threshold model | Three-tier (green/yellow/red) | Binary (go/no-go) | Allows warm-up period for incoming supervisor |
| Clock type | Hybrid (time floor + usage ceiling) | Pure time-based or pure usage-based | Time misses burst usage; usage misses idle periods |
| Handover initiation | External daemon decides, supervisor participates | Supervisor self-initiates | Separation of monitoring from monitored entity |
| Handover timing | Natural seam (task boundary) | Immediate on threshold | Prevents mid-operation disruption |

## Open Questions (for later sessions)

- What exactly does the supervisor serialize? (Session 4)
- What if the daemon detects degradation but supervisor can't respond? (Session 5)
- How does the incoming supervisor inherit running workers? (Session 2)
