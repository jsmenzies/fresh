package listing

import "fresh/internal/domain"

type RepoUpdatedMsg struct {
	Repo domain.Repository
}

type pullLineMsg struct {
	repoPath string
	line     string
	state    *pullWorkState
}

type pullCompleteMsg struct {
	repoPath string
	exitCode int
	Repo     domain.Repository
}

type pullWorkState struct {
	repoPath string
	lineChan chan string
	doneChan chan pullCompleteMsg
}
