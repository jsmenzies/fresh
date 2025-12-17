package listing

import (
	"fresh/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

func performRefresh(repoPath string) tea.Cmd {
	return func() tea.Msg {
		repo := git.BuildRepository(repoPath)

		// First do a fetch to update remote tracking branches
		err := git.RefreshRemoteStatusWithFetch(&repo)
		if err != nil {
			repo.ErrorMessage = err.Error()
		} else {
			repo.ErrorMessage = ""
		}

		// Re-get status after fetch
		repo.RemoteState = git.GetStatus(repoPath)

		return RepoUpdatedMsg{
			Repo: repo,
		}
	}
}

func performPull(repoPath string) tea.Cmd {
	return func() tea.Msg {
		lineChan := make(chan string, 10)
		doneChan := make(chan pullCompleteMsg, 1)

		go func() {
			exitCode := git.Pull(repoPath, func(line string) {
				lineChan <- line
			})

			close(lineChan)

			repo := git.BuildRepository(repoPath)

			doneChan <- pullCompleteMsg{
				repoPath: repoPath,
				exitCode: exitCode,
				Repo:     repo,
			}
			close(doneChan)
		}()

		return pullWorkState{
			repoPath: repoPath,
			lineChan: lineChan,
			doneChan: doneChan,
		}
	}
}

func listenForPullProgress(state pullWorkState) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-state.lineChan:
			if ok {
				return pullLineMsg{
					repoPath: state.repoPath,
					line:     line,
					state:    &state,
				}
			}
			return <-state.doneChan
		case complete := <-state.doneChan:
			return complete
		}
	}
}
