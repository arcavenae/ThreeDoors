# Shift Handover State File Schema

> Reference for the `shift-state.yaml` file used during supervisor shift handover.
> See [synthesis](../../_bmad-output/planning-artifacts/supervisor-shift-handover/synthesis-supervisor-shift-handover.md) for design rationale.

## Location

```
~/.multiclaude/handover/<repo>/shift-state.yaml
```

Archived after handover to: `~/.multiclaude/handover/<repo>/history/<timestamp>-shift-state.yaml`

## Schema (v1)

```yaml
# Required metadata
version: 1                              # Schema version (integer)
timestamp: "2026-03-11T14:30:00Z"       # ISO 8601 UTC when file was written
outgoing_supervisor: "gentle-hawk"       # Name of the outgoing supervisor agent

# Section 1: Observable State (daemon-maintained)
# Updated every 5 minutes by the daemon refresh loop from external sources.

workers:
  active:                                # Currently running workers
    - name: "bold-eagle"                 # Worker agent name
      task: "Implement story 42.3"       # Human-readable task description
      story_file: "docs/stories/42.3.story.md"  # Path to story file (if story work)
      branch: "work/bold-eagle"          # Git branch
      pr: null                           # PR number (null if not yet created)
      dispatched_at: "2026-03-11T13:45:00Z"  # ISO 8601 UTC
  recently_completed:                    # Workers that finished in the last refresh cycle
    - name: "calm-deer"
      task: "Implement story 42.2"
      pr: "#565"
      completed_at: "2026-03-11T14:20:00Z"
      result: "PR created, CI passing"   # Brief outcome

persistent_agents:                       # Long-running agent status
  - name: "merge-queue"
    status: "active"                     # "active" or "inactive"
    notes: "Processing #565, #566 in queue"  # Optional context
  - name: "pr-shepherd"
    status: "active"
    notes: "Rebasing #564 onto latest main"
  - name: "project-watchdog"
    status: "active"
  - name: "arch-watchdog"
    status: "active"
  - name: "envoy"
    status: "active"

open_prs:                                # Open PRs from gh pr list
  - number: 564
    title: "feat: keybinding auto-fade (Story 39.12)"
    status: "needs-rebase"               # "ready", "needs-rebase", "changes-requested", "draft"
    ci: "pending"                        # "passing", "failing", "pending"
  - number: 565
    title: "feat: task pool analytics (Story 42.2)"
    status: "ready"
    ci: "passing"

# Section 2: Supervisor Delta (written by outgoing supervisor)
# Contains context only the outgoing supervisor had. May be empty in emergency handover.

pending_decisions:                       # Unresolved decisions needing follow-up
  - context: "Worker bold-eagle asked whether AC #3 means per-session or global analytics"
    recommendation: "Told worker to implement per-session for now, can extend later"
    resolved: false                      # true if decided but not yet acted on

priorities:                              # Ordered by urgency, max 5 items
  - "Epic 42 completion is the sprint goal — stories 42.2 and 42.3 are critical path"
  - "PR #564 (story 39.12) is stale — needs attention today"
  - "Issue #89 triage is in PM examination phase — follow up with envoy"

issue_triage:                            # In-progress issue triage state
  - issue: "#89"
    stage: "PM examination"              # "acknowledged", "PM examination", "party mode", "story creation", "complete"
    assigned_to: "envoy"
    notes: "User reported crash on empty task file. Envoy acknowledged."

blockers:                                # Dependency blocks — incoming must respect these
  - "Story 42.4 depends on 42.3 — don't dispatch until bold-eagle's PR merges"

# Section 3: Operational Notes
# Known limitations, warnings, and gotchas.

warnings:                                # Operational warnings
  - "merge-queue can't merge workflow PRs (OAuth scope limitation)"
  - "pr-shepherd definition was updated yesterday — verify it loaded correctly"
```

## Field Requirements

| Field | Required | Maintained By | Notes |
|-------|----------|---------------|-------|
| `version` | Yes | Daemon | Always `1` for v1 schema |
| `timestamp` | Yes | Daemon | ISO 8601 UTC, no relative times |
| `outgoing_supervisor` | Yes | Daemon | Agent name of outgoing instance |
| `workers.active` | Yes | Daemon | May be empty list `[]` |
| `workers.recently_completed` | No | Daemon | Omit if none |
| `persistent_agents` | Yes | Daemon | All 5 expected agents listed |
| `open_prs` | Yes | Daemon | From `gh pr list` |
| `pending_decisions` | No | Supervisor | Empty in emergency handover |
| `priorities` | No | Supervisor | Empty in emergency handover |
| `issue_triage` | No | Supervisor | Empty in emergency handover |
| `blockers` | No | Supervisor | Empty in emergency handover |
| `warnings` | No | Either | Daemon adds known limitations; supervisor adds discovered ones |

## Two-Phase Write Protocol

1. **Daemon writes base snapshot** — Observable state from external sources (`multiclaude worker list`, `gh pr list`, `multiclaude message list`, `tmux list-windows`). Updated every 5 minutes as part of the refresh loop.

2. **Supervisor writes delta** — Only supervisor-unique context: pending decisions, priorities, blockers, warnings. The daemon prompts with "Anything to add that only you know?" and the supervisor appends its delta.

In emergency handover (outgoing unresponsive), only the daemon snapshot exists. The incoming supervisor operates with reduced context but is not blocked.

## Constraints

- **Maximum file size:** 10KB. Alert if exceeded.
- **Priorities:** Maximum 5 items, ordered by urgency.
- **Timestamps:** Always absolute UTC. Never relative ("2 hours ago").
- **Paths/links:** Include full paths so incoming supervisor can verify.
- **Write-once per shift:** After handover, the file becomes an audit trail (archived to history).
