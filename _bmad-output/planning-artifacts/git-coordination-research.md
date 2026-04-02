# Git Coordination Research: Keeping Working Copies Current

**Date:** 2026-04-02
**Researcher:** cool-platypus (worker)
**Status:** Complete

---

## Executive Summary

ThreeDoors has **two distinct staleness problems** that require different solutions:

1. **Persistent agents** (merge-queue, pr-shepherd, etc.) share the main checkout at `~/.multiclaude/repos/ThreeDoors/` — this checkout gets perpetually stuck behind `origin/main` because nobody refreshes it
2. **Workspace/director** is a worktree on branch `workspace/default` — also never refreshed by the daemon

The upstream multiclaude daemon has a worktree refresh loop, but it **only refreshes worker worktrees** (`AgentTypeWorker`). Persistent agents, workspace, and supervisor are explicitly excluded. This is by design — but the design has a gap.

---

## Part 1: How Upstream Multiclaude Works

### Daemon Refresh Loop

Source: `/Users/skippy/work/multiclaude/internal/daemon/daemon.go:490-607`

- Runs every **5 minutes** via `time.Ticker`
- First run after 30-second startup delay
- Can be triggered manually via `multiclaude refresh` CLI command

**Algorithm per repo:**
1. Fetch from remote (once per repo, not per worktree)
2. Iterate agents — **skip all non-worker types** (line 555-556: `if agent.Type != state.AgentTypeWorker { continue }`)
3. Check worktree state via `GetWorktreeState()` — skip if detached HEAD, mid-rebase, mid-merge, on main, or already up-to-date
4. Call `RefreshWorktree()` which: stashes uncommitted changes → rebases onto `remote/main` → pops stash
5. On conflict: abort rebase, restore stash, log warning
6. On success: notify the agent via inter-agent messaging

### Agent Types and Their Checkout Model

Source: `/Users/skippy/work/multiclaude/internal/daemon/daemon.go:1602-1612`

| Agent Type | Checkout Location | Refreshed by Daemon? |
|---|---|---|
| `supervisor` | Main repo dir | No |
| `workspace` | Dedicated worktree on `workspace/default` | No |
| `worker` | Dedicated worktree on `work/<name>` | **Yes** |
| `merge-queue` | Main repo dir (persistent) | No |
| `pr-shepherd` | Main repo dir (persistent) | No |
| `generic-persistent` | Main repo dir (persistent) | No |
| `review` | Ephemeral worktree | No |

**Key insight:** Persistent agents (`agentClass == "persistent"`) are explicitly given `worktreePath = repoPath` (line 1604-1605). They all share the same directory. Workers get isolated worktrees with their own branches.

### RefreshWorktree Implementation

Source: `/Users/skippy/work/multiclaude/internal/worktree/worktree.go:742-885`

The `RefreshWorktree()` function is robust:
- Checks for mid-rebase, mid-merge, detached HEAD before attempting anything
- Stashes uncommitted changes (including untracked) before rebase
- Aborts rebase on conflict and restores stash
- Returns detailed `RefreshResult` with conflict file list

**Critical limitation:** This function uses `git rebase`, which rewrites history. For shared checkouts (where multiple agents operate), a rebase while another agent is mid-operation could cause data loss — exactly the INC-002 scenario.

---

## Part 2: How Our System Works Today

### ADR-0030: CI Churn Reduction

- Intentionally disabled "require branches to be up-to-date" for PRs
- merge-queue does NOT rebase PRs before merging
- pr-shepherd ONLY rebases when there are actual merge conflicts
- Post-merge CI on main is the safety net

### Git Safety Hook Scope

Source: `scripts/hooks/git-safety.sh`

The hook detects worker worktrees via path matching (`/.multiclaude/wts/`):
- **Workers:** BLOCKED from `git fetch`, `git pull`, `git rebase`, `git merge`
- **Persistent agents & workspace:** NOT blocked — they CAN run these commands
- **Universal blocks:** unsigned commits, push to main, Co-Authored-By

### Current Agent Behavior

**merge-queue** (line 68 in `agents/merge-queue.md`):
- CAN "Sync local main after merges (`git fetch origin main:main`)"
- Uses `gh api` for branch updates, not direct git operations

**pr-shepherd** (line 51, 76-84 in `agents/pr-shepherd.md`):
- CAN "Keep local main in sync with remote"
- Uses temporary worktrees for ALL rebases (`git worktree add /tmp/pr-rebase-NNN`)
- Explicitly NEVER operates in the shared checkout (INC-001 guardrail)

**Neither agent actively refreshes the shared checkout.** merge-queue's authority includes `git fetch origin main:main` but this only updates the local ref, not the working tree.

---

## Part 3: External Approaches

### Gastown (gastownhall/gastown) — Integration Branches

