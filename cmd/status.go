package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "現在のサブプロジェクトの状態を表示する",
	Long:  "現在のディレクトリに関連するサブプロジェクトの状態を表示します。",
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

		// Load project config
		projectCfg, err := config.LoadProjectConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("プロジェクト設定の読み込みに失敗しました: %w", err)
		}

		w := cmd.OutOrStdout()

		// Check if we're in a subproject
		subprojectName, err := project.FindCurrentSubproject(projectRoot, cwd)
		if err != nil {
			if errors.Is(err, project.ErrNotInSubproject) {
				// Show project-level status
				_, _ = fmt.Fprintf(w, "プロジェクト: %s\n", projectCfg.Name)
				_, _ = fmt.Fprintf(w, "モデル: %s\n", projectCfg.Model)
				_, _ = fmt.Fprintln(w, "")
				_, _ = fmt.Fprintln(w, "サブプロジェクト内にいません。")
				_, _ = fmt.Fprintln(w, "サブプロジェクトに移動するか、新しく作成してください:")
				_, _ = fmt.Fprintln(w, "  cd subprojects/<name>")
				_, _ = fmt.Fprintln(w, "  banago subproject create <name>")
				return nil
			}
			return err
		}

		// Show subproject-level status
		subprojectDir := config.GetSubprojectDir(projectRoot, subprojectName)
		subprojectCfg, err := config.LoadSubprojectConfig(subprojectDir)
		if err != nil {
			return fmt.Errorf("サブプロジェクト設定の読み込みに失敗しました: %w", err)
		}

		_, _ = fmt.Fprintf(w, "プロジェクト: %s\n", projectCfg.Name)
		_, _ = fmt.Fprintf(w, "サブプロジェクト: %s\n", subprojectCfg.Name)
		if subprojectCfg.Description != "" {
			_, _ = fmt.Fprintf(w, "説明: %s\n", subprojectCfg.Description)
		}
		_, _ = fmt.Fprintln(w, "")

		// Context file
		contextPath := filepath.Join(subprojectDir, subprojectCfg.ContextFile)
		if _, err := os.Stat(contextPath); err == nil {
			relPath, _ := filepath.Rel(cwd, contextPath)
			_, _ = fmt.Fprintf(w, "コンテキスト: %s\n", relPath)
		}

		// Character file
		if subprojectCfg.CharacterFile != "" {
			characterPath := filepath.Join(projectRoot, config.CharactersDir, subprojectCfg.CharacterFile)
			relPath, _ := filepath.Rel(cwd, characterPath)
			if _, err := os.Stat(characterPath); err == nil {
				_, _ = fmt.Fprintf(w, "キャラクター: %s\n", relPath)
			} else {
				_, _ = fmt.Fprintf(w, "キャラクター: %s (見つかりません)\n", relPath)
			}
		}
		_, _ = fmt.Fprintln(w, "")

		// Input images
		_, _ = fmt.Fprintln(w, "入力画像:")
		if len(subprojectCfg.InputImages) == 0 {
			_, _ = fmt.Fprintln(w, "  (なし)")
		} else {
			inputsDir := config.GetInputsDir(subprojectDir)
			for _, img := range subprojectCfg.InputImages {
				imgPath := filepath.Join(inputsDir, img)
				relPath, _ := filepath.Rel(cwd, imgPath)
				if _, err := os.Stat(imgPath); err == nil {
					_, _ = fmt.Fprintf(w, "  %s\n", relPath)
				} else {
					_, _ = fmt.Fprintf(w, "  %s (見つかりません)\n", relPath)
				}
			}
		}
		_, _ = fmt.Fprintln(w, "")

		// History summary
		historyDir := config.GetHistoryDir(subprojectDir)
		entries, err := history.ListEntries(historyDir)
		if err != nil {
			_, _ = fmt.Fprintln(w, "履歴: (読み込みエラー)")
		} else if len(entries) == 0 {
			_, _ = fmt.Fprintln(w, "履歴: なし")
		} else {
			_, _ = fmt.Fprintf(w, "履歴: %d 件\n", len(entries))
			// Show latest entry
			latest := entries[len(entries)-1]
			_, _ = fmt.Fprintf(w, "  最新: %s (%s)\n", latest.ID[:8]+"...", latest.CreatedAt[:10])
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
