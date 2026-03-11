# LLM CLI Services Architecture — Synthesis Document

**Date:** 2026-03-11
**Sessions:** 5 party mode rounds with PM, Architect, Dev, UX Designer, TEA/QA, SM
**Scope:** How ThreeDoors (as client) calls LLM CLIs (as service providers) for intelligent task services

---

## Executive Summary

ThreeDoors should invoke LLM CLI tools (Claude Code CLI, Gemini CLI, Ollama CLI) as subprocess-based service providers for intelligent task operations. This is **Direction 1** — ThreeDoors as CLIENT calling LLMs. It complements the existing **Direction 2** (Epic 24) where ThreeDoors is a SERVER exposing MCP tools for LLM agents.

The architecture extends the existing `LLMBackend` interface with CLI-based implementations using a declarative `CLISpec` model. A two-layer design separates **services** (what: extract, enrich, breakdown) from **backends** (how: which CLI tool). Auto-discovery and fallback chains ensure LLM features work with minimal configuration.

## Two Directions of Integration

```
Direction 1 (THIS RESEARCH): ThreeDoors → LLM CLIs
  ThreeDoors invokes claude/gemini/ollama CLIs via os/exec
  for task extraction, enrichment, breakdown, recommendation.
  ThreeDoors is the CLIENT. LLM CLIs are service providers.

Direction 2 (EXISTING — Epic 24): LLM Agents → ThreeDoors MCP Server
  Claude Desktop, Cursor, etc. connect via MCP (JSON-RPC).
  ThreeDoors is the SERVER. LLM agents are clients.

These are complementary. Direction 1 gives users intelligent task services.
Direction 2 gives LLM agents access to task data.
They compose: MCP tools can delegate to the LLM Service Layer (Direction 1).
```

## Prioritized Services

| Priority | Service | Description | Value | Complexity |
|----------|---------|-------------|-------|------------|
| **P0** | Task Extraction | Natural language → structured tasks from any text source | HIGH — feeds entire pipeline | MEDIUM |
| **P0** | Task Breakdown | Decompose large tasks into actionable subtasks | HIGH — reduces paralysis | LOW (extends Epic 14) |
| **P1** | Task Enrichment | Add context, tags, estimates to sparse tasks | MEDIUM — improves task quality | LOW |
| **P1** | Task Recommendation | Suggest which door to pick based on context | MEDIUM — addresses re-roll problem | MEDIUM |
| **P2** | Task Organization | Suggest groupings, priorities, sequences | LOW — fights SOUL.md | HIGH |
| **P2** | Pattern Analysis | Observe behavior, identify trends | LOW — not actionable | HIGH |

**Rejected for P0/P1:** Organization and Pattern Analysis fight SOUL.md's "Not a second brain" principle.

## Architecture

### Layer Diagram

```
┌─────────────────────────────────────────────────┐
│                  TUI / CLI Layer                 │
│  :extract, :enrich, :breakdown commands          │
│  threedoors extract/enrich/breakdown CLI          │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│           LLM Service Layer (NEW)                │
│  internal/intelligence/services/                 │
│                                                  │
│  TaskExtractor   — raw text → []ExtractedTask    │
│  TaskEnricher    — Task → enriched Task          │
│  TaskBreakdown   — Task → []subtasks             │
│  TaskRecommender — []Task → ranked []Task        │
│                                                  │
│  Each service owns prompt templates +            │
│  response parsing. Backend-agnostic.             │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│           LLM Backend Layer (EXTENDED)           │
│  internal/intelligence/llm/                      │
│                                                  │
│  LLMBackend interface (unchanged):               │
│    Name() string                                 │
│    Complete(ctx, prompt) (string, error)          │
│    Available(ctx) bool                            │
│                                                  │
│  Existing HTTP backends:                         │
│    ClaudeBackend (Anthropic API)                  │
│    OllamaBackend (Ollama HTTP API)                │
│                                                  │
│  New CLI backends:                               │
│    CLIProvider (generic, spec-driven)             │
│      - ClaudeCLISpec                              │
│      - GeminiCLISpec                              │
│      - OllamaCLISpec                              │
│      - CustomCLISpec (any stdin→stdout tool)      │
│                                                  │
└──────────────────────┬──────────────────────────┘
                       │
             ┌─────────▼─────────┐
             │   CommandRunner    │
             │   (os/exec)       │
             │   shared w/       │
             │   dispatch pkg    │
             └───────────────────┘
```

### Key Design Decisions

1. **Extend, don't replace.** The existing `LLMBackend` interface (`Complete(ctx, prompt) → (string, error)`) is sufficient for all P0 services. CLI backends implement the same interface. No new abstractions needed.

2. **Declarative CLISpec model.** Each CLI tool is described by a `CLISpec` struct (command, args, input method, output format, timeout). The generic `CLIProvider` executor handles all specs uniformly. Adding a new provider = defining a spec.

