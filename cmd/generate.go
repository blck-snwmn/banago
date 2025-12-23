package cmd

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/generation"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

type generateOptions struct {
	prompt     string
	promptFile string
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

// collectImagePaths gathers image paths from subproject config.
func collectImagePaths(subprojectDir string, subprojectCfg *config.SubprojectConfig) []string {
	var imagePaths []string
	if len(subprojectCfg.InputImages) > 0 {
		inputsDir := project.GetInputsDir(subprojectDir)
		for _, img := range subprojectCfg.InputImages {
			imagePaths = append(imagePaths, filepath.Join(inputsDir, img))
		}
	}
	return imagePaths
}

// resolveGenerationParams determines aspect ratio and size from flags and config.
func resolveGenerationParams(flagAspect, flagSize string, subprojectCfg *config.SubprojectConfig) (aspect, size string) {
	return cmp.Or(flagAspect, subprojectCfg.AspectRatio), cmp.Or(flagSize, subprojectCfg.ImageSize)
}

var genOpts generateOptions

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate images",
	Long: `Generate images using the Gemini API.

Must be run inside a subproject directory:
  - input_images from config.yaml are automatically used
  - Results are saved to history/`,
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

		subprojectDir := project.GetSubprojectDir(projectRoot, subprojectName)
		subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			return fmt.Errorf("failed to load subproject config: %w", err)
		}

		// Collect image paths
		imagePaths := collectImagePaths(subprojectDir, subprojectCfg)
		if len(imagePaths) == 0 {
			return errors.New("no images specified. Set input_images in subproject config.yaml")
		}

		// Determine aspect ratio and size
		aspect, size := resolveGenerationParams(genOpts.aspect, genOpts.size, subprojectCfg)

		// Build generation spec
		spec := generation.Spec{
			Model:           model,
			Prompt:          promptText,
			ImagePaths:      imagePaths,
			AspectRatio:     aspect,
			ImageSize:       size,
			InputImageNames: subprojectCfg.InputImages,
		}

		// Run generation
		historyDir := history.GetHistoryDir(subprojectDir)
		_, err = generation.Run(cmd.Context(), cfg.apiKey, spec, historyDir, cmd.OutOrStdout())
		return err
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&genOpts.prompt, "prompt", "p", "", "Prompt for generation")
	generateCmd.Flags().StringVarP(&genOpts.promptFile, "prompt-file", "F", "", "Path to text file containing prompt")
	generateCmd.Flags().StringVar(&genOpts.aspect, "aspect", "", "Output image aspect ratio (e.g., 1:1, 16:9)")
	generateCmd.Flags().StringVar(&genOpts.size, "size", "", "Output image size (1K / 2K / 4K)")

	generateCmd.MarkFlagsOneRequired("prompt", "prompt-file")
	generateCmd.MarkFlagsMutuallyExclusive("prompt", "prompt-file")
}
