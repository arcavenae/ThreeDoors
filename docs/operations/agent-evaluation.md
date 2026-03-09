# BMAD Agent Evaluation Framework

Evaluation criteria, tuning guidelines, and escalation procedures for the persistent BMAD agent infrastructure (Epic 37). Use this document to assess whether agents are providing value during Phase 1 deployment and to guide ongoing adjustments.

## 1. Success Metrics

### Story Status Accuracy

**Metric:** Percentage of merged PRs with corresponding story status updates within 1 hour of merge.

**Target:** 90% within 1 hour, 100% within 4 hours.

**How to measure:** After each merge, check whether the story file referenced by the PR has been updated to `Done (PR #NNN)`. Track hits and misses over a 2-week window.

| Rating | Threshold | Interpretation |
|--------|-----------|----------------|
| Excellent | >= 95% within 1 hour | project-watchdog is keeping up reliably |
| Acceptable | 80-94% within 1 hour | Minor gaps, likely polling interval or catch-up delays |
| Needs tuning | < 80% within 1 hour | Polling too slow, agent crashing, or PR-to-story mapping failing |

**What success looks like:** A PR merges at 2:15 PM. By 2:30 PM the story file reads `Status: Done (PR #NNN)` and ROADMAP.md epic progress count has incremented. No human intervention required.

### ROADMAP Freshness

**Metric:** Time between an epic progress change (story completion or new story creation) and ROADMAP.md reflecting that change.

**Target:** ROADMAP.md matches actual epic progress within 24 hours.

**How to measure:** Compare the story file statuses in `docs/stories/` against the progress counts in ROADMAP.md. Any discrepancy older than 24 hours is a miss.

| Rating | Threshold | Interpretation |
|--------|-----------|----------------|
| Excellent | Always current within 4 hours | project-watchdog updates ROADMAP as part of its merge cascade |
| Acceptable | Current within 24 hours | Acceptable lag for low-merge periods |
| Needs tuning | Stale for > 24 hours | Agent may not be detecting story completions |

**What success looks like:** Epic 36 has 3 stories. After Story 36.2 merges, ROADMAP.md shows "1/3 stories done" updated to "2/3 stories done" without anyone asking.

### Architecture Doc Currency

**Metric:** New code patterns documented in `docs/architecture/` within 48 hours of the introducing PR merging.

**Target:** 80% of architecture-significant PRs have corresponding doc updates within 48 hours.

**How to measure:** Review PRs that introduced new packages, interfaces, or design patterns. Check whether `docs/architecture/` was updated to reflect them. Baseline expectation from research: 80% architecture divergence coverage.

| Rating | Threshold | Interpretation |
|--------|-----------|----------------|
| Excellent | >= 90% covered within 48 hours | arch-watchdog is catching nearly everything |
| Acceptable | 70-89% covered within 48 hours | Some patterns slipping through, acceptable for Phase 1 |
| Needs tuning | < 70% covered within 48 hours | Detection heuristics need improvement or polling too slow |

**What success looks like:** A PR adds a new `internal/notifications/` package. Within 48 hours, arch-watchdog has either updated architecture docs or opened an issue flagging the undocumented package.

### False Positive Rate

**Metric:** Percentage of agent-generated messages or actions that did not require any follow-up action.

**Target:** Below 20% false positive rate.

**How to measure:** Review all messages sent to supervisor and other agents over a 2-week period. Count messages where the flagged issue was not actually an issue (no action taken, no real drift).

| Rating | Threshold | Interpretation |
|--------|-----------|----------------|
| Excellent | < 10% false positives | Agent judgments are well-calibrated |
| Acceptable | 10-20% false positives | Some noise, but value outweighs cost |
| Needs tuning | > 20% false positives | Agent is too aggressive; tighten detection thresholds |

**What success looks like:** Of 20 messages sent to supervisor in a week, 17 identified genuine issues that needed attention. 3 were false alarms where the agent flagged a doc that was already current.

### Noise Level

**Metric:** Average messages per day sent to supervisor across all agents.

**Target:** 2-5 messages per day during active development periods.

