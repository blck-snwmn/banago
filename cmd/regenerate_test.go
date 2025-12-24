package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScenario_Regenerate_Latest tests regeneration using --latest flag.
func TestScenario_Regenerate_Latest(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
	historyDir := filepath.Join(subprojectDir, "history")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Step 1: Generate initial entry
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "original prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	firstEntryID := entries[0].ID

	// Step 2: Regenerate using --latest
	regenMock := newSuccessMock(pngData)
	regenHandler := &regenerateHandler{generator: regenMock}

	var regenBuf bytes.Buffer
	err = regenHandler.run(context.Background(), regenerateOptions{
		latest: true,
	}, subprojectDir, &regenBuf)
	require.NoError(t, err)

	// Verify output
	output := regenBuf.String()
	assert.Contains(t, output, "Regenerating from history:")
	assert.Contains(t, output, firstEntryID)
	assert.Contains(t, output, "History ID:")

	// Verify new entry was created
	entries, err = history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "expected 2 entries (original + regenerated)")

	// Find the regenerated entry (not the first one)
	var regenEntryID string
	for _, e := range entries {
		if e.ID != firstEntryID {
			regenEntryID = e.ID
			break
		}
	}
	require.NotEmpty(t, regenEntryID)

	// Verify regenerated entry structure
	regenEntryDir := filepath.Join(historyDir, regenEntryID)
	assert.DirExists(t, regenEntryDir)
	assert.FileExists(t, filepath.Join(regenEntryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(regenEntryDir, "prompt.txt"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(regenEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)

	// Verify regeneration used the same prompt
	assert.Equal(t, "original prompt", regenMock.lastCall().Prompt)
}

// TestScenario_Regenerate_ByID tests regeneration using specific ID.
func TestScenario_Regenerate_ByID(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
	historyDir := filepath.Join(subprojectDir, "history")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Step 1: Generate initial entry
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "specific prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	entryID := entries[0].ID

	// Step 2: Regenerate using specific ID
	regenMock := newSuccessMock(pngData)
	regenHandler := &regenerateHandler{generator: regenMock}

	var regenBuf bytes.Buffer
	err = regenHandler.run(context.Background(), regenerateOptions{
		id: entryID,
	}, subprojectDir, &regenBuf)
	require.NoError(t, err)

	// Verify mock was called with correct prompt
	assert.Equal(t, "specific prompt", regenMock.lastCall().Prompt)

	// Verify new entry was created with correct structure
	entries, err = history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// Find the regenerated entry
	var regenEntryID string
	for _, e := range entries {
		if e.ID != entryID {
			regenEntryID = e.ID
			break
		}
	}
	require.NotEmpty(t, regenEntryID)

	// Verify regenerated entry structure
	regenEntryDir := filepath.Join(historyDir, regenEntryID)
	assert.DirExists(t, regenEntryDir)
	assert.FileExists(t, filepath.Join(regenEntryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(regenEntryDir, "prompt.txt"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(regenEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)
}

// TestScenario_Regenerate_APIError tests cleanup when API fails.
func TestScenario_Regenerate_APIError(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
	historyDir := filepath.Join(subprojectDir, "history")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Step 1: Generate initial entry
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "original prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	// Step 2: Regenerate with API error
	apiErr := errors.New("API error")
	regenMock := newErrorMock(apiErr)
	regenHandler := &regenerateHandler{generator: regenMock}

	var regenBuf bytes.Buffer
	err = regenHandler.run(context.Background(), regenerateOptions{
		latest: true,
	}, subprojectDir, &regenBuf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate image")

	// Verify only original entry remains (cleanup worked)
	entries, err = history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "expected only original entry after API error")
}

// TestScenario_Regenerate_NoHistory tests error when no history exists.
func TestScenario_Regenerate_NoHistory(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &regenerateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), regenerateOptions{
		latest: true,
	}, subprojectDir, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get latest history")
}

// TestScenario_Regenerate_InvalidID tests error when ID doesn't exist.
func TestScenario_Regenerate_InvalidID(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &regenerateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), regenerateOptions{
		id: "nonexistent-id",
	}, subprojectDir, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get history entry")
}

// TestScenario_Regenerate_WithOverrides tests regeneration with aspect/size overrides.
func TestScenario_Regenerate_WithOverrides(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
	historyDir := filepath.Join(subprojectDir, "history")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Step 1: Generate initial entry
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	firstEntryID := entries[0].ID

	// Step 2: Regenerate with overrides
	regenMock := newSuccessMock(pngData)
	regenHandler := &regenerateHandler{generator: regenMock}

	var regenBuf bytes.Buffer
	err = regenHandler.run(context.Background(), regenerateOptions{
		latest: true,
		aspect: "16:9",
		size:   "4K",
	}, subprojectDir, &regenBuf)
	require.NoError(t, err)

	// Verify mock received overridden parameters
	lastCall := regenMock.lastCall()
	assert.Equal(t, "16:9", lastCall.AspectRatio)
	assert.Equal(t, "4K", lastCall.ImageSize)

	// Find the regenerated entry
	entries, err = history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	var regenEntryID string
	for _, e := range entries {
		if e.ID != firstEntryID {
			regenEntryID = e.ID
			break
		}
	}
	require.NotEmpty(t, regenEntryID)

	// Verify regenerated entry structure
	regenEntryDir := filepath.Join(historyDir, regenEntryID)
	assert.DirExists(t, regenEntryDir)
	assert.FileExists(t, filepath.Join(regenEntryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(regenEntryDir, "prompt.txt"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(regenEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)
}

// TestScenario_Regenerate_Multiple tests multiple consecutive generations and regenerations.
func TestScenario_Regenerate_Multiple(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
	historyDir := filepath.Join(subprojectDir, "history")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Step 1: Generate initial entry
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "initial generation",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	firstEntryID := entries[0].ID

	// Step 2: Regenerate from the first entry
	regenMock := newSuccessMock(pngData)
	regenHandler := &regenerateHandler{generator: regenMock}

	var regenBuf bytes.Buffer
	err = regenHandler.run(context.Background(), regenerateOptions{
		latest: true,
	}, subprojectDir, &regenBuf)
	require.NoError(t, err)

	// Verify two entries exist with different IDs
	entries, err = history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	var secondEntryID string
	for _, e := range entries {
		if e.ID != firstEntryID {
			secondEntryID = e.ID
			break
		}
	}
	assert.NotEmpty(t, secondEntryID, "second entry should have different ID")
	assert.NotEqual(t, firstEntryID, secondEntryID)

	// Verify regeneration used the same prompt
	assert.Equal(t, "initial generation", regenMock.lastCall().Prompt)

	// Verify both entries have correct structure
	for _, entryID := range []string{firstEntryID, secondEntryID} {
		entryDir := filepath.Join(historyDir, entryID)
		assert.DirExists(t, entryDir)
		assert.FileExists(t, filepath.Join(entryDir, "meta.yaml"))
		assert.FileExists(t, filepath.Join(entryDir, "prompt.txt"))

		outputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
		require.NoError(t, err)
		assert.Len(t, outputFiles, 1)
	}
}
