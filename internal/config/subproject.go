package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// SubprojectConfig represents a subproject configuration (config.yaml)
type SubprojectConfig struct {
	Version       string   `yaml:"version"`
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description,omitempty"`
	CreatedAt     string   `yaml:"created_at"`
	CharacterFile string   `yaml:"character_file,omitempty"`
	ContextFile   string   `yaml:"context_file"`
	AspectRatio   string   `yaml:"aspect_ratio,omitempty"`
	ImageSize     string   `yaml:"image_size,omitempty"`
	InputImages   []string `yaml:"input_images,omitempty"`
}

const (
	subprojectConfigFile = "config.yaml"
	SubprojectsDir       = "subprojects"
	CharactersDir        = "characters"
	inputsDir            = "inputs"
	historyDir           = "history"
	DefaultContextFile   = "context.md"
)

// NewSubprojectConfig creates a new subproject configuration with defaults
func NewSubprojectConfig(name string) *SubprojectConfig {
	return &SubprojectConfig{
		Version:     configVersion,
		Name:        name,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		ContextFile: DefaultContextFile,
	}
}

// Save writes the subproject configuration to the specified directory
func (c *SubprojectConfig) Save(dir string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal subproject config: %w", err)
	}

	path := filepath.Join(dir, subprojectConfigFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write subproject config: %w", err)
	}

	return nil
}

// LoadSubprojectConfig reads a subproject configuration from the specified directory
func LoadSubprojectConfig(dir string) (*SubprojectConfig, error) {
	path := filepath.Join(dir, subprojectConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read subproject config: %w", err)
	}

	var config SubprojectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse subproject config: %w", err)
	}

	return &config, nil
}

// SubprojectConfigExists checks if a subproject configuration exists in the specified directory
func SubprojectConfigExists(dir string) bool {
	path := filepath.Join(dir, subprojectConfigFile)
	_, err := os.Stat(path)
	return err == nil
}

// GetSubprojectDir returns the path to a subproject directory
func GetSubprojectDir(projectRoot, name string) string {
	return filepath.Join(projectRoot, SubprojectsDir, name)
}

// GetInputsDir returns the path to the inputs directory of a subproject
func GetInputsDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, inputsDir)
}

// GetHistoryDir returns the path to the history directory of a subproject
func GetHistoryDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, historyDir)
}
