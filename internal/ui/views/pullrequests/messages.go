package pullrequests

import (
	"time"

	"fresh/internal/domain"

	tea "charm.land/bubbletea/v2"
)

type BackToRepoListMsg struct{}

type PullRequestsLoadedMsg struct {
	RepoPath     string
	PullRequests []domain.PullRequestDetails
	Unsupported  bool
	Error        string
}

type pulseTickMsg struct{}

func schedulePulseTick(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = 350 * time.Millisecond
	}

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return pulseTickMsg{}
	})
}
