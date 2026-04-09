package pullrequests

import (
	"fmt"
	"fresh/internal/notifications"
)

type NotificationSink interface {
	Upsert(notification notifications.Notification)
	Resolve(key notifications.PRKey)
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
		key := notifications.PRKey{
			Owner:  change.Key.Owner,
			Repo:   change.Key.Repo,
			Number: change.Key.Number,
		}

		switch change.Kind {
		case ChangeBecameBlocked:
			sink.Upsert(notifications.Notification{
				Key:              key,
				Kind:             notifications.KindBlocked,
				Reason:           fmt.Sprintf("%s is blocked", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBecameMergeable:
			sink.Upsert(notifications.Notification{
				Key:              key,
				Kind:             notifications.KindProgress,
				Reason:           fmt.Sprintf("%s is mergeable", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBecameUnblocked:
			sink.Resolve(key)
			sink.Upsert(notifications.Notification{
				Key:              key,
				Kind:             notifications.KindProgress,
				Reason:           fmt.Sprintf("%s is no longer blocked", change.Key.String()),
				PullRequestTitle: change.Title,
			})
		case ChangeBlockedRemoved:
			sink.Resolve(key)
		}
	}

	return changes
}
