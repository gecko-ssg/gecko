/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove generated build artifacts",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		for _, name := range []string{"dist", "public", ".gecko-dev-staging", ".gecko-preview-staging", "dist.gecko-previous", "public.gecko-previous"} {
			if err := os.RemoveAll(filepath.Join(root, name)); err != nil {
				return fmt.Errorf("remove %s: %w", name, err)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Removed generated build artifacts.")
		return nil
	},
}

func init() { rootCmd.AddCommand(cleanCmd) }