**How to measure:** Count `multiclaude message list` entries addressed to supervisor, averaged over the evaluation period.

| Rating | Threshold | Interpretation |
|--------|-----------|----------------|
| Excellent | 2-5 messages/day | Enough signal to stay informed without being overwhelmed |
| Acceptable | 6-10 messages/day | Manageable during high-merge sprints |
| Needs tuning | > 10 messages/day or < 1 message/day | Too noisy (reduce polling/sensitivity) or too quiet (agent may be stuck) |

**What success looks like:** The supervisor checks messages twice a day and finds 2-3 actionable updates each time, each clearly explaining what changed and what (if anything) needs attention.

## 2. Tuning Guidelines

### When to Increase Polling Frequency

**Trigger:** Multiple PRs merging within a 1-hour window, and story status updates are lagging behind.

**Action:** Reduce project-watchdog polling interval from 10-15 minutes to 5-7 minutes. Reduce arch-watchdog from 20-30 minutes to 10-15 minutes.

**How:** Update the polling interval in the agent definition file (`agents/project-watchdog.md` or `agents/arch-watchdog.md`) and restart the agent.

**Revert when:** Merge rate returns to normal (< 3 PRs/hour). Restore original intervals to conserve API budget.

### When to Decrease Polling Frequency

**Trigger:** No merges for 2+ hours, or GitHub API rate limit warnings appearing in agent logs.

**Action:** Increase polling intervals by 50-100%. project-watchdog: 15-20 minutes. arch-watchdog: 30-45 minutes.

**How:** Update agent definition files and restart. Monitor API usage via `gh api rate_limit`.

**Revert when:** Development activity resumes.

### When to Adjust Authority Boundaries

**Trigger:** An agent consistently escalates the same type of change that could be handled autonomously.

**Examples:**
- project-watchdog escalates every ROADMAP scope question, but 90% of them are straightforward epic progress updates that don't involve scope changes.
- arch-watchdog flags every new test helper file as an undocumented pattern.

**Action:** Expand the agent's authority in its definition file to handle the routine case directly. Keep escalation for genuinely ambiguous situations.

**Guardrail:** Never expand authority to include scope decisions, priority changes, or cross-agent domain writes. Those remain supervisor-only.

### When to Promote a Cron Job to Persistent

**Trigger:** The cron job's 4-hour (SM) or weekly (QA) interval proves insufficient for catching issues in time.

**Indicators:**
- SM sprint health: Stale PRs go unnoticed for a full day because the 4-hour window missed the staleness threshold.
- QA coverage: A coverage regression ships to main and sits for a week before the audit catches it.

**Action:** Convert the cron job to a persistent agent with its own polling loop (15-30 minute intervals). Create a new agent definition file following the project-watchdog/arch-watchdog pattern.

**Guardrail:** Adding a persistent agent increases the total from 5 to 6. The recommended maximum is 6-7 before coordination overhead dominates. Evaluate consolidation before adding.

### API Budget Management

The 5 persistent agents consume an estimated 19-31 GitHub API calls per hour. SM and QA cron jobs add minimal overhead (~6 calls/day and ~1 call/week respectively).

**If approaching rate limits:**
1. Increase polling intervals across all agents by 50%
2. Prioritize project-watchdog (highest governance value) over arch-watchdog
3. Consider combining project-watchdog and arch-watchdog polling into a single API call batch (fetch merged PRs once, dispatch to both agents)

## 3. Phase 1 Evaluation Checklist

Use this checklist after 2 weeks of deployment. All items should be checked before declaring Phase 1 successful.

### Agent Stability

- [ ] Both persistent agents (project-watchdog, arch-watchdog) running for 2+ weeks without crash
- [ ] No manual restarts required beyond initial deployment
- [ ] Agents recover gracefully from GitHub API errors or rate limits

### Automated Governance

- [ ] Story status updates happening automatically on PR merge
- [ ] ROADMAP.md staying current without manual intervention
- [ ] Architecture docs updated when code patterns change
- [ ] No governance gaps (no merged PRs with untracked story updates for > 24 hours)

