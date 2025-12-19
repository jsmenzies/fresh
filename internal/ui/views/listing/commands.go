package listing

import (
	"fresh/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

func performRefresh(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		repo := git.BuildRepository(repoPath)

		err := git.RefreshRemoteStatusWithFetch(&repo)
		if err != nil {
			repo.ErrorMessage = err.Error()
		} else {
			repo.ErrorMessage = ""
		}

		repo.RemoteState = git.GetStatus(repoPath)

		return RepoUpdatedMsg{
			Repo:  repo,
			Index: index,
		}
	}
}

func performPull(index int, repoPath string) tea.Cmd {
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
				Index:    index,
				exitCode: exitCode,
				Repo:     repo,
			}
			close(doneChan)
		}()

		return pullWorkState{
			Index:    index,
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
					Index: state.Index,
					line:  line,
					state: &state,
				}
			}
			return <-state.doneChan
		case complete := <-state.doneChan:
			return complete
		}
	}
}
