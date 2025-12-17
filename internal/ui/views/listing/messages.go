package listing

type refreshStartMsg struct {
	repoPath string
}

type refreshCompleteMsg struct {
	repoPath     string
	aheadCount   int
	behindCount  int
	hasError     bool
	errorMessage string
}

type pullStartMsg struct {
	repoPath string
}

type pullLineMsg struct {
	repoPath string
	line     string
	state    *pullWorkState
}

type pullCompleteMsg struct {
	repoPath string
	exitCode int
}

type pullWorkState struct {
	repoPath string
	lineChan chan string
	doneChan chan pullCompleteMsg
}
