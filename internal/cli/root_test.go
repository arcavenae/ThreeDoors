package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd_Structure(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()

	if root.Use != "threedoors" {
		t.Errorf("Use = %q, want %q", root.Use, "threedoors")
	}
	if root.Short == "" {
		t.Error("Short description should not be empty")
	}
	if root.Long == "" {
		t.Error("Long description should not be empty")
	}
	if !root.SilenceUsage {
		t.Error("SilenceUsage should be true")
	}
	if !root.SilenceErrors {
		t.Error("SilenceErrors should be true")
	}
}

func TestNewRootCmd_HasJSONFlag(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	flag := root.PersistentFlags().Lookup("json")
	if flag == nil {
		t.Fatal("missing --json persistent flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("json flag default = %q, want %q", flag.DefValue, "false")
	}
}

func TestNewRootCmd_Subcommands(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	subCmds := root.Commands()
	names := make(map[string]bool)
	for _, cmd := range subCmds {
		names[cmd.Name()] = true
	}

	expected := []string{"task", "doors", "doctor", "version", "completion", "mood", "stats", "config", "plan", "doc-audit"}
	for _, want := range expected {
		if !names[want] {
			t.Errorf("missing %q subcommand", want)
		}
	}
}

func TestIsJSONOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json bool
		want bool
	}{
		{"default false", false, false},
		{"set true", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{}
			cmd.Flags().Bool("json", tt.json, "")
			got := isJSONOutput(cmd)
			if got != tt.want {
				t.Errorf("isJSONOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsJSONOutput_NoFlag(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	got := isJSONOutput(cmd)
	if got {
		t.Error("isJSONOutput() should return false when flag is missing")
	}
}

func TestKnownSubcommands(t *testing.T) {
	t.Parallel()

	names := KnownSubcommands()
	if len(names) == 0 {
		t.Fatal("expected non-empty subcommand list")
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	required := []string{"task", "doors", "completion", "mood", "stats", "config", "doctor", "version", "help", "plan"}
	for _, want := range required {
		if !nameSet[want] {
			t.Errorf("KnownSubcommands() missing %q", want)
		}
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	t.Parallel()

	// Execute with an unknown command should return non-zero
	// But Execute() calls os.Exit path, so we test NewRootCmd directly
	root := NewRootCmd()
	root.SetArgs([]string{"nonexistent-command"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}
