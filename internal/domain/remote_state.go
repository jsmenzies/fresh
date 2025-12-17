package domain

type RemoteState interface {
	isRemoteState()
}

type Synced struct{}

type Ahead struct {
	Count int
}

type Behind struct {
	Count int
}

type Diverged struct {
	AheadCount  int
	BehindCount int
}

type NoUpstream struct {
}

type DetachedRemote struct{}

type RemoteError struct {
	Message string
}

func (Synced) isRemoteState()         {}
func (Ahead) isRemoteState()          {}
func (Behind) isRemoteState()         {}
func (Diverged) isRemoteState()       {}
func (NoUpstream) isRemoteState()     {}
func (DetachedRemote) isRemoteState() {}
func (RemoteError) isRemoteState()    {}
