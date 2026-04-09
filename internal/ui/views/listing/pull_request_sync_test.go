package listing

import (
	"testing"

	"fresh/internal/domain"

	"charm.land/bubbles/v2/spinner"
)

func TestStartPullRequestSyncMarksInFlightAndReturnsCommands(t *testing.T) {
	m := New(nil)
	cmds := m.startPullRequestSync(pullRequestSyncManual)

	if m.PRSyncInFlight != 1 {
		t.Fatalf("PRSyncInFlight = %d, want 1", m.PRSyncInFlight)
	}
	if len(cmds) != 2 {
		t.Fatalf("len(cmds) = %d, want 2", len(cmds))
	}
	if cmds[0] == nil || cmds[1] == nil {
		t.Fatal("expected non-nil sync commands")
	}
}

func TestCompletePullRequestSyncDoesNotGoNegative(t *testing.T) {
	m := New(nil)
	m.completePullRequestSync()

	if m.PRSyncInFlight != 0 {
		t.Fatalf("PRSyncInFlight = %d, want 0", m.PRSyncInFlight)
	}

	m.PRSyncInFlight = 1
	m.completePullRequestSync()
	if m.PRSyncInFlight != 0 {
		t.Fatalf("PRSyncInFlight = %d, want 0", m.PRSyncInFlight)
	}
}

func TestUpdateSpinnerTickAdvancesPRSpinnerWhenSyncing(t *testing.T) {
	m := New(nil)
	_ = m.startPullRequestSync(pullRequestSyncManual)

	newM, cmd := m.Update(spinner.TickMsg{})
	if newM == nil {
		t.Fatal("expected model")
	}
	if cmd == nil {
		t.Fatal("expected follow-up spinner tick command")
	}
}

func TestUpdatePullRequestStatesCompletesInFlightSync(t *testing.T) {
	m := New(nil)
	_ = m.startPullRequestSync(pullRequestSyncManual)

	newM, _ := m.Update(PullRequestStatesUpdatedMsg{Trigger: pullRequestSyncManual})
	if newM == nil {
		t.Fatal("expected model")
	}
	if m.PRSyncInFlight != 0 {
		t.Fatalf("PRSyncInFlight = %d, want 0", m.PRSyncInFlight)
	}
}

func TestPullRequestSpinnerViewEmptyWhenNotSyncing(t *testing.T) {
	m := New(nil)
	if got := m.pullRequestSpinnerView(); got != "" {
		t.Fatalf("pullRequestSpinnerView() = %q, want empty", got)
	}
}

func TestStartRefreshCycleStartsPRSyncAndRepoRefresh(t *testing.T) {
	m := New(sampleWatchRepos())

	cmd := m.startRefreshCycle(pullRequestSyncManual)
	if cmd == nil {
		t.Fatal("expected batch command")
	}
	if m.PRSyncInFlight != 1 {
		t.Fatalf("PRSyncInFlight = %d, want 1", m.PRSyncInFlight)
	}

	if _, ok := m.Repositories[0].Activity.(*domain.RefreshingActivity); !ok {
		t.Fatalf("repo activity = %T, want *domain.RefreshingActivity", m.Repositories[0].Activity)
	}
}

func sampleWatchRepos() []domain.Repository {
	return []domain.Repository{
		{
			Name:        "demo",
			Path:        "/tmp/demo",
			RemoteURL:   "https://github.com/org/demo.git",
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			Activity:    &domain.IdleActivity{},
		},
	}
}
