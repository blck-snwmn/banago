package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/templates"
)

// InitProject initializes a new banago project in the specified directory
func InitProject(dir, name string, force bool) error {
	// Check if already initialized
	if config.ProjectConfigExists(dir) && !force {
		return ErrAlreadyInitialized
	}

	// Create project configuration
	cfg := config.NewProjectConfig(name)
	if err := cfg.Save(dir); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	// Create AI guide files
	if err := writeAIGuides(dir); err != nil {
		return fmt.Errorf("failed to write AI guides: %w", err)
	}

	// Create directories
	dirs := []string{
		filepath.Join(dir, charactersDir),
		filepath.Join(dir, subprojectsDir),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	return nil
}

func writeAIGuides(dir string) error {
	files := map[string]string{
		"CLAUDE.md": templates.ClaudeMD,
		"GEMINI.md": templates.GeminiMD,
		"AGENTS.md": templates.AgentsMD,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}
