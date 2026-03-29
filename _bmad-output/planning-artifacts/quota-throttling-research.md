# Claude API/Plan Quota Throttling & Multi-Session Management

**Date:** 2026-03-29
**Type:** Research Spike
**Status:** Complete

---

## 1. Claude Usage Limits & Quota Discovery

### Max Plan Limits (as of March 2026)

Claude Max plans operate on **two independent limit dimensions**:

| Dimension | Reset Cadence | Notes |
|-----------|--------------|-------|
| **5-hour session window** | Rolling 5h from first message | Limits messages/tokens per window |
| **Weekly limit** | 7 days from session start | Two weekly caps: all-models + Sonnet-only |

**Estimated token allocations per 5-hour window** (community-derived, not official):
- **Pro**: ~44,000 tokens / ~45 prompts
- **Max 5x** ($100/mo): ~88,000 tokens / ~225 messages
- **Max 20x** ($200/mo): ~220,000 tokens / ~900 messages

**Weekly limits** (estimated):
- Pro: ~40-80 active Sonnet hours/week
- Max tiers: up to ~480 Sonnet hours or ~40 Opus hours/week

**March 2026 update**: Anthropic reduced session limits during peak hours (05:00-11:00 PT) but expanded off-peak capacity. Weekly totals unchanged — redistribution, not reduction.

### Quota Discovery Methods

**No official API endpoint exists for checking remaining Max plan quota.** Available heuristics:

1. **`/stats` command** — Shows usage patterns for Max/Pro subscribers (not cost)
2. **`/cost` command** — Shows API token usage (relevant for API users, not Max subscribers)
3. **`/context` command** — Shows current context window consumption
4. **Settings > Usage** (claude.ai) — Progress bars for 5-hour and weekly limits
5. **Claude Usage Tracker browser extension** — Monitors remaining quota in claude.ai

**For API users** (not Max plan), rich headers are available:
- `anthropic-ratelimit-requests-remaining` — Requests left in window
- `anthropic-ratelimit-tokens-remaining` — Tokens left (rounded to nearest 1K)
- `anthropic-ratelimit-tokens-reset` — RFC 3339 reset timestamp
- `anthropic-ratelimit-input-tokens-remaining` / `output-tokens-remaining`
- `retry-after` header on 429 responses

**Key insight**: Claude Code on Max plan does NOT go through the API rate limit system — it uses the consumer subscription quota system, which has NO programmatic headers. The only signals are:
- Messages stop being accepted (hard block, no partial throttle)
- No manual reset or support override available

### Third-Party Monitoring Tools

