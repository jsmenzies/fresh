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

func performPrune(index int, repoPath string, branches []string) tea.Cmd {
	return func() tea.Msg {
		// Note: branches are pre-fetched and passed in

		lineChan := make(chan string, 10)
		doneChan := make(chan pruneCompleteMsg, 1)

		go func() {
			deletedCount := 0
			lineCallback := func(line string) {
				lineChan <- line
				if len(line) > 9 && line[:9] == "Deleted: " {
					deletedCount++
				}
			}

			_, deleted := git.DeleteBranches(repoPath, branches, lineCallback)

			close(lineChan)

			repo := git.BuildRepository(repoPath)

			doneChan <- pruneCompleteMsg{
				Index:        index,
				exitCode:     0,
				Repo:         repo,
				DeletedCount: deleted,
			}
			close(doneChan)
		}()

		return pruneWorkState{
			Index:    index,
			lineChan: lineChan,
			doneChan: doneChan,
		}
	}
}

func listenForPruneProgress(state pruneWorkState) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-state.lineChan:
			if ok {
				return pruneLineMsg{
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
