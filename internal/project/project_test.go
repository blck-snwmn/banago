package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/config"
)

func setupTestProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()

	// Create project config
	cfg := config.NewProjectConfig("test-project")
	if err := cfg.Save(projectRoot); err != nil {
		t.Fatalf("failed to save project config: %v", err)
	}

	return projectRoot
}

func setupTestSubproject(t *testing.T, projectRoot, name string) string {
	t.Helper()

	subprojectDir := GetSubprojectDir(projectRoot, name)
	if err := os.MkdirAll(subprojectDir, 0o755); err != nil {
		t.Fatalf("failed to create subproject dir: %v", err)
	}

	subCfg := config.NewSubprojectConfig(name)
	if err := subCfg.Save(subprojectDir); err != nil {
		t.Fatalf("failed to save subproject config: %v", err)
	}

	return subprojectDir
}

func TestFindProjectRoot(t *testing.T) {
	t.Parallel()

	t.Run("finds project in current dir", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		found, err := FindProjectRoot(projectRoot)
		if err != nil {
			t.Fatalf("FindProjectRoot() error = %v", err)
		}
		if found != projectRoot {
			t.Errorf("FindProjectRoot() = %q, want %q", found, projectRoot)
		}
	})

	t.Run("finds project in parent dir", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		// Create nested directory
		nestedDir := filepath.Join(projectRoot, "some", "nested", "dir")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		found, err := FindProjectRoot(nestedDir)
		if err != nil {
			t.Fatalf("FindProjectRoot() error = %v", err)
		}
		if found != projectRoot {
			t.Errorf("FindProjectRoot() = %q, want %q", found, projectRoot)
		}
	})

	t.Run("returns error when not found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		_, err := FindProjectRoot(tmpDir)
		if !errors.Is(err, ErrProjectNotFound) {
			t.Errorf("FindProjectRoot() error = %v, want %v", err, ErrProjectNotFound)
		}
	})
}

func TestFindCurrentSubproject(t *testing.T) {
	t.Parallel()

	t.Run("finds subproject", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)
		subprojectDir := setupTestSubproject(t, projectRoot, "my-subproject")

		name, err := FindCurrentSubproject(projectRoot, subprojectDir)
		if err != nil {
			t.Fatalf("FindCurrentSubproject() error = %v", err)
		}
		if name != "my-subproject" {
			t.Errorf("FindCurrentSubproject() = %q, want %q", name, "my-subproject")
		}
	})

	t.Run("finds subproject from nested dir", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)
		subprojectDir := setupTestSubproject(t, projectRoot, "my-subproject")

		// Create nested directory inside subproject
		nestedDir := filepath.Join(subprojectDir, "some", "nested")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		name, err := FindCurrentSubproject(projectRoot, nestedDir)
		if err != nil {
			t.Fatalf("FindCurrentSubproject() error = %v", err)
		}
		if name != "my-subproject" {
			t.Errorf("FindCurrentSubproject() = %q, want %q", name, "my-subproject")
		}
	})

	t.Run("returns error at project root", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		_, err := FindCurrentSubproject(projectRoot, projectRoot)
		if !errors.Is(err, ErrNotInSubproject) {
			t.Errorf("FindCurrentSubproject() error = %v, want %v", err, ErrNotInSubproject)
		}
	})

	t.Run("returns error for non-subproject dir", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		// Create a dir that is not a subproject
		otherDir := filepath.Join(projectRoot, "other")
		if err := os.MkdirAll(otherDir, 0o755); err != nil {
			t.Fatalf("failed to create other dir: %v", err)
		}

		_, err := FindCurrentSubproject(projectRoot, otherDir)
		if !errors.Is(err, ErrNotInSubproject) {
			t.Errorf("FindCurrentSubproject() error = %v, want %v", err, ErrNotInSubproject)
		}
	})

	t.Run("returns error for dir without config", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		// Create subprojects dir without config
		subprojectDir := filepath.Join(projectRoot, subprojectsDir, "no-config")
		if err := os.MkdirAll(subprojectDir, 0o755); err != nil {
			t.Fatalf("failed to create subproject dir: %v", err)
		}

		_, err := FindCurrentSubproject(projectRoot, subprojectDir)
		if !errors.Is(err, ErrNotInSubproject) {
			t.Errorf("FindCurrentSubproject() error = %v, want %v", err, ErrNotInSubproject)
		}
	})
}

