package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createHistoryEntryForCLI creates a history entry for testing CLI commands that read history.
// This is intentionally a local helper because:
// - CLI integration tests can't use DI (they run the compiled binary)
// - This should ONLY be used for testing read-only commands like "history"
// - For generate/regenerate/edit tests, use the handler pattern with mock generators instead.
func createHistoryEntryForCLI(t *testing.T, historyDir, prompt string) *history.Entry {
	t.Helper()

	// Read sample PNG
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err, "failed to read sample.png")

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

	// Create input image in entry directory
	inputPath := filepath.Join(entryDir, "test.png")
	if err := os.WriteFile(inputPath, pngData, 0o644); err != nil {
		t.Fatalf("failed to create input image: %v", err)
	}

	// Create output image
	outputPath := filepath.Join(entryDir, "output-test-1.png")
	if err := os.WriteFile(outputPath, pngData, 0o644); err != nil {
		t.Fatalf("failed to create output image: %v", err)
	}

	// Save entry metadata
	if err := entry.Save(historyDir); err != nil {
		t.Fatalf("failed to save entry: %v", err)
	}

	return entry
}

func TestIntegration_Init(t *testing.T) {
	t.Parallel()

	t.Run("creates project structure", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cmd := exec.Command(testBinPath, "init")
		cmd.Dir = tmpDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init failed: %v\noutput: %s", err, output)
		}

		// Verify output
		assert.Contains(t, string(output), "Initialized banago project")

		// Verify project structure inline
		assert.FileExists(t, filepath.Join(tmpDir, "banago.yaml"))
		assert.DirExists(t, filepath.Join(tmpDir, "characters"))
		assert.DirExists(t, filepath.Join(tmpDir, "subprojects"))
	})

	t.Run("fails if already initialized", func(t *testing.T) {
		t.Parallel()

		// Create a project first
		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))

		cmd := exec.Command(testBinPath, "init")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("init should fail when project already exists")
		}

		assert.Contains(t, string(output), "already exists")
	})

	t.Run("force overwrites existing", func(t *testing.T) {
		t.Parallel()

		// Create a project first
		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))

		cmd := exec.Command(testBinPath, "init", "--force")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init --force failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "Initialized banago project")
	})

	t.Run("uses custom name", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cmd := exec.Command(testBinPath, "init", "--name", "custom-project")
		cmd.Dir = tmpDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "custom-project")
	})
}

func TestIntegration_SubprojectCreate(t *testing.T) {
	t.Parallel()

	t.Run("creates subproject", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))

		cmd := exec.Command(testBinPath, "subproject", "create", "my-subproject")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("subproject create failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "my-subproject")

		// Verify subproject structure inline
		subprojectDir := filepath.Join(projectRoot, "subprojects", "my-subproject")
		assert.FileExists(t, filepath.Join(subprojectDir, "config.yaml"))
		assert.DirExists(t, filepath.Join(subprojectDir, "inputs"))
		assert.DirExists(t, filepath.Join(subprojectDir, "history"))
	})

	t.Run("fails for duplicate name", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "existing-sub", ""))

		cmd := exec.Command(testBinPath, "subproject", "create", "existing-sub")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("subproject create should fail for duplicate name")
		}

		assert.Contains(t, string(output), "already exists")
	})

	t.Run("fails outside project", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		cmd := exec.Command(testBinPath, "subproject", "create", "test-sub")
		cmd.Dir = tmpDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("subproject create should fail outside project")
		}

		assert.Contains(t, string(output), "banago project not found")
	})
}

func TestIntegration_SubprojectList(t *testing.T) {
	t.Parallel()

	t.Run("lists subprojects", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "sub-one", ""))
		require.NoError(t, project.CreateSubproject(projectRoot, "sub-two", ""))

		cmd := exec.Command(testBinPath, "subproject", "list")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("subproject list failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "sub-one")
		assert.Contains(t, string(output), "sub-two")
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))

		cmd := exec.Command(testBinPath, "subproject", "list")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("subproject list failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "No subprojects")
	})
}

func TestIntegration_History(t *testing.T) {
	t.Parallel()

	t.Run("shows history entries", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
		subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")
		historyDir := filepath.Join(subprojectDir, "history")

		// Create some history entries (using local helper for read-only CLI tests)
		entry1 := createHistoryEntryForCLI(t, historyDir, "first prompt")
		entry2 := createHistoryEntryForCLI(t, historyDir, "second prompt")

		cmd := exec.Command(testBinPath, "history")
		cmd.Dir = subprojectDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("history failed: %v\noutput: %s", err, output)
		}

		// Verify entries are listed (at least partial IDs)
		assert.Contains(t, string(output), entry1.ID[:8])
		assert.Contains(t, string(output), entry2.ID[:8])
	})

	t.Run("empty history", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
		subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

		cmd := exec.Command(testBinPath, "history")
		cmd.Dir = subprojectDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("history failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "No history")
	})

	t.Run("fails outside subproject", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))

		cmd := exec.Command(testBinPath, "history")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("history should fail outside subproject")
		}

		assert.Contains(t, string(output), "subproject")
	})
}

func TestIntegration_Status(t *testing.T) {
	t.Parallel()

	t.Run("shows subproject status", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
		subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

		cmd := exec.Command(testBinPath, "status")
		cmd.Dir = subprojectDir
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("status failed: %v\noutput: %s", err, output)
		}

		// Should show subproject name and status info
		assert.Contains(t, string(output), "test-sub")
	})

	t.Run("shows project status from root", func(t *testing.T) {
		t.Parallel()

		projectRoot := t.TempDir()
		require.NoError(t, project.InitProject(projectRoot, "test-project", false))
		require.NoError(t, project.CreateSubproject(projectRoot, "sub-one", ""))

		cmd := exec.Command(testBinPath, "status")
		cmd.Dir = projectRoot
		cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("status failed: %v\noutput: %s", err, output)
		}

		assert.Contains(t, string(output), "test-project")
	})
}

func TestIntegration_ServeHelp(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinPath, "serve", "--help")
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("serve --help failed: %v\noutput: %s", err, output)
	}

	assert.Contains(t, string(output), "serve")
	assert.Contains(t, string(output), "port")
}

func TestIntegration_MigrateHelp(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinPath, "migrate", "--help")
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("migrate --help failed: %v\noutput: %s", err, output)
	}

	assert.Contains(t, string(output), "migrate")
}

func TestIntegration_EditRequiresAPIKey(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinPath, "edit", "--latest", "--prompt", "test")
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("edit command should fail without API key")
	}

	if !strings.Contains(string(output), "API key") {
		t.Errorf("expected error message about API key, got: %s", output)
	}
}

func TestIntegration_RegenerateRequiresAPIKey(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(testBinPath, "regenerate", "--latest")
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("regenerate command should fail without API key")
	}

	if !strings.Contains(string(output), "API key") {
		t.Errorf("expected error message about API key, got: %s", output)
	}
}
