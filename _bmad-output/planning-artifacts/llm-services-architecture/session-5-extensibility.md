# Party Mode Session 5: Extensibility & Provider Abstraction

**Date:** 2026-03-11
**Participants:** Architect, Dev, TEA/QA, PM, UX Designer
**Topic:** How do we support Claude Code as primary, but make it easy to add Gemini, Ollama, or future LLMs? What's the right abstraction layer?

---

## Round 1: The Provider Landscape

### Architect (Opening)

Let me catalog the CLI tools we need to support, their invocation patterns, and their capabilities:

**Claude Code CLI (`claude`)**
```bash
# Non-interactive mode (--print)
echo "Extract tasks from this text..." | claude --print
# With system prompt
echo "..." | claude --print --system-prompt "You are a task extractor..."
# With output format
echo "..." | claude --print --output-format json
# With allowed tools (can restrict capabilities)
echo "..." | claude --print --allowedTools "none"
```
- Auth: OAuth via `claude auth login` (already configured by user)
- Strengths: Best at structured extraction, strong instruction following
- Weaknesses: Requires internet, costs money

**Gemini CLI (`gemini`)**
```bash
# Basic invocation
echo "Extract tasks..." | gemini
# With context packaging
gemini --include-directories ./notes/ "Extract tasks from these notes"
# JSON output
echo "..." | gemini --output-format json
```
- Auth: OAuth via `gemini auth` (Google account)
- Strengths: Large context window, good at documents, free tier (50 Pro/day)
- Weaknesses: Less reliable at structured output than Claude

**Ollama CLI (`ollama`)**
```bash
# Basic invocation
echo "Extract tasks..." | ollama run llama3.2
# With system prompt
ollama run llama3.2 --system "You are a task extractor..." "Extract from: ..."
```
- Auth: None (local)
- Strengths: Private, free, no internet required, fast inference on Apple Silicon
- Weaknesses: Lower quality than cloud models, model download required

**Generic CLI (any tool)**
```bash
# Any tool that reads stdin, writes stdout
echo "Extract tasks..." | my-custom-llm --json
```

### Dev

The key differences across CLIs:

| Feature | Claude CLI | Gemini CLI | Ollama CLI | Generic |
|---------|-----------|-----------|------------|---------|
| Input method | stdin pipe | stdin pipe / args | args / stdin | stdin pipe |
| Output format flag | `--output-format` | `--output-format` | None | Varies |
| System prompt | `--system-prompt` | N/A | `--system` | Varies |
| Auth mechanism | OAuth (pre-configured) | OAuth (pre-configured) | None | Varies |
| Process startup | ~1-2s | ~1-2s | ~0.5s | Varies |
| Cost | Per-token | Free tier / per-token | Free | Varies |

The abstraction must handle these differences without leaking provider details to the service layer.

---

## Round 2: The CLIProvider Interface

### Architect

Here's my proposed abstraction:

```go
// Package: internal/intelligence/llm/

// CLIProvider wraps an LLM CLI tool for use as an LLMBackend.
// It handles the provider-specific invocation details (flags, input method,
// output parsing) behind the standard LLMBackend interface.
type CLIProvider struct {
    spec   CLISpec
    runner CommandRunner
}

// CLISpec defines how to invoke a specific CLI tool.
type CLISpec struct {
    Name          string        // "claude", "gemini", "ollama", "custom"
    Command       string        // binary name or path
    BaseArgs      []string      // always-present args (e.g., ["--print"] for claude)
    SystemPrompt  ArgTemplate   // how to pass system prompt (flag + value)
    OutputFormat  ArgTemplate   // how to request JSON output
    InputMethod   InputMethod   // stdin, arg, or file
    Timeout       time.Duration // per-invocation timeout
    ResponseParse ResponseParser // how to extract text from output
}

// ArgTemplate describes how to construct a CLI argument.
type ArgTemplate struct {
    Flag    string // e.g., "--system-prompt", "--output-format"
    Value   string // e.g., "json" (for --output-format json)
    Enabled bool   // whether this feature is supported
}

// InputMethod determines how the prompt is delivered.
type InputMethod int

const (
    InputStdin InputMethod = iota // pipe prompt to stdin
    InputArg                       // pass prompt as positional arg
    InputFile                      // write prompt to temp file, pass path
)

// ResponseParser extracts usable text from CLI output.
type ResponseParser interface {
    Parse(stdout []byte) (string, error)
}
```

