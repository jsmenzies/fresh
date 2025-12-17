package domain

type LocalState interface {
	isLocal()
}

type CleanLocalState struct{}

// DirtyLocalState TODO: potentially add number of new files, modified files, deleted files, etc to display in UI
type DirtyLocalState struct {
}

type UntrackedLocalState struct {
}

type LocalStateError struct {
	Message string
}

func (CleanLocalState) isLocal()     {}
func (DirtyLocalState) isLocal()     {}
func (UntrackedLocalState) isLocal() {}
func (LocalStateError) isLocal()     {}
