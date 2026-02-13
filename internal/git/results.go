package git

// PullResult contains the result of a pull operation.
type PullResult struct {
	Success  bool
	ExitCode int
	Lines    []string
}

// DeleteResult contains the result of deleting branches.
type DeleteResult struct {
	ExitCode     int
	DeletedCount int
	Messages     []string
}
