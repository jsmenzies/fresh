package domain

type PullState struct {
	InProgress bool
	Lines      []string
	ExitCode   int
	Completed  bool
}

func NewPullState() *PullState {
	return &PullState{
		InProgress: true,
		Lines:      make([]string, 0),
		ExitCode:   0,
		Completed:  false,
	}
}

func (ps *PullState) AddLine(line string) {
	if ps != nil {
		ps.Lines = append(ps.Lines, line)
	}
}

func (ps *PullState) Complete(exitCode int) {
	if ps != nil {
		ps.InProgress = false
		ps.Completed = true
		ps.ExitCode = exitCode
	}
}

func (ps *PullState) GetLastLine() string {
	if ps != nil && len(ps.Lines) > 0 {
		return ps.Lines[len(ps.Lines)-1]
	}
	return ""
}

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
