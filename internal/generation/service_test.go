package generation

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Run_Success(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	var buf bytes.Buffer
	result, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf)

	require.NoError(t, err)
	assert.NotEmpty(t, result.EntryID, "expected entry ID")
	assert.Len(t, result.OutputImages, 1)

	// Verify history entry
	testutil.VerifyHistoryEntry(t, historyDir, result.EntryID)

	// Verify output images
	entryDir := filepath.Join(historyDir, result.EntryID)
	testutil.VerifyOutputImages(t, entryDir, 1)

	// Verify stdout output
	output := buf.String()
	assert.Contains(t, output, "History ID:")
	assert.Contains(t, output, "Generated files:")

	// Verify mock was called correctly
	assert.Equal(t, 1, mock.CallCount())
	lastCall := mock.LastCall()
	assert.Equal(t, "test-model", lastCall.Model)
	assert.Equal(t, "test prompt", lastCall.Prompt)
}

func TestService_Run_APIError(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	apiErr := errors.New("API error")
	mock := testutil.NewErrorMock(apiErr)
	svc := NewService(mock)

	var buf bytes.Buffer
	_, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
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

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewMultiImageMock(testutil.SamplePNG, 3)
	svc := NewService(mock)

	var buf bytes.Buffer
	result, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf)

	require.NoError(t, err)
	assert.Len(t, result.OutputImages, 3)

	entryDir := filepath.Join(historyDir, result.EntryID)
	testutil.VerifyOutputImages(t, entryDir, 3)
}

func TestService_Run_Regenerate(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	// First generation
	var buf1 bytes.Buffer
	firstResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "first prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &buf1)
	require.NoError(t, err)

	// Regeneration with source entry
	var buf2 bytes.Buffer
	secondResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "first prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
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
	assert.Equal(t, 2, mock.CallCount())
}

func TestService_Run_ValidationError(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	t.Run("invalid aspect ratio", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := svc.Run(context.Background(), Spec{
			Model:           "test-model",
			Prompt:          "test prompt",
			ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
			InputImageNames: []string{"test.png"},
			AspectRatio:     "invalid",
		}, historyDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "aspect ratio")

		// Verify mock was NOT called (validation failed before API call)
		assert.Equal(t, 0, mock.CallCount())
	})

	t.Run("invalid image size", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := svc.Run(context.Background(), Spec{
			Model:           "test-model",
			Prompt:          "test prompt",
			ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
			InputImageNames: []string{"test.png"},
			ImageSize:       "5K",
		}, historyDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "image size")
	})
}

func TestService_Run_WithAspectRatioAndSize(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	var buf bytes.Buffer
	_, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "test prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
		AspectRatio:     "16:9",
		ImageSize:       "2K",
	}, historyDir, &buf)

	require.NoError(t, err)

	// Verify mock received correct parameters
	lastCall := mock.LastCall()
	assert.Equal(t, "16:9", lastCall.AspectRatio)
	assert.Equal(t, "2K", lastCall.ImageSize)
}

// TestScenario_MultipleGenerations tests multiple consecutive generations.
func TestScenario_MultipleGenerations(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	// Generate multiple times
	var entryIDs []string
	for i := 0; i < 3; i++ {
		var buf bytes.Buffer
		result, err := svc.Run(context.Background(), Spec{
			Model:           "test-model",
			Prompt:          "test prompt",
			ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
			InputImageNames: []string{"test.png"},
		}, historyDir, &buf)
		require.NoError(t, err)
		entryIDs = append(entryIDs, result.EntryID)
	}

	// Verify all entries are unique
	assert.Len(t, entryIDs, 3)
	assert.NotEqual(t, entryIDs[0], entryIDs[1])
	assert.NotEqual(t, entryIDs[1], entryIDs[2])
	assert.NotEqual(t, entryIDs[0], entryIDs[2])

	// Verify all entries exist in history
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	// Verify mock was called 3 times
	assert.Equal(t, 3, mock.CallCount())
}

