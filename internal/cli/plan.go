package cli

import (
	"github.com/spf13/cobra"
)

// newPlanCmd creates the "plan" subcommand that launches the TUI in planning mode.
func newPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Start a daily planning session",
		Long: `Launch the ThreeDoors daily planning flow.

The planning session is a 3-step guided process:
  1. Review — look at incomplete tasks from yesterday
  2. Select — pick up to 5 tasks to focus on today
  3. Confirm — review your choices and commit

After the session, focus-boosted tasks will appear more frequently
in your door selection. The TUI exits after planning completes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Signal to main that planning mode was requested.
			// The actual TUI launch happens in main.go since it needs
			// the full TUI initialization pipeline.
			return nil
		},
	}

	return cmd
}

// IsPlanCommand returns true if the CLI args request planning mode.
func IsPlanCommand() bool {
	if len(planCommandArgs) > 1 {
		return planCommandArgs[1] == "plan"
	}
	return false
}

// planCommandArgs is set during init for early command detection.
// This avoids importing os in the check function.
var planCommandArgs []string

// SetPlanCommandArgs stores os.Args for plan command detection.
func SetPlanCommandArgs(args []string) {
	planCommandArgs = args
}
