package pullrequests

import "time"

type NotificationKind string

const (
	NotificationKindProgress NotificationKind = "progress"
	NotificationKindBlocked  NotificationKind = "blocked"
)

type Notification struct {
	Key              Key
	Kind             NotificationKind
	Reason           string
	PullRequestTitle string
	Repeat           bool
	RepeatEvery      time.Duration
}
