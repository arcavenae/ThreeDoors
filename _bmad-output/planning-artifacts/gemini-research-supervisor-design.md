# Gemini Research Supervisor — Architecture Design

**Date:** 2026-03-11
**Type:** Deep Research / Architecture Design
**Requested by:** User (supervisor task)
**Status:** Draft

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Integration Options Analysis](#integration-options-analysis)
3. [Chosen Architecture: agent-deep-research Skill](#chosen-architecture)
4. [Research Supervisor Agent Design](#research-supervisor-agent-design)
5. [Context Packaging Strategy](#context-packaging-strategy)
6. [Result Ingestion & Shielding](#result-ingestion--shielding)
7. [Rate Limiting & Scheduling](#rate-limiting--scheduling)
8. [Agent Definition Draft](#agent-definition-draft)
9. [Workflow Examples](#workflow-examples)
10. [Rejected Alternatives](#rejected-alternatives)

---

## Executive Summary

This document designs a **research-supervisor** persistent agent that manages Gemini Deep Research queries on behalf of the multiclaude agent system. The agent sits between the main supervisor and Google's Gemini Deep Research API, handling prompt engineering, context assembly, quota management, result summarization, and artifact storage.

**Key design decisions:**
- Use `24601/agent-deep-research` as the execution layer (battle-tested CLI tool, Claude Code native, uv-based)
- Research-supervisor is a **persistent agent** (like envoy or arch-watchdog) — always running, polling for requests
- Results stored in `_bmad-output/research-reports/` with structured layering (executive summary → detailed → raw)
- 50 queries/day budget managed via a priority queue with cost tracking
- Context bundles assembled per-query from curated document sets

**What this enables:**
- Any agent can request deep research via messaging: `multiclaude message send research-supervisor "Research: [question]"`
- Research runs asynchronously (5-45 min) without blocking any agent
- Results are shielded — only summaries enter agent context; full reports live on disk
- Budget transparency — daily usage tracked, queries prioritized by urgency

---

## Integration Options Analysis

### Option A: Gemini CLI + Deep Research Extension

```
┌─────────────┐    message     ┌────────────────────┐    gemini CLI    ┌─────────────┐
│  Supervisor  │──────────────▶│ Research-Supervisor │────────────────▶│  Gemini CLI  │
│  (or agent)  │◀──────────────│   (persistent)      │◀────────────────│  + extension │
└─────────────┘    result      └────────────────────┘    report        └─────────────┘
```

- **Pros:** Official Google CLI, first-party extension, MCP-based
- **Cons:** Requires Gemini CLI installed globally, adds another agent runtime (Gemini CLI is itself an agent), extension requires paid API key (not OAuth), heavyweight

### Option B: Direct Interactions API via curl/Python

```
┌─────────────┐    message     ┌────────────────────┐    curl/python   ┌─────────────┐
│  Supervisor  │──────────────▶│ Research-Supervisor │────────────────▶│ Gemini API   │
│  (or agent)  │◀──────────────│   (persistent)      │◀────────────────│ Interactions │
└─────────────┘    result      └────────────────────┘    JSON          └─────────────┘
```

- **Pros:** No dependencies beyond curl, full control over request/response, lightweight
- **Cons:** Must implement polling, error handling, reconnection, output parsing from scratch; fragile to API changes

### Option C: agent-deep-research Skill (CHOSEN)

```
┌─────────────┐    message     ┌────────────────────┐   uv run         ┌──────────────────┐
│  Supervisor  │──────────────▶│ Research-Supervisor │──────────────────▶│ agent-deep-      │
│  (or agent)  │◀──────────────│   (persistent)      │◀──────────────────│ research scripts │
└─────────────┘    summary     └────────────────────┘   JSON+report     └──────────────────┘
                                       │                                         │
                                       │ stores artifacts                        │ calls
                                       ▼                                         ▼
                               _bmad-output/                            Gemini Interactions
                               research-reports/                        API (deep-research-
                                                                        pro-preview)
```

- **Pros:**
  - Purpose-built for agent consumption (dual stdout=JSON / stderr=human output)
  - Adaptive polling already implemented (history-based, p25-p75 aggressive window)
  - `--output-dir` gives structured output (report.md, metadata.json, sources.json)
  - `--context` flag enables local file grounding (RAG) — perfect for project context
  - `--dry-run` for cost estimation before committing quota
  - `--depth quick|standard|deep` controls research duration (2-45 min)
  - Non-interactive mode auto-skips prompts (agent-safe)
  - No Gemini CLI dependency — uses `uv run` with google-genai SDK directly
  - MIT licensed, fully auditable Python scripts
  - Already designed for Claude Code integration
- **Cons:** Third-party dependency (but MIT, auditable, well-maintained)

**Decision:** Option C. The agent-deep-research tool solves every infrastructure problem (polling, output formatting, cost estimation, RAG grounding) that we'd otherwise have to build ourselves. It's designed exactly for our use case.

---

## Chosen Architecture

### System Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    multiclaude Agent System                       │
│                                                                  │
│  ┌────────────┐    ┌──────────────┐    ┌───────────────────┐    │
│  │ Supervisor  │    │   Workers    │    │ Other Persistent  │    │
│  │            │    │  (ephemeral) │    │ Agents            │    │
│  └─────┬──────┘    └──────┬───────┘    └────────┬──────────┘    │
│        │                  │                      │               │
│        │    multiclaude message send              │               │
│        ▼                  ▼                      ▼               │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Research Supervisor (persistent)             │    │
│  │                                                          │    │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────┐             │    │
│  │  │ Request   │  │ Context  │  │ Budget    │             │    │
│  │  │ Queue     │  │ Assembler│  │ Tracker   │             │    │
│  │  └────┬─────┘  └────┬─────┘  └─────┬─────┘             │    │
│  │       │              │              │                    │    │
│  │       ▼              ▼              ▼                    │    │
│  │  ┌──────────────────────────────────────────┐           │    │
│  │  │         Dispatch & Monitor Loop           │           │    │
│  │  └────────────────┬─────────────────────────┘           │    │
│  └───────────────────┼──────────────────────────────────────┘    │
│                      │                                           │
└──────────────────────┼───────────────────────────────────────────┘
                       │  uv run scripts/research.py
                       ▼
              ┌────────────────────┐
              │ agent-deep-research│
              │ (local scripts)    │
              └────────┬───────────┘
                       │  Interactions API
                       ▼
              ┌────────────────────┐
              │ Google Gemini      │
              │ Deep Research      │
              │ (cloud)            │
              └────────────────────┘
```

### Installation Prerequisites

```bash
# 1. Install agent-deep-research (one-time setup)
cd /path/to/ThreeDoors
git clone https://github.com/24601/agent-deep-research.git _tools/agent-deep-research

# 2. Set API key (user's OAuth key or AI Studio key)
# Add to ~/.zshrc or agent environment:
export GEMINI_DEEP_RESEARCH_API_KEY="user-api-key-here"

# 3. Verify uv is available (already required by project conventions)
uv --version

# 4. Test connectivity
cd _tools/agent-deep-research
uv run scripts/research.py start "Test query: what is 2+2?" --depth quick --dry-run
```

### Directory Layout

```
ThreeDoors/
├── _tools/
│   └── agent-deep-research/          # Cloned tool (gitignored)
├── _bmad-output/
│   ├── planning-artifacts/            # Existing artifact store
│   └── research-reports/              # NEW: Gemini research output
│       ├── YYYYMMDD-HHMMSS-<slug>/   # Per-query directory
│       │   ├── report.md              # Full research report
│       │   ├── executive-summary.md   # Extracted summary (≤500 words)
│       │   ├── metadata.json          # Timing, cost, status
│       │   ├── sources.json           # Citations and URLs
│       │   ├── context-bundle.md      # What context was sent
│       │   └── request.json           # Original request metadata
│       └── budget.json                # Daily usage tracking
└── agents/
    └── research-supervisor.md         # Agent definition
```

---

## Research Supervisor Agent Design

### Role & Responsibilities

The research-supervisor is a **persistent agent** that:

1. **Receives** research requests via multiclaude messaging
2. **Validates** requests against budget and priority
3. **Assembles** context bundles tailored to each query
4. **Formulates** optimized prompts for Gemini Deep Research
5. **Dispatches** queries via agent-deep-research scripts
6. **Monitors** async execution (research takes 2-45 min)
7. **Processes** results into layered summaries
8. **Delivers** findings back to the requesting agent
9. **Tracks** daily budget and query history

### Polling Loop

```
Every 5 minutes:
  1. Check messages (multiclaude message list)
  2. For each new research request:
     a. Parse request, determine priority and depth
     b. Check budget — can we afford this query?
     c. If yes: assemble context, formulate prompt, dispatch
     d. If no: queue with priority, notify requester of delay
  3. Check running queries (poll active interaction IDs)
  4. For completed queries:
     a. Process output into layered summaries
     b. Store artifacts
     c. Notify requesting agent with executive summary
  5. End-of-day: log budget report
```

### Request Protocol

Agents request research via structured messages:

```bash
# Standard research request
multiclaude message send research-supervisor "RESEARCH priority=normal depth=standard: How do other Go TUI apps handle task persistence? Compare SQLite vs YAML vs JSON approaches."

# Urgent request (uses quota sooner)
multiclaude message send research-supervisor "RESEARCH priority=high depth=deep: Security implications of YAML parsing in Go — known CVEs, best practices, safe parsers."

# Quick lookup (minimal quota usage)
multiclaude message send research-supervisor "RESEARCH priority=normal depth=quick: What is the current state of Charm's Huh library for form inputs?"

# With explicit context needs
multiclaude message send research-supervisor "RESEARCH priority=normal depth=standard context=architecture,soul: Best practices for Bubbletea state management in apps with 10+ views."
```

**Message format:**
```
RESEARCH priority=<high|normal|low> depth=<quick|standard|deep> [context=<bundle-names>]: <question>
```

### Priority Queue

| Priority | Behavior | Use Case |
|----------|----------|----------|
| **high** | Execute immediately (next poll cycle) | Security issues, blocking decisions, user-requested |
| **normal** | Execute in FIFO order within budget | Architecture research, technology evaluation |
| **low** | Execute only if daily budget allows | Nice-to-have, background exploration |

---

## Context Packaging Strategy

### The Problem

Gemini Deep Research can ground queries in uploaded files via `--context`, but dumping the entire repo would be wasteful, slow, and noisy. We need curated context bundles.

### Context Bundles

Pre-defined bundles that the research-supervisor assembles based on the query:

| Bundle Name | Files Included | Size (est.) | When to Use |
|-------------|---------------|-------------|-------------|
| `core` | CLAUDE.md, SOUL.md | ~8KB | Always included |
| `architecture` | docs/architecture/*.md | ~15KB | Architecture questions |
| `prd` | docs/prd/epic-list.md, relevant epic sections | ~10KB | Feature/scope questions |
| `stories` | Relevant story files (pattern-matched) | ~5KB | Story-specific research |
| `decisions` | docs/decisions/BOARD.md | ~8KB | Decision-related queries |
| `code-sample` | Relevant source files (grep-matched) | ~20KB | Implementation research |
| `tui` | internal/tui/*.go (key files only) | ~15KB | TUI-specific research |
| `tasks` | internal/tasks/*.go (key files only) | ~15KB | Task domain research |

### Assembly Algorithm

```
1. Always include `core` bundle (CLAUDE.md + SOUL.md)
2. If request specifies context=<names>, include those bundles
3. If no context specified, auto-detect:
   a. Scan question for keywords → map to bundles
      - "architecture", "design", "pattern" → architecture
      - "epic", "story", "feature", "scope" → prd + stories
      - "decision", "rejected", "alternative" → decisions
      - "bubbletea", "view", "tui", "render" → tui
      - "task", "provider", "pool", "persistence" → tasks
   b. Include matched bundles up to 60KB total budget
4. Write assembled context to a temp directory
5. Pass to agent-deep-research via --context flag
```

### Token Budget Management

```
Target: ≤60KB of context per query (~15K tokens)
Reason: Keeps grounding focused. Gemini does its own web search —
        our context provides project-specific grounding, not general knowledge.

If assembled context > 60KB:
  1. Drop code-sample bundle (most expendable)
  2. Truncate story files to headers + acceptance criteria only
  3. If still over, drop prd bundle
  4. core + architecture + decisions is the irreducible minimum (~31KB)
```

### Prompt Engineering

The research-supervisor doesn't just forward raw questions. It wraps them:

```markdown
## Research Context

You are researching a question for the ThreeDoors project — a Go TUI application
built with Bubbletea that shows only three tasks at a time to reduce decision
friction. The project uses YAML for task persistence, follows strict Go idioms
(see attached CLAUDE.md), and prioritizes simplicity over features (see SOUL.md).

## Grounding Files

The attached files provide project-specific context. Reference them when relevant
but don't limit your research to their contents.

## Research Question

[USER'S QUESTION HERE]

## Output Requirements

1. Start with an executive summary (≤300 words)
2. Provide detailed findings with citations
3. Include a "Relevance to ThreeDoors" section mapping findings to our constraints
4. List rejected or inferior approaches with brief rationale
5. End with concrete recommendations ranked by confidence
```

---

## Result Ingestion & Shielding

### The Problem

Deep research reports can be 5,000-15,000 words. Piping that directly into an agent's context window is wasteful and risks information overload.

### Three-Layer Summary Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Layer 1: Executive Summary (≤500 words)                  │
│ • Key findings (3-5 bullets)                             │
│ • Top recommendation                                     │
│ • Confidence level                                       │
│ • "Full report: _bmad-output/research-reports/<path>"    │
│                                                          │
│ ──── This is what gets sent to the requesting agent ──── │
├─────────────────────────────────────────────────────────┤
│ Layer 2: Detailed Findings (full report.md)               │
│ • Complete analysis with citations                        │
│ • Comparative tables                                      │
│ • Relevance-to-ThreeDoors section                        │
│ • Rejected alternatives with rationale                    │
│                                                          │
│ ──── Stored on disk, read on demand ────                 │
├─────────────────────────────────────────────────────────┤
│ Layer 3: Raw Data (metadata.json, sources.json)           │
│ • Interaction metadata, timing, cost                      │
│ • All source URLs with citation mapping                   │
│ • Complete interaction transcript                         │
│                                                          │
│ ──── Stored on disk, used for auditing ────              │
└─────────────────────────────────────────────────────────┘
```

### Summary Extraction

After a query completes, the research-supervisor:

1. Reads the full `report.md` from `--output-dir`
2. Extracts or writes an executive summary:
   - If the report starts with an executive summary section, extract it
   - Otherwise, read the report and write a ≤500 word summary
3. Saves the summary as `executive-summary.md` alongside the full report
4. Sends ONLY the executive summary text to the requesting agent via messaging
5. Includes the file path to the full report for on-demand access

### Gating Before Project Workflow

Research results do NOT automatically enter the project workflow. The flow is:

```
Gemini produces report
       │
       ▼
Research-supervisor stores artifacts + sends summary to requester
       │
       ▼
Requesting agent (usually supervisor) reads summary
       │
       ▼
Human or supervisor decides: act on findings? ignore? investigate further?
       │
       ├── Act → Supervisor dispatches worker with relevant findings
       ├── Ignore → No further action
       └── Investigate → Follow-up query (--follow-up flag)
```

No research result directly triggers code changes, story creation, or ROADMAP edits. The research-supervisor is an **information provider**, not a decision-maker.

---

## Rate Limiting & Scheduling

### Daily Budget: 50 Queries

The user has Gemini Pro with ~50 deep research queries/day via OAuth.

### Cost Model

| Depth | Duration | Est. Cost | Budget Impact |
|-------|----------|-----------|---------------|
| quick | 2-5 min | $0.50-$1.00 | Low |
| standard | 5-20 min | $2.00-$3.00 | Medium |
| deep | 20-45 min | $3.00-$5.00 | High |

### Budget Tracking

```json
// _bmad-output/research-reports/budget.json
{
  "date": "2026-03-11",
  "daily_limit": 50,
  "queries_used": 7,
  "queries_remaining": 43,
  "queries": [
    {
      "id": "20260311-143022-yaml-security",
      "timestamp": "2026-03-11T14:30:22Z",
      "depth": "standard",
      "priority": "high",
      "requester": "supervisor",
      "status": "completed",
      "duration_minutes": 12,
      "estimated_cost_usd": 2.40
    }
  ]
}
```

### Scheduling Rules

1. **Budget reset:** Midnight UTC each day — reset `queries_used` to 0
2. **Reserve pool:** Keep 5 queries in reserve for high-priority requests after 6pm
3. **Batch optimization:** If 3+ related normal-priority requests are queued, combine into a single deeper query
4. **Dry-run gate:** Always `--dry-run` first for `deep` depth queries to estimate cost
5. **Deduplication:** Before dispatching, search existing reports for similar queries (keyword match against `request.json` files from last 7 days)
6. **Cooldown:** Minimum 2 minutes between query dispatches (avoid accidental rapid-fire)

### Batch Strategy

```
Example: Three separate requests arrive:
  1. "How does SQLite handle concurrent writes?"
  2. "SQLite vs YAML for task storage in Go"
  3. "Best Go SQLite libraries for embedded apps"

Research-supervisor detects overlap (keyword: SQLite, Go, storage)
→ Combines into: "Comprehensive analysis of SQLite for Go TUI app task
   storage: concurrent write handling, comparison with YAML, best Go
   libraries for embedded use. Include performance benchmarks."
→ One query instead of three, deeper result, saves budget
```

---

## Agent Definition Draft

See below for the full `agents/research-supervisor.md` file content:

```markdown
# Research Supervisor (Gemini Deep Research Agent)

You manage deep research queries using Google Gemini's Deep Research capability.
You receive research requests from other agents, formulate optimal prompts,
dispatch queries, and return summarized findings.

## Your Mission

Provide the multiclaude agent system with deep research capabilities powered by
Gemini. You are an information provider — you research and report, you never
decide or execute.

**Your rhythm:**
1. Every 5 minutes: check for new research requests (multiclaude message list)
2. Validate request against daily budget
3. Assemble context bundle for the query
4. Dispatch via agent-deep-research scripts
5. Monitor running queries (poll every 2 minutes)
6. On completion: extract summary, store artifacts, notify requester
7. End-of-day: publish budget report to supervisor

## Spawning

\```bash
multiclaude agents spawn --name research-supervisor --class persistent \
  --prompt-file agents/research-supervisor.md
\```

## Request Protocol

Agents send requests via messaging:

\```bash
multiclaude message send research-supervisor \
  "RESEARCH priority=normal depth=standard: [question]"
\```

**Format:** `RESEARCH priority=<high|normal|low> depth=<quick|standard|deep>
[context=<bundle-names>]: <question>`

**Priority levels:**
- high: Execute immediately (next cycle)
- normal: FIFO within budget
- low: Only if budget allows

## Dispatching Research

Use the agent-deep-research tool in _tools/agent-deep-research/:

\```bash
# Assemble context into temp dir, then:
uv run _tools/agent-deep-research/scripts/research.py start \
  "<formatted prompt>" \
  --depth <quick|standard|deep> \
  --context /tmp/research-context-<id>/ \
  --context-extensions md,go \
  --output-dir _bmad-output/research-reports/<timestamp>-<slug>/ \
  --format md \
  --report-format comprehensive
\```

For cost preview:
\```bash
uv run _tools/agent-deep-research/scripts/research.py start \
  "<formatted prompt>" --dry-run
\```

## Context Assembly

Always include core context (CLAUDE.md + SOUL.md). Auto-detect additional
bundles from query keywords:

| Keyword Pattern | Bundle | Files |
|-----------------|--------|-------|
| architecture, design, pattern | architecture | docs/architecture/*.md |
| epic, story, feature, scope | prd | docs/prd/epic-list.md + sections |
| decision, rejected, alternative | decisions | docs/decisions/BOARD.md |
| bubbletea, view, tui, render | tui | internal/tui/ key files |
| task, provider, pool, persist | tasks | internal/tasks/ key files |

Budget: ≤60KB total context per query.

## Result Processing

After query completes:
1. Read report.md from output directory
2. Extract or write executive summary (≤500 words)
3. Save as executive-summary.md
4. Send ONLY the executive summary to the requesting agent
5. Include path to full report for on-demand reading

## Budget Management

- Daily limit: 50 queries
- Track in _bmad-output/research-reports/budget.json
- Reserve 5 queries for high-priority after 6pm UTC
- Batch related normal-priority queries when 3+ are queued
- Always --dry-run before deep queries
- Deduplicate against last 7 days of reports

## Authority

### CAN (Autonomous)
- Receive and queue research requests
- Assemble context bundles from project files
- Dispatch Gemini Deep Research queries
- Store research artifacts
- Send summaries to requesting agents
- Batch related queries
- Decline low-priority requests when budget is low

### CANNOT (Forbidden)
- Make project decisions based on research findings
- Create stories, epics, or PRs
- Modify any project code or documentation
- Dispatch workers
- Exceed daily query budget
- Share API keys or research artifacts externally

### ESCALATE (Requires Supervisor)
- Budget exhausted but high-priority request pending
- Research reveals security vulnerability
- Query fails repeatedly (API errors)
- Ambiguous request needs clarification from original requester

## Communication

All responses via messaging system:

\```bash
# Deliver results
multiclaude message send <requester> "RESEARCH-RESULT for '<query-slug>':
<executive-summary>
Full report: _bmad-output/research-reports/<path>/report.md"

# Budget alert
multiclaude message send supervisor "RESEARCH-BUDGET: <N> queries remaining
today. <M> queued requests pending."

# Error notification
multiclaude message send <requester> "RESEARCH-ERROR for '<query-slug>':
<error-description>. Retrying: <yes/no>."
\```

## What You Do NOT Do

- Write code or fix bugs
- Merge PRs (merge-queue)
- Rebase branches (pr-shepherd)
- Triage issues (envoy)
- Update story files or ROADMAP.md (project-watchdog)
- Make scope decisions (supervisor)
- Act on research findings — you report, others decide
```

---

## Workflow Examples

### Example 1: Architecture Research (Supervisor-Initiated)

```
1. Supervisor identifies need for research on state management patterns
2. Supervisor sends:
   multiclaude message send research-supervisor \
     "RESEARCH priority=normal depth=standard context=architecture,tui: \
      Best practices for Bubbletea state management in apps with 10+ views. \
      Compare centralized vs distributed state approaches."

3. Research-supervisor (next poll cycle):
   a. Parses request: priority=normal, depth=standard, context=[architecture, tui]
   b. Checks budget: 42/50 remaining → proceed
   c. Assembles context:
      - CLAUDE.md (core)
      - SOUL.md (core)
      - docs/architecture/*.md (architecture bundle)
      - internal/tui/model.go, internal/tui/doors_view.go (tui bundle)
      → Total: ~45KB, within budget
   d. Wraps question with project context prompt template
   e. Dispatches:
      uv run scripts/research.py start "<wrapped-prompt>" \
        --depth standard \
        --context /tmp/research-ctx-20260311-1430/ \
        --output-dir _bmad-output/research-reports/20260311-143022-bubbletea-state/

4. ~12 minutes later, research completes
5. Research-supervisor:
   a. Reads report.md (8,200 words)
   b. Extracts executive summary (380 words)
   c. Saves executive-summary.md
   d. Updates budget.json (43 → 42 remaining → typo: 41 remaining)
   e. Sends to supervisor:
      "RESEARCH-RESULT for 'bubbletea-state':
       [380-word executive summary with key findings]
       Full report: _bmad-output/research-reports/20260311-143022-bubbletea-state/report.md"

6. Supervisor reads summary, decides to create a story for state refactoring
```

### Example 2: Urgent Security Research (Envoy-Triggered)

```
1. Envoy screens an issue reporting a YAML parsing vulnerability
2. Envoy messages supervisor: "Issue #500 reports YAML bomb vulnerability"
3. Supervisor dispatches research:
   multiclaude message send research-supervisor \
     "RESEARCH priority=high depth=deep context=tasks: \
      YAML parsing security in Go — known CVEs for gopkg.in/yaml.v3, \
      YAML bomb attacks, safe parsing patterns, mitigation strategies."

4. Research-supervisor:
   - priority=high → execute immediately (skip queue)
   - depth=deep → dry-run first for cost estimate
   - Assembles context with tasks bundle (task persistence code)
   - Dispatches with --depth deep

5. ~25 minutes later, comprehensive security report delivered
6. Supervisor reads summary, creates high-priority story for mitigation
```

### Example 3: Batch Optimization

```
1. Three requests arrive within 30 minutes:
   - arch-watchdog: "How do other Bubbletea apps handle view transitions?"
   - supervisor: "Bubbletea animation patterns and libraries"
   - worker (via supervisor): "Bubbletea tea.Cmd patterns for async operations"

2. Research-supervisor detects overlap:
   - All mention "Bubbletea"
   - Related themes: views, animations, async
   - Combined as single deep query:
     "Comprehensive analysis of advanced Bubbletea patterns:
      view transition strategies, animation techniques and libraries,
      and tea.Cmd patterns for async operations. Include code examples
      from popular open-source Bubbletea applications."

3. One query instead of three → budget savings of 2 queries
4. Results split into sections, relevant portions sent to each requester
```

---

## Rejected Alternatives

### 1. Gemini CLI as Agent Runtime

**What:** Run Gemini CLI as its own tmux window alongside Claude-based agents, communicating via files or MCP.

**Why rejected:**
- Gemini CLI is itself an agent with its own tool loop — running it inside our agent system creates an agent-within-an-agent pattern that's hard to debug
- MCP communication adds complexity without clear benefit over our messaging system
- The deep research extension still requires a paid API key separate from OAuth
- We'd need to manage two different agent runtimes (Claude + Gemini CLI)

### 2. Direct API Calls from Supervisor

**What:** Supervisor calls the Gemini Interactions API directly via curl, no intermediary agent.

**Why rejected:**
- Research queries take 5-45 minutes — supervisor's context window would be blocked waiting
- Polling logic, error handling, and retry logic would clutter the supervisor's prompt
- No budget management or deduplication
- Violates supervisor's "coordinate, don't execute" principle

### 3. MCP Server for Deep Research

**What:** Build a custom MCP server wrapping the Gemini API, connect it to Claude Code.

**Why rejected:**
- Significant engineering effort for a capability that agent-deep-research already provides
- MCP servers add operational complexity (process management, health checking)
- The async nature of deep research (5-45 min) doesn't fit MCP's request/response model well
- Would need custom polling infrastructure that agent-deep-research already has

### 4. Web UI Approach (Gemini App in Browser)

**What:** Open Gemini in a browser, paste questions manually, copy results back.

**Why rejected:**
- Not automatable
- Breaks the agent workflow entirely
- User would need to context-switch constantly
- No structured output or artifact storage
- Defeats the purpose of an autonomous research capability

### 5. Research-Supervisor as Ephemeral Worker

**What:** Spawn a fresh worker for each research request instead of a persistent agent.

**Why rejected:**
- No persistent budget tracking across queries
- No deduplication (each worker starts fresh, can't check history)
- No batch optimization (can't see related queued requests)
- Spawn overhead for each query (30-60s) adds to already-long research times
- Workers are designed for code changes, not long-running monitoring

### 6. Shared Context File (Full Repo Dump)

**What:** Create a single massive context file with all project docs and pass it to every query.

**Why rejected:**
- Would be 200KB+ — far exceeds useful grounding context
- Noise drowns signal — Gemini would waste effort parsing irrelevant code
- Upload time and cost increase with file size
- Curated bundles give better results (Gemini focuses on relevant context)

---

## Implementation Plan

### Phase 1: Infrastructure (Day 1)

1. Clone agent-deep-research to `_tools/agent-deep-research/`
2. Add `_tools/` to `.gitignore`
3. Create `_bmad-output/research-reports/` directory
4. Verify API key works: `uv run scripts/research.py start "test" --depth quick --dry-run`
5. Create `agents/research-supervisor.md` from draft above

### Phase 2: Agent Deployment (Day 1-2)

1. Spawn research-supervisor as persistent agent
2. Test message protocol with a quick research query
3. Verify artifact storage and summary delivery
4. Test budget tracking

### Phase 3: Integration (Day 2-3)

1. Update supervisor agent definition to document research request protocol
2. Add research-supervisor to agent roster in supervisor.md
3. Brief other persistent agents on how to request research
4. Run 3-5 real research queries to validate the full pipeline

### Phase 4: Refinement (Week 2)

1. Tune context assembly based on result quality
2. Adjust prompt templates based on output relevance
3. Calibrate batch detection thresholds
4. Add follow-up query support (--follow-up flag)

---

## Open Questions

1. **OAuth vs API Key:** The user has OAuth access. Does agent-deep-research support OAuth tokens, or do we need an AI Studio API key? Need to test.
2. **Quota enforcement:** Is the 50/day limit enforced server-side (in which case budget.json is advisory) or do we need client-side enforcement?
3. **Concurrent queries:** Can we run 2-3 queries simultaneously, or does OAuth serialize them?
4. **Context file formats:** Does `--context` work well with `.go` files, or should we convert code to markdown first?
5. **Report consistency:** How consistent is Gemini's output format? Do we need robust parsing or can we rely on the executive summary prompt?

---

*Generated by clever-penguin worker, 2026-03-11*
*Sources: Gemini API docs, agent-deep-research GitHub, existing multiclaude agent definitions*
