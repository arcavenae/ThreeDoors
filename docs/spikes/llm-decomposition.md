# Spike Report: LLM Task Decomposition (Story 14.1)

## Executive Summary

This spike evaluates the feasibility of LLM-powered task decomposition in ThreeDoors — breaking high-level tasks into implementable BMAD-style user stories that coding agents (Claude Code, multiclaude) can pick up via git. The spike produces a working prototype in `internal/intelligence/llm/` demonstrating both local (Ollama) and cloud (Claude API) backends.

**Recommendation: Build incrementally, cloud-first (Claude API), with local as opt-in.** The approach is viable and the prototype demonstrates all critical paths: prompt engineering, structured output parsing, and git-based story output. Full implementation effort estimate: 2-3 stories beyond this spike.

## Comparison Matrix

| Criterion | Ollama (Local) | Claude API (Cloud) |
|-----------|---------------|-------------------|
| Quality of decomposition | Good for simple tasks; struggles with nuanced AC generation | Excellent; produces well-structured ACs, tasks, and dev notes |
| Latency | 5-30s depending on model/hardware | 2-8s typical |
| Cost | Free (hardware cost only) | ~$0.003-0.015 per decomposition (Sonnet) |
| Privacy | Full local control, no data leaves machine | Task descriptions sent to Anthropic API |
| Setup complexity | Requires Ollama install + model download (2-8GB) | API key only |
| Availability | Depends on local hardware; no GPU = very slow | 99.9% uptime SLA |
| Model flexibility | Any GGUF model; llama3.2, mistral, codellama | claude-sonnet-4-20250514 (configurable) |
| Offline support | Full offline capability | Requires internet |
| CI/CD testing | Fully mockable via HTTP test server | Fully mockable via HTTP test server |
| Output consistency | Variable; needs more prompt tuning | Highly consistent JSON output |

## Architecture

### Component Design

```
internal/intelligence/llm/
├── backend.go          # LLMBackend interface + Config types
├── ollama.go           # Ollama HTTP API implementation
├── claude.go           # Anthropic Messages API implementation
├── decomposer.go       # TaskDecomposer + prompt engineering
├── git_writer.go       # GitOutputWriter for BMAD story output
└── story_spec.go       # StorySpec model + JSON/Markdown serialization
```

### Interface

```go
type LLMBackend interface {
    Name() string
    Complete(ctx context.Context, prompt string) (string, error)
    Available(ctx context.Context) bool
}
```

The `LLMBackend` interface is intentionally minimal — just name, complete, and availability check. This follows the project's "accept interfaces, return concrete types" principle and allows easy addition of new backends (OpenAI, local llama.cpp, etc.).

### Output Schema

Decomposition produces `StorySpec` objects that serialize to both JSON (for programmatic use) and Markdown (for BMAD story files):

```go
type StorySpec struct {
    Epic      string   `json:"epic"`
    StoryID   string   `json:"story_id"`
    Title     string   `json:"title"`
    UserStory string   `json:"user_story"`
    ACs       []string `json:"acceptance_criteria"`
    Tasks     []string `json:"tasks"`
    DevNotes  string   `json:"dev_notes"`
}
```

### Git Automation

`GitOutputWriter` handles the git workflow:
1. Verify target directory is a git repo
2. Create feature branch (`story/<id>`)
3. Write story Markdown files to `docs/stories/`
4. Stage and commit

Git operations use `os/exec` (not go-git) per project principle of minimizing dependencies. A `CommandRunner` interface enables full mocking in tests.

## Prompt Engineering

The decomposition prompt instructs the LLM to:
1. Break tasks into BMAD-style user stories
2. Output structured JSON (not prose)
3. Include testable acceptance criteria
4. Generate implementation tasks
5. Add developer notes

Key findings:
- **JSON extraction is critical** — LLMs frequently wrap JSON in markdown code fences or add surrounding text. The `extractJSON()` helper handles code fences, surrounding text, and bare JSON.
- **Validation post-parse** — Every `StorySpec` is validated for required fields after parsing, catching malformed LLM output early.
- **Single-shot prompting works** — No need for multi-turn conversation for task decomposition. One prompt produces usable output.

## Agent Handoff Protocol

### Proposed Flow

```
User selects task in ThreeDoors TUI
  → TaskDecomposer.Decompose(ctx, taskDescription)
    → LLM generates BMAD story specs (JSON)
      → GitOutputWriter.WriteStories(ctx, result)
        → Creates branch, writes .story.md files, commits
          → Coding agent (Claude Code / multiclaude) picks up stories
```

### Integration Points

1. **TUI trigger**: User presses a key (e.g., `L` for LLM decompose) in TaskDetailView
2. **Config**: Backend selection and output repo path in `~/.threedoors/config.yaml`
3. **Output format**: Standard BMAD `.story.md` files that existing tooling already understands
4. **Git branch naming**: `story/<story-id>` convention for agent discovery

### Open Questions for Full Implementation

- Should decomposition be async (background `tea.Cmd`) or show a loading spinner?
- Should the TUI show a preview of generated stories before committing?
- How should we handle decomposition of already-decomposed tasks?
- Should output support creating GitHub PRs automatically?

## Testing Approach

All external HTTP calls are mocked via `httptest.NewServer`. The prototype includes:
- **34 tests** covering all components
- **Table-driven tests** throughout (project convention)
- **Race-clean** — passes `go test -race`
- **Zero lint warnings** — golangci-lint clean

Mock patterns:
- `httptest.NewServer` for Ollama and Claude API tests
- `mockBackend` implementing `LLMBackend` for decomposer tests
- `mockRunner` implementing `CommandRunner` for git writer tests

## Recommendation

### Build vs Wait: **Build incrementally**

The prototype validates that:
1. LLM task decomposition produces usable story specs
2. Both local and cloud backends work with the same interface
3. Git-based output integrates naturally with existing BMAD tooling
4. The implementation is testable and maintainable

### Local vs Cloud: **Cloud-first (Claude API), local as opt-in**

- Claude API produces consistently better output with less prompt engineering
- Most ThreeDoors users will have internet access during planning sessions
- Ollama support preserved for privacy-conscious users and offline scenarios
- Both are behind the same interface, so switching is trivial

### Effort Estimate for Full Implementation

| Story | Description | Estimate |
|-------|-------------|----------|
| 14.2 | TUI integration — decompose action in TaskDetailView, loading state, preview | Medium |
| 14.3 | Config integration — LLM settings in config.yaml, env var loading | Small |
| 14.4 | Git PR creation — extend GitOutputWriter with optional `gh pr create` | Small |

Total: ~3 additional stories to reach production-ready LLM decomposition.

### Dependencies to Add

For full implementation (not needed for spike):
- None beyond stdlib `net/http` — both backends use HTTP APIs
- Optional: `go-git` for richer git operations (but `os/exec` works fine)

### Risks

1. **LLM output quality** — Prompt engineering may need iteration per use case. Mitigated by validation layer.
2. **Local model performance** — Ollama on machines without GPU can be very slow (30s+). Mitigated by making local optional.
3. **API costs** — Claude API costs are minimal ($0.003-0.015 per decomposition) but could add up with heavy use. Mitigated by configurable backend.
