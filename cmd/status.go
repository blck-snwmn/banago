package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

const (
	uuidShortLen = 8  // Length for shortened UUID display (e.g., "01234567...")
	datePrefixLen = 10 // Length for date prefix from RFC3339 (YYYY-MM-DD)
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current subproject status",
	Long:  "Display the status of the current directory's subproject.",
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

		// Load project config
		projectCfg, err := config.LoadProjectConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}

		w := cmd.OutOrStdout()

		// Check if we're in a subproject
		subprojectName, err := project.FindCurrentSubproject(projectRoot, cwd)
		if err != nil {
			if errors.Is(err, project.ErrNotInSubproject) {
				// Show project-level status
				_, _ = fmt.Fprintf(w, "Project: %s\n", projectCfg.Name)
				_, _ = fmt.Fprintf(w, "Model: %s\n", projectCfg.Model)
				_, _ = fmt.Fprintln(w, "")
				_, _ = fmt.Fprintln(w, "Not in a subproject.")
				_, _ = fmt.Fprintln(w, "Navigate to a subproject or create one:")
				_, _ = fmt.Fprintln(w, "  cd subprojects/<name>")
				_, _ = fmt.Fprintln(w, "  banago subproject create <name>")
				return nil
			}
			return err
		}

		// Show subproject-level status
		subprojectDir := project.GetSubprojectDir(projectRoot, subprojectName)
		subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			return fmt.Errorf("failed to load subproject config: %w", err)
		}

		_, _ = fmt.Fprintf(w, "Project: %s\n", projectCfg.Name)
		_, _ = fmt.Fprintf(w, "Subproject: %s\n", subprojectCfg.Name)
		if subprojectCfg.Description != "" {
			_, _ = fmt.Fprintf(w, "Description: %s\n", subprojectCfg.Description)
		}
		_, _ = fmt.Fprintln(w, "")

		// Context file
		contextPath := filepath.Join(subprojectDir, subprojectCfg.ContextFile)
		if _, err := os.Stat(contextPath); err == nil {
			relPath, _ := filepath.Rel(cwd, contextPath)
			_, _ = fmt.Fprintf(w, "Context: %s\n", relPath)
		}

		// Character file
		if subprojectCfg.CharacterFile != "" {
			characterPath := project.GetCharacterPath(projectRoot, subprojectCfg.CharacterFile)
			relPath, _ := filepath.Rel(cwd, characterPath)
			if _, err := os.Stat(characterPath); err == nil {
				_, _ = fmt.Fprintf(w, "Character: %s\n", relPath)
			} else {
				_, _ = fmt.Fprintf(w, "Character: %s (not found)\n", relPath)
			}
		}
		_, _ = fmt.Fprintln(w, "")

		// Input images
		_, _ = fmt.Fprintln(w, "Input images:")
		if len(subprojectCfg.InputImages) == 0 {
			_, _ = fmt.Fprintln(w, "  (none)")
		} else {
			inputsDir := project.GetInputsDir(subprojectDir)
			for _, img := range subprojectCfg.InputImages {
				imgPath := filepath.Join(inputsDir, img)
				relPath, _ := filepath.Rel(cwd, imgPath)
				if _, err := os.Stat(imgPath); err == nil {
					_, _ = fmt.Fprintf(w, "  %s\n", relPath)
				} else {
					_, _ = fmt.Fprintf(w, "  %s (not found)\n", relPath)
				}
			}
		}
		_, _ = fmt.Fprintln(w, "")

		// History summary
		historyDir := history.GetHistoryDir(subprojectDir)
		entries, err := history.ListEntries(historyDir)
		if err != nil {
			_, _ = fmt.Fprintln(w, "History: (load error)")
		} else if len(entries) == 0 {
			_, _ = fmt.Fprintln(w, "History: none")
		} else {
			_, _ = fmt.Fprintf(w, "History: %d entries\n", len(entries))
			// Show latest entry
			latest := entries[len(entries)-1]
			_, _ = fmt.Fprintf(w, "  Latest: %s (%s)\n", latest.ID[:uuidShortLen]+"...", latest.CreatedAt[:datePrefixLen])
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