Gastown's most relevant innovation is **integration branches** for epic-scoped work:

- Workers ("polecats") spawn worktrees from an integration branch, not main
- The "Refinery" (merge queue) merges completed MRs into the integration branch
- When all epic children are closed, the integration branch lands on main as a single merge commit
- Polecats starting later get sibling work already present on the integration branch

**Key coordination patterns:**
1. **Session-per-step model:** Each molecule step gets one polecat session. The sandbox (branch, worktree) persists across sessions. Sessions cycle; sandboxes persist.
2. **No shared checkouts:** Every agent (polecat) operates in its own worktree/sandbox. The Refinery uses temporary worktrees for merges. The "Witness" monitors health.
3. **Three-layer safety:** Formula/role instructions (soft) + pre-push hook (hard) + authorized code path (hard) prevent unauthorized integration branch merges.
4. **Auto-detection:** Workers auto-detect their target branch based on epic hierarchy.

**Applicability to ThreeDoors:**
- Integration branches are relevant for grouped epic work but don't solve the persistent-agent staleness problem
- The "no shared checkout" principle is the key lesson — Gastown avoids our problem entirely by never having multiple agents share a working directory
- The temporary worktree pattern for merge operations (like pr-shepherd already uses) is validated by Gastown

### OpenClaw (openclaw/openclaw) — Single-Agent Architecture

OpenClaw is a single-agent CLI tool, not a multi-agent system. It uses skills/plugins for extensibility but doesn't have multi-agent git coordination. Not directly applicable.

---

## Part 4: Proposed Solutions

### Problem A: Persistent Agent Checkout Stale

**Root cause:** All persistent agents share `~/.multiclaude/repos/ThreeDoors/`. Nobody runs `git pull` or equivalent. The daemon refresh loop skips non-worker agents.

**Proposed solution: Ref-only fetch + message notification**

**WHO:** The daemon refresh loop (upstream multiclaude change) OR a dedicated persistent agent

**WHEN:** Every 5 minutes (piggyback on existing refresh loop)

**HOW:**
```bash
# Safe: updates refs without touching working tree
git fetch origin main:main
```

This is the critical insight: `git fetch origin main:main` updates the local `main` ref to match `origin/main` WITHOUT touching the working tree or any checked-out branch. The persistent agents don't need their working tree updated because:
- merge-queue uses `gh api` and `gh pr` commands (GitHub API, not local git)
- pr-shepherd uses temporary worktrees for all operations
- Other persistent agents primarily read files, not build/run code

**What persistent agents actually need:**
1. Fresh `origin/main` ref for `git log`, `git diff`, and branch comparisons
2. Up-to-date remote tracking refs for `gh pr` operations (gh handles this itself)
3. Local file contents at HEAD for reading CLAUDE.md, agent definitions, etc.

For #3 (local file staleness), persistent agents would need a working tree update. Options:

**Option A — Do nothing for working tree (recommended for now):**
Persistent agents read CLAUDE.md and agent definitions at spawn time. Context is baked in. They don't dynamically re-read files. `git fetch origin main:main` is sufficient.

**Option B — Periodic safe reset of shared checkout:**
```bash
# Only safe if no agent has uncommitted changes (persistent agents shouldn't)
git checkout main
git reset --hard origin/main
```
**Risk:** INC-001 and INC-002 scenarios. If ANY persistent agent is mid-operation, this destroys state. NOT recommended unless we can coordinate agent quiescence.

**Option C — Give each persistent agent its own worktree (Gastown model):**
Upstream multiclaude would need to change `persistent` agent spawning to create isolated worktrees like workers. This is the most robust long-term solution but requires upstream changes.

### Problem B: Workspace/Director Stale

**Root cause:** Workspace is a worktree on `workspace/default` branch. The daemon doesn't refresh it (type is `AgentTypeWorkspace`, not `AgentTypeWorker`).

**Proposed solution: On-demand refresh via `/refresh` slash command**

**WHO:** The human operator (or future director agent)

**WHEN:** On demand, before starting new work

**HOW:** The `/refresh` slash command already exists and works:
```bash
git fetch origin main
git rebase origin/main
```

This is safe for the workspace because:
- It's an isolated worktree (not shared)
- The operator controls timing (no mid-operation conflicts)
- If conflicts occur, the operator can resolve them

**For the future director agent:**
The director should periodically self-refresh, similar to how workers are refreshed. Two approaches:

1. **Extend daemon refresh to include workspace type:**
   - Modify upstream multiclaude to also refresh `AgentTypeWorkspace` worktrees
   - Use the same stash/rebase/restore pattern
   - Advantage: consistent with worker refresh, daemon-managed

