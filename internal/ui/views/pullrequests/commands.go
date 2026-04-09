package pullrequests

import (
	"errors"

	"fresh/internal/domain"
	"fresh/internal/git"

	tea "charm.land/bubbletea/v2"
)

func performPullRequestLoad(repo domain.Repository) tea.Cmd {
	return func() tea.Msg {
		rows, err := git.GetRepositoryPullRequests(repo)
		msg := PullRequestsLoadedMsg{
			RepoPath:     repo.Path,
			PullRequests: rows,
		}
		if err != nil {
			msg.Error = err.Error()
			msg.Unsupported = errors.Is(err, git.ErrPullRequestDetailsUnsupported)
		}
		return msg
	}
}

func backToRepoList() tea.Cmd {
	return func() tea.Msg {
		return BackToRepoListMsg{}
	}
}
