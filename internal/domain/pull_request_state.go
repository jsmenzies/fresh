package domain

type PullRequestState interface {
	isPullRequestState()
}

type PullRequestCount struct {
	Open      int
	MyOpen    int
	MyReady   int
	MyBlocked int
	MyReview  int
	MyChecks  int
}

type PullRequestUnavailable struct{}

type PullRequestError struct {
	Message string
}

func (PullRequestCount) isPullRequestState()       {}
func (PullRequestUnavailable) isPullRequestState() {}
func (PullRequestError) isPullRequestState()       {}