| Tool | Method | Features |
|------|--------|----------|
| [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | Reads JSONL session files | P90 analysis, multi-level alerts, plan detection |
| [ccusage](https://github.com/ryoppippi/ccusage) | Reads local JSONL files | Usage by date/session/project |
| Claude Usage Tracker (Chrome) | Browser extension | Remaining quota display |

The Claude-Code-Usage-Monitor is most relevant — it:
- Reads `~/.claude/` session data (JSONL files)
- Uses P90 percentile analysis over 192 hours (8 days) of history
- Estimates token thresholds: Pro ~44K, Max5 ~88K, Max20 ~220K per window
- Provides color-coded warnings as limits approach

---

## 2. Throttling Protocol Design

### The Fundamental Problem

Claude Code on Max plan provides **no intermediate throttle signal**. It's binary: either messages are accepted or they're blocked until the window resets. This means any throttling must be **self-imposed and heuristic-based**.

### Proposed Throttle Protocol

#### Token Budget Estimation

Track consumption by reading JSONL session logs:
```
~/.claude/projects/<project-hash>/<session-id>.jsonl
```

Each entry contains input/output token counts per interaction. Sum these to estimate window consumption.

#### Throttle Tiers

| Threshold | Action | Rationale |
|-----------|--------|-----------|
| **< 70%** | Normal operation | Full agent activity |
| **70-80%** | Reduce heartbeat frequency | Heartbeats are low-value, high-frequency |
| **80-90%** | Pause non-critical agents | Only workers with active stories continue |
| **90-95%** | Emergency triage | Only P0 work proceeds; all polling stops |
| **> 95%** | Full stop | Queue all work; wait for window reset |

#### Implementation Approaches

**Option A: Daemon-Level Throttle (Recommended)**

The multiclaude daemon already manages agent lifecycles. Add:
1. A token counter that reads JSONL files every 60 seconds
2. Threshold-based message routing: daemon delays or drops heartbeats when above 70%
3. Agent priority classification: P0 (active workers) > P1 (merge-queue) > P2 (heartbeats, polling)
4. Override mechanism: `multiclaude throttle override` for human bypass

**Option B: Agent-Level Self-Throttle**

Each agent tracks its own consumption and self-limits. Simpler but uncoordinated — agents can't see each other's usage.

**Option C: Cron-Based Budget Check**

A cron job runs every 5 minutes, reads total consumption, and sends throttle/resume messages to agents. Lightweight but adds latency.

**Recommended: Hybrid (A + C)**
- Daemon tracks global budget (Option A)
- Cron sends periodic budget snapshots to agents (Option C)
- Agents can self-throttle based on snapshots even if daemon is slow

#### Time-of-Day Awareness

Given March 2026 changes (reduced peak-hour capacity), the throttle should:
- Be more aggressive during 05:00-11:00 PT (reduce threshold by ~20%)
- Allow more aggressive burst usage during off-peak hours (23:00-05:00 PT)
- Track which "regime" is active and adjust thresholds accordingly

#### Window Reset Detection

Since the 5-hour window starts from first message:
1. Record timestamp of first message each window
2. Calculate expected reset time (first_message + 5h)
3. As reset approaches, allow pending high-priority work to queue
4. Resume immediately after reset

---

## 3. Multi-Agent Quota Budgeting

### Current multiclaude Agent Inventory

| Agent | Type | Quota Priority | Activity Pattern |
|-------|------|---------------|-----------------|
| Workers (N) | Ephemeral | P0 — active story work | Burst, high token usage |
| merge-queue | Persistent | P1 — merge operations | Periodic, medium usage |
| pr-shepherd | Persistent | P1 — PR maintenance | Periodic, medium usage |
| envoy | Persistent | P2 — issue triage | Periodic, low-medium usage |
| project-watchdog | Persistent | P2 — governance | Periodic, low usage |
| arch-watchdog | Persistent | P2 — architecture review | Periodic, low usage |
| retrospector | Persistent | P3 — analysis | Periodic, low usage |

### Budget Allocation Strategy

**Fixed Budget Approach** (simpler, less efficient):
- Reserve 60% for workers
- Reserve 20% for P1 agents (merge-queue, pr-shepherd)
- Reserve 10% for P2 agents
- Reserve 10% buffer

**Dynamic Budget Approach** (recommended):
- Workers get unlimited priority up to 80% of window budget
- When workers are idle, P1/P2 agents can use surplus
- Heartbeats are the first thing cut — they're polling, not real work
- Active story work always preempts idle polling

### Heartbeat Optimization

Current heartbeat intervals (from MEMORY.md):
```
merge-queue:      every 7 min
pr-shepherd:      every 7 min (offset 3)
envoy:            every 11 min
retrospector:     every 13 min
project-watchdog: every 23 min
arch-watchdog:    every 23 min (offset 5)
```

**Under quota pressure, dynamically adjust:**

| Quota Used | Heartbeat Multiplier | Effective Intervals |
|------------|---------------------|-------------------|
| < 70% | 1x (normal) | As configured |
| 70-80% | 2x (slower) | 14/22/26/46 min |
| 80-90% | 4x (much slower) | 28/44/52/92 min |
| > 90% | Suspended | No heartbeats |

### Useful Work Detection

Distinguish between:
- **Productive messages**: Tool calls, code changes, PR operations
- **Idle polling**: Heartbeat responses that say "nothing to do"
- **Administrative**: Status checks, message acks

If an agent's last N heartbeats produced no action, deprioritize it further.

---

## 4. Multiple Max Plans on Same Machine

### The CLAUDE_CONFIG_DIR Method

**This is the officially endorsed approach** (confirmed by Anthropic engineer in [claude-code#261](https://github.com/anthropics/claude-code/issues/261)).

```bash
# Account 1 (default)
claude

# Account 2
CLAUDE_CONFIG_DIR=~/.claude-account2 claude

# Account 3
CLAUDE_CONFIG_DIR=~/.claude-account3 claude
```

Each config directory maintains independent:
- OAuth credentials
- Session history (JSONL files)
- `settings.json`
- `CLAUDE.md` (user-level)

### Setup for multiclaude

```bash
# In multiclaude daemon or agent spawn scripts:
export CLAUDE_CONFIG_DIR=~/.claude-plan-${PLAN_NUMBER}

# Example: 3 Max plans
# Plan 1: supervisor + merge-queue + pr-shepherd
# Plan 2: workers 1-3
# Plan 3: workers 4-6 + envoy + watchdogs
```

### Requirements

1. **Separate Anthropic accounts** (different email addresses)
2. **Separate Max subscriptions** ($100-200/mo each)
3. **Separate login** per config dir: `CLAUDE_CONFIG_DIR=~/.claude-plan2 claude` → authenticate
4. **Shared CLAUDE.md**: Symlink `~/.claude-plan-N/CLAUDE.md` → `~/.claude/CLAUDE.md`
5. **Shared settings.json**: Symlink similarly, or maintain separate copies

### Limitations & Risks

- **IDE integration**: `CLAUDE_CONFIG_DIR` may not work properly with VS Code/JetBrains extensions
- **Terms of Service**: Multiple accounts are in a gray area. Anthropic's TOS doesn't explicitly prohibit it, but they reserve the right to enforce "fair use"
- **Per-config-dir session files**: Token tracking must aggregate across config dirs
- **No profile switching**: Each tmux session/agent is locked to one config dir for its lifetime

### Cost Analysis

| Setup | Monthly Cost | Effective Quota |
|-------|-------------|----------------|
| 1x Max 5x | $100 | ~88K tokens/5h, ~480 Sonnet hrs/wk |
| 1x Max 20x | $200 | ~220K tokens/5h, ~480 Sonnet hrs/wk |
| 2x Max 5x | $200 | ~176K tokens/5h combined |
| 3x Max 5x | $300 | ~264K tokens/5h combined |
| 1x Max 20x + 1x Max 5x | $300 | ~308K tokens/5h combined |

**Note**: Two Max 5x plans ($200) provide ~176K tokens vs one Max 20x ($200) at ~220K tokens. A single Max 20x is more cost-efficient than two Max 5x plans. However, multiple plans provide **independent windows** — if one plan hits its limit, the other still has quota.

---

## 5. Session Coordination

### Cross-Session Quota Tracking

**Option A: Shared File Tracker (Recommended for multiclaude)**

```
~/.multiclaude/quota/
├── plan-1.json    # {"tokens_used": 45000, "window_start": "2026-03-29T10:00:00Z"}
├── plan-2.json    # {"tokens_used": 12000, "window_start": "2026-03-29T10:15:00Z"}
└── aggregate.json # Total across plans, computed by daemon
```

The multiclaude daemon reads JSONL files from each `CLAUDE_CONFIG_DIR`, sums usage, and writes aggregate stats. Agents query this before major operations.

**Option B: Daemon-Mediated API**

The daemon exposes a local socket/file-based API:
- `multiclaude quota status` — shows per-plan and aggregate usage
- `multiclaude quota budget <agent>` — returns tokens available for this agent
- `multiclaude quota pause <plan>` — manually pause a plan's agents

**Option C: Independent Throttling (Simplest)**

Each plan manages its own agents independently. No cross-plan coordination. Simpler but wastes quota if one plan is idle while another is maxed.

### Graceful Degradation

When total quota is running low across all plans:

1. **Consolidate**: Move all active work to the plan with the most remaining quota
2. **Defer**: Queue non-urgent work for the next window reset
3. **Notify**: `multiclaude message send supervisor "QUOTA_WARNING: 90% consumed across all plans, next reset in 2h15m"`
4. **Degrade**: Keep only merge-queue alive (it handles the pipeline); pause everything else

### Priority-Based Plan Assignment

Assign agents to plans based on priority, not just round-robin:

| Plan | Agents | Rationale |
|------|--------|-----------|
| Primary (Max 20x) | Supervisor, active workers | Highest quota for highest-value work |
| Secondary (Max 5x) | merge-queue, pr-shepherd | Pipeline work is important but lower volume |
| Tertiary (Max 5x) | envoy, watchdogs, retrospector | Polling agents that can tolerate delays |

---

## 6. Implementation Options

### Phase 1: Passive Monitoring (Low effort, immediate value)

1. Install [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) or build equivalent
2. Add `multiclaude quota status` command that reads JSONL files and reports usage
3. Add quota percentage to agent status output
4. Create a cron that warns supervisor at 80% usage

**Estimated effort**: 1-2 days

### Phase 2: Adaptive Heartbeats (Medium effort)

1. Daemon tracks total token consumption across agents
2. Dynamic heartbeat frequency based on quota consumption (table in Section 3)
3. `multiclaude throttle` command family (status, override, set-threshold)
4. Time-of-day awareness for peak/off-peak

**Estimated effort**: 3-5 days

### Phase 3: Multi-Plan Support (Higher effort)

1. `CLAUDE_CONFIG_DIR` integration in multiclaude agent spawn
2. Per-plan quota tracking
3. Agent-to-plan assignment in multiclaude config
4. Cross-plan coordination via shared quota file
5. `multiclaude quota rebalance` for shifting agents between plans

**Estimated effort**: 5-8 days

### Phase 4: Full Budget System (Highest effort)

1. Per-agent token budgets with enforcement
2. Priority-based preemption (worker takes quota from idle agent)
3. Predictive throttling (ML-based, like the Usage Monitor's P90 approach)
4. Automated plan purchasing/scaling recommendations
5. Historical usage analytics and optimization suggestions

**Estimated effort**: 2-3 weeks

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| QT-1 | Quota tracking method | JSONL file reading (local) | API headers (not available for Max plan), Dashboard scraping (fragile) | Max plan has no programmatic quota API; JSONL files are the only reliable local data source |
| QT-2 | Throttle implementation layer | Hybrid daemon + cron | Agent-only self-throttle, Pure daemon | Daemon provides coordination; cron provides resilience if daemon is slow |
| QT-3 | Heartbeat throttle approach | Dynamic multiplier based on % consumed | Fixed schedules, Binary on/off | Graduated response avoids cliff-edge behavior |
| QT-4 | Multi-plan session isolation | CLAUDE_CONFIG_DIR per plan | Multiple OS users, Docker containers | Official Anthropic-endorsed approach; simplest credential isolation |
| QT-5 | Cross-plan coordination | Shared file tracker | Socket API, No coordination | File-based is simple, daemon-readable, crash-safe |
| QT-6 | Implementation phasing | 4-phase incremental | Big-bang full implementation | Each phase delivers standalone value; later phases can be deferred based on actual need |

## Open Questions

| ID | Question | Context |
|----|----------|---------|
| QT-Q1 | Does Anthropic's TOS explicitly allow multiple Max subscriptions on the same machine? | Gray area — not prohibited but "fair use" clause exists |
| QT-Q2 | Can the daemon detect window resets by monitoring JSONL file activity patterns? | If messages suddenly succeed after a period of 429s, the window likely reset |
| QT-Q3 | Should multiclaude support API key mode alongside Max plan mode for burst capacity? | API keys have explicit rate limit headers and could supplement Max plan during crunch |
| QT-Q4 | What is the actual token budget per 5-hour window? | Community estimates (~88K/~220K) may be inaccurate; needs empirical measurement |

## Sources

- [Claude Max Plan](https://support.claude.com/en/articles/11049741-what-is-the-max-plan)
- [Claude Usage Limits](https://support.claude.com/en/articles/11647753-understanding-usage-and-length-limits)
- [Claude API Rate Limits](https://platform.claude.com/docs/en/api/rate-limits)
- [Claude Code Cost Management](https://code.claude.com/docs/en/costs)
- [Claude Code Multi-Account Issue #261](https://github.com/anthropics/claude-code/issues/261)
- [Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor)
- [ccusage](https://github.com/ryoppippi/ccusage)
- [Multiple Claude Code Accounts Guide](https://medium.com/@buwanekasumanasekara/setting-up-multiple-claude-code-accounts-on-your-local-machine-f8769a36d1b1)
- [Anthropic March 2026 Usage Changes](https://www.theregister.com/2026/03/26/anthropic_tweaks_usage_limits/)
- [Claude Code Limits Guide](https://www.truefoundry.com/blog/claude-code-limits-explained)
