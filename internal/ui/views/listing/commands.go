package listing

import (
	"fresh/internal/config"
	"fresh/internal/git"

	tea "charm.land/bubbletea/v2"
)

var cfg = config.DefaultConfig()

type checkoutFn func(repoPath string, lineCallback func(string)) (targetBranch string, exitCode int, err error)

func performRefresh(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		// Fetch first, then build the repo to get fresh status (avoids 3x GetStatus calls)
		git.Fetch(repoPath)
		repo := git.BuildRepository(repoPath, cfg)

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

			repo := git.BuildRepository(repoPath, cfg)

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
			lineCallback := func(line string) {
				lineChan <- line
			}

			_, deleted := git.DeleteBranches(repoPath, branches, lineCallback)

			close(lineChan)

			repo := git.BuildRepository(repoPath, cfg)

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

func performCheckoutIntegration(index int, repoPath string) tea.Cmd {
	return performCheckout(index, repoPath, git.CheckoutIntegration)
}

func performCheckoutPrimary(index int, repoPath string) tea.Cmd {
	return performCheckout(index, repoPath, git.CheckoutPrimary)
}

func performCheckout(index int, repoPath string, fn checkoutFn) tea.Cmd {
	return func() tea.Msg {
		lineChan := make(chan string, 10)
		doneChan := make(chan checkoutCompleteMsg, 1)

		go func() {
			targetBranch, exitCode, _ := fn(repoPath, func(line string) {
				lineChan <- line
			})

			close(lineChan)

			repo := git.BuildRepository(repoPath, cfg)

			doneChan <- checkoutCompleteMsg{
				Index:        index,
				exitCode:     exitCode,
				targetBranch: targetBranch,
				Repo:         repo,
			}
			close(doneChan)
		}()

		return checkoutWorkState{
			Index:    index,
			lineChan: lineChan,
			doneChan: doneChan,
		}
	}
}

func listenForCheckoutProgress(state checkoutWorkState) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-state.lineChan:
			if ok {
				return checkoutLineMsg{
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
