package domain

import "time"

type Repository struct {
	Name             string
	Path             string
	RemoteURL        string
	Branch           Branch
	LocalState       LocalState
	RemoteState      RemoteState
	LastCommitTime   time.Time
	Activity         Activity
	ErrorMessage     string
	MergedBranches   []string
	SquashedBranches []string
}
