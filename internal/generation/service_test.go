package generation

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

func TestService_Run_Success(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	svc := NewService(mock)

	var buf bytes.Buffer
	result, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf)

	require.NoError(t, err)
	assert.NotEmpty(t, result.EntryID, "expected entry ID")
	assert.Len(t, result.OutputImages, 1)

	// Verify history entry inline
	entryDir := filepath.Join(historyDir, result.EntryID)
	assert.DirExists(t, entryDir)
	assert.FileExists(t, filepath.Join(entryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(entryDir, "prompt.txt"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)

	// Verify stdout output
	output := buf.String()
	assert.Contains(t, output, "History ID:")
	assert.Contains(t, output, "Generated files:")

	// Verify mock was called correctly
	assert.Equal(t, 1, mock.callCount())
	lastCall := mock.lastCall()
	assert.Equal(t, "test-model", lastCall.Model)
	assert.Equal(t, "test prompt", lastCall.Prompt)
}

func TestService_Run_APIError(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	apiErr := errors.New("API error")
	mock := newErrorMock(apiErr)
	svc := NewService(mock)

	var buf bytes.Buffer
	_, err = svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate image")

	// Verify no history entries remain (cleanup)
	entries, listErr := history.ListEntries(historyDir)
	require.NoError(t, listErr)
	assert.Empty(t, entries, "expected no history entries after error")
}

func TestService_Run_MultipleImages(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	mock := newMultiImageMock(pngData, 3)
	svc := NewService(mock)

	var buf bytes.Buffer
	result, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf)

	require.NoError(t, err)
	assert.Len(t, result.OutputImages, 3)

	entryDir := filepath.Join(historyDir, result.EntryID)
	outputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 3)
}

func TestService_Run_Regenerate(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	svc := NewService(mock)

	// First generation
	var buf1 bytes.Buffer
	firstResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "first prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf1)
	require.NoError(t, err)

	// Regeneration with source entry
	var buf2 bytes.Buffer
	secondResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "first prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
		SourceEntryID:   firstResult.EntryID,
	}, historyDir, &buf2)
	require.NoError(t, err)

	// Verify different entry IDs
	assert.NotEqual(t, firstResult.EntryID, secondResult.EntryID, "expected different entry IDs for regeneration")

	// Verify both entries exist
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify mock was called twice
	assert.Equal(t, 2, mock.callCount())
}

func TestService_Run_ValidationError(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	svc := NewService(mock)

	t.Run("invalid aspect ratio", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := svc.Run(context.Background(), Spec{
			Model:           "test-model",
			Prompt:          "test prompt",
			ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
			InputImageNames: []string{"test.png"},
			AspectRatio:     "invalid",
		}, historyDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "aspect ratio")

		// Verify mock was NOT called (validation failed before API call)
		assert.Equal(t, 0, mock.callCount())
	})

	t.Run("invalid image size", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := svc.Run(context.Background(), Spec{
			Model:           "test-model",
			Prompt:          "test prompt",
			ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
			InputImageNames: []string{"test.png"},
			ImageSize:       "5K",
		}, historyDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "image size")
	})
}

func TestService_Run_WithAspectRatioAndSize(t *testing.T) {
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
	inputsDir := project.GetInputsDir(subprojectDir)
	require.NoError(t, os.WriteFile(filepath.Join(inputsDir, "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	svc := NewService(mock)

	var buf bytes.Buffer
	_, err = svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(inputsDir, "test.png")},
		InputImageNames: []string{"test.png"},
		AspectRatio:     "16:9",
		ImageSize:       "2K",
	}, historyDir, &buf)

	require.NoError(t, err)

	// Verify mock received correct parameters
	lastCall := mock.lastCall()
	assert.Equal(t, "16:9", lastCall.AspectRatio)
	assert.Equal(t, "2K", lastCall.ImageSize)
}

// Note: Scenario tests (TestScenario_*) have been moved to cmd/ layer
// where they can be tested through the handler pattern with DI support.
