package domain

import "github.com/charmbracelet/bubbles/spinner"

type Activity interface {
	isActivity()
	IsInProgress() bool
}

type IdleActivity struct{}

func (IdleActivity) isActivity() {}

func (IdleActivity) IsInProgress() bool { return false }

type RefreshingActivity struct {
	Spinner  spinner.Model
	Complete bool
}

func (*RefreshingActivity) isActivity() {}

func (r *RefreshingActivity) MarkComplete() {
	r.Complete = true
}

func (r *RefreshingActivity) IsInProgress() bool { return !r.Complete }

type LineBuffer struct {
	Lines []string
}

func (lb *LineBuffer) AddLine(line string) {
	lb.Lines = append(lb.Lines, line)
}

func (lb *LineBuffer) GetLastLine() string {
	if len(lb.Lines) == 0 {
		return ""
	}
	return lb.Lines[len(lb.Lines)-1]
}

type PullingActivity struct {
	LineBuffer
	Spinner  spinner.Model
	ExitCode int
	Complete bool
}

func (*PullingActivity) isActivity() {}

func (p *PullingActivity) MarkComplete(exitCode int) {
	p.Complete = true
	p.ExitCode = exitCode
}

func (p *PullingActivity) IsInProgress() bool { return !p.Complete }

type PruningActivity struct {
	LineBuffer
	Spinner      spinner.Model
	DeletedCount int
	ExitCode     int
	Complete     bool
}

func (*PruningActivity) isActivity() {}

func (p *PruningActivity) MarkComplete(exitCode int, deletedCount int) {
	p.Complete = true
	p.ExitCode = exitCode
	p.DeletedCount = deletedCount
}

func (p *PruningActivity) IsInProgress() bool { return !p.Complete }
