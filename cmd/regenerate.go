package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/generation"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

type regenerateOptions struct {
	id     string
	latest bool
}

var regenOpts regenerateOptions

var regenerateCmd = &cobra.Command{
	Use:   "regenerate",
	Short: "Regenerate images from history",
	Long: `Regenerate images using a previous history entry.

Uses the prompt and input images from the specified history entry
to generate new images. Results are saved as a new history entry.

Examples:
  banago regenerate --latest           # Use the latest history entry
  banago regenerate --id <uuid>        # Use a specific history entry`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := requireAPIKey(); err != nil {
			return err
		}

		if !regenOpts.latest && regenOpts.id == "" {
			return errors.New("specify --latest or --id <uuid>")
		}

		// Must be in a subproject
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return errors.New("banago project not found. Run 'banago init' first")
			}
			return err
		}

		// Load project config
		projectCfg, err := config.LoadProjectConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		model := projectCfg.Model

		subprojectName, err := project.FindCurrentSubproject(projectRoot, cwd)
		if err != nil {
			if errors.Is(err, project.ErrNotInSubproject) {
				return errors.New("not in a subproject. Navigate to a subproject directory")
			}
			return err
		}

		subprojectDir := project.GetSubprojectDir(projectRoot, subprojectName)
		subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			return fmt.Errorf("failed to load subproject config: %w", err)
		}

		historyDir := history.GetHistoryDir(subprojectDir)

		// Load history entry
		var sourceEntry *history.Entry
		if regenOpts.latest {
			sourceEntry, err = history.GetLatestEntry(historyDir)
			if err != nil {
				return fmt.Errorf("failed to get latest history: %w", err)
			}
		} else {
			sourceEntry, err = history.GetEntryByID(historyDir, regenOpts.id)
			if err != nil {
				return fmt.Errorf("failed to get history entry: %w", err)
			}
		}

		// Load prompt from history
		sourceEntryDir := filepath.Join(historyDir, sourceEntry.ID)
		promptText, err := history.LoadPrompt(sourceEntryDir)
		if err != nil {
			return fmt.Errorf("failed to load prompt from history: %w", err)
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Regenerating from history: %s\n", sourceEntry.ID)
		_, _ = fmt.Fprintln(w, "")

		// Get input images from history entry directory
		imagePaths := history.GetInputImagePaths(sourceEntryDir, sourceEntry.Generation.InputImages)

		if len(imagePaths) == 0 {
			return errors.New("no input images found in history entry. Run 'banago migrate' first")
		}

		// Build generation context
		genCtx := &generation.Context{
			Model:           model,
			Prompt:          promptText,
			ImagePaths:      imagePaths,
			AspectRatio:     subprojectCfg.AspectRatio,
			ImageSize:       subprojectCfg.ImageSize,
			InputImageNames: sourceEntry.Generation.InputImages,
			SourceEntryID:   sourceEntry.ID,
		}

		// Run generation
		_, err = generation.Run(context.Background(), cfg.apiKey, genCtx, historyDir, w)
		return err
	},
}

func init() {
	rootCmd.AddCommand(regenerateCmd)

	regenerateCmd.Flags().StringVar(&regenOpts.id, "id", "", "History entry ID to regenerate from")
	regenerateCmd.Flags().BoolVar(&regenOpts.latest, "latest", false, "Use the latest history entry")

	regenerateCmd.MarkFlagsMutuallyExclusive("id", "latest")
}
