# Graph Database Backend Research — ThreeDoors

**Date:** 2026-03-15
**Type:** Blue-sky research (no stories/epics produced)
**Status:** Complete

---

## Executive Summary

This research explores two scopes: (1) using a graph database as a task storage backend for ThreeDoors, and (2) evolving that into a cross-project universal knowledge repository. The findings indicate that **an embedded graph database is a natural fit for ThreeDoors' existing multi-source architecture**, particularly for modeling task relationships (dependencies, parent/child, cross-source linking) that are currently awkward in flat YAML. The existing `TaskProvider` interface and `Registry` pattern accommodate a graph backend cleanly — it would be registered as another adapter alongside textfile, Jira, GitHub, etc.

For the broader knowledge repository vision, a graph model excels at answering cross-cutting questions ("what decisions led to this task?", "what research from Project A informs Project B?") that flat files fundamentally cannot. However, the SOUL.md principle **"Not a second brain (no knowledge graph, no linking, no tagging taxonomy)"** is directly in tension with Scope 2. Any knowledge repository evolution would need to carefully re-examine this principle, distinguishing between "the user manually curates a knowledge graph" (which ThreeDoors rejects) and "the system automatically builds relationship context behind the scenes" (which could enhance the three-doors UX without adding user-facing complexity).

**Recommendation:** Start with an embedded graph database (CayleyGraph or BadgerDB-backed custom solution) as an internal indexing layer that sits behind the existing providers — not replacing them, but enriching the task selection algorithm with relationship awareness. This preserves the flat-file simplicity users see while enabling powerful cross-source linking internally.

---

## Scope 1: Task Tracking Backend

### Current Architecture

ThreeDoors uses a well-designed provider/adapter pattern:

- **`TaskProvider` interface** (`internal/core/provider.go`): 7-method contract (Name, LoadTasks, SaveTask, SaveTasks, DeleteTask, MarkComplete, Watch, HealthCheck)
- **`Registry`** (`internal/core/registry.go`): Factory-based adapter registration with runtime discovery
- **`MultiSourceAggregator`** (`internal/core/aggregator.go`): Merges tasks from multiple providers, routes writes back to originating provider via `taskOrigins` map
- **`SourceRef`** (`internal/core/task.go`): Links a task to its identity in a specific provider (provider name + native ID)
- **Storage backends**: textfile (YAML), Apple Notes, Apple Reminders, Obsidian, Jira, Linear, GitHub Issues, ClickUp, Todoist
- **Session data**: JSONL append-only log (`sessions.jsonl`) for analytics
- **Sync**: WAL-based sync queue (`sync-queue.jsonl`), vector clocks and field-level versioning on Task struct

### What a Graph DB Would Add

The Task struct already has graph-like fields that are awkward in flat storage:

| Field | Current Representation | Graph Representation |
|-------|----------------------|---------------------|
| `ParentID *string` | String FK, manually resolved | `CHILD_OF` edge to parent node |
| `DependsOn []string` | String array of IDs, no validation | `DEPENDS_ON` edge with ordering |
| `SourceRefs []SourceRef` | Embedded array | `SOURCED_FROM` edges to provider nodes |
| `DevDispatch` | Embedded struct | `DISPATCHED_TO` edge to agent node |
| Cross-source links | Not modeled | `SAME_AS` edges across providers |
| Task history | Separate JSONL file | `COMPLETED_IN` edges to session nodes |

**Key wins:**
1. **Dependency traversal**: "What tasks are blocked by this one?" is O(1) with edges, requires full scan with flat files
2. **Cross-source identity**: A Jira ticket and a GitHub issue about the same thing can be linked with a `SAME_AS` edge — currently handled by SourceRef matching which breaks for unrelated-but-related items
3. **History queries**: "Show me all tasks I completed this week from Linear" requires joining sessions.jsonl with task YAML — trivial in a graph
4. **Cycle detection**: DependsOn cycles are a correctness hazard with string arrays; graph DBs provide native cycle detection

### Graph Database Options for Go

