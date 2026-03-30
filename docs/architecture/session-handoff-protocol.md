# Session Handoff Protocol for Persistent Agents

> Design specification for agent state persistence across restarts.
> Extends Epic 58 (Supervisor Shift Handover) to all persistent agents.
> Decision: Q-C-010 — session handoff is a daemon feature.

## Problem

Persistent agents (merge-queue, pr-shepherd, envoy, project-watchdog, arch-watchdog, retrospector) lose all accumulated context when restarted. This means:
- merge-queue forgets which PRs it was tracking and merge validation state
- pr-shepherd loses rebase queue and conflict resolution progress
- envoy forgets triage progress on open issues
- project-watchdog loses its processed-PR correlation list
- arch-watchdog loses its processed-PR list and pending architecture reviews
- retrospector already has partial persistence (checkpoint.json) but session-scoped counters are conflated with analytical state

Restarts happen frequently: every 4-6 hours proactively (context exhaustion prevention), plus ad-hoc restarts for definition updates or crashes.

## Design

### State Directory Structure (AC1)

Each persistent agent gets a state directory outside the git repo:

```
~/.multiclaude/agent-state/<repo>/<agent-name>/
  handoff.md        -- structured handoff notes (human-readable)
  session.jsonl     -- breadcrumb actions from current/recent sessions
  context.json      -- machine-readable state (tracked PRs, pending work, etc.)
```

For ThreeDoors:
```
~/.multiclaude/agent-state/ThreeDoors/merge-queue/
~/.multiclaude/agent-state/ThreeDoors/pr-shepherd/
~/.multiclaude/agent-state/ThreeDoors/envoy/
~/.multiclaude/agent-state/ThreeDoors/project-watchdog/
~/.multiclaude/agent-state/ThreeDoors/arch-watchdog/
~/.multiclaude/agent-state/ThreeDoors/retrospector/
```

**Why outside the git repo:** State files are per-machine, per-session artifacts. Committing them would pollute the repo with ephemeral data and create merge conflicts between agents writing to the same files. The `~/.multiclaude/` directory is already the home for multiclaude state.

### File Formats

#### `handoff.md` — Structured Handoff Notes (AC3)

Written by the agent before shutdown. Read by the daemon and injected into the agent's startup prompt on restart.

```markdown
# Session Handoff — <agent-name>
## Timestamp: <ISO 8601 UTC>

## In Progress
<!-- Work currently being done. Be specific: PR numbers, story IDs, branch names. -->

## Recently Completed
<!-- Work finished this session. Include outcomes: merged, created, resolved. -->

## Blocked/Waiting
<!-- Items awaiting external action. Include who/what is blocking. -->

## Key Decisions
<!-- Decisions made this session that affect future work. Include rationale. -->

## Warnings
<!-- Issues or risks the next session should know about. -->
```

**Size limit:** 5KB max. The handoff must fit in the startup prompt without consuming excessive context. Agents should be concise — include PR numbers and story IDs, not full descriptions.

**Rotation:** On startup, the daemon moves the previous `handoff.md` to `handoff-<timestamp>.md` before injecting it. Keep the last 3 handoffs; delete older ones.

#### `session.jsonl` — Breadcrumb Log

Append-only log of significant actions during the session. Agents write breadcrumbs as they work; the log provides an audit trail and helps the next session understand what happened.

```jsonl
{"ts":"2026-03-29T14:30:00Z","action":"merge","detail":"Merged PR #850, CI green, scope valid"}
{"ts":"2026-03-29T14:35:00Z","action":"emergency","detail":"Main CI red after PR #851, entered emergency mode"}
{"ts":"2026-03-29T14:50:00Z","action":"spawn","detail":"Spawned worker to fix CI: multiclaude work 'Fix CI for PR #851'"}
{"ts":"2026-03-29T15:10:00Z","action":"resolve","detail":"Emergency resolved, main CI green, resuming merges"}
```

**Fields:**
- `ts` — ISO 8601 UTC timestamp
- `action` — category: `merge`, `rebase`, `triage`, `escalate`, `emergency`, `spawn`, `resolve`, `decision`, `sync`, `checkpoint`, `warning`
- `detail` — human-readable description (one line, <200 chars)

