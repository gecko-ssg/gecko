/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.2"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Gecko version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "gecko version %s\n", version)
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