func TestListSubprojects(t *testing.T) {
	t.Parallel()

	t.Run("lists subprojects", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)
		setupTestSubproject(t, projectRoot, "sub1")
		setupTestSubproject(t, projectRoot, "sub2")
		setupTestSubproject(t, projectRoot, "sub3")

		names, err := listSubprojects(projectRoot)
		if err != nil {
			t.Fatalf("listSubprojects() error = %v", err)
		}
		if len(names) != 3 {
			t.Errorf("listSubprojects() returned %d items, want 3", len(names))
		}
	})

	t.Run("returns empty for no subprojects", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		names, err := listSubprojects(projectRoot)
		if err != nil {
			t.Fatalf("listSubprojects() error = %v", err)
		}
		if len(names) != 0 {
			t.Errorf("listSubprojects() returned %d items, want 0", len(names))
		}
	})

	t.Run("ignores dirs without config", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)
		setupTestSubproject(t, projectRoot, "valid")

		// Create dir without config
		invalidDir := filepath.Join(projectRoot, subprojectsDir, "invalid")
		if err := os.MkdirAll(invalidDir, 0o755); err != nil {
			t.Fatalf("failed to create invalid dir: %v", err)
		}

		names, err := listSubprojects(projectRoot)
		if err != nil {
			t.Fatalf("listSubprojects() error = %v", err)
		}
		if len(names) != 1 {
			t.Errorf("listSubprojects() returned %d items, want 1", len(names))
		}
		if names[0] != "valid" {
			t.Errorf("listSubprojects()[0] = %q, want %q", names[0], "valid")
		}
	})
}

func TestInitProject(t *testing.T) {
	t.Parallel()

	t.Run("initializes project", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := InitProject(dir, "test-project", false)
		if err != nil {
			t.Fatalf("InitProject() error = %v", err)
		}

		// Verify project config exists
		if !config.ProjectConfigExists(dir) {
			t.Error("project config should exist")
		}

		// Verify directories exist
		if _, err := os.Stat(filepath.Join(dir, charactersDir)); os.IsNotExist(err) {
			t.Error("characters directory should exist")
		}
		if _, err := os.Stat(filepath.Join(dir, subprojectsDir)); os.IsNotExist(err) {
			t.Error("subprojects directory should exist")
		}

		// Verify AI guides exist
		for _, filename := range []string{"CLAUDE.md", "GEMINI.md", "AGENTS.md"} {
			if _, err := os.Stat(filepath.Join(dir, filename)); os.IsNotExist(err) {
				t.Errorf("%s should exist", filename)
			}
		}
	})

	t.Run("fails if already initialized", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Initialize first time
		if err := InitProject(dir, "test-project", false); err != nil {
			t.Fatalf("first InitProject() error = %v", err)
		}

		// Try to initialize again
		err := InitProject(dir, "test-project", false)
		if !errors.Is(err, ErrAlreadyInitialized) {
			t.Errorf("InitProject() error = %v, want %v", err, ErrAlreadyInitialized)
		}
	})

	t.Run("force overwrites existing", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Initialize first time
		if err := InitProject(dir, "old-name", false); err != nil {
			t.Fatalf("first InitProject() error = %v", err)
		}

		// Initialize again with force
		err := InitProject(dir, "new-name", true)
		if err != nil {
			t.Fatalf("InitProject() with force error = %v", err)
		}

		// Verify new name
		cfg, err := config.LoadProjectConfig(dir)
		if err != nil {
			t.Fatalf("LoadProjectConfig() error = %v", err)
		}
		if cfg.Name != "new-name" {
			t.Errorf("config.Name = %q, want %q", cfg.Name, "new-name")
		}
	})
}

