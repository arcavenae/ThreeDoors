# Agentic Knowledge System — Concept Research

**Date:** 2026-03-15
**Type:** Blue-sky research (non-actionable)
**Status:** Draft

---

## Executive Summary

This document explores how a graph-backed knowledge system with RAG and reranking could improve agentic engineering workflows like multiclaude, and how ThreeDoors would interface with it as a client. The core thesis: flat-file knowledge (PRDs, BOARD.md, story files, architecture docs) works remarkably well at small scale but hits fundamental limits around cross-referencing, staleness detection, and semantic retrieval. A knowledge graph doesn't replace files — it indexes them, connects them, and makes them queryable in ways grep cannot achieve.

---

## SCOPE 1: Agentic Engineering Knowledge System

### 1.1 The Problem with Flat Files at Scale

ThreeDoors/multiclaude currently uses ~100+ planning artifacts, 60+ BOARD decisions, 40+ story files, 11 agent definitions, and growing. Agents interact with this knowledge through:

- **Direct file reads** — Agent knows the exact path (`docs/decisions/BOARD.md`)
- **Grep searches** — Agent searches for keywords across files
- **Hardcoded references** — Agent definitions embed specific file paths

This works because:
- The corpus is small enough that agents can read most of it
- File naming conventions are disciplined (`X.Y.story.md`)
- CLAUDE.md provides a roadmap to key files

But it breaks when:
- **Cross-referencing** — "What decisions affected this epic?" requires reading BOARD.md, finding entries that mention the epic, then reading referenced ADRs. A human does this by memory; an agent does it by expensive sequential reads.
- **Staleness** — project-watchdog polls for drift between story files and planning docs. The polling interval (SLA: 6 hours) means agents may act on stale data. A knowledge system could enforce consistency at write time.
- **Semantic queries** — "What prior decisions are relevant to implementing OAuth?" can't be answered by grep because the word "OAuth" may not appear in all relevant decisions (some may discuss "authentication", "token management", or "session handling").
- **Relationship awareness** — "Which stories are blocked by this architectural decision?" requires understanding dependency chains that span multiple file types.
- **Context window pressure** — Agents read entire files to find relevant paragraphs. A retrieval layer would return only the relevant chunks.

### 1.2 Knowledge Graph Schema

#### Node Types

| Node Type | Source | Properties |
|-----------|--------|------------|
| **Epic** | `docs/prd/epic-list.md`, `epics-and-stories.md` | id, title, phase, status, description |
| **Story** | `docs/stories/*.story.md` | id, title, status, acceptance_criteria[], epic_ref |
| **Decision** | `docs/decisions/BOARD.md` | id, type (Open/Active/Pending/Decided), date, adopted_approach, rejected_alternatives[] |
| **ADR** | `docs/decisions/adr-*.md` | id, title, status, context, decision, consequences |
| **PRD Section** | `docs/prd/*.md` | section_id, content_hash, last_modified |
| **Architecture Doc** | `docs/architecture/*.md` | doc_id, content_hash, components_mentioned[] |
| **Research Artifact** | `_bmad-output/planning-artifacts/*` | id, type, date, findings_summary |
| **Agent Definition** | `agents/*.md` | name, class, authority_boundaries, guardrails[] |
| **Incident** | Extracted from agent definitions | id (INC-NNN), description, guardrail_produced |
| **Sprint Report** | `docs/operations/sprint-*.md` | sprint_id, date_range, metrics |
| **PR** | GitHub API | number, title, status, story_ref, files_changed[] |
| **Issue** | GitHub API | number, title, status, labels[], triage_status |

#### Edge Types

| Edge | From | To | Semantics |
|------|------|----|-----------|
| `BELONGS_TO` | Story | Epic | Story is part of epic |
| `DECIDED_BY` | Epic/Story | Decision | Work guided by this decision |
| `REJECTED_IN` | Alternative | Decision | Option was considered but rejected |
| `SUPERSEDES` | Decision | Decision | Newer decision overrides older |
| `BLOCKS` | Story | Story | Dependency |
| `IMPLEMENTS` | PR | Story | PR fulfills story |
| `REFERENCES` | Any | Any | Generic cross-reference |
| `PRODUCED_BY` | Artifact | Event (party mode, spike) | Provenance |
| `GUARDS_AGAINST` | Guardrail | Incident | Defensive relationship |
| `OWNED_BY` | Node | Agent | Authority boundary |
| `MODIFIED_IN` | File | PR | Change tracking |

