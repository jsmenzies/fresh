package listing

import "fresh/internal/domain"

type RepoUpdatedMsg struct {
	Repo  domain.Repository
	Index int
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

type checkoutWorkState struct {
	Index    int
	lineChan chan string
	doneChan chan checkoutCompleteMsg
}

type checkoutLineMsg struct {
	Index int
	line  string
	state *checkoutWorkState
}

type checkoutCompleteMsg struct {
	Index        int
	exitCode     int
	targetBranch string
	Repo         domain.Repository
}
