package ui

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

func withRepoPath(path string) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.Path = path
	}
}

func withRepoRemoteURL(remoteURL string) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.RemoteURL = remoteURL
	}
}

func withRepoMutator(mutator func(*domain.Repository)) testRepositoryOption {
	return mutator
}