| Database | Type | Go Client | Embedded? | License | Maturity | Notes |
|----------|------|-----------|-----------|---------|----------|-------|
| **CayleyGraph** | Triple store | Native Go | Yes | Apache 2.0 | Medium | Written in Go, embeds cleanly. Backed by BoltDB/Badger/Mongo. Gremlin + GraphQL query. |
| **BadgerDB + custom** | KV store | Native Go (dgraph-io/badger) | Yes | Apache 2.0 | High | Build adjacency lists on top. Maximum control, minimum dependency. |
| **Dgraph** | Native graph | `dgo` client | No (server) | Apache 2.0 | High | Overkill for local-first. Requires separate process. |
| **Neo4j** | Native graph | `neo4j/neo4j-go-driver` | No (server) | GPL/Commercial | Very High | Industry standard but violates local-first (needs JVM, server process). |
| **SurrealDB** | Multi-model | `surrealdb/surrealdb.go` | Embeddable (new) | BSL | Medium | Interesting multi-model (doc + graph + SQL). Go embedding is new/experimental. |
| **BoltDB/bbolt** | KV store | Native Go | Yes | MIT | Very High | Battle-tested. Would need custom graph layer. |
| **SQLite + recursive CTEs** | Relational | `modernc.org/sqlite` (pure Go) | Yes | Public domain | Very High | Surprisingly capable for graphs via WITH RECURSIVE. No CGO needed with modernc. |

### Evaluation Against SOUL.md Principles

- **Local-first, privacy-always**: Eliminates Dgraph, Neo4j (server processes). CayleyGraph, BadgerDB, BoltDB, SQLite all embed cleanly.
- **No cloud dependencies**: All embedded options satisfy this.
- **Simplicity**: CayleyGraph adds a dependency but it's pure Go. BadgerDB or BoltDB + custom graph layer is more code but fewer deps.
- **Meet users where they are**: Graph DB is internal — users still see YAML files and their existing tools. The graph is an optimization layer, not a user-facing change.

### Recommended Approach: CayleyGraph as Internal Index

CayleyGraph is the strongest fit because:

1. **Pure Go, embeddable** — no CGO, no external process
2. **Multiple storage backends** — can use BoltDB (lightweight) or BadgerDB (higher write throughput)
3. **Gremlin query language** — well-documented, good for traversals
4. **Graph-native operations** — shortest path, cycle detection, reachability are built in
5. **Apache 2.0 license** — no viral concerns

**Architecture:**
```
┌──────────────────────────────────────────────┐
│  MultiSourceAggregator                        │
│  (existing — merges providers)                │
├──────────────────────────────────────────────┤
│        │                    │                  │
│  [textfile]  [jira]  [github]  [linear] ...   │
│        │                    │                  │
│        └──────────┬─────────┘                  │
│                   ▼                            │
│        GraphIndexProvider                      │
│        (new — indexes relationships)           │
│        ┌─────────────────┐                     │
│        │   CayleyGraph    │                    │
│        │   (BoltDB)       │                    │
│        └─────────────────┘                     │
└──────────────────────────────────────────────┘
```

The `GraphIndexProvider` would:
- **Not replace** existing providers — they remain source-of-truth for task CRUD
- **Index relationships** when tasks are loaded: ParentID → `CHILD_OF`, DependsOn → `DEPENDS_ON`, SourceRefs → `SOURCED_FROM`
- **Answer graph queries**: "blocked tasks", "dependency chains", "cross-source links"
- **Feed the selection algorithm**: When picking 3 doors, prefer tasks without unresolved dependencies

### Migration Path

1. **Phase 1: Read-only index** — Graph indexes existing flat-file data. No writes go through the graph. If graph is unavailable, everything works as before. Zero user impact.
2. **Phase 2: Relationship CRUD** — Add/remove dependencies through the graph. Sync edges back to Task struct fields (ParentID, DependsOn). Providers still own task persistence.
3. **Phase 3: Cross-source linking** — `SAME_AS` edges between tasks from different providers. This is new functionality that flat files can't model well.
4. **Phase 4: Session integration** — Index session JSONL events as graph nodes. Enable history traversal queries.

### Provider Pattern Compatibility

The `GraphIndexProvider` does NOT need to implement `TaskProvider`. It's a separate concern:

```go
// GraphIndex provides relationship queries over tasks.
// It does not own task persistence — providers do that.
type GraphIndex interface {
    // IndexTask adds/updates a task node and its relationship edges.
    IndexTask(task *core.Task) error

    // DependencyChain returns ordered IDs from task to all transitive deps.
    DependencyChain(taskID string) ([]string, error)

    // BlockedBy returns tasks that this task transitively depends on.
    BlockedBy(taskID string) ([]string, error)

    // RelatedAcrossProviders returns tasks linked via SAME_AS edges.
    RelatedAcrossProviders(taskID string) ([]string, error)

    // HasCycle detects if adding a dependency would create a cycle.
    HasCycle(fromID, toID string) (bool, error)
}
```

