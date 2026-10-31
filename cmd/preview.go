/*
Copyright © 2026 dos <debasmitbos22@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use: "preview", Short: "Build and preview the generated site", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, _ := cmd.Flags().GetString("dir")
		output, _ := cmd.Flags().GetString("output")
		port, _ := cmd.Flags().GetInt("port")
		root, err := filepath.Abs(directory)
		if err != nil {
			return fmt.Errorf("resolve project directory: %w", err)
		}
		if !filepath.IsAbs(output) {
			output = filepath.Join(root, output)
		}
		output, err = filepath.Abs(output)
		if err != nil {
			return fmt.Errorf("resolve output directory: %w", err)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		staging := filepath.Join(root, ".gecko-preview-staging")
		if output == staging {
			return fmt.Errorf("output directory conflicts with preview staging directory")
		}
		if err := validateDevOutput(root, output); err != nil {
			return err
		}
		if err := rebuildAndPublish(root, staging, output); err != nil {
			return fmt.Errorf("build site: %w", err)
		}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		server := &http.Server{Addr: fmt.Sprintf("localhost:%d", port), Handler: http.FileServer(http.Dir(output))}
		serverErrors := make(chan error, 1)
		go func() { serverErrors <- server.ListenAndServe() }()
		fmt.Fprintf(cmd.OutOrStdout(), "Previewing %s at http://localhost:%d\nPress Ctrl+C to stop.\n", output, port)
		select {
		case <-ctx.Done():
			return server.Shutdown(context.Background())
		case err := <-serverErrors:
			if err == http.ErrServerClosed {
				return nil
			}
			return fmt.Errorf("preview server: %w", err)
		}
	},
}

func init() {
	previewCmd.Flags().String("dir", ".", "Project directory")
	previewCmd.Flags().Int("port", 3000, "Preview server port")
	previewCmd.Flags().String("output", "dist", "Generated site directory")
	rootCmd.AddCommand(previewCmd)
}
