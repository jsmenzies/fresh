package listing

import "fresh/internal/domain"

type testRepositoryOption func(*domain.Repository)

func makeTestRepository(name string, opts ...testRepositoryOption) domain.Repository {
	repo := domain.Repository{
		Name:        name,
		Path:        "/tmp/" + name,
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}

	for _, opt := range opts {
		opt(&repo)
	}

	return repo
}

func withRepo(overrides domain.Repository) testRepositoryOption {
	return func(repo *domain.Repository) {
		if overrides.Name != "" {
			repo.Name = overrides.Name
		}
		if overrides.Path != "" {
			repo.Path = overrides.Path
		}
		if overrides.RemoteURL != "" {
			repo.RemoteURL = overrides.RemoteURL
		}
		if overrides.Activity != nil {
			repo.Activity = overrides.Activity
		}
		if overrides.LocalState != nil {
			repo.LocalState = overrides.LocalState
		}
		if overrides.RemoteState != nil {
			repo.RemoteState = overrides.RemoteState
		}
		if overrides.PullRequests != nil {
			repo.PullRequests = overrides.PullRequests
		}
		if overrides.Branches.Current != nil {
			repo.Branches.Current = overrides.Branches.Current
		}
		if overrides.Branches.Merged != nil {
			repo.Branches.Merged = append([]string(nil), overrides.Branches.Merged...)
		}
		if overrides.StashCount != 0 {
			repo.StashCount = overrides.StashCount
		}
	}
}
