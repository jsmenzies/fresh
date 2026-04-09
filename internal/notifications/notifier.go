package notifications

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Kind string

const (
	KindProgress Kind = "progress"
	KindBlocked  Kind = "blocked"
)

type PRKey struct {
	Owner  string
	Repo   string
	Number int
}

func (k PRKey) String() string {
	return fmt.Sprintf("%s/%s#%d", k.Owner, k.Repo, k.Number)
}

type Notification struct {
	Key              PRKey
	Kind             Kind
	Reason           string
	PullRequestTitle string
	Repeat           bool
	RepeatEvery      time.Duration
}

type scheduledNotification struct {
	notification Notification
	nextDue      time.Time
}

type Notifier struct {
	mu      sync.Mutex
	entries map[string]scheduledNotification
	stopCh  chan struct{}
	doneCh  chan struct{}
	running bool
}

const (
	glassSoundPath = "/System/Library/Sounds/Glass.aiff"
	blowSoundPath  = "/System/Library/Sounds/Blow.aiff"
)

func NewNotifier() *Notifier {
	return &Notifier{
		entries: make(map[string]scheduledNotification),
	}
}

func (n *Notifier) Start() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return
	}

	n.stopCh = make(chan struct{})
	n.doneCh = make(chan struct{})
	n.running = true

	go n.loop()
}

func (n *Notifier) Stop() {
	n.mu.Lock()
	if !n.running {
		n.mu.Unlock()
		return
	}
	stopCh := n.stopCh
	doneCh := n.doneCh
	n.running = false
	n.stopCh = nil
	n.doneCh = nil
	n.mu.Unlock()

	close(stopCh)
	<-doneCh
}

func (n *Notifier) Upsert(notification Notification) {
	if notification.Key.Owner == "" || notification.Key.Repo == "" || notification.Key.Number <= 0 {
		return
	}

	if notification.Repeat && notification.RepeatEvery <= 0 {
		notification.RepeatEvery = 10 * time.Second
	}

	n.send(notification)

	if !notification.Repeat {
		n.Resolve(notification.Key)
		return
	}

	entry := scheduledNotification{
		notification: notification,
		nextDue:      time.Now().Add(notification.RepeatEvery),
	}

	n.mu.Lock()
	n.entries[notification.Key.String()] = entry
	n.mu.Unlock()
}

func (n *Notifier) Resolve(key PRKey) {
	n.mu.Lock()
	delete(n.entries, key.String())
	n.mu.Unlock()
}

func (n *Notifier) loop() {
	defer close(n.doneCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			n.sendDue()
		}
	}
}

func (n *Notifier) sendDue() {
	now := time.Now()

	n.mu.Lock()
	due := make([]scheduledNotification, 0)
	for key, entry := range n.entries {
		if now.Before(entry.nextDue) {
			continue
		}
		due = append(due, entry)
		entry.nextDue = now.Add(entry.notification.RepeatEvery)
		n.entries[key] = entry
	}
	n.mu.Unlock()

	for _, entry := range due {
		n.send(entry.notification)
	}
}

func (n *Notifier) send(notification Notification) {
	title, body, soundPath := buildPayload(notification)

	cmd := exec.Command(
		"osascript",
		"-e",
		fmt.Sprintf(`display notification "%s" with title "%s"`, escapeAppleScriptString(body), escapeAppleScriptString(title)),
	)

	if err := cmd.Run(); err != nil {
		fmt.Print("\a")
	}

	if err := exec.Command("afplay", soundPath).Start(); err != nil {
		fmt.Print("\a")
	}
}

func buildPayload(notification Notification) (title, body, soundPath string) {
	title = notification.Key.Repo
	if prTitle := strings.TrimSpace(notification.PullRequestTitle); prTitle != "" {
		title = fmt.Sprintf("%s: %s", notification.Key.Repo, prTitle)
	}

	switch notification.Kind {
	case KindBlocked:
		body = fmt.Sprintf("Blocked - %s", notification.Reason)
		soundPath = blowSoundPath
	default:
		body = fmt.Sprintf("Progress - %s", notification.Reason)
		soundPath = glassSoundPath
	}

	if strings.TrimSpace(notification.Reason) == "" {
		if notification.Kind == KindBlocked {
			body = "Blocked"
		} else {
			body = "Progress"
		}
	}

	return title, body, soundPath
}

func escapeAppleScriptString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	return strings.ReplaceAll(value, "\"", "\\\"")
}
