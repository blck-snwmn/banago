package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var historyOpts struct {
	limit int
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "生成履歴を表示する",
	Long:  "現在のサブプロジェクトの生成履歴を表示します。",
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

		subprojectName, err := project.FindCurrentSubproject(projectRoot, cwd)
		if err != nil {
			if errors.Is(err, project.ErrNotInSubproject) {
				return fmt.Errorf("サブプロジェクト内にいません。サブプロジェクトディレクトリに移動してください")
			}
			return err
		}

		subprojectDir := config.GetSubprojectDir(projectRoot, subprojectName)
		historyDir := config.GetHistoryDir(subprojectDir)

		entries, err := history.ListEntries(historyDir)
		if err != nil {
			return fmt.Errorf("履歴の読み込みに失敗しました: %w", err)
		}

		w := cmd.OutOrStdout()

		if len(entries) == 0 {
			_, _ = fmt.Fprintln(w, "履歴がありません")
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "画像を生成するには:")
			_, _ = fmt.Fprintln(w, "  banago generate --prompt \"...\"")
			return nil
		}

		_, _ = fmt.Fprintf(w, "履歴 (%d 件):\n", len(entries))
		_, _ = fmt.Fprintln(w, "")

		// Show entries in reverse order (newest first)
		start := 0
		if historyOpts.limit > 0 && historyOpts.limit < len(entries) {
			start = len(entries) - historyOpts.limit
		}

		for i := len(entries) - 1; i >= start; i-- {
			entry := entries[i]
			status := "✓"
			if !entry.Result.Success {
				status = "✗"
			}
			_, _ = fmt.Fprintf(w, "  %s %s\n", status, entry.ID)
			_, _ = fmt.Fprintf(w, "      日時: %s\n", entry.CreatedAt)
			if entry.Result.Success && len(entry.Result.OutputImages) > 0 {
				_, _ = fmt.Fprintf(w, "      出力: %d 枚\n", len(entry.Result.OutputImages))
			}
			if !entry.Result.Success && entry.Result.ErrorMessage != "" {
				_, _ = fmt.Fprintf(w, "      エラー: %s\n", entry.Result.ErrorMessage)
			}
			_, _ = fmt.Fprintln(w, "")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().IntVar(&historyOpts.limit, "limit", 10, "表示する履歴の件数")
}
