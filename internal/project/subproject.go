package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/templates"
)

// CreateSubproject creates a new subproject in the specified project
func CreateSubproject(projectRoot, name, description string) error {
	subprojectDir := GetSubprojectDir(projectRoot, name)

	// Check if subproject already exists
	if config.SubprojectConfigExists(subprojectDir) {
		return fmt.Errorf("subproject '%s' already exists", name)
	}

	// Create subproject directory structure
	dirs := []string{
		subprojectDir,
		GetInputsDir(subprojectDir),
		history.GetHistoryDir(subprojectDir),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Create subproject configuration
	cfg := config.NewSubprojectConfig(name)
	cfg.Description = description
	if err := cfg.Save(subprojectDir); err != nil {
		return fmt.Errorf("failed to save subproject config: %w", err)
	}

	// Create default context.md
	contextPath := filepath.Join(subprojectDir, config.DefaultContextFile)
	if err := os.WriteFile(contextPath, []byte(templates.DefaultContextMD), 0o644); err != nil {
		return fmt.Errorf("failed to write context.md: %w", err)
	}

	return nil
}

// SubprojectInfo contains information about a subproject
type SubprojectInfo struct {
	Name        string
	Description string
}

// getSubprojectInfo returns information about a specific subproject
func getSubprojectInfo(projectRoot, name string) (*SubprojectInfo, error) {
	subprojectDir := GetSubprojectDir(projectRoot, name)

	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	if err != nil {
		return nil, err
	}

	return &SubprojectInfo{
		Name:        cfg.Name,
		Description: cfg.Description,
	}, nil
}

// ListSubprojectInfos returns information about all subprojects
func ListSubprojectInfos(projectRoot string) ([]*SubprojectInfo, error) {
	names, err := listSubprojects(projectRoot)
	if err != nil {
		return nil, err
	}

	var infos []*SubprojectInfo
	for _, name := range names {
		info, err := getSubprojectInfo(projectRoot, name)
		if err != nil {
			continue // Skip invalid subprojects
		}
		infos = append(infos, info)
	}

	return infos, nil
}
