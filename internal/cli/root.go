package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// isJSONOutput reads the --json flag from the command's inherited flags.
func isJSONOutput(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("json")
	return v
}

// NewRootCmd creates the top-level Cobra command for ThreeDoors CLI.
func NewRootCmd() *cobra.Command {
	var jsonOutput bool

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
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newMoodCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newPlanCmd())
	cmd.AddCommand(newDocAuditCmd())
	cmd.AddCommand(newSourcesCmd())
	cmd.AddCommand(newConnectCmd())
	cmd.AddCommand(newExtractCmd())
	cmd.AddCommand(newLLMCmd())
	cmd.AddCommand(newDevicesCmd())
	cmd.AddCommand(newSyncCmd())
	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newMergeDriverCmd())

	registerFlagCompletions(cmd)

	return cmd
}

// Execute runs the root command. It returns the exit code to use.
func Execute() int {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		jsonFlag, _ := root.PersistentFlags().GetBool("json")
		formatter := NewOutputFormatter(os.Stderr, jsonFlag)
		if jsonFlag {
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
	return append(names, "task", "doors", "completion", "mood", "stats", "config", "provider", "health", "doctor", "version", "help", "plan", "sources", "connect", "extract", "llm")
}
