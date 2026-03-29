# /rollcall Command Design

**Date:** 2026-03-29
**Status:** Research Complete — Ready for Implementation

---

## 1. Data Source Audit

### Available Now

| Source | What It Provides | Access Method | Latency |
|--------|-----------------|---------------|---------|
| **tmux list-windows** | Agent names, window IDs, active/inactive panes | `tmux list-windows -t mc-ThreeDoors` | <50ms |
| **multiclaude state.json** | Agent type, PID, session_id, task description, created_at, last_nudge, worktree_path, branch | `cat ~/.multiclaude/state.json \| jq` | <50ms |
| **multiclaude worker list** | Worker names, status (running/stopped), branch, task summary | `multiclaude worker list` | ~200ms |
| **multiclaude status** | Daemon status, repo count, agent/worker counts | `multiclaude status` | ~200ms |
| **Process info (ps)** | RSS memory per Claude process, CPU usage | `ps -o rss,pcpu -p <pid>` | <50ms |
| **JSONL transcripts** | Per-message token usage (input, output, cache_create, cache_read) | Parse `~/.claude/projects/<project>/<session_id>.jsonl` | 200ms-2s depending on file size |
| **Git branch** | Current branch per worktree | `git -C <worktree> branch --show-current` | <50ms per worktree |
| **JSONL file size** | Rough proxy for conversation length | `stat` or `du` on the JSONL file | <50ms |
| **tmux capture-pane** | Last visible lines of agent's terminal (status bar, recent output) | `tmux capture-pane -t <window> -p` | <50ms |

### Available But Expensive

| Source | What It Provides | Cost |
|--------|-----------------|------|
| **Full JSONL token aggregation** | Exact cumulative input/output tokens, cache usage | Must read entire JSONL file; files range 0.5MB-10MB+. Takes 200ms-2s per agent. For 12 agents = up to 24s. |
| **tmux capture-pane content parsing** | What the agent is "currently doing" (by reading last few lines) | Fragile — output format varies, may capture mid-render state |

### Not Available

| Source | What We Wished For | Why Not |
|--------|-------------------|---------|
| **Live token count / context remaining** | Real-time context window usage (e.g., "450K / 1M tokens used") | Claude Code doesn't expose this. JSONL has per-message usage but no running total or context window max. The `/stats` command shows session info but cannot be captured programmatically (it's interactive). |
| **Agent "thinking about" status** | What the agent is currently processing | No API — Claude's internal state isn't exposed. Best proxy: tmux capture-pane last lines. |
| **Heartbeat timestamps** | Last time agent responded to a heartbeat | multiclaude doesn't track this. `last_nudge` is when the daemon last nudged the agent, not when the agent last responded. |
| **Message queue depth** | How many unread messages per agent | `multiclaude message list` shows messages for the CURRENT agent only, not cross-agent. state.json has a `msgs` field but it's not reliably populated. |

### Key Finding: Context Usage Feasibility

**JSONL transcripts contain full token usage per message.** Each `assistant` type entry has:
```json
{
  "usage": {
    "input_tokens": 3,
    "cache_creation_input_tokens": 17861,
    "cache_read_input_tokens": 11407,
    "output_tokens": 44
  }
}
```

However:
- **No "context window remaining" field** — we'd have to assume model context limit (1M for Opus) and sum usage
- **Cumulative sums are misleading** — `cache_read_input_tokens` accumulates across turns but doesn't mean context is "used up" (cache reads are re-reads of prior context)
- **The last message's `input_tokens + cache_read_input_tokens`** is the best single-number proxy for "how full is the context" — it represents what the model saw on its most recent turn
- **Newly spawned workers may not have JSONL files yet** — the JSONL appears after the first assistant response
- **JSONL file size is a decent proxy**: 0.5MB = light session, 3MB+ = heavy session, 10MB+ = approaching context limits

**Recommendation:** Use JSONL file size as a fast proxy (< 50ms), with an optional `--detail` flag that parses the last message's token usage for precise numbers.

---

## 2. Implementation Recommendation

### Recommended: Option D — Hybrid (Shell Script + Optional Skill)

**Primary tool:** Shell script at `scripts/rollcall.sh`
**Optional rich view:** Claude Code skill at `.claude/commands/rollcall.md`

#### Rationale

| Option | Speed | Token Cost | Data Richness | Maintenance |
|--------|-------|-----------|---------------|-------------|
| A: Shell Script | <1s | 0 | Good (table) | Low |
| B: Skill Only | 5-15s | ~5K tokens | Rich (interpreted) | Medium |
| C: multiclaude Subcommand | <500ms | 0 | Best | High (upstream) |
| **D: Hybrid** | **<1s (script) or 5-15s (skill)** | **0 (script) or ~5K (skill)** | **Good to Rich** | **Low** |

