# Persistent Agent Communication Investigation

**Date:** 2026-03-08
**Investigator:** swift-elephant (worker agent)
**Trigger:** Supervisor reported persistent agents failing to respond to messages and appearing stuck

## Executive Summary

Three persistent agents (merge-queue, pr-shepherd, envoy) are running Claude processes but exhibiting communication failures. The root cause is a combination of: (1) Claude cannot hot-reload system prompts mid-session, making definition updates ineffective without restart; (2) stale cached state causing agents to loop on outdated information; and (3) a missing `## Communication` section in envoy's original definition (fix pending in PR #266).

## Findings

### Agent Process Status

All three agents have live Claude processes (version 2.1.71) in their tmux panes. The daemon is healthy (PID 80640, 5 agents tracked). No crashes detected.

| Agent | PID | Process | Status |
|-------|-----|---------|--------|
| merge-queue | 82434 | claude 2.1.71 | Running, stuck in poll loop |
| pr-shepherd | 82500 | claude 2.1.71 | Running, functional but idle |
| envoy | 73261 | claude 2.1.71 | Running, responding to status checks only |

### Issue 1: merge-queue — Stale State Loop

**Symptoms:** Repeatedly reports "Still conflicting. Standing by." for PR #262, even though pr-shepherd already rebased the branch and the PR is now MERGEABLE.

**Root cause:** merge-queue cached the CONFLICTING status from an earlier check and keeps re-polling with the same result. It received the supervisor's "sync git" message but appears to be in a "Baking..." state (Claude thinking indicator) without completing. The agent may have hit a context window limit or is stuck in a long cogitation cycle (one log entry shows "Cogitated for 38m 41s").

**Additional factor:** Earlier in the session, merge-queue hit an OAuth `workflow` scope limitation — it couldn't merge PRs #262 and #263 because they modify `.github/workflows/ci.yml` and the token lacks `workflow` scope. This error was never escalated to the supervisor.

**Actual PR states (as of investigation):**
- PR #262: MERGEABLE, BLOCKED (Performance Benchmarks pending — not conflicting)
- PR #266: MERGEABLE, CLEAN (ready to merge)

**Fix:** Restart merge-queue. The agent has accumulated too much stale context. A fresh start will pick up current PR states correctly.

### Issue 2: envoy — Missing Messaging Instructions (Design Gap)

**Symptoms:** envoy processes work (triage, issue checks, cross-references) and outputs results to its tmux pane, but never sends messages back via `multiclaude message send`.

