package ui

import "fresh/internal/domain"

func makeTestRepository(name string) domain.Repository {
	return newTestRepository(name).Build()
}

type testRepositoryBuilder struct {
	repo domain.Repository
}

func newTestRepository(name string) *testRepositoryBuilder {
	return &testRepositoryBuilder{
		repo: domain.Repository{
			Name:        name,
			Path:        "/tmp/" + name,
			Activity:    domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}
}

func (b *testRepositoryBuilder) Name(name string) *testRepositoryBuilder {
	b.repo.Name = name
	return b
}

func (b *testRepositoryBuilder) Path(path string) *testRepositoryBuilder {
	b.repo.Path = path
	return b
}

func (b *testRepositoryBuilder) RemoteURL(remoteURL string) *testRepositoryBuilder {
	b.repo.RemoteURL = remoteURL
	return b
}

func (b *testRepositoryBuilder) Activity(activity domain.Activity) *testRepositoryBuilder {
	b.repo.Activity = activity
	return b
}

func (b *testRepositoryBuilder) LocalState(state domain.LocalState) *testRepositoryBuilder {
	b.repo.LocalState = state
	return b
}

func (b *testRepositoryBuilder) RemoteState(state domain.RemoteState) *testRepositoryBuilder {
	b.repo.RemoteState = state
	return b
}

func (b *testRepositoryBuilder) PullRequests(state domain.PullRequestState) *testRepositoryBuilder {
	b.repo.PullRequests = state
	return b
}

func (b *testRepositoryBuilder) CurrentBranch(branch domain.Branch) *testRepositoryBuilder {
	b.repo.Branches.Current = branch
	return b
}

func (b *testRepositoryBuilder) MergedBranches(merged ...string) *testRepositoryBuilder {
	b.repo.Branches.Merged = append([]string(nil), merged...)
	return b
}

func (b *testRepositoryBuilder) StashCount(count int) *testRepositoryBuilder {
	b.repo.StashCount = count
	return b
}

func (b *testRepositoryBuilder) Build() domain.Repository {
	return b.repo
}
