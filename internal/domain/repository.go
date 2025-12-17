package domain

import "time"

type Repository struct {
	Name           string
	Path           string
	RemoteURL      string
	Branch         Branch
	LocalState     LocalState
	RemoteState    RemoteState
	LastCommitTime time.Time

	// UI state
	Activity     Activity
	ErrorMessage string
}
