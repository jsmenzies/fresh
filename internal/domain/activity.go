package domain

import "charm.land/bubbles/v2/spinner"

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

type CommandCompletion struct {
	Outcome  CommandOutcome
	Complete bool
}

func (c *CommandCompletion) MarkComplete(outcome CommandOutcome) {
	c.Complete = true
	c.Outcome = outcome
}

func (c *CommandCompletion) IsInProgress() bool { return !c.Complete }

type PullingActivity struct {
	LineBuffer
	Spinner spinner.Model
	CommandCompletion
}

func (*PullingActivity) isActivity() {}

func (p *PullingActivity) MarkComplete(outcome CommandOutcome) {
	p.CommandCompletion.MarkComplete(outcome)
}

func (p *PullingActivity) IsInProgress() bool { return p.CommandCompletion.IsInProgress() }

type PruningActivity struct {
	LineBuffer
	Spinner spinner.Model
	CommandCompletion
	DeletedCount int
	FailedCount  int
}

func (*PruningActivity) isActivity() {}

func (p *PruningActivity) MarkComplete(outcome PruneOutcome) {
	p.CommandCompletion.MarkComplete(outcome.CommandOutcome)
	p.DeletedCount = outcome.DeletedCount
	p.FailedCount = outcome.FailedCount
}

func (p *PruningActivity) IsInProgress() bool { return p.CommandCompletion.IsInProgress() }
