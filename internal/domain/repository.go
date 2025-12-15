package domain

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

type Repository struct {
	Name           string
	Path           string
	LastCommitTime time.Time
	RemoteURL      string
	CurrentBranch  string
	HasModified    bool
	AheadCount     int
	BehindCount    int

	// UI state (will be refactored out later)
	Fetching         bool
	Done             bool
	Refreshing       bool
	RefreshSpinner   spinner.Model
	HasRemoteUpdates bool
	HasError         bool
	ErrorMessage     string
	PullState        *PullState
	PullSpinner      spinner.Model
}
