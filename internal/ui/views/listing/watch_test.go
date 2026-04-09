package listing

import (
	"strings"
	"testing"
	"time"

	"fresh/internal/domain"
)

func TestToggleWatchModeTurnsOnAndSchedulesTick(t *testing.T) {
	m := New(nil)
	var gotInterval time.Duration
	withImmediateTickMock(t, func(interval time.Duration) {
		gotInterval = interval
	})
	if m.WatchEnabled {
		t.Fatal("watch mode should start disabled")
	}

	cmd := m.toggleWatchMode()
	if !m.WatchEnabled {
		t.Fatal("watch mode should be enabled")
	}
	if cmd == nil {
		t.Fatal("expected watch tick command")
	}

	msg := cmd()
	if _, ok := msg.(watchTickMsg); !ok {
		t.Fatalf("command message = %T, want watchTickMsg", msg)
	}
	if gotInterval != m.WatchEvery {
		t.Fatalf("scheduled interval = %s, want %s", gotInterval, m.WatchEvery)
	}
}

func TestToggleWatchModeTurnsOffAndCancelsPendingTicks(t *testing.T) {
	m := New(nil)
	_ = m.toggleWatchMode()
	tokenOn := m.WatchToken

	cmd := m.toggleWatchMode()
	if m.WatchEnabled {
		t.Fatal("watch mode should be disabled")
	}
	if cmd != nil {
		t.Fatal("disabling watch should not schedule new command")
	}
	if m.WatchToken <= tokenOn {
		t.Fatal("watch token should advance to invalidate pending ticks")
	}
}

func TestCurrentWatchIntervalBackoffDoublesToMax(t *testing.T) {
	m := New(nil)
	m.WatchEvery = time.Minute
	m.WatchMaxEvery = 8 * time.Minute

	intervals := []time.Duration{
		time.Minute,
		2 * time.Minute,
		4 * time.Minute,
		8 * time.Minute,
		8 * time.Minute,
	}

	for i, want := range intervals {
		m.WatchBackoff = i
		if got := m.currentWatchInterval(); got != want {
			t.Fatalf("backoff %d interval = %s, want %s", i, got, want)
		}
	}
}

func TestHasPullRequestSyncError(t *testing.T) {
	noError := map[string]domain.PullRequestState{
		"/repo/a": domain.PullRequestCount{Open: 1},
		"/repo/b": domain.PullRequestUnavailable{},
	}
	if hasPullRequestSyncError(noError) {
		t.Fatal("expected false when no PullRequestError present")
	}

	withError := map[string]domain.PullRequestState{
		"/repo/a": domain.PullRequestCount{Open: 1},
		"/repo/b": domain.PullRequestError{Message: "rate limited"},
	}
	if !hasPullRequestSyncError(withError) {
		t.Fatal("expected true when PullRequestError is present")
	}
}

func TestUpdateWatchTickIgnoresStaleToken(t *testing.T) {
	m := New(nil)
	_ = m.toggleWatchMode()
	currentToken := m.WatchToken

	newM, cmd := m.Update(watchTickMsg{Token: currentToken + 1})
	if newM == nil {
		t.Fatal("expected model")
	}
	if cmd != nil {
		t.Fatal("stale tick should not schedule work")
	}
}

func TestUpdateWatchTickStartsWatchRefreshCycle(t *testing.T) {
	m := New([]domain.Repository{
		{
			Name:        "demo",
			Path:        "/tmp/demo",
			RemoteURL:   "https://github.com/org/demo.git",
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			RemoteState: domain.Synced{},
			LocalState:  domain.CleanLocalState{},
			Activity:    &domain.IdleActivity{},
		},
	})
	_ = m.toggleWatchMode()

	newM, cmd := m.Update(watchTickMsg{Token: m.WatchToken})
	if newM == nil {
		t.Fatal("expected model")
	}
	if cmd == nil {
		t.Fatal("expected watch refresh command batch")
	}

	if _, ok := newM.Repositories[0].Activity.(*domain.RefreshingActivity); !ok {
		t.Fatalf("repo activity = %T, want *domain.RefreshingActivity", newM.Repositories[0].Activity)
	}
	if m.PRSyncInFlight != 1 {
		t.Fatalf("PRSyncInFlight = %d, want 1", m.PRSyncInFlight)
	}
}

func TestBuildFooterShowsWatchStatus(t *testing.T) {
	m := New(nil)
	footer := m.buildFooter()

	if !strings.Contains(footer, "w watch off") {
		t.Fatalf("footer = %q, want watch off label", footer)
	}

	_ = m.toggleWatchMode()
	footer = m.buildFooter()
	if !strings.Contains(footer, "w watch on") {
		t.Fatalf("footer = %q, want watch on label", footer)
	}
}

func TestPullRequestWatchErrorBackoffSchedulesLongerInterval(t *testing.T) {
	m := New(nil)
	_ = m.toggleWatchMode()

	if got := m.currentWatchInterval(); got != time.Minute {
		t.Fatalf("currentWatchInterval() = %s, want %s", got, time.Minute)
	}

	newM, cmd := m.Update(PullRequestStatesUpdatedMsg{
		States: map[string]domain.PullRequestState{
			"/repo/a": domain.PullRequestError{Message: "gh: rate limit"},
		},
		Trigger: pullRequestSyncWatch,
	})

	if newM == nil {
		t.Fatal("expected updated model")
	}
	if cmd == nil {
		t.Fatal("expected watch reschedule command")
	}
	if m.WatchBackoff != 1 {
		t.Fatalf("WatchBackoff = %d, want 1", m.WatchBackoff)
	}
	if got := m.currentWatchInterval(); got != 2*time.Minute {
		t.Fatalf("currentWatchInterval() = %s, want %s", got, 2*time.Minute)
	}
}

func TestPullRequestWatchSuccessResetsBackoff(t *testing.T) {
	m := New(nil)
	_ = m.toggleWatchMode()
	m.WatchBackoff = 3

	newM, cmd := m.Update(PullRequestStatesUpdatedMsg{
		States: map[string]domain.PullRequestState{
			"/repo/a": domain.PullRequestCount{Open: 1},
		},
		Trigger: pullRequestSyncWatch,
	})

	if newM == nil {
		t.Fatal("expected updated model")
	}
	if cmd == nil {
		t.Fatal("expected watch reschedule command")
	}
	if m.WatchBackoff != 0 {
		t.Fatalf("WatchBackoff = %d, want 0", m.WatchBackoff)
	}
	if got := m.currentWatchInterval(); got != time.Minute {
		t.Fatalf("currentWatchInterval() = %s, want %s", got, time.Minute)
	}
}

func TestStaleWatchSyncStillReschedulesTick(t *testing.T) {
	m := New(nil)
	_ = m.toggleWatchMode()
	m.WatchBackoff = 2
	m.PRSyncGeneration = 3

	newM, cmd := m.Update(PullRequestStatesUpdatedMsg{
		Generation: 2, // stale
		States: map[string]domain.PullRequestState{
			"/repo/a": domain.PullRequestError{Message: "gh: timeout"},
		},
		Trigger: pullRequestSyncWatch,
	})

	if newM == nil {
		t.Fatal("expected updated model")
	}
	if cmd == nil {
		t.Fatal("expected watch reschedule command from stale watch response")
	}
	// Stale responses should not modify backoff state.
	if m.WatchBackoff != 2 {
		t.Fatalf("WatchBackoff = %d, want 2", m.WatchBackoff)
	}
}