**Why Hybrid wins:**
1. The shell script handles 95% of use cases — quick status check, no tokens burned
2. The skill can call the shell script and add interpretation ("merge-queue has been idle for 2 hours, may need restart")
3. The shell script is runnable from ANY tmux window, even outside Claude — useful for the human operator
4. The skill format is already the standard for project commands (`.claude/commands/`)
5. Future: if multiclaude adds a native `rollcall` subcommand, the script becomes the fallback

#### Rejected Options

- **Option A (Script Only):** Good but misses the rich interpretation the skill can add. A skill that calls the script gets both.
- **Option B (Skill Only):** Too slow and expensive for a status check you might run 10x/hour. Burns tokens every time.
- **Option C (multiclaude Subcommand):** Best option architecturally but requires multiclaude source changes. We don't control that codebase. Should be proposed upstream as a feature request after the script proves the concept.

---

## 3. Mock Output

### Shell Script Output (fast, default)

```
ThreeDoors Rollcall — 2026-03-29 18:34 CDT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 AGENT              TYPE        STATUS    BRANCH               JSONL    TASK
──────────────────────────────────────────────────────────────────────────────────
 supervisor         supervisor  ● alive   main                  3.1MB   —
 merge-queue        persistent  ● alive   main                  1.3MB   —
 pr-shepherd        persistent  ● alive   main                  2.1MB   —
 arch-watchdog      persistent  ● alive   main                  0.5MB   —
 envoy              persistent  ● alive   main                  0.6MB   —
 project-watchdog   persistent  ● alive   main                  0.8MB   —
 retrospector       persistent  ● alive   main                  0.7MB   —
 calm-bear          worker      ● alive   work/calm-bear          —     CronCreate Heartbeat Viability Study
 fancy-dolphin      worker      ● alive   work/fancy-dolphin      —     Message Queue for Dark Factory
 bright-koala       worker      ● alive   work/bright-koala       —     Design /rollcall Slash Command
 proud-rabbit       worker      ● alive   work/proud-rabbit       —     /implement-story 74.3
──────────────────────────────────────────────────────────────────────────────────
 Daemon: running (PID 41378) | Agents: 7 persistent + 4 workers | Repo: ThreeDoors
```

### With `--detail` flag

```
ThreeDoors Rollcall — 2026-03-29 18:34 CDT (detailed)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 AGENT              TYPE        STATUS    MEM(MB)  JSONL    CTX(K)  UP SINCE     TASK
──────────────────────────────────────────────────────────────────────────────────────
 supervisor         supervisor  ● alive    766     3.1MB    ~580K   11:09        —
 merge-queue        persistent  ● alive    591     1.3MB    ~210K   11:09        —
 pr-shepherd        persistent  ● alive    622     2.1MB    ~340K   11:09        —
 arch-watchdog      persistent  ● alive    437     0.5MB     ~80K   11:09        —
 envoy              persistent  ● alive    457     0.6MB     ~95K   11:09        —
 project-watchdog   persistent  ● alive    439     0.8MB    ~130K   11:09        —
 retrospector       persistent  ● alive    457     0.7MB    ~110K   11:09        —
 calm-bear          worker      ● alive    486       —        —     18:31        CronCreate Heartbeat Via...
 fancy-dolphin      worker      ● alive       —      —        —     18:24        Message Queue for Dark F...
 bright-koala       worker      ● alive       —      —        —     18:33        Design /rollcall Slash C...
 proud-rabbit       worker      ● alive    494       —        —     18:33        /implement-story 74.3
──────────────────────────────────────────────────────────────────────────────────────
 Daemon: running (PID 41378) | Agents: 7 persistent + 4 workers | Repo: ThreeDoors
 CTX(K) = estimated context tokens from last JSONL message (input + cache_read) in thousands
```

### Skill Output (rich, interpreted)

The skill would call the shell script, then add interpretation:

```
## Agent Fleet Status

All 11 agents alive. No issues detected.

### Persistent Agents (7)
- **supervisor** — 3.1MB transcript, heaviest context usage (~580K tokens). Running since 11:09.
- **pr-shepherd** — 2.1MB transcript. Running since 11:09.
- **merge-queue** — 1.3MB transcript. Running since 11:09.
- Others: arch-watchdog, envoy, project-watchdog, retrospector — all light (<1MB).

### Workers (4)
- **proud-rabbit** — implementing Story 74.3 (Provenance Tagging)
- **fancy-dolphin** — researching message queue for dark factory
- **calm-bear** — researching CronCreate heartbeat viability
- **bright-koala** — designing /rollcall command (this task)

### Observations
- No stale workers (all spawned within last 30 min)
- No missing JSONL files for persistent agents
- supervisor context is heaviest at ~580K — consider whether context compression is being triggered
```

