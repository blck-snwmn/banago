package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/blck-snwmn/banago/internal/history"
	"google.golang.org/genai"
)

// generationParams holds parameters for image generation
type generationParams struct {
	Model       string
	Prompt      string
	ImagePaths  []string
	AspectRatio string
	ImageSize   string
}

// generationResult holds the result of image generation
type generationResult struct {
	Response   *genai.GenerateContentResponse
	Error      error
	TokenUsage history.TokenUsage
}

// generateImages calls the Gemini API to generate images
func generateImages(ctx context.Context, apiKey string, params generationParams) *generationResult {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return &generationResult{Error: fmt.Errorf("failed to initialize client: %w", err)}
	}

	parts := []*genai.Part{genai.NewPartFromText(params.Prompt)}
	for _, imgPath := range params.ImagePaths {
		part, err := imagePartFromFile(imgPath)
		if err != nil {
			return &generationResult{Error: err}
		}
		parts = append(parts, part)
	}

	gcfg := &genai.GenerateContentConfig{ResponseModalities: []string{"IMAGE"}}
	if params.AspectRatio != "" || params.ImageSize != "" {
		gcfg.ImageConfig = &genai.ImageConfig{}
		if params.AspectRatio != "" {
			gcfg.ImageConfig.AspectRatio = params.AspectRatio
		}
		if params.ImageSize != "" {
			gcfg.ImageConfig.ImageSize = strings.ToUpper(params.ImageSize)
		}
	}

	contents := []*genai.Content{{Parts: parts}}
	resp, err := client.Models.GenerateContent(ctx, params.Model, contents, gcfg)

	result := &generationResult{
		Response: resp,
		Error:    err,
	}

	if err == nil && resp != nil && resp.UsageMetadata != nil {
		result.TokenUsage = history.TokenUsage{
			Prompt:     int(resp.UsageMetadata.PromptTokenCount),
			Candidates: int(resp.UsageMetadata.CandidatesTokenCount),
			Total:      int(resp.UsageMetadata.TotalTokenCount),
			Cached:     int(resp.UsageMetadata.CachedContentTokenCount),
			Thoughts:   int(resp.UsageMetadata.ThoughtsTokenCount),
		}
	}

	return result
}

// printGenerationOutput prints the generation result to the writer
func printGenerationOutput(w io.Writer, resp *genai.GenerateContentResponse, model string) {
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
}
