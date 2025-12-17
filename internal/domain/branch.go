package domain

type Branch interface {
	isBranch()
}

type OnBranch struct {
	Name string
}

type DetachedHead struct {
	CommitSHA string
}

type NoBranch struct {
	Reason string
}

func (OnBranch) isBranch()     {}
func (DetachedHead) isBranch() {}
func (NoBranch) isBranch()     {}
