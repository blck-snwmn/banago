package project

import "path/filepath"

const (
	subprojectsDir = "subprojects"
	charactersDir  = "characters"
	inputsDirName  = "inputs"
)

// GetSubprojectsDir returns the path to the subprojects directory
func GetSubprojectsDir(projectRoot string) string {
	return filepath.Join(projectRoot, subprojectsDir)
}

// GetSubprojectDir returns the path to a subproject directory
func GetSubprojectDir(projectRoot, name string) string {
	return filepath.Join(projectRoot, subprojectsDir, name)
}

// GetCharactersDir returns the path to the characters directory
func GetCharactersDir(projectRoot string) string {
	return filepath.Join(projectRoot, charactersDir)
}

// GetCharacterPath returns the path to a character file
func GetCharacterPath(projectRoot, characterFile string) string {
	return filepath.Join(projectRoot, charactersDir, characterFile)
}

// GetInputsDir returns the path to the inputs directory of a subproject
func GetInputsDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, inputsDirName)
}
