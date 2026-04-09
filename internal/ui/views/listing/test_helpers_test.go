package listing

import "fresh/internal/domain"

func makeTestRepository(name string) domain.Repository {
	return domain.Repository{
		Name:        name,
		Path:        "/tmp/" + name,
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}
}
