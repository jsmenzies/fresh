package domain

import "time"

type Repository struct {
	Name           string
	Path           string
	RemoteURL      string
	Branches       Branches
	LocalState     LocalState
	RemoteState    RemoteState
	LastCommitTime time.Time
	Activity       Activity
}

func (r Repository) IsBusy() bool {
	return r.Activity.IsInProgress()
}

func (r Repository) CanPull() bool {
	return r.RemoteState.CanPull()
}
