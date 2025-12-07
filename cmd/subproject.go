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
	Short: "サブプロジェクトを管理する",
	Long:  "サブプロジェクトの作成・一覧表示などを行います。",
}

var subprojectCreateOpts struct {
	description string
}

var subprojectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "新しいサブプロジェクトを作成する",
	Long: `指定した名前で新しいサブプロジェクトを作成します。

以下のファイル・ディレクトリが作成されます:
  - subprojects/<name>/config.yaml (サブプロジェクト設定)
  - subprojects/<name>/context.md (付加情報ファイル)
  - subprojects/<name>/inputs/ (入力画像ディレクトリ)
  - subprojects/<name>/history/ (生成履歴ディレクトリ)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return fmt.Errorf("banago プロジェクトが見つかりません。先に banago init を実行してください")
			}
			return err
		}

		if err := project.CreateSubproject(projectRoot, name, subprojectCreateOpts.description); err != nil {
			return fmt.Errorf("サブプロジェクトの作成に失敗しました: %w", err)
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "サブプロジェクト '%s' を作成しました\n", name)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "次のステップ:")
		_, _ = fmt.Fprintf(w, "  1. subprojects/%s/config.yaml でキャラクター参照を設定\n", name)
		_, _ = fmt.Fprintf(w, "  2. subprojects/%s/context.md に付加情報を記載\n", name)
		_, _ = fmt.Fprintf(w, "  3. subprojects/%s/inputs/ に参照画像を配置\n", name)

		return nil
	},
}

var subprojectListCmd = &cobra.Command{
	Use:   "list",
	Short: "サブプロジェクト一覧を表示する",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		projectRoot, err := project.FindProjectRoot(cwd)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				return fmt.Errorf("banago プロジェクトが見つかりません。先に banago init を実行してください")
			}
			return err
		}

		infos, err := project.ListSubprojectInfos(projectRoot)
		if err != nil {
			return fmt.Errorf("サブプロジェクト一覧の取得に失敗しました: %w", err)
		}

		w := cmd.OutOrStdout()
		if len(infos) == 0 {
			_, _ = fmt.Fprintln(w, "サブプロジェクトがありません")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "新しいサブプロジェクトを作成するには:")
			_, _ = fmt.Fprintln(w, "  banago subproject create <name>")
			return nil
		}

		_, _ = fmt.Fprintln(w, "サブプロジェクト一覧:")
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

	subprojectCreateCmd.Flags().StringVar(&subprojectCreateOpts.description, "description", "", "サブプロジェクトの説明")
}