**Retention:** Rolling 500 entries max. Truncate oldest on overflow. On agent shutdown, the daemon preserves the current `session.jsonl` and starts a fresh one on next startup.

#### `context.json` — Machine-Readable State

Agent-specific structured state that enables fast startup without reprocessing. Each agent defines its own schema (see Agent-Specific State Requirements below).

**Atomic writes:** Always write to `context.json.tmp`, then rename to `context.json`. Prevents corruption on crash.

**Version field:** Every `context.json` includes a `version` integer. Schema changes bump the version; agents that encounter an unknown version start fresh rather than misinterpreting data.

### Agent-Specific State Requirements (Task 2)

#### merge-queue

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "tracked_prs": [
    {
      "number": 850,
      "title": "feat: session handoff (Story 73.4)",
      "status": "ready",
      "ci": "passing",
      "last_checked": "2026-03-29T14:25:00Z"
    }
  ],
  "emergency_mode": false,
  "emergency_since": null,
  "last_merge": {
    "pr": 849,
    "timestamp": "2026-03-29T14:00:00Z",
    "ci_result": "success"
  },
  "processed_prs": [845, 846, 847, 848, 849],
  "stale_branches_cleaned": ["work/old-worker-1"]
}
```

**What to persist:** Tracked PR list with validation state, emergency mode flag, post-merge CI status, processed PR correlation IDs (last 50), recently cleaned branches.

**What NOT to persist:** Full PR diffs, review contents, CI logs — these are available from GitHub on demand.

#### pr-shepherd

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "conflict_queue": [
    {
      "pr": 852,
      "branch": "work/fancy-cat",
      "conflict_detected": "2026-03-29T13:00:00Z",
      "resolution_attempted": false
    }
  ],
  "active_worktrees": [],
  "stale_prs": [
    {"pr": 840, "stale_since": "2026-03-22T00:00:00Z", "labeled": true}
  ],
  "spawned_workers": [
    {"pr": 851, "task": "Fix CI", "spawned_at": "2026-03-29T14:50:00Z"}
  ]
}
```

**What to persist:** Conflict resolution queue, active worktree tracking, stale PR list, spawned worker tracking.

#### envoy

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "triage_state": [
    {
      "issue": 95,
      "stage": "PM examination",
      "labels_applied": ["triage.in-progress", "type.feature", "priority.p1"],
      "escalated_to": "supervisor",
      "escalated_at": "2026-03-29T13:00:00Z"
    }
  ],
  "recently_screened": [90, 91, 92, 93, 94, 95],
  "cross_check_last_merged_pr": 849,
  "pending_reporter_updates": [
    {"issue": 93, "update": "Fix merged in PR #848, awaiting verification"}
  ]
}
```

**What to persist:** Issue triage progress (stage, labels applied, escalation state), recently screened issue numbers, cross-check watermark, pending reporter updates.

#### project-watchdog

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "processed_prs": [845, 846, 847, 848, 849],
  "pending_story_updates": [
    {"story": "73.4", "new_status": "Done (PR #860)", "pr": 860}
  ],
  "allocated_numbers": {
    "last_epic": 73,
    "pending_allocations": []
  },
  "recommendation_queue_cursor": "REC-045"
}
```

**What to persist:** Processed PR correlation IDs (last 50), pending story status updates, number allocation state (last allocated epic/story), recommendation queue cursor.

