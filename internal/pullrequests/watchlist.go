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
	ChangeRemoved         ChangeKind = "removed"
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
	tracked map[string]Snapshot
}

func NewWatchlist() *Watchlist {
	return &Watchlist{
		tracked: make(map[string]Snapshot),
	}
}

func (w *Watchlist) Apply(current []Snapshot, options ApplyOptions) []Change {
	if w == nil {
		return nil
	}
	if w.tracked == nil {
		w.tracked = make(map[string]Snapshot)
	}

	currentByKey := make(map[string]Snapshot, len(current))
	for _, snapshot := range current {
		if !snapshot.Key.IsValid() {
			continue
		}
		currentByKey[snapshot.Key.String()] = snapshot
	}

	if !w.seeded && options.Seed {
		w.tracked = currentByKey
		w.seeded = true
		return nil
	}

	emitNewBlocked := w.seeded || !options.Seed
	changes := make([]Change, 0)

	for key, snapshot := range currentByKey {
		previous, existed := w.tracked[key]
		switch {
		case !existed:
			if emitNewBlocked && snapshot.Status == StatusBlocked {
				changes = append(changes, Change{
					Kind:    ChangeBecameBlocked,
					Key:     snapshot.Key,
					Current: snapshot.Status,
				})
			}
		case previous.Status != snapshot.Status:
			switch {
			case snapshot.Status == StatusBlocked:
				changes = append(changes, Change{
					Kind:     ChangeBecameBlocked,
					Key:      snapshot.Key,
					Previous: previous.Status,
					Current:  snapshot.Status,
				})
			case previous.Status == StatusBlocked:
				changes = append(changes, Change{
					Kind:     ChangeBecameUnblocked,
					Key:      snapshot.Key,
					Previous: previous.Status,
					Current:  snapshot.Status,
				})
			}
		}
	}

	for key, previous := range w.tracked {
		if _, ok := currentByKey[key]; ok {
			continue
		}
		changes = append(changes, Change{
			Kind:     ChangeRemoved,
			Key:      previous.Key,
			Previous: previous.Status,
		})
	}

	w.tracked = currentByKey
	w.seeded = true

	sort.Slice(changes, func(i, j int) bool {
		left := changes[i].Key.String()
		right := changes[j].Key.String()
		if left == right {
			return changes[i].Kind < changes[j].Kind
		}
		return left < right
	})

	return changes
}
