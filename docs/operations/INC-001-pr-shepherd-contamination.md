# INC-001: pr-shepherd Contamination of Shared Checkout

**Date:** 2026-03-09
**Severity:** High
**Duration:** ~2 hours (discovery + recovery)
**Status:** Resolved

---

## Summary

The pr-shepherd persistent agent performed `git checkout` and `git rebase` operations directly in the shared main repository checkout at `/Users/skippy/.multiclaude/repos/ThreeDoors/`. This caused uncommitted supervisor work to be lost and contaminated the working tree state for all agents sharing that checkout.

---

## What Happened

The pr-shepherd persistent agent, whose job is to keep PRs up-to-date with main, ran branch-switching and rebase commands directly in the shared repository checkout. This checkout is shared by the supervisor, pr-shepherd, merge-queue, and other agents running in tmux sessions. The branch switch destroyed uncommitted work the supervisor had in progress.

### Sequence of Events

1. Multiple PRs were open and needed rebasing onto main (including PRs #418, #419, and others).
2. The pr-shepherd agent ran `git checkout work/witty-owl` in the shared repo checkout.
3. This switched the branch that the supervisor was working on, without the supervisor knowing.
4. pr-shepherd then ran `git rebase origin/main` on that branch.
5. The supervisor had been editing three agent definition files (`agents/envoy.md`, `agents/project-watchdog.md`, `agents/pr-shepherd.md`) with critical policy changes — these edits were uncommitted.
6. The branch switch caused all uncommitted changes to be lost or carried to the wrong branch.
7. Additionally, two custom skill files (`.claude/commands/plan-work.md` and `.claude/commands/sync-enhancements.md`) that were untracked were at risk of loss.
8. The supervisor initially did not realize the contamination had happened — it accepted system reminders about file modifications as intentional, further masking the damage.
9. When the damage was discovered, the rebase was aborted and the supervisor's branch was restored, but all uncommitted edits were permanently lost.

### What Was Lost

1. **`agents/envoy.md`** — Comprehensive rewrite scoping envoy down from full triage authority to go-between with screen-out-only authority.
2. **`agents/project-watchdog.md`** — Epic number mutex protocol (project-watchdog as sole authority for allocating epic/story numbers).
3. **`agents/pr-shepherd.md`** — Git worktree isolation requirement (the very change that would have prevented this incident).
4. **`.claude/commands/plan-work.md`** — Custom slash command for the research-to-stories pipeline.
5. **`.claude/commands/sync-enhancements.md`** — Custom slash command for syncing multiclaude enhancements.

---

## Root Cause Analysis

### Primary Cause

The pr-shepherd agent definition did not prohibit `git checkout` or `git rebase` in the shared repo checkout. It contained examples showing direct `git checkout main && git merge` and `git rebase upstream/main` without any worktree isolation requirement. The agent followed its definition faithfully — the definition was wrong.

### Secondary Cause

The agent definitions assumed each agent had its own isolated git state, but in the multiclaude architecture, persistent agents share the same filesystem checkout. There was no architectural enforcement of isolation — it relied on agents following conventions that did not exist yet.

### Contributing Factors

- **Uncommitted work in the shared checkout was vulnerable** to any agent switching branches. The supervisor held critical policy changes as uncommitted edits for an extended period.
- **Supervisor notification masking** — The supervisor's notification about file modifications was misinterpreted as intentional changes rather than contamination damage. This delayed detection.
- **No automated guard** against `git checkout` in shared checkouts. Nothing in the git configuration or hooks prevented one agent from switching the branch out from under another.
- **Untracked files invisible to git** — Custom skill files in `.claude/commands/` were not version-controlled, making them especially vulnerable to loss.

---

## Impact

| Category | Impact |
|---|---|
| **Work lost** | ~2 hours of policy work (agent definition edits needed to be recreated from memory) |
| **Cascade effect** | Epic number collision (four parallel `/plan-work` workers all claimed Epic 42) went undetected longer because the project-watchdog mutex was not yet in place |
| **Custom skills** | Two custom slash commands lost and needed recreation |
| **Context window** | Supervisor context window consumed by diagnosis and recovery instead of productive work |
| **Trust** | Reduced confidence in persistent agent isolation — required immediate architectural review |

---

## Timeline

| Time | Event |
|---|---|
| T+0 | pr-shepherd runs `git checkout work/witty-owl` in shared checkout |
| T+0 | pr-shepherd runs `git rebase origin/main` on the checked-out branch |
| T+? | Supervisor continues working, unaware branch has changed |
| T+? | Supervisor receives system reminders about file modifications, interprets them as intentional |
| T+~1h | Supervisor discovers branch is wrong, working tree is contaminated |
| T+~1h | Rebase aborted, supervisor branch restored |
| T+~1h | Assessment: all uncommitted edits permanently lost |
| T+~2h | Agent definitions recreated from memory, recovery complete |

---

## Recommendations

### Immediate (Applied in this incident's follow-up PR)

1. **`agents/pr-shepherd.md`** — Added CRITICAL section requiring git worktree isolation for ALL branch operations. Added `git checkout` and `git rebase` in shared checkout to the CANNOT list.
2. **`agents/envoy.md`** — Rebuilt with scoped-down authority: go-between only, screen-out authority, no BMAD pipelines or story creation.
3. **`agents/project-watchdog.md`** — Added epic number mutex protocol: project-watchdog is sole authority for allocating epic/story numbers.

### Preventive

1. **All persistent agent definitions should prohibit `git checkout` in the shared repo.** Add this to a shared "agent rules" preamble that all agent definitions include or reference.
2. **Commit early, commit often.** Agent definition changes should be committed immediately, even as work-in-progress, to prevent loss from any cause.
3. **Supervisor should never hold uncommitted policy changes.** Either commit to a branch immediately or dispatch a worker to make the changes via PR.
4. **Consider git hooks.** A `pre-checkout` hook in the shared repo could warn or block when other tmux sessions are active.
5. **Custom skills should be committed.** Untracked files in `.claude/commands/` are invisible to git and vulnerable to loss. Track them in version control.

### Architectural

1. **Each persistent agent should get its own worktree.** Instead of sharing a checkout, each persistent agent (merge-queue, pr-shepherd, envoy, etc.) should run in a dedicated git worktree. This provides filesystem isolation by default and eliminates this entire class of incident.
2. **Supervisor's checkout should be read-only to other agents.** No agent should modify files in the supervisor's working directory.
3. **Worktree lifecycle management.** If agents use dedicated worktrees, the multiclaude framework should create and clean up worktrees as part of agent spawn/despawn lifecycle.

---

## Resolution

- All three agent definitions have been re-applied with corrective policies.
- Custom skills are being recreated.
- Agents will be restarted with updated definitions after corrective PRs merge.
- This incident report documents the failure mode for future reference and onboarding.

---

## Lessons Learned

1. **Shared mutable state is the root of all evil.** The shared checkout is shared mutable state. Every agent that can write to it is a potential source of contamination.
2. **Agent definitions are the firewall.** Until architectural isolation exists (dedicated worktrees), agent definitions are the only thing preventing cross-contamination. They must be explicit and restrictive.
3. **Uncommitted work is unprotected work.** Git only protects what has been committed. Policy changes, skill files, and any other valuable edits should be committed to a branch immediately.
4. **Silent failures are worse than loud failures.** The supervisor did not receive an obvious error when the branch was switched — the contamination was masked by normal-looking system reminders. Loud failures (hooks, warnings) would have caught this immediately.
