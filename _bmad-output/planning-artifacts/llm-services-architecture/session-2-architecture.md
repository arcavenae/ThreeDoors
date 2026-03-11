# Party Mode Session 2: Architecture — LLM CLI Service Integration

**Date:** 2026-03-11
**Participants:** Architect, Dev, PM, TEA/QA
**Topic:** What's the broadest, most elegant way to connect LLM CLI services? How does it compose with existing Epic 24 MCP work?

---

## Round 1: Two Directions of Integration

### Architect (Opening)

Let me establish the critical architectural distinction the user clarified:

```
Direction 1: ThreeDoors → LLM CLIs (NEW — this research)
  ThreeDoors calls out to claude/gemini/ollama CLIs as service providers.
  ThreeDoors is the CLIENT. LLM CLIs are SERVERS (subprocess invocation).

Direction 2: LLM Agents → ThreeDoors MCP Server (EXISTING — Epic 24)
  LLM agents (Claude Desktop, Cursor) call ThreeDoors via MCP protocol.
  ThreeDoors is the SERVER. LLM agents are CLIENTS (JSON-RPC over stdio/SSE).
```

These are **complementary, not competing**. Direction 1 gives ThreeDoors users intelligent task services. Direction 2 gives LLM agents access to ThreeDoors' task data.

The architecture question is: **what abstraction layer lets ThreeDoors invoke LLM services via their CLIs?**

### Dev

Let's look at what we have now in `internal/intelligence/llm/`:

```go
// Current LLMBackend interface (HTTP-based)
type LLMBackend interface {
    Name() string
    Complete(ctx context.Context, prompt string) (string, error)
    Available(ctx context.Context) bool
}
```

This is clean and narrow. Two implementations: `ClaudeBackend` (HTTP to Anthropic API) and `OllamaBackend` (HTTP to local Ollama API). Both use `net/http`.

For CLI backends, we need a parallel path. The `Complete(ctx, prompt) (string, error)` signature still works — we just implement it differently (subprocess instead of HTTP).

### Architect

I see three architecture options:

**Option A: Extend LLMBackend with CLI implementations**
Add `ClaudeCLIBackend`, `GeminiCLIBackend`, `OllamaCLIBackend` that implement the same `LLMBackend` interface but use `os/exec` instead of `net/http`.

**Option B: New CLIService abstraction**
Create a separate `CLIService` interface with richer semantics (streaming, conversation context, tool use) that CLI tools support natively.

**Option C: Unified LLMService with capability negotiation**
Single interface with optional capabilities — some backends support streaming, some support tool use, some support conversation history.

### PM

Option A is the simplest and follows our existing patterns. SOUL.md says "simplest thing that works." We already have a working interface. Why not just add more implementations?

### Dev

Option A works for the `Complete()` use case, but CLI tools have capabilities HTTP APIs don't expose easily:

1. **Claude Code CLI** has `--print` mode (non-interactive), reads from stdin, supports `--allowedTools`, and has rich tool-use capabilities
2. **Gemini CLI** has `--include-directories` for context packaging, outputs JSON
3. **Ollama CLI** is the simplest — `ollama run model "prompt"` or pipe stdin

The `Complete(prompt) → string` signature is sufficient for our P0 use cases (extraction, breakdown, enrichment). We don't need streaming, tool use, or conversation for those.

### TEA/QA

I favor Option A for now. We can always evolve to Option B or C later if we need richer capabilities. YAGNI applies here.

But I want to define the CLI contract clearly:

```go
// CLI invocation contract:
// 1. Prompt goes to stdin
// 2. Response comes from stdout
// 3. Errors go to stderr
// 4. Non-zero exit code = error
// 5. Timeout via context cancellation
```

This is the Unix philosophy — it works with any CLI tool, not just LLM-specific ones.

---

## Round 2: Concrete Architecture

### Architect

Here's my proposal. We extend the existing `LLMBackend` interface with CLI implementations, and add a new **service layer** on top that encodes specific use cases:

```
┌─────────────────────────────────────────────────┐
│                  TUI / CLI Layer                 │
│  `:extract`, `:enrich`, `:breakdown` commands    │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│              LLM Service Layer                   │
│  internal/intelligence/services/                 │
│                                                  │
│  TaskExtractor   — raw text → []Task             │
│  TaskEnricher    — Task → enriched Task          │
│  TaskBreakdown   — Task → []subtasks             │
│  TaskRecommender — []Task → ranked []Task        │
│                                                  │
│  Each service owns its prompt templates and       │
│  response parsing. Backend-agnostic.              │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│              LLM Backend Layer                   │
│  internal/intelligence/llm/                      │
│                                                  │
│  LLMBackend interface:                           │
│    Name() string                                 │
│    Complete(ctx, prompt) (string, error)          │
│    Available(ctx) bool                            │
│                                                  │
│  HTTP Backends (existing):                       │
│    ClaudeBackend    (Anthropic Messages API)      │
│    OllamaBackend    (Ollama HTTP API)             │
│                                                  │
│  CLI Backends (new):                             │
│    ClaudeCLIBackend  (claude --print via os/exec) │
│    GeminiCLIBackend  (gemini via os/exec)         │
│    OllamaCLIBackend  (ollama run via os/exec)     │
│    GenericCLIBackend (any CLI tool)               │
│                                                  │
└─────────────────────────────────────────────────┘
```

### Dev

I like this layered approach. The key insight is separating **what** (service layer — extract, enrich, breakdown) from **how** (backend layer — which LLM, via HTTP or CLI).

For the CLI backends, here's a concrete implementation sketch:

```go
// CLIBackend executes an LLM CLI tool as a subprocess.
type CLIBackend struct {
    name       string
    command    string   // e.g., "claude", "gemini", "ollama"
    args       []string // e.g., ["--print"], ["run", "llama3.2"]
    timeout    time.Duration
    runner     CommandRunner
}

// CommandRunner abstracts subprocess execution (same pattern as dispatch package).
type CommandRunner interface {
    Run(ctx context.Context, name string, args ...string) (stdout []byte, stderr []byte, err error)
}

func (c *CLIBackend) Complete(ctx context.Context, prompt string) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    // Pipe prompt to stdin, read response from stdout
    stdout, stderr, err := c.runner.RunWithStdin(ctx, prompt, c.command, c.args...)
    if err != nil {
        return "", fmt.Errorf("%s CLI error: %w (stderr: %s)", c.name, err, stderr)
    }

    text := strings.TrimSpace(string(stdout))
    if text == "" {
        return "", ErrEmptyResponse
    }
    return text, nil
}
```

### Architect

Note that we already have a `CommandRunner` interface in `internal/dispatch/` (Epic 22). We should reuse that pattern, not duplicate it. Move it to a shared package or accept it as a constructor parameter.

### PM

What about the existing `AgentService` in `internal/intelligence/agent_service.go`? It currently hardcodes backend selection:

```go
func newBackendFromConfig(cfg llm.Config) (llm.LLMBackend, error) {
    switch cfg.Backend {
    case "ollama", "":
        return llm.NewOllamaBackend(cfg.Ollama), nil
    case "claude":
        return llm.NewClaudeBackend(claudeCfg), nil
    }
}
```

We'd extend this to include CLI backends:

```go
case "claude-cli":
    return llm.NewClaudeCLIBackend(cfg.ClaudeCLI), nil
case "gemini-cli":
    return llm.NewGeminiCLIBackend(cfg.GeminiCLI), nil
case "ollama-cli":
    return llm.NewOllamaCLIBackend(cfg.OllamaCLI), nil
```

---

## Round 3: Configuration and Discovery

### Dev

How does the user configure which backend to use? Current config:

```yaml
llm:
  backend: "ollama"  # or "claude"
  ollama:
    endpoint: "http://localhost:11434"
    model: "llama3.2"
  claude:
    model: "claude-sonnet-4-20250514"
```

Extended for CLI backends:

```yaml
llm:
  backend: "claude-cli"  # "ollama" | "claude" | "claude-cli" | "gemini-cli" | "ollama-cli"

  # HTTP backends (existing)
  ollama:
    endpoint: "http://localhost:11434"
    model: "llama3.2"
  claude:
    model: "claude-sonnet-4-20250514"

  # CLI backends (new)
  claude_cli:
    command: "claude"         # binary name or full path
    args: ["--print"]         # additional CLI flags
    timeout: "120s"           # per-invocation timeout
  gemini_cli:
    command: "gemini"
    args: []
    timeout: "120s"
  ollama_cli:
    command: "ollama"
    args: ["run", "llama3.2"]
    timeout: "120s"
```

### Architect

I want to add auto-discovery. On startup, ThreeDoors can check which CLI tools are available:

```go
func DiscoverAvailableCLIs() []string {
    var available []string
    for _, cmd := range []string{"claude", "gemini", "ollama"} {
        if _, err := exec.LookPath(cmd); err == nil {
            available = append(available, cmd)
        }
    }
    return available
}
```

This powers:
1. **First-run setup** — "Found: claude, ollama. Which would you like to use?"
2. **Fallback chain** — If preferred CLI is unavailable, try next in chain
3. **`:doctor` integration** — ThreeDoors Doctor (Epic 49) can check LLM CLI availability

### TEA/QA

The fallback chain is important. Proposed order:

1. User's explicitly configured backend
2. `claude` CLI (if available — highest quality)
3. `ollama` CLI (if available — local/private)
4. `ollama` HTTP API (existing, works with running server)
5. Claude HTTP API (if API key configured)
6. No LLM available — disable LLM features gracefully

This means LLM services **always have a chance of working** if any backend is available, without requiring explicit configuration.

---

## Round 4: Composing with Epic 24 MCP

### Architect

The two directions compose naturally:

```
User's LLM Agent (Claude Desktop)
    │
    │ MCP Protocol (JSON-RPC)
    ▼
ThreeDoors MCP Server (Epic 24)
    │
    │ internal/ packages
    ▼
ThreeDoors Core (TaskPool, Adapters, etc.)
    │
    │ LLM Service Layer
    ▼
LLM CLI Backend (claude/gemini/ollama)
```

An LLM agent could, via MCP, ask ThreeDoors to extract tasks from a document. ThreeDoors' MCP server would delegate to the LLM Service Layer, which would call out to a CLI backend. The agent doesn't need to know which backend is used.

This creates a powerful composability:
- **MCP tool: `extract_tasks`** — accepts raw text, returns structured tasks
- **MCP tool: `enrich_task`** — accepts task ID, returns enriched version
- **MCP tool: `breakdown_task`** — accepts task ID, returns subtasks

These MCP tools become thin wrappers around the LLM Service Layer.

### Dev

We should be careful about recursion. If Claude Desktop calls ThreeDoors MCP `extract_tasks`, and ThreeDoors calls `claude` CLI for extraction, we're using Claude to call Claude. This works fine (they're different Claude instances), but we should:

1. Document this clearly
2. Allow the MCP server to specify which backend to use (avoid calling the same provider)
3. Consider cost implications

### PM

I don't think recursion is a real problem. The MCP server is for external agents. The CLI backend is for ThreeDoors-initiated services. They serve different users. Let's not over-engineer the separation.

---

## Round 5: The GenericCLIBackend Escape Hatch

### Architect

One more architectural piece: a `GenericCLIBackend` that works with ANY CLI tool that follows the stdin→stdout contract. This future-proofs us for tools we haven't heard of yet.

```yaml
llm:
  backend: "custom"
  custom:
    command: "/usr/local/bin/my-local-llm"
    args: ["--format", "json"]
    timeout: "60s"
```

```go
type GenericCLIBackend struct {
    name    string
    command string
    args    []string
    timeout time.Duration
    runner  CommandRunner
}
```

This is the adapter pattern applied to LLM tools. Any CLI that reads prompt from stdin and writes response to stdout can be a ThreeDoors LLM backend.

### TEA/QA

For the generic backend, we need a way to verify it works. Add a `--test` flag to LLM CLI configuration:

```bash
threedoors llm test  # sends a simple prompt to configured backend, shows response
```

This validates the CLI is installed, accessible, and produces usable output.

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| S2-D1 | Extend LLMBackend interface with CLI implementations | Option A: same interface, new implementations | Option B (new CLIService interface), Option C (capability negotiation) | `Complete(ctx, prompt) → (string, error)` is sufficient for P0 services; YAGNI on richer semantics |
| S2-D2 | Two-layer architecture: Services + Backends | Service layer (what) atop Backend layer (how) | Single monolithic layer | Separation lets us swap backends without touching service logic |
| S2-D3 | Subprocess via CommandRunner interface | Reuse dispatch package's pattern; inject via constructor | Direct `os/exec` calls | Testability; already proven pattern in codebase |
| S2-D4 | Auto-discovery of CLI tools on startup | `exec.LookPath` check for known CLIs | Manual-only configuration | Reduces setup friction; fallback chain ensures resilience |
| S2-D5 | Fallback chain: configured → claude-cli → ollama-cli → HTTP backends | Priority-ordered fallback | Single backend, fail if unavailable | Maximizes chance of LLM availability |
| S2-D6 | GenericCLIBackend for arbitrary CLI tools | Stdin→stdout contract; yaml-configurable | Only support known CLI tools | Future-proofs for tools we don't know about yet |
| S2-D7 | MCP tools as thin wrappers around Service Layer | MCP server delegates to same service layer | MCP implements its own LLM logic | DRY; single implementation serves both TUI and MCP clients |
| S2-D8 | CLI backend config in existing config.yaml | Extend `llm:` section with `claude_cli:`, `gemini_cli:`, etc. | Separate config file | Follows existing configuration pattern |
