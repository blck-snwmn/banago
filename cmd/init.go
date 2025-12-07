package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var initOpts struct {
	name  string
	force bool
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new banago project",
	Long: `Initialize a banago project in the current directory.

The following files and directories will be created:
  - banago.yaml (project config)
  - CLAUDE.md (Claude Code guide)
  - GEMINI.md (Gemini CLI guide)
  - AGENTS.md (common AI agent guide)
  - characters/ (character definitions directory)
  - subprojects/ (subprojects directory)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		name := initOpts.name
		if name == "" {
			name = filepath.Base(cwd)
		}

		if err := project.InitProject(cwd, name, initOpts.force); err != nil {
			if errors.Is(err, project.ErrAlreadyInitialized) {
				return fmt.Errorf("banago project already exists in this directory. Use --force to overwrite")
			}
			return fmt.Errorf("failed to initialize project: %w", err)
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Initialized banago project '%s'\n", name)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Created files:")
		_, _ = fmt.Fprintln(w, "  banago.yaml")
		_, _ = fmt.Fprintln(w, "  CLAUDE.md")
		_, _ = fmt.Fprintln(w, "  GEMINI.md")
		_, _ = fmt.Fprintln(w, "  AGENTS.md")
		_, _ = fmt.Fprintln(w, "  characters/")
		_, _ = fmt.Fprintln(w, "  subprojects/")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Next steps:")
		_, _ = fmt.Fprintln(w, "  1. Create character definition files in characters/")
		_, _ = fmt.Fprintln(w, "  2. Run 'banago subproject create <name>' to create a subproject")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initOpts.name, "name", "", "Project name (default: directory name)")
	initCmd.Flags().BoolVar(&initOpts.force, "force", false, "Overwrite existing project")
}
