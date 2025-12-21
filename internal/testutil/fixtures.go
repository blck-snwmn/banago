package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
)

// SamplePNG is a minimal valid PNG image for testing (1x1 transparent pixel).
var SamplePNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk header
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 dimensions
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, // bit depth, color type, CRC
	0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, // IDAT chunk
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, // compressed data
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, // more data
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, // IEND chunk
	0x42, 0x60, 0x82, // CRC
}

// CreateTestProject creates a test project structure in a temporary directory.
// Returns the project root path.
func CreateTestProject(t *testing.T, name string) string {
	t.Helper()

	projectRoot := t.TempDir()

	// Create project config
	cfg := config.NewProjectConfig(name)
	if err := cfg.Save(projectRoot); err != nil {
		t.Fatalf("failed to save project config: %v", err)
	}

	// Create required directories
	dirs := []string{"characters", "subprojects"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectRoot, dir), 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	return projectRoot
}

// CreateTestSubproject creates a subproject with an input image.
// Returns the subproject directory path.
func CreateTestSubproject(t *testing.T, projectRoot, name string) string {
	t.Helper()

	subprojectDir := filepath.Join(projectRoot, "subprojects", name)

	// Create subproject directories
	dirs := []string{
		subprojectDir,
		filepath.Join(subprojectDir, "inputs"),
		filepath.Join(subprojectDir, "history"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create subproject config
	cfg := config.NewSubprojectConfig(name)
	cfg.InputImages = []string{"test.png"}
	if err := cfg.Save(subprojectDir); err != nil {
		t.Fatalf("failed to save subproject config: %v", err)
	}

	// Create test input image
	inputPath := filepath.Join(subprojectDir, "inputs", "test.png")
	if err := os.WriteFile(inputPath, SamplePNG, 0o644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	return subprojectDir
}

// GetHistoryDir returns the history directory path for a subproject.
func GetHistoryDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, "history")
}

// GetInputsDir returns the inputs directory path for a subproject.
func GetInputsDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, "inputs")
}

// CreateHistoryEntry creates a history entry with the given prompt.
// Returns the entry and entry directory path.
func CreateHistoryEntry(t *testing.T, historyDir, prompt string) (*history.Entry, string) {
	t.Helper()

	entry := history.NewEntry()
	entry.Generation.PromptFile = history.PromptFile
	entry.Generation.InputImages = []string{"test.png"}
	entry.Result.Success = true
	entry.Result.OutputImages = []string{"output-test-1.png"}

	entryDir := entry.GetEntryDir(historyDir)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry directory: %v", err)
	}

	// Save prompt
	if err := entry.SavePrompt(historyDir, prompt); err != nil {
		t.Fatalf("failed to save prompt: %v", err)
	}

	// Create output image
	outputPath := filepath.Join(entryDir, "output-test-1.png")
	if err := os.WriteFile(outputPath, SamplePNG, 0o644); err != nil {
		t.Fatalf("failed to create output image: %v", err)
	}

	// Save entry metadata
	if err := entry.Save(historyDir); err != nil {
		t.Fatalf("failed to save entry: %v", err)
	}

	return entry, entryDir
}
