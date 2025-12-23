package cmd

import (
	"cmp"
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
	Short: "Image generation CLI powered by Gemini",
	Long:  "CLI tool to generate images using Gemini 3 Pro Image Preview with prompts and reference images",
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
	rootCmd.PersistentFlags().StringVar(&cfg.apiKey, "api-key", "", "Gemini API key (defaults to GEMINI_API_KEY env var)")
}

// requireAPIKey checks if the API key is set and returns an error if not.
// Should be called by commands that require the API key (generate, regenerate).
func requireAPIKey() error {
	cfg.apiKey = cmp.Or(strings.TrimSpace(cfg.apiKey), strings.TrimSpace(os.Getenv("GEMINI_API_KEY")))
	if cfg.apiKey == "" {
		return errors.New("API key is required. Set --api-key or GEMINI_API_KEY environment variable")
	}
	return nil
}
