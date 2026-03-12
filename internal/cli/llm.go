package cli

import "github.com/spf13/cobra"

// newLLMCmd creates the "llm" subcommand group for LLM-related operations.
func newLLMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "LLM backend management",
		Long:  `Manage LLM backends used for task decomposition, enrichment, and other intelligence features.`,
	}

	cmd.AddCommand(newLLMStatusCmd())

	return cmd
}
