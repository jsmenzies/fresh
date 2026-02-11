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
	ErrorMessage   string
}