#### Example Queries This Enables

```
# "What decisions affect Epic 24 (MCP)?"
MATCH (e:Epic {id: 24})-[:DECIDED_BY]->(d:Decision) RETURN d

# "What was rejected when we chose the MCP transport?"
MATCH (d:Decision)-[:REJECTED_IN]->(a:Alternative)
WHERE d.id = 'D-042' RETURN a

# "Which stories are blocked and what's blocking them?"
MATCH (s:Story)-[:BLOCKS]->(blocked:Story)
WHERE blocked.status = 'Not Started' RETURN s, blocked

# "What incidents led to current agent guardrails?"
MATCH (i:Incident)-[:GUARDS_AGAINST]-(g:Guardrail)-[:DEFINED_IN]->(a:AgentDef)
RETURN i, g, a

# "What prior art is relevant to implementing a keybinding system?"
# (This is where RAG shines — semantic similarity, not keyword match)
SEMANTIC_SEARCH("keybinding configuration user customization", type=Decision|Architecture)
```

### 1.3 RAG Layer: From Grep to Semantic Retrieval

**Current flow (grep-based):**
```
Agent needs context → grep for keywords → read matching files → parse mentally → act
```

**Proposed flow (RAG-based):**
```
Agent needs context → semantic query → retrieve ranked chunks → act on top-k results
```

#### Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Flat Files  │────▶│  Indexer      │────▶│  Vector DB  │
│  (source of  │     │  (on write/   │     │  (embeddings│
│   truth)     │     │   commit)     │     │   + graph)  │
└─────────────┘     └──────────────┘     └──────┬──────┘
                                                 │
┌─────────────┐     ┌──────────────┐            │
│  Agent       │────▶│  Query API   │◀───────────┘
│  (multiclaude│     │  (MCP server │
│   or Claude) │     │   or REST)   │
└─────────────┘     └──────────────┘
```

Key design: **files remain the source of truth**. The graph/vector DB is a derived index, not a primary store. This preserves:
- Git history and version control
- Human readability (files are still markdown)
- SOUL.md's local-first principle (no cloud dependency)
- Existing workflows (CLAUDE.md, agent definitions, story files all still work)

#### RAG Benefits for Specific Agent Workflows

**Story Implementation (worker agents):**
- Before: Worker reads story file, CLAUDE.md, maybe BOARD.md. Might miss relevant architecture decisions or prior research.
- After: Worker queries "context for implementing Story 24.5" → gets story file + relevant architecture sections + related BOARD decisions + similar prior implementations, ranked by relevance.

**Issue Triage (envoy):**
- Before: Envoy reads issue, greps for related keywords, may miss connections.
- After: Envoy queries "similar issues and resolutions" → finds past issues with similar symptoms, their root causes, and the PRs that fixed them.

**Decision Making (party mode/supervisor):**
- Before: Party mode participants read whatever files they're pointed to. Easy to miss rejected alternatives from prior decisions.
- After: Query "prior decisions about X" → gets full decision history including rejected alternatives and their rationale.

**Architecture Review (arch-watchdog):**
- Before: Watches for architectural drift by reading PRs and comparing against docs.
- After: Queries "architectural constraints relevant to files changed in PR #NNN" → gets specific constraints that apply to the changed components.

### 1.4 Reranker: Why Grep's Ranking Is Insufficient

Grep returns results ordered by file position, not relevance. A semantic search returns results ordered by embedding similarity, which is better but still noisy. A reranker takes the top-N results from retrieval and re-scores them with a more expensive model that considers:

- **Query-document interaction** — Cross-attention between query and result (vs. independent embedding)
- **Recency** — Recent decisions should outweigh old ones (especially for superseded decisions)
- **Authority** — BOARD decisions outweigh casual mentions in research artifacts
- **Agent-specific relevance** — A worker implementing a story cares about different things than arch-watchdog reviewing a PR

**Local reranker options (SOUL.md local-first compatible):**

| Reranker | Size | Quality | Speed | Notes |
|----------|------|---------|-------|-------|
| Cohere rerank-v3 (API) | Cloud | Excellent | Fast | Violates local-first |
| `bge-reranker-v2-m3` | 560MB | Very good | ~50ms/query | Best local option |
| `ms-marco-MiniLM-L-6-v2` | 80MB | Good | ~10ms/query | Lightweight, good enough for code |
| ColBERT v2 | 400MB | Excellent | ~30ms/query | Token-level interaction, great for code |
| `jina-reranker-v2-base-multilingual` | 560MB | Very good | ~40ms/query | Good multilingual support |

**Recommendation:** `bge-reranker-v2-m3` for quality, `ms-marco-MiniLM` if size/speed matter. Both run locally on Apple Silicon with good performance.

### 1.5 Agent Stubs: Knowledge-Aware Agent Definitions

Current agent definitions embed file paths:
```markdown
## Workflow
1. Read `docs/decisions/BOARD.md` for relevant decisions
2. Check `docs/prd/epics-and-stories.md` for scope
3. Read `docs/stories/X.Y.story.md` for acceptance criteria
```

Knowledge-aware definitions would reference queries:
```markdown
## Workflow
1. Query knowledge system: "active decisions affecting {current_epic}"
2. Query knowledge system: "scope boundaries for {task_description}"
3. Query knowledge system: "acceptance criteria and related context for Story {X.Y}"
```

#### Stub Design Pattern

```markdown
# Agent: worker

