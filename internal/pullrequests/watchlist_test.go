package pullrequests

import "testing"

func TestWatchlist_SeedInitialSnapshotWithoutAlerts(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	if len(changes) != 0 {
		t.Fatalf("changes = %d, want 0", len(changes))
	}
}

func TestWatchlist_NoAlertWhenBlockedStateUnchanged(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{})

	if len(changes) != 0 {
		t.Fatalf("changes = %d, want 0", len(changes))
	}
}

func TestWatchlist_EmitsBlockedTransition(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameBlocked {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameBlocked)
	}
}

func TestWatchlist_EmitsUnblockedTransition(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusChecks},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameUnblocked {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameUnblocked)
	}
}

func TestWatchlist_EmitsMergeableTransition(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameMergeable)
	}
}

func TestWatchlist_BlockedToReadyEmitsMergeableOnly(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameMergeable)
	}
}

func TestWatchlist_EmitsBlockedRemovedForTerminalBlockedPullRequest(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBlockedRemoved {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBlockedRemoved)
	}
	if changes[0].Key.Number != 12 {
		t.Fatalf("number = %d, want 12", changes[0].Key.Number)
	}
}

func TestWatchlist_DoesNotEmitRemovalForNonBlockedPullRequest(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply(nil, ApplyOptions{})
	if len(changes) != 0 {
		t.Fatalf("changes = %d, want 0", len(changes))
	}
}

func TestWatchlist_NewBlockedAfterSeedEmitsAlert(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameBlocked {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameBlocked)
	}
	if changes[0].Key.Number != 12 {
		t.Fatalf("number = %d, want 12", changes[0].Key.Number)
	}
}

func TestWatchlist_NewReadyAfterSeedEmitsMergeableAlert(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "web", Number: 44}, Status: StatusReview},
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameMergeable)
	}
	if changes[0].Key.Number != 12 {
		t.Fatalf("number = %d, want 12", changes[0].Key.Number)
	}
}

func TestWatchlist_FirstApplyWithoutSeedStillAlertsBlocked(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: false})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameBlocked {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameBlocked)
	}
}

func TestWatchlist_FirstApplyWithoutSeedAlertsMergeable(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{Seed: false})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameMergeable)
	}
}

func TestWatchlist_IgnoresInvalidKeys(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	changes := watchlist.Apply([]Snapshot{
		{Key: Key{Owner: "", Repo: "api", Number: 12}, Status: StatusBlocked},
		{Key: Key{Owner: "acme", Repo: "", Number: 12}, Status: StatusBlocked},
		{Key: Key{Owner: "acme", Repo: "api", Number: 0}, Status: StatusBlocked},
	}, ApplyOptions{Seed: false})

	if len(changes) != 0 {
		t.Fatalf("changes = %d, want 0", len(changes))
	}
}

func TestWatchlist_ChangeIncludesPullRequestTitle(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{
			Key:    Key{Owner: "acme", Repo: "api", Number: 12},
			Status: StatusReview,
			Title:  "Improve API docs",
		},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{
			Key:    Key{Owner: "acme", Repo: "api", Number: 12},
			Status: StatusBlocked,
			Title:  "Improve API docs",
		},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Title != "Improve API docs" {
		t.Fatalf("title = %q, want %q", changes[0].Title, "Improve API docs")
	}
}

func TestWatchlist_ChangeRetainsPreviousTitleWhenCurrentSnapshotOmitsIt(t *testing.T) {
	t.Parallel()

	watchlist := NewWatchlist()
	watchlist.Apply([]Snapshot{
		{
			Key:    Key{Owner: "acme", Repo: "api", Number: 12},
			Status: StatusBlocked,
			Title:  "Fix flaky checks",
		},
	}, ApplyOptions{Seed: true})

	changes := watchlist.Apply([]Snapshot{
		{
			Key:    Key{Owner: "acme", Repo: "api", Number: 12},
			Status: StatusChecks,
		},
	}, ApplyOptions{})

	if len(changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(changes))
	}
	if changes[0].Kind != ChangeBecameUnblocked {
		t.Fatalf("kind = %q, want %q", changes[0].Kind, ChangeBecameUnblocked)
	}
	if changes[0].Title != "Fix flaky checks" {
		t.Fatalf("title = %q, want %q", changes[0].Title, "Fix flaky checks")
	}
}
