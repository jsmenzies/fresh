package listing

import (
	"fresh/internal/config"
	"fresh/internal/domain"
	"fresh/internal/git"

	tea "charm.land/bubbletea/v2"
)

var cfg = config.DefaultConfig()

func performInitialRefresh(index int, existingRepo domain.Repository) tea.Cmd {
	return func() tea.Msg {
		repo := git.RefreshRepository(existingRepo.Path, cfg, git.RefreshRepositoryOptions{
			Mode:     git.RefreshModeFetchRemoteOnly,
			Existing: &existingRepo,
		})

		return RepoUpdatedMsg{
			Repo:  repo,
			Index: index,
		}
	}
}

func performRefresh(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		repo := git.RefreshRepository(repoPath, cfg, git.RefreshRepositoryOptions{
			Mode: git.RefreshModeFetchAndBuild,
		})

		return RepoUpdatedMsg{
			Repo:  repo,
			Index: index,
		}
	}
}

func performPull(index int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		return startStreamedRepoCommand(
			index,
			repoPath,
			func(lineCallback func(string)) domain.CommandOutcome {
				return git.Pull(repoPath, lineCallback)
			},
			func(index int, repo domain.Repository, outcome domain.CommandOutcome) pullCompleteMsg {
				return pullCompleteMsg{
					Index:   index,
					outcome: outcome,
					Repo:    repo,
				}
			},
		)
	}
}

func listenForPullProgress(state pullWorkState) tea.Cmd {
	return listenForStreamedProgress(state, func(index int, line string, next *pullWorkState) tea.Msg {
		return pullLineMsg{
			Index: index,
			line:  line,
			state: next,
		}
	})
}

func performPrune(index int, repoPath string, branches []string) tea.Cmd {
	return func() tea.Msg {
		return startStreamedRepoCommand(
			index,
			repoPath,
			func(lineCallback func(string)) domain.PruneOutcome {
				return git.DeleteBranches(repoPath, branches, lineCallback)
			},
			func(index int, repo domain.Repository, outcome domain.PruneOutcome) pruneCompleteMsg {
				return pruneCompleteMsg{
					Index:   index,
					outcome: outcome,
					Repo:    repo,
				}
			},
		)
	}
}

func listenForPruneProgress(state pruneWorkState) tea.Cmd {
	return listenForStreamedProgress(state, func(index int, line string, next *pruneWorkState) tea.Msg {
		return pruneLineMsg{
			Index: index,
			line:  line,
			state: next,
		}
	})
}

func performPullRequestSync(repos []domain.Repository, trigger PullRequestSyncTrigger, generation uint64) tea.Cmd {
	snapshot := append([]domain.Repository(nil), repos...)

	return func() tea.Msg {
		sync := git.GetPullRequestSync(snapshot)
		return PullRequestStatesUpdatedMsg{
			Generation: generation,
			States:     sync.States,
			Tracked:    sync.Tracked,
			Trigger:    trigger,
		}
	}
}

func openPullRequestsView(repo domain.Repository) tea.Cmd {
	return func() tea.Msg {
		return OpenPullRequestsMsg{Repo: repo}
	}
}