// TestScenario_GenerateAndEdit tests generate followed by edit.
func TestScenario_GenerateAndEdit(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	// Step 1: Generate
	var genBuf bytes.Buffer
	genResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "generate prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &genBuf)
	require.NoError(t, err)
	assert.NotEmpty(t, genResult.EntryID)
	assert.Len(t, genResult.OutputImages, 1)

	// Get the generated image path
	entryDir := filepath.Join(historyDir, genResult.EntryID)
	sourceImagePath := filepath.Join(entryDir, genResult.OutputImages[0])

	// Step 2: Edit
	var editBuf bytes.Buffer
	editResult, err := svc.Edit(context.Background(), EditSpec{
		Model:           "test-model",
		Prompt:          "edit prompt",
		SourceImagePath: sourceImagePath,
		HistoryDir:      historyDir,
		EntryID:         genResult.EntryID,
		SourceType:      "generate",
		SourceOutput:    genResult.OutputImages[0],
	}, &editBuf)
	require.NoError(t, err)
	assert.NotEmpty(t, editResult.EditID)
	assert.Len(t, editResult.OutputImages, 1)

	// Verify edit entry exists
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 1)

	// Verify mock was called twice (generate + edit)
	assert.Equal(t, 2, mock.CallCount())

	// Verify output contains expected text
	assert.Contains(t, genBuf.String(), "History ID:")
	assert.Contains(t, editBuf.String(), "Edit ID:")
}

// TestScenario_EditChain tests editing from a previous edit.
func TestScenario_EditChain(t *testing.T) {
	t.Parallel()

	projectRoot := testutil.CreateTestProject(t, "test-project")
	subprojectDir := testutil.CreateTestSubproject(t, projectRoot, "test-sub")
	historyDir := testutil.GetHistoryDir(subprojectDir)

	mock := testutil.NewSuccessMock(testutil.SamplePNG)
	svc := NewService(mock)

	// Step 1: Generate initial image
	var genBuf bytes.Buffer
	genResult, err := svc.Run(context.Background(), Spec{
		Model:           "test-model",
		Prompt:          "initial prompt",
		ImagePaths:      []string{filepath.Join(testutil.GetInputsDir(subprojectDir), "test.png")},
		InputImageNames: []string{"test.png"},
	}, historyDir, &genBuf)
	require.NoError(t, err)

	entryDir := filepath.Join(historyDir, genResult.EntryID)
	sourceImagePath := filepath.Join(entryDir, genResult.OutputImages[0])

	// Step 2: First edit (from generate)
	var edit1Buf bytes.Buffer
	edit1Result, err := svc.Edit(context.Background(), EditSpec{
		Model:           "test-model",
		Prompt:          "first edit",
		SourceImagePath: sourceImagePath,
		HistoryDir:      historyDir,
		EntryID:         genResult.EntryID,
		SourceType:      "generate",
		SourceOutput:    genResult.OutputImages[0],
	}, &edit1Buf)
	require.NoError(t, err)

	// Get first edit output path
	edit1OutputPath := history.GetEditOutputPath(entryDir, edit1Result.EditID, edit1Result.OutputImages[0])

	// Step 3: Second edit (from first edit)
	var edit2Buf bytes.Buffer
	edit2Result, err := svc.Edit(context.Background(), EditSpec{
		Model:           "test-model",
		Prompt:          "second edit",
		SourceImagePath: edit1OutputPath,
		HistoryDir:      historyDir,
		EntryID:         genResult.EntryID,
		SourceType:      "edit",
		SourceEditID:    edit1Result.EditID,
		SourceOutput:    edit1Result.OutputImages[0],
	}, &edit2Buf)
	require.NoError(t, err)

	// Verify all IDs are different
	assert.NotEqual(t, edit1Result.EditID, edit2Result.EditID)

	// Verify two edit entries exist
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 2)

	// Verify mock was called 3 times (1 generate + 2 edits)
	assert.Equal(t, 3, mock.CallCount())
}
