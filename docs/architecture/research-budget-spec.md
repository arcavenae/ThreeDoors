# Research Budget Management Specification

**Story:** 54.5 — Rate Limiting, Budget Management & Model Selection
**Epic:** 54 — Gemini Research Supervisor
**Status:** Active specification — authoritative for research-supervisor agent behavior

---

## Overview

The research-supervisor agent operates under free-tier rate limits for the Gemini API. This specification defines the budget tracking schema, daily reset logic, model selection heuristics, priority queue behavior, batch optimization, deduplication, cooldown rules, and escalation protocols.

All budget management is **convention-based** — the research-supervisor agent maintains `budget.json` through its own reasoning. There is no Go code enforcing these rules.

---

## 1. Budget Schema (`budget.json`)

**Location:** `_bmad-output/research-reports/budget.json`

```json
{
  "date": "2026-03-13",
  "pro_limit": 50,
  "pro_used": 7,
  "pro_remaining": 43,
  "flash_limit": 1000,
  "flash_used": 12,
  "flash_remaining": 988,
  "queries": [
    {
      "id": "20260313-143022-yaml-security",
      "timestamp": "2026-03-13T14:30:22Z",
      "depth": "standard",
      "model": "gemini-2.5-pro",
      "priority": "high",
      "requester": "supervisor",
      "status": "completed"
    }
  ]
}
```

### Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `date` | `string` (YYYY-MM-DD) | Current UTC date. Used to detect day rollover. |
| `pro_limit` | `int` | Daily Pro query limit. Fixed at 50. |
| `pro_used` | `int` | Pro queries dispatched today. |
| `pro_remaining` | `int` | `pro_limit - pro_used`. Maintained for quick reads. |
| `flash_limit` | `int` | Daily Flash query limit. Fixed at 1000. |
| `flash_used` | `int` | Flash queries dispatched today. |
| `flash_remaining` | `int` | `flash_limit - flash_used`. Maintained for quick reads. |
| `queries` | `array` | Ordered list of all queries dispatched today. |

### Query Record Fields

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `id` | `string` | `YYYYMMDD-HHMMSS-<slug>` | Unique query identifier. Slug derived from first 40 chars of question. |
| `timestamp` | `string` (ISO 8601) | UTC timestamp | When the query was dispatched. Always UTC. |
| `depth` | `string` | `quick`, `standard`, `deep` | Research depth requested. |
| `model` | `string` | `gemini-2.5-pro`, `gemini-2.5-flash` | Model actually used (may differ from depth default after fallback). |
| `priority` | `string` | `high`, `normal`, `low` | Request priority level. |
| `requester` | `string` | Agent name | Which agent requested this query. |
| `status` | `string` | `completed`, `failed`, `downgraded` | Final status. `downgraded` means Pro was requested but Flash was used. |

---

## 2. Daily Reset Logic

**Trigger:** Midnight UTC (detected by comparing `budget.json.date` to current UTC date).

