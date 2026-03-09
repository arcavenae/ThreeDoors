# LLM Integration & MCP Server Research

**Date:** 2026-03-06
**Status:** Research Complete
**Method:** Analyst review + 3-round BMAD Party Mode brainstorming (12 agents)
**Authors:** Mary (Analyst), Winston (Architect), John (PM), Amelia (Dev), Sally (UX), Quinn (QA), Murat (TEA), Bob (SM), Victor (Innovation), Carson (Brainstorming), Dr. Quinn (Problem Solver), Maya (Design Thinking), Sophia (Storyteller), Paige (Tech Writer), Barry (Quick Flow)

---

## Executive Summary

ThreeDoors is architecturally well-positioned to expose task management services to LLMs through the Model Context Protocol (MCP). The existing `TaskProvider` interface, multi-source aggregation, session analytics, and enrichment database provide a robust foundation that requires no breaking changes.

**Key Constraint:** LLMs must NEVER directly edit task data. All modifications flow through a proposal/approval pattern — LLMs propose, users approve. This is not a limitation but a core design principle that builds trust and maintains data integrity.

**Competitive Advantage:** No existing task manager offers MCP integration with multi-provider aggregation. ThreeDoors can uniquely answer cross-system questions ("What's blocking me across Jira and local tasks?") and mine unified productivity patterns.

This document covers 10 research areas spanning read-only queries, controlled enrichment, analytics mining, dependency graphing, and advanced interaction patterns.

---

## Table of Contents