2. **Director self-refresh on HEARTBEAT:**
   - Director's polling loop includes a self-refresh step
   - Uses the `/refresh` slash command logic
   - Advantage: no upstream changes needed, agent-controlled timing

**Recommended: Option 1 (daemon-managed) with Option 2 as fallback.**

### Problem C: PR Branch Updates Before Merge

**Current state:** ADR-0030 explicitly chose NOT to require PRs to be up-to-date. The push-to-main CI is the safety net.

**WHO:** pr-shepherd (for conflict resolution) or merge-queue (for pre-merge update)

**WHEN:**
- pr-shepherd: When `gh pr view --json mergeable` shows `CONFLICTING`
- merge-queue: Optionally, just before merge if we want the safety of a rebase

**HOW:** pr-shepherd already does this correctly via temporary worktrees:
```bash
git worktree add /tmp/pr-rebase-NNN <branch-name>
cd /tmp/pr-rebase-NNN
git fetch origin main
git rebase origin/main
git push --force-with-lease origin <branch-name>
cd -
git worktree remove /tmp/pr-rebase-NNN
```

**One-at-a-time coordination:** merge-queue should process PRs sequentially:
1. Check PR is mergeable (no conflicts)
2. If conflicts: message pr-shepherd, wait
3. If clean: merge via `gh pr merge`
4. Wait for push-to-main CI
5. If CI red: emergency mode
6. If CI green: next PR

This is already the design in `agents/merge-queue.md`. No change needed.

---

## Part 5: Recommendation Summary

| Problem | Who | When | How | INC-002 Safe? |
|---|---|---|---|---|
| **A: Persistent agents stale refs** | Daemon or merge-queue self-initiated | Every 5 min | `git fetch origin main:main` (ref-only) | Yes — no working tree changes |
| **A: Persistent agents stale files** | Agent restart | On agent restart (every 4-6 hours) | Agents re-read CLAUDE.md at spawn | Yes — clean restart |
| **B: Workspace stale** | Human via `/refresh` (now) or daemon (future) | On demand / every 5 min | Stash → rebase → restore (isolated worktree, safe) | Yes — isolated worktree |
| **C: PR branches** | pr-shepherd (conflicts only) | When `mergeable == CONFLICTING` | Temp worktree rebase + force-push-with-lease | Yes — temp worktree |
| **C: PR branch pre-merge** | merge-queue | Not required (ADR-0030) | N/A — push-to-main CI is safety net | N/A |

### Immediate Actions (No Upstream Changes)

1. **merge-queue should run `git fetch origin main:main` in its polling loop** — it already has authority to do this (line 68 of merge-queue.md). This updates the shared checkout's `main` ref so all persistent agents see fresh history.

2. **Human/director should run `/refresh` before starting sessions** — the slash command already exists and works for the workspace worktree.

3. **No change to PR branch handling** — ADR-0030's approach is working. pr-shepherd handles conflicts via temp worktrees.

### Medium-Term (Upstream multiclaude Changes)

4. **Extend daemon refresh loop to include `AgentTypeWorkspace`** — same stash/rebase/restore logic, just widen the type filter. This gives the workspace automatic sync like workers get.

5. **Add ref-only fetch for persistent agents in daemon** — a new code path that does `git fetch origin` in the shared checkout without touching the working tree. This keeps all agents' remote refs fresh.

### Long-Term (Architecture)

6. **Consider per-agent worktrees for persistent agents** (Gastown model) — eliminates the shared checkout entirely. Each persistent agent gets its own branch/worktree. Most robust, but requires significant upstream changes and increases disk usage.

7. **Consider integration branches for epic work** (Gastown model) — workers within an epic target a shared integration branch, land atomically. Solves the cross-PR dependency problem but adds complexity. Evaluate if post-merge CI failures become frequent.

---

## Appendix: Key Source References

| File | What It Shows |
|---|---|
| `multiclaude/internal/daemon/daemon.go:490-607` | Daemon refresh loop — workers only |
| `multiclaude/internal/daemon/daemon.go:1602-1612` | Persistent agents share repo dir |
| `multiclaude/internal/worktree/worktree.go:742-885` | `RefreshWorktree()` — stash/rebase/restore |
| `multiclaude/internal/state/state.go:16-22` | Agent type definitions |
| `agents/merge-queue.md:68` | merge-queue CAN sync local main |
| `agents/pr-shepherd.md:76-84` | pr-shepherd conflict resolution via temp worktree |
| `scripts/hooks/git-safety.sh` | Only blocks workers, not persistent agents |
| `docs/ADRs/ADR-0030` | Intentional no-rebase-before-merge policy |
| `gastownhall/gastown: docs/concepts/integration-branches.md` | Integration branch approach |
| `gastownhall/gastown: docs/design/polecat-lifecycle-patrol.md` | Session-per-step, sandbox persistence |
