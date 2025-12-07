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
	Short: "新しい banago プロジェクトを初期化する",
	Long: `カレントディレクトリに banago プロジェクトを初期化します。

以下のファイル・ディレクトリが作成されます:
  - banago.yaml (プロジェクト設定)
  - CLAUDE.md (Claude Code 向けガイド)
  - GEMINI.md (Gemini CLI 向けガイド)
  - AGENTS.md (共通AIエージェントガイド)
  - characters/ (キャラクター定義ディレクトリ)
  - subprojects/ (サブプロジェクトディレクトリ)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
		}

		name := initOpts.name
		if name == "" {
			name = filepath.Base(cwd)
		}

		if err := project.InitProject(cwd, name, initOpts.force); err != nil {
			if errors.Is(err, project.ErrAlreadyInitialized) {
				return fmt.Errorf("このディレクトリには既に banago プロジェクトが存在します。上書きする場合は --force を指定してください")
			}
			return fmt.Errorf("プロジェクトの初期化に失敗しました: %w", err)
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "banago プロジェクト '%s' を初期化しました\n", name)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "作成されたファイル:")
		_, _ = fmt.Fprintln(w, "  banago.yaml")
		_, _ = fmt.Fprintln(w, "  CLAUDE.md")
		_, _ = fmt.Fprintln(w, "  GEMINI.md")
		_, _ = fmt.Fprintln(w, "  AGENTS.md")
		_, _ = fmt.Fprintln(w, "  characters/")
		_, _ = fmt.Fprintln(w, "  subprojects/")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "次のステップ:")
		_, _ = fmt.Fprintln(w, "  1. characters/ にキャラクター定義ファイルを作成")
		_, _ = fmt.Fprintln(w, "  2. banago subproject create <name> でサブプロジェクトを作成")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initOpts.name, "name", "", "プロジェクト名 (デフォルト: ディレクトリ名)")
	initCmd.Flags().BoolVar(&initOpts.force, "force", false, "既存のプロジェクトを上書きする")
}
