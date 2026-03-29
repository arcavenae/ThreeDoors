# Perplexity API Integration & Research Supervisor Architecture

**Date:** 2026-03-29
**Type:** Research / Architecture Design
**Status:** Complete
**Prior Art:** [Gemini Research Supervisor Design](gemini-research-supervisor-design.md) (D-154/D-164)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Perplexity API Capabilities](#perplexity-api-capabilities)
3. [Integration Options Analysis](#integration-options-analysis)
4. [API Key Management Architecture](#api-key-management-architecture)
5. [Research Supervisor Agent Design](#research-supervisor-agent-design)
6. [Research Coordination Protocol](#research-coordination-protocol)
7. [Dark Factory Enhancement](#dark-factory-enhancement)
8. [Implementation Plan](#implementation-plan)
9. [Decisions Summary](#decisions-summary)
10. [Open Questions](#open-questions)

---

## Executive Summary

This document designs the integration of Perplexity's Sonar API into the multiclaude agent system and proposes a research-supervisor persistent agent to coordinate AI-powered web research across all agents.

**Key findings:**
- Perplexity offers an **official MCP server** (`@perplexity-ai/mcp-server`) with 4 tools (search, ask, research, reason) — this is the integration path of least resistance
- The MCP server installs via a single `claude mcp add` command and works natively with Claude Code
- API authentication is Bearer token via `PERPLEXITY_API_KEY` environment variable
- Pricing is usage-based: Sonar at $1/$1 per M tokens + $5/1K requests; Sonar Pro at $3/$15 per M tokens
- **Recommended architecture:** MCP server configured at project level in `.claude/settings.json`, with research-supervisor as the primary consumer and gatekeeper
- This supersedes the Gemini-only research design (D-164) by adding Perplexity as a **complementary** tool — Perplexity for web search/citations, Gemini for deep multi-step research

**What this enables:**
- Any agent can perform web-grounded research with citations
- Research-supervisor deduplicates queries, manages budget, caches results
- The dark factory gains "thinking" capabilities — research before coding
- Citation provenance maintained for all research-informed decisions

---

## Perplexity API Capabilities

### Available Models

| Model | Use Case | Input Cost | Output Cost | Request Fee |
|-------|----------|-----------|-------------|-------------|
| `sonar` | Quick web search, fact-checking | $1/1M | $1/1M | $5-12/1K |
| `sonar-pro` | Multi-step queries, detailed analysis | $3/1M | $15/1M | $6-14/1K |
| `sonar-reasoning-pro` | Complex analytical reasoning | $2/1M | $8/1M | $6-14/1K |
| `sonar-deep-research` | Comprehensive multi-step research | $2/1M | $8/1M | $5/1K + $2/1M citations |

### API Details

- **Endpoint:** `POST https://api.perplexity.ai/v1/sonar`
- **Auth:** Bearer token via `Authorization` header
- **Format:** OpenAI-compatible (works with OpenAI SDK by swapping base URL)
- **Streaming:** Supported via `stream: true`
- **Citations:** Included in responses with URLs and metadata
- **Search options:** Configurable source filters

### Official MCP Server

Perplexity provides `@perplexity-ai/mcp-server` with 4 tools:

| Tool | Model Used | Purpose |
|------|-----------|---------|
| `perplexity_search` | Search API | Direct web search, ranked results with metadata |
| `perplexity_ask` | `sonar-pro` | Conversational Q&A with web grounding |
| `perplexity_research` | `sonar-deep-research` | Comprehensive research reports with citations |
| `perplexity_reason` | `sonar-reasoning-pro` | Complex reasoning and problem-solving |

**Installation:**
```bash
claude mcp add perplexity --env PERPLEXITY_API_KEY="pplx-xxx" -- npx -y @perplexity-ai/mcp-server
```

**Configuration (`.claude/settings.json`):**
```json
{
  "mcpServers": {
    "perplexity": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@perplexity-ai/mcp-server"],
      "env": {
        "PERPLEXITY_API_KEY": "${PERPLEXITY_API_KEY}"
      }
    }
  }
}
```

**Environment variables:**
- `PERPLEXITY_API_KEY` (required) — API key from api.perplexity.ai
- `PERPLEXITY_TIMEOUT_MS` (optional, default 5 min)
- `PERPLEXITY_BASE_URL` (optional, default https://api.perplexity.ai)
- `PERPLEXITY_LOG_LEVEL` (optional: DEBUG|INFO|WARN|ERROR)

### Rate Limits

Rate limits based on RPM (Requests Per Minute), TPD (Tokens Per Day), and bandwidth. Specific limits vary by usage tier (high/medium/low). Perplexity throttles/queues on exceeding limits.

**Pro subscribers:** $5/month automatic API credit included.

---

## Integration Options Analysis

### Option A: MCP Server at Project Level (RECOMMENDED)

```
┌──────────────────────────────────────────────────────────┐
│                Claude Code Agent Session                   │
│                                                            │
│  .claude/settings.json → mcpServers.perplexity             │
│                    ↓                                       │
│  ┌─────────────────────────────┐                          │
│  │ @perplexity-ai/mcp-server   │                          │
│  │ (stdio, auto-started)       │                          │
│  │                             │                          │
│  │ Tools:                      │                          │
│  │  perplexity_search          │                          │
│  │  perplexity_ask             │                          │
│  │  perplexity_research        │                          │
│  │  perplexity_reason          │                          │
│  └──────────┬──────────────────┘                          │
│             │                                              │
└─────────────┼──────────────────────────────────────────────┘
              │ HTTPS (Bearer token)
              ▼
    ┌────────────────────┐
    │ Perplexity Sonar   │
    │ API (cloud)        │
    └────────────────────┘
```

- **Pros:**
  - Official first-party MCP server — maintained by Perplexity
  - Single `claude mcp add` command installs it
  - All Claude Code agents (workers, persistent agents) get access automatically
  - 4 purpose-built tools map perfectly to research use cases
  - OpenAI-compatible — familiar interface
  - `strip_thinking` parameter conserves context tokens
  - No custom code needed
- **Cons:**
  - Every agent has direct access (no centralized gatekeeping without research-supervisor)
  - No built-in budget tracking or deduplication
  - npx requires Node.js (already available via Homebrew)

### Option B: Direct API Calls via curl

```bash
curl -X POST https://api.perplexity.ai/v1/sonar \
  -H "Authorization: Bearer $PERPLEXITY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "sonar-pro", "messages": [{"role": "user", "content": "..."}]}'
```

- **Pros:** No dependencies beyond curl, full control
- **Cons:** Must parse JSON responses manually, no streaming, no MCP integration, reinventing the wheel
- **Verdict:** Rejected — MCP server provides all this for free

### Option C: Custom MCP Server Wrapping Perplexity

Build our own MCP server that wraps Perplexity API with budget tracking, caching, deduplication.

- **Pros:** Full control over budget, caching, rate limiting
- **Cons:** Significant development effort, must maintain as Perplexity API evolves, premature optimization
- **Verdict:** Rejected for Phase 1 — revisit if budget/deduplication becomes a real problem

### Option D: OpenAI SDK with Perplexity Base URL

Since Perplexity is OpenAI-compatible, use the OpenAI Go/Python SDK with a swapped base URL.

- **Pros:** Familiar API, could share code with other LLM integrations
- **Cons:** Loses MCP integration, must build tool abstractions, more code to maintain
- **Verdict:** Rejected — MCP server is strictly better for agent integration

**Decision:** Option A. The official MCP server is the path of least resistance, provides all needed tools, and integrates natively with Claude Code. Budget tracking and deduplication are handled by the research-supervisor agent (a coordination layer, not an infrastructure layer).

---

## API Key Management Architecture

### Options Evaluated

| Strategy | How It Works | Pros | Cons |
|----------|-------------|------|------|
| **A: Project-level MCP config** | Key in `.claude/settings.json` env block, referencing `$PERPLEXITY_API_KEY` from shell env | All agents inherit; single config point; key not in repo | Requires env var set before Claude starts |
| **B: Per-agent key injection** | Each agent spawned with its own key via env var at spawn time | Fine-grained control, per-agent budgets | Complex management, key proliferation |
| **C: Daemon-managed shared service** | Daemon proxies all Perplexity requests, holds key centrally | Best budget control, single point of rate limiting | Requires daemon code changes, adds latency |
| **D: Supervisor holds key, delegates** | Only research-supervisor has the MCP server; others request via messaging | Centralized gatekeeping | Bottleneck; workers can't do quick lookups |

**Recommended: Strategy A (project-level MCP) + research-supervisor as soft gatekeeper.**

Rationale:
- Project-level MCP config means all agents CAN access Perplexity directly
- Research-supervisor acts as the **preferred** path (protocol convention, not enforcement)
- Quick lookups (workers verifying a library exists, envoy checking a CVE) can go direct
- Complex research (multi-step, needs deduplication, should be cached) routes through research-supervisor
- This matches the existing pattern: all agents CAN push to git, but convention routes through merge-queue

**Key management:**
```bash
# User sets in their shell profile (~/.zshrc)
export PERPLEXITY_API_KEY="pplx-xxxxxxxxxxxxxxxx"

# .claude/settings.json references it (not committed to repo)
# OR: use .claude/settings.local.json for local-only config
```

**Security:**
- API key NEVER committed to repo
- `.claude/settings.json` uses `${PERPLEXITY_API_KEY}` environment variable reference
- Workers inherit the key from the shell environment at spawn time
- Cost visibility: Perplexity dashboard shows usage per API key

---

## Research Supervisor Agent Design

### Role & Responsibilities

The research-supervisor is a **persistent agent** that:

1. **Receives** research requests via multiclaude messaging
2. **Deduplicates** against recent research results (avoids re-querying)
3. **Routes** to the appropriate Perplexity tool based on query complexity
4. **Caches** results in `_bmad-output/research-reports/` with citations
5. **Distributes** findings back to requesting agents with summaries
6. **Tracks** daily query budget and usage
7. **Maintains** a research knowledge base for the project

### When to Use Research-Supervisor vs Direct Perplexity

| Scenario | Route | Why |
|----------|-------|-----|
| Worker needs to verify a Go library exists | Direct `perplexity_search` | Quick, no caching value |
| Envoy checking a CVE for issue triage | Direct `perplexity_ask` | Time-sensitive, simple |
| Architecture decision: compare 3 approaches | Research-supervisor | Multi-step, should be cached, deduplicated |
| Dark factory research before implementation | Research-supervisor | Complex, results feed into planning |
| Understanding a new technology for epic planning | Research-supervisor | Results should be stored as artifacts |

### Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    multiclaude Agent System                  │
│                                                              │
│  ┌───────────┐  ┌──────────┐  ┌───────────┐  ┌──────────┐ │
│  │ Supervisor │  │ Workers  │  │ Architect │  │  Envoy   │ │
│  │           │  │ (ephem)  │  │ Watchdog  │  │          │ │
│  └─────┬─────┘  └────┬─────┘  └─────┬─────┘  └────┬─────┘ │
│        │              │              │              │        │
│        │    multiclaude message send "RESEARCH: ..." │        │
│        ▼              ▼              ▼              ▼        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Research Supervisor (persistent)             │   │
│  │                                                       │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐  │   │
│  │  │ Dedup    │ │ Router   │ │ Cache    │ │ Budget │  │   │
│  │  │ Engine   │ │ (model   │ │ Manager  │ │Tracker │  │   │
│  │  │          │ │ selector)│ │          │ │        │  │   │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └───┬────┘  │   │
│  │       │             │            │            │       │   │
│  │       ▼             ▼            ▼            ▼       │   │
│  │  ┌──────────────────────────────────────────────┐    │   │
│  │  │        MCP: @perplexity-ai/mcp-server         │    │   │
│  │  │  perplexity_search | _ask | _research | _reason│    │   │
│  │  └──────────────────┬───────────────────────────┘    │   │
│  └─────────────────────┼────────────────────────────────┘   │
│                        │                                     │
└────────────────────────┼─────────────────────────────────────┘
                         │ HTTPS
                         ▼
               ┌────────────────────┐
               │ Perplexity Sonar   │
               │ API (cloud)        │
               └────────────────────┘
```

### Persistent vs Ephemeral

**Recommendation: Persistent agent.**

Rationale:
- Research requests come at unpredictable times from any agent
- Deduplication requires memory of recent queries (in-process state)
- Budget tracking needs continuity across requests
- Cache management is ongoing
- Matches the pattern of envoy (always listening for messages)

### Authority Boundaries

| CAN (Autonomous) | CANNOT | ESCALATE |
|---|---|---|
| Execute research queries via Perplexity MCP tools | Make architectural decisions based on findings | Research results that contradict existing decisions (D-xxx) |
| Cache and distribute results | Create/modify code or story files | Budget approaching daily limit |
| Choose which Perplexity model to use | Update BOARD.md directly | Queries that could expose sensitive project info |
| Deduplicate identical/similar queries | Merge PRs or create branches | Requests that seem to duplicate ongoing research |
| Prioritize query queue | Override agent's explicit model choice | |
| Report findings with citations | | |

### Request Protocol

```bash
# Standard research request (routes through research-supervisor)
multiclaude message send research-supervisor "RESEARCH priority=normal tool=auto: What are best practices for YAML schema migration in Go CLIs?"

# Specify tool explicitly
multiclaude message send research-supervisor "RESEARCH priority=high tool=research: Comprehensive analysis of Bubbletea vs Brick vs tview for complex TUI layouts"

# Quick lookup (may bypass research-supervisor)
multiclaude message send research-supervisor "RESEARCH priority=low tool=search: Does charmbracelet/huh support dynamic form fields?"
```

**Message format:**
```
RESEARCH priority=<high|normal|low> tool=<auto|search|ask|research|reason> [cache=<yes|no>]: <question>
```

### Routing Logic (tool=auto)

```
IF question is a factual lookup → perplexity_search
IF question needs conversational answer → perplexity_ask
IF question requires deep analysis (>3 aspects) → perplexity_research
IF question requires logical reasoning/comparison → perplexity_reason
```

### Deduplication Strategy

1. **Exact match:** Hash the question, check against recent query log
2. **Semantic similarity:** If question is substantially similar to a cached result (<24h old), return cached result with "from cache" flag
3. **Partial overlap:** If cached result partially answers the question, supplement with a focused follow-up query

Cache stored in `_bmad-output/research-reports/` as individual markdown files with metadata YAML frontmatter.

### Cache Structure

```
_bmad-output/
└── research-reports/
    ├── perplexity/
    │   ├── 20260329-143000-yaml-migration-go.md     # Individual result
    │   ├── 20260329-150000-bubbletea-comparison.md
    │   └── budget.json                               # Usage tracking
    └── gemini/                                        # Existing Gemini results
        └── ...
```

**Result file format:**
```markdown
---
query: "What are best practices for YAML schema migration in Go CLIs?"
tool: perplexity_ask
model: sonar-pro
requested_by: arch-watchdog
timestamp: 2026-03-29T14:30:00Z
tokens_used: { input: 450, output: 2100 }
cost_estimate: "$0.033"
citations: 8
cache_hit: false
---

# YAML Schema Migration Best Practices in Go

[research content with inline citations]

## Sources
1. [Title](url) — snippet
2. [Title](url) — snippet
...
```

### Budget Tracking

```json
{
  "date": "2026-03-29",
  "queries": 12,
  "tokens": { "input": 15000, "output": 42000 },
  "estimated_cost": "$0.85",
  "by_tool": {
    "search": { "queries": 5, "cost": "$0.025" },
    "ask": { "queries": 4, "cost": "$0.12" },
    "research": { "queries": 2, "cost": "$0.55" },
    "reason": { "queries": 1, "cost": "$0.15" }
  },
  "daily_budget": "$5.00",
  "remaining": "$4.15"
}
```

**Budget thresholds:**
- Green: <60% of daily budget
- Yellow: 60-80% — reduce non-essential queries
- Red: >80% — high-priority queries only
- Hard stop: 100% — all queries rejected until next day

---

## Research Coordination Protocol

### Flow: Agent Needs Research

```
1. Agent identifies need for web research
2. Agent sends RESEARCH message to research-supervisor
3. Research-supervisor:
   a. Checks dedup cache — if hit, returns cached result immediately
   b. Checks budget — if over threshold, queues or rejects
   c. Routes to appropriate Perplexity tool
   d. Stores result in research-reports/
   e. Sends executive summary back to requesting agent
4. Requesting agent incorporates findings
```

### Flow: Research Informs Planning

```
1. Supervisor identifies need for research before story creation
2. Supervisor sends RESEARCH message with priority=high
3. Research-supervisor executes, stores artifact
4. Supervisor reads artifact, makes planning decisions
5. Research artifact linked in BOARD.md or story file
```

### Interaction with Existing Agents

| Agent | How They Use Research-Supervisor |
|-------|------|
| **Supervisor** | Requests research before dispatching work; informs story planning |
| **Workers** | Quick technology lookups; verify library/API details during implementation |
| **Arch-watchdog** | Research emerging patterns; validate architecture decisions |
| **Envoy** | CVE lookups for issue triage; verify external tool status |
| **Merge-queue** | Rarely — might verify a dependency security advisory |
| **pr-shepherd** | Rarely — might research a merge conflict resolution approach |
| **Retrospector** | Research best practices to compare against project patterns |
| **Project-watchdog** | Research for story planning and roadmap decisions |

### Complementary Research Tools

Perplexity and Gemini serve different purposes:

| Capability | Perplexity (NEW) | Gemini Deep Research (D-164) |
|------------|-----------------|--------------------------|
| **Speed** | Seconds (sonar), minutes (deep-research) | Minutes to hours |
| **Web search** | Native, every query is web-grounded | GoogleSearch grounding, but primarily synthesis |
| **Citations** | First-class, always included | Available but less structured |
| **Deep multi-step research** | `sonar-deep-research` for complex queries | Gemini's strength — autonomous multi-step investigation |
| **Reasoning** | `sonar-reasoning-pro` for analytical queries | Gemini Thinking for complex reasoning |
| **Cost** | $1-15/1M tokens + request fees | Free tier (50 Pro/day via OAuth) |
| **Integration** | MCP server (native Claude Code) | Gemini CLI (`gemini -p`) |
| **Best for** | Quick web-grounded answers with citations | Deep autonomous research projects |

**Recommendation:** Use both. Research-supervisor routes to the appropriate backend:
- Quick web lookups → Perplexity `perplexity_search` / `perplexity_ask`
- Technology comparisons → Perplexity `perplexity_reason`
- Comprehensive research → Perplexity `perplexity_research` OR Gemini Deep Research (based on depth needed)
- Multi-hour autonomous investigation → Gemini Deep Research

---

## Dark Factory Enhancement

### Current Dark Factory Capabilities (from R-003)

The dark factory research (R-003) established:
- L0-L4 autonomy spectrum
- Separate-repo architecture with golden repo
- Gallery model with 3-5 variants per generation
- Dispose-and-rebuild cycles
- AI judges panel for variant evaluation

### How Perplexity Elevates the Dark Factory

**Current state:** Dark factory agents can only code using their training knowledge + project context. They're "coding factories" — they execute specs but can't research novel approaches.

**With Perplexity:** Agents gain real-time web access. This transforms the dark factory from a "coding factory" to a "thinking factory":

1. **Research-before-coding:** Before implementing a story, the dark factory variant researches current best practices, recent library updates, known pitfalls
2. **Citation-backed decisions:** Architecture decisions in variant code include links to sources
3. **Novel approach discovery:** Variants can discover approaches not in Claude's training data
4. **Up-to-date dependency selection:** Variants verify dependencies are current, maintained, and secure
5. **Competitive analysis:** Factory can research how other projects solve similar problems

### Research-Driven Development Flow

```
┌──────────────────────────────────────────────────────────┐
│                Dark Factory Variant Run                    │
│                                                            │
│  1. Read spec + project context                           │
│  2. Research phase (NEW):                                 │
│     a. Research current best practices for each AC        │
│     b. Check for recent library updates/deprecations      │
│     c. Verify approach against current standards          │
│     d. Store findings as variant-local research artifacts │
│  3. Design phase (informed by research)                   │
│  4. Implement phase                                       │
│  5. Test phase                                            │
│  6. Self-review (cross-check against research findings)   │
│                                                            │
│  Output: code + tests + research provenance               │
└──────────────────────────────────────────────────────────┘
```

### Research Provenance in Dark Factory

Each dark factory variant's output should include:
```
variant-output/
├── code/                    # Implementation
├── tests/                   # Test suite
├── research-log.md          # What was researched, findings, how they influenced decisions
├── citations.json           # All Perplexity citations used
└── variant-metadata.json    # Includes research_queries_count, research_cost
```

This enables the AI judges panel to evaluate not just code quality but research quality — did the variant explore the right questions? Did it incorporate findings?

---

## Implementation Plan

### Phase 1: MCP Server Integration (Immediate)

**Effort:** ~30 minutes
**Prerequisite:** Perplexity API key

1. User creates Perplexity API account at https://api.perplexity.ai
2. Generate API key (starts with `pplx-`)
3. Add to shell profile: `export PERPLEXITY_API_KEY="pplx-xxx"`
4. Install MCP server:
   ```bash
   claude mcp add perplexity --env PERPLEXITY_API_KEY="$PERPLEXITY_API_KEY" -- npx -y @perplexity-ai/mcp-server
   ```
5. Verify: all Claude Code sessions now have `perplexity_search`, `perplexity_ask`, `perplexity_research`, `perplexity_reason` tools available

**Gate:** Any agent can call Perplexity tools. No research-supervisor yet — this is "direct access" mode.

### Phase 2: Research Supervisor Agent Definition (Story-sized)

**Effort:** One story (~2-4 hours)

1. Create `agents/research-supervisor.md` agent definition
2. Define polling loop (check messages every 5 min)
3. Implement routing logic (tool=auto selection)
4. Create `_bmad-output/research-reports/perplexity/` directory structure
5. Add HEARTBEAT cron to supervisor startup checklist
6. Spawn and validate

### Phase 3: Deduplication & Caching (Story-sized)

**Effort:** One story (~2-4 hours)

1. Implement query hash cache
2. Build result storage in markdown with YAML frontmatter
3. Cache lookup on new requests
4. TTL-based cache expiry (24 hours default)

### Phase 4: Budget Tracking & Knowledge Base (Story-sized)

**Effort:** One story (~2-4 hours)

1. Implement `budget.json` daily tracking
2. Add threshold-based query throttling
3. Build research knowledge base index (searchable catalog of past research)
4. Integration with BOARD.md (research-supervisor can suggest entries)

### Phase 5: Dark Factory Research Integration (Epic-sized)

**Effort:** Multiple stories, dependent on dark factory implementation

1. Research phase in variant execution pipeline
2. Citation provenance in variant output
3. Research quality evaluation by AI judges

### Security Considerations

- **API key rotation:** Perplexity keys can be regenerated at any time; update shell profile
- **Cost controls:** Budget tracking in Phase 4; Perplexity dashboard for real-time monitoring
- **Rate limiting:** Perplexity enforces per-key limits; budget tracker prevents runaway costs
- **Data privacy:** Queries sent to Perplexity cloud — don't include sensitive project secrets in research queries
- **Key exposure:** Never commit `PERPLEXITY_API_KEY` to repo; use environment variables only

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| 1 | Integration method | Official Perplexity MCP server (`@perplexity-ai/mcp-server`) | Direct curl API calls (reinventing the wheel), Custom MCP server wrapper (premature optimization), OpenAI SDK with swapped base URL (loses MCP integration) | MCP server provides 4 purpose-built tools, native Claude Code integration, zero custom code |
| 2 | Key management | Project-level MCP config with `$PERPLEXITY_API_KEY` env var; all agents inherit | Per-agent key injection (complexity), Daemon-managed proxy (requires daemon code changes), Supervisor-only access (bottleneck) | Matches existing pattern (all agents CAN git push, convention routes through merge-queue); research-supervisor as soft gatekeeper |
| 3 | Research-supervisor type | Persistent agent (always running) | Ephemeral/on-demand (loses dedup state, no continuity), Cron job (research needs are unpredictable) | Requests come unpredictably; dedup/cache/budget require continuity; matches envoy pattern |
| 4 | Authority model | Report findings only — never make decisions based on research | Decision-making authority based on research | Research-supervisor is an information service; decisions belong to the requesting agent or supervisor |
| 5 | Perplexity + Gemini relationship | Complementary — Perplexity for web search/citations, Gemini for deep autonomous research | Replace Gemini with Perplexity (different strengths), Use only one (limits capabilities) | Different tools for different jobs; research-supervisor routes to the best backend per query |
| 6 | Direct access policy | All agents MAY use Perplexity directly for quick lookups; research-supervisor PREFERRED for complex research | Strictly centralized (bottleneck), Fully decentralized (no dedup/budget) | Convention over enforcement; quick CVE checks shouldn't require message round-trips |
| 7 | Cache storage | Markdown files with YAML frontmatter in `_bmad-output/research-reports/perplexity/` | Database (over-engineering), In-memory only (lost on restart), JSON (less readable) | Matches project patterns (markdown + YAML); human-readable; git-trackable if desired |

---

## Open Questions

| ID | Question | Context | Recommended Default |
|----|----------|---------|-------------------|
| OQ-1 | Daily budget cap for Perplexity API? | $5/day seems reasonable for development; production dark factory may need more | Start at $5/day, adjust based on actual usage |
| OQ-2 | Should research results be committed to repo or gitignored? | Cached results are ephemeral but research artifacts have long-term value | Gitignore cache (`research-reports/perplexity/`); manually commit valuable artifacts to `_bmad-output/planning-artifacts/` |
| OQ-3 | Should workers be allowed to call `perplexity_research` (deep research) directly, or only via research-supervisor? | Deep research is expensive ($0.50+ per query) and benefits most from deduplication | Workers use `perplexity_search` and `perplexity_ask` directly; route `perplexity_research` through research-supervisor |
| OQ-4 | Perplexity Pro subscription or pay-as-you-go only? | Pro is $20/month but includes $5 API credit and higher rate limits | Start pay-as-you-go; upgrade to Pro if monthly spend exceeds $15 |
| OQ-5 | Should research-supervisor have its own HEARTBEAT cron, or respond to supervisor polling? | Other persistent agents use HEARTBEAT crons | Add HEARTBEAT cron at prime interval (e.g., `*/17 * * * *`) |
| OQ-6 | How should the Perplexity MCP server be installed — project-level `.claude/settings.json` or user-level `~/.claude/settings.json`? | Project-level keeps it scoped to ThreeDoors; user-level makes it available everywhere | Project-level for now (ThreeDoors-specific budget tracking); user can add globally if desired |
