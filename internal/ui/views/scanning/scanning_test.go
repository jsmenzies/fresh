package scanning

import (
	"testing"
	"time"

	"fresh/internal/domain"
	"fresh/internal/git"
)

type fakeScanningGitClient struct {
	reposByPath map[string]domain.Repository
}

var _ git.Client = (*fakeScanningGitClient)(nil)

func newFakeScanningGitClient() *fakeScanningGitClient {
	return &fakeScanningGitClient{
		reposByPath: make(map[string]domain.Repository),
	}
}

func (f *fakeScanningGitClient) BuildRepository(path string) domain.Repository {
	if repo, ok := f.reposByPath[path]; ok {
		return repo
	}
	return domain.Repository{
		Name:        "repo",
		Path:        path,
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}
}

func (f *fakeScanningGitClient) Fetch(repoPath string) error { return nil }

func (f *fakeScanningGitClient) Pull(repoPath string, lineCallback func(string)) int { return 0 }

func (f *fakeScanningGitClient) DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (int, int) {
	return 0, 0
}

type fakeRepoScanner struct {
	ch   chan string
	done chan struct{}
}

func newFakeRepoScanner(paths ...string) *fakeRepoScanner {
	ch := make(chan string, len(paths))
	for _, path := range paths {
		ch <- path
	}
	close(ch)

	return &fakeRepoScanner{
		ch:   ch,
		done: make(chan struct{}),
	}
}

func (f *fakeRepoScanner) Scan() {
	close(f.done)
}

func (f *fakeRepoScanner) GetRepoChannel() chan string {
	return f.ch
}

func TestNewWithDependencies_PanicsOnNilDependencies(t *testing.T) {
	t.Parallel()

	t.Run("nil git client", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil git client")
			}
		}()

		NewWithDependencies(nil, newFakeRepoScanner())
	})

	t.Run("nil scanner", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for nil scanner")
			}
		}()

		NewWithDependencies(newFakeScanningGitClient(), nil)
	})
}

func TestInit_StartsScannerAndReturnsCmd(t *testing.T) {
	t.Parallel()

	gitClient := newFakeScanningGitClient()
	repoScanner := newFakeRepoScanner("/tmp/a")
	m := NewWithDependencies(gitClient, repoScanner)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected non-nil init cmd")
	}

	select {
	case <-repoScanner.done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected scanner.Scan to be called")
	}
}

func TestUpdate_RepoFoundMsg_AppendsBuiltRepository(t *testing.T) {
	t.Parallel()

	gitClient := newFakeScanningGitClient()
	gitClient.reposByPath["/tmp/project"] = domain.Repository{
		Name:        "project",
		Path:        "/tmp/project",
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}

	repoScanner := newFakeRepoScanner()
	m := NewWithDependencies(gitClient, repoScanner)

	result, cmd := m.Update(repoFoundMsg("/tmp/project"))
	model := result.(*Model)

	if len(model.Repositories) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(model.Repositories))
	}
	if model.Repositories[0].Name != "project" {
		t.Fatalf("repo name = %q, want project", model.Repositories[0].Name)
	}
	if cmd == nil {
		t.Fatal("expected non-nil follow-up wait cmd")
	}
}

func TestUpdate_ScanCompleteMsg_EmitsScanFinishedMsg(t *testing.T) {
	t.Parallel()

	m := NewWithDependencies(newFakeScanningGitClient(), newFakeRepoScanner())
	m.Repositories = append(m.Repositories, domain.Repository{Name: "a", Path: "/tmp/a"})

	_, cmd := m.Update(scanCompleteMsg{})
	if cmd == nil {
		t.Fatal("expected cmd for scan completion")
	}

	msg := cmd()
	finished, ok := msg.(ScanFinishedMsg)
	if !ok {
		t.Fatalf("expected ScanFinishedMsg, got %T", msg)
	}
	if len(finished.Repos) != 1 {
		t.Fatalf("expected 1 repo in completion msg, got %d", len(finished.Repos))
	}
}