## Context Acquisition
On task start, execute these knowledge queries:

### Required Context
- `KG.story({story_id})` → story file + acceptance criteria
- `KG.decisions_for(story.epic)` → relevant BOARD decisions
- `KG.architecture_constraints(story.files_likely_touched)` → applicable architecture rules
- `KG.similar_implementations(story.description, limit=3)` → prior stories with similar scope

### Optional Context (fetch if needed during implementation)
- `KG.semantic("how does {component} work", scope=architecture)` → architecture docs
- `KG.incidents_for(story.files_likely_touched)` → known incidents affecting these files
- `KG.rejected_alternatives(decision_id)` → why other approaches were rejected
```

This is more resilient than file paths because:
- If files are reorganized, the graph updates; agent queries still work
- Agents get only relevant chunks, not entire files
- New relationships (a new decision affecting an old epic) are automatically discoverable

### 1.6 Storage Strategy: Dual-Write with File Primary

**Recommended: File-primary with derived graph.**

```
Write path:  Agent → writes file → git commit → post-commit hook → indexer → graph + vectors
Read path:   Agent → query API → graph traversal + vector search → reranked chunks
```

Why not graph-primary:
- Loses git history, diffs, PRs
- Breaks all existing tooling (CLAUDE.md references, grep, human review)
- Violates SOUL.md local-first (graph DBs are harder to back up/version than files)
- Migration cost is enormous

The indexer runs as a post-commit hook or filesystem watcher:
1. Parse changed markdown files into structured chunks
2. Extract entities (epic IDs, story IDs, decision IDs, file paths mentioned)
3. Generate embeddings for each chunk
4. Update graph nodes/edges
5. Update vector index

**Consistency guarantee:** The graph is always rebuildable from files. If the graph gets corrupted, `just rebuild-knowledge-index` regenerates it from the git working tree.

---

## SCOPE 2: ThreeDoors as Client

### 2.1 Integration Architecture

ThreeDoors already has a clean provider pattern (`TaskProvider` interface) and an MCP server (Epic 24). The knowledge system integration should follow both patterns:

```
┌─────────────────────┐         ┌──────────────────────┐
│  ThreeDoors TUI     │         │  Knowledge System     │
│                     │         │  (MCP Server)         │
│  ┌───────────────┐  │  MCP    │  ┌────────────────┐  │
│  │ MCP Client    │──┼────────▶│  │ Query Engine   │  │
│  │ (new)         │  │         │  │ (graph + RAG)  │  │
│  └───────────────┘  │         │  └────────────────┘  │
│                     │         │                      │
│  ┌───────────────┐  │         │  ┌────────────────┐  │
│  │ TaskProviders │  │         │  │ Flat File      │  │
│  │ (existing)    │  │         │  │ Index          │  │
│  └───────────────┘  │         │  └────────────────┘  │
└─────────────────────┘         └──────────────────────┘
```

**Not a new TaskProvider.** A TaskProvider loads/saves tasks. The knowledge system provides _context about_ tasks — related decisions, architecture constraints, similar past work. This is a different concern. It's more like a `ContextProvider` or `KnowledgeClient`.

### 2.2 What ThreeDoors Consumes

| Data | Use Case | UI Impact |
|------|----------|-----------|
| **Enriched task context** | When user selects a task from the three doors, show related decisions, blocking stories, relevant architecture constraints | Info panel below doors |
| **Cross-project relationships** | "This task relates to 3 tasks in Project X" | Relationship indicator |
| **Urgency signals** | Knowledge system knows sprint deadlines, dependent stories, PR states | Sort/priority hints |
| **Semantic search results** | User searches "auth" and finds tasks about "login", "session", "OAuth" | Enhanced search view |
| **Decision history** | "This task was shaped by Decision D-042 (rejected SSE, chose stdio)" | Context on why task exists |

### 2.3 What ThreeDoors Pushes Back

| Data | Destination | Purpose |
|------|-------------|---------|
| **Task completions** | Knowledge graph node status update | Keep graph current |
| **Session decisions** | New edges: "User chose Task A over Task B at time T" | Decision pattern analysis |
| **Engagement patterns** | Which tasks get deferred repeatedly | Retrospector insights |
| **Context requests** | "User asked for more context on Task X" | Demand signal for indexing priority |

### 2.4 MCP as the Protocol (Epic 24 Alignment)

ThreeDoors already implements an MCP server (Epic 24). The knowledge system should be a _separate_ MCP server that ThreeDoors connects to as a client. This follows the MCP design philosophy: composable tool servers.

```
ThreeDoors ──MCP client──▶ threedoors-mcp (task management tools)
ThreeDoors ──MCP client──▶ knowledge-mcp (context/search tools)
Any AI tool ──MCP client──▶ knowledge-mcp (same server, different client)
```

#### Knowledge MCP Server — Tool Surface

```json
{
  "tools": [
    {
      "name": "knowledge.query",
      "description": "Semantic search across all indexed knowledge",
      "parameters": {
        "query": "string — natural language query",
        "scope": "enum — decisions|architecture|stories|all",
        "limit": "int — max results (default 5)"
      }
    },
    {
      "name": "knowledge.related",
      "description": "Find nodes related to a given entity",
      "parameters": {
        "entity_id": "string — e.g., 'story:24.5', 'decision:D-042', 'epic:24'",
        "relationship": "string — optional filter (blocks, decided_by, implements)",
        "depth": "int — traversal depth (default 1)"
      }
    },
    {
      "name": "knowledge.context_for_story",
      "description": "Get full implementation context for a story",
      "parameters": {
        "story_id": "string — e.g., '24.5'"
      }
    },
    {
      "name": "knowledge.decisions",
      "description": "Find decisions affecting a given scope",
      "parameters": {
        "scope": "string — epic ID, component name, or topic",
        "include_rejected": "bool — include rejected alternatives"
      }
    },
    {
      "name": "knowledge.similar",
      "description": "Find similar past work",
      "parameters": {
        "description": "string — describe what you're looking for",
        "type": "enum — story|decision|incident|architecture"
      }
    }
  ],
  "resources": [
    {
      "name": "knowledge://graph/stats",
      "description": "Current graph statistics (node/edge counts, freshness)"
    },
    {
      "name": "knowledge://graph/entity/{id}",
      "description": "Full node data for any entity"
    }
  ]
}
```

### 2.5 Protocol Choice: MCP over REST/GraphQL/gRPC

| Protocol | Pros | Cons | Verdict |
|----------|------|------|---------|
| **MCP** | Already implemented in ThreeDoors; composable; any AI tool can use it; stdio transport = local-first | Young protocol; limited ecosystem tooling | **Recommended** |
| REST | Universal; simple; well-tooled | No streaming; no tool/resource semantics; another API to maintain | Good fallback |
| GraphQL | Natural fit for graph data; flexible queries | Heavy client library; overkill for this use case; schema maintenance burden | Over-engineered |
| gRPC | Fast; typed; streaming | Protobuf complexity; poor fit for exploratory queries; hard to debug | Wrong tool |

MCP wins because:
1. ThreeDoors already speaks MCP (Epic 24)
2. Other AI tools (Claude Code, Cursor, etc.) already speak MCP
3. Stdio transport satisfies local-first (no network, no ports, no auth)
4. Tool/resource semantics map naturally to knowledge queries
5. One server serves both ThreeDoors and multiclaude agents

---

## SCOPE 3: Technology Landscape

### 3.1 Existing Tools and Approaches

| Tool | Approach | Graph? | RAG? | Local? | Agent-aware? | Assessment |
|------|----------|--------|------|--------|--------------|------------|
| **Obsidian** | Markdown + link graph | Visual only | No | Yes | No | Graph view is display-only, not queryable. No API for agents. |
| **Notion** | Databases + relations | Implicit | No | No | No | Cloud-only, violates local-first. Relations are manual. |
| **Roam Research** | Block-level graph | Yes (block refs) | No | No | No | Graph is real but cloud-only. No programmatic access for agents. |
| **Logseq** | Local-first block graph | Yes | No | Yes | No | Closest to our values. But graph is for human navigation, not agent queries. |
| **Mem.ai** | AI-native knowledge | Implicit | Yes | No | No | Cloud-only, proprietary. Good UX ideas but wrong architecture. |
| **Khoj** | Local AI search | No | Yes | Yes | Partial | RAG over local files. No graph. Good RAG reference implementation. |
| **Anything LLM** | Local RAG chatbot | No | Yes | Yes | No | RAG pipeline reference. No graph or structured knowledge. |
| **Continue.dev** | IDE context for AI | No | Partial | Yes | Partial | @codebase indexing is similar concept. No graph, just vector search. |
| **Cursor** | Codebase indexing | No | Yes | Yes | Partial | Good vector search over code. No graph, no structured knowledge. |
| **Pieces.app** | Snippet + context management | Partial | Yes | Yes | Partial | Closest to "knowledge system for developers". But snippet-focused, not project-knowledge-focused. |

**Gap analysis:** No existing tool combines (1) local-first, (2) graph structure, (3) RAG retrieval, (4) agent-queryable API, (5) MCP interface. This is a greenfield opportunity.

### 3.2 Embedded Graph DBs with RAG Capabilities

| Database | Type | Embedding Support | Local? | Go SDK | Notes |
|----------|------|-------------------|--------|--------|-------|
| **SQLite + FTS5** | Relational + full-text | Via extension (sqlite-vec) | Yes | Yes (mattn/go-sqlite3) | Simplest option. FTS5 for keyword search, sqlite-vec for vectors. Graph via adjacency tables. Battle-tested. |
| **DuckDB** | Analytical | Via extension | Yes | Yes | Better for analytics than graph queries. Good for sprint metrics. |
| **Milvus Lite** | Vector DB | Native | Yes | Yes | Pure vector search, no graph structure. |
| **Qdrant** | Vector DB | Native | Yes (binary) | Yes | Better than Milvus for filtering. No graph. |
| **Neo4j** | Graph DB | Via plugin | Embedded option | Yes | Full graph but heavy. Embedded mode is Java-based. |
| **Dgraph** | Graph DB | No | Self-hosted | Yes | Go-native but heavy. No vector support. |
| **Kùzu** | Embedded graph | No | Yes | Go bindings (CGo) | Lightweight embedded graph DB. Fast. No vector support yet. |
| **LanceDB** | Vector DB | Native | Yes | Partial | Rust-based, fast, good for RAG. No graph. |
| **TypeDB** | Knowledge graph | No | Self-hosted | Partial | Strong knowledge modeling but heavy. |

**Recommended stack:**

```
SQLite (graph via adjacency tables + FTS5 for keyword)
  + sqlite-vec (vector search for semantic)
  + Go stdlib (no heavy dependencies)