3. **Auto-discovery with fallback.** On startup, ThreeDoors checks which CLIs are available via `exec.LookPath`. Priority: user-configured → claude-cli → gemini-cli → ollama-cli → HTTP backends. This ensures LLM features work with zero configuration.

4. **Service layer separation.** Each service (extract, enrich, breakdown) owns its prompt templates and response parsing. Services are backend-agnostic — they work identically whether the backend is Claude CLI, Gemini CLI, or Ollama.

5. **Privacy tiers.** Local-first by default (Ollama). Cloud CLI backends are opt-in. No data leaves the machine unless the user explicitly configures a cloud backend.

## CLISpec Model

```go
type CLISpec struct {
    Name          string        // "claude", "gemini", "ollama", "custom"
    Command       string        // binary name or path
    BaseArgs      []string      // always-present args
    SystemPrompt  ArgTemplate   // how to pass system prompt
    OutputFormat  ArgTemplate   // how to request structured output
    InputMethod   InputMethod   // stdin, arg, or file
    Timeout       time.Duration // per-invocation timeout
    ResponseParse ResponseParser
}
```

Pre-built specs for known providers:
- `ClaudeCLISpec()` — `claude --print`, stdin input, `--system-prompt`, `--output-format json`
- `GeminiCLISpec()` — `gemini`, stdin input, `--output-format json`
- `OllamaCLISpec(model)` — `ollama run <model>`, arg input, `--system`
- `CustomCLISpec(cmd, args)` — any tool, stdin input, no special flags

## Task Extraction Pipeline

```
Source (Apple Notes, Obsidian, transcript, clipboard, file)
  │
  ▼ get raw text
extractFromText(ctx, text)
  │
  ▼ build prompt (extraction template + text)
LLMBackend.Complete(ctx, prompt)
  │
  ▼ parse JSON response → []ExtractedTask
Dedup against existing task pool
  │
  ▼
User review screen (TUI: toggle, edit, confirm)
  │
  ▼
Import selected tasks into TaskPool
```

**Input sources and how text is obtained:**
- **Apple Notes** — existing AppleScript bridge, fetch note body
- **Obsidian** — read .md file from vault path
- **File** — `os.ReadFile(path)`
- **Clipboard** — `pbpaste` (macOS)
- **Stdin** — `cat transcript.txt | threedoors extract`
- **Paste** — TUI text input area

**Input size limit:** 32KB for MVP. Chunking deferred to P2.

## User Experience

### Interaction Modes

| Mode | Trigger | MVP? | Example |
|------|---------|------|---------|
| Explicit commands | User types `:extract`, `:enrich`, `:breakdown` | Yes | `:extract` → paste text → review → import |
| Contextual suggestions | Heuristic rule shows suggestion line | P1 | Effort ≥ 4: "This task seems large. Break it down? (B)" |
| Ambient enrichment | Background processing when idle | P2 (local-only) | Tags appear on tasks automatically |

### UX Principles

1. **User-initiated** — LLM never acts without user triggering it
2. **Transparent** — always show what the LLM produced (before/after diff for enrichment)
3. **Reversible** — user can discard any LLM output
4. **Optional** — app works perfectly without any LLM
5. **Fast** — latency targets: extract <5s, enrich <3s, breakdown <8s
6. **Cancellable** — Esc cancels any in-progress LLM operation

### TUI Commands

| Command | Context | Keybind | Description |
|---------|---------|---------|-------------|
| `:extract` | Global | — | Extract tasks from text/file/clipboard |
| `:enrich` | Detail view | E | Enrich current task with LLM |
| `:breakdown` | Detail view | B | Break current task into subtasks |
| `:llm-status` | Global | — | Show LLM backend status |

### CLI Commands

```bash
threedoors extract --file notes.txt        # from file
threedoors extract --clipboard             # from clipboard
cat transcript.txt | threedoors extract    # from stdin
threedoors enrich <task-id>                # enrich specific task
threedoors breakdown <task-id>             # break down specific task
threedoors llm status                      # show backend status
```

## Configuration

```yaml
llm:
  # Backend selection (auto-detected if omitted)
  backend: "claude-cli"  # "ollama" | "claude" | "claude-cli" | "gemini-cli" | "ollama-cli" | "custom"

  # Existing HTTP backends
  ollama:
    endpoint: "http://localhost:11434"
    model: "llama3.2"
  claude:
    model: "claude-sonnet-4-20250514"

  # New CLI backends
  claude_cli:
    command: "claude"
    args: ["--print"]
    timeout: "120s"
  gemini_cli:
    command: "gemini"
    timeout: "120s"
  ollama_cli:
    command: "ollama"
    model: "llama3.2"
    timeout: "120s"
  custom:
    command: "/path/to/my-llm"
    args: ["--json"]
    timeout: "60s"

  # Decomposition output (existing, from Epic 14)
  decomposition:
    output_repo: "/path/to/repo"
    output_branch_prefix: "story/"
```

