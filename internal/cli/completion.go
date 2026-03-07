package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCompletionCmd creates the "completion" command that generates shell completion scripts.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for ThreeDoors.

To load completions:

Bash:
  $ source <(threedoors completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ threedoors completion bash > /etc/bash_completion.d/threedoors
  # macOS:
  $ threedoors completion bash > $(brew --prefix)/etc/bash_completion.d/threedoors

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ threedoors completion zsh > "${fpath[1]}/_threedoors"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ threedoors completion fish | source

  # To load completions for each session, execute once:
  $ threedoors completion fish > ~/.config/fish/completions/threedoors.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
