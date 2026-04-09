package notifications

import (
	"testing"
	"time"
)

func TestNotifierSendDueOnlyProcessesDueEntries(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	notifier, _, sent := newTestNotifier(base)

	first := Notification{
		Key: PRKey{
			Owner:  "octo",
			Repo:   "fresh",
			Number: 1,
		},
		Repeat:      true,
		RepeatEvery: 5 * time.Second,
	}
	second := Notification{
		Key: PRKey{
			Owner:  "octo",
			Repo:   "fresh",
			Number: 2,
		},
		Repeat:      true,
		RepeatEvery: 10 * time.Second,
	}

	notifier.Upsert(first)
	notifier.Upsert(second)

	if got := len(*sent); got != 2 {
		t.Fatalf("immediate sends = %d, want 2", got)
	}
	*sent = (*sent)[:0]

	notifier.sendDue(base.Add(6 * time.Second))

	if got := len(*sent); got != 1 {
		t.Fatalf("due sends = %d, want 1", got)
	}
	if (*sent)[0].Key.Number != 1 {
		t.Fatalf("sent key = %d, want 1", (*sent)[0].Key.Number)
	}

	notifier.mu.Lock()
	firstDue := notifier.entries[first.Key.String()].nextDue
	secondDue := notifier.entries[second.Key.String()].nextDue
	head := notifier.dueHeap[0].notification.Key.Number
	notifier.mu.Unlock()

	if !firstDue.Equal(base.Add(11 * time.Second)) {
		t.Fatalf("first next due = %s, want %s", firstDue, base.Add(11*time.Second))
	}
	if !secondDue.Equal(base.Add(10 * time.Second)) {
		t.Fatalf("second next due = %s, want %s", secondDue, base.Add(10*time.Second))
	}
	if head != 2 {
		t.Fatalf("heap head = %d, want 2", head)
	}
}

func TestNotifierResolveRemovesScheduledEntry(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	notifier, _, sent := newTestNotifier(base)

	notification := Notification{
		Key: PRKey{
			Owner:  "octo",
			Repo:   "fresh",
			Number: 7,
		},
		Repeat:      true,
		RepeatEvery: 5 * time.Second,
	}

	notifier.Upsert(notification)
	*sent = (*sent)[:0]

	notifier.Resolve(notification.Key)
	notifier.sendDue(base.Add(1 * time.Hour))

	if got := len(*sent); got != 0 {
		t.Fatalf("due sends after resolve = %d, want 0", got)
	}

	notifier.mu.Lock()
	entryCount := len(notifier.entries)
	heapCount := len(notifier.dueHeap)
	notifier.mu.Unlock()

	if entryCount != 0 {
		t.Fatalf("entries = %d, want 0", entryCount)
	}
	if heapCount != 0 {
		t.Fatalf("heap entries = %d, want 0", heapCount)
	}
}

func TestNotifierUpsertUpdatesExistingSchedule(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	notifier, now, sent := newTestNotifier(base)

	key := PRKey{
		Owner:  "octo",
		Repo:   "fresh",
		Number: 9,
	}

	notifier.Upsert(Notification{
		Key:         key,
		Kind:        KindProgress,
		Repeat:      true,
		RepeatEvery: 10 * time.Second,
	})
	*sent = (*sent)[:0]

	*now = base.Add(2 * time.Second)
	notifier.Upsert(Notification{
		Key:         key,
		Kind:        KindBlocked,
		Repeat:      true,
		RepeatEvery: 20 * time.Second,
	})

	if got := len(*sent); got != 1 {
		t.Fatalf("immediate sends after update = %d, want 1", got)
	}
	if (*sent)[0].Kind != KindBlocked {
		t.Fatalf("updated kind = %s, want %s", (*sent)[0].Kind, KindBlocked)
	}
	*sent = (*sent)[:0]

	notifier.sendDue(base.Add(11 * time.Second))
	if got := len(*sent); got != 0 {
		t.Fatalf("due sends before updated schedule = %d, want 0", got)
	}

	notifier.sendDue(base.Add(22 * time.Second))
	if got := len(*sent); got != 1 {
		t.Fatalf("due sends at updated schedule = %d, want 1", got)
	}
	if (*sent)[0].Kind != KindBlocked {
		t.Fatalf("sent kind = %s, want %s", (*sent)[0].Kind, KindBlocked)
	}

	notifier.mu.Lock()
	entryCount := len(notifier.entries)
	heapCount := len(notifier.dueHeap)
	notifier.mu.Unlock()

	if entryCount != 1 {
		t.Fatalf("entries = %d, want 1", entryCount)
	}
	if heapCount != 1 {
		t.Fatalf("heap entries = %d, want 1", heapCount)
	}
}

func newTestNotifier(base time.Time) (*Notifier, *time.Time, *[]Notification) {
	now := base
	sent := make([]Notification, 0)

	notifier := NewNotifier()
	notifier.now = func() time.Time {
		return now
	}
	notifier.sendFn = func(notification Notification) {
		sent = append(sent, notification)
	}

	return notifier, &now, &sent
}
