package ui

import (
	"fresh/internal/domain"
	"time"
)

// Message types for Bubble Tea

type scanProgressMsg struct {
	reposFound int
}

type repoFoundMsg domain.Repository

type scanCompleteMsg []domain.Repository

type scanTickMsg time.Time

type refreshStartMsg struct {
	repoIndex int
	repoPath  string
}

type refreshCompleteMsg struct {
	repoIndex    int
	aheadCount   int
	behindCount  int
	hasError     bool
	errorMessage string
}

type pullStartMsg struct {
	repoIndex int
}

type pullLineMsg struct {
	repoIndex int
	line      string
	state     *pullWorkState
}

type pullCompleteMsg struct {
	repoIndex int
	exitCode  int
}

type pullWorkState struct {
	repoIndex int
	lineChan  chan string
	doneChan  chan pullCompleteMsg
}
