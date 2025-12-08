package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
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

		subprojectDir := config.GetSubprojectDir(projectRoot, subprojectName)
		subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			return fmt.Errorf("failed to load subproject config: %w", err)
		}

		historyDir := config.GetHistoryDir(subprojectDir)

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

		// Get input images from history entry
		inputsDir := config.GetInputsDir(subprojectDir)
		var imagePaths []string
		for _, img := range sourceEntry.Generation.InputImages {
			imagePaths = append(imagePaths, filepath.Join(inputsDir, img))
		}

		if len(imagePaths) == 0 {
			return errors.New("no input images found in history entry")
		}

		// Determine aspect ratio and size from subproject config
		aspect := subprojectCfg.AspectRatio
		size := subprojectCfg.ImageSize

		// Generate images
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.apiKey, Backend: genai.BackendGeminiAPI})
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		parts := []*genai.Part{genai.NewPartFromText(promptText)}
		for _, imgPath := range imagePaths {
			part, err := imagePartFromFile(imgPath)
			if err != nil {
				return err
			}
			parts = append(parts, part)
		}

		gcfg := &genai.GenerateContentConfig{ResponseModalities: []string{"IMAGE"}}
		if aspect != "" || size != "" {
			gcfg.ImageConfig = &genai.ImageConfig{}
			if aspect != "" {
				gcfg.ImageConfig.AspectRatio = aspect
			}
			if size != "" {
				gcfg.ImageConfig.ImageSize = strings.ToUpper(size)
			}
		}

		contents := []*genai.Content{{Parts: parts}}
		resp, err := client.Models.GenerateContent(ctx, model, contents, gcfg)

		// Save to new history entry
		entry := history.NewEntry()
		entry.Generation.PromptFile = history.PromptFile
		entry.Generation.InputImages = append(entry.Generation.InputImages, sourceEntry.Generation.InputImages...)

		entryDir := entry.GetEntryDir(historyDir)

		// Save prompt
		if mkErr := os.MkdirAll(entryDir, 0o755); mkErr != nil {
			return fmt.Errorf("failed to create history directory: %w", mkErr)
		}
		if saveErr := entry.SavePrompt(historyDir, promptText); saveErr != nil {
			return fmt.Errorf("failed to save prompt: %w", saveErr)
		}

		// Copy context file from source if exists
		if sourceEntry.Generation.ContextFile != "" {
			srcContextPath := filepath.Join(sourceEntryDir, history.ContextFile)
			if _, statErr := os.Stat(srcContextPath); statErr == nil {
				data, readErr := os.ReadFile(srcContextPath)
				if readErr == nil {
					dstContextPath := filepath.Join(entryDir, history.ContextFile)
					if writeErr := os.WriteFile(dstContextPath, data, 0o644); writeErr == nil {
						entry.Generation.ContextFile = history.ContextFile
					}
				}
			}
		}

		// Copy character file from source if exists
		if sourceEntry.Generation.CharacterFile != "" {
			srcCharPath := filepath.Join(sourceEntryDir, history.CharacterFile)
			if _, statErr := os.Stat(srcCharPath); statErr == nil {
				data, readErr := os.ReadFile(srcCharPath)
				if readErr == nil {
					dstCharPath := filepath.Join(entryDir, history.CharacterFile)
					if writeErr := os.WriteFile(dstCharPath, data, 0o644); writeErr == nil {
						entry.Generation.CharacterFile = history.CharacterFile
					}
				}
			}
		}

		if err != nil {
			// Save failed entry
			entry.Result.Success = false
			entry.Result.ErrorMessage = err.Error()
			if saveErr := entry.Save(historyDir); saveErr != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to save history: %v\n", saveErr)
			}
			return fmt.Errorf("failed to generate image: %w", err)
		}

		// Save generated images
		saved, saveErr := saveInlineImages(resp, entryDir, "output")
		if saveErr != nil {
			entry.Result.Success = false
			entry.Result.ErrorMessage = saveErr.Error()
			if err := entry.Save(historyDir); err != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to save history: %v\n", err)
			}
			return saveErr
		}

		// Update entry with results
		entry.Result.Success = true
		for _, s := range saved {
			entry.Result.OutputImages = append(entry.Result.OutputImages, filepath.Base(s))
		}
		if resp.UsageMetadata != nil {
			entry.Result.TokenUsage = history.TokenUsage{
				Prompt:     int(resp.UsageMetadata.PromptTokenCount),
				Candidates: int(resp.UsageMetadata.CandidatesTokenCount),
				Total:      int(resp.UsageMetadata.TotalTokenCount),
				Cached:     int(resp.UsageMetadata.CachedContentTokenCount),
				Thoughts:   int(resp.UsageMetadata.ThoughtsTokenCount),
			}
		}

		if err := entry.Save(historyDir); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to save history: %v\n", err)
		}

		// Output
		_, _ = fmt.Fprintf(w, "History ID: %s\n", entry.ID)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Generated files:")
		for _, s := range saved {
			_, _ = fmt.Fprintf(w, "  %s\n", filepath.Base(s))
		}

		if text := strings.TrimSpace(resp.Text()); text != "" {
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "Text response:")
			_, _ = fmt.Fprintln(w, text)
		}

		if resp.UsageMetadata != nil {
			usage := resp.UsageMetadata
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "Token usage:")
			_, _ = fmt.Fprintf(w, "  prompt: %d\n", usage.PromptTokenCount)
			_, _ = fmt.Fprintf(w, "  candidates: %d\n", usage.CandidatesTokenCount)
			_, _ = fmt.Fprintf(w, "  total: %d\n", usage.TotalTokenCount)
			if usage.CachedContentTokenCount > 0 {
				_, _ = fmt.Fprintf(w, "  cached: %d\n", usage.CachedContentTokenCount)
			}
			if usage.ThoughtsTokenCount > 0 {
				_, _ = fmt.Fprintf(w, "  thoughts: %d\n", usage.ThoughtsTokenCount)
			}
		}

		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintf(w, "Model: %s\n", model)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(regenerateCmd)

	regenerateCmd.Flags().StringVar(&regenOpts.id, "id", "", "History entry ID to regenerate from")
	regenerateCmd.Flags().BoolVar(&regenOpts.latest, "latest", false, "Use the latest history entry")

	regenerateCmd.MarkFlagsMutuallyExclusive("id", "latest")
}
