package notifications

import (
	"container/heap"
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
	Key         PRKey
	Kind        Kind
	Reason      string
	Repeat      bool
	RepeatEvery time.Duration
}

type scheduledNotification struct {
	notification Notification
	nextDue      time.Time
	index        int
}

type notificationHeap []*scheduledNotification

func (h notificationHeap) Len() int {
	return len(h)
}

func (h notificationHeap) Less(i, j int) bool {
	return h[i].nextDue.Before(h[j].nextDue)
}

func (h notificationHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *notificationHeap) Push(value any) {
	entry := value.(*scheduledNotification)
	entry.index = len(*h)
	*h = append(*h, entry)
}

func (h *notificationHeap) Pop() any {
	entries := *h
	last := len(entries) - 1
	entry := entries[last]
	entry.index = -1
	*h = entries[:last]
	return entry
}

type Notifier struct {
	mu       sync.Mutex
	entries  map[string]*scheduledNotification
	dueHeap  notificationHeap
	stopCh   chan struct{}
	wakeupCh chan struct{}
	doneCh   chan struct{}
	running  bool

	now    func() time.Time
	sendFn func(Notification)
}

const (
	glassSoundPath = "/System/Library/Sounds/Glass.aiff"
	blowSoundPath  = "/System/Library/Sounds/Blow.aiff"
)

func NewNotifier() *Notifier {
	notifier := &Notifier{
		entries: make(map[string]*scheduledNotification),
		now:     time.Now,
	}
	notifier.sendFn = notifier.send
	return notifier
}

func (n *Notifier) Start() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return
	}

	n.stopCh = make(chan struct{})
	n.wakeupCh = make(chan struct{}, 1)
	n.doneCh = make(chan struct{})
	n.running = true

	go n.loop(n.stopCh, n.wakeupCh, n.doneCh)
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
	n.wakeupCh = nil
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

	n.sendFn(notification)

	if !notification.Repeat {
		n.Resolve(notification.Key)
		return
	}

	key := notification.Key.String()
	nextDue := n.now().Add(notification.RepeatEvery)
	n.mu.Lock()
	entry, exists := n.entries[key]
	if exists {
		entry.notification = notification
		entry.nextDue = nextDue
		heap.Fix(&n.dueHeap, entry.index)
	} else {
		entry = &scheduledNotification{
			notification: notification,
			nextDue:      nextDue,
			index:        -1,
		}
		heap.Push(&n.dueHeap, entry)
		n.entries[key] = entry
	}
	wakeupCh := n.wakeupCh
	running := n.running
	n.mu.Unlock()

	if running {
		notifyLoop(wakeupCh)
	}
}

func (n *Notifier) Resolve(key PRKey) {
	n.mu.Lock()
	entry, exists := n.entries[key.String()]
	if exists {
		heap.Remove(&n.dueHeap, entry.index)
		delete(n.entries, key.String())
	}
	wakeupCh := n.wakeupCh
	running := n.running
	n.mu.Unlock()

	if running {
		notifyLoop(wakeupCh)
	}
}

func (n *Notifier) loop(stopCh <-chan struct{}, wakeupCh <-chan struct{}, doneCh chan struct{}) {
	defer close(doneCh)

	var timer *time.Timer
	var timerCh <-chan time.Time

	for {
		wait, hasDue := n.nextDueWait()
		if hasDue {
			if timer == nil {
				timer = time.NewTimer(wait)
			} else {
				resetTimer(timer, wait)
			}
			timerCh = timer.C
		} else {
			stopTimer(timer)
			timerCh = nil
		}

		select {
		case <-stopCh:
			stopTimer(timer)
			return
		case <-wakeupCh:
			continue
		case <-timerCh:
			n.sendDue(n.now())
		}
	}
}

func (n *Notifier) nextDueWait() (time.Duration, bool) {
	now := n.now()

	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.dueHeap) == 0 {
		return 0, false
	}

	wait := n.dueHeap[0].nextDue.Sub(now)
	if wait < 0 {
		wait = 0
	}

	return wait, true
}

func (n *Notifier) sendDue(now time.Time) {
	due := make([]Notification, 0)

	n.mu.Lock()
	for len(n.dueHeap) > 0 {
		entry := n.dueHeap[0]
		if entry.nextDue.After(now) {
			break
		}

		due = append(due, entry.notification)
		entry.nextDue = now.Add(entry.notification.RepeatEvery)
		heap.Fix(&n.dueHeap, entry.index)
	}
	n.mu.Unlock()

	for _, notification := range due {
		n.sendFn(notification)
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

func notifyLoop(wakeupCh chan struct{}) {
	if wakeupCh == nil {
		return
	}

	select {
	case wakeupCh <- struct{}{}:
	default:
	}
}

func stopTimer(timer *time.Timer) {
	if timer == nil {
		return
	}

	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	stopTimer(timer)
	timer.Reset(duration)
}
