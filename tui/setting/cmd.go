package setting

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"kanba/git"
	"kanba/update"
)

func GitDiffCmd(repoPath string, args []string) tea.Cmd {
	return func() tea.Msg {
		runner := git.NewGitRunner(repoPath)
		cmd := &git.DiffCommand{
			DiffArgs: git.DiffArgs{Show: false, Args: args},
			Parser:   git.NewUnifiedParser(),
			Aligner:  &git.UnifiedAligner{},
		}
		diffs, err := git.Execute(context.Background(), runner, cmd)
		if err != nil {
			return DiffMsg{nil, err}
		}

		untracked, err := git.Execute(context.Background(), runner, &git.LsFilesCommand{})
		if err != nil {
			return DiffMsg{diffs, nil}
		}

		for _, fp := range untracked {
			d, err := git.UntrackedToSideBySideDiff(repoPath, fp)
			if err != nil {
				continue
			}
			if d.NewPath != "" {
				diffs = append(diffs, d)
			}
		}

		return DiffMsg{diffs, nil}
	}
}

func UpdateCheckCmd(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		latest, available, err := update.CheckLatest(context.Background(), currentVersion)
		if err != nil || !available {
			return nil
		}
		return UpdateCheckMsg{Version: latest, Available: true}
	}
}

func UpdateInstallCmd() tea.Cmd {
	return func() tea.Msg {
		return UpdateInstallMsg{Err: update.Run()}
	}
}
