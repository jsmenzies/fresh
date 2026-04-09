package listing

import (
	"time"

	"fresh/internal/domain"

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
	States  map[string]domain.PullRequestState
	Trigger PullRequestSyncTrigger
}

type pullLineMsg struct {
	Index int
	line  string
	state *pullWorkState
}

type pullCompleteMsg struct {
	Index    int
	exitCode int
	Repo     domain.Repository
}

type pullWorkState struct {
	Index    int
	lineChan chan string
	doneChan chan pullCompleteMsg
}

type pruneWorkState struct {
	Index    int
	lineChan chan string
	doneChan chan pruneCompleteMsg
}

type pruneLineMsg struct {
	Index int
	line  string
	state *pruneWorkState
}

type pruneCompleteMsg struct {
	Index        int
	exitCode     int
	Repo         domain.Repository
	DeletedCount int
}

type infoRotateTickMsg struct{}

func scheduleInfoRotateTick(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return infoRotateTickMsg{}
	})
}
