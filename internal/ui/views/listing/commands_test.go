package listing

import (
	"testing"

	"fresh/internal/domain"
)

func TestPerformRefresh_FetchBeforeBuildRepository(t *testing.T) {
	t.Parallel()

	client := newFakeGitClient()
	client.reposByPath["/tmp/repo"] = domain.Repository{
		Name:        "repo",
		Path:        "/tmp/repo",
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
	}

	m := New([]domain.Repository{}, client)

	cmd := performRefresh(m, 3, "/tmp/repo")
	msg := cmd()

	if len(client.callLog) != 2 {
		t.Fatalf("expected 2 client calls, got %d (%v)", len(client.callLog), client.callLog)
	}
	if client.callLog[0] != "Fetch:/tmp/repo" {
		t.Fatalf("first call = %q, want Fetch first", client.callLog[0])
	}
	if client.callLog[1] != "BuildRepository:/tmp/repo" {
		t.Fatalf("second call = %q, want BuildRepository second", client.callLog[1])
	}

	updated, ok := msg.(RepoUpdatedMsg)
	if !ok {
		t.Fatalf("expected RepoUpdatedMsg, got %T", msg)
	}
	if updated.Index != 3 {
		t.Fatalf("index = %d, want 3", updated.Index)
	}
	if updated.Repo.Path != "/tmp/repo" {
		t.Fatalf("repo path = %q, want /tmp/repo", updated.Repo.Path)
	}
}
