package generator

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/history"
	"github.com/google/uuid"
	"google.golang.org/genai"
)

const DefaultModel = "gemini-3-pro-image-preview"

// Options contains generation options
type Options struct {
	AspectRatio string
	ImageSize   string
}

// GenerateResult contains the result of a generation
type GenerateResult struct {
	OutputImages []string
	TextResponse string
	TokenUsage   history.TokenUsage
}

// Generate performs image generation using Gemini API
func Generate(ctx context.Context, apiKey, prompt string, imagePaths []string, opts Options) (*GenerateResult, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("クライアント初期化に失敗しました: %w", err)
	}

	parts := []*genai.Part{genai.NewPartFromText(prompt)}
	for _, imgPath := range imagePaths {
		part, err := imagePartFromFile(imgPath)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	gcfg := &genai.GenerateContentConfig{ResponseModalities: []string{"IMAGE"}}
	if opts.AspectRatio != "" || opts.ImageSize != "" {
		gcfg.ImageConfig = &genai.ImageConfig{}
		if opts.AspectRatio != "" {
			gcfg.ImageConfig.AspectRatio = opts.AspectRatio
		}
		if opts.ImageSize != "" {
			gcfg.ImageConfig.ImageSize = strings.ToUpper(opts.ImageSize)
		}
	}

	contents := []*genai.Content{{Parts: parts}}
	resp, err := client.Models.GenerateContent(ctx, DefaultModel, contents, gcfg)
	if err != nil {
		return nil, fmt.Errorf("画像生成に失敗しました: %w", err)
	}

	result := &GenerateResult{
		TextResponse: strings.TrimSpace(resp.Text()),
	}

	if resp.UsageMetadata != nil {
		result.TokenUsage = history.TokenUsage{
			Prompt:     int(resp.UsageMetadata.PromptTokenCount),
			Candidates: int(resp.UsageMetadata.CandidatesTokenCount),
			Total:      int(resp.UsageMetadata.TotalTokenCount),
			Cached:     int(resp.UsageMetadata.CachedContentTokenCount),
			Thoughts:   int(resp.UsageMetadata.ThoughtsTokenCount),
		}
	}

	return result, nil
}

// SaveImages saves inline images from response to the specified directory
func SaveImages(resp *genai.GenerateContentResponse, dir, prefix string) ([]string, error) {
	if resp == nil {
		return nil, fmt.Errorf("レスポンスが空です")
	}

	runID := uuid.Must(uuid.NewV7())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
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
				return nil, fmt.Errorf("画像の保存に失敗しました (%s): %w", fullPath, err)
			}
			saved = append(saved, fileName)
			imageIndex++
		}
	}

	if len(saved) == 0 {
		return nil, fmt.Errorf("画像レスポンスが見つかりませんでした")
	}

	return saved, nil
}

func imagePartFromFile(path string) (*genai.Part, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("画像の読み込みに失敗しました (%s): %w", path, err)
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("画像 MIME を判定できませんでした (%s): %s", path, mimeType)
	}
	return genai.NewPartFromBytes(data, mimeType), nil
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
