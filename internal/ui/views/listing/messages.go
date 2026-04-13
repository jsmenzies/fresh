package listing

import (
	"time"

	"fresh/internal/domain"
	"fresh/internal/pullrequests"

	tea "charm.land/bubbletea/v2"
)

type RepoUpdatedMsg struct {
	Repo  domain.Repository
	Index int
}

type PullRequestSyncTrigger int

const (
	pullRequestSyncStartup PullRequestSyncTrigger = iota + 1
	pullRequestSyncManual
	pullRequestSyncWatch
)

type PullRequestStatesUpdatedMsg struct {
	Generation uint64
	States     map[string]domain.PullRequestState
	Tracked    []pullrequests.Snapshot
	Trigger    PullRequestSyncTrigger
}

type OpenPullRequestsMsg struct {
	Repo domain.Repository
}

type pullLineMsg struct {
	Index int
	line  string
	state *pullWorkState
}

type pullCompleteMsg struct {
	Index   int
	outcome domain.CommandOutcome
	Repo    domain.Repository
}

type pruneLineMsg struct {
	Index int
	line  string
	state *pruneWorkState
}

type pruneCompleteMsg struct {
	Index   int
	outcome domain.PruneOutcome
	Repo    domain.Repository
}

type infoRotateTickMsg struct{}

var scheduleTick = tea.Tick

func scheduleInfoRotateTick(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	return scheduleTick(interval, func(time.Time) tea.Msg {
		return infoRotateTickMsg{}
	})
}
