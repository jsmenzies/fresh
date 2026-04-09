package pullrequests

import (
	"fmt"
	"sort"
)

type Status string

const (
	StatusReady   Status = "ready"
	StatusBlocked Status = "blocked"
	StatusReview  Status = "review"
	StatusChecks  Status = "checks"
)

type Key struct {
	Owner  string
	Repo   string
	Number int
}

func (k Key) String() string {
	return fmt.Sprintf("%s/%s#%d", k.Owner, k.Repo, k.Number)
}

func (k Key) IsValid() bool {
	return k.Owner != "" && k.Repo != "" && k.Number > 0
}

type Snapshot struct {
	Key    Key
	Status Status
}

type ChangeKind string

const (
	ChangeBecameBlocked   ChangeKind = "became_blocked"
	ChangeBecameUnblocked ChangeKind = "became_unblocked"
	ChangeBlockedRemoved  ChangeKind = "blocked_removed"
)

type Change struct {
	Kind     ChangeKind
	Key      Key
	Previous Status
	Current  Status
}

type ApplyOptions struct {
	// Seed suppresses change events the first time Apply is called, so the
	// initial sync becomes a baseline instead of generating alerts.
	Seed bool
}

type Watchlist struct {
	seeded  bool
	tracked map[Key]Status
}

func NewWatchlist() *Watchlist {
	return &Watchlist{
		tracked: make(map[Key]Status),
	}
}

func (w *Watchlist) Apply(current []Snapshot, options ApplyOptions) []Change {
	if w == nil {
		return nil
	}
	if w.tracked == nil {
		w.tracked = make(map[Key]Status)
	}

	currentByKey := make(map[Key]Status, len(current))
	for _, snapshot := range current {
		if !snapshot.Key.IsValid() {
			continue
		}
		currentByKey[snapshot.Key] = snapshot.Status
	}

	if !w.seeded && options.Seed {
		w.tracked = currentByKey
		w.seeded = true
		return nil
	}

	emitNewBlocked := w.seeded || !options.Seed
	changes := make([]Change, 0)

	for key, currentStatus := range currentByKey {
		previousStatus, existed := w.tracked[key]
		switch {
		case !existed:
			if emitNewBlocked && currentStatus == StatusBlocked {
				changes = append(changes, Change{
					Kind:    ChangeBecameBlocked,
					Key:     key,
					Current: currentStatus,
				})
			}
		case previousStatus != currentStatus:
			switch {
			case currentStatus == StatusBlocked:
				changes = append(changes, Change{
					Kind:     ChangeBecameBlocked,
					Key:      key,
					Previous: previousStatus,
					Current:  currentStatus,
				})
			case previousStatus == StatusBlocked:
				changes = append(changes, Change{
					Kind:     ChangeBecameUnblocked,
					Key:      key,
					Previous: previousStatus,
					Current:  currentStatus,
				})
			}
		}
	}

	for key, previousStatus := range w.tracked {
		if _, ok := currentByKey[key]; ok {
			continue
		}
		if previousStatus != StatusBlocked {
			continue
		}
		changes = append(changes, Change{
			Kind:     ChangeBlockedRemoved,
			Key:      key,
			Previous: previousStatus,
		})
	}

	w.tracked = currentByKey
	w.seeded = true

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Key == changes[j].Key {
			return changes[i].Kind < changes[j].Kind
		}
		return keyLess(changes[i].Key, changes[j].Key)
	})

	return changes
}

func keyLess(left, right Key) bool {
	if left.Owner != right.Owner {
		return left.Owner < right.Owner
	}
	if left.Repo != right.Repo {
		return left.Repo < right.Repo
	}
	return left.Number < right.Number
}
