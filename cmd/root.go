package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var cfg = struct {
	apiKey string
}{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "banago",
	Short: "Gemini ベースの画像生成 CLI",
	Long:  "プロンプトや手元の画像を指定して Gemini 3 Pro Image Preview に画像生成を依頼する CLI",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.apiKey, "api-key", "", "Gemini API キー。未指定なら環境変数 GEMINI_API_KEY を利用")
}

// requireAPIKey checks if the API key is set and returns an error if not.
// Should be called by commands that require the API key (generate, regenerate).
func requireAPIKey() error {
	if strings.TrimSpace(cfg.apiKey) == "" {
		cfg.apiKey = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	}
	if cfg.apiKey == "" {
		return errors.New("API キーがありません。--api-key か環境変数 GEMINI_API_KEY を設定してください")
	}
	return nil
}
