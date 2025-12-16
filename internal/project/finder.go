package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/blck-snwmn/banago/internal/config"
)

var (
	ErrProjectNotFound    = errors.New("banago project not found (no banago.yaml in current or parent directories)")
	ErrNotInSubproject    = errors.New("not inside a subproject directory")
	ErrAlreadyInitialized = errors.New("banago project already initialized in this directory")
)

// FindProjectRoot searches for the project root directory by looking for banago.yaml
// starting from startDir and traversing up to parent directories
func FindProjectRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if config.ProjectConfigExists(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", ErrProjectNotFound
		}
		dir = parent
	}
}

// FindCurrentSubproject determines if the current directory is inside a subproject
// Returns the subproject name if found
func FindCurrentSubproject(projectRoot, cwd string) (string, error) {
	rel, err := filepath.Rel(projectRoot, cwd)
	if err != nil {
		return "", err
	}

	// Check if we're inside subprojects/<name>/...
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) >= 2 && parts[0] == subprojectsDir {
		subprojectName := parts[1]
		subprojectDir := GetSubprojectDir(projectRoot, subprojectName)
		if config.SubprojectConfigExists(subprojectDir) {
			return subprojectName, nil
		}
	}

	return "", ErrNotInSubproject
}

// listSubprojects returns a list of all subproject names in the project
func listSubprojects(projectRoot string) ([]string, error) {
	spDir := filepath.Join(projectRoot, subprojectsDir)

	entries, err := os.ReadDir(spDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			subprojectDir := filepath.Join(spDir, entry.Name())
			if config.SubprojectConfigExists(subprojectDir) {
				names = append(names, entry.Name())
			}
		}
	}

	return names, nil
}
