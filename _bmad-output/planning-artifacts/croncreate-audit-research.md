# CronCreate Audit: What Scheduled Events Did We Lose?

**Date:** 2026-03-29
**Researcher:** worker/bold-wolf
**Provenance:** L3 (AI-autonomous research)
**Story:** 73.2 (post-deployment audit)

---

## Executive Summary

Story 73.2 removed 6 CronCreate heartbeat jobs and retained the SYNC_OPERATIONAL_DATA cron. The heartbeat removal was correct — the daemon wake loop replaced them. However, the SYNC_OPERATIONAL_DATA cron that was **retained in the startup checklist** relies on CronCreate, which is session-scoped and fragile. Additionally, two other scheduled behaviors documented in the codebase (BMAD PM audits via `/loop`, future QUOTA_CHECK cron for Story 76.6) were never CronCreate jobs but represent scheduling needs that the same fragility affects.

**Key finding:** The SYNC_OPERATIONAL_DATA cron is the only non-heartbeat CronCreate job that existed and was retained, but its CronCreate-based delivery mechanism has known reliability gaps (session-scoping, no restart survival). The retrospector persistence research (zealous-badger, 2026-03-29) already identified this as Gap 1.

---

## Question 1: What CronCreate Jobs Were Previously Registered?

### Complete CronCreate Inventory (pre-73.2 removal)

From Story 0.55 (commit `56dbc01c`, 2026-03-12) and Story 67.1 (commit `1a07dc02`):

| # | Cron Expression | Command | Purpose | Category |
|---|----------------|---------|---------|----------|
| 1 | `*/7 * * * *` | `multiclaude message send merge-queue HEARTBEAT` | Wake merge-queue polling loop | Heartbeat |
| 2 | `3-59/7 * * * *` | `multiclaude message send pr-shepherd HEARTBEAT` | Wake pr-shepherd polling loop | Heartbeat |
| 3 | `*/11 * * * *` | `multiclaude message send envoy HEARTBEAT` | Wake envoy polling rhythm | Heartbeat |
| 4 | `*/13 * * * *` | `multiclaude message send retrospector HEARTBEAT` | Wake retrospector analysis loop | Heartbeat |
| 5 | `*/23 * * * *` | `multiclaude message send project-watchdog HEARTBEAT` | Wake project-watchdog polling loop | Heartbeat |
| 6 | `5-59/23 * * * *` | `multiclaude message send arch-watchdog HEARTBEAT` | Wake arch-watchdog polling loop | Heartbeat |
| 7 | `0 */3 * * *` | `multiclaude message send project-watchdog SYNC_OPERATIONAL_DATA` | Trigger data sync pipeline | **Workflow trigger** |

**Source:** `docs/operations/persistent-agent-ops.md` (lines 79-104), `docs/stories/0.55.story.md` (lines 148-153), supervisor MEMORY.md startup checklist.

---

## Question 2: Which Were Heartbeats vs. Other Purposes?

### Pure Heartbeats (6 jobs — CORRECTLY REMOVED)

Jobs 1-6 above were pure heartbeats. They sent a generic "HEARTBEAT" message that triggered each agent's polling loop. The daemon wake loop (`wakeAgents()`, daemon.go:437) already sends role-specific prompts to every agent every 2 minutes, making these redundant. The viability study (`_bmad-output/planning-artifacts/croncreate-viability-study.md`) confirmed removal was safe.

**Replacement:** Daemon wake loop (2-min cycle, daemon-native, survives restarts, has deduplication).

### Non-Heartbeat Workflow Triggers (1 job — RETAINED BUT FRAGILE)

Job 7 (`SYNC_OPERATIONAL_DATA`) serves a distinct purpose: it triggers project-watchdog to check `docs/operations/` for uncommitted retrospector data files and create a data-sync PR if changes exist. This is a **workflow trigger**, not a heartbeat — it initiates a specific multi-step workflow (check → branch → commit → push → PR).

**Current status:** Retained in the supervisor MEMORY.md startup checklist as a CronCreate job. The supervisor must manually create this CronCreate entry on each restart.

### Documented But Not Yet Implemented

Two additional scheduled behaviors are documented in planning/architecture docs but were never CronCreate jobs:

| Behavior | Source | Interval | Status |
|----------|--------|----------|--------|
| BMAD PM sprint audit (`/loop 30m /bmad-bmm-sprint-status`) | MEMORY.md "BMAD PM Audit Loop" section | 30 min via `/loop` skill | Documented, manually invoked |
| QUOTA_CHECK cron | Story 76.6 (Not Started) | Every 5 min | Future — depends on Stories 76.1, 76.3, 76.5 |

---

## Question 3: Impact of Lost/Fragile Non-Heartbeat Jobs

### SYNC_OPERATIONAL_DATA — The Only Affected Job

**Purpose:** Ensures retrospector findings, checkpoint, and recommendations are committed to git every 3 hours so worker agents in isolated worktrees can access current operational data.

**Interval:** Every 3 hours (`0 */3 * * *`).

