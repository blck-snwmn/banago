package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProjectConfig(t *testing.T) {
	t.Parallel()

	cfg := NewProjectConfig("test-project")

	if cfg.Name != "test-project" {
		t.Errorf("Name = %q, want %q", cfg.Name, "test-project")
	}
	if cfg.Version != configVersion {
		t.Errorf("Version = %q, want %q", cfg.Version, configVersion)
	}
	if cfg.Model != defaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, defaultModel)
	}
	if cfg.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
}

func TestProjectConfig_SaveAndLoad(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := NewProjectConfig("test-project")

	// Save
	if err := cfg.Save(tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadProjectConfig() error = %v", err)
	}

	if loaded.Name != cfg.Name {
		t.Errorf("loaded.Name = %q, want %q", loaded.Name, cfg.Name)
	}
	if loaded.Version != cfg.Version {
		t.Errorf("loaded.Version = %q, want %q", loaded.Version, cfg.Version)
	}
	if loaded.Model != cfg.Model {
		t.Errorf("loaded.Model = %q, want %q", loaded.Model, cfg.Model)
	}
}

func TestLoadProjectConfig_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	_, err := LoadProjectConfig(tmpDir)
	if err == nil {
		t.Error("LoadProjectConfig() expected error for missing file")
	}
}

func TestProjectConfigExists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		cfg := NewProjectConfig("test")
		if err := cfg.Save(tmpDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if !ProjectConfigExists(tmpDir) {
			t.Error("ProjectConfigExists() = false, want true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		if ProjectConfigExists(tmpDir) {
			t.Error("ProjectConfigExists() = true, want false")
		}
	})
}

func TestNewSubprojectConfig(t *testing.T) {
	t.Parallel()

	cfg := NewSubprojectConfig("test-subproject")

	if cfg.Name != "test-subproject" {
		t.Errorf("Name = %q, want %q", cfg.Name, "test-subproject")
	}
	if cfg.Version != configVersion {
		t.Errorf("Version = %q, want %q", cfg.Version, configVersion)
	}
	if cfg.ContextFile != DefaultContextFile {
		t.Errorf("ContextFile = %q, want %q", cfg.ContextFile, DefaultContextFile)
	}
	if cfg.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
}

func TestSubprojectConfig_SaveAndLoad(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := NewSubprojectConfig("test-subproject")
	cfg.Description = "Test description"
	cfg.CharacterFile = "char.md"
	cfg.InputImages = []string{"img1.png", "img2.jpg"}

	// Save
	if err := cfg.Save(tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := LoadSubprojectConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadSubprojectConfig() error = %v", err)
	}

	if loaded.Name != cfg.Name {
		t.Errorf("loaded.Name = %q, want %q", loaded.Name, cfg.Name)
	}
	if loaded.Description != cfg.Description {
		t.Errorf("loaded.Description = %q, want %q", loaded.Description, cfg.Description)
	}
	if loaded.CharacterFile != cfg.CharacterFile {
		t.Errorf("loaded.CharacterFile = %q, want %q", loaded.CharacterFile, cfg.CharacterFile)
	}
	if len(loaded.InputImages) != len(cfg.InputImages) {
		t.Errorf("loaded.InputImages length = %d, want %d", len(loaded.InputImages), len(cfg.InputImages))
	}
}

func TestSubprojectConfigExists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		cfg := NewSubprojectConfig("test")
		if err := cfg.Save(tmpDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if !SubprojectConfigExists(tmpDir) {
			t.Error("SubprojectConfigExists() = false, want true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		if SubprojectConfigExists(tmpDir) {
			t.Error("SubprojectConfigExists() = true, want false")
		}
	})
}

func TestLoadSubprojectConfig_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	_, err := LoadSubprojectConfig(tmpDir)
	if err == nil {
		t.Error("LoadSubprojectConfig() expected error for missing file")
	}
}

func TestLoadProjectConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "banago.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0o644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	_, err := LoadProjectConfig(tmpDir)
	if err == nil {
		t.Error("LoadProjectConfig() expected error for invalid YAML")
	}
}
