# Research: Retrospector Data Persistence Problems

**Date:** 2026-03-29
**Researcher:** worker/zealous-badger
**Provenance:** L3 (AI-autonomous research)

## Problem Summary

Three interrelated problems cause retrospector operational data to be unreliable across multiclaude sessions:

1. **Stale checkpoint counter** — `prs_since_restart` persists across sessions, triggering false self-restart requests
2. **Uncommitted operational data** — findings and checkpoint accumulate in the main checkout but aren't reliably committed to git
3. **RBAC gap** — retrospector cannot commit its own data, and the existing sync pipeline has gaps

---

## Issue 1: Stale Checkpoint Counter Bug

### Current State

`docs/operations/retrospector-checkpoint.json` currently reads:
```json
{
  "version": 1,
  "last_pr": 841,
  "prs_since_restart": 19,
  "hours_since_restart": 0,
  ...
}
```

### The Bug

The retrospector agent definition (lines 274-278) specifies self-restart triggers:
- **20 PRs processed since last restart** → request restart
- **8 hours of continuous operation** → request restart

On startup, retrospector reads the checkpoint file (line 37-39). If `prs_since_restart` is 19 (as it is now), the very first PR processed will push it to 20, triggering an immediate restart request — even though the agent just started fresh with a full context window.

The counter reflects the *previous session's* cumulative work, not the current session's. A fresh spawn should start at 0 for both `prs_since_restart` and `hours_since_restart`.

### Root Cause

The checkpoint schema conflates two purposes:
1. **Resumable analytical state** (last_pr, mode_rotation_index, rolling_windows) — SHOULD persist across sessions
2. **Session-scoped resource tracking** (prs_since_restart, hours_since_restart) — SHOULD NOT persist across sessions

Both are written to the same file with no mechanism to distinguish "restored from checkpoint" vs. "continuing same session."

### Recommended Fix

**Option A (Minimal — recommended):** On startup, retrospector should reset session-scoped fields after reading the checkpoint:
- Set `prs_since_restart = 0`
- Set `hours_since_restart = 0`
- Preserve all other fields (last_pr, mode_rotation_index, rolling_windows)

This requires a one-line change to the startup sequence in `agents/retrospector.md` (after step 1 "Read checkpoint file"):
> After restoring checkpoint, reset `prs_since_restart` and `hours_since_restart` to 0 — these track current session resource usage, not historical state.

**Option B (Structural):** Split the checkpoint into two files:
- `retrospector-checkpoint.json` — analytical state (persists across sessions)
- `retrospector-session.json` — session metrics (ephemeral, not committed to git)

Option B is cleaner architecturally but requires more changes. Option A solves the immediate bug.

**Option C (Schema version bump):** Add a `session_id` field to the checkpoint. On startup, if `session_id` doesn't match the current spawn, reset session-scoped counters. This makes the single-file approach robust without splitting files.

---

## Issue 2: Uncommitted Operational Data Risk

### Current Pipeline

The data flow is:
1. **Retrospector** writes to `docs/operations/retrospector-{findings,checkpoint,recommendations}.jsonl/.json` in the main checkout
2. **CronCreate** fires every 3 hours: `multiclaude message send project-watchdog SYNC_OPERATIONAL_DATA`
3. **project-watchdog** checks for uncommitted changes, creates a `data-sync/<timestamp>` branch, commits, creates PR
4. **merge-queue** merges the data-sync PR

### Failure Points Observed

**Gap 1: CronCreate is session-scoped.** CronCreate jobs expire after 3 days and are lost on supervisor restart. If the supervisor restarts and doesn't recreate the cron (or if the startup checklist is missed), the entire sync pipeline stops. Evidence: the JSONL findings log shows data-sync PRs cluster around active sessions — PR #760, #802, #812, #819, #822 — with gaps of days between them (e.g., nothing between PR #822 on 2026-03-19 and PR #823 on 2026-03-26).

**Gap 2: 3-hour interval means up to 3 hours of data loss.** If multiclaude shuts down between sync cycles, any data written since the last sync is lost. The 6-hour staleness SLA acknowledges this but doesn't prevent it.

**Gap 3: project-watchdog must be alive and responsive.** If project-watchdog is context-exhausted or stuck when the SYNC message arrives, the message may be lost or delayed. The message queue has no guaranteed delivery.

**Gap 4: Main checkout state is the single point of failure.** All operational data lives only in the main checkout's working tree until committed. There's no backup, no WAL, no replication. A `git checkout .` or `git clean -f` would destroy all accumulated data.

### Evidence of Data Loss

The current checkpoint shows `prs_since_restart: 19` with `last_pr: 841`, but the committed JSONL in git only has entries up to PR #841 (109 lines total, covering PRs 713-841). The gap between PR #822 (2026-03-19) and PR #823 (2026-03-26) — a full week — suggests the sync pipeline was not running during that period. Any retrospector analysis during that week would have been lost if the main checkout was reset.

### Recommended Fix

**Short-term (Story-sized):**
1. Move SYNC_OPERATIONAL_DATA trigger from CronCreate to the daemon wake loop. The daemon already nudges agents every 2 minutes; it could send SYNC_OPERATIONAL_DATA to project-watchdog on a configurable interval (e.g., every 6th wake cycle = ~12 min). This eliminates the CronCreate session-scoping problem.

2. Add a shutdown hook: before killing a persistent agent session, the daemon should trigger a final SYNC_OPERATIONAL_DATA. This captures data accumulated since the last sync.

