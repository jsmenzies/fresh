package domain

import "time"

type PullRequestDetails struct {
	Number    int
	Title     string
	IsMine    bool
	UpdatedAt time.Time
	Checks    PullRequestChecks
}

type PullRequestChecks struct {
	Total   int
	Passed  int
	Waiting int
	Running int
	Failed  int
	Skipped int
}