**Impact of not running:**
- Retrospector operational data accumulates only in the main checkout working tree
- Worker agents in worktrees cannot see current findings/recommendations
- Data is at risk of loss if the working tree is reset (`git checkout .`, `git clean -f`)
- The 6-hour staleness SLA (Standing Order #9) will be violated

**Evidence of past failures:** The retrospector persistence research (zealous-badger) found data-sync PRs cluster around active supervisor sessions with multi-day gaps between them (e.g., nothing between PR #822 on 2026-03-19 and PR #823 on 2026-03-26). This is consistent with the CronCreate job not being recreated after supervisor restarts.

**Known fragility (from `docs/operations/persistent-agent-ops.md` lines 107-112):**
1. **Session-scoped:** CronCreate jobs are lost on supervisor exit and auto-expire after 3 days
2. **Idle-only firing:** Crons only fire while the supervisor REPL is idle
3. **No guaranteed delivery:** If supervisor is busy when cron fires, that cycle is skipped
4. **Manual recreation required:** Supervisor must remember to run the CronCreate command on every restart

---

## Question 4: How Should We Restore These Scheduled Events?

### Option A: Keep CronCreate (Status Quo)

The SYNC_OPERATIONAL_DATA cron is already retained in the startup checklist. It works when the supervisor is running and remembers to create it.

**Pros:** No code changes needed. Already documented.
**Cons:** All the known fragility issues persist. Relies on supervisor discipline.

### Option B: Daemon-Native Periodic Triggers (Recommended — Story 73.6 Scope)

The daemon wake loop already runs every 2 minutes. It could be extended to send `SYNC_OPERATIONAL_DATA` to project-watchdog on a configurable interval (e.g., every 90th wake cycle = ~3 hours, or a separate timer).

**Pros:** Survives restarts. No supervisor involvement. Eliminates CronCreate dependency entirely.
**Cons:** Requires multiclaude daemon code changes. Story 73.6 is `Not Started`.

This was already identified as the recommended approach by:
- Story 73.2 AC2 (migrate to daemon-native in 73.6)
- Retrospector persistence research R2 (move SYNC trigger to daemon wake loop)
- Viability study recommendation ("Migrate to daemon-native in Story 73.6 if desired")

### Option C: Agent Definition Startup Task

Add a standing order to project-watchdog's definition: "Every 3 hours (approximately every 90th polling cycle), run the SYNC_OPERATIONAL_DATA workflow automatically."

**Pros:** No daemon changes. Agent-native.
**Cons:** Claude agents cannot track elapsed time. Without an external trigger, the agent has no way to know "3 hours have passed." The daemon wake nudge doesn't include timing information. This only works if the daemon wake prompt explicitly says "run data sync" every N cycles.

### Option D: project-watchdog Self-Triggers on Daemon Wake

Modify project-watchdog's HEARTBEAT/wake response to include a conditional data sync check. On every Nth wake (tracked by a counter), run the SYNC_OPERATIONAL_DATA workflow.

**Pros:** Leverages existing daemon wake loop. No daemon code changes.
**Cons:** Claude agents cannot reliably maintain counters across context compressions. The counter would reset unpredictably, leading to either too-frequent or too-infrequent syncs.

### Option E: Hybrid — Daemon Wake + Keyword Trigger

Add `SYNC_OPERATIONAL_DATA` to the daemon's wake message for project-watchdog, but only on certain cycles (e.g., configurable in daemon config). The daemon already sends role-specific messages; this would add a workflow trigger to certain messages.

**Pros:** Clean separation. Daemon handles timing, agent handles workflow.
**Cons:** Requires daemon config changes (similar effort to Option B).

---

## Recommendation

### Short-term (Now): Status quo is acceptable

The SYNC_OPERATIONAL_DATA CronCreate in the startup checklist works. The supervisor just needs to actually run it on startup. The staleness SLA standing order (#9) catches failures.

### Medium-term (Story 73.6): Migrate to daemon-native

Story 73.6 ("Daemon-Native Configurable Heartbeats") is already scoped to replace CronCreate with daemon-native triggers. The SYNC_OPERATIONAL_DATA cron should be migrated as part of this work. The story's AC4 explicitly calls for "Migration path from CronCreate documented."

### Future (Story 76.6): QUOTA_CHECK cron

When Story 76.6 is implemented, it should use the daemon-native mechanism from 73.6 rather than CronCreate. This is already implied by the story's background section ("Phase 1 implements only the cron component; daemon integration is Phase 2").

---

## Summary Table

| CronCreate Job | Status After 73.2 | Replacement | Action Needed |
|---------------|-------------------|-------------|---------------|
| 6x HEARTBEAT jobs | Removed | Daemon wake loop (2-min) | None — correctly replaced |
| SYNC_OPERATIONAL_DATA | Retained (CronCreate) | Daemon-native (Story 73.6) | Migrate when 73.6 is implemented |
| BMAD PM audit (`/loop`) | Never was CronCreate | N/A — manual `/loop` invocation | No change needed |
| QUOTA_CHECK (76.6) | Not yet implemented | Should use daemon-native from 73.6 | Plan for daemon-native from the start |

---

## Files Studied

- `docs/stories/73.2.story.md` — CronCreate heartbeat removal story
- `docs/stories/67.1.story.md` — SYNC_OPERATIONAL_DATA pipeline story
- `docs/stories/0.55.story.md` — Original heartbeat system story
- `docs/stories/76.6.story.md` — Future QUOTA_CHECK cron story
- `docs/operations/persistent-agent-ops.md` — Heartbeat schedule, sync pipeline docs
- `agents/project-watchdog.md` — SYNC_OPERATIONAL_DATA handler
- `_bmad-output/planning-artifacts/croncreate-viability-study.md` — 73.2 viability research
- `_bmad-output/planning-artifacts/retrospector-persistence-research.md` — Data persistence gaps
- `_bmad-output/planning-artifacts/multiclaude-operator-ux-research.md` — Operator UX research (R-007)
- Supervisor MEMORY.md — Current startup checklist
- Git history: commits `56dbc01c`, `6c14309b`, `35cecd8e`, `1a07dc02`
