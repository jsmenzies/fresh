package domain

import "github.com/charmbracelet/bubbles/spinner"

type Activity interface {
	isActivity()
}

type IdleActivity struct{}

func (IdleActivity) isActivity() {}

type RefreshingActivity struct {
	Spinner  spinner.Model
	Complete bool
}

func (RefreshingActivity) isActivity() {}

func (r *RefreshingActivity) MarkComplete() {
	r.Complete = true
}

type PullingActivity struct {
	Spinner  spinner.Model
	Lines    []string
	ExitCode int
	Complete bool
}

func (PullingActivity) isActivity() {}

func (p *PullingActivity) AddLine(line string) {
	p.Lines = append(p.Lines, line)
}

func (p *PullingActivity) GetLastLine() string {
	if len(p.Lines) == 0 {
		return ""
	}
	return p.Lines[len(p.Lines)-1]
}

func (p *PullingActivity) MarkComplete(exitCode int) {
	p.Complete = true
	p.ExitCode = exitCode
}

func (p *PullingActivity) GetAllOutput() string {
	if len(p.Lines) == 0 {
		return ""
	}

	var result string
	for _, line := range p.Lines {
		result += line + "\n"
	}
	return result
}

type PruningActivity struct {
	Spinner      spinner.Model
	Lines        []string
	DeletedCount int
	TotalCount   int
	ExitCode     int
	Complete     bool
}

func (PruningActivity) isActivity() {}

func (p *PruningActivity) AddLine(line string) {
	p.Lines = append(p.Lines, line)
}

func (p *PruningActivity) GetLastLine() string {
	if len(p.Lines) == 0 {
		return ""
	}
	return p.Lines[len(p.Lines)-1]
}

func (p *PruningActivity) MarkComplete(exitCode int, deletedCount int) {
	p.Complete = true
	p.ExitCode = exitCode
	p.DeletedCount = deletedCount
}

func (p *PruningActivity) GetAllOutput() string {
	if len(p.Lines) == 0 {
		return ""
	}

	var result string
	for _, line := range p.Lines {
		result += line + "\n"
	}
	return result
}