---

## 4. What's Feasible Now vs Needs Marvel

### Feasible Now (Week 1)

| Feature | Method |
|---------|--------|
| Agent names + types | state.json |
| Alive/dead status | `kill -0 <pid>` from state.json |
| Task description | state.json `task` field |
| Branch per agent | state.json `worktree_path` + `git branch` |
| Worker vs persistent | state.json `type` field |
| JSONL size as context proxy | `stat` on JSONL file |
| RSS memory | `ps -o rss -p <pid>` |
| Uptime | state.json `created_at` |
| Daemon status | `multiclaude daemon status` |

### Feasible Now (Week 2 — with JSONL parsing)

| Feature | Method |
|---------|--------|
| Approximate context usage | Parse last `assistant` message in JSONL for `input_tokens + cache_read_input_tokens` |
| Message count / conversation length | Count `assistant` entries in JSONL |
| Cost estimate | Sum all usage entries, multiply by model pricing |

### Needs Marvel / Upstream Changes

| Feature | Dependency |
|---------|-----------|
| Real-time context window percentage | Claude Code API or `/stats` programmatic output |
| Agent "currently thinking about" | Claude Code internal state exposure |
| Cross-agent message queue depth | multiclaude enhancement to expose per-agent message counts |
| Heartbeat response timestamps | multiclaude enhancement to track agent responses (not just nudges) |
| Native `multiclaude rollcall` subcommand | multiclaude source contribution |

---

## 5. Draft Implementation

### A. Shell Script: `scripts/rollcall.sh`

```bash
#!/usr/bin/env bash
# rollcall.sh — Quick overview of all running multiclaude agents
# Usage: ./scripts/rollcall.sh [--detail] [--repo REPO]
set -euo pipefail

REPO="${2:-ThreeDoors}"
SESSION="mc-${REPO}"
STATE_FILE="$HOME/.multiclaude/state.json"
JSONL_DIR="$HOME/.claude/projects/-Users-$(whoami)--multiclaude-repos-${REPO}"
DETAIL=false

[[ "${1:-}" == "--detail" ]] && DETAIL=true

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

# Header
echo -e "${BOLD}${REPO} Rollcall — $(date '+%Y-%m-%d %H:%M %Z')${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check prerequisites
if ! command -v jq &>/dev/null; then
    echo "Error: jq required. Install with: brew install jq"
    exit 1
fi

if [[ ! -f "$STATE_FILE" ]]; then
    echo "Error: state.json not found at $STATE_FILE"
    exit 1
fi

# Parse agents from state.json
agents=$(jq -r --arg repo "$REPO" '
    .repos[$repo].agents // {} | to_entries[] |
    [.key, .value.type, .value.pid, .value.session_id, .value.created_at, .value.task // ""] |
    @tsv
' "$STATE_FILE")

# Count agents
persistent_count=0
worker_count=0

# Print header
if $DETAIL; then
    printf " ${DIM}%-18s %-12s %-9s %7s %8s %7s %-8s %s${NC}\n" \
        "AGENT" "TYPE" "STATUS" "MEM(MB)" "JSONL" "CTX(K)" "UP" "TASK"
    echo "──────────────────────────────────────────────────────────────────────────────────────"
else
    printf " ${DIM}%-18s %-12s %-9s %8s %s${NC}\n" \
        "AGENT" "TYPE" "STATUS" "JSONL" "TASK"
    echo "──────────────────────────────────────────────────────────────────────────────────"
fi

while IFS=$'\t' read -r name atype pid session_id created_at task; do
    [[ -z "$name" ]] && continue

    # Status check
    if kill -0 "$pid" 2>/dev/null; then
        status="${GREEN}● alive${NC}"
    else
        status="${RED}● dead${NC}"
    fi

    # Type label
    case "$atype" in
        worker) type_label="worker"; ((worker_count++)) ;;
        supervisor) type_label="supervisor"; ((persistent_count++)) ;;
        workspace) type_label="workspace"; ((persistent_count++)) ;;
        *) type_label="persistent"; ((persistent_count++)) ;;
    esac

    # JSONL size
    jsonl_path="${JSONL_DIR}/${session_id}.jsonl"
    if [[ -f "$jsonl_path" ]]; then
        jsonl_bytes=$(stat -f%z "$jsonl_path" 2>/dev/null || echo 0)
        jsonl_mb=$(echo "scale=1; $jsonl_bytes / 1048576" | bc)
        jsonl_str="${jsonl_mb}MB"
    else
        jsonl_str="—"
    fi

    # Task summary (truncate)
    task_short=$(echo "$task" | head -1 | sed 's/^Research task: //' | sed 's/^\/implement-story /▶ /' | cut -c1-35)

    if $DETAIL; then
        # Memory (RSS in MB)
        claude_pid=$(pgrep -P "$pid" -f "claude" 2>/dev/null | head -1 || true)
        if [[ -n "$claude_pid" ]]; then
            rss_kb=$(ps -o rss= -p "$claude_pid" 2>/dev/null | tr -d ' ')
            mem_mb=$((rss_kb / 1024))
        else
            mem_mb="—"
        fi

        # Context estimate from JSONL last message
        ctx_str="—"
        if [[ -f "$jsonl_path" ]]; then
            ctx_str=$(python3 -c "
import json, sys
last_usage = None
with open('$jsonl_path') as f:
    for line in f:
        d = json.loads(line)
        if d.get('type') == 'assistant' and 'message' in d:
            u = d['message'].get('usage', {})
            if u:
                last_usage = u
if last_usage:
    total = last_usage.get('input_tokens', 0) + last_usage.get('cache_read_input_tokens', 0)
    print(f'~{total // 1000}K')
else:
    print('—')
" 2>/dev/null || echo "—")
        fi

        # Uptime
        up_time=$(echo "$created_at" | sed 's/.*T//' | cut -c1-5)

        printf " %-18s %-12s %-19s %7s %8s %7s %-8s %s\n" \
            "$name" "$type_label" "$(echo -e "$status")" "$mem_mb" "$jsonl_str" "$ctx_str" "$up_time" "$task_short"
    else
        printf " %-18s %-12s %-19s %8s %s\n" \
            "$name" "$type_label" "$(echo -e "$status")" "$jsonl_str" "$task_short"
    fi
done <<< "$agents"

echo "──────────────────────────────────────────────────────────────────────────────────"

# Daemon status
daemon_pid=$(pgrep -f "multiclaude.*daemon" 2>/dev/null | head -1 || echo "?")
echo -e " Daemon: ${GREEN}running${NC} (PID ${daemon_pid}) | Agents: ${persistent_count} persistent + ${worker_count} workers | Repo: ${REPO}"
```

