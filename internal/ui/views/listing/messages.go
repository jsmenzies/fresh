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