### Dev

And the pre-built specs for known providers:

```go
func ClaudeCLISpec() CLISpec {
    return CLISpec{
        Name:     "claude",
        Command:  "claude",
        BaseArgs: []string{"--print"},
        SystemPrompt: ArgTemplate{
            Flag:    "--system-prompt",
            Enabled: true,
        },
        OutputFormat: ArgTemplate{
            Flag:    "--output-format",
            Value:   "json",
            Enabled: true,
        },
        InputMethod:   InputStdin,
        Timeout:       120 * time.Second,
        ResponseParse: &PlainTextParser{},
    }
}

func GeminiCLISpec() CLISpec {
    return CLISpec{
        Name:     "gemini",
        Command:  "gemini",
        BaseArgs: []string{},
        SystemPrompt: ArgTemplate{Enabled: false}, // Gemini CLI doesn't support system prompts
        OutputFormat: ArgTemplate{
            Flag:    "--output-format",
            Value:   "json",
            Enabled: true,
        },
        InputMethod:   InputStdin,
        Timeout:       120 * time.Second,
        ResponseParse: &PlainTextParser{},
    }
}

func OllamaCLISpec(model string) CLISpec {
    if model == "" {
        model = "llama3.2"
    }
    return CLISpec{
        Name:     "ollama",
        Command:  "ollama",
        BaseArgs: []string{"run", model},
        SystemPrompt: ArgTemplate{
            Flag:    "--system",
            Enabled: true,
        },
        OutputFormat: ArgTemplate{Enabled: false}, // Ollama CLI doesn't have output format flag
        InputMethod:   InputArg, // ollama run model "prompt" — prompt as positional arg
        Timeout:       120 * time.Second,
        ResponseParse: &PlainTextParser{},
    }
}

func CustomCLISpec(command string, args []string) CLISpec {
    return CLISpec{
        Name:        "custom",
        Command:     command,
        BaseArgs:    args,
        InputMethod: InputStdin,
        Timeout:     60 * time.Second,
        ResponseParse: &PlainTextParser{},
    }
}
```

### Architect

The `CLIProvider` then implements `LLMBackend`:

```go
func (p *CLIProvider) Name() string {
    return p.spec.Name
}

func (p *CLIProvider) Complete(ctx context.Context, prompt string) (string, error) {
    if prompt == "" {
        return "", ErrEmptyPrompt
    }

    args := p.buildArgs(prompt)
    var stdin string
    if p.spec.InputMethod == InputStdin {
        stdin = prompt
    }

    ctx, cancel := context.WithTimeout(ctx, p.spec.Timeout)
    defer cancel()

    stdout, stderr, err := p.runner.RunWithStdin(ctx, stdin, p.spec.Command, args...)
    if err != nil {
        return "", fmt.Errorf("%s CLI: %w (stderr: %s)", p.spec.Name, err, string(stderr))
    }

    return p.spec.ResponseParse.Parse(stdout)
}

func (p *CLIProvider) Available(ctx context.Context) bool {
    _, err := exec.LookPath(p.spec.Command)
    return err == nil
}

func (p *CLIProvider) buildArgs(prompt string) []string {
    args := make([]string, len(p.spec.BaseArgs))
    copy(args, p.spec.BaseArgs)

    if p.spec.SystemPrompt.Enabled && p.spec.SystemPrompt.Flag != "" {
        // System prompt is injected by the service layer
        // via a WithSystemPrompt option, not hardcoded here
    }

    if p.spec.OutputFormat.Enabled {
        args = append(args, p.spec.OutputFormat.Flag, p.spec.OutputFormat.Value)
    }

    if p.spec.InputMethod == InputArg {
        args = append(args, prompt)
    }

    return args
}
```

---

## Round 3: The Factory and Discovery

### Dev

Bringing it all together with a factory:

```go
// NewCLIBackend creates a CLI-based LLM backend from configuration.
func NewCLIBackend(cfg CLIConfig) (*CLIProvider, error) {
    var spec CLISpec
    switch cfg.Provider {
    case "claude-cli":
        spec = ClaudeCLISpec()
    case "gemini-cli":
        spec = GeminiCLISpec()
    case "ollama-cli":
        spec = OllamaCLISpec(cfg.Model)
    case "custom":
        spec = CustomCLISpec(cfg.Command, cfg.Args)
    default:
        return nil, fmt.Errorf("unknown CLI provider: %q", cfg.Provider)
    }

    // Override defaults from config
    if cfg.Command != "" {
        spec.Command = cfg.Command
    }
    if cfg.Timeout > 0 {
        spec.Timeout = cfg.Timeout
    }

    runner := &ExecRunner{} // real os/exec implementation
    return &CLIProvider{spec: spec, runner: runner}, nil
}
```

And the auto-discovery with fallback:

```go
// DiscoverBackend finds the best available LLM backend.
// Priority: user-configured > claude-cli > ollama-cli > ollama-http > claude-http
func DiscoverBackend(cfg Config) (LLMBackend, error) {
    // 1. User explicitly configured
    if cfg.Backend != "" {
        return newBackendFromConfig(cfg)
    }

    // 2. Auto-discover CLI tools
    cliOrder := []string{"claude", "gemini", "ollama"}
    for _, cmd := range cliOrder {
        if _, err := exec.LookPath(cmd); err == nil {
            switch cmd {
            case "claude":
                return NewCLIBackend(CLIConfig{Provider: "claude-cli"})
            case "gemini":
                return NewCLIBackend(CLIConfig{Provider: "gemini-cli"})
            case "ollama":
                return NewCLIBackend(CLIConfig{Provider: "ollama-cli"})
            }
        }
    }

    // 3. Fall back to HTTP backends
    if cfg.Ollama.Endpoint != "" {
        b := NewOllamaBackend(cfg.Ollama)
        if b.Available(context.Background()) {
            return b, nil
        }
    }

    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        claudeCfg := cfg.Claude
        claudeCfg.APIKey = apiKey
        return NewClaudeBackend(claudeCfg), nil
    }

    return nil, ErrBackendUnavailable
}
```

### TEA/QA

Testing strategy for CLI providers:

```go
// MockRunner implements CommandRunner for testing
type MockRunner struct {
    responses map[string]RunResult
}

type RunResult struct {
    Stdout []byte
    Stderr []byte
    Err    error
}

func (m *MockRunner) RunWithStdin(ctx context.Context, stdin string, name string, args ...string) ([]byte, []byte, error) {
    key := name + " " + strings.Join(args, " ")
    result, ok := m.responses[key]
    if !ok {
        return nil, nil, fmt.Errorf("unexpected command: %s", key)
    }
    return result.Stdout, result.Stderr, result.Err
}
```

Test scenarios:
1. **Happy path** — CLI returns valid JSON → parsed correctly
2. **CLI not found** — `Available()` returns false, `Complete()` returns helpful error
3. **CLI timeout** — Context cancelled → error with timeout message
4. **Non-zero exit** — CLI returns error code → error with stderr content
5. **Malformed output** — CLI returns non-JSON → parse error, retry
6. **Empty output** — CLI returns nothing → ErrEmptyResponse
7. **Large output** — CLI returns huge response → truncation or streaming

---

## Round 4: Adding a New Provider (The 5-Minute Test)

### PM

The "5-minute test": Can a developer add support for a new LLM CLI in 5 minutes? Let's walk through adding hypothetical "Llama CLI" support:

### Dev

**Step 1: Define the spec (2 minutes)**

```go
// In internal/intelligence/llm/specs.go
func LlamaCLISpec() CLISpec {
    return CLISpec{
        Name:     "llama",
        Command:  "llama",
        BaseArgs: []string{"chat", "--no-interactive"},
        SystemPrompt: ArgTemplate{
            Flag:    "--system",
            Enabled: true,
        },
        InputMethod:   InputStdin,
        Timeout:       60 * time.Second,
        ResponseParse: &PlainTextParser{},
    }
}
```

**Step 2: Register in factory (1 minute)**

```go
// In NewCLIBackend switch:
case "llama-cli":
    spec = LlamaCLISpec()
```

**Step 3: Add to discovery (1 minute)**

```go
// In DiscoverBackend:
cliOrder := []string{"claude", "gemini", "ollama", "llama"}
```

**Step 4: Add config type (1 minute)**

```yaml
llm:
  backend: "llama-cli"
  llama_cli:
    command: "llama"
    model: "llama-3.2-8b"
```

Total: 5 minutes, 4 files touched, zero changes to service layer.

### TEA/QA

And the test is just:

```go
func TestLlamaCLISpec(t *testing.T) {
    spec := LlamaCLISpec()
    if spec.Command != "llama" {
        t.Errorf("expected command 'llama', got %q", spec.Command)
    }
    // ... verify args, input method, etc.
}

func TestLlamaCLIIntegration(t *testing.T) {
    if _, err := exec.LookPath("llama"); err != nil {
        t.Skip("llama CLI not installed")
    }
    // ... send test prompt, verify response
}
```

### Architect

The 5-minute test passes. The abstraction is right-sized: specific enough to handle real CLI differences, generic enough that adding a new provider is mechanical.

---

## Round 5: Future Capabilities Without Over-Engineering

### Architect

Some future capabilities we should plan for but NOT implement now:

1. **Streaming responses** — Show partial results as they arrive. The `LLMBackend` interface would gain a `CompleteStream(ctx, prompt) (<-chan string, error)` method. CLI tools support this (line-by-line stdout). But our P0 services don't need it — extraction and enrichment are request-response.

2. **Conversation context** — Multi-turn conversations for iterative refinement ("No, I meant break it into smaller pieces"). Would require session state management. Not needed for P0 — each service call is stateless.

3. **Tool use** — Claude CLI supports `--allowedTools`. Could let the LLM call ThreeDoors tools (read tasks, check calendar) during processing. Powerful but complex. Defer.

4. **Model selection per service** — Use a fast/cheap model for enrichment, a powerful model for extraction. Would add a `model` field to service config. Easy to add later.

5. **Cost tracking** — Track API usage per service, show in `:llm-status`. Would require parsing CLI output for token counts. Nice-to-have.

### PM

The right approach: build the `CLISpec` model expressive enough to support these features later (the fields are there), but don't implement the logic until we need it. The spec is data — cheap to extend. The logic is code — expensive to maintain prematurely.

### Dev

One thing we SHOULD build now: the `threedoors llm status` command. It should show:

```
LLM Service Status
──────────────────
Backend: claude-cli (auto-detected)
Command: /usr/local/bin/claude
Available: Yes
Fallbacks: ollama-cli (available), ollama-http (unavailable)

Service Capabilities:
  Extract tasks:    Ready
  Enrich task:      Ready
  Break down task:  Ready
  Recommend tasks:  Not configured (requires pattern data)
```

This gives the user visibility into what's working and what's not, without them having to debug CLI tool issues manually.

---

## Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| S5-D1 | CLISpec struct for provider-specific invocation details | Declarative spec + generic CLIProvider executor | Per-provider classes with duplicated exec logic | DRY; adding a provider = defining a spec, not writing a class |
| S5-D2 | CLIProvider implements existing LLMBackend interface | Same interface, new implementation | New CLIBackend interface | Service layer doesn't care how backend works; interface is sufficient |
| S5-D3 | Pre-built specs for claude, gemini, ollama + generic custom | Known providers get optimized specs; custom for anything else | Only support known providers | Balance between convenience (known) and flexibility (custom) |
| S5-D4 | Auto-discovery with priority-ordered fallback chain | claude > gemini > ollama > HTTP backends | Manual-only configuration | Zero-config experience for users who have CLI tools installed |
| S5-D5 | 5-minute provider addition test | New provider = spec + factory case + discovery entry | Complex registration / plugin system | YAGNI; Go's compile-time type checking is sufficient |
| S5-D6 | Streaming, conversation, tool use deferred | CLISpec fields reserved but logic not implemented | Build now for future use | YAGNI; P0 services are request-response only |
| S5-D7 | `threedoors llm status` command from day one | Shows backend, availability, fallbacks, service readiness | No status/diagnostics | Essential for debugging; complements Epic 49 Doctor |
| S5-D8 | CommandRunner interface shared with dispatch package | Extract to common package or accept as constructor param | Duplicate interface | DRY; same pattern already exists in `internal/dispatch/` |
