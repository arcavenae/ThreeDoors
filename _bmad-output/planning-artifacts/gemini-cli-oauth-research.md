# Gemini CLI OAuth Rearchitecture — Research & Plan

**Date:** 2026-03-11
**Type:** Research / Architecture Rework
**Requested by:** User (supervisor task)
**Status:** Complete
**Supersedes:** D-154 (agent-deep-research as execution layer)

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Current Architecture (Epic 54)](#current-architecture-epic-54)
3. [Gemini CLI Capabilities](#gemini-cli-capabilities)
4. [Proposed Architecture: Gemini CLI + OAuth](#proposed-architecture)
5. [Deep Research Options](#deep-research-options)
6. [Migration Plan](#migration-plan)
7. [Risks & Mitigations](#risks--mitigations)
8. [Rejected Approaches](#rejected-approaches)
9. [Open Questions](#open-questions)

---

## Problem Statement

Epic 54's current design (D-154) chose `24601/agent-deep-research` — a third-party Python tool using `uv run` and the Google Gemini Interactions API with a **paid API key**. The user wants to rearchitect to:

1. **Wrap the official Gemini CLI** (`@google/gemini-cli`) instead of Python scripts
2. **Use OAuth authentication** — user already has a Google account, no API key needed
3. **Target Gemini Pro free tier** — 50 requests/day for Pro, 1,000/day for Flash
4. **Eliminate Python dependency** — no `uv`, no `pip`, no Python SDK

### Why Change?

| Current (D-154) | Proposed |
|---|---|
| Python + `uv run` + google-genai SDK | Node.js Gemini CLI (npm/npx) |
| Requires paid Gemini API key | OAuth with Google account (free) |
| Third-party tool (24601/agent-deep-research) | Official Google CLI (`google-gemini/gemini-cli`) |
| Deep Research Interactions API ($2-5/query) | Standard Gemini Pro queries (free tier) |
| 50 queries/day (paid budget) | 50 req/day Pro free tier, 1,000/day Flash |

---

## Current Architecture (Epic 54)

The existing design (in `gemini-research-supervisor-design.md`) specifies:

- **Execution layer:** `24601/agent-deep-research` (MIT, Python, uses `uv run`)
- **Authentication:** `GEMINI_DEEP_RESEARCH_API_KEY` environment variable (paid)
- **API:** Gemini Interactions API (`deep-research-pro-preview-12-2025` model)
- **Features used:** `--context` (RAG grounding), `--dry-run`, `--depth`, `--output-dir`
- **5 stories:** 54.1 (agent def), 54.2 (CLI setup), 54.3 (context packaging), 54.4 (result shielding), 54.5 (rate limiting)

### What Must Be Preserved

These design elements from the existing Epic 54 are still valuable:
- Research-supervisor as a persistent agent (messaging-based request/response)
- Context packaging strategy (8 bundles: core, architecture, PRD, stories, decisions, code-sample, TUI, tasks)
- Result shielding (executive summary → detailed → raw)
- Budget tracking and priority queue
- Artifact storage at `_bmad-output/research-reports/`

---

## Gemini CLI Capabilities

### Overview

The [Gemini CLI](https://github.com/google-gemini/gemini-cli) (`@google/gemini-cli`) is Google's official open-source AI agent for the terminal. Current version: 0.32.1.

**Installation:**
```bash
npm install -g @google/gemini-cli
# or without installation:
npx @google/gemini-cli
```

### Authentication Methods

| Method | Setup | Rate Limits | API Key Required? |
|---|---|---|---|
| **Google Account OAuth** | Run `gemini`, browser opens, sign in | 60 RPM, 1,000 RPD | No |
| Gemini API Key | `GEMINI_API_KEY` env var | Varies by tier | Yes |
| Vertex AI (ADC) | `gcloud auth application-default login` | Enterprise limits | No (uses ADC) |
| Service Account | `GOOGLE_APPLICATION_CREDENTIALS` env var | Enterprise limits | No (uses SA key) |

**OAuth is the default and simplest method.** First run opens browser, user signs in, token is cached locally and refreshed automatically. No API keys, no environment variables, no project configuration.

### Headless/Programmatic Mode

The CLI supports full non-interactive usage:

```bash
# Basic prompt execution
gemini -p "Analyze the architecture of this Go project"

# With JSON output (structured, parseable)
gemini -p "Summarize this codebase" --output-format json

# With streaming JSON (JSONL events)
gemini -p "Review this code" --output-format stream-json

# Pipe input
cat docs/architecture/*.md | gemini -p "Summarize these architecture docs"

# Custom system prompt
GEMINI_SYSTEM_MD=/path/to/custom-system.md gemini -p "Research query"

# Include additional directories for context
gemini -p "Analyze test coverage" --include-directories ../tests,../docs

# Model selection
gemini -m gemini-2.5-pro -p "Deep analysis query"
```

### Output Formats

**JSON output** (`--output-format json`):
```json
{
  "response": "The model's answer as a string",
  "stats": { "token_usage": ..., "latency": ... },
  "error": null
}
```

**Stream JSON** (`--output-format stream-json`):
- `init`: Session metadata (session ID, model)
- `message`: User/assistant message chunks
- `tool_use`: Tool call requests
- `tool_result`: Tool outputs
- `error`: Non-fatal warnings
- `result`: Final outcome with aggregated stats

### Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error / API failure |
| 42 | Input error (invalid prompt) |
| 53 | Turn limit exceeded |

### Built-in Tools

The CLI includes tools that enhance research capabilities:
- **GoogleSearch**: Web search grounding for real-time information
- **WebFetch**: Fetch and analyze web content
- **ReadFile/ReadFolder**: Local file access
- **SearchText/FindFiles**: Codebase investigation
- **Shell**: Execute shell commands

### GEMINI.md Context Files

Like `CLAUDE.md` for Claude Code, `GEMINI.md` files provide persistent project context. This is the Gemini CLI's native context injection mechanism.

---

## Proposed Architecture

### Architecture Diagram

```
┌─────────────┐    message     ┌────────────────────┐    gemini -p      ┌─────────────┐
│  Supervisor  │──────────────▶│ Research-Supervisor │─────────────────▶│  Gemini CLI  │
│  (or agent)  │◀──────────────│   (persistent)      │◀─────────────────│  (OAuth)     │
└─────────────┘    summary     └────────────────────┘    JSON output    └─────────────┘
                                       │                                        │
                                       │ stores artifacts                       │ uses
                                       ▼                                        ▼
                               _bmad-output/                           Google Account
                               research-reports/                       OAuth (cached)
                                                                       + GoogleSearch
                                                                       + WebFetch
```

### Key Design Decisions

**1. Gemini CLI as execution layer (replaces agent-deep-research)**

The research-supervisor agent invokes `gemini -p "<query>" --output-format json` as a subprocess. This replaces the `uv run agent-deep-research` pattern entirely.

**2. OAuth authentication (replaces API key)**

The user runs `gemini` once interactively to complete the OAuth flow. After that, tokens are cached and refreshed automatically. No API key management, no environment variables to set, no secrets to protect.

**3. Gemini Pro for complex queries, Flash for simple ones**

| Query Type | Model | Daily Limit | Use Case |
|---|---|---|---|
| Deep analysis | `gemini-2.5-pro` | ~50 RPD | Architecture research, technology evaluation |
| Quick lookups | `gemini-2.5-flash` | ~1,000 RPD | Fact-checking, summarization, quick answers |

The research-supervisor selects the model based on query priority/depth.

**4. Context injection via file piping + `--include-directories`**

Instead of `--context` flag from agent-deep-research, use:
```bash
# Pipe assembled context directly
cat _bmad-output/research-context/bundle.md | \
  gemini -m gemini-2.5-pro -p "Given the project context above, research: <query>" \
  --output-format json

# Or use include-directories for broader context
gemini -m gemini-2.5-pro \
  -p "<query>" \
  --include-directories docs/architecture,docs/prd,docs/decisions \
  --output-format json
```

**5. Custom system prompt via GEMINI_SYSTEM_MD**

Create a research-focused system prompt that:
- Instructs Gemini to use GoogleSearch for grounding
- Formats output as structured research reports
- Requests citations and source links
- Specifies the three-layer output format (summary, detailed, raw)

### Wrapper Script

Create `scripts/gemini-research.sh` — a thin shell wrapper that:

1. Assembles context bundle from the 8 context categories
2. Selects model (Pro vs Flash) based on depth parameter
3. Invokes `gemini -p` with JSON output
4. Parses response and stores artifacts
5. Returns summary to caller

```bash
#!/usr/bin/env bash
# scripts/gemini-research.sh
# Usage: ./scripts/gemini-research.sh --depth deep|standard|quick --query "Research question"
#
# Wraps Gemini CLI for structured research queries.
# Requires: gemini CLI authenticated via OAuth (run `gemini` once first).

set -euo pipefail

DEPTH="${1:---depth}"  # deep|standard|quick
QUERY="$2"
OUTPUT_DIR="_bmad-output/research-reports/$(date -u +%Y%m%d-%H%M%S)-$(echo "$QUERY" | tr ' ' '-' | head -c 40)"

# Select model based on depth
case "$DEPTH" in
  deep)    MODEL="gemini-2.5-pro" ;;
  standard) MODEL="gemini-2.5-pro" ;;
  quick)   MODEL="gemini-2.5-flash" ;;
esac

mkdir -p "$OUTPUT_DIR"

# Invoke Gemini CLI in headless mode
gemini -m "$MODEL" \
  -p "$QUERY" \
  --include-directories docs/architecture,docs/prd,docs/decisions \
  --output-format json \
  > "$OUTPUT_DIR/response.json" 2>"$OUTPUT_DIR/stderr.log"

# Extract response text
jq -r '.response' "$OUTPUT_DIR/response.json" > "$OUTPUT_DIR/report.md"

echo "Research complete: $OUTPUT_DIR/report.md"
```

### Dependencies

| Dependency | Type | Installation | Already in Project? |
|---|---|---|---|
| `gemini` CLI | npm package | `npm install -g @google/gemini-cli` | No — new |
| `jq` | System tool | `brew install jq` | Likely present |
| Google Account | OAuth | One-time browser sign-in | User has one |

**Eliminated dependencies:**
- ~~Python~~ (no longer needed)
- ~~uv/uvx~~ (no longer needed for research)
- ~~google-genai SDK~~ (replaced by CLI)
- ~~GEMINI_DEEP_RESEARCH_API_KEY~~ (replaced by OAuth)
- ~~24601/agent-deep-research~~ (replaced by Gemini CLI)

---

## Deep Research Options

### Option A: Standard Gemini Pro Queries with GoogleSearch Grounding (RECOMMENDED)

Use Gemini Pro in headless mode with its built-in GoogleSearch tool for web-grounded research. This isn't the Interactions API "Deep Research" product — it's standard Gemini Pro with web search enabled.

**Pros:**
- Free tier (50 Pro queries/day, 1,000 Flash/day)
- OAuth authentication (no API key)
- Synchronous — no polling needed (responses in seconds to minutes)
- GoogleSearch grounding provides real-time web data
- Simple invocation via `gemini -p`

**Cons:**
- Not as thorough as Deep Research Interactions API (no multi-step iterative research)
- Single-turn queries (no autonomous research planning)
- May need to decompose complex research into multiple focused queries

**Mitigation for depth:** The research-supervisor can decompose complex queries into a series of focused sub-queries, synthesize results, and produce a combined report. This is what the agent architecture is for — the supervisor orchestrates, Gemini executes.

### Option B: Deep Research Extension (REJECTED for now)

The `allenhutchison/gemini-cli-deep-research` extension provides true deep research via the Interactions API.

**Why rejected:**
- Requires a **paid API key** — contradicts the OAuth-only goal
- Deep Research API costs $2-5/query
- Free-tier keys get 429 errors
- The extension uses `GEMINI_DEEP_RESEARCH_API_KEY`, not OAuth

**Future option:** If the user later wants true Deep Research, this extension can be added without changing the base architecture. The wrapper script just needs a `--deep-research` flag that delegates to the extension instead of standard `gemini -p`.

### Option C: Gemini Interactions API Direct (REJECTED)

Direct REST calls to the Interactions API for deep research.

**Why rejected:** Same paid-API-key requirement as Option B, plus must implement polling/retry from scratch.

---

## Migration Plan

### Phase 1: Gemini CLI Setup & Authentication

**Replaces Story 54.2** (agent-deep-research CLI Integration)

- Install Gemini CLI: `npm install -g @google/gemini-cli`
- Run OAuth flow: `gemini` → browser sign-in → token cached
- Verify: `gemini -p "Hello" --output-format json` returns valid JSON
- Document setup in project README or agent definition
- No `_tools/` directory needed (CLI is installed globally via npm)
- No `.gitignore` changes for `_tools/`

### Phase 2: Wrapper Script & Context Assembly

**Replaces Stories 54.2 + 54.3** (CLI setup + context packaging)

- Create `scripts/gemini-research.sh` wrapper script
- Implement context bundle assembly (reuse the 8-bundle strategy from existing design)
- Create `GEMINI.md` research system prompt for project context
- Test with representative queries at `--depth quick` and `--depth deep`

### Phase 3: Research Supervisor Agent Update

**Replaces Story 54.1** (agent definition)

- Update the research-supervisor agent definition to invoke `gemini -p` instead of `uv run`
- Update the polling loop — **no longer needed** for standard queries (synchronous)
- For long-running Pro queries, add a timeout (Gemini CLI has built-in timeout handling)
- Update messaging protocol (same request format, updated execution backend)

### Phase 4: Result Shielding & Storage

**Minimal changes to Story 54.4** (result shielding)

- Parse JSON output from `gemini -p --output-format json`
- Extract `.response` field as the research report
- Apply the three-layer shielding (executive summary → detailed → raw)
- Store in `_bmad-output/research-reports/YYYYMMDD-HHMMSS-<slug>/`

### Phase 5: Rate Limiting & Budget

**Simplifies Story 54.5** (budget management)

- Track daily Pro query count (50/day limit)
- Track daily Flash query count (1,000/day limit)
- Priority queue remains useful for Pro query allocation
- Remove cost estimation (`--dry-run`) — queries are free
- Budget file tracks counts, not dollars

### Story Mapping

| Original Story | New Scope | Changes |
|---|---|---|
| 54.1 Agent Definition | Update to use Gemini CLI | Remove Python refs, add npm/gemini setup |
| 54.2 CLI Setup | Much simpler — `npm install -g @google/gemini-cli` | No `_tools/`, no `uv`, no API key |
| 54.3 Context Packaging | Mostly unchanged | Use `--include-directories` and stdin piping instead of `--context` |
| 54.4 Result Shielding | Parse JSON instead of agent-deep-research output | Simpler parsing, same storage strategy |
| 54.5 Rate Limiting | Count-based instead of cost-based | No dollars, just request counts |

---

## Risks & Mitigations

### 1. OAuth Token Expiry During Long Sessions

**Risk:** OAuth token expires mid-research, causing 401 errors.
**Mitigation:** Gemini CLI handles token refresh automatically. If refresh fails, the research-supervisor retries once, then escalates to supervisor with "re-authenticate" message.

### 2. Rate Limit Exhaustion

**Risk:** 50 Pro queries/day may be insufficient for heavy research days.
**Mitigation:**
- Use Flash (1,000/day) for simple queries, reserve Pro for deep analysis
- Priority queue ensures high-priority queries get Pro slots
- Budget tracking warns at 80% daily usage
- Research-supervisor can defer low-priority queries to next day

### 3. Research Depth vs. Deep Research API

**Risk:** Standard Gemini Pro + GoogleSearch may not match the depth of the Interactions API Deep Research.
**Mitigation:**
- Research-supervisor decomposes complex queries into focused sub-queries
- Multi-turn research: first query identifies key areas, follow-ups drill down
- For truly complex research, the architecture supports adding the deep-research extension later (paid API key, opt-in)

### 4. Node.js / npm Dependency

**Risk:** Adding npm as a dependency for a Go project.
**Mitigation:**
- Gemini CLI can run via `npx @google/gemini-cli` without global install
- The dependency is for the research agent infrastructure, not the Go application itself
- Alternatively, if Gemini CLI releases a standalone binary in the future, switch to that
- Consider `brew install gemini-cli` if/when a Homebrew formula is available

### 5. Gemini CLI Breaking Changes

**Risk:** The CLI is actively developed (weekly preview releases), flags may change.
**Mitigation:**
- Pin to a specific version in the wrapper script: `npx @google/gemini-cli@0.32.1`
- Wrapper script isolates the rest of the system from CLI changes
- Monitor release notes for breaking changes

---

## Rejected Approaches

### 1. Keep agent-deep-research (Status Quo)

**Why rejected:** Requires paid API key, Python dependency, third-party tool. User explicitly requested eliminating all three.

### 2. Gemini CLI Deep Research Extension

**Why rejected:** Still requires paid API key (`GEMINI_DEEP_RESEARCH_API_KEY`). Free-tier keys get 429 errors. Contradicts the OAuth-only goal.

### 3. Direct Interactions API via curl

**Why rejected:** Requires paid API key. Also must implement polling, retry, error handling from scratch. The Gemini CLI already handles all of this.

### 4. Python google-genai SDK with OAuth

**Why rejected:** While technically possible to use OAuth with the Python SDK, this retains the Python dependency. The user specifically wants to eliminate Python from the research stack.

### 5. MCP Server Approach

**Why rejected:** (Same as original D-154 rejection) — engineering effort is high for a protocol that's designed for real-time tool use, not 5-45 minute async research tasks. The CLI headless mode is a better fit.

---

## Open Questions

1. **Should we pin the Gemini CLI version or use latest?** Pinning prevents surprises; latest gets improvements. Recommend: pin in wrapper script, update manually.

2. **Should research reports be git-tracked?** They can be large. Recommend: same as original design — gitignore report contents, track `budget.json` for transparency.

3. **How should the research-supervisor handle multi-turn research?** When a single query isn't enough, should it automatically issue follow-up queries? Recommend: yes, with a configurable max-turns (default 3) and total Pro budget per research task (default 5 queries).

4. **Should we create a GEMINI.md in the project root?** This would give Gemini CLI automatic project context. Pro: automatic context for any research query. Con: another context file to maintain alongside CLAUDE.md. Recommend: yes, create a minimal GEMINI.md that references the same project docs.

5. **What happens when the user's Google account doesn't have access to Gemini Pro?** The free tier should include Pro access, but limits may vary. The wrapper script should detect model availability and fall back to Flash if Pro is unavailable.

---

## Summary

The rearchitecture replaces a Python-based, paid-API-key, third-party tool approach with the official Gemini CLI using OAuth authentication. This eliminates the Python dependency, API key requirement, and per-query costs while retaining the valuable architectural patterns from the original Epic 54 design (persistent agent, context packaging, result shielding, budget tracking).

**Key trade-off:** We lose the Interactions API "Deep Research" capability (multi-step autonomous research) in exchange for zero cost, simpler authentication, and fewer dependencies. The research-supervisor's orchestration layer compensates by decomposing complex queries into focused sub-queries — leveraging the agent architecture to achieve depth through composition rather than a single API call.
