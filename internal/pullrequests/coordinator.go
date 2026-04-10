package pullrequests

import (
	"fmt"
)

type NotificationSink interface {
	Upsert(notification Notification)
	Resolve(key Key)
}

type NotificationCoordinator struct {
	watchlist *Watchlist
}

func NewNotificationCoordinator(watchlist *Watchlist) *NotificationCoordinator {
	if watchlist == nil {
		watchlist = NewWatchlist()
	}

	return &NotificationCoordinator{
		watchlist: watchlist,
	}
}

func (c *NotificationCoordinator) Sync(tracked []Snapshot, options ApplyOptions, sink NotificationSink) []Change {
	if c == nil || c.watchlist == nil {
		return nil
	}

	changes := c.watchlist.Apply(tracked, options)
	if sink == nil || len(changes) == 0 {
		return changes
	}

	for _, change := range changes {
		switch change.Kind {
		case ChangeBecameBlocked:
			sink.Upsert(Notification{
				Key:              change.Key,
				Kind:             NotificationKindBlocked,
				Reason:           fmt.Sprintf("%s is blocked", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBecameMergeable:
			sink.Upsert(Notification{
				Key:              change.Key,
				Kind:             NotificationKindProgress,
				Reason:           fmt.Sprintf("%s is mergeable", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBecameUnblocked:
			sink.Resolve(change.Key)
			sink.Upsert(Notification{
				Key:              change.Key,
				Kind:             NotificationKindProgress,
				Reason:           fmt.Sprintf("%s is no longer blocked", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBlockedRemoved:
			sink.Resolve(change.Key)
		}
	}

	return changes
}