**Medium-term (Epic 73.4 integration):**
Story 73.4 (Session Handoff Protocol) already designs per-agent state persistence at `~/.multiclaude/agent-state/<repo>/<agent-name>/`. Retrospector's checkpoint.json and findings buffer should migrate to this location, with the daemon managing persistence. This is the right long-term home.

**Long-term (Dark Factory):**
Replace file-based persistence with a proper data store. The checkpoint is effectively a key-value store; findings are an append-only log. Both map naturally to SQLite or similar embedded databases that provide atomicity and crash recovery.

---

## Issue 3: Retrospector Storage and RBAC

### Current Authority Model

From `agents/retrospector.md` (lines 319-340), retrospector CAN:
- Write to `docs/operations/retrospector-findings.jsonl`
- Write to `docs/operations/retrospector-checkpoint.json`
- Write to `docs/operations/retrospector-recommendations.jsonl`

Retrospector CANNOT:
- Create PRs (explicitly deferred to Phase 2)
- Run git commands that modify working tree state
- Commit changes

### The RBAC Problem

Retrospector can *write files* but cannot *commit them*. This creates a dependency chain:
```
retrospector writes → project-watchdog commits → merge-queue merges
```

Three agents must be alive and coordinated for data to reach git. If any one fails, data accumulates in the working tree with no durability guarantee.

### Who Should Commit Retrospector's Data?

**Current answer (project-watchdog via SYNC_OPERATIONAL_DATA):** This works but is fragile due to the dependency chain described above.

**Alternative 1: Grant retrospector limited commit permissions for its own data files only.**
- Pros: Eliminates the 3-agent dependency chain. Retrospector can commit after each checkpoint write.
- Cons: Violates the observer/participant separation principle (party mode decision D-ODP1). Retrospector would need to create branches and PRs, participating in the merge queue it monitors.
- **Recommendation: REJECT** — the architectural principle is sound. An observer that participates in what it observes creates feedback loops.

**Alternative 2: Staging area pattern — retrospector writes to a staging location, daemon commits.**
- Retrospector writes to `~/.multiclaude/agent-state/ThreeDoors/retrospector/` (outside the git repo)
- The daemon periodically copies staged data into the repo and triggers a commit via project-watchdog
- This separates the write (fast, local, always works) from the commit (requires git coordination)
- Pros: No RBAC change for retrospector. Data survives even if git working tree is reset. Daemon is always running.
- Cons: Adds complexity. Data exists in two places temporarily.
- **Recommendation: ADOPT for 73.4** — this aligns with the Session Handoff Protocol design.

**Alternative 3: project-watchdog self-triggers on file modification (inotify/fswatch).**
- project-watchdog watches `docs/operations/` for changes and commits when files are modified
- Pros: Eliminates cron dependency entirely. Near-real-time persistence.
- Cons: Requires filesystem watching capability that Claude agents don't have. Would need a daemon feature.
- **Recommendation: DEFER** — good idea but requires daemon work beyond current scope.

### Recommended Approach

Keep the current RBAC model (retrospector writes, project-watchdog commits) but strengthen the pipeline:

1. **Immediate:** Fix the checkpoint counter bug (Issue 1, Option A)
2. **Story 73.4 integration:** Move retrospector's working data to `~/.multiclaude/agent-state/` with daemon-managed persistence
3. **Daemon enhancement:** Replace CronCreate SYNC trigger with daemon-native periodic trigger (eliminates session-scoping fragility)
4. **Daemon enhancement:** Add shutdown-triggered final sync (eliminates inter-cycle data loss)

---

## Summary of Recommendations

| # | Issue | Fix | Effort | Priority |
|---|-------|-----|--------|----------|
| R1 | Stale checkpoint counter | Reset session counters on startup (Option A) | Trivial — agent def change | P0 (causes false restarts NOW) |
| R2 | CronCreate session-scoping | Move SYNC trigger to daemon wake loop | Small — daemon change | P1 |
| R3 | Shutdown data loss | Add daemon shutdown hook for final sync | Small — daemon change | P1 |
| R4 | Working tree SPOF | Migrate to `~/.multiclaude/agent-state/` staging area | Medium — Story 73.4 | P2 |
| R5 | 3-agent dependency chain | Daemon-managed persistence (bypasses project-watchdog for commits) | Medium — daemon feature | P2 |

### Rejected Approaches

| Approach | Why Rejected |
|----------|-------------|
| Grant retrospector commit authority | Violates observer/participant separation (D-ODP1) |
| Supervisor commits directly | Violates coordination/execution separation |
| Split checkpoint into two files (Option B) | Overengineered for the immediate bug; Option A suffices |
| Filesystem watching (inotify) | Requires daemon capability not yet available |

---

## Files Studied

- `agents/retrospector.md` — agent definition, checkpoint schema, restart triggers
- `agents/project-watchdog.md` — SYNC_OPERATIONAL_DATA handler
- `docs/operations/retrospector-checkpoint.json` — current checkpoint state
- `docs/operations/retrospector-findings.jsonl` — 109 entries, PRs 713-841
- `docs/operations/persistent-agent-ops.md` — operational guide, heartbeat schedule, sync pipeline docs
- `docs/stories/67.1.story.md` — cron-triggered sync pipeline (Done, PR #757)
- `docs/stories/73.2.story.md` — CronCreate heartbeat removal (Deployed)
- `docs/stories/73.4.story.md` — Session Handoff Protocol (Not Started)
- `_bmad-output/planning-artifacts/retrospector-data-pipeline-party-mode.md` — original pipeline design decisions
