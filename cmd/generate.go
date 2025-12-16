package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/generator"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

type generateOptions struct {
	prompt     string
	promptFile string
	images     []string
	aspect     string
	size       string
}

// resolvePrompt returns the prompt text from either inline prompt or file.
func resolvePrompt(prompt, promptFile string) (string, error) {
	if prompt != "" {
		text := strings.TrimSpace(prompt)
		if text == "" {
			return "", errors.New("prompt is empty. Specify with --prompt or --prompt-file")
		}
		return text, nil
	}
	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file: %w", err)
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			return "", errors.New("prompt is empty. Specify with --prompt or --prompt-file")
		}
		return text, nil
	}
	return "", errors.New("prompt is empty. Specify with --prompt or --prompt-file")
}

// collectImagePaths gathers image paths from subproject config and command flags.
func collectImagePaths(subprojectDir string, subprojectCfg *config.SubprojectConfig, additionalImages []string) []string {
	var imagePaths []string
	if len(subprojectCfg.InputImages) > 0 {
		inputsDir := config.GetInputsDir(subprojectDir)
		for _, img := range subprojectCfg.InputImages {
			imagePaths = append(imagePaths, filepath.Join(inputsDir, img))
		}
	}
	imagePaths = append(imagePaths, additionalImages...)
	return imagePaths
}

// resolveGenerationParams determines aspect ratio and size from flags and config.
func resolveGenerationParams(flagAspect, flagSize string, subprojectCfg *config.SubprojectConfig) (aspect, size string) {
	aspect = flagAspect
	size = flagSize
	if aspect == "" && subprojectCfg.AspectRatio != "" {
		aspect = subprojectCfg.AspectRatio
	}
	if size == "" && subprojectCfg.ImageSize != "" {
		size = subprojectCfg.ImageSize
	}
	return aspect, size
}

var genOpts generateOptions

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate images",
	Long: `Generate images using the Gemini API.

Must be run inside a subproject directory:
  - input_images from config.yaml are automatically used
  - Results are saved to history/
  - Additional images can be specified with --image`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := requireAPIKey(); err != nil {
			return err
		}

		// Get prompt
		promptText, err := resolvePrompt(genOpts.prompt, genOpts.promptFile)
		if err != nil {
			return err
		}

		// Must be in a project
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

		// Must be in a subproject
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

		// Collect image paths
		imagePaths := collectImagePaths(subprojectDir, subprojectCfg, genOpts.images)
		if len(imagePaths) == 0 {
			return errors.New("no images specified. Use --image or set input_images in subproject config.yaml")
		}

		// Determine aspect ratio and size
		aspect, size := resolveGenerationParams(genOpts.aspect, genOpts.size, subprojectCfg)

		// Generate images
		ctx := context.Background()
		result := generator.Generate(ctx, cfg.apiKey, generator.Params{
			Model:       model,
			Prompt:      promptText,
			ImagePaths:  imagePaths,
			AspectRatio: aspect,
			ImageSize:   size,
		})
		resp := result.Response
		err = result.Error

		w := cmd.OutOrStdout()

		// Save to history
		entry := history.NewEntry()
		entry.Generation.PromptFile = history.PromptFile

		// Extract input image filenames
		entry.Generation.InputImages = append(entry.Generation.InputImages, subprojectCfg.InputImages...)

		historyDir := config.GetHistoryDir(subprojectDir)
		entryDir := entry.GetEntryDir(historyDir)

		// Save prompt
		if err := os.MkdirAll(entryDir, 0o755); err != nil {
			return fmt.Errorf("failed to create history directory: %w", err)
		}
		if err := entry.SavePrompt(historyDir, promptText); err != nil {
			return fmt.Errorf("failed to save prompt: %w", err)
		}

		// Save input images
		if err := entry.SaveInputImages(historyDir, imagePaths); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to save input images: %v\n", err)
		}

		if err != nil {
			// Clean up history directory on generation failure
			if err := entry.Cleanup(historyDir); err != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to clean up history directory: %v\n", err)
			}
			return fmt.Errorf("failed to generate image: %w", err)
		}

		// Save generated images
		saved, saveErr := generator.SaveImages(resp, entryDir)
		if saveErr != nil {
			// Clean up history directory on save failure
			if err := entry.Cleanup(historyDir); err != nil {
				_, _ = fmt.Fprintf(w, "Warning: failed to clean up history directory: %v\n", err)
			}
			return saveErr
		}

		// Update entry with results
		entry.Result.Success = true
		for _, s := range saved {
			entry.Result.OutputImages = append(entry.Result.OutputImages, filepath.Base(s))
		}
		entry.Result.TokenUsage = result.TokenUsage

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

		generator.PrintOutput(w, resp, model)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&genOpts.prompt, "prompt", "p", "", "Prompt for generation")
	generateCmd.Flags().StringVarP(&genOpts.promptFile, "prompt-file", "F", "", "Path to text file containing prompt")
	generateCmd.Flags().StringSliceVarP(&genOpts.images, "image", "i", nil, "Additional image files to send with prompt (can specify multiple)")
	generateCmd.Flags().StringVar(&genOpts.aspect, "aspect", "", "Output image aspect ratio (e.g., 1:1, 16:9)")
	generateCmd.Flags().StringVar(&genOpts.size, "size", "", "Output image size (1K / 2K / 4K)")

	generateCmd.MarkFlagsOneRequired("prompt", "prompt-file")
	generateCmd.MarkFlagsMutuallyExclusive("prompt", "prompt-file")
}
