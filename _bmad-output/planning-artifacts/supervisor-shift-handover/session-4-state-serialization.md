# Party Mode Session 4: State Serialization

**Date:** 2026-03-11
**Topic:** Supervisor Shift Handover — What Does the Incoming Supervisor Need to Know?
**Participants:** Winston (Architect), John (PM), Murat (TEA), Amelia (Dev), Paige (Tech Writer)

---

## Problem Statement

When a supervisor hands off, the incoming instance needs enough context to continue seamlessly. What state must be serialized, what can be discovered, and what format should it take?

## Three Categories of State

### 1. Observable State (Daemon-Maintained)
Things the daemon can discover externally without supervisor cooperation:
- Active tmux sessions / worker names
- Worker branches (from git)
- Open PRs (from `gh pr list`)
- Pending messages (from `multiclaude message list`)
- Persistent agent status (from tmux session list)

**Updated every 5 minutes by daemon refresh loop.**

### 2. Conversational State (Supervisor Context Only)
What the supervisor has discussed with workers, context from escalations, nuance about blockers. Lives only in the supervisor's context window.

**This is the hardest to capture.** Example: worker asks "should I interpret AC #3 as X or Y?" and supervisor says "X" — that decision exists only in conversation.

### 3. Decision State (Partially External)
Pending decisions, priorities, judgment calls. Partially in context, partially in MEMORY.md.

## Key Design Decision: Externalize Decisions Immediately

**Adopted approach:** Every time the supervisor makes a decision that a future supervisor would need to know, write it down immediately — not at handover time.

**Rationale:** Converts the hard problem (serialize conversational state at handover) into a discipline problem (write things down as they happen). A degraded supervisor trying to dump everything at the end is unreliable.

**Risk (TEA):** Write fatigue — if the supervisor writes to MEMORY.md after every decision, the writes consume context and time. Must be selective.

## What MUST Be Serialized

| Category | Examples | Why |
|---|---|---|
| Active worker tasks + status | "bold-eagle implementing story 42.3, last said tests 3/5 passing" | Incoming needs to know what's in flight |
| Pending decisions | "Worker asked about AC #3 scope — told them per-session" | Decisions not yet acted on may need follow-up |
| Blocking dependencies | "Story 42.4 depends on 42.3 — don't dispatch until PR merges" | Prevents premature dispatch |
| Sprint priorities (ordered) | "Epic 42 completion is sprint goal" | Guides incoming supervisor's decision-making |
| Issue triage state | "Issue #89 in PM examination phase" | Prevents re-triaging or dropping in-progress triage |
| Operational warnings | "merge-queue can't merge workflow PRs" | Known limitations that affect operations |

## What Does NOT Need Serialization

| Category | Why Not |
|---|---|
| Worker implementation details | Worker has its own context |
| Routine status updates | Message log has these |
| Completed/acknowledged work | Done; git has the record |
| Persistent agent configuration | They have their own prompts |
| Codebase structure | CLAUDE.md and reading the code covers this |

## Complete State File Schema

```yaml
# shift-state.yaml — Complete handover state
version: 1
timestamp: "2026-03-11T14:30:00Z"
outgoing_supervisor: "gentle-hawk"

# Section 1: Observable state (daemon-maintained)
workers:
  active:
    - name: "bold-eagle"
      task: "Implement story 42.3"
      story_file: "docs/stories/42.3.story.md"
      branch: "work/bold-eagle"
      pr: null
      dispatched_at: "2026-03-11T13:45:00Z"
    - name: "swift-fox"
      task: "Fix CI lint failures on PR #567"
      branch: "work/swift-fox"
      pr: "#567"
      dispatched_at: "2026-03-11T14:10:00Z"
  recently_completed:
    - name: "calm-deer"
      task: "Implement story 42.2"
      pr: "#565"
      completed_at: "2026-03-11T14:20:00Z"
      result: "PR created, CI passing"

persistent_agents:
  - name: "merge-queue"
    status: "active"
    notes: "Processing #565, #566 in queue"
  - name: "pr-shepherd"
    status: "active"
    notes: "Rebasing #564 onto latest main"
  - name: "project-watchdog"
    status: "active"
  - name: "arch-watchdog"
    status: "active"
  - name: "envoy"
    status: "active"

open_prs:
  - number: 564
    title: "feat: keybinding auto-fade (Story 39.12)"
    status: "needs-rebase"
    ci: "pending"
  - number: 565
    title: "feat: task pool analytics (Story 42.2)"
    status: "ready"
    ci: "passing"

# Section 2: Supervisor-unique context (written by outgoing supervisor)
pending_decisions:
  - context: "Worker bold-eagle asked whether AC #3 means per-session or global analytics"
    recommendation: "Told worker to implement per-session for now, can extend later"
    resolved: false

priorities:
  - "Epic 42 completion is the sprint goal — stories 42.2 and 42.3 are critical path"
  - "PR #564 (story 39.12) is stale — needs attention today"
  - "Issue #89 triage is in PM examination phase — follow up with envoy"

issue_triage:
  - issue: "#89"
    stage: "PM examination"
    assigned_to: "envoy"
    notes: "User reported crash on empty task file. Envoy acknowledged."

blockers:
  - "Story 42.4 depends on 42.3 — don't dispatch until bold-eagle's PR merges"

# Section 3: Operational notes
warnings:
  - "merge-queue can't merge workflow PRs (OAuth scope limitation)"
  - "pr-shepherd definition was updated yesterday — verify it loaded correctly"
```

## Design Principles

1. **Schema versioned** — Include version number for future evolution
2. **Self-explanatory fields** — Readable without external context
3. **Absolute timestamps only** — Never "2 hours ago"
4. **Include paths/links** — So incoming supervisor can verify
5. **Priorities limited to 3-5 items** — Ordered by urgency
6. **Write-once per shift** — After handover, becomes an audit trail
7. **Archived to history** — `~/.multiclaude/handover/ThreeDoors/history/` with timestamps for debugging and process improvement

## Two-Phase Write Protocol

1. **Daemon writes base snapshot** — Observable state from external sources (worker list, PR list, message list, tmux sessions). Updated every 5 minutes.
2. **Supervisor writes delta** — Only supervisor-unique context: pending decisions, priorities, blockers, warnings. Minimal cognitive load on potentially degraded instance.

The daemon asks: "Anything to add that only you know?" The supervisor appends its delta and exits.