#### arch-watchdog

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "processed_prs": [845, 846, 847, 848, 849],
  "pending_reviews": [
    {
      "pr": 850,
      "files": ["internal/tasks/new_package/"],
      "concern": "New package introduced without architecture doc"
    }
  ],
  "flagged_patterns": [
    {"pattern": "direct ANSI escapes in internal/tui/", "first_seen": "2026-03-28T10:00:00Z", "issue": null}
  ]
}
```

**What to persist:** Processed PR correlation IDs (last 50), pending architecture review queue, flagged pattern tracking.

#### retrospector

```json
{
  "version": 1,
  "last_updated": "2026-03-29T14:30:00Z",
  "last_pr": 849,
  "mode_rotation_index": 2,
  "rolling_windows": {
    "ci_failure_rate_10pr": 0.1,
    "conflict_rate_10pr": 0.2,
    "rebase_avg_10pr": 1.3
  },
  "prs_since_restart": 0,
  "hours_since_restart": 0,
  "recommendation_cursor": "REC-045",
  "messaging_fallback": false
}
```

**Migration path:** Retrospector already uses `docs/operations/retrospector-checkpoint.json` for analytical state. The `context.json` here supersedes that file, adding clean separation of session-scoped counters (always reset to 0 on startup) from analytical state (persisted). See the research at `_bmad-output/planning-artifacts/retrospector-persistence-research.md` for the detailed analysis of the stale counter bug this fixes.

### Daemon Integration Design (AC5)

#### Graceful Shutdown — "Prepare for Shutdown" Signal

Before killing an agent, the daemon sends a structured message:

```bash
multiclaude message send <agent-name> "SESSION_HANDOFF_PREPARE"
```

**Agent response protocol:**
1. Agent receives `SESSION_HANDOFF_PREPARE` message
2. Agent writes `handoff.md` with current state (In Progress, Recently Completed, Blocked, Decisions, Warnings)
3. Agent writes final `context.json` checkpoint
4. Agent flushes any pending `session.jsonl` entries
5. Agent responds: `multiclaude message send supervisor "SESSION_HANDOFF_READY"`
6. Daemon waits up to 30 seconds for `SESSION_HANDOFF_READY`, then terminates

**Timeout behavior:** If the agent doesn't respond within 30 seconds (frozen, context exhausted, crashed), the daemon terminates it anyway. The previous `context.json` and `session.jsonl` entries written during normal operation provide partial recovery — the handoff note is the only thing lost.

#### Startup — Handoff Injection (AC4)

When spawning an agent, the daemon:

1. Checks for `~/.multiclaude/agent-state/<repo>/<agent-name>/handoff.md`
2. If it exists, prepends its contents to the agent's startup prompt:
   ```
   --- SESSION HANDOFF FROM PREVIOUS SESSION ---
   <contents of handoff.md>
   --- END SESSION HANDOFF ---

   <normal agent definition prompt>
   ```
3. Also makes `context.json` available by including a note:
   ```
   Your previous session state is available at:
   ~/.multiclaude/agent-state/<repo>/<agent-name>/context.json
   Read this file on startup to restore tracked PRs, correlation IDs, and other state.
   ```

**Why prepend, not append:** The handoff context is most important for the agent's first actions. Prepending ensures it's in the agent's attention window before it starts its polling loop.

#### Breadcrumb Logging

Agents write to `session.jsonl` autonomously during operation. No daemon involvement needed — the agent simply appends entries as significant actions occur. The daemon's only role is:
- Preserving `session.jsonl` across restarts (it's outside the git repo, so no risk of loss from git operations)
- Rotating old session logs if they exceed the 500-entry limit

#### State File Lifecycle

| Event | handoff.md | session.jsonl | context.json |
|-------|-----------|--------------|-------------|
| Agent startup | Read and inject into prompt, then archive to `handoff-<ts>.md` | Start fresh file | Read for state restoration |
| During operation | Not touched | Agent appends breadcrumbs | Agent writes periodically (every 5th action or 2 hours) |
| Graceful shutdown | Agent writes final version | Agent flushes pending entries | Agent writes final checkpoint |
| Crash (no graceful shutdown) | Stale from previous session — still useful | Contains entries up to crash point | Contains last periodic checkpoint |
| Rotation | Keep last 3 archived handoffs | Keep current + 1 previous | Keep current only (versioned schema handles migration) |

### Breadcrumb Action Categories

Standardized action types for `session.jsonl` entries:

| Action | Used By | Description |
|--------|---------|-------------|
| `merge` | merge-queue | PR merged successfully |
| `merge_blocked` | merge-queue | PR blocked from merge (scope, CI, review) |
| `emergency` | merge-queue | Emergency mode entered/exited |
| `rebase` | pr-shepherd | Branch rebased to resolve conflicts |
| `conflict` | pr-shepherd | Merge conflict detected on a PR |
| `triage` | envoy | Issue screened and classified |
| `escalate` | envoy, all | Issue or decision escalated to supervisor |
| `cross_check` | envoy | Merged PR cross-checked against open issues |
| `sync` | project-watchdog | Story status or planning doc synced |
| `allocate` | project-watchdog | Epic/story number allocated |
| `drift` | arch-watchdog | Architecture drift detected |
| `doc_update` | arch-watchdog | Architecture doc updated |
| `finding` | retrospector | New finding recorded to JSONL |
| `recommendation` | retrospector | Recommendation filed to queue |
| `saga` | retrospector | Saga condition detected |
| `spawn` | any | Worker spawned for a task |
| `decision` | any | Significant decision made during session |
| `checkpoint` | any | State checkpoint written |
| `warning` | any | Operational warning logged |

## Prototype: Merge-Queue Handoff (Task 6)

Until daemon integration is built, the handoff protocol can be validated manually:

### Manual Handoff Procedure

**Before restarting merge-queue:**
1. Send `SESSION_HANDOFF_PREPARE` message:
   ```bash
   multiclaude message send merge-queue "SESSION_HANDOFF_PREPARE"
   ```
2. Wait for merge-queue to write its handoff files
3. Kill the agent tmux window
4. Respawn with the handoff note included in the prompt

**On respawn, the supervisor (or operator) can inject handoff context:**
```bash
# Read the handoff note
cat ~/.multiclaude/agent-state/ThreeDoors/merge-queue/handoff.md

