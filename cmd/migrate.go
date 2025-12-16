package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate history entries to new format",
	Long: `Migrate history entries from old format to new format.

This command:
  - Copies input images from inputs/ to each history entry directory
  - Removes context.md and character.md from history entries
  - Updates the project version to 2

The migration is idempotent - running it multiple times is safe.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return errors.New("banago project not found. Run 'banago init' first")
			}
			return err
		}

		// Load project config
		projectCfg, err := config.LoadProjectConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}

		// Check version
		version, err := parseVersion(projectCfg.Version)
		if err != nil {
			return fmt.Errorf("failed to parse project version: %w", err)
		}
		if version >= 2 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Already migrated (version >= 2)")
			return nil
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintln(w, "Starting migration...")
		_, _ = fmt.Fprintln(w, "")

		// List all subprojects
		subprojectInfos, err := project.ListSubprojectInfos(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to list subprojects: %w", err)
		}

		var failedPaths []string
		var migratedCount int
		var skippedSubprojects int
		allSubprojectsSuccess := true

		for _, info := range subprojectInfos {
			subprojectName := info.Name
			subprojectDir := config.GetSubprojectDir(projectRoot, subprojectName)

			// Load subproject config and check version
			subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
			if err != nil {
				failedPaths = append(failedPaths, fmt.Sprintf("%s/config.yaml: %v", subprojectDir, err))
				allSubprojectsSuccess = false
				continue
			}

			subVersion, err := parseVersion(subprojectCfg.Version)
			if err != nil {
				failedPaths = append(failedPaths, fmt.Sprintf("%s: invalid version %q", subprojectDir, subprojectCfg.Version))
				allSubprojectsSuccess = false
				continue
			}

			// Skip if subproject already migrated
			if subVersion >= 2 {
				skippedSubprojects++
				continue
			}

			historyDir := config.GetHistoryDir(subprojectDir)
			inputsDir := config.GetInputsDir(subprojectDir)

			entries, err := history.ListEntries(historyDir)
			if err != nil {
				failedPaths = append(failedPaths, fmt.Sprintf("%s: %v", historyDir, err))
				allSubprojectsSuccess = false
				continue
			}

			subprojectFailed := false

			for _, entry := range entries {
				entryDir := entry.GetEntryDir(historyDir)

				// Copy input images from inputs/ to history entry
				for _, imgName := range entry.Generation.InputImages {
					srcPath := filepath.Join(inputsDir, imgName)
					dstPath := filepath.Join(entryDir, imgName)

					// Skip if already exists
					if _, err := os.Stat(dstPath); err == nil {
						continue
					}

					// Copy file
					data, err := os.ReadFile(srcPath)
					if err != nil {
						failedPaths = append(failedPaths, fmt.Sprintf("%s -> %s: %v", srcPath, dstPath, err))
						subprojectFailed = true
						continue
					}
					if err := os.WriteFile(dstPath, data, 0o644); err != nil {
						failedPaths = append(failedPaths, fmt.Sprintf("%s: %v", dstPath, err))
						subprojectFailed = true
						continue
					}
				}

				// Remove context.md if exists
				contextPath := filepath.Join(entryDir, "context.md")
				if _, err := os.Stat(contextPath); err == nil {
					if err := os.Remove(contextPath); err != nil {
						failedPaths = append(failedPaths, fmt.Sprintf("%s: %v", contextPath, err))
						subprojectFailed = true
					}
				}

				// Remove character.md if exists
				characterPath := filepath.Join(entryDir, "character.md")
				if _, err := os.Stat(characterPath); err == nil {
					if err := os.Remove(characterPath); err != nil {
						failedPaths = append(failedPaths, fmt.Sprintf("%s: %v", characterPath, err))
						subprojectFailed = true
					}
				}

				migratedCount++
			}

			// Update subproject version if successful
			if !subprojectFailed {
				subprojectCfg.Version = "2"
				if err := subprojectCfg.Save(subprojectDir); err != nil {
					failedPaths = append(failedPaths, fmt.Sprintf("%s/config.yaml: %v", subprojectDir, err))
					allSubprojectsSuccess = false
				}
			} else {
				allSubprojectsSuccess = false
			}
		}

		// Update project version only if all subprojects succeeded
		if allSubprojectsSuccess {
			projectCfg.Version = "2"
			if err := projectCfg.Save(projectRoot); err != nil {
				return fmt.Errorf("failed to update project version: %w", err)
			}
		}

		// Print summary
		_, _ = fmt.Fprintf(w, "Migration completed: %d entries migrated", migratedCount)
		if skippedSubprojects > 0 {
			_, _ = fmt.Fprintf(w, " (%d subprojects skipped)", skippedSubprojects)
		}
		_, _ = fmt.Fprintln(w, "")

		if len(failedPaths) > 0 {
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "Failed paths:")
			for _, p := range failedPaths {
				_, _ = fmt.Fprintf(w, "  - %s\n", p)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

// parseVersion parses version string and returns major version number
func parseVersion(version string) (int, error) {
	// Handle versions like "1.0", "2", "2.0"
	if version == "" {
		return 0, errors.New("empty version")
	}

	// Extract major version (before first dot)
	major := version
	for i, c := range version {
		if c == '.' {
			major = version[:i]
			break
		}
	}

	return strconv.Atoi(major)
}
