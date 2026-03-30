package quota

import (
	"fmt"
	"os/exec"
)

// Notifier sends warning messages to the appropriate destination.
type Notifier struct {
	// CLIMode sends warnings to stdout instead of multiclaude messaging.
	CLIMode bool
}

// NewNotifier creates a Notifier with the given configuration.
func NewNotifier(cliMode bool) *Notifier {
	return &Notifier{CLIMode: cliMode}
}

// Send delivers a warning message. In CLI mode it prints to stdout.
// Otherwise it sends via multiclaude message send supervisor.
// This function is ADVISORY ONLY — it never blocks or throttles agents.
func (n *Notifier) Send(message string) error {
	if message == "" {
		return nil
	}
	if n.CLIMode {
		fmt.Println(message)
		return nil
	}
	return sendMulticlaudeMessage("supervisor", message)
}

// sendMulticlaudeMessage sends a message via the multiclaude CLI.
func sendMulticlaudeMessage(recipient, message string) error {
	cmd := exec.Command("multiclaude", "message", "send", recipient, message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("multiclaude message send %s: %w (output: %s)", recipient, err, output)
	}
	return nil
}
