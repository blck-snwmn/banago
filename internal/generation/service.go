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

// Generator defines the interface for image generation.
// This allows for dependency injection and testing with mocks.
type Generator interface {
	Generate(ctx context.Context, params gemini.Params) *gemini.Result
}

// Result contains the output of a generation run.
type Result struct {
	EntryID      string
	OutputImages []string
}

// EditResult contains the output of an edit operation.
type EditResult struct {
	EditID       string
	OutputImages []string
}

// Service handles image generation with dependency injection support.
type Service struct {
	generator Generator
}

// NewService creates a new Service with the given generator.
func NewService(generator Generator) *Service {
	return &Service{generator: generator}
}

// Run executes the generation workflow and saves the result to history.
func (s *Service) Run(ctx context.Context, spec Spec, historyDir string, w io.Writer) (*Result, error) {
	// Validate inputs before any work
	if err := validateSpec(spec); err != nil {
		return nil, err
	}

	// Create history entry
	var entry *history.Entry
	if spec.SourceEntryID != "" {
		// For regeneration, create entry linked to source
		sourceEntry := &history.Entry{ID: spec.SourceEntryID}
		entry = history.NewEntryFromSource(sourceEntry)
	} else {
		entry = history.NewEntry()
	}

	entry.Generation.PromptFile = history.PromptFile
	entry.Generation.InputImages = spec.InputImageNames

	entryDir := entry.GetEntryDir(historyDir)

	// Create history directory and save prompt
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}
	if err := entry.SavePrompt(historyDir, spec.Prompt); err != nil {
		return nil, fmt.Errorf("failed to save prompt: %w", err)
	}

	// Save input images
	if err := entry.SaveInputImages(historyDir, spec.ImagePaths); err != nil {
		_, _ = fmt.Fprintf(w, "Warning: failed to save input images: %v\n", err)
	}

	// Call Gemini API
	result := s.generator.Generate(ctx, gemini.Params{
		Model:       spec.Model,
		Prompt:      spec.Prompt,
		ImagePaths:  spec.ImagePaths,
		AspectRatio: spec.AspectRatio,
		ImageSize:   spec.ImageSize,
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

	gemini.PrintOutput(w, result.Response, spec.Model)

	return &Result{
		EntryID:      entry.ID,
		OutputImages: entry.Result.OutputImages,
	}, nil
}

// Run executes the generation workflow (backward compatible wrapper).
func Run(ctx context.Context, apiKey string, spec Spec, historyDir string, w io.Writer) (*Result, error) {
	return NewService(gemini.NewClient(apiKey)).Run(ctx, spec, historyDir, w)
}

// Edit executes an edit operation on an existing image.
func (s *Service) Edit(ctx context.Context, spec EditSpec, historyDir string, w io.Writer) (*EditResult, error) {
	// Create edit entry
	editEntry := history.NewEditEntry()
	editEntry.Source = history.EditSource{
		Type:   spec.SourceType,
		EditID: spec.SourceEditID,
		Output: spec.SourceOutput,
	}

	entryDir := filepath.Join(historyDir, spec.EntryID)
	editDir := editEntry.GetEditEntryDir(entryDir)

	// Create edit directory and save prompt
	if err := os.MkdirAll(editDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create edit directory: %w", err)
	}
	if err := editEntry.SavePrompt(entryDir, spec.Prompt); err != nil {
		return nil, fmt.Errorf("failed to save edit prompt: %w", err)
	}

	// Call Gemini API
	result := s.generator.Generate(ctx, gemini.Params{
		Model:      spec.Model,
		Prompt:     spec.Prompt,
		ImagePaths: []string{spec.SourceImagePath},
	})

	if result.Error != nil {
		// Clean up edit directory on failure
		if err := editEntry.Cleanup(entryDir); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to clean up edit directory: %v\n", err)
		}
		return nil, fmt.Errorf("failed to edit image: %w", result.Error)
	}

	// Save edited images
	saved, saveErr := gemini.SaveImages(result.Response, editDir)
	if saveErr != nil {
		if err := editEntry.Cleanup(entryDir); err != nil {
			_, _ = fmt.Fprintf(w, "Warning: failed to clean up edit directory: %v\n", err)
		}
		return nil, saveErr
	}

	// Update entry with results
	editEntry.Result.Success = true
	for _, savedPath := range saved {
		editEntry.Result.OutputImages = append(editEntry.Result.OutputImages, filepath.Base(savedPath))
	}
	editEntry.Result.TokenUsage = result.TokenUsage

	if err := editEntry.Save(entryDir); err != nil {
		_, _ = fmt.Fprintf(w, "Warning: failed to save edit metadata: %v\n", err)
	}

	// Print output
	_, _ = fmt.Fprintf(w, "Edit ID: %s\n", editEntry.ID)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Edited files:")
	for _, savedPath := range saved {
		_, _ = fmt.Fprintf(w, "  %s\n", filepath.Base(savedPath))
	}

	gemini.PrintOutput(w, result.Response, spec.Model)

	return &EditResult{
		EditID:       editEntry.ID,
		OutputImages: editEntry.Result.OutputImages,
	}, nil
}

// Edit executes an edit operation (backward compatible wrapper).
func Edit(ctx context.Context, apiKey string, spec EditSpec, historyDir string, w io.Writer) (*EditResult, error) {
	return NewService(gemini.NewClient(apiKey)).Edit(ctx, spec, historyDir, w)
}
