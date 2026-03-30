# Perplexity MCP Server — Usage Guidelines

## Overview

Perplexity's official MCP server (`@perplexity-ai/mcp-server`) provides web-grounded research
capabilities to Claude Code agents. It is **disabled by default** to prevent uncontrolled API spend.

## Available Tools

| Tool | Model | Cost (per M tokens) | Use Case |
|------|-------|---------------------|----------|
| `perplexity_search` | sonar | $1 input / $1 output + $5/1K requests | Quick web search with ranked results |
| `perplexity_ask` | sonar-pro | Higher than sonar | Conversational Q&A with web grounding |
| `perplexity_research` | sonar-deep-research | Significantly higher | Comprehensive research reports — **expensive** |
| `perplexity_reason` | sonar-reasoning-pro | Higher than sonar | Complex reasoning and problem-solving |

## Cost Awareness

- **Cheap:** `perplexity_search` — use freely for quick lookups
- **Moderate:** `perplexity_ask`, `perplexity_reason` — use when search isn't enough
- **Expensive:** `perplexity_research` — use sparingly, only for deep research tasks that justify the cost
- **Budget check:** Monitor usage at [perplexity.ai/settings/api](https://perplexity.ai/settings/api) before heavy use

## Agent Access Recommendations

| Agent | Should Have Access | Rationale |
|-------|-------------------|-----------|
| Research workers | Yes (when enabled) | Primary consumers of web research |
| Supervisor | Yes (when enabled) | May need to verify external information |
| merge-queue | No | Git operations only — no research needed |
| pr-shepherd | No | Branch management only — no research needed |
| arch-watchdog | No | Internal codebase analysis only |
| project-watchdog | No | Internal docs management only |
| envoy | Maybe | Issue triage may benefit from external context |
| retrospector | No | Internal analysis only |

## Per-Session Enablement

### Prerequisites

1. npx must be available (verified: v11.10.1 on this system)
2. `PERPLEXITY_API_KEY` must be set in the environment

### Enable for Current Session

1. Set the API key in your environment:
   ```bash
   export PERPLEXITY_API_KEY="pplx-your-key-here"
   ```

2. Copy the example config to the active config:
   ```bash
   cp .claude/settings.local.json.perplexity-example .claude/settings.local.json
   ```
   If `.claude/settings.local.json` already exists, merge the `mcpServers.perplexity` key
   into the existing file rather than overwriting.

3. Restart Claude Code (or the agent) to pick up the new MCP server.

4. Verify tools are available — the agent should have access to `perplexity_search`,
   `perplexity_ask`, `perplexity_research`, and `perplexity_reason`.

### Disable After Session

Remove the perplexity entry from `.claude/settings.local.json`, or delete the file
if it only contained the Perplexity config:
```bash
# If settings.local.json only has Perplexity config:
rm .claude/settings.local.json

# If it has other settings, edit to remove the perplexity mcpServers entry
```

### Enable for Specific Workers

Workers run in isolated worktrees. Three approaches, in order of preference:

**Option A: Environment variable at spawn time (recommended)**
Set `PERPLEXITY_API_KEY` in the supervisor's environment before spawning research workers.
The worker inherits the env var. The MCP config template must exist in the worktree's
`.claude/settings.local.json` for this to work.

**Option B: Worker-specific settings.local.json**
After spawning a worker, copy the example config into the worker's worktree:
```bash
cp .claude/settings.local.json.perplexity-example \
   /Users/skippy/.multiclaude/wts/ThreeDoors/<worker-name>/.claude/settings.local.json
```
The worker's Claude session picks it up on next restart.

**Option C: Supervisor enables globally before spawning**
Enable Perplexity in the main checkout's `.claude/settings.local.json`. All new worktrees
inherit from HEAD, so newly spawned workers get the config automatically. Disable after
the research session.

## When to Use Perplexity

**Good use cases:**
- Researching external libraries, APIs, or tools before integration
- Verifying current best practices for a technology decision
- Investigating bug reports that reference external services
- Deep research for architecture decisions or technology evaluations

**Bad use cases:**
- Questions answerable from the codebase or git history
- General Go or Bubbletea questions (Claude already knows these)
- Anything that doesn't require current/live web data
- Routine implementation work

## Security

- The API key is NEVER committed to the repository
- `.claude/settings.local.json` is gitignored
- The example template references `${PERPLEXITY_API_KEY}` from the environment
- Rotate the key if it is ever accidentally exposed
