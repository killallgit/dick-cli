/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(dick completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ dick completion bash > /etc/bash_completion.d/dick
  # macOS:
  $ dick completion bash > $(brew --prefix)/etc/bash_completion.d/dick

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ dick completion zsh > "${fpath[1]}/_dick"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ dick completion fish | source

  # To load completions for each session, execute once:
  $ dick completion fish > ~/.config/fish/completions/dick.fish

PowerShell:

  PS> dick completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> dick completion powershell > dick.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}