**Reset behavior:**
1. Compare `date` field in `budget.json` to `time.Now().UTC().Format("2006-01-02")`
2. If dates differ:
   - Set `date` to today's UTC date
   - Set `pro_used` to 0, `pro_remaining` to 50
   - Set `flash_used` to 0, `flash_remaining` to 1000
   - Clear the `queries` array (yesterday's queries are preserved in individual `request.json` files per query directory)
3. Write the reset `budget.json` to disk

**On agent startup/restart:** Always check for day rollover before processing any requests. If the agent was down across midnight, the first poll cycle handles the reset.

---

## 3. Reserve Pool Rule

**Rule:** After 18:00 UTC, reserve the last 5 Pro queries for high-priority requests only.

**Implementation:**

```
if current_utc_hour >= 18 AND pro_remaining <= 5:
    if request.priority == "high":
        → dispatch with Pro (reserve pool allows high-priority)
    else:
        → downgrade to Flash (if depth allows) OR queue for next day
```

**Rationale:** Late in the day, remaining Pro budget should be preserved for urgent requests. Normal and low-priority queries can use Flash or wait until the budget resets at midnight UTC.

**Edge cases:**
- If Pro is already exhausted (0 remaining), the reserve rule is moot — escalate per Section 10.
- Flash queries are unaffected by the reserve rule (1,000/day is effectively unlimited).

---

## 4. Priority Queue Behavior

| Priority | Dispatch Behavior | Budget Gate |
|----------|-------------------|-------------|
| `high` | Execute immediately on next poll cycle, skip queue | Always dispatched (even from reserve pool after 18:00 UTC). If Pro exhausted, escalate to supervisor. |
| `normal` | FIFO order within budget | Dispatched if Pro remaining > 5 (or > 0 before 18:00 UTC). After 18:00 UTC with <=5 Pro remaining, downgrade to Flash or queue. |
| `low` | Execute only if daily budget is comfortable | Dispatched only if Pro remaining > 10. Otherwise queued or downgraded to Flash. |

**Queue persistence:** The queue lives in agent memory (not on disk). If the agent restarts, queued requests are lost. Requesters should re-send if no response within 30 minutes.

**Queue ordering:** Within the same priority level, FIFO. Across priority levels: high > normal > low.

---

## 5. Model Selection Heuristic

### Depth-to-Model Mapping

| Depth | Default Model | Rationale |
|-------|---------------|-----------|
| `quick` | `gemini-2.5-flash` | Always Flash. Quick lookups don't justify Pro budget. |
| `standard` | `gemini-2.5-pro` | Pro for synthesis/analysis. Falls back to Flash if Pro budget low. |
| `deep` | `gemini-2.5-pro` | Pro required for comprehensive research. If Pro exhausted, inform requester (see Section 10). |

### Fallback Logic

```
1. Determine default model from depth
2. If default is Pro and Pro budget is exhausted:
   a. depth=standard → downgrade to Flash, notify requester of downgrade
   b. depth=deep → do NOT silently downgrade. Notify requester:
      "Pro is exhausted. Options: (1) run with Flash (reduced quality),
       (2) queue for next day when Pro resets."
3. If default is Flash → always dispatch (effectively unlimited budget)
```

### 429 Rate-Limit Fallback (Pro to Flash)

If the Gemini API returns a 429 (rate limit) error for a Pro query:

1. Log the 429 with timestamp and query ID in `budget.json` query record (set status to `downgraded`)
2. Automatically retry the same query with Flash
3. Notify the requester: "Pro rate-limited, query dispatched with Flash instead"
4. Do NOT count the failed Pro attempt against the Pro budget (the API rejected it)
5. Count the successful Flash retry against the Flash budget
6. If Flash also returns 429, halt and escalate to supervisor

---

## 6. Batch Optimization

### Detection Criteria

Batch optimization triggers when **3 or more** queued normal-priority requests share overlapping keywords.

**Keyword overlap detection:**
1. Tokenize each queued request's question text (split on spaces, lowercase, remove stop words)
2. Compute pairwise Jaccard similarity between token sets
3. If any group of 3+ requests has pairwise similarity > 0.3, flag as batchable

**Stop words to exclude:** the, a, an, is, are, was, were, be, been, being, have, has, had, do, does, did, will, would, could, should, may, might, can, shall, for, and, but, or, nor, not, no, so, yet, both, either, neither, each, every, all, any, few, more, most, other, some, such, than, too, very, just, how, what, when, where, which, who, whom, why, this, that, these, those, it, its

### Merge Strategy

When a batch is detected:

1. Extract the union of all unique non-stop-word keywords from the batch
2. Synthesize a merged prompt that covers all original questions:
   ```
   "Comprehensive analysis covering: [merged question topics].
    Original questions for reference:
    1. [question 1]
    2. [question 2]
    3. [question 3]"
   ```
3. Dispatch as a single `depth=deep` query (the merged query is broader, justifying deeper research)
4. Use the highest priority from any request in the batch
5. Distribute the result to all original requesters
6. Log in `budget.json` as one query with a note: `"batch_size": 3, "merged_from": ["id1", "id2", "id3"]`

### Savings Reporting

When a batch is executed, include in the result notification:
```
"[BATCH] Merged 3 queries into 1. Budget savings: 2 Pro queries."
```

---

## 7. Deduplication

### 7-Day Lookback

Before dispatching any query, search existing research reports for similar prior results.

**Search procedure:**
1. List all directories in `_bmad-output/research-reports/` from the last 7 days (by directory name prefix `YYYYMMDD-`)
2. Read the `request.json` file from each directory
3. Compare the new query's keywords against the stored query text
4. If keyword overlap > 50% (Jaccard similarity on non-stop-word tokens), flag as potential duplicate

### Requester Notification

When a duplicate is found:
```bash
multiclaude message send <requester> "RESEARCH-DUPLICATE: Your query '<slug>'
is similar to existing research from <date>:
  Report: _bmad-output/research-reports/<path>/report.md
  Summary: _bmad-output/research-reports/<path>/executive-summary.md
If this doesn't cover your needs, reply with 'RESEARCH-OVERRIDE priority=<P> depth=<D>: <refined question>' to force a new query."
```

**Override mechanism:** The requester can force a new query by sending a `RESEARCH-OVERRIDE` message. Overrides skip deduplication but still respect budget limits.

---

## 8. Cooldown Rules

### Pro Query Cooldown

**Rule:** Minimum **2 minutes** between consecutive Pro query dispatches.

**Rationale:** Prevents accidental rapid-fire usage of the limited Pro budget. Also respects Google's per-minute rate limits for the free tier.

**Implementation:**
1. After dispatching a Pro query, record the dispatch timestamp
2. Before dispatching the next Pro query, check elapsed time
3. If < 2 minutes elapsed, hold the query in queue until the cooldown expires
4. Process the next item from the queue (which may be a Flash query — those aren't cooldown-restricted)

### Flash Query Cooldown

**Rule:** No explicit cooldown between Flash queries, but respect the API's 60 RPM (requests per minute) limit.

**Implementation:** If dispatching more than 1 Flash query per second, add a 1-second delay between dispatches. In practice, the 5-minute poll cycle means multiple Flash queries per second is unlikely.

---

## 9. Budget Warning Threshold

**Rule:** At **80% daily Pro usage** (40 out of 50 queries used), send a budget warning to the supervisor.

**Warning message:**
```bash
multiclaude message send supervisor "RESEARCH-BUDGET: WARNING — Pro queries at 80%
(40/50 used, 10 remaining). Flash queries: <M>/1000 used. <P> requests queued.
Consider: prioritize remaining Pro for high-priority requests only."
```

**When to send:**
- Immediately after the 40th Pro query completes
- Only send once per day (track `warning_sent` flag, reset with daily reset)
- Do NOT send additional warnings at 90% or 100% — the supervisor is already informed

---

## 10. RESEARCH-BUDGET Escalation Message Format

### Pro Budget Exhausted + High-Priority Request

When Pro is fully exhausted (0 remaining) and a high-priority request arrives:

```bash
multiclaude message send supervisor "RESEARCH-BUDGET: ESCALATION — Pro budget
exhausted (50/50 used today). High-priority request pending from <requester>:
  Query: '<first 100 chars of question>...'
  Options: (1) Approve Flash fallback (reduced depth), (2) Queue for tomorrow's
  Pro reset at midnight UTC, (3) Override — the supervisor can instruct the
  research-supervisor to dispatch anyway if server-side limits allow.
Awaiting instruction."
```

### End-of-Day Budget Report

At the last poll cycle before midnight UTC (or when the agent detects day rollover):

```bash
multiclaude message send supervisor "RESEARCH-BUDGET: Daily Report for <date>
  Pro:   <N>/50 used (<remaining> remaining)
  Flash: <M>/1000 used (<remaining> remaining)
  Total queries: <total>
  Batches: <batch_count> (saved <savings> queries)
  Duplicates caught: <dedup_count>
  Downgrades (Pro→Flash): <downgrade_count>
  Queued for tomorrow: <queue_count>
Full query log: _bmad-output/research-reports/budget.json"
```

---

## Decision Record

**Adopted:** Convention-based budget management (agent self-tracks via `budget.json`).

**Why:** Queries are free-tier — no financial risk from missed enforcement. The agent's reasoning is sufficient for advisory budget tracking. Server-side limits provide a hard backstop.

**Rejected alternatives:**
- **Go code enforcement:** Over-engineered for free-tier queries. Would require a new Go package, HTTP middleware, and testing infrastructure for zero financial upside.
- **Shared database for budget:** Multi-agent budget contention is not a real problem — only one agent (research-supervisor) dispatches queries.
- **Cost-per-query tracking:** Removed from original design. Queries are free; cost fields would be misleading zeros.
- **`--dry-run` gate:** Removed. Dry-run estimated per-query cost for paid API. Free tier has no per-query cost to estimate.
