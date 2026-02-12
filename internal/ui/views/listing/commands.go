package listing

import (
	tea "github.com/charmbracelet/bubbletea"
)

func performRefresh(m *Model, index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		// Fetch first, then build the repo to get fresh status (avoids 3x GetStatus calls)
		_ = m.GitClient.Fetch(repoPath)
		repo := m.GitClient.BuildRepository(repoPath)

		return RepoUpdatedMsg{
			Repo:  repo,
			Index: index,
		}
	}
}

func performPull(m *Model, index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		lineChan := make(chan string, 10)
		doneChan := make(chan pullCompleteMsg, 1)

		go func() {
			exitCode := m.GitClient.Pull(repoPath, func(line string) {
				lineChan <- line
			})

			close(lineChan)

			repo := m.GitClient.BuildRepository(repoPath)

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

func performPrune(m *Model, index int, repoPath string, branches []string) tea.Cmd {
	return func() tea.Msg {
		// Note: branches are pre-fetched and passed in

		lineChan := make(chan string, 10)
		doneChan := make(chan pruneCompleteMsg, 1)

		go func() {
			lineCallback := func(line string) {
				lineChan <- line
			}

			_, deleted := m.GitClient.DeleteBranches(repoPath, branches, lineCallback)

			close(lineChan)

			repo := m.GitClient.BuildRepository(repoPath)

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