```

Why SQLite:
- Single file, zero configuration, local-first by nature
- Already used everywhere in the Go ecosystem
- `sqlite-vec` extension adds vector similarity search
- FTS5 adds high-quality full-text search
- Graph queries via adjacency tables + recursive CTEs
- Entire knowledge base is a single `.db` file in the project root
- Rebuildable from flat files (no data loss if DB deleted)
- Excellent Go bindings (`mattn/go-sqlite3` or `modernc.org/sqlite` for pure Go)

**Alternative for graph-heavy queries:** Kùzu (embedded, C++ with Go bindings) if recursive CTE performance on SQLite becomes a bottleneck. But start with SQLite — it handles most graph patterns well enough.

### 3.3 Local Embedding Models

| Model | Size | Dims | Quality | Speed (M1) | Notes |
|-------|------|------|---------|------------|-------|
| `all-MiniLM-L6-v2` | 80MB | 384 | Good | ~5ms/doc | Lightweight, fast, good baseline |
| `bge-small-en-v1.5` | 130MB | 384 | Better | ~8ms/doc | Better quality, still small |
| `bge-base-en-v1.5` | 440MB | 768 | Very good | ~15ms/doc | Best quality/size tradeoff |
| `nomic-embed-text-v1.5` | 550MB | 768 | Very good | ~15ms/doc | Open source, good for code |
| `mxbai-embed-large-v1` | 1.3GB | 1024 | Excellent | ~30ms/doc | Best quality, but large |

**Recommendation:** `bge-small-en-v1.5` for the initial build (fast, small, good enough for structured markdown). Upgrade to `bge-base` if quality needs improvement.

**Runtime:** Use `ollama` for embedding inference — it's already the standard local model runner, handles model management, and has a simple HTTP API. No need to bundle ONNX runtime or custom inference code.

### 3.4 Could This Be an MCP Server for Any AI Tool?

**Yes, and this is the strongest argument for building it.**

The knowledge system as an MCP server is not ThreeDoors-specific. Any AI coding tool could benefit:

```
Claude Code ──MCP──▶ knowledge-mcp ──▶ project knowledge graph
Cursor      ──MCP──▶ knowledge-mcp ──▶ same graph
Windsurf    ──MCP──▶ knowledge-mcp ──▶ same graph
ThreeDoors  ──MCP──▶ knowledge-mcp ──▶ same graph
multiclaude ──MCP──▶ knowledge-mcp ──▶ same graph
```

This means:
- A user working in Claude Code gets "what decisions affect this file?" context
- A Cursor user gets "similar past implementations" when starting a new feature
- ThreeDoors gets enriched task context
- multiclaude agents get structured knowledge queries

**The knowledge MCP server could be a standalone open-source project** — not tied to ThreeDoors or multiclaude. ThreeDoors and multiclaude would be early adopters/test cases, but the tool serves any project that uses markdown-based knowledge management.

---

## Architecture Sketch: Putting It Together

```
┌─────────────────────────────────────────────────────────────┐
│                    Local Machine                             │
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ Claude   │  │ Cursor   │  │ Three    │  │ multi    │  │
│  │ Code     │  │          │  │ Doors    │  │ claude   │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │              │              │              │        │
│       └──────────────┴──────┬───────┴──────────────┘        │
│                             │ MCP (stdio)                   │
│                      ┌──────┴──────┐                        │
│                      │ knowledge-  │                        │
│                      │ mcp server  │                        │
│                      └──────┬──────┘                        │
│                             │                               │
│                 ┌───────────┼───────────┐                   │
│                 │           │           │                    │
│           ┌─────┴────┐ ┌───┴───┐ ┌────┴─────┐             │
│           │ Graph    │ │Vector │ │ Reranker │             │
│           │ (SQLite  │ │(sqlite│ │(bge-     │             │
│           │  adj.    │ │ -vec) │ │ reranker)│             │
│           │  tables) │ │       │ │          │             │
│           └─────┬────┘ └───┬───┘ └──────────┘             │
│                 │          │                                │
│                 └────┬─────┘                                │
│                      │                                      │
│              ┌───────┴───────┐                              │
│              │ knowledge.db  │  (single SQLite file)        │
│              └───────┬───────┘                              │
│                      │                                      │
│              ┌───────┴───────┐                              │
│              │   Indexer     │  (post-commit hook or        │
│              │   (watcher)   │   filesystem watcher)        │
│              └───────┬───────┘                              │
│                      │                                      │
│              ┌───────┴───────┐                              │
│              │  Flat Files   │  (git-tracked markdown,      │
│              │  (source of   │   YAML, JSONL — unchanged)   │
│              │   truth)      │                              │
│              └───────────────┘                              │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Write:** Agent/human edits a markdown file → git commit
2. **Index:** Post-commit hook triggers indexer → parses file → extracts entities & relationships → generates embeddings → updates SQLite (graph tables + vector index)
3. **Query:** Agent sends MCP tool call → knowledge-mcp server → query planner routes to graph traversal, vector search, or hybrid → reranker scores results → returns top-k chunks with metadata
4. **Rebuild:** `just rebuild-knowledge-index` regenerates entire DB from flat files (idempotent, ~30s for ThreeDoors-scale corpus)

