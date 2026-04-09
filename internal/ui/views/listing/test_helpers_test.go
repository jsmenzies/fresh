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

func withRepoPath(path string) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.Path = path
	}
}

func withRepoActivity(activity domain.Activity) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.Activity = activity
	}
}

func withRepoLocalState(state domain.LocalState) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.LocalState = state
	}
}

func withRepoRemoteState(state domain.RemoteState) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.RemoteState = state
	}
}

func withRepoPullRequests(state domain.PullRequestState) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.PullRequests = state
	}
}

func withRepoCurrentBranch(branch domain.Branch) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.Branches.Current = branch
	}
}

func withRepoMergedBranches(merged ...string) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.Branches.Merged = append([]string(nil), merged...)
	}
}

func withRepoStashCount(count int) testRepositoryOption {
	return func(repo *domain.Repository) {
		repo.StashCount = count
	}
}

func withRepoMutator(mutator func(*domain.Repository)) testRepositoryOption {
	return mutator
}