func TestCreateSubproject(t *testing.T) {
	t.Parallel()

	t.Run("creates subproject", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		err := CreateSubproject(projectRoot, "my-sub", "A test subproject")
		if err != nil {
			t.Fatalf("CreateSubproject() error = %v", err)
		}

		subprojectDir := GetSubprojectDir(projectRoot, "my-sub")

		// Verify config exists
		if !config.SubprojectConfigExists(subprojectDir) {
			t.Error("subproject config should exist")
		}

		// Verify directories exist
		if _, err := os.Stat(GetInputsDir(subprojectDir)); os.IsNotExist(err) {
			t.Error("inputs directory should exist")
		}
		if _, err := os.Stat(filepath.Join(subprojectDir, "history")); os.IsNotExist(err) {
			t.Error("history directory should exist")
		}

		// Verify context.md exists
		if _, err := os.Stat(filepath.Join(subprojectDir, "context.md")); os.IsNotExist(err) {
			t.Error("context.md should exist")
		}

		// Verify description is set
		cfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			t.Fatalf("LoadSubprojectConfig() error = %v", err)
		}
		if cfg.Description != "A test subproject" {
			t.Errorf("config.Description = %q, want %q", cfg.Description, "A test subproject")
		}
	})

	t.Run("fails if subproject exists", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		// Create first time
		if err := CreateSubproject(projectRoot, "existing", ""); err != nil {
			t.Fatalf("first CreateSubproject() error = %v", err)
		}

		// Try to create again
		err := CreateSubproject(projectRoot, "existing", "")
		if err == nil {
			t.Error("CreateSubproject() should fail for existing subproject")
		}
	})
}

func TestListSubprojectInfos(t *testing.T) {
	t.Parallel()

	t.Run("returns subproject infos", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		// Create subprojects with descriptions
		if err := CreateSubproject(projectRoot, "sub1", "First subproject"); err != nil {
			t.Fatalf("CreateSubproject() error = %v", err)
		}
		if err := CreateSubproject(projectRoot, "sub2", "Second subproject"); err != nil {
			t.Fatalf("CreateSubproject() error = %v", err)
		}

		infos, err := ListSubprojectInfos(projectRoot)
		if err != nil {
			t.Fatalf("ListSubprojectInfos() error = %v", err)
		}
		if len(infos) != 2 {
			t.Fatalf("ListSubprojectInfos() returned %d items, want 2", len(infos))
		}

		// Check that descriptions are included
		foundSub1 := false
		for _, info := range infos {
			if info.Name == "sub1" && info.Description == "First subproject" {
				foundSub1 = true
			}
		}
		if !foundSub1 {
			t.Error("sub1 with description not found in infos")
		}
	})

	t.Run("returns empty for no subprojects", func(t *testing.T) {
		t.Parallel()
		projectRoot := setupTestProject(t)

		infos, err := ListSubprojectInfos(projectRoot)
		if err != nil {
			t.Fatalf("ListSubprojectInfos() error = %v", err)
		}
		if len(infos) != 0 {
			t.Errorf("ListSubprojectInfos() returned %d items, want 0", len(infos))
		}
	})
}

// Path function tests

func TestGetSubprojectsDir(t *testing.T) {
	t.Parallel()

	projectRoot := "/path/to/project"
	expected := "/path/to/project/subprojects"
	got := GetSubprojectsDir(projectRoot)

	if got != expected {
		t.Errorf("GetSubprojectsDir() = %q, want %q", got, expected)
	}
}

func TestGetSubprojectDir(t *testing.T) {
	t.Parallel()

	projectRoot := "/path/to/project"
	name := "my-sub"
	expected := "/path/to/project/subprojects/my-sub"
	got := GetSubprojectDir(projectRoot, name)

	if got != expected {
		t.Errorf("GetSubprojectDir() = %q, want %q", got, expected)
	}
}

func TestGetCharactersDir(t *testing.T) {
	t.Parallel()

	projectRoot := "/path/to/project"
	expected := "/path/to/project/characters"
	got := GetCharactersDir(projectRoot)

	if got != expected {
		t.Errorf("GetCharactersDir() = %q, want %q", got, expected)
	}
}

func TestGetCharacterPath(t *testing.T) {
	t.Parallel()

	projectRoot := "/path/to/project"
	characterFile := "character.md"
	expected := "/path/to/project/characters/character.md"
	got := GetCharacterPath(projectRoot, characterFile)

	if got != expected {
		t.Errorf("GetCharacterPath() = %q, want %q", got, expected)
	}
}

func TestGetInputsDir(t *testing.T) {
	t.Parallel()

	subprojectDir := "/path/to/project/subprojects/my-sub"
	expected := "/path/to/project/subprojects/my-sub/inputs"
	got := GetInputsDir(subprojectDir)

	if got != expected {
		t.Errorf("GetInputsDir() = %q, want %q", got, expected)
	}
}
