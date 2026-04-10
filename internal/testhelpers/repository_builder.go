package testhelpers

import "fresh/internal/domain"

type RepositoryBuilder struct {
	repo domain.Repository
}

func NewTestRepository(name string) *RepositoryBuilder {
	return &RepositoryBuilder{
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

func (b *RepositoryBuilder) Name(name string) *RepositoryBuilder {
	b.repo.Name = name
	return b
}

func (b *RepositoryBuilder) Path(path string) *RepositoryBuilder {
	b.repo.Path = path
	return b
}

func (b *RepositoryBuilder) RemoteURL(remoteURL string) *RepositoryBuilder {
	b.repo.RemoteURL = remoteURL
	return b
}

func (b *RepositoryBuilder) Activity(activity domain.Activity) *RepositoryBuilder {
	b.repo.Activity = activity
	return b
}

func (b *RepositoryBuilder) LocalState(state domain.LocalState) *RepositoryBuilder {
	b.repo.LocalState = state
	return b
}

func (b *RepositoryBuilder) RemoteState(state domain.RemoteState) *RepositoryBuilder {
	b.repo.RemoteState = state
	return b
}

func (b *RepositoryBuilder) PullRequests(state domain.PullRequestState) *RepositoryBuilder {
	b.repo.PullRequests = state
	return b
}

func (b *RepositoryBuilder) CurrentBranch(branch domain.Branch) *RepositoryBuilder {
	b.repo.Branches.Current = branch
	return b
}

func (b *RepositoryBuilder) MergedBranches(merged ...string) *RepositoryBuilder {
	b.repo.Branches.Merged = append([]string(nil), merged...)
	return b
}

func (b *RepositoryBuilder) StashCount(count int) *RepositoryBuilder {
	b.repo.StashCount = count
	return b
}

func (b *RepositoryBuilder) Build() domain.Repository {
	return b.repo
}
