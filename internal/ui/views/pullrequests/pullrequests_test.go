package pullrequests

import (
	"strings"
	"testing"
	"time"

	"fresh/internal/domain"

	tea "charm.land/bubbletea/v2"
)

func TestUpdate_PullRequestsLoadedMsg_UpdatesRowsAndStopsLoading(t *testing.T) {
	t.Parallel()

	repo := domain.Repository{Name: "demo", Path: "/tmp/demo", RemoteURL: "https://github.com/octo/demo"}
	m := New(repo, nil)
	m.Loading = true

	rows := []domain.PullRequestDetails{
		{
			Number:    42,
			Title:     "Fix flaky test",
			IsMine:    true,
			UpdatedAt: time.Now(),
			Checks: domain.PullRequestChecks{
				Total:   7,
				Passed:  5,
				Failed:  1,
				Skipped: 1,
			},
		},
	}

	newM, cmd := m.Update(PullRequestsLoadedMsg{RepoPath: repo.Path, PullRequests: rows})
	if cmd != nil {
		t.Fatalf("expected nil cmd, got non-nil")
	}
	if newM.Loading {
		t.Fatalf("expected loading=false after loaded message")
	}
	if len(newM.PullRequests) != 1 {
		t.Fatalf("rows = %d, want 1", len(newM.PullRequests))
	}
	if newM.PullRequests[0].Number != 42 {
		t.Fatalf("row number = %d, want 42", newM.PullRequests[0].Number)
	}
}

func TestUpdate_EscapeReturnsBackCommand(t *testing.T) {
	t.Parallel()

	repo := domain.Repository{Name: "demo", Path: "/tmp/demo", RemoteURL: "https://github.com/octo/demo"}
	m := New(repo, nil)

	_, cmd := m.Update(tea.KeyPressMsg{Code: 27})
	if cmd == nil {
		t.Fatal("expected back command on escape")
	}

	msg := cmd()
	if _, ok := msg.(BackToRepoListMsg); !ok {
		t.Fatalf("command message = %T, want BackToRepoListMsg", msg)
	}
}

func TestBuildCheckSummary_CompleteAndFailing(t *testing.T) {
	t.Parallel()

	summary := buildCheckSummary(domain.PullRequestChecks{
		Total:   7,
		Passed:  5,
		Failed:  1,
		Skipped: 1,
	}, 80)

	if !strings.Contains(summary, "6/7 complete") {
		t.Fatalf("summary = %q, want complete count", summary)
	}
	if !strings.Contains(summary, "1 failing") {
		t.Fatalf("summary = %q, want failing count", summary)
	}
}
