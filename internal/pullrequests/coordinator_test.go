package pullrequests

import (
	"fresh/internal/notifications"
	"testing"
)

type fakeNotificationSink struct {
	upserts  []notifications.Notification
	resolves []notifications.PRKey
}

func (f *fakeNotificationSink) Upsert(notification notifications.Notification) {
	f.upserts = append(f.upserts, notification)
}

func (f *fakeNotificationSink) Resolve(key notifications.PRKey) {
	f.resolves = append(f.resolves, key)
}

func TestNotificationCoordinator_SeedSuppressesNotifications(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	sink := &fakeNotificationSink{}

	changes := coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true}, sink)

	if len(changes) != 0 {
		t.Fatalf("changes = %d, want 0", len(changes))
	}
	if len(sink.upserts) != 0 || len(sink.resolves) != 0 {
		t.Fatalf("notifications sent on seed: upserts=%d resolves=%d", len(sink.upserts), len(sink.resolves))
	}
}

func TestNotificationCoordinator_BlockedTransitionUpsertsBlocked(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	_ = coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReview},
	}, ApplyOptions{Seed: true}, nil)

	sink := &fakeNotificationSink{}
	changes := coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{}, sink)

	if len(changes) != 1 || changes[0].Kind != ChangeBecameBlocked {
		t.Fatalf("changes = %+v, want one %q", changes, ChangeBecameBlocked)
	}
	if len(sink.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(sink.upserts))
	}
	if sink.upserts[0].Kind != notifications.KindBlocked {
		t.Fatalf("kind = %q, want %q", sink.upserts[0].Kind, notifications.KindBlocked)
	}
}

func TestNotificationCoordinator_UnblockedTransitionResolvesThenProgress(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	_ = coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true}, nil)

	sink := &fakeNotificationSink{}
	changes := coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusChecks},
	}, ApplyOptions{}, sink)

	if len(changes) != 1 || changes[0].Kind != ChangeBecameUnblocked {
		t.Fatalf("changes = %+v, want one %q", changes, ChangeBecameUnblocked)
	}
	if len(sink.resolves) != 1 {
		t.Fatalf("resolves = %d, want 1", len(sink.resolves))
	}
	if len(sink.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(sink.upserts))
	}
	if sink.upserts[0].Kind != notifications.KindProgress {
		t.Fatalf("kind = %q, want %q", sink.upserts[0].Kind, notifications.KindProgress)
	}
}

func TestNotificationCoordinator_MergeableTransitionUpsertsProgress(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	_ = coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReview},
	}, ApplyOptions{Seed: true}, nil)

	sink := &fakeNotificationSink{}
	changes := coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{}, sink)

	if len(changes) != 1 || changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("changes = %+v, want one %q", changes, ChangeBecameMergeable)
	}
	if len(sink.resolves) != 0 {
		t.Fatalf("resolves = %d, want 0", len(sink.resolves))
	}
	if len(sink.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(sink.upserts))
	}
	if sink.upserts[0].Kind != notifications.KindProgress {
		t.Fatalf("kind = %q, want %q", sink.upserts[0].Kind, notifications.KindProgress)
	}
}

func TestNotificationCoordinator_BlockedToReadyDoesNotEmitUnblocked(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	_ = coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true}, nil)

	sink := &fakeNotificationSink{}
	changes := coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusReady},
	}, ApplyOptions{}, sink)

	if len(changes) != 1 || changes[0].Kind != ChangeBecameMergeable {
		t.Fatalf("changes = %+v, want one %q", changes, ChangeBecameMergeable)
	}
	if len(sink.resolves) != 0 {
		t.Fatalf("resolves = %d, want 0", len(sink.resolves))
	}
	if len(sink.upserts) != 1 {
		t.Fatalf("upserts = %d, want 1", len(sink.upserts))
	}
}

func TestNotificationCoordinator_BlockedRemovedOnlyResolves(t *testing.T) {
	t.Parallel()

	coordinator := NewNotificationCoordinator(nil)
	_ = coordinator.Sync([]Snapshot{
		{Key: Key{Owner: "acme", Repo: "api", Number: 12}, Status: StatusBlocked},
	}, ApplyOptions{Seed: true}, nil)

	sink := &fakeNotificationSink{}
	changes := coordinator.Sync(nil, ApplyOptions{}, sink)

	if len(changes) != 1 || changes[0].Kind != ChangeBlockedRemoved {
		t.Fatalf("changes = %+v, want one %q", changes, ChangeBlockedRemoved)
	}
	if len(sink.resolves) != 1 {
		t.Fatalf("resolves = %d, want 1", len(sink.resolves))
	}
	if len(sink.upserts) != 0 {
		t.Fatalf("upserts = %d, want 0", len(sink.upserts))
	}
}
