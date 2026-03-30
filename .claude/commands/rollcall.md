# /rollcall — Agent Fleet Status

Show the status of all running multiclaude agents with health interpretation.

## Instructions

1. Run the rollcall shell script to gather raw data:
   ```bash
   bash scripts/rollcall.sh --detail 2>/dev/null
   ```

2. If the script doesn't exist or fails, gather data manually:
   ```bash
   cat ~/.multiclaude/state.json | python3 -c "
   import json, sys, os
   data = json.load(sys.stdin)
   repos = data.get('repos', {})
   repo = next(iter(repos), None)
   if not repo:
       print('No repos found in state.json')
       sys.exit(0)
   agents = repos[repo].get('agents', {})
   for name in sorted(agents):
       info = agents[name]
       pid = info.get('pid', 0)
       alive = os.system(f'kill -0 {pid} 2>/dev/null') == 0
       atype = info.get('type', '?')
       task = (info.get('task', '') or '—')[:50].split('\n')[0]
       status = '● alive' if alive else '● dead'
       print(f'{name:20s} {atype:12s} {status:10s} {task}')
   "
   ```

3. Present the results in a clean, readable format.

4. Add brief interpretation after the table:
   - Flag any **dead agents** and suggest: `multiclaude agent restart <name>`
   - Note if any persistent agent JSONL is unusually large (>5MB) — indicates context pressure, may need restart
   - Note if any workers have been running for >2 hours (check created_at vs current time) — may be stuck
   - Note if daemon is not running
   - If all agents are healthy, say so concisely

5. Do NOT suggest actions unless there are problems. A clean rollcall should be brief.
