package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

// PullState holds all data from a git pull operation
type PullState struct {
	InProgress bool
	Lines      []string // All output lines from git pull (both stdout and stderr)
	ExitCode   int      // 0 for success, non-zero for error
	Completed  bool
}

// repository represents a Git repository and its current state
type repository struct {
	name             string
	path             string
	fetching         bool
	done             bool
	lastCommitTime   time.Time
	remoteURL        string
	hasModified      bool
	aheadCount       int
	behindCount      int
	currentBranch    string
	refreshing       bool
	refreshSpinner   spinner.Model
	hasRemoteUpdates bool
	hasError         bool
	errorMessage     string
	pullState        *PullState    // nil = no pull performed yet
	pullSpinner      spinner.Model // Spinner for git pull progress
}

// NewPullState creates a new PullState in progress
func NewPullState() *PullState {
	return &PullState{
		InProgress: true,
		Lines:      make([]string, 0),
		ExitCode:   0,
		Completed:  false,
	}
}

// AddLine appends a new output line to the pull state
func (ps *PullState) AddLine(line string) {
	if ps != nil {
		ps.Lines = append(ps.Lines, line)
	}
}

// Complete marks the pull as finished with the given exit code
func (ps *PullState) Complete(exitCode int) {
	if ps != nil {
		ps.InProgress = false
		ps.Completed = true
		ps.ExitCode = exitCode
	}
}

// GetLastLine returns the most recent output line, or empty string
func (ps *PullState) GetLastLine() string {
	if ps != nil && len(ps.Lines) > 0 {
		return ps.Lines[len(ps.Lines)-1]
	}
	return ""
}

// GetAllOutput returns all lines joined as a single string
func (ps *PullState) GetAllOutput() string {
	if ps == nil || len(ps.Lines) == 0 {
		return ""
	}

	result := ""
	for _, line := range ps.Lines {
		result += line + "\n"
	}
	return result
}
