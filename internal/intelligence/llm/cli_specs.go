package llm

import "time"

// ClaudeCLISpec returns a CLISpec for the Claude CLI (claude --print).
func ClaudeCLISpec() CLISpec {
	return CLISpec{
		Name:     "claude-cli",
		Command:  "claude",
		BaseArgs: []string{"--print"},
		SystemPrompt: ArgTemplate{
			Flag:    "--system-prompt",
			Enabled: false,
		},
		OutputFormat: ArgTemplate{
			Flag:    "--output-format",
			Value:   "json",
			Enabled: false,
		},
		InputMethod: InputStdin,
		Timeout:     120 * time.Second,
		Parser:      PlainTextParser{},
	}
}

// GeminiCLISpec returns a CLISpec for the Gemini CLI.
func GeminiCLISpec() CLISpec {
	return CLISpec{
		Name:    "gemini-cli",
		Command: "gemini",
		OutputFormat: ArgTemplate{
			Flag:    "--output-format",
			Value:   "json",
			Enabled: false,
		},
		InputMethod: InputStdin,
		Timeout:     120 * time.Second,
		Parser:      PlainTextParser{},
	}
}

// OllamaCLISpec returns a CLISpec for the Ollama CLI with the given model.
func OllamaCLISpec(model string) CLISpec {
	return CLISpec{
		Name:     "ollama-cli",
		Command:  "ollama",
		BaseArgs: []string{"run", model},
		SystemPrompt: ArgTemplate{
			Flag:    "--system",
			Enabled: false,
		},
		InputMethod: InputArg,
		Timeout:     120 * time.Second,
		Parser:      PlainTextParser{},
	}
}

// CustomCLISpec returns a CLISpec for an arbitrary CLI tool.
func CustomCLISpec(cmd string, args []string) CLISpec {
	return CLISpec{
		Name:        "custom",
		Command:     cmd,
		BaseArgs:    args,
		InputMethod: InputStdin,
		Timeout:     120 * time.Second,
		Parser:      PlainTextParser{},
	}
}
