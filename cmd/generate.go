package cmd

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cmp"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

const defaultModel = "gemini-3-pro-image-preview" // Nano Banana Pro (image preview)

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
	Short: "Nano Banana Pro に画像生成を依頼する",
	Long:  "Gemini 公式 SDK を用いて gemini-3-pro-image-preview (Nano Banana Pro) で画像を生成します。プロンプト必須、参考画像は複数指定可能です。",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		var promptText string
		if genOpts.prompt != "" {
			promptText = strings.TrimSpace(genOpts.prompt)
		}
		if genOpts.promptFile != "" {
			data, err := os.ReadFile(genOpts.promptFile)
			if err != nil {
				return fmt.Errorf("プロンプトファイルの読み込みに失敗しました: %w", err)
			}
			promptText = strings.TrimSpace(string(data))
		}
		if promptText == "" {
			return errors.New("プロンプトが空です。--prompt か --prompt-file で内容を指定してください")
		}
		if len(genOpts.images) == 0 {
			return errors.New("--image で最低 1 枚の画像を指定してください")
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.apiKey, Backend: genai.BackendGeminiAPI})
		if err != nil {
			return fmt.Errorf("クライアント初期化に失敗しました: %w", err)
		}

		parts := []*genai.Part{genai.NewPartFromText(promptText)}
		for _, imgPath := range genOpts.images {
			part, err := imagePartFromFile(imgPath)
			if err != nil {
				return err
			}
			parts = append(parts, part)
		}

		gcfg := &genai.GenerateContentConfig{ResponseModalities: []string{"IMAGE"}}
		if genOpts.aspect != "" || genOpts.size != "" {
			gcfg.ImageConfig = &genai.ImageConfig{}
			if genOpts.aspect != "" {
				gcfg.ImageConfig.AspectRatio = genOpts.aspect
			}
			if genOpts.size != "" {
				gcfg.ImageConfig.ImageSize = strings.ToUpper(genOpts.size)
			}
		}

		contents := []*genai.Content{{Parts: parts}}
		resp, err := client.Models.GenerateContent(ctx, defaultModel, contents, gcfg)
		if err != nil {
			return fmt.Errorf("画像生成に失敗しました: %w", err)
		}

		saved, err := saveInlineImages(resp, genOpts.outputDir, genOpts.prefix)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "生成完了: %s\n", strings.Join(saved, ", "))
		if text := strings.TrimSpace(resp.Text()); text != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "テキスト応答:")
			fmt.Fprintln(cmd.OutOrStdout(), text)
		}
		if resp.UsageMetadata != nil {
			usage := resp.UsageMetadata
			fmt.Fprintln(cmd.OutOrStdout(), "トークン使用量:")
			fmt.Fprintf(cmd.OutOrStdout(), "  prompt: %d\n", usage.PromptTokenCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  candidates: %d\n", usage.CandidatesTokenCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  total: %d\n", usage.TotalTokenCount)
			if usage.CachedContentTokenCount > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  cached: %d\n", usage.CachedContentTokenCount)
			}
			if usage.ThoughtsTokenCount > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  thoughts: %d\n", usage.ThoughtsTokenCount)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&genOpts.prompt, "prompt", "p", "", "生成に使うプロンプト (標準入力でなくフラグ指定)")
	generateCmd.Flags().StringVarP(&genOpts.promptFile, "prompt-file", "F", "", "プロンプトを含むテキストファイルへのパス")
	generateCmd.Flags().StringSliceVarP(&genOpts.images, "image", "i", nil, "プロンプトと一緒に送る画像ファイル (複数指定可)")
	generateCmd.Flags().StringVarP(&genOpts.outputDir, "output-dir", "o", "dist", "生成画像を保存するディレクトリ")
	generateCmd.Flags().StringVar(&genOpts.prefix, "prefix", "generated", "保存ファイル名のプレフィックス")
	generateCmd.Flags().StringVar(&genOpts.aspect, "aspect", "", "出力画像のアスペクト比 (例: 1:1, 16:9)")
	generateCmd.Flags().StringVar(&genOpts.size, "size", "", "出力画像のサイズ (1K / 2K / 4K。未指定ならモデル既定)")

	generateCmd.MarkFlagsOneRequired("prompt", "prompt-file")
	generateCmd.MarkFlagsMutuallyExclusive("prompt", "prompt-file")
	_ = generateCmd.MarkFlagRequired("image")
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

func saveInlineImages(resp *genai.GenerateContentResponse, dir, prefix string) ([]string, error) {
	if resp == nil {
		return nil, errors.New("レスポンスが空です")
	}
	runID := uuid.Must(uuid.NewV7())
	dir = cmp.Or(dir, "dist")
	prefix = cmp.Or(prefix, "generated")
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
			saved = append(saved, fullPath)
			imageIndex++
		}
	}

	if len(saved) == 0 {
		return nil, errors.New("画像レスポンスが見つかりませんでした")
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
	// .jpe などの揺れを防ぐため明示的に jpg に寄せる
	if strings.Contains(strings.ToLower(mimeType), "jpeg") {
		return ".jpg"
	}
	return ".bin"
}
