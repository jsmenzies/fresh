package domain

type LocalState interface {
	isLocal()
}

type CleanLocalState struct{}

type DirtyLocalState struct {
	Added     int
	Modified  int
	Deleted   int
	Untracked int
}

type LocalStateError struct {
	Message string
}

func (CleanLocalState) isLocal() {}
func (DirtyLocalState) isLocal() {}
func (LocalStateError) isLocal() {}