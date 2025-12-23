package cmd

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/gemini"
	"github.com/blck-snwmn/banago/internal/generation"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

type editOptions struct {
	id         string
	latest     bool
	editID     string
	editLatest bool
	prompt     string
	promptFile string
	aspect     string
	size       string
}

var editOpts editOptions

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a generated image",
	Long: `Edit a generated image using Gemini's image editing capabilities.

Uses an existing output image as input and applies the edit prompt.
Results are saved in the edits/ subdirectory of the history entry.

Examples:
  banago edit --latest -p "Change the button color to red"
  banago edit --latest --edit-latest -p "Further adjust the background"
  banago edit --id <uuid> -p "Fix the background"
  banago edit --id <uuid> --edit-id <edit-uuid> -p "Additional adjustments"`,
	Args: cobra.NoArgs,
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, _ []string) error {
	if err := requireAPIKey(); err != nil {
		return err
	}

	// Validate flags
	if !editOpts.latest && editOpts.id == "" {
		return errors.New("specify --latest or --id <uuid>")
	}

	promptText, err := resolveEditPrompt(editOpts.prompt, editOpts.promptFile)
	if err != nil {
		return err
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

	// Load generate entry
	var genEntry *history.Entry
	if editOpts.latest {
		genEntry, err = history.GetLatestEntry(historyDir)
		if err != nil {
			return fmt.Errorf("failed to get latest history: %w", err)
		}
	} else {
		genEntry, err = history.GetEntryByID(historyDir, editOpts.id)
		if err != nil {
			return fmt.Errorf("failed to get history entry: %w", err)
		}
	}

	entryDir := filepath.Join(historyDir, genEntry.ID)

	// Determine source image path and source info
	var sourceImagePath string
	var sourceType string
	var sourceEditID string
	var sourceOutput string
	var editEntry *history.EditEntry

	if editOpts.editLatest || editOpts.editID != "" {
		// Edit from an existing edit
		if editOpts.editLatest {
			editEntry, err = history.GetLatestEditEntry(entryDir)
			if err != nil {
				return fmt.Errorf("failed to get latest edit: %w", err)
			}
		} else {
			editEntry, err = history.GetEditEntryByID(entryDir, editOpts.editID)
			if err != nil {
				return fmt.Errorf("failed to get edit entry: %w", err)
			}
		}

		if len(editEntry.Result.OutputImages) == 0 {
			return errors.New("no output images in edit entry")
		}

		sourceOutput = editEntry.Result.OutputImages[0]
		sourceImagePath = history.GetEditOutputPath(entryDir, editEntry.ID, sourceOutput)
		sourceType = "edit"
		sourceEditID = editEntry.ID
	} else {
		// Edit from generate output
		if len(genEntry.Result.OutputImages) == 0 {
			return errors.New("no output images in history entry")
		}

		sourceOutput = genEntry.Result.OutputImages[0]
		sourceImagePath = filepath.Join(entryDir, sourceOutput)
		sourceType = "generate"
	}

	// Verify source image exists
	if _, err := os.Stat(sourceImagePath); err != nil {
		return fmt.Errorf("source image not found: %s", sourceImagePath)
	}

	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(w, "Editing from %s: %s\n", sourceType, sourceOutput)
	_, _ = fmt.Fprintln(w, "")

	// Resolve aspect ratio and image size: flag > source edit history > generate history > config
	var editAspect, editSize string
	if editEntry != nil {
		editAspect = editEntry.Generation.AspectRatio
		editSize = editEntry.Generation.ImageSize
	}
	aspect := cmp.Or(editOpts.aspect, editAspect, genEntry.Generation.AspectRatio, subprojectCfg.AspectRatio)
	size := cmp.Or(editOpts.size, editSize, genEntry.Generation.ImageSize, subprojectCfg.ImageSize)

	// Build edit spec
	spec := generation.EditSpec{
		Model:           model,
		Prompt:          promptText,
		AspectRatio:     aspect,
		ImageSize:       size,
		SourceImagePath: sourceImagePath,
		EntryID:         genEntry.ID,
		SourceType:      sourceType,
		SourceEditID:    sourceEditID,
		SourceOutput:    sourceOutput,
	}

	// Create Gemini client
	client, err := gemini.NewClient(cmd.Context(), cfg.apiKey)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Run edit
	_, err = generation.NewService(client).Edit(cmd.Context(), spec, historyDir, w)
	return err
}

func resolveEditPrompt(prompt, promptFile string) (string, error) {
	if prompt != "" && promptFile != "" {
		return "", errors.New("cannot specify both --prompt and --prompt-file")
	}
	if prompt == "" && promptFile == "" {
		return "", errors.New("specify --prompt or --prompt-file")
	}

	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file: %w", err)
		}
		prompt = string(data)
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	return prompt, nil
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVar(&editOpts.id, "id", "", "History entry ID to edit")
	editCmd.Flags().BoolVar(&editOpts.latest, "latest", false, "Use the latest history entry")
	editCmd.Flags().StringVar(&editOpts.editID, "edit-id", "", "Edit entry ID to edit from")
	editCmd.Flags().BoolVar(&editOpts.editLatest, "edit-latest", false, "Use the latest edit entry")
	editCmd.Flags().StringVarP(&editOpts.prompt, "prompt", "p", "", "Edit prompt")
	editCmd.Flags().StringVarP(&editOpts.promptFile, "prompt-file", "F", "", "Path to edit prompt file")
	editCmd.Flags().StringVar(&editOpts.aspect, "aspect", "", "Output image aspect ratio (overrides history/config)")
	editCmd.Flags().StringVar(&editOpts.size, "size", "", "Output image size (overrides history/config)")

	editCmd.MarkFlagsMutuallyExclusive("id", "latest")
	editCmd.MarkFlagsMutuallyExclusive("edit-id", "edit-latest")
	editCmd.MarkFlagsMutuallyExclusive("prompt", "prompt-file")
}
