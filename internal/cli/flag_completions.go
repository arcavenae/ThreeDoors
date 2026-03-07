package cli

import "github.com/spf13/cobra"

// registerFlagCompletions sets up completion functions for enum flags on subcommands.
func registerFlagCompletions(root *cobra.Command) {
	statusValues := []string{"todo", "in-progress", "blocked", "in-review", "complete", "deferred", "archived"}
	typeValues := []string{"creative", "administrative", "technical", "physical"}
	effortValues := []string{"quick-win", "medium", "deep-work"}

	taskCmd, _, _ := root.Find([]string{"task"})
	if taskCmd == nil {
		return
	}

	for _, sub := range taskCmd.Commands() {
		registerEnumFlag(sub, "status", statusValues)
		registerEnumFlag(sub, "type", typeValues)
		registerEnumFlag(sub, "effort", effortValues)
	}
}

func registerEnumFlag(cmd *cobra.Command, flagName string, values []string) {
	if cmd.Flags().Lookup(flagName) == nil {
		return
	}
	_ = cmd.RegisterFlagCompletionFunc(flagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	})
}
