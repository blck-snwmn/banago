package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var historyOpts struct {
	limit int
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show generation history",
	Long:  "Display the generation history of the current subproject.",
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

		subprojectName, err := project.FindCurrentSubproject(projectRoot, cwd)
		if err != nil {
			if errors.Is(err, project.ErrNotInSubproject) {
				return fmt.Errorf("not in a subproject. Navigate to a subproject directory")
			}
			return err
		}

		subprojectDir := project.GetSubprojectDir(projectRoot, subprojectName)
		historyDir := history.GetHistoryDir(subprojectDir)

		entries, err := history.ListEntries(historyDir)
		if err != nil {
			return fmt.Errorf("failed to load history: %w", err)
		}

		w := cmd.OutOrStdout()

		if len(entries) == 0 {
			_, _ = fmt.Fprintln(w, "No history found")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "To generate images:")
			_, _ = fmt.Fprintln(w, "  banago generate --prompt \"...\"")
			return nil
		}

		_, _ = fmt.Fprintf(w, "History (%d entries):\n", len(entries))
		_, _ = fmt.Fprintln(w, "")

		// Show entries in reverse order (newest first)
		start := 0
		if historyOpts.limit > 0 && historyOpts.limit < len(entries) {
			start = len(entries) - historyOpts.limit
		}

		for i := len(entries) - 1; i >= start; i-- {
			entry := entries[i]
			status := "✓"
			if !entry.Result.Success {
				status = "✗"
			}
			_, _ = fmt.Fprintf(w, "  %s %s\n", status, entry.ID)
			_, _ = fmt.Fprintf(w, "      Date: %s\n", entry.CreatedAt)
			if entry.Result.Success && len(entry.Result.OutputImages) > 0 {
				_, _ = fmt.Fprintf(w, "      Output: %d images\n", len(entry.Result.OutputImages))
			}

			// List edits
			entryDir := filepath.Join(historyDir, entry.ID)
			edits, _ := history.ListEditEntries(entryDir)
			if len(edits) > 0 {
				_, _ = fmt.Fprintf(w, "      Edits:\n")
				for _, edit := range edits {
					editStatus := "✓"
					if !edit.Result.Success {
						editStatus = "✗"
					}
					_, _ = fmt.Fprintf(w, "        %s %s\n", editStatus, edit.ID)
					_, _ = fmt.Fprintf(w, "            Date: %s\n", edit.CreatedAt)
				}
			}

			if !entry.Result.Success && entry.Result.ErrorMessage != "" {
				_, _ = fmt.Fprintf(w, "      Error: %s\n", entry.Result.ErrorMessage)
			}
			_, _ = fmt.Fprintln(w, "")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().IntVar(&historyOpts.limit, "limit", 10, "Number of history entries to show")
}