This interface composes with the existing `TaskProvider` and `Registry` without modifying either. The aggregator would gain an optional `GraphIndex` field.

### Rejected Alternatives (Scope 1)

| Alternative | Reason for Rejection |
|-------------|---------------------|
| **Neo4j** | Requires JVM + server process. Violates local-first. Massively over-provisioned for single-user task data (~100-1000 tasks). |
| **Dgraph** | Requires separate gRPC server. Same local-first violation. Its distributed design is for multi-node clusters, not laptops. |
| **SQLite with recursive CTEs** | Could work. But graph queries become verbose SQL. Loses the graph-native operations (cycle detection, shortest path). Would be a reasonable fallback if CayleyGraph proves problematic. |
| **Replace all providers with graph** | Violates "meet users where they are." Users want their Jira/GitHub/textfile data in those tools. The graph should index, not own. |
| **Custom graph on BoltDB** | Maximum control but significant implementation effort. Re-inventing traversal algorithms, cycle detection, etc. CayleyGraph provides this for free. |
| **SurrealDB embedded** | Go embedding is experimental (as of mid-2026). Multi-model is appealing but the immaturity risk is too high for a core storage layer. |

---

## Scope 2: General Knowledge Repository

### Vision

A graph database as a universal knowledge store that connects information across multiple projects, enriching task selection with cross-cutting context:

```
                    ┌─────────────┐
                    │   Calendar   │
                    │  (events)    │
                    └──────┬──────┘
                           │ BLOCKS_TIME_FOR
    ┌──────────┐    ┌──────┴──────┐    ┌───────────┐
    │ Research  │────│    Task     │────│   Story   │
    │ Artifact  │    │   (node)   │    │   File    │
    └──────────┘    └──────┬──────┘    └─────┬─────┘
     INFORMS_DECISION      │ PART_OF         │ IMPLEMENTS
                    ┌──────┴──────┐    ┌─────┴─────┐
                    │   Project   │    │    PRD     │
                    │  (node)     │    │  (node)    │
                    └──────┬──────┘    └─────┬─────┘
                           │ DECIDED_IN      │ SHAPED_BY
                    ┌──────┴──────┐    ┌─────┴─────┐
                    │  Decision   │    │ Meeting   │
                    │  (BOARD.md) │    │ Transcript│
                    └─────────────┘    └───────────┘
```

### Node Types

| Node Type | Properties | Source |
|-----------|-----------|--------|
| **Task** | id, text, status, effort, type, timestamps | TaskProvider (existing) |
| **Project** | name, path, status, description | Config files / user declaration |
| **Story** | id, title, status, epic, acceptance_criteria | `docs/stories/*.story.md` |
| **Epic** | number, title, status, stories[] | `docs/prd/epics-and-stories.md` |
| **Decision** | id, adopted, rejected[], rationale, date | `docs/decisions/BOARD.md` |
| **PRD** | title, version, sections[] | `docs/prd/` files |
| **Architecture** | title, patterns[], components[] | `docs/architecture/` files |
| **PR** | number, title, status, branch, author | GitHub API |
| **Research** | title, findings, recommendations | `_bmad-output/planning-artifacts/` |
| **Meeting** | date, participants, transcript, decisions[] | External (transcript service) |
| **CalendarEvent** | title, start, end, recurrence | Calendar integration (existing `internal/calendar/`) |
| **Person** | name, role, projects[] | User declaration |
| **Sprint** | id, start, end, goals[], stories[] | Sprint planning docs |

### Edge Types

| Edge | From → To | Semantics |
|------|-----------|-----------|
| `DEPENDS_ON` | Task → Task | Completion dependency |
| `CHILD_OF` | Task → Task | Parent-child decomposition |
| `SAME_AS` | Task → Task | Cross-provider identity |
| `PART_OF` | Task/Story → Epic/Project | Containment |
| `IMPLEMENTS` | Story → PRD Section | Traceability |
| `DECIDED_IN` | Decision → Meeting/PR | Provenance |
| `INFORMS` | Research → Decision/Story | Knowledge flow |
| `BLOCKS_TIME_FOR` | CalendarEvent → Task | Scheduling context |
| `RELATED_TO` | Any → Any | Weak semantic link |
| `CREATED_BY` | PR/Decision → Person | Attribution |
| `REVIEWED_BY` | PR → Person | Review tracking |
| `SHAPED_BY` | PRD → Meeting/Research | Input provenance |
| `COMPLETED_IN` | Task → Session | Completion history |
| `FORKED_FROM` | Task → Task | Fork lineage |

### Query Patterns — Questions This Answers

**Cross-project context:**
- "What tasks across all my projects are blocked?" → `MATCH (t:Task)-[:DEPENDS_ON]->(blocker:Task) WHERE blocker.status != 'complete'`
- "What did I decide about X in Project A that affects Project B?" → Traverse `INFORMS` and `DECIDED_IN` edges across project boundaries
- "Show me all tasks related to the Linear integration, across planning docs, PRs, and implementation" → Fan out from any node matching "Linear" through all edge types

**Schedule-informed prioritization:**
- "I have a 2-hour block this afternoon — what fits?" → Cross-reference CalendarEvent gaps with Task effort estimates
- "What should I review before tomorrow's meeting?" → Traverse `BLOCKS_TIME_FOR` edges from tomorrow's calendar events to incomplete tasks
- "Which tasks have context that will decay if I don't do them this week?" → Time-based relevance scoring using `DeferUntil`, meeting dates, sprint boundaries

**Research reuse:**
- "Has anyone researched graph databases before?" → Traverse `Research` nodes across all projects (this document would be a node!)
- "What research informed the decision to use the provider pattern?" → `SHAPED_BY` / `INFORMS` edges from architecture docs to research artifacts

**Meeting → action:**
- "What decisions were made in last Tuesday's meeting?" → `DECIDED_IN` edges from meeting transcript node
- "What stories came out of the brainstorming session?" → `SHAPED_BY` edges from meeting to story nodes

**History and patterns:**
- "What types of tasks do I complete fastest?" → Aggregate `COMPLETED_IN` edges with duration and task type
- "Which projects have the most stale tasks?" → Count tasks with old `UpdatedAt` per project node

### Architectural Considerations

#### Data Locality and Privacy

Per SOUL.md's "Local-first, privacy-always":

- **Embedded graph only** — CayleyGraph with BoltDB/BadgerDB backend, stored in `~/.threedoors/graph.db`
- **No cloud sync of the graph** — individual providers handle their own sync (Jira syncs to Jira, GitHub to GitHub)
- **Graph is a derived index** — can be rebuilt from source providers at any time. If `graph.db` is deleted, reindex from flat files + provider APIs
- **Cross-project graphs are local** — multiple project graphs can be federated via a top-level graph that links project subgraphs, all on the user's machine
- **Optional export** — user can export subgraphs for sharing, but never automatic

#### Ingestion Architecture

```
┌─────────────────────────────────────┐
│         Graph Ingestion Layer        │
├─────────────────────────────────────┤
│  FileWatcher    → story files,       │
│                   decisions,         │
│                   research artifacts │
│  ProviderSync   → tasks from all     │
│                   TaskProviders      │
│  CalendarReader → events from        │
│                   internal/calendar/ │
│  GitHubPoller   → PRs, issues       │
│  TranscriptParser → meeting notes   │
└──────────────────┬──────────────────┘
                   ▼
            ┌──────────────┐
            │  CayleyGraph  │
            │  (BoltDB)     │
            └──────────────┘
```

Most of this infrastructure already exists in ThreeDoors:
- `internal/calendar/` has CalDAV, AppleScript, and ICS readers
- `internal/adapters/` has file watchers (Obsidian, textfile)
- The Watch() channel on TaskProvider already emits ChangeEvents

The graph ingestion layer would subscribe to these existing channels and index changes as they arrive.

#### Schema Evolution

Graph databases are inherently schema-flexible — adding new node/edge types doesn't require migrations. This is a significant advantage over relational schemas for a system that will evolve as new data sources are integrated.

### SOUL.md Tension: "Not a second brain"

SOUL.md explicitly states ThreeDoors is "Not a second brain (no knowledge graph, no linking, no tagging taxonomy)." This is the most important constraint for Scope 2.

**Resolution path:**
- The knowledge graph is **internal infrastructure**, not a user-facing feature
- Users never see nodes, edges, or query results directly
- The graph informs the **three-doors selection algorithm** — "here are three things you could do right now" gets smarter
- Think of it as the plumbing, not the faucet. The user still sees three doors. But the system is better at picking which three doors to show.
- If we added a UI for exploring the graph, that WOULD violate SOUL.md. The graph must remain invisible.

**Boundary to maintain:** The graph is a **behind-the-scenes optimization**. The moment it becomes something the user needs to curate, tag, or browse, it's violating ThreeDoors' soul. The graph should be self-maintaining — auto-indexed from existing data sources, never requiring user input to maintain relationships.

