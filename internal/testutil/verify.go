package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VerifyHistoryEntry checks that a history entry exists and has valid structure.
func VerifyHistoryEntry(t *testing.T, historyDir, entryID string) {
	t.Helper()

	entryDir := filepath.Join(historyDir, entryID)

	// Check entry directory exists
	assert.DirExists(t, entryDir, "entry directory should exist")

	// Check meta.yaml exists
	metaPath := filepath.Join(entryDir, "meta.yaml")
	assert.FileExists(t, metaPath, "meta.yaml should exist")

	// Check prompt.txt exists
	promptPath := filepath.Join(entryDir, "prompt.txt")
	assert.FileExists(t, promptPath, "prompt.txt should exist")
}

// VerifyOutputImages checks that the expected number of output images exist.
func VerifyOutputImages(t *testing.T, entryDir string, expectedCount int) {
	t.Helper()

	// Load entry to get output image names
	historyDir := filepath.Dir(entryDir)
	entryID := filepath.Base(entryDir)
	entry, err := history.GetEntryByID(historyDir, entryID)
	require.NoError(t, err, "failed to load entry")

	assert.Len(t, entry.Result.OutputImages, expectedCount, "output images count mismatch")

	// Check each output image file exists
	for _, img := range entry.Result.OutputImages {
		imgPath := filepath.Join(entryDir, img)
		assert.FileExists(t, imgPath, "output image should exist: %s", img)
	}
}

// VerifyProjectStructure checks that a project has the expected structure.
func VerifyProjectStructure(t *testing.T, projectRoot string) {
	t.Helper()

	assert.FileExists(t, filepath.Join(projectRoot, "banago.yaml"), "banago.yaml should exist")
	assert.DirExists(t, filepath.Join(projectRoot, "characters"), "characters directory should exist")
	assert.DirExists(t, filepath.Join(projectRoot, "subprojects"), "subprojects directory should exist")
}

// VerifySubprojectStructure checks that a subproject has the expected structure.
func VerifySubprojectStructure(t *testing.T, subprojectDir string) {
	t.Helper()

	assert.FileExists(t, filepath.Join(subprojectDir, "config.yaml"), "config.yaml should exist")
	assert.DirExists(t, filepath.Join(subprojectDir, "inputs"), "inputs directory should exist")
	assert.DirExists(t, filepath.Join(subprojectDir, "history"), "history directory should exist")
}

// AssertNoError fails the test if err is not nil.
// Deprecated: Use require.NoError directly.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// AssertError fails the test if err is nil.
// Deprecated: Use require.Error directly.
func AssertError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
}

// AssertErrorContains checks that the error message contains the expected substring.
// Deprecated: Use require.Error and assert.Contains directly.
func AssertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), expected)
}

// FileExists checks if a file exists (helper for tests).
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists (helper for tests).
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
