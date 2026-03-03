package intelligence

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

// AgentService orchestrates LLM task decomposition and git output writing.
// It bridges the LLM spike code with the TUI, providing a single entry point
// for decomposing a task description into BMAD story specs and writing them
// to a target git repository.
type AgentService struct {
	decomposer *llm.TaskDecomposer
	writer     *llm.GitOutputWriter
	cfg        llm.Config
}

// NewAgentService creates an AgentService from the given LLM config.
// Returns an error if the config is invalid (e.g., missing output repo).
func NewAgentService(cfg llm.Config) (*AgentService, error) {
	if cfg.Output.OutputRepo == "" {
		return nil, fmt.Errorf("new agent service: output_repo is required")
	}

	backend, err := newBackendFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("new agent service: %w", err)
	}

	decomposer := llm.NewTaskDecomposer(backend)
	writer := llm.NewGitOutputWriter(cfg.Output.OutputRepo, cfg.Output.OutputBranchPrefix, nil)

	return &AgentService{
		decomposer: decomposer,
		writer:     writer,
		cfg:        cfg,
	}, nil
}

// NewAgentServiceWithDeps creates an AgentService with injected dependencies for testing.
func NewAgentServiceWithDeps(decomposer *llm.TaskDecomposer, writer *llm.GitOutputWriter) *AgentService {
	return &AgentService{
		decomposer: decomposer,
		writer:     writer,
	}
}

// DecomposeAndWrite decomposes a task description into stories and writes them to git.
func (s *AgentService) DecomposeAndWrite(ctx context.Context, taskDescription string) (*llm.DecompositionResult, error) {
	taskDescription = strings.TrimSpace(taskDescription)
	if taskDescription == "" {
		return nil, fmt.Errorf("decompose: task description must not be empty")
	}

	result, err := s.decomposer.Decompose(ctx, taskDescription)
	if err != nil {
		return nil, fmt.Errorf("decompose task: %w", err)
	}

	if len(result.Stories) == 0 {
		return nil, fmt.Errorf("decompose: LLM returned no stories")
	}

	if err := s.writer.WriteStories(ctx, result); err != nil {
		return nil, fmt.Errorf("write stories: %w", err)
	}

	return result, nil
}

// newBackendFromConfig creates the appropriate LLM backend based on config.
func newBackendFromConfig(cfg llm.Config) (llm.LLMBackend, error) {
	switch cfg.Backend {
	case "ollama", "":
		return llm.NewOllamaBackend(cfg.Ollama), nil
	case "claude":
		claudeCfg := cfg.Claude
		if claudeCfg.APIKey == "" {
			claudeCfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if claudeCfg.APIKey == "" {
			return nil, fmt.Errorf("claude backend requires ANTHROPIC_API_KEY environment variable")
		}
		return llm.NewClaudeBackend(claudeCfg), nil
	default:
		return nil, fmt.Errorf("unknown LLM backend: %q", cfg.Backend)
	}
}
