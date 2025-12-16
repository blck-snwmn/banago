package history

import "path/filepath"

const historyDirName = "history"

// GetHistoryDir returns the path to the history directory of a subproject
func GetHistoryDir(subprojectDir string) string {
	return filepath.Join(subprojectDir, historyDirName)
}
