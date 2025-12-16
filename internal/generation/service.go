package generation

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/gemini"
	"github.com/blck-snwmn/banago/internal/history"
)

// Result contains the output of a generation run.
type Result struct {
	EntryID      string
	OutputImages []string
}

// Run executes the generation workflow and saves the result to history.
func Run(ctx context.Context, apiKey string, genCtx *Context, historyDir string, w io.Writer) (*Result, error) {
	// Create history entry
	var entry *history.Entry
	if genCtx.SourceEntryID != "" {
		// For regeneration, create entry linked to source
		sourceEntry := &history.Entry{ID: genCtx.SourceEntryID}
		entry = history.NewEntryFromSource(sourceEntry)
	} else {
		entry = history.NewEntry()
	}

	entry.Generation.PromptFile = history.PromptFile
	entry.Generation.InputImages = genCtx.InputImageNames

	entryDir := entry.GetEntryDir(historyDir)

	// Create history directory and save prompt
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}
	if err := entry.SavePrompt(historyDir, genCtx.Prompt); err != nil {
		return nil, fmt.Errorf("failed to save prompt: %w", err)
	}

	// Save input images
	if err := entry.SaveInputImages(historyDir, genCtx.ImagePaths); err != nil {
		_, _ = fmt.Fprintf(w, "Warning: failed to save input images: %v\n", err)
	}

	// Call Gemini API
	result := gemini.Generate(ctx, apiKey, gemini.Params{
		Model:       genCtx.Model,
		Prompt:      genCtx.Prompt,
		ImagePaths:  genCtx.ImagePaths,
		AspectRatio: genCtx.AspectRatio,
		ImageSize:   genCtx.ImageSize,
	})

	if result.Error != nil {
		// Clean up history directory on generation failure
		if err := entry.Cleanup(historyDir); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to clean up history directory: %v\n", err)
		}
		return nil, fmt.Errorf("failed to generate image: %w", result.Error)
	}

	// Save generated images
	saved, saveErr := gemini.SaveImages(result.Response, entryDir)
	if saveErr != nil {
		// Clean up history directory on save failure
		if err := entry.Cleanup(historyDir); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to clean up history directory: %v\n", err)
		}
		return nil, saveErr
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

	// Print output
	_, _ = fmt.Fprintf(w, "History ID: %s\n", entry.ID)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Generated files:")
	for _, s := range saved {
		_, _ = fmt.Fprintf(w, "  %s\n", filepath.Base(s))
	}

	gemini.PrintOutput(w, result.Response, genCtx.Model)

	return &Result{
		EntryID:      entry.ID,
		OutputImages: entry.Result.OutputImages,
	}, nil
}
