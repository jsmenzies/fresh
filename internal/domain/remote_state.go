package domain

type RemoteState interface {
	isRemoteState()
	CanPull() bool
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

func (Synced) CanPull() bool         { return false }
func (Ahead) CanPull() bool          { return false }
func (b Behind) CanPull() bool       { return b.Count > 0 }
func (d Diverged) CanPull() bool     { return d.BehindCount > 0 }
func (NoUpstream) CanPull() bool     { return false }
func (DetachedRemote) CanPull() bool { return false }
func (RemoteError) CanPull() bool    { return false }
