package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the root project configuration (banago.yaml)
type ProjectConfig struct {
	Version   string          `yaml:"version"`
	Name      string          `yaml:"name"`
	Model     string          `yaml:"model"`
	CreatedAt string          `yaml:"created_at"`
	Defaults  ProjectDefaults `yaml:"defaults,omitempty"`
}

// ProjectDefaults contains default generation parameters
type ProjectDefaults struct {
	AspectRatio string `yaml:"aspect_ratio,omitempty"`
	ImageSize   string `yaml:"image_size,omitempty"`
}

const (
	ProjectConfigFile = "banago.yaml"
	DefaultModel      = "gemini-3-pro-image-preview"
	ConfigVersion     = "1.0"
)

// NewProjectConfig creates a new project configuration with defaults
func NewProjectConfig(name string) *ProjectConfig {
	return &ProjectConfig{
		Version:   ConfigVersion,
		Name:      name,
		Model:     DefaultModel,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// Save writes the project configuration to the specified directory
func (c *ProjectConfig) Save(dir string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	path := filepath.Join(dir, ProjectConfigFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

// LoadProjectConfig reads a project configuration from the specified directory
func LoadProjectConfig(dir string) (*ProjectConfig, error) {
	path := filepath.Join(dir, ProjectConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	return &config, nil
}

// Exists checks if a project configuration exists in the specified directory
func ProjectConfigExists(dir string) bool {
	path := filepath.Join(dir, ProjectConfigFile)
	_, err := os.Stat(path)
	return err == nil
}