### Rejected Alternatives (Scope 2)

| Alternative | Reason for Rejection |
|-------------|---------------------|
| **Hosted graph service (Neo4j Aura, Amazon Neptune)** | Violates local-first and privacy-always. Single-user data doesn't need cloud infrastructure. |
| **User-facing knowledge graph UI** | Directly contradicts SOUL.md "not a second brain." The graph must be invisible plumbing. |
| **Replacing flat files with graph-only storage** | Users want to see their tasks in YAML, their decisions in Markdown. The graph supplements, never replaces. |
| **Full RDF/SPARQL stack** | Overkill for the data volumes involved (~1000s of nodes, not millions). SPARQL's learning curve adds maintenance burden. CayleyGraph supports Gremlin which is more practical. |
| **LLM-powered auto-linking** | Tempting but adds latency, cost, and non-determinism. Start with explicit relationships (file references, task IDs) before considering semantic similarity. Could be a Phase 3+ enhancement. |
| **Obsidian-style user-created links** | This IS the "second brain" anti-pattern. ThreeDoors should infer relationships, not ask users to create them. |

---

## Recommended Approach

### Phase 1: Graph-Indexed Task Relationships (Scope 1, Minimal)

- Add CayleyGraph as a Go dependency
- Create `internal/graph/index.go` implementing `GraphIndex` interface
- On task load, index ParentID, DependsOn, and SourceRefs as edges
- Expose cycle detection for the dependency creation UI
- Use graph traversal to surface "blocked tasks" in the three-doors algorithm
- **Effort:** 1 epic, ~4-5 stories
- **Risk:** Low — read-only index, providers remain authoritative

### Phase 2: Cross-Source Linking (Scope 1, Extended)

- Add `SAME_AS` edge creation when tasks from different providers match (heuristic: same title within time window, or explicit user linking)
- Surface cross-source context in task detail view ("This task also appears in Jira as PROJ-123")
- **Effort:** 2-3 stories
- **Risk:** Low-medium — heuristic matching may produce false positives

### Phase 3: Knowledge Ingestion (Scope 2, Initial)

- Index story files, BOARD.md decisions, and research artifacts as graph nodes
- Create `IMPLEMENTS`, `DECIDED_IN`, `INFORMS` edges from file references
- Feed cross-cutting context to the three-doors selection: tasks with recent decision context get priority
- **Effort:** 1 epic, ~5-6 stories
- **Risk:** Medium — scope of "what to index" can creep. Strict boundary: only index files in known locations with known formats.

### Phase 4: Calendar + Schedule Integration (Scope 2, Extended)

- Leverage existing `internal/calendar/` infrastructure
- Create `BLOCKS_TIME_FOR` edges between events and tasks
- Time-aware door selection: "you have 30 minutes before your next meeting, here are 3 quick tasks"
- **Effort:** 3-4 stories (calendar infra already exists)
- **Risk:** Medium — schedule integration is high-value but easy to over-engineer

### Phase 5: Cross-Project Federation (Scope 2, Future)

- Support multiple project graphs linked by a root graph
- Research from Project A appears when relevant to Project B tasks
- **Effort:** Unknown — depends on multi-project UX decisions
- **Risk:** High — this is where "second brain" pressure is strongest. Must be auto-inferred, never user-curated.

---

## Open Questions

1. **CayleyGraph maintenance status** — Last significant release was 2022. Is the project actively maintained enough for a production dependency? Fallback: SQLite with recursive CTEs.
2. **Graph size and performance** — For a single user with 5 projects, ~1000 tasks, and ~500 documents, CayleyGraph on BoltDB should be negligible overhead. But should be benchmarked.
3. **SOUL.md amendment** — Does the project owner want to formally amend the "not a second brain" line to clarify that an internal knowledge index is acceptable? Or does that line intentionally prevent this entire direction?
4. **Multi-project UX** — ThreeDoors currently operates on a single task file / provider set. How would multi-project federation surface in the three-doors UI without overwhelming the user?
5. **LLM enhancement** — In Phase 5+, could an LLM infer relationships between documents that don't explicitly reference each other? (e.g., "this research about graph databases is relevant to the story about task dependencies"). Deferred to future research.

---

## Next Steps

- **No stories or epics created** — this is research only
- Supervisor should review findings and decide whether to proceed with Phase 1
- If proceeding, a party mode session to design the `GraphIndex` interface and CayleyGraph integration would be appropriate
- The SOUL.md question (Q3 above) should be resolved before any Scope 2 work
