package listing

import (
	"path/filepath"
	"sync"

	"fresh/internal/domain"
	"fresh/internal/git"
)

type fakeGitClient struct {
	mu sync.Mutex

	reposByPath map[string]domain.Repository
	callLog     []string

	fetchErr      error
	pullExitCode  int
	pruneExitCode int
	pruneDeleted  int
	pullProgress  []string
	pruneProgress []string
}

var _ git.Client = (*fakeGitClient)(nil)

func newFakeGitClient() *fakeGitClient {
	return &fakeGitClient{
		reposByPath: make(map[string]domain.Repository),
	}
}

func (f *fakeGitClient) BuildRepository(path string) domain.Repository {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, "BuildRepository:"+path)
	if repo, ok := f.reposByPath[path]; ok {
		return repo
	}

	name := filepath.Base(path)
	return domain.Repository{
		Name:        name,
		Path:        path,
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}
}

func (f *fakeGitClient) Fetch(repoPath string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, "Fetch:"+repoPath)
	return f.fetchErr
}

func (f *fakeGitClient) Pull(repoPath string, lineCallback func(string)) int {
	f.mu.Lock()
	f.callLog = append(f.callLog, "Pull:"+repoPath)
	progress := append([]string(nil), f.pullProgress...)
	exitCode := f.pullExitCode
	f.mu.Unlock()

	if lineCallback != nil {
		for _, line := range progress {
			lineCallback(line)
		}
	}

	return exitCode
}

func (f *fakeGitClient) DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (int, int) {
	f.mu.Lock()
	f.callLog = append(f.callLog, "DeleteBranches:"+repoPath)
	progress := append([]string(nil), f.pruneProgress...)
	exitCode := f.pruneExitCode
	deleted := f.pruneDeleted
	f.mu.Unlock()

	if lineCallback != nil {
		for _, line := range progress {
			lineCallback(line)
		}
	}

	return exitCode, deleted
}

func newTestModel(repos []domain.Repository) *Model {
	return New(repos, newFakeGitClient())
}
