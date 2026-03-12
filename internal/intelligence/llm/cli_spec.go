package llm

import (
	"strings"
	"time"
)

// InputMethod defines how prompts are delivered to CLI tools.
type InputMethod int

const (
	// InputStdin pipes the prompt to the process's standard input.
	InputStdin InputMethod = iota
	// InputArg appends the prompt as the final positional argument.
	InputArg
)

// ArgTemplate represents an optional CLI flag with a value.
type ArgTemplate struct {
	Flag    string
	Value   string
	Enabled bool
}

// ResponseParser extracts the LLM response text from raw CLI stdout.
type ResponseParser interface {
	Parse(stdout []byte) (string, error)
}

// CLISpec declaratively describes how to invoke a CLI-based LLM tool.
type CLISpec struct {
	Name         string
	Command      string
	BaseArgs     []string
	SystemPrompt ArgTemplate
	OutputFormat ArgTemplate
	InputMethod  InputMethod
	Timeout      time.Duration
	Parser       ResponseParser
}

// PlainTextParser trims whitespace from stdout and returns it as the response.
type PlainTextParser struct{}

// Parse trims whitespace from stdout. Returns ErrEmptyResponse if the result is empty.
func (p PlainTextParser) Parse(stdout []byte) (string, error) {
	text := strings.TrimSpace(string(stdout))
	if text == "" {
		return "", ErrEmptyResponse
	}
	return text, nil
}
