package domain

type Repository struct {
	Name         string
	Path         string
	RemoteURL    string
	Branches     Branches
	StashCount   int
	LocalState   LocalState
	RemoteState  RemoteState
	PullRequests PullRequestState
	Activity     Activity
}

func (r Repository) IsBusy() bool {
	return r.Activity.IsInProgress()
}

func (r Repository) CanPull() bool {
	return r.RemoteState.CanPull()
}
