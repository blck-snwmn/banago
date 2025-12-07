package cmd

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

const defaultModel = "gemini-3-pro-image-preview"

type generateOptions struct {
	prompt     string
	promptFile string
	images     []string
	outputDir  string
	prefix     string
	aspect     string
	size       string
}

var genOpts generateOptions

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate images",
	Long: `Generate images using the Gemini API.

When running inside a subproject:
  - input_images from config.yaml are automatically used
  - Results are saved to history/
  - Additional images can be specified with --image

When running outside a subproject:
  - Images must be specified with --image
  - Results are saved to --output-dir`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := requireAPIKey(); err != nil {
			return err
		}

		// Get prompt
		var promptText string
		if genOpts.prompt != "" {
			promptText = strings.TrimSpace(genOpts.prompt)
		}
		if genOpts.promptFile != "" {
			data, err := os.ReadFile(genOpts.promptFile)
			if err != nil {
				return fmt.Errorf("failed to read prompt file: %w", err)
			}
			promptText = strings.TrimSpace(string(data))
		}
		if promptText == "" {
			return errors.New("prompt is empty. Specify with --prompt or --prompt-file")
		}

		// Check if we're in a subproject
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, projectErr := project.FindProjectRoot(cwd)
		var subprojectName string
		var subprojectDir string
		var subprojectCfg *config.SubprojectConfig

		if projectErr == nil {
			subprojectName, err = project.FindCurrentSubproject(projectRoot, cwd)
			if err == nil {
				subprojectDir = config.GetSubprojectDir(projectRoot, subprojectName)
				subprojectCfg, err = config.LoadSubprojectConfig(subprojectDir)
				if err != nil {
					return fmt.Errorf("failed to load subproject config: %w", err)
				}
			}
		}

		// Collect image paths
		var imagePaths []string
		if subprojectCfg != nil && len(subprojectCfg.InputImages) > 0 {
			inputsDir := config.GetInputsDir(subprojectDir)
			for _, img := range subprojectCfg.InputImages {
				imagePaths = append(imagePaths, filepath.Join(inputsDir, img))
			}
		}
		// Add any additional images from --image flag
		imagePaths = append(imagePaths, genOpts.images...)

		if len(imagePaths) == 0 {
			return errors.New("no images specified. Use --image or set input_images in subproject config.yaml")
		}

		// Determine aspect ratio and size
		aspect := genOpts.aspect
		size := genOpts.size
		if subprojectCfg != nil {
			if aspect == "" && subprojectCfg.AspectRatio != "" {
				aspect = subprojectCfg.AspectRatio
			}
			if size == "" && subprojectCfg.ImageSize != "" {
				size = subprojectCfg.ImageSize
			}
		}

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
		resp, err := client.Models.GenerateContent(ctx, defaultModel, contents, gcfg)

		w := cmd.OutOrStdout()

		// Handle generation result
		if subprojectCfg != nil {
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

			// Save context file
			if subprojectCfg.ContextFile != "" {
				contextPath := filepath.Join(subprojectDir, subprojectCfg.ContextFile)
				if _, statErr := os.Stat(contextPath); statErr == nil {
					if err := entry.SaveContextFile(historyDir, contextPath); err != nil {
						_, _ = fmt.Fprintf(w, "Warning: failed to save context file: %v\n", err)
					} else {
						entry.Generation.ContextFile = history.ContextFile
					}
				}
			}

			// Save character file
			if subprojectCfg.CharacterFile != "" {
				characterPath := filepath.Join(projectRoot, config.CharactersDir, subprojectCfg.CharacterFile)
				if _, statErr := os.Stat(characterPath); statErr == nil {
					if err := entry.SaveCharacterFile(historyDir, characterPath); err != nil {
						_, _ = fmt.Fprintf(w, "Warning: failed to save character file: %v\n", err)
					} else {
						entry.Generation.CharacterFile = history.CharacterFile
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
		} else {
			// Legacy mode: save to output directory
			if err != nil {
				return fmt.Errorf("failed to generate image: %w", err)
			}

			saved, err := saveInlineImages(resp, genOpts.outputDir, genOpts.prefix)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(w, "Generated files:")
			for _, path := range saved {
				_, _ = fmt.Fprintf(w, "  %s\n", filepath.Base(path))
			}
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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&genOpts.prompt, "prompt", "p", "", "Prompt for generation")
	generateCmd.Flags().StringVarP(&genOpts.promptFile, "prompt-file", "F", "", "Path to text file containing prompt")
	generateCmd.Flags().StringSliceVarP(&genOpts.images, "image", "i", nil, "Image files to send with prompt (can specify multiple)")
	generateCmd.Flags().StringVarP(&genOpts.outputDir, "output-dir", "o", "dist", "Directory to save generated images (outside subproject)")
	generateCmd.Flags().StringVar(&genOpts.prefix, "prefix", "generated", "Filename prefix for saved files (outside subproject)")
	generateCmd.Flags().StringVar(&genOpts.aspect, "aspect", "", "Output image aspect ratio (e.g., 1:1, 16:9)")
	generateCmd.Flags().StringVar(&genOpts.size, "size", "", "Output image size (1K / 2K / 4K)")

	generateCmd.MarkFlagsOneRequired("prompt", "prompt-file")
	generateCmd.MarkFlagsMutuallyExclusive("prompt", "prompt-file")
}

func imagePartFromFile(path string) (*genai.Part, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image (%s): %w", path, err)
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("could not determine image MIME type (%s): %s", path, mimeType)
	}
	return genai.NewPartFromBytes(data, mimeType), nil
}

func saveInlineImages(resp *genai.GenerateContentResponse, dir, prefix string) ([]string, error) {
	if resp == nil {
		return nil, errors.New("response is empty")
	}
	runID := uuid.Must(uuid.NewV7())
	dir = cmp.Or(dir, "dist")
	prefix = cmp.Or(prefix, "generated")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var saved []string
	imageIndex := 0
	for _, cand := range resp.Candidates {
		if cand == nil || cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part == nil || part.InlineData == nil || len(part.InlineData.Data) == 0 {
				continue
			}
			mimeType := part.InlineData.MIMEType
			ext := normalizeExt(mimeType)

			fileName := fmt.Sprintf("%s-%s-%d%s", prefix, runID, imageIndex+1, ext)
			fullPath := filepath.Join(dir, fileName)
			if err := os.WriteFile(fullPath, part.InlineData.Data, 0o644); err != nil {
				return nil, fmt.Errorf("failed to save image (%s): %w", fullPath, err)
			}
			saved = append(saved, fullPath)
			imageIndex++
		}
	}

	if len(saved) == 0 {
		return nil, errors.New("no image response found")
	}

	return saved, nil
}

func normalizeExt(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	case "image/avif":
		return ".avif"
	case "image/heic":
		return ".heic"
	case "image/heif":
		return ".heif"
	case "image/tiff", "image/tif":
		return ".tiff"
	}
	if strings.Contains(strings.ToLower(mimeType), "jpeg") {
		return ".jpg"
	}
	return ".bin"
}
