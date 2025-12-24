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

func TestEditHandler_Run_Success(t *testing.T) {
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
	genEntryID := entries[0].ID

	// Step 2: Edit from the generated entry
	editMock := newSuccessMock(pngData)
	handler := &editHandler{generator: editMock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), editOptions{
		latest: true,
		prompt: "edit prompt",
	}, subprojectDir, &buf)

	require.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Editing from generate:")
	assert.Contains(t, output, "Edit ID:")
	assert.Contains(t, output, "Edited files:")

	// Verify edit entry was created
	entryDir := filepath.Join(historyDir, genEntryID)
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 1)

	// Verify edit entry structure
	editsDir := filepath.Join(entryDir, "edits")
	editEntryDir := filepath.Join(editsDir, editEntries[0].ID)
	assert.DirExists(t, editEntryDir)
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-prompt.txt"))
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-meta.yaml"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(editEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)

	// Verify mock was called with correct prompt
	assert.Equal(t, 1, editMock.callCount())
	lastCall := editMock.lastCall()
	assert.Equal(t, "edit prompt", lastCall.Prompt)
}

func TestEditHandler_Run_ByID(t *testing.T) {
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
	entryID := entries[0].ID

	// Step 2: Edit by specific ID
	editMock := newSuccessMock(pngData)
	handler := &editHandler{generator: editMock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), editOptions{
		id:     entryID,
		prompt: "edit by id",
	}, subprojectDir, &buf)

	require.NoError(t, err)

	// Verify mock was called
	lastCall := editMock.lastCall()
	assert.Equal(t, "edit by id", lastCall.Prompt)

	// Verify edit entry was created with correct structure
	entryDir := filepath.Join(historyDir, entryID)
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 1)

	// Verify edit entry structure
	editsDir := filepath.Join(entryDir, "edits")
	editEntryDir := filepath.Join(editsDir, editEntries[0].ID)
	assert.DirExists(t, editEntryDir)
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-prompt.txt"))
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-meta.yaml"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(editEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)
}

func TestEditHandler_Run_APIError(t *testing.T) {
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
	genEntryID := entries[0].ID

	// Step 2: Edit with API error
	apiErr := errors.New("API error")
	editMock := newErrorMock(apiErr)
	handler := &editHandler{generator: editMock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), editOptions{
		latest: true,
		prompt: "edit prompt",
	}, subprojectDir, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to edit image")

	// Verify no edit entries remain (cleanup worked)
	entryDir := filepath.Join(historyDir, genEntryID)
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Empty(t, editEntries, "expected no edit entries after API error")
}

func TestEditHandler_Run_NoHistory(t *testing.T) {
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
	handler := &editHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), editOptions{
		latest: true,
		prompt: "edit prompt",
	}, subprojectDir, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get latest history")
}

func TestEditHandler_Run_EmptyPrompt(t *testing.T) {
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
	handler := &editHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), editOptions{
		latest: true,
		prompt: "",
	}, subprojectDir, &buf)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "specify --prompt or --prompt-file")
}

// TestScenario_GenerateAndEdit tests generate followed by edit.
func TestScenario_GenerateAndEdit(t *testing.T) {
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

	// Step 1: Generate
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "generate prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	// Verify generate entry created
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	genEntryID := entries[0].ID

	// Step 2: Edit
	editMock := newSuccessMock(pngData)
	editHandler := &editHandler{generator: editMock}

	var editBuf bytes.Buffer
	err = editHandler.run(context.Background(), editOptions{
		latest: true,
		prompt: "edit prompt",
	}, subprojectDir, &editBuf)
	require.NoError(t, err)

	// Verify edit entry created
	entryDir := filepath.Join(historyDir, genEntryID)
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 1)

	// Verify generate entry structure
	assert.DirExists(t, entryDir)
	assert.FileExists(t, filepath.Join(entryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(entryDir, "prompt.txt"))
	genOutputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, genOutputFiles, 1)

	// Verify edit entry structure
	editsDir := filepath.Join(entryDir, "edits")
	editEntryDir := filepath.Join(editsDir, editEntries[0].ID)
	assert.DirExists(t, editEntryDir)
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-prompt.txt"))
	assert.FileExists(t, filepath.Join(editEntryDir, "edit-meta.yaml"))
	editOutputFiles, err := filepath.Glob(filepath.Join(editEntryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, editOutputFiles, 1)

	// Verify both mocks were called
	assert.Equal(t, 1, genMock.callCount())
	assert.Equal(t, 1, editMock.callCount())

	// Verify output contains expected text
	assert.Contains(t, genBuf.String(), "History ID:")
	assert.Contains(t, editBuf.String(), "Edit ID:")
}

// TestScenario_EditChain tests editing from a previous edit.
func TestScenario_EditChain(t *testing.T) {
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

	// Step 1: Generate initial image
	genMock := newSuccessMock(pngData)
	genHandler := &generateHandler{generator: genMock}

	var genBuf bytes.Buffer
	err = genHandler.run(context.Background(), generateOptions{
		prompt: "initial prompt",
	}, subprojectDir, &genBuf)
	require.NoError(t, err)

	// Get the generate entry
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	genEntryID := entries[0].ID
	entryDir := filepath.Join(historyDir, genEntryID)

	// Step 2: First edit (from generate)
	edit1Mock := newSuccessMock(pngData)
	edit1Handler := &editHandler{generator: edit1Mock}

	var edit1Buf bytes.Buffer
	err = edit1Handler.run(context.Background(), editOptions{
		latest: true,
		prompt: "first edit",
	}, subprojectDir, &edit1Buf)
	require.NoError(t, err)

	// Get first edit entry
	editEntries, err := history.ListEditEntries(entryDir)
	require.NoError(t, err)
	require.Len(t, editEntries, 1)
	firstEditID := editEntries[0].ID

	// Step 3: Second edit (from first edit using --edit-latest)
	edit2Mock := newSuccessMock(pngData)
	edit2Handler := &editHandler{generator: edit2Mock}

	var edit2Buf bytes.Buffer
	err = edit2Handler.run(context.Background(), editOptions{
		latest:     true,
		editLatest: true,
		prompt:     "second edit",
	}, subprojectDir, &edit2Buf)
	require.NoError(t, err)

	// Verify two edit entries exist
	editEntries, err = history.ListEditEntries(entryDir)
	require.NoError(t, err)
	assert.Len(t, editEntries, 2)

	// Verify all IDs are different
	var secondEditID string
	for _, e := range editEntries {
		if e.ID != firstEditID {
			secondEditID = e.ID
			break
		}
	}
	assert.NotEmpty(t, secondEditID)
	assert.NotEqual(t, firstEditID, secondEditID)

	// Verify mocks were called
	assert.Equal(t, 1, genMock.callCount())
	assert.Equal(t, 1, edit1Mock.callCount())
	assert.Equal(t, 1, edit2Mock.callCount())

	// Verify second edit output shows "edit" source type
	assert.Contains(t, edit2Buf.String(), "Editing from edit:")

	// Verify all edit entries have correct structure
	editsDir := filepath.Join(entryDir, "edits")
	for _, editID := range []string{firstEditID, secondEditID} {
		editEntryDir := filepath.Join(editsDir, editID)
		assert.DirExists(t, editEntryDir)
		assert.FileExists(t, filepath.Join(editEntryDir, "edit-prompt.txt"))
		assert.FileExists(t, filepath.Join(editEntryDir, "edit-meta.yaml"))

		outputFiles, err := filepath.Glob(filepath.Join(editEntryDir, "output-*.png"))
		require.NoError(t, err)
		assert.Len(t, outputFiles, 1)
	}
}
