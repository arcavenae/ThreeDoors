package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRegisterFlagCompletions(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	// registerFlagCompletions is already called in NewRootCmd, so verify it ran
	taskCmd, _, _ := root.Find([]string{"task"})
	if taskCmd == nil {
		t.Fatal("task command not found")
	}

	// The list subcommand has status, type, effort flags
	listCmd, _, _ := taskCmd.Find([]string{"list"})
	if listCmd == nil {
		t.Fatal("task list command not found")
	}

	for _, name := range []string{"status", "type", "effort"} {
		f := listCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("task list missing flag %q", name)
		}
	}
}

func TestRegisterEnumFlag_NoFlag(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	// Should not panic when flag doesn't exist
	registerEnumFlag(cmd, "nonexistent", []string{"a", "b"})
}

func TestRegisterEnumFlag_WithFlag(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("status", "", "filter status")

	registerEnumFlag(cmd, "status", []string{"todo", "in-progress", "complete"})

	// Verify registration didn't error (we can't easily test completion output
	// but we verified the function doesn't panic)
}
