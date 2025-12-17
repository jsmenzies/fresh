package listing

import (
	"fresh/internal/domain"
	"fresh/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

func startBackgroundRefresh(repos []domain.Repository) tea.Cmd {
	var cmds []tea.Cmd
	for _, repo := range repos {
		cmd := func(repoPath string) tea.Cmd {
			return func() tea.Msg {
				return refreshStartMsg{repoPath: repoPath}
			}
		}(repo.Path)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func performRefresh(repoPath string) tea.Cmd {
	return func() tea.Msg {
		ahead, behind, err := git.RefreshRemoteStatus(repoPath)
		if err != nil {
			return refreshCompleteMsg{
				repoPath:     repoPath,
				aheadCount:   0,
				behindCount:  0,
				hasError:     true,
				errorMessage: err.Error(),
			}
		}
		return refreshCompleteMsg{
			repoPath:     repoPath,
			aheadCount:   ahead,
			behindCount:  behind,
			hasError:     false,
			errorMessage: "",
		}
	}
}

func performPull(repoPath string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return pullStartMsg{repoPath: repoPath}
		},
		func() tea.Msg {
			lineChan := make(chan string, 10)
			doneChan := make(chan pullCompleteMsg, 1)

			go func() {
				exitCode := git.Pull(repoPath, func(line string) {
					lineChan <- line
				})

				close(lineChan)

				doneChan <- pullCompleteMsg{
					repoPath: repoPath,
					exitCode: exitCode,
				}
				close(doneChan)
			}()

			return pullWorkState{
				repoPath: repoPath,
				lineChan: lineChan,
				doneChan: doneChan,
			}
		},
	)
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
