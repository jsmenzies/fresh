package ui

import (
	"time"

	"fresh/internal/domain"
	"fresh/internal/git"
	"fresh/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
)

func startScanning(scanner *scanner.Scanner, scanDir string) tea.Cmd {
	return func() tea.Msg {
		scanner.StartScanning(scanDir)
		return scanProgressMsg{reposFound: 0}
	}
}

func scanStep(scanner *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		repo, hasMore, found := scanner.ScanStep()

		if !hasMore {
			return scanCompleteMsg(scanner.GetFoundRepositories())
		}

		if found {
			return repoFoundMsg(repo)
		}

		return scanProgressMsg{
			reposFound: scanner.GetRepoCount(),
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return scanTickMsg(t)
	})
}

func startBackgroundRefresh(repos []domain.Repository) tea.Cmd {
	var cmds []tea.Cmd
	for i, repo := range repos {
		cmds = append(cmds, func(repoIndex int, repoPath string) tea.Cmd {
			return func() tea.Msg {
				return refreshStartMsg{
					repoIndex: repoIndex,
					repoPath:  repoPath,
				}
			}
		}(i, repo.Path))
	}
	return tea.Batch(cmds...)
}

func performRefresh(repoIndex int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		ahead, behind, err := git.RefreshRemoteStatus(repoPath)

		if err != nil {
			return refreshCompleteMsg{
				repoIndex:    repoIndex,
				aheadCount:   0,
				behindCount:  0,
				hasError:     true,
				errorMessage: err.Error(),
			}
		}

		return refreshCompleteMsg{
			repoIndex:    repoIndex,
			aheadCount:   ahead,
			behindCount:  behind,
			hasError:     false,
			errorMessage: "",
		}
	}
}

func performPull(repoIndex int, repoPath string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return pullStartMsg{repoIndex: repoIndex}
		},
		func() tea.Msg {
			// Create channels for lines and completion
			lineChan := make(chan string, 10)
			doneChan := make(chan pullCompleteMsg, 1)

			// Start git pull in a goroutine
			go func() {
				exitCode := git.Pull(repoPath, func(line string) {
					lineChan <- line
				})

				close(lineChan)

				doneChan <- pullCompleteMsg{
					repoIndex: repoIndex,
					exitCode:  exitCode,
				}
				close(doneChan)
			}()

			// Start listening for line updates
			return pullWorkState{
				repoIndex: repoIndex,
				lineChan:  lineChan,
				doneChan:  doneChan,
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
					repoIndex: state.repoIndex,
					line:      line,
					state:     &state,
				}
			}
			return <-state.doneChan
		case complete := <-state.doneChan:
			return complete
		}
	}
}