## Fallback Chain

```
User-configured backend
  ↓ (if unavailable)
claude CLI (if installed)
  ↓ (if unavailable)
gemini CLI (if installed)
  ↓ (if unavailable)
ollama CLI (if installed)
  ↓ (if unavailable)
Ollama HTTP API (if running)
  ↓ (if unavailable)
Claude HTTP API (if ANTHROPIC_API_KEY set)
  ↓ (if unavailable)
LLM features disabled (graceful degradation)
```

## MCP Composition (Direction 1 + Direction 2)

The LLM Service Layer is shared. MCP tools become thin wrappers:

```
MCP Client (Claude Desktop)
  │ MCP JSON-RPC: call tool "extract_tasks"
  ▼
ThreeDoors MCP Server (Epic 24)
  │ delegates to Service Layer
  ▼
TaskExtractor.ExtractFromText(ctx, text)
  │ calls configured LLM backend
  ▼
CLIProvider.Complete(ctx, prompt)
  │ executes: echo "prompt" | claude --print
  ▼
Returns extracted tasks via MCP response
```

## Testing Strategy

- **Unit tests**: Mock `CommandRunner` for all CLI providers. Test spec construction, arg building, response parsing.
- **Integration tests**: Skip if CLI not installed (`t.Skip`). Send test prompts, verify structured output.
- **Contract tests**: Verify all CLI providers satisfy `LLMBackend` interface contract.
- **Prompt tests**: Validate prompt templates produce parseable output with fixture data.
- **E2E tests**: TUI tests for `:extract` flow with mock backend.

## Epic Decomposition (Proposed)

### Epic: LLM CLI Services (suggested 6-8 stories)

| Story | Title | Priority | Depends On |
|-------|-------|----------|------------|
| X.1 | CLIProvider + CLISpec + CommandRunner abstraction | P0 | None |
| X.2 | Auto-discovery and fallback chain | P0 | X.1 |
| X.3 | TaskExtractor service + extraction prompt | P0 | X.1 |
| X.4 | Extraction TUI (`:extract` command + review screen) | P0 | X.3 |
| X.5 | Extraction CLI (`threedoors extract`) | P0 | X.3 |
| X.6 | TaskEnricher service + enrichment TUI | P1 | X.1 |
| X.7 | TaskBreakdown service (extend Epic 14 decomposer) | P1 | X.1 |
| X.8 | `threedoors llm status` command | P1 | X.1, X.2 |

**Dependency graph:** X.1 is foundation. X.2 builds on X.1. X.3-X.5 parallelize after X.1. X.6-X.8 parallelize after X.1.

## All Decisions (Consolidated)

| ID | Decision | Rationale |
|----|----------|-----------|
| S1-D1 | Task Extraction is P0 | LLMs understand natural language intent; pattern matching can't |
| S1-D2 | Task Breakdown is P0 | Extends existing Epic 14; reduces task paralysis |
| S1-D7 | Privacy-tiered LLM model (local default, cloud opt-in) | SOUL.md: "Local-First, Privacy-Always" |
| S1-D8 | All LLM services user-initiated | SOUL.md: reduce friction, not add automation |
| S2-D1 | Extend LLMBackend with CLI implementations (not new interface) | `Complete(ctx, prompt)` is sufficient; YAGNI on richer semantics |
| S2-D2 | Two-layer: Services (what) + Backends (how) | Swap backends without touching service logic |
| S2-D4 | Auto-discovery of CLI tools on startup | Reduces setup friction; fallback chain ensures resilience |
| S2-D6 | GenericCLIBackend for arbitrary CLI tools | Future-proofs for unknown tools |
| S2-D7 | MCP tools as thin wrappers around Service Layer | DRY; single implementation serves TUI and MCP |
| S3-D1 | All sources reduce to extractFromText(text) | Simpler architecture; source metadata carried separately |
| S3-D2 | User review required before import | Trust building; LLMs hallucinate; SOUL.md user control |
| S3-D5 | 32KB input size limit for MVP | YAGNI on chunking |
| S4-D1 | Explicit commands for MVP; contextual P1; ambient P2 | Incremental rollout reduces risk |
| S4-D3 | Always show before/after for enrichment | Transparency builds trust |
| S4-D7 | Latency targets: extract <5s, enrich <3s, breakdown <8s | Slow responses kill the "magical" feeling |
| S5-D1 | CLISpec struct for declarative provider description | Adding a provider = defining a spec (5-minute test) |
| S5-D5 | 5-minute provider addition test | New provider = spec + factory case + discovery entry |
| S5-D6 | Streaming, conversation, tool use deferred | P0 services are request-response only |
| S5-D7 | `threedoors llm status` from day one | Essential for debugging; complements Epic 49 Doctor |