### B. Claude Code Skill: `.claude/commands/rollcall.md`

```markdown
# /rollcall — Agent Fleet Status

Show the status of all running multiclaude agents.

## Instructions

1. Run the rollcall shell script to gather raw data:
   ```bash
   bash scripts/rollcall.sh --detail 2>/dev/null
   ```

2. If the script doesn't exist or fails, gather data manually:
   ```bash
   # Agent list from state.json
   cat ~/.multiclaude/state.json | python3 -c "
   import json, sys, os
   data = json.load(sys.stdin)
   repo = 'ThreeDoors'
   agents = data['repos'][repo]['agents']
   for name, info in sorted(agents.items()):
       pid = info.get('pid', '?')
       alive = os.system(f'kill -0 {pid} 2>/dev/null') == 0
       atype = info.get('type', '?')
       task = (info.get('task', '') or '—')[:50].split('\n')[0]
       status = '● alive' if alive else '● dead'
       print(f'{name:20s} {atype:20s} {status:10s} {task}')
   "
   ```

3. Present the results in a clean table format.

4. Add brief interpretation:
   - Flag any dead agents
   - Note if any persistent agent JSONL is unusually large (>5MB = context pressure)
   - Note if any workers have been running for >2 hours (may be stuck)
   - Note if daemon is not running

5. If agents are dead, suggest: `multiclaude agent restart <name>`
```

---

## 6. Implementation Plan

### Phase 1: Shell Script (immediate)
1. Create `scripts/rollcall.sh` with the draft above
2. Make it executable, test with current agent fleet
3. Add to `justfile`: `rollcall: scripts/rollcall.sh`

### Phase 2: Skill File (same PR)
1. Create `.claude/commands/rollcall.md`
2. Skill calls the shell script, adds interpretation

### Phase 3: Refinements (follow-up)
1. Add `--json` output mode for machine consumption
2. Add `--watch` mode (re-runs every N seconds, like `watch`)
3. Propose `multiclaude rollcall` as upstream feature request

---

## 7. Decision Record

**Decision:** Implement /rollcall as a hybrid shell script + Claude Code skill.

**Adopted:** Option D (Hybrid) — shell script for speed (0 tokens, <1s), skill for rich interpretation.

**Rejected:**
- **Option A (Script Only):** Lacks the interpretive layer that makes the skill valuable for less-obvious status issues (e.g., "this agent's context is getting heavy, consider restart").
- **Option B (Skill Only):** Too expensive for a status check run frequently. Burns ~5K tokens per invocation.
- **Option C (multiclaude Subcommand):** Best long-term option but requires upstream changes. The shell script can serve as the prototype and specification for a future native command.

**Context usage approach:** JSONL file size as fast proxy; optional `--detail` flag parses last JSONL message for precise token counts. Real-time context percentage not feasible without Claude Code API changes.