**Root cause:** The original `agents/envoy.md` definition (committed in PR #227) used vague language like "message supervisor" and "report triage results" without providing the actual `multiclaude message send` CLI command. By contrast, merge-queue and pr-shepherd both have explicit `## Communication` sections with copy-pasteable commands.

PR #266 (branch: `work/bold-squirrel`) fixes this by adding:
- A `## Communication` section with explicit `multiclaude message send` examples
- Updated "Your rhythm" section referencing the specific command
- Bold instruction: "All responses to supervisor and other agents MUST use the messaging system — not tmux output"

**Critical finding:** The supervisor told envoy to "sync git and re-read its agent definition." The envoy did both — but when it re-read `agents/envoy.md`, it saw the *old* version on main (PR #266 hasn't merged). Even if PR #266 had merged, re-reading a file mid-session does NOT change the agent's behavioral patterns. Claude's system prompt is set at spawn time. The envoy explicitly said: "no behavioral changes needed" after re-reading, completely missing the key Communication section that would be added by PR #266.

**Fix:**
1. Merge PR #266 first (it's CLEAN and ready)
2. Then restart envoy (`multiclaude agent restart envoy`) so it loads the new definition as its system prompt

### Issue 3: pr-shepherd — Fork Mode Definition Mismatch

**Symptoms:** pr-shepherd appears functional but references `upstream` remote and fork workflows. The project switched to direct push mode.

**Root cause:** The `agents/pr-shepherd.md` definition still says "You are the PR shepherd for a fork" and references `upstream` throughout. The project switched from fork to direct push (2026-03-07 per MEMORY.md). The pr-shepherd's definition was never updated for direct push mode.

**Current impact:** Low — pr-shepherd is successfully rebasing branches and resolving conflicts despite the mismatch. It adapted because the actual git remote configuration doesn't have an `upstream` remote, so commands fall back to `origin`. However, it creates conceptual confusion and may cause issues if `upstream` is ever configured for another purpose.

**Fix:** Update `agents/pr-shepherd.md` to remove fork references and use `origin` directly. This is a separate task — not blocking.

### Issue 4: Hot-Reload Impossibility (Systemic)

**Root finding:** Claude Code agents cannot hot-reload their system prompts. The agent definition file (`agents/*.md`) is read once at spawn time and becomes part of the system prompt. Telling an agent to "re-read your definition file" only gives it the file contents as conversation context — it does NOT change the behavioral instructions that were baked in at launch.

This means:
- Any definition update requires an agent restart to take effect
- Asking agents to "sync and re-read" is a no-op for behavioral changes
- Agents will acknowledge reading the file but continue with old behavior patterns

**Implications for future operations:**
- Always restart agents after definition updates
- Don't waste time asking agents to re-read definitions — just restart them
- Consider adding a version/hash check so agents can detect when their running definition differs from disk

### Issue 5: merge-queue Not Merging PR #266 (CLEAN)

**Symptoms:** PR #266 is MERGEABLE with mergeStateStatus CLEAN, but merge-queue hasn't attempted to merge it.

**Root cause:** merge-queue is focused on PR #262 (which it thinks is CONFLICTING) and hasn't re-scanned the full PR list to notice #266. Its poll loop checks the same cached PR set.

**Fix:** Restarting merge-queue (Issue 1 fix) will cause it to scan all open PRs fresh.

## Message Delivery Analysis

Messages ARE being delivered by the daemon. The issue is on the receiving end:

- **merge-queue:** Received supervisor's "sync git" message. Displayed it in tmux. Appears to be processing but stuck in cogitation. Never acknowledged or acted on it.
- **envoy:** Received supervisor's "sync git" message. Executed `git fetch` and `git rebase`. Re-read definition file. But continued with old behavior (no `multiclaude message send` responses). Also received "community status" request — generated a detailed report in tmux but never sent it via messaging.
- **pr-shepherd:** Functioning normally. Not involved in the communication failures.

## Recommended Actions (Priority Order)

### Immediate

1. **Merge PR #266** — It's CLEAN and ready. Fixes envoy's missing Communication section. Can be merged manually: `gh pr merge 266 --squash`
2. **Restart merge-queue** — `multiclaude agent restart merge-queue` — clears stale state, picks up fresh PR list
3. **Restart envoy** — `multiclaude agent restart envoy` — loads updated definition with Communication section (after PR #266 merge)

### Short-term

4. **Update pr-shepherd definition** — Remove fork references, align with direct push mode (separate story/PR)
5. **Investigate OAuth workflow scope** — merge-queue can't merge PRs modifying `.github/workflows/` files. Need to update the GitHub token with `workflow` scope or handle this as a known limitation
6. **Check PR #262 Performance Benchmarks** — Currently pending; may need manual intervention if stuck

### Process Improvements

7. **Establish restart-after-update protocol** — Document that agent definition changes require `multiclaude agent restart <name>`, not "re-read your file"
8. **Add definition version checking** — Agents could hash their loaded definition at start and periodically compare against disk to detect drift
9. **Add startup self-test** — Agents send a test message on startup to verify messaging works (noted as opportunity in PR #266)

## Appendix: Agent Definition Comparison

| Feature | merge-queue | pr-shepherd | envoy (current) | envoy (PR #266) |
|---------|-------------|-------------|-----------------|-----------------|
| `## Communication` section | Yes | Yes (implicit) | No | **Yes** |
| Explicit `multiclaude message send` examples | Yes | Yes | No | **Yes** |
| `multiclaude message list` / `ack` instructions | Yes | No | No | **Yes** |
| Loop/polling behavior defined | Yes | Yes | Yes | Yes |
| Authority section | Yes | Yes | Yes | Yes |
| Fork/direct mode alignment | Correct (direct) | **Wrong (fork)** | Correct (direct) | Correct (direct) |
