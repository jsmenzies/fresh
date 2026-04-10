package listing

import (
	"testing"

	"fresh/internal/domain"

	"charm.land/bubbles/v2/spinner"
)

func TestStartPullRequestSyncMarksInFlightAndReturnsCommand(t *testing.T) {
	m := New(nil)
	cmd := m.startPullRequestSync(pullRequestSyncManual)

	if m.PRSyncInFlight != 1 {
		t.Fatalf("PRSyncInFlight = %d, want 1", m.PRSyncInFlight)
	}
	if cmd == nil {
		t.Fatal("expected non-nil sync command")
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

	newM, _ := m.Update(PullRequestStatesUpdatedMsg{
		Generation: m.PRSyncGeneration,
		Trigger:    pullRequestSyncManual,
	})
	if newM == nil {
		t.Fatal("expected model")
	}
	if m.PRSyncInFlight != 0 {
		t.Fatalf("PRSyncInFlight = %d, want 0", m.PRSyncInFlight)
	}
}

func TestUpdatePullRequestStatesIgnoresStaleGeneration(t *testing.T) {
	m := New(sampleWatchRepos())
	_ = m.startPullRequestSync(pullRequestSyncManual)
	staleGeneration := m.PRSyncGeneration
	_ = m.startPullRequestSync(pullRequestSyncManual)
	latestGeneration := m.PRSyncGeneration

	repoPath := m.Repositories[0].Path
	m.Repositories[0].PullRequests = domain.PullRequestCount{Open: 7}

	newM, cmd := m.Update(PullRequestStatesUpdatedMsg{
		Generation: staleGeneration,
		States: map[string]domain.PullRequestState{
			repoPath: domain.PullRequestCount{Open: 99},
		},
		Trigger: pullRequestSyncManual,
	})
	if newM == nil {
		t.Fatal("expected model")
	}
	if cmd != nil {
		t.Fatal("stale manual sync should not schedule work")
	}
	if m.PRSyncInFlight != 1 {
		t.Fatalf("PRSyncInFlight = %d, want 1", m.PRSyncInFlight)
	}

	state, ok := m.Repositories[0].PullRequests.(domain.PullRequestCount)
	if !ok {
		t.Fatalf("pull request state type = %T, want domain.PullRequestCount", m.Repositories[0].PullRequests)
	}
	if state.Open != 7 {
		t.Fatalf("stale sync overwrote state: Open = %d, want 7", state.Open)
	}

	_, _ = m.Update(PullRequestStatesUpdatedMsg{
		Generation: latestGeneration,
		States: map[string]domain.PullRequestState{
			repoPath: domain.PullRequestCount{Open: 5},
		},
		Trigger: pullRequestSyncManual,
	})
	state, ok = m.Repositories[0].PullRequests.(domain.PullRequestCount)
	if !ok {
		t.Fatalf("pull request state type = %T, want domain.PullRequestCount", m.Repositories[0].PullRequests)
	}
	if state.Open != 5 {
		t.Fatalf("latest sync not applied: Open = %d, want 5", state.Open)
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
