package llm

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// DefaultCLITimeout is the default timeout for CLI-based LLM calls.
const DefaultCLITimeout = 120 * time.Second

// CLIProvider implements LLMBackend by invoking an LLM CLI tool as a subprocess.
type CLIProvider struct {
	spec   CLISpec
	runner CLIRunner
}

// NewCLIProvider creates a CLIProvider from a CLISpec and CLIRunner.
func NewCLIProvider(spec CLISpec, runner CLIRunner) *CLIProvider {
	if spec.Timeout == 0 {
		spec.Timeout = DefaultCLITimeout
	}
	if spec.Parser == nil {
		spec.Parser = PlainTextParser{}
	}
	return &CLIProvider{
		spec:   spec,
		runner: runner,
	}
}

// Name returns the spec's name as the provider identifier.
func (p *CLIProvider) Name() string {
	return p.spec.Name
}

// Complete sends the prompt to the CLI tool and returns the parsed response.
func (p *CLIProvider) Complete(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrEmptyPrompt
	}

	ctx, cancel := context.WithTimeout(ctx, p.spec.Timeout)
	defer cancel()

	args := p.buildArgs(prompt)

	var stdin string
	if p.spec.InputMethod == InputStdin {
		stdin = prompt
	}

	stdout, stderr, err := p.runner.RunWithStdin(ctx, stdin, p.spec.Command, args...)
	if err != nil {
		return "", fmt.Errorf("%s cli: %s: %w", p.spec.Name, stderr, err)
	}

	return p.spec.Parser.Parse([]byte(stdout))
}

// Available reports whether the CLI binary exists on PATH.
func (p *CLIProvider) Available(_ context.Context) bool {
	_, err := exec.LookPath(p.spec.Command)
	return err == nil
}

// buildArgs constructs the full argument list from the spec and prompt.
func (p *CLIProvider) buildArgs(prompt string) []string {
	args := make([]string, 0, len(p.spec.BaseArgs)+4)
	args = append(args, p.spec.BaseArgs...)

	if p.spec.SystemPrompt.Enabled {
		args = append(args, p.spec.SystemPrompt.Flag, p.spec.SystemPrompt.Value)
	}
	if p.spec.OutputFormat.Enabled {
		args = append(args, p.spec.OutputFormat.Flag, p.spec.OutputFormat.Value)
	}

	if p.spec.InputMethod == InputArg {
		args = append(args, prompt)
	}

	return args
}
