/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gecko",
	Short:   "A static site generator",
	Long:    "Gecko builds and serves static websites from Markdown content and HTML templates.\n\nShell completion setup:\n  Bash: add 'source <(gecko completion bash)' to ~/.bashrc\n  Zsh:  add 'source <(gecko completion zsh)' to ~/.zshrc\n  Fish: gecko completion fish > ~/.config/fish/completions/gecko.fish",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.InitDefaultHelpFlag()
	rootCmd.InitDefaultVersionFlag()
	rootCmd.InitDefaultCompletionCmd()
}
