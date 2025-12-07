package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var subprojectCmd = &cobra.Command{
	Use:   "subproject",
	Short: "Manage subprojects",
	Long:  "Create and list subprojects.",
}

var subprojectCreateOpts struct {
	description string
}

var subprojectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new subproject",
	Long: `Create a new subproject with the specified name.

The following files and directories will be created:
  - subprojects/<name>/config.yaml (subproject config)
  - subprojects/<name>/context.md (additional info file)
  - subprojects/<name>/inputs/ (input images directory)
  - subprojects/<name>/history/ (generation history directory)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return fmt.Errorf("banago project not found. Run 'banago init' first")
			}
			return err
		}

		if err := project.CreateSubproject(projectRoot, name, subprojectCreateOpts.description); err != nil {
			return fmt.Errorf("failed to create subproject: %w", err)
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "Created subproject '%s'\n", name)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Next steps:")
		_, _ = fmt.Fprintf(w, "  1. Configure character reference in subprojects/%s/config.yaml\n", name)
		_, _ = fmt.Fprintf(w, "  2. Add context info to subprojects/%s/context.md\n", name)
		_, _ = fmt.Fprintf(w, "  3. Place reference images in subprojects/%s/inputs/\n", name)

		return nil
	},
}

var subprojectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subprojects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return fmt.Errorf("banago project not found. Run 'banago init' first")
			}
			return err
		}

		infos, err := project.ListSubprojectInfos(projectRoot)
		if err != nil {
			return fmt.Errorf("failed to list subprojects: %w", err)
		}

		w := cmd.OutOrStdout()
		if len(infos) == 0 {
			_, _ = fmt.Fprintln(w, "No subprojects found")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "To create a new subproject:")
			_, _ = fmt.Fprintln(w, "  banago subproject create <name>")
			return nil
		}

		_, _ = fmt.Fprintln(w, "Subprojects:")
		for _, info := range infos {
			_, _ = fmt.Fprintf(w, "  %s", info.Name)
			if info.Description != "" {
				_, _ = fmt.Fprintf(w, " - %s", info.Description)
			}
			_, _ = fmt.Fprintln(w)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(subprojectCmd)
	subprojectCmd.AddCommand(subprojectCreateCmd)
	subprojectCmd.AddCommand(subprojectListCmd)

	subprojectCreateCmd.Flags().StringVar(&subprojectCreateOpts.description, "description", "", "Subproject description")
}
