package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// Params holds parameters for image generation
type Params struct {
	Model       string
	Prompt      string
	ImagePaths  []string
	AspectRatio string
	ImageSize   string
}

// Result holds the result of image generation
type Result struct {
	Response   *genai.GenerateContentResponse
	Error      error
	TokenUsage TokenUsage
}

// Client calls the real Gemini API for image generation
type Client struct {
	apiKey string
}

// NewClient creates a new Client with the given API key
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

// Generate calls the Gemini API to generate images
func (c *Client) Generate(ctx context.Context, params Params) *Result {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return &Result{Error: fmt.Errorf("failed to initialize client: %w", err)}
	}

	parts := []*genai.Part{genai.NewPartFromText(params.Prompt)}
	for _, imgPath := range params.ImagePaths {
		part, err := ImagePartFromFile(imgPath)
		if err != nil {
			return &Result{Error: err}
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

	result := &Result{
		Response: resp,
		Error:    err,
	}

	if err == nil && resp != nil && resp.UsageMetadata != nil {
		result.TokenUsage = TokenUsage{
			Prompt:     int(resp.UsageMetadata.PromptTokenCount),
			Candidates: int(resp.UsageMetadata.CandidatesTokenCount),
			Total:      int(resp.UsageMetadata.TotalTokenCount),
			Cached:     int(resp.UsageMetadata.CachedContentTokenCount),
			Thoughts:   int(resp.UsageMetadata.ThoughtsTokenCount),
		}
	}

	return result
}

// Generate calls the Gemini API to generate images (backward compatible wrapper)
func Generate(ctx context.Context, apiKey string, params Params) *Result {
	return NewClient(apiKey).Generate(ctx, params)
}

// PrintOutput prints the generation result to the writer
func PrintOutput(w io.Writer, resp *genai.GenerateContentResponse, model string) {
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

// ImagePartFromFile reads an image file and returns a genai.Part
func ImagePartFromFile(path string) (*genai.Part, error) {
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

// SaveImages saves generated images from the response to the specified directory
func SaveImages(resp *genai.GenerateContentResponse, dir string) ([]string, error) {
	if resp == nil {
		return nil, errors.New("response is empty")
	}
	runID := uuid.Must(uuid.NewV7())
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
			ext := NormalizeExt(part.InlineData.MIMEType)
			fileName := fmt.Sprintf("output-%s-%d%s", runID, imageIndex+1, ext)
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

// NormalizeExt returns the file extension for a given MIME type
func NormalizeExt(mimeType string) string {
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
