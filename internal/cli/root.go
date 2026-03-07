package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var jsonOutput bool

// NewRootCmd creates the top-level Cobra command for ThreeDoors CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threedoors",
		Short: "ThreeDoors — reduce decision friction, three tasks at a time",
		Long: `ThreeDoors is a task management tool that reduces decision friction
by presenting only three tasks at a time. Run without arguments to
launch the interactive TUI, or use subcommands for scriptable access.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	cmd.AddCommand(newTaskCmd())
	cmd.AddCommand(NewDoorsCmd())

	return cmd
}

// Execute runs the root command. It returns the exit code to use.
func Execute() int {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		formatter := NewOutputFormatter(os.Stderr, jsonOutput)
		if jsonOutput {
			_ = formatter.WriteJSONError("", ExitGeneralError, err.Error(), "")
		} else {
			_ = formatter.Writef("Error: %v\n", err)
		}
		return ExitGeneralError
	}
	return ExitSuccess
}

// KnownSubcommands returns the list of registered subcommand names.
// Used by main.go to decide between TUI and CLI routing.
func KnownSubcommands() []string {
	root := NewRootCmd()
	var names []string
	for _, cmd := range root.Commands() {
		names = append(names, cmd.Name())
	}
	// Include subcommands that will be added in future stories
	return append(names, "task", "doors", "mood", "stats", "config", "provider", "health", "version", "help")
}