### What Changes for multiclaude Agents

**Before (grep-based):**
```markdown
# Agent: worker
## On task start:
1. Read docs/stories/{story_id}.story.md
2. Read docs/decisions/BOARD.md, search for relevant decisions
3. Read CLAUDE.md for project rules
4. Maybe read architecture docs if you think they're relevant
```

**After (knowledge-aware):**
```markdown
# Agent: worker
## On task start:
1. Call knowledge.context_for_story({story_id})
   → Returns: story file, relevant decisions, architecture constraints,
     similar past implementations, known incidents for affected files
2. Read CLAUDE.md for project rules (still direct — these are meta-rules, not knowledge)
3. If stuck, call knowledge.query("how does {component} work")
```

The agent gets better context with fewer reads, and discovers relationships it wouldn't have found by grepping.

---

## Open Questions for Future Exploration

1. **Incremental vs. full reindex** — Can we efficiently update only changed nodes on each commit, or is full reindex simpler and fast enough?

2. **Embedding model updates** — When a better embedding model comes out, all vectors need regeneration. How to handle this gracefully?

3. **Multi-project support** — If the knowledge system indexes multiple projects, how are cross-project relationships handled? Shared entities (same user, same tech stack)?

4. **Freshness guarantees** — The current 6-hour SLA for planning doc sync is already a pain point. Can the knowledge system enforce stricter consistency? (Post-commit hook = near-instant, but what about uncommitted changes?)