# Include it in the spawn task
multiclaude agents spawn --name merge-queue --class persistent \
  --prompt-file agents/merge-queue.md \
  --task "Previous session handoff: $(cat ~/.multiclaude/agent-state/ThreeDoors/merge-queue/handoff.md)"
```

**Note:** The `--task` flag appends to the prompt. If multiclaude doesn't support this flag yet, the daemon integration (separate PR) will add proper handoff injection.

### Validation Criteria

The prototype is successful if:
1. merge-queue writes a handoff note that captures its current PR tracking state
2. After restart with the handoff note injected, merge-queue doesn't re-process already-merged PRs
3. Emergency mode state survives restart (if active at shutdown time)

## Relationship to Existing Systems

### Epic 58 — Supervisor Shift Handover

Epic 58 designed `shift-state.yaml` for supervisor-to-supervisor handover. That schema captures the full system view (workers, persistent agents, open PRs, pending decisions). Session handoff is complementary:
- **shift-state.yaml**: System-wide snapshot maintained by the daemon, used for supervisor rotation
- **agent handoff**: Per-agent state maintained by each agent, used for individual agent restarts

The two systems don't overlap — shift-state.yaml doesn't capture per-agent internal state (correlation IDs, triage progress, etc.), and agent handoff doesn't capture the full system view.

### Retrospector Checkpoint

Retrospector's existing `docs/operations/retrospector-checkpoint.json` is a precursor to the `context.json` pattern. The migration path:
1. Continue using `retrospector-checkpoint.json` for now (it's committed to git, which has value for the data pipeline)
2. Add `context.json` for session-scoped state that should NOT be in git (correlation IDs, session counters)
3. Eventually, the daemon manages persistence for both — `context.json` becomes the single source, and the daemon handles syncing analytical state to `docs/operations/` for the data pipeline

## Rejected Approaches

| Approach | Why Rejected |
|----------|-------------|
| Store handoff state in git repo | Ephemeral per-machine state pollutes repo; merge conflicts between agents |
| Shared state file for all agents | Agent state schemas are too different; single file creates contention |
| Database (SQLite) for agent state | Over-engineered for the current scale; files are simpler and human-readable |
| Eager handoff (write on every action) | Too much I/O; periodic checkpointing (every 5th action or 2 hours) is sufficient |
| Agent self-manages state without daemon | Agents can't inject state into their own startup prompt; daemon must mediate |
