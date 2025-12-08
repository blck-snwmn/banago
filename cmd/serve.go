package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/blck-snwmn/banago/internal/project"
	"github.com/blck-snwmn/banago/internal/server"
	"github.com/spf13/cobra"
)

var serveOpts struct {
	port int
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a web server to browse generated images",
	Long:  "Launch a local web server to view generation history and images in a browser.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return fmt.Errorf("banago project not found. Run 'banago init' first")
			}
			return err
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Starting server at http://localhost:%d\n", serveOpts.port)
		_, _ = fmt.Fprintln(w, "Press Ctrl+C to stop")

		srv := server.New(projectRoot, serveOpts.port)
		return srv.Start()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVar(&serveOpts.port, "port", 8080, "Port to listen on")
}