5. **Query cost model** — Some queries are cheap (graph traversal), some are expensive (vector search + rerank). Should the MCP server expose cost hints so agents can choose?

6. **Privacy boundaries** — If the knowledge system indexes multiple projects, how to ensure one project's agents can't access another project's knowledge? (Probably: one DB per project, but then cross-project queries need a federation layer.)

7. **Human-in-the-loop corrections** — When the graph has a wrong relationship (e.g., story wrongly linked to a decision), how does a human correct it? (Probably: a `knowledge.correct` MCP tool that adds an override edge.)

8. **ollama dependency** — Requiring ollama for embeddings adds a system dependency. Alternative: ship a small ONNX model and use Go ONNX runtime directly, but this adds binary size and build complexity.

---

## Recommendations

1. **Start with MCP server architecture.** Even before building the knowledge system, defining the MCP tool surface forces clarity on what agents actually need.

2. **SQLite + sqlite-vec is the right foundation.** No heavy dependencies, single file, rebuildable from source files, excellent Go support.

3. **File-primary, graph-derived.** Never let the graph become the source of truth. Files + git remain authoritative.

4. **Build as a standalone project.** The knowledge MCP server is valuable beyond ThreeDoors/multiclaude. It's a general-purpose "project knowledge server" for any AI coding tool.

5. **ThreeDoors connects as an MCP client, not a tightly-coupled integration.** Use the same MCP transport ThreeDoors already supports from Epic 24.

6. **Prototype with ThreeDoors' own docs.** The ~200 files in ThreeDoors' docs/ and _bmad-output/ are a perfect test corpus — small enough to iterate fast, complex enough to test real queries.

7. **Don't build this yet.** This is research. The current flat-file approach works. Build this when agents start consistently failing to find relevant context — that's the signal that the grep-based approach has hit its limits.

---

## Appendix: Comparison with CLAUDE.md / Memory Approach

Claude Code's memory system (`CLAUDE.md` + `.claude/` memory files) is a lightweight version of what's described here. Key differences:

| Aspect | CLAUDE.md / Memory | Knowledge System |
|--------|-------------------|------------------|
| Structure | Flat text, manually curated | Graph + vectors, auto-indexed |
| Query | Full file read | Semantic search + graph traversal |
| Relationships | Implicit (human remembers) | Explicit (edges in graph) |
| Freshness | Manual updates | Auto-indexed on commit |
| Cross-referencing | Grep | Graph traversal |
| Scope | Single conversation context | Full project history |
| Portability | Claude Code only | Any MCP client |

The knowledge system subsumes CLAUDE.md's role for project knowledge while CLAUDE.md retains its role for behavioral instructions and preferences.