1. [MCP Server Architecture](#1-mcp-server-architecture)
2. [Controlled Enrichment API](#2-controlled-enrichment-api)
3. [Task Population from External Sources](#3-task-population-from-external-sources)
4. [Read-Only Query Interface](#4-read-only-query-interface)
5. [Guardrails & Data Integrity](#5-guardrails--data-integrity)
6. [Natural Language Task Queries](#6-natural-language-task-queries)
7. [Task Relationship Graphs](#7-task-relationship-graphs)
8. [Mood-Execution Correlation Analysis](#8-mood-execution-correlation-analysis)
9. [Cross-Provider Dependency Mapping](#9-cross-provider-dependency-mapping)
10. [Advanced Interaction Patterns](#10-advanced-interaction-patterns)
11. [Implementation Roadmap](#11-implementation-roadmap)
12. [Party Mode Insights](#12-party-mode-insights)

---

## 1. MCP Server Architecture

### How MCP Maps to ThreeDoors

The Model Context Protocol defines three primitives that map cleanly to ThreeDoors' existing architecture:

| MCP Primitive | ThreeDoors Mapping | Access Level |
|---|---|---|
| **Resources** (read-only data) | Task lists, session history, analytics, provider status | Read |
| **Tools** (callable functions) | Query tasks, propose enrichments, analyze patterns | Read + Propose |
| **Prompts** (template queries) | Common question templates for task analysis | Read |

### Proposed MCP Resources

```
threedoors://tasks                     # All tasks across providers
threedoors://tasks/{id}                # Single task detail
threedoors://tasks/status/{status}     # Tasks by status
threedoors://tasks/provider/{name}     # Tasks by provider
threedoors://providers                 # Provider registry + health
threedoors://providers/{name}/status   # Individual provider health
threedoors://session/current           # Current session metrics
threedoors://session/history           # Historical session data
threedoors://analytics/patterns        # Discovered patterns
threedoors://analytics/mood            # Mood-execution correlations
threedoors://analytics/productivity    # Productivity trends
threedoors://graph/dependencies        # Task dependency graph
threedoors://graph/cross-provider      # Cross-provider relationships
threedoors://proposals/pending         # Pending proposals for review
```

### Architectural Design

```
┌─────────────────────────────────────────────────────┐
│                   MCP Server                         │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │ Resources│  │  Tools   │  │     Prompts       │  │
│  └────┬─────┘  └────┬─────┘  └────────┬──────────┘  │
│       │              │                 │             │
│  ┌────┴──────────────┴─────────────────┴──────────┐  │
│  │           Security Middleware                   │  │
│  │  RateLimiter → AuditLogger → SchemaValidator    │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │                              │
│  ┌────────────────────┴───────────────────────────┐  │
│  │              Service Layer                      │  │
│  │  ┌──────────┐ ┌────────────┐ ┌──────────────┐  │  │
│  │  │TaskQuery │ │ProposalMgr │ │AnalyticsMiner│  │  │
│  │  └────┬─────┘ └─────┬──────┘ └──────┬───────┘  │  │
│  └───────┼─────────────┼───────────────┼──────────┘  │
│          │             │               │             │
└──────────┼─────────────┼───────────────┼─────────────┘
           │             │               │
┌──────────┴─────────────┴───────────────┴─────────────┐
│              Existing ThreeDoors Core                  │
│  ┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────────┐  │
│  │Registry│ │Aggregator│ │TaskPool │ │SessionTrack│  │
│  └────────┘ └──────────┘ └─────────┘ └────────────┘  │
│  ┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────────┐  │
│  │  WAL   │ │EnrichDB  │ │SyncLog  │ │PatternAnlz │  │
│  └────────┘ └──────────┘ └─────────┘ └────────────┘  │
└──────────────────────────────────────────────────────┘
```

### Go Implementation Pattern

```go
// MCPServer wraps existing core components — no new storage layer needed
type MCPServer struct {
    registry   *core.Registry
    aggregator *core.MultiSourceAggregator
    pool       *core.TaskPool
    session    *core.SessionTracker
    enrichDB   *enrichment.Store
    proposals  *ProposalStore
    analytics  *AnalyticsMiner
    middleware []MCPMiddleware
}

// MCPMiddleware follows the decorator pattern already used by WALProvider
type MCPMiddleware func(Handler) Handler

// Resources are read-only — they expose existing data
func (s *MCPServer) ListResources() []Resource { ... }
func (s *MCPServer) ReadResource(uri string) ([]byte, error) { ... }

// Tools are proposal-only — they cannot directly modify tasks
func (s *MCPServer) CallTool(name string, args map[string]any) (Result, error) { ... }
```

### Key Design Decisions

1. **The MCP server is a separate binary** (`cmd/threedoors-mcp/`) that shares the `internal/` packages. This keeps the TUI and MCP server independently deployable.
2. **No new storage layer** — the MCP server reads from the same YAML files, JSONL logs, and SQLite enrichment DB that the TUI uses.
3. **Proposal store is append-only JSONL** (`~/.threedoors/proposals.jsonl`), consistent with session tracking patterns.
4. **The server runs as a stdio MCP server** for Claude Desktop integration, or as an SSE server for remote access.

---

## 2. Controlled Enrichment API

### The Proposal/Approval Pattern

LLMs can propose changes but never execute them directly. This mirrors how code review works — suggest, review, merge.

### Proposal Types

| Type | Description | Payload |
|---|---|---|
| `enrich-metadata` | Add/update effort, type, location | `{effort: "deep-work", type: "research"}` |
| `add-subtasks` | Break a task into subtasks | `{subtasks: [{text, effort, type}]}` |
| `add-context` | Add context or background info | `{context: "Related to PR #42"}` |
| `add-note` | Suggest a note to append | `{note: "Consider edge case X"}` |
| `suggest-blocker` | Flag a potential blocker | `{blocker: "Depends on API v2 release"}` |
| `suggest-relationship` | Link two tasks | `{targetID, relationType}` |
| `suggest-category` | Propose categorization | `{categories: ["backend", "auth"]}` |
| `update-effort` | Refine effort estimate | `{effort: "deep-work", rationale: "..."}` |

### Proposal Data Model

```go
type Proposal struct {
    ID          string          `json:"id"`          // UUID
    Type        ProposalType    `json:"type"`        // enrich-metadata, add-subtasks, etc.
    TaskID      string          `json:"task_id"`     // Target task (empty for new task proposals)
    BaseVersion time.Time       `json:"base_version"` // Task's UpdatedAt when proposal created
    Payload     json.RawMessage `json:"payload"`     // Type-specific data
    Status      ProposalStatus  `json:"status"`      // pending|approved|rejected|expired|stale
    Source      string          `json:"source"`      // "mcp:claude-desktop", "mcp:cursor", etc.
    Rationale   string          `json:"rationale"`   // Why the LLM suggests this
    CreatedAt   time.Time       `json:"created_at"`
    ReviewedAt  *time.Time      `json:"reviewed_at,omitempty"`
    ExpiresAt   time.Time       `json:"expires_at"`  // Auto-expire after 7 days
}

type ProposalStatus string
const (
    ProposalPending  ProposalStatus = "pending"
    ProposalApproved ProposalStatus = "approved"
    ProposalRejected ProposalStatus = "rejected"
    ProposalExpired  ProposalStatus = "expired"
    ProposalStale    ProposalStatus = "stale"    // Task changed since proposal
)
```

### Optimistic Concurrency Control

Proposals include a `BaseVersion` (the task's `UpdatedAt` timestamp at proposal creation time). When reviewing a proposal:

1. Compare `BaseVersion` to the task's current `UpdatedAt`
2. If they differ, mark the proposal as `stale` — the task has changed
3. Stale proposals require re-evaluation, not automatic rejection
4. This prevents conflicting proposals from silently corrupting state

### TUI Integration (UX Design)

The approval flow should feel lightweight, not bureaucratic:

- **Badge indicator**: A subtle count on the doors view: `[3 suggestions]`
- **Suggestion view** (press `S`): Split pane — proposals on left, affected task on right
- **Quick actions**: Enter = approve, Backspace = reject, Tab = skip, Ctrl+A = approve all
- **Batch operations**: Group proposals by task for efficient review
- **Preview mode**: Show what the task would look like after applying the proposal

---

## 3. Task Population from External Sources

### Intake Pipeline Architecture

```
External Sources                    Intake Pipeline                    User Review
┌──────────────┐     ┌──────────────────────────────┐     ┌──────────────────┐
│ Codebase     │────→│                              │     │                  │
│ Analysis     │     │   IntakeChannel Interface     │     │  Proposal Store  │
├──────────────┤     │                              │────→│  (proposals.jsonl)│
│ GitHub PRs   │────→│   Suggest(ctx, Proposal)     │     │                  │
│ & Issues     │     │                              │     │  ┌────────────┐  │
├──────────────┤     │   Validates:                 │     │  │ TUI Review │  │
│ Calendar     │────→│   - Dedup against existing   │     │  │    View    │  │
│ Events       │     │   - Schema validation        │     │  └────────────┘  │
├──────────────┤     │   - Rate limiting            │     │        │         │
│ LLM Agent    │────→│   - Source attribution        │     │   Approve/Reject │
│ Suggestions  │     │                              │     │        │         │
└──────────────┘     └──────────────────────────────┘     │   TaskProvider   │
                                                          │   .SaveTask()    │
                                                          └──────────────────┘
```

### Source-Specific Intake Strategies

**Codebase Analysis (LLM-powered):**
- LLM scans for TODO/FIXME/HACK comments → suggests tasks
- LLM reviews recent git commits → identifies follow-up work
- LLM analyzes test coverage gaps → suggests testing tasks
- All suggestions include file path, line number, and rationale

**GitHub PRs & Issues:**
- LLM monitors configured repos for new issues → maps to task proposals
- PR review comments → suggest follow-up tasks
- Merged PRs → suggest documentation/cleanup tasks

**Calendar Events:**
- Meeting prep tasks ("Review agenda for X meeting")
- Follow-up tasks after meetings
- Deadline-aware task suggestions ("Feature X due Friday — 3 subtasks remaining")

### IntakeChannel Interface

```go
type IntakeChannel interface {
    Name() string
    Suggest(ctx context.Context, proposal Proposal) error
    HealthCheck() HealthCheckResult
}

// LLMIntakeChannel handles proposals from MCP-connected LLMs
type LLMIntakeChannel struct {
    proposals *ProposalStore
    dedup     *core.DuplicateDetector
    limiter   *rate.Limiter
}
```

### Deduplication at Intake

Before accepting a proposal, check against:
1. Existing tasks (text similarity using existing `DuplicateDetector`)
2. Other pending proposals (prevent duplicate suggestions)
3. Recently rejected proposals (don't re-suggest what the user said no to)
4. Task history (don't suggest tasks that were already completed)

---

## 4. Read-Only Query Interface

### MCP Tools for Querying

```
query_tasks       — Search tasks with filters (status, type, effort, provider, text)
get_task          — Get full detail for a single task
list_providers    — List configured providers and their health status
get_session       — Get current or historical session metrics
get_analytics     — Get pattern analysis results
get_mood_data     — Get mood entries and correlations
walk_graph        — Traverse task relationship graph
search_history    — Search across session history
get_completions   — Get completion statistics for a time range
```

### Query Capabilities

**Structured Queries:**
```json
{
  "tool": "query_tasks",
  "args": {
    "status": ["blocked", "in-progress"],
    "type": "bug",
    "provider": "jira",
    "effort": "quick-win",
    "created_after": "2026-03-01",
    "text_contains": "authentication",
    "limit": 20,
    "sort_by": "updated_at",
    "sort_order": "desc"
  }
}
```

**Time-Range Analytics:**
```json
{
  "tool": "get_completions",
  "args": {
    "from": "2026-02-01",
    "to": "2026-03-01",
    "group_by": "week",
    "include_mood": true,
    "include_patterns": true
  }
}
```

### Response Format

All query responses include metadata for LLM context:

```json
{
  "result": { ... },
  "metadata": {
    "total_count": 47,
    "returned_count": 20,
    "query_time_ms": 12,
    "providers_queried": ["textfile", "jira"],
    "data_freshness": "2026-03-06T10:30:00Z"
  }
}
```

---

## 5. Guardrails & Data Integrity

### Defense-in-Depth Strategy

```
Layer 1: Protocol Enforcement
├── MCP tools are categorized: read-only, propose-only, analyze-only
├── No tool endpoint can directly call TaskProvider.SaveTask()
└── Write operations ONLY through ProposalStore → user approval → SaveTask

Layer 2: Input Validation
├── Task IDs: must match UUID v4 format
├── Text fields: max 500 characters, sanitized for injection
├── Status values: must be valid TaskStatus enum
├── Timestamps: must be valid UTC, not in the future
└── Proposal payloads: validated against per-type JSON schemas

Layer 3: Rate Limiting
├── Global: 100 requests/minute per MCP connection
├── Per-tool: 20 proposals/minute, 60 queries/minute
├── Per-task: max 5 pending proposals per task
└── Burst: allow 10-request burst, then throttle

Layer 4: Audit Trail
├── Every MCP request logged to ~/.threedoors/mcp-audit.jsonl
├── Fields: timestamp, tool, args, result, source, duration_ms
├── Proposal lifecycle: created → reviewed → outcome
├── Tamper detection: SHA-256 hash chain on audit entries
└── Rotation: daily log files, 30-day retention

Layer 5: Proposal Governance
├── Proposals expire after 7 days if unreviewed
├── Stale detection: BaseVersion vs current task version
├── Conflict detection: multiple proposals for same task field
├── Source tracking: which LLM/client made each proposal
└── Rejection memory: don't re-suggest recently rejected proposals
```

### Audit Log Entry

```go
type AuditEntry struct {
    Timestamp  time.Time       `json:"ts"`
    RequestID  string          `json:"req_id"`
    Tool       string          `json:"tool"`
    Args       json.RawMessage `json:"args"`
    Result     string          `json:"result"`    // "ok", "error", "rate_limited"
    Source     string          `json:"source"`    // MCP client identifier
    DurationMs int64           `json:"duration_ms"`
    Error      string          `json:"error,omitempty"`
    PrevHash   string          `json:"prev_hash"` // SHA-256 of previous entry
}
```

### Security Middleware Chain

```go
func NewMCPServer(opts ...Option) *MCPServer {
    handler := coreHandler
    handler = SchemaValidator(handler)    // Validate inputs
    handler = AuditLogger(handler)        // Log everything
    handler = RateLimiter(handler)        // Throttle abuse
    handler = ReadOnlyEnforcer(handler)   // Block direct writes
    return &MCPServer{handler: handler}
}
```

---

## 6. Natural Language Task Queries

### Query Translation Pipeline

```
User Question                    Query Parser                    Execution
"What am I blocked on?"    →    {status: "blocked"}        →    TaskPool.Filter()
"Auth-related tasks"       →    {text_search: "auth*"}     →    Semantic search
"What did I finish today?" →    {status: "complete",       →    TaskPool.Filter()
                                 completed_after: today}
"My deepest tasks"         →    {effort: "deep-work"}      →    TaskPool.Filter()
"Jira bugs"                →    {provider: "jira",         →    Provider-scoped query
                                 type: "bug"}
```

### Common Question Templates (MCP Prompts)

The MCP server exposes prompt templates that LLMs can use for common queries:

```yaml
prompts:
  - name: blocked_tasks
    description: "Show all blocked tasks with their blockers"
    template: |
      Query the user's blocked tasks using query_tasks with status=blocked.
      For each blocked task, show the task text and blocker reason.
      Suggest potential unblocking actions based on the blocker description.

  - name: daily_summary
    description: "Summarize today's task activity"
    template: |
      1. Get today's completions using get_completions(from=today)
      2. Get current in-progress tasks using query_tasks(status=in-progress)
      3. Get current mood data using get_mood_data(from=today)
      4. Summarize: completed count, in-progress count, mood trend, suggested next task

  - name: weekly_retrospective
    description: "Generate a weekly productivity retrospective"
    template: |
      1. Get this week's completions, grouped by day
      2. Get mood data for the week
      3. Get pattern analysis for the week
      4. Narrate: best day, completion trend, mood correlation, recommendations

  - name: task_deep_dive
    description: "Deep analysis of a specific task"
    template: |
      1. Get task detail using get_task(id)
      2. Walk its relationship graph using walk_graph(id, depth=2)
      3. Check session history for interactions with this task
      4. Report: full context, relationships, history, suggestions
```

### Semantic Search Approach

For text-based queries that go beyond exact matching:

1. **Token overlap scoring** — split query and task text into tokens, compute Jaccard similarity
2. **Field-weighted search** — Text field weighted 3x, Context weighted 2x, Notes weighted 1x
3. **Synonym expansion** — "auth" matches "authentication", "login", "credentials"
4. **Recency boost** — recently updated tasks score higher for ambiguous queries
5. **Provider context** — include provider name in searchable text ("jira bug" matches Jira-sourced bugs)

```go
type SearchResult struct {
    Task       *core.Task `json:"task"`
    Score      float64    `json:"score"`      // 0.0 - 1.0
    MatchedOn  []string   `json:"matched_on"` // ["text", "context", "notes"]
    Highlights []string   `json:"highlights"` // Matched fragments
}

func (q *TaskQueryEngine) Search(query string, opts SearchOptions) ([]SearchResult, error) {
    // Tokenize, score, rank, return
}
```

---

## 7. Task Relationship Graphs

### Graph Data Model

Tasks don't currently have explicit dependency fields, but relationships can be both explicit (via proposals) and inferred (via pattern analysis).

```go
type TaskGraph struct {
    Nodes map[string]*TaskNode
    Edges []TaskEdge
}

type TaskNode struct {
    Task     *core.Task
    Provider string
    Depth    int  // Distance from query root
}

type TaskEdge struct {
    FromID   string
    ToID     string
    Type     EdgeType
    Weight   float64  // Confidence for inferred edges
    Source   string   // "explicit", "inferred:text", "inferred:temporal"
}

type EdgeType string
const (
    EdgeBlocks    EdgeType = "blocks"
    EdgeRelatedTo EdgeType = "related-to"
    EdgeSubtaskOf EdgeType = "subtask-of"
    EdgeDuplicateOf EdgeType = "duplicate-of"
    EdgeSequential  EdgeType = "sequential"  // Inferred from completion order
    EdgeCrossRef    EdgeType = "cross-ref"   // Same SourceRef
)
```

### Relationship Inference Engine

When explicit relationships don't exist, infer them from data:

1. **Text Similarity** — Tasks with >70% token overlap are `related-to` (weight = similarity score)
2. **Temporal Sequences** — Tasks completed within 30 minutes of each other with similar text are `sequential`
3. **Cross-Provider Refs** — Tasks sharing `SourceRef` values are `cross-ref`
4. **Blocker Chains** — If Task A's blocker text mentions Task B's text, infer `blocks`
5. **Subtask Patterns** — Tasks where one text is a subset of another suggest `subtask-of`
6. **Duplicate Detection** — Leverage existing `DuplicateDetector` for `duplicate-of` edges

### MCP Graph Tools

```
walk_graph(task_id, depth=2, direction="both", edge_types=["blocks", "related-to"])
  → Returns subgraph centered on task_id

find_paths(from_id, to_id, max_depth=5)
  → Returns all paths between two tasks

get_critical_path(root_id)
  → Returns longest dependency chain from root

get_orphans()
  → Returns tasks with no relationships (potential isolation risk)

get_clusters()
  → Returns groups of related tasks (community detection)
```

### Graph Traversal Example

```
User: "What depends on the authentication refactor?"

LLM calls: walk_graph(task_id="auth-refactor-123", depth=3, direction="outgoing")

Response:
{
  "root": "auth-refactor-123",
  "nodes": [
    {"id": "auth-refactor-123", "text": "Refactor auth middleware", "status": "in-progress"},
    {"id": "api-tests-456", "text": "Update API integration tests", "status": "todo", "depth": 1},
    {"id": "deploy-789", "text": "Deploy auth changes to staging", "status": "todo", "depth": 2},
    {"id": "jira-FEAT-42", "text": "User login flow redesign", "provider": "jira", "depth": 1}
  ],
  "edges": [
    {"from": "auth-refactor-123", "to": "api-tests-456", "type": "blocks", "source": "explicit"},
    {"from": "api-tests-456", "to": "deploy-789", "type": "sequential", "source": "inferred:temporal"},
    {"from": "auth-refactor-123", "to": "jira-FEAT-42", "type": "cross-ref", "source": "inferred:text", "weight": 0.82}
  ]
}
```

---

## 8. Mood-Execution Correlation Analysis

### Data Sources

ThreeDoors session tracking already captures the raw data needed for mood-execution analysis:

- **Mood entries**: Timestamped mood ratings during sessions
- **Task completions**: Which tasks were completed, when, and their attributes
- **Door selections**: Which tasks were chosen vs bypassed
- **Session durations**: How long users work in each session
- **Task bypass patterns**: What was skipped and why

### Correlation Analyses

#### 8.1 Mood × Completion Rate

```
Mood Rating | Avg Completions/Session | Avg Session Duration
────────────┼─────────────────────────┼─────────────────────
Energized   | 4.2                     | 45 min
Focused     | 3.8                     | 52 min
Neutral     | 2.5                     | 30 min
Tired       | 1.8                     | 22 min
Frustrated  | 1.2                     | 18 min
```

LLM insight: "When you report feeling energized, you complete 3.5x more tasks than when frustrated. Your energized sessions also last 2.5x longer."

#### 8.2 Mood × Task Type Preference

```
Mood        | Feature | Bug Fix | Research | Chore
────────────┼─────────┼─────────┼──────────┼──────
Energized   | 45%     | 20%     | 25%      | 10%
Focused     | 30%     | 35%     | 25%      | 10%
Tired       | 10%     | 15%     | 20%      | 55%
Frustrated  | 5%      | 50%     | 10%      | 35%
```

LLM insight: "When tired, you gravitate toward chores (55%). When frustrated, you tackle bugs (50%) — possibly channeling frustration into fixing things."

#### 8.3 Time-of-Day Productivity

```go
type ProductivityProfile struct {
    HourlyCompletionRate map[int]float64    // hour (0-23) → avg completions
    PeakHours           []int               // Top 3 productive hours
    SlumpHours          []int               // Bottom 3 productive hours
    EffortByHour        map[int]string      // hour → most common effort level
    MoodByHour          map[int]string      // hour → most common mood
}
```

LLM insight: "Your peak productivity is 9-11 AM (3.2 completions/hour). After 3 PM, your completion rate drops 60%. Schedule deep-work tasks before noon."

#### 8.4 Streak Analysis

```go
type StreakData struct {
    CurrentStreak    int       // Consecutive days with completions
    LongestStreak    int       // All-time record
    StreakHistory     []Streak  // Past streaks with start/end dates
    AvgStreakLength   float64
    StreakBreakers    []string  // Common reasons streaks end
}
```

LLM insight: "You're on a 4-day streak! Your average streak is 3.2 days. Streaks typically end on Fridays (60% of breaks). Your longest was 11 days."

#### 8.5 Burnout Early Warning

```go
type BurnoutIndicators struct {
    CompletionTrend     float64  // Slope over last 10 sessions (negative = declining)
    MoodTrend           float64  // Slope of mood ratings
    SessionLengthTrend  float64  // Getting shorter?
    BypassRate          float64  // % of shown tasks being skipped
    TaskAvoidanceTypes  []string // Task types being systematically avoided
    DaysWithNoActivity  int      // Recent gap days
    Score               float64  // 0-1 composite score (>0.7 = warning)
}
```

LLM insight: "Burnout risk: MODERATE (0.6). Your completion rate has dropped 30% over 5 sessions, sessions are getting shorter, and you've been avoiding 'deep-work' tasks for 8 days. Consider: shorter sessions, quick-win tasks, or a break."

### MCP Analytics Resources

```
threedoors://analytics/mood-correlation     # Full mood × completion matrix
threedoors://analytics/time-of-day          # Hourly productivity profile
threedoors://analytics/streaks              # Streak data and history
threedoors://analytics/burnout-risk         # Current burnout indicators
threedoors://analytics/task-preferences     # Mood-driven task type preferences
threedoors://analytics/weekly-summary       # Auto-generated weekly summary
```

### Implementation: PatternMiner

```go
type PatternMiner struct {
    sessions *core.SessionStore   // Historical session data
    tasks    *core.TaskPool       // Current task state
    enrichDB *enrichment.Store    // Cross-reference data
}

func (m *PatternMiner) MoodCorrelation(from, to time.Time) (*MoodCorrelation, error)
func (m *PatternMiner) ProductivityProfile(from, to time.Time) (*ProductivityProfile, error)
func (m *PatternMiner) StreakAnalysis() (*StreakData, error)
func (m *PatternMiner) BurnoutRisk() (*BurnoutIndicators, error)
func (m *PatternMiner) WeeklySummary(weekOf time.Time) (*WeeklySummary, error)
func (m *PatternMiner) TaskTypePreferences(mood string) (map[string]float64, error)
```

---

## 9. Cross-Provider Dependency Mapping

### The Multi-Provider Challenge

Users with multiple providers have tasks spread across systems:
- A Jira epic "Launch Feature X" has subtasks in Jira
- Local tasks include "Research X approach" and "Write X tests"
- An Apple Note contains "X meeting notes"
- An Obsidian note has "X architecture decisions"

These are logically related but currently siloed. Cross-provider dependency mapping connects them.

### Cross-Provider Graph

```
┌─────────────────────────────────────────────────────────┐
│                Cross-Provider Task Graph                 │
│                                                         │
│  ┌──────────┐     blocks      ┌──────────────┐         │
│  │ Jira     │────────────────→│ Local        │         │
│  │ FEAT-123 │                 │ "Write tests"│         │
│  │ (epic)   │                 └──────────────┘         │
│  └────┬─────┘                                          │
│       │ subtask-of                                      │
│       ↓                                                 │
│  ┌──────────┐    related-to   ┌──────────────┐         │
│  │ Jira     │────────────────→│ Apple Notes  │         │
│  │ FEAT-124 │                 │ "X meeting"  │         │
│  │ (story)  │                 └──────────────┘         │
│  └──────────┘                                          │
│                                                         │
│  Legend: ── explicit  ╌╌ inferred                       │
│  Colors: 🔵 Jira  🟢 Local  🟡 Apple Notes  🟣 Obsidian│
└─────────────────────────────────────────────────────────┘
```

### Cross-Provider Link Discovery

```go
type CrossProviderLinker struct {
    aggregator *core.MultiSourceAggregator
    enrichDB   *enrichment.Store
    dedup      *core.DuplicateDetector
}

func (l *CrossProviderLinker) DiscoverLinks() ([]TaskEdge, error) {
    // 1. Match by SourceRef (explicit cross-references)
    // 2. Match by text similarity across providers
    // 3. Match by temporal proximity (created/completed around same time)
    // 4. Match by shared categories/tags in enrichment DB
    // 5. Return discovered edges with confidence weights
}

func (l *CrossProviderLinker) SuggestLinks() ([]Proposal, error) {
    // Discover links and wrap as proposals for user review
    // User approves → links stored in enrichment DB cross_refs
}
```

### MCP Tools for Cross-Provider Analysis

```
get_provider_overlap(provider_a, provider_b)
  → Tasks that appear related across two providers

get_unified_view(topic)
  → All tasks, notes, and references across all providers for a topic

get_provider_health_matrix()
  → Health status of all providers with sync freshness

suggest_cross_links()
  → Propose new cross-provider relationships for review
```

### Cross-Provider Analytics

```go
type CrossProviderMetrics struct {
    ProviderTaskCounts   map[string]int      // Tasks per provider
    OverlapScore         float64             // % of tasks with cross-refs
    OrphanedTasks        int                 // Tasks with no cross-refs
    CrossRefsByType      map[EdgeType]int    // Relationship type distribution
    SyncFreshness        map[string]time.Time // Last sync per provider
    ConflictCount        int                 // Unresolved cross-provider conflicts
}
```

---

## 10. Advanced Interaction Patterns

### 10.1 Automated Summaries

**Daily Digest:**
```
📊 ThreeDoors Daily Summary — March 6, 2026

Completed: 5 tasks (2 bugs, 2 features, 1 chore)
In Progress: 3 tasks
Blocked: 1 task ("API rate limit issue" — blocked since Mar 4)
Mood: Started energized → ended focused
Session: 2 sessions, 67 minutes total
Streak: Day 4 ✓

Top achievement: Closed the authentication refactor (deep-work, 3 days)
Watch out: "API rate limit issue" has been blocked for 2 days

Tomorrow suggestion: Start with the quick-win bug fixes while energized,
then tackle the remaining feature work before your afternoon slump.
```

**Weekly Retrospective:**
```
📈 Week of March 2-6, 2026

Velocity: 18 tasks completed (up 20% from last week)
Best day: Tuesday (6 completions, energized mood)
Worst day: Thursday (1 completion, tired mood)
Type mix: 40% features, 30% bugs, 20% chores, 10% research

Patterns noticed:
- You complete 3x more tasks when you start before 10 AM
- Bug fixes happen faster when you're slightly frustrated (channeling energy?)
- You've been avoiding research tasks for 2 weeks
- Your Tuesday-Wednesday combo is consistently your most productive

Recommendation: Consider scheduling a research task for next Tuesday
when you're typically energized and productive.
```

### 10.2 Prioritization Suggestions

The LLM can analyze tasks and suggest prioritization based on multiple signals:

```go
type PrioritySuggestion struct {
    TaskID     string   `json:"task_id"`
    Score      float64  `json:"score"`      // 0-100
    Rationale  string   `json:"rationale"`
    Factors    []Factor `json:"factors"`
}

type Factor struct {
    Name   string  `json:"name"`
    Weight float64 `json:"weight"`
    Value  float64 `json:"value"`
}
```

**Prioritization signals:**
- **Blocking score** — How many other tasks does this unblock?
- **Age** — How long has this been in the backlog?
- **Effort fit** — Does the effort level match the user's current energy?
- **Type fit** — Does the task type match the user's current mood preference?
- **Time-of-day fit** — Is this the right time for this effort level?
- **Streak impact** — Will completing this extend or break a streak?
- **Provider urgency** — Does the source provider have deadline signals?

### 10.3 Workload Analysis

```go
type WorkloadAnalysis struct {
    TotalTasks       int
    ByStatus         map[string]int
    ByEffort         map[string]int
    ByProvider       map[string]int
    EstimatedHours   float64          // Based on historical effort → time mapping
    OverloadRisk     float64          // 0-1 based on historical capacity
    RecommendedFocus []string         // Top 3 task IDs to focus on
    DeferCandidates  []string         // Tasks that could be deferred
}
```

LLM insight: "You have 23 active tasks across 3 providers. Based on your average completion rate of 4 tasks/day, this represents about 6 days of work. However, 8 of these are deep-work tasks that historically take you 3x longer. Adjusted estimate: 10 days. Consider deferring the 5 chore tasks to focus on the blocking features."

### 10.4 Burnout Detection & Prevention

Beyond the indicators in Section 8.5, advanced burnout detection includes:

- **Context switching cost** — How often does the user switch between task types within a session? High switching = cognitive load
- **Incomplete session ratio** — Sessions where the user quits without completing any task
- **Mood trajectory** — Is the overall mood trending downward over weeks?
- **Avoidance patterns** — Systematic avoidance of certain task types or providers
- **Session gap analysis** — Increasing gaps between sessions

```
⚠️ Burnout Alert: Your burnout risk score is 0.72 (elevated)

Signals:
- Completion rate down 35% over 7 sessions
- Average mood dropped from "focused" to "tired"
- You've avoided deep-work tasks for 12 days
- Last 3 sessions ended in under 15 minutes
- Context switching increased 40% (jumping between task types)

Suggestions:
1. Try a "quick wins only" session — just chores and small bugs
2. Take a 2-day break (your best recoveries follow 2-day gaps)
3. Pick ONE deep-work task and commit to just 25 minutes (Pomodoro)
4. Your Tuesday sessions are historically best — schedule focused time then
```

### 10.5 Focus Time Recommendations

```go
type FocusRecommendation struct {
    OptimalStartTime  string   `json:"optimal_start"`  // "9:00 AM"
    OptimalDuration   int      `json:"optimal_duration"` // minutes
    SuggestedTasks    []string `json:"suggested_tasks"`
    TaskOrder         []string `json:"task_order"`       // Recommended sequence
    BreakAfter        int      `json:"break_after"`      // Minutes before break
    Rationale         string   `json:"rationale"`
}
```

### 10.6 "What If" Scenario Modeling

LLMs can model hypothetical scenarios without changing data:

```
User: "What if I complete tasks A and B today?"

LLM response:
"If you complete tasks A and B today:
- You'll unblock 3 downstream tasks (C, D, E)
- Your streak extends to 5 days (personal record approaching: 11 days)
- Your weekly velocity reaches 20 tasks (22% above average)
- The 'Authentication Epic' would be 75% complete
- Estimated time: 2.5 hours based on your historical pace with these effort levels"
```

### 10.7 Context Switching Analysis

```go
type ContextSwitchAnalysis struct {
    SwitchesPerSession  float64              // Average type switches per session
    CostPerSwitch       float64              // Estimated minutes lost per switch
    MostExpensivePairs  [][2]string          // Task type pairs with highest cost
    Recommendation      string
    OptimalBatching     map[string][]string  // Suggested task groupings
}
```

---

## 11. Implementation Roadmap

### Phase 1: Read-Only MCP Server (Foundation)

**Scope:** Core resources + basic query tools

| Component | Description |
|---|---|
| `cmd/threedoors-mcp/` | MCP server binary (stdio mode) |
| MCP Resources | tasks, providers, session/current |
| MCP Tools | `query_tasks`, `get_task`, `list_providers`, `get_session` |
| Middleware | AuditLogger, RateLimiter |
| Tests | Contract tests for all endpoints |

**Value delivered:** Claude can see and query tasks. Users get AI-assisted task understanding immediately.

### Phase 2: Proposals + Enrichment

**Scope:** Proposal store, enrichment tools, TUI review

| Component | Description |
|---|---|
| `ProposalStore` | JSONL-backed proposal persistence |
| MCP Tools | `propose_enrichment`, `suggest_task`, `suggest_relationship` |
| TUI View | Proposal review with approve/reject |
| IntakeChannel | Interface for external task suggestions |
| Middleware | SchemaValidator, ProposalGovernance |

**Value delivered:** LLMs can suggest improvements. Users maintain full control.

### Phase 3: Analytics + Pattern Mining

**Scope:** PatternMiner, mood correlation, productivity analysis

| Component | Description |
|---|---|
| `PatternMiner` | Session data analysis engine |
| MCP Resources | analytics/mood, analytics/productivity, analytics/streaks |
| MCP Tools | `get_mood_correlation`, `get_productivity_profile`, `burnout_risk` |
| MCP Prompts | daily_summary, weekly_retrospective |

**Value delivered:** LLMs provide data-driven productivity insights.

### Phase 4: Relationship Graphs + Cross-Provider

**Scope:** TaskGraph, inference engine, cross-provider linking

| Component | Description |
|---|---|
| `TaskGraph` | Graph data structure with traversal |
| `RelationshipInferencer` | Infer edges from patterns |
| `CrossProviderLinker` | Discover cross-provider relationships |
| MCP Tools | `walk_graph`, `find_paths`, `get_clusters`, `suggest_cross_links` |

**Value delivered:** LLMs understand task relationships and cross-system dependencies.

### Phase 5: Advanced Interactions

**Scope:** Summaries, prioritization, burnout detection, what-if modeling

| Component | Description |
|---|---|
| Summary Generator | Daily/weekly digest templates |
| Priority Scorer | Multi-signal prioritization |
| Burnout Detector | Composite risk scoring |
| Scenario Modeler | What-if analysis without mutation |
| MCP Prompts | Full prompt template library |

**Value delivered:** LLMs become a personal productivity coach.

---

## 12. Party Mode Insights

### Round 1 Key Insights

- **Winston (Architect):** The existing `TaskProvider` interface requires zero changes for MCP. The MCP server is just another consumer of the same abstractions.
- **Mary (Analyst):** No task manager currently offers MCP + multi-provider aggregation. This is a genuine blue ocean competitive advantage.
- **John (PM):** Three core jobs-to-be-done: understand task landscape, get next-task suggestions, auto-enrich without effort. The read-only constraint builds trust.
- **Quinn (QA):** Rate limiting and chaos testing are day-one requirements, not afterthoughts.

### Round 2 Key Insights

- **Victor (Innovation):** The real disruption is *predictive* — personal productivity oracle, not just query tool. ThreeDoors becomes a platform.
- **Carson (Brainstorming):** Auto-generated weekly retrospectives transform dry metrics into coaching. This is the emotional connection point.
- **Bob (SM):** Burnout detection fills a genuine gap — no task manager actively monitors for overwork patterns.
- **Dr. Quinn (Problem Solver):** Optimistic concurrency control on proposals prevents conflicting enrichments from corrupting state.

### Round 3 Key Insights

- **Sophia (Storyteller):** Narrate productivity as a story, not a dashboard. "Chapter 1: The Monday Sprint" resonates more than "Completion Rate: 4.2/session."
- **Maya (Design Thinking):** Users don't want dashboards — they want *understanding*. The LLM should answer unasked questions.
- **Barry (Quick Flow):** Ship read-only queries first. Smallest thing that validates the assumption. Proposals are phase 2.
- **Paige (Tech Writer):** Document capability levels (Read → Propose → Analyze) for progressive onboarding.
- **Dr. Quinn:** Edge case: conflicting proposals need `BaseVersion` for optimistic concurrency. Stale proposals should be flagged, not auto-rejected.

### Cross-Cutting Themes

1. **Trust through constraint** — The read-only/proposal pattern isn't a limitation, it's the product's core trust proposition.
2. **Existing architecture is ready** — No breaking changes needed. Provider pattern, registry, aggregator, session tracking all support MCP naturally.
3. **Data richness is the moat** — Session JSONL + enrichment DB + multi-provider aggregation = unique analytical depth no competitor can match.
4. **Narrative over numbers** — LLMs should tell stories about productivity, not just report statistics.
5. **Progressive capability** — Start with reads, graduate to proposals, unlock analytics. Each phase delivers independent value.

---

## Appendix A: MCP Protocol Quick Reference

The Model Context Protocol (MCP) is a JSON-RPC-based protocol for LLM-tool integration:

- **Resources**: URI-addressable read-only data (`threedoors://tasks/123`)
- **Tools**: Callable functions with typed arguments and return values
- **Prompts**: Template strings that guide LLM behavior
- **Transport**: stdio (local) or SSE (remote)
- **Discovery**: Server declares capabilities; client queries them

ThreeDoors MCP server capabilities:

```json
{
  "capabilities": {
    "resources": { "subscribe": true, "listChanged": true },
    "tools": {},
    "prompts": { "listChanged": true }
  }
}
```

## Appendix B: Existing Code Leverage Map

| MCP Feature | Existing Code to Reuse | New Code Needed |
|---|---|---|
| Task queries | `TaskPool.Filter()`, `MultiSourceAggregator` | `TaskQueryEngine` with search scoring |
| Provider status | `HealthChecker`, `Registry` | MCP resource wrapper |
| Session data | `SessionTracker`, `MetricsWriter` | MCP resource wrapper |
| Proposals | — | `ProposalStore`, `ProposalReviewer` |
| Analytics | `PatternAnalyzer` (partial) | `PatternMiner` expansion |
| Graph | `DuplicateDetector` (partial) | `TaskGraph`, `RelationshipInferencer` |
| Audit trail | `SyncLog` pattern | `MCPAuditLog` |
| Rate limiting | — | `RateLimiterMiddleware` |
| Cross-provider | `MultiSourceAggregator`, `SourceRef` | `CrossProviderLinker` |

## Appendix C: Related Research

- [Task Source Expansion Research](task-source-expansion-research.md) — Provider integration patterns
- [Mood Correlation Research](mood-correlation.md) — Existing mood analysis findings
- [Sync Architecture Scaling](sync-architecture-scaling-research.md) — Multi-provider sync patterns
- [AI Tooling Findings](ai-tooling-findings.md) — Previous AI integration research
- [Self-Driving Development Pipeline](self-driving-development-pipeline.md) — Automation patterns
