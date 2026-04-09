package listing

import (
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	tea "charm.land/bubbletea/v2"
)

const (
	defaultWatchInterval    = time.Minute
	defaultWatchMaxInterval = 8 * time.Minute
	maxWatchBackoff         = 8
)

type watchTickMsg struct {
	Token uint64
}

func scheduleWatchTick(interval time.Duration, token uint64) tea.Cmd {
	if interval <= 0 {
		interval = defaultWatchInterval
	}

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return watchTickMsg{Token: token}
	})
}

func (m *Model) toggleWatchMode() tea.Cmd {
	if m.WatchEnabled {
		m.WatchEnabled = false
		m.WatchBackoff = 0
		m.WatchToken++
		return nil
	}

	m.WatchEnabled = true
	m.WatchBackoff = 0
	m.WatchToken++

	return scheduleWatchTick(m.currentWatchInterval(), m.WatchToken)
}

func (m *Model) startRefreshCycle(trigger PullRequestSyncTrigger) tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.startPullRequestSync(trigger))
	for i := range m.Repositories {
		repo := &m.Repositories[i]
		if repo.IsBusy() {
			continue
		}
		repo.Activity = &domain.RefreshingActivity{Spinner: common.NewRefreshSpinner()}
		cmds = append(cmds, performRefresh(i, repo.Path))
		cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick)
	}

	return tea.Batch(cmds...)
}

func (m *Model) currentWatchInterval() time.Duration {
	base, max := m.watchIntervals()
	interval := base

	for i := 0; i < m.WatchBackoff; i++ {
		if interval >= max {
			return max
		}
		next := interval * 2
		if next > max {
			return max
		}
		interval = next
	}

	return interval
}

func (m *Model) watchIntervals() (time.Duration, time.Duration) {
	base := m.WatchEvery
	if base <= 0 {
		base = defaultWatchInterval
	}

	max := m.WatchMaxEvery
	if max < base {
		max = base
	}

	return base, max
}

func (m *Model) updateWatchBackoff(hasError bool) {
	if !hasError {
		m.WatchBackoff = 0
		return
	}

	if m.WatchBackoff < maxWatchBackoff {
		m.WatchBackoff++
	}
}

func hasPullRequestSyncError(states map[string]domain.PullRequestState) bool {
	for _, state := range states {
		if _, ok := state.(domain.PullRequestError); ok {
			return true
		}
	}
	return false
}
