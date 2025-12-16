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
