package domain

type CommandOutcome struct {
	ExitCode      int
	FailureReason string
}

func (o CommandOutcome) IsSuccess() bool {
	return o.ExitCode == 0
}

type PruneOutcome struct {
	CommandOutcome
	DeletedCount int
	FailedCount  int
}