### Safety and Coordination

- [ ] No circular notification loops observed (agent A messages agent B, agent B messages agent A about the same PR indefinitely)
- [ ] No authority boundary violations observed (agents editing files outside their domain)
- [ ] Correlation ID tracking preventing duplicate processing
- [ ] Separate worktrees preventing file conflicts between agents

### Resource Usage

- [ ] API usage within budget (19-31 calls/hour across all persistent agents)
- [ ] No GitHub API rate limit hits during normal operation
- [ ] tmux sessions stable (no runaway processes or memory leaks)

### Cron Jobs

- [ ] SM cron producing useful sprint summaries every 4 hours
- [ ] SM reports identifying stale PRs before they become blockers
- [ ] QA cron baseline established from first run
- [ ] QA first comparison run completed, identifying any coverage regressions

### Value Assessment

- [ ] Supervisor time spent on manual doc maintenance reduced compared to pre-deployment
- [ ] Story status accuracy metric at "Acceptable" or better (>= 80% within 1 hour)
- [ ] ROADMAP freshness metric at "Acceptable" or better (current within 24 hours)
- [ ] False positive rate at "Acceptable" or better (< 20%)
- [ ] Noise level at "Acceptable" or better (2-10 messages/day)

### Overall Verdict

After checking all items above:

- **All checked, metrics Excellent:** Proceed to Phase 2 (tuning and optimization). Consider expanding agent responsibilities.
- **All checked, metrics Acceptable:** Continue Phase 1 for another 2 weeks with targeted tuning. Address any metrics below "Acceptable".
- **Critical items unchecked:** Investigate root cause. If stability or safety items fail, consider stopping affected agents until issues are resolved.
- **Value items unchecked:** Agents are running but not providing enough value. Evaluate whether the problem is configuration (tunable) or fundamental (rethink approach).

## 4. Escalation Criteria

### When to Stop an Agent

**Immediate stop (within minutes):**
- Agent is creating incorrect file edits that break the build or corrupt docs
- Agent is stuck in a loop sending messages to itself or another agent
- Agent is consuming excessive API calls (> 60/hour sustained)
- Agent is editing files outside its authority domain

**Planned stop (within hours):**
- False positive rate exceeds 30% for 3+ consecutive days
- Noise level exceeds 15 messages/day for 3+ consecutive days with no corresponding increase in merge activity
- Agent consistently produces low-quality doc updates that require manual correction

**How to stop:**
```bash
multiclaude worker rm <agent-name>
```

**Post-stop:** Investigate root cause in agent logs (`multiclaude logs <agent-name>`). Fix the agent definition or underlying issue before restarting.

### When to Add a Third Persistent Agent

**Criteria (all must be true):**
1. A clear governance gap exists that neither project-watchdog nor arch-watchdog covers
2. The gap cannot be addressed by expanding an existing agent's authority
3. The gap requires continuous monitoring (not periodic — cron would suffice for periodic)
4. The current 5 persistent agents are stable and within API budget

**Examples of valid reasons:**
- A dedicated "dependency watchdog" that monitors `go.mod` changes and validates compatibility across the dependency graph — if dependency issues become frequent
- A dedicated "test watchdog" that monitors coverage regressions in real-time — if the weekly QA cron proves too slow

**Examples of invalid reasons:**
- "It would be nice to have" — not a valid justification; wait for demonstrated need
- Duplicating an existing agent's work with slightly different rules — consolidate instead

### When to Consolidate Agents

**Trigger:** Two agents are doing overlapping work or one agent has very little to do.

**Indicators:**
- arch-watchdog processes < 2 PRs per week (not enough code changes to justify a persistent agent)
- Both watchdogs are checking the same PRs and sending redundant messages about the same issues
- One agent is effectively idle for > 50% of its uptime

**Action:** Merge the less-active agent's responsibilities into the more-active one. Update the surviving agent's definition file to include the merged responsibilities. Remove the redundant agent.

**Guardrail:** After consolidation, monitor the surviving agent's noise level and false positive rate for 1 week. If quality degrades, split them back apart.
