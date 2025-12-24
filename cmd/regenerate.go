package cmd

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/gemini"
	"github.com/blck-snwmn/banago/internal/generation"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

type regenerateOptions struct {
	id     string
	latest bool
	aspect string
	size   string
}

// regenerateHandler handles the regenerate command with dependency injection support.
type regenerateHandler struct {
	generator generation.Generator
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

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Create Gemini client and inject into handler
		client, err := gemini.NewClient(cmd.Context(), cfg.apiKey)
		if err != nil {
			return fmt.Errorf("failed to create Gemini client: %w", err)
		}

		handler := &regenerateHandler{generator: client}
		return handler.run(cmd.Context(), regenOpts, cwd, cmd.OutOrStdout())
	},
}

// run executes the regenerate command logic.
// This method is independent of cobra.Command for testability.
func (h *regenerateHandler) run(ctx context.Context, opts regenerateOptions, workDir string, w io.Writer) error {
	projectRoot, err := project.FindProjectRoot(workDir)
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

	subprojectName, err := project.FindCurrentSubproject(projectRoot, workDir)
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
	if opts.latest {
		sourceEntry, err = history.GetLatestEntry(historyDir)
		if err != nil {
			return fmt.Errorf("failed to get latest history: %w", err)
		}
	} else {
		sourceEntry, err = history.GetEntryByID(historyDir, opts.id)
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

	_, _ = fmt.Fprintf(w, "Regenerating from history: %s\n", sourceEntry.ID)
	_, _ = fmt.Fprintln(w, "")

	// Get input images from history entry directory
	imagePaths := history.GetInputImagePaths(sourceEntryDir, sourceEntry.Generation.InputImages)

	if len(imagePaths) == 0 {
		return errors.New("no input images found in history entry. Run 'banago migrate' first")
	}

	// Resolve aspect ratio and image size: flag > history > config
	aspect := cmp.Or(opts.aspect, sourceEntry.Generation.AspectRatio, subprojectCfg.AspectRatio)
	size := cmp.Or(opts.size, sourceEntry.Generation.ImageSize, subprojectCfg.ImageSize)

	// Build generation spec
	spec := generation.Spec{
		Model:           model,
		Prompt:          promptText,
		ImagePaths:      imagePaths,
		AspectRatio:     aspect,
		ImageSize:       size,
		InputImageNames: sourceEntry.Generation.InputImages,
		SourceEntryID:   sourceEntry.ID,
	}

	// Run generation with injected generator
	_, err = generation.NewService(h.generator).Run(ctx, spec, historyDir, w)
	return err
}

func init() {
	rootCmd.AddCommand(regenerateCmd)

	regenerateCmd.Flags().StringVar(&regenOpts.id, "id", "", "History entry ID to regenerate from")
	regenerateCmd.Flags().BoolVar(&regenOpts.latest, "latest", false, "Use the latest history entry")
	regenerateCmd.Flags().StringVar(&regenOpts.aspect, "aspect", "", "Output image aspect ratio (overrides history/config)")
	regenerateCmd.Flags().StringVar(&regenOpts.size, "size", "", "Output image size (overrides history/config)")

	regenerateCmd.MarkFlagsOneRequired("id", "latest")
	regenerateCmd.MarkFlagsMutuallyExclusive("id", "latest")
}
