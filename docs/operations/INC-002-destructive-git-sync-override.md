# INC-002: Destructive Git Sync Override in Worker Dispatch

**Date:** 2026-03-10
**Severity:** Medium
**Duration:** Ongoing since ~2026-03-07 (3 days), discovered 2026-03-10
**Status:** Resolved (PR #424)

---

## Summary

The supervisor's persistent memory contained a mandatory instruction to prepend `git fetch origin main && git rebase origin/main` to every worker task description. This overrode multiclaude's built-in worktree management, which already handles git sync automatically. The instruction was redundant in most cases and destructive in at least one confirmed case (worker stuck mid-rebase with conflicts, branch detached to bare HEAD).

---

## What Happened

### The Bad Instruction

Supervisor MEMORY.md contained this in the "Worker Dispatch Checklist":

> **Git sync FIRST** — Every worker task description MUST start with: `FIRST: run 'git fetch origin main && git rebase origin/main' before doing anything else.` Workers skip this if it's buried at the end. Put it at the top.

This was applied to **every worker dispatch** — visible in 100+ worker completion messages across 3 days. The instruction was also reinforced by Standing Order #2: "All agents must sync git before starting work."

### What multiclaude Actually Does

Research into multiclaude's source code (daemon.go, worktree.go) revealed the instruction was entirely unnecessary:

1. **At worker creation:** multiclaude runs `git worktree add -b work/{name} {path} HEAD`, creating a fresh isolated worktree from the current HEAD of the main repo (which the daemon keeps updated).

2. **Every 5 minutes:** The daemon's refresh loop (`RefreshWorktree()`) automatically fetches from remote, stashes uncommitted changes, rebases the worktree onto `origin/main`, and restores the stash.

Workers never needed to sync git manually. The daemon handles it.

### Confirmed Damage

Worker `witty-hawk` (dispatched 2026-03-10) executed the `git rebase origin/main` instruction and got stuck mid-rebase with merge conflicts. Its branch was left in a detached HEAD state. The worktree had to be manually cleaned up (`git branch -D`, tmux window killed).

### Probable Damage

Over 3 days and 100+ worker dispatches, an unknown number of workers likely experienced:
- Wasted time executing a no-op rebase (best case)
- Transient rebase conflicts that the worker had to resolve before starting real work (moderate case)
- Failed rebases requiring manual intervention or worker respawn (worst case)

---

## Root Cause

The instruction was cargo-culted from an earlier era of the project. Possible origins:

1. **Fork-mode era (pre 2026-03-07):** When ThreeDoors used a fork workflow, workers may have needed to sync with upstream. The switch to direct-push mode eliminated this need, but the instruction was never re-evaluated.

2. **Pre-worktree era:** If multiclaude previously created workers without worktrees (or with a simpler branching strategy), manual sync may have been necessary. The instruction persisted after multiclaude matured.

3. **Reinforcement loop:** The instruction was encoded as a "MUST" rule in supervisor memory, applied without question, and never failed loudly enough to trigger re-evaluation. Workers that experienced issues likely just retried or worked around it silently.

---

## Resolution

### Immediate Fixes (2026-03-10)

| Location | Change |
|----------|--------|
| Supervisor MEMORY.md | Deleted "Git sync FIRST" from Worker Dispatch Checklist. Added explicit warning: "NEVER prepend git fetch/rebase to worker tasks." Documented correct worktree model. |
| MEMORY.md Standing Order #2 | Changed from "All agents must sync git" to distinguishing workers (auto-managed) from persistent agents (may sync). |
| `agents/supervisor.md` | Standing Order #2 updated. Worker Dispatch Checklist fixed. (PR #424) |
| `agents/worker.md` | Added "Git Worktree (Managed by multiclaude)" section. (PR #424) |

### Pending

- `/sync-enhancements` to propagate fix to `multiclaude-enhancements` repo
- PR #419 (Epic 42 planning) still has unrelated merge conflicts

---

## Lessons Learned

### 1. Validate infrastructure assumptions against actual tool behavior

The supervisor assumed workers needed manual git sync without ever verifying how multiclaude handles worktrees. A 5-minute read of the daemon's behavior would have prevented 3 days of unnecessary overhead. **Action:** When adopting or relying on a tool's behavior, read the source or docs — don't guess and encode the guess as policy.

### 2. "MUST" rules in agent memory calcify into unquestioned dogma

Once encoded as item #0 in the Worker Dispatch Checklist with the word "MUST," the instruction was never re-evaluated. It was applied 100+ times without anyone asking "is this still necessary?" **Action:** Periodically audit MEMORY.md for stale or cargo-culted rules. Every "MUST" should have a rationale that can be re-validated.

### 3. Don't override the platform — extend it

multiclaude's worktree management is a well-designed abstraction. By overriding it with manual git commands, we broke the abstraction, created redundant work, and introduced failure modes that the platform was specifically designed to prevent. **Action:** When a platform provides a capability, use it. Only override with explicit justification and documentation of why the platform's behavior is insufficient.

### 4. Silent failures hide systemic problems

Workers that hit rebase conflicts likely just worked around them or failed and were respawned. The cost was invisible — spread across many workers as small delays rather than concentrated in one visible failure. **Action:** Log and surface friction. If workers are doing unexpected git operations, that should be visible in their output or metrics.

### 5. Incident patterns repeat — check for echoes

INC-001 (pr-shepherd contamination) was about an agent doing git operations in the wrong context. INC-002 is the same category — agents doing git operations they shouldn't. Both stem from insufficient understanding of multiclaude's git model. **Action:** After an incident, check for the same anti-pattern in other parts of the system. INC-001 fixed pr-shepherd but didn't audit whether the same confusion existed in worker dispatch.

---

## Related

- [INC-001: pr-shepherd Contamination of Shared Checkout](INC-001-pr-shepherd-contamination.md)
- PR #424: fix: remove destructive git sync overrides from agent definitions
- multiclaude source: `daemon.go:1600-1612` (worktree creation), `worktree.go:742-885` (refresh loop)
