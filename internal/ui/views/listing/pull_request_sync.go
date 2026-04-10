package listing

import tea "charm.land/bubbletea/v2"

func (m *Model) startPullRequestSync(trigger PullRequestSyncTrigger) tea.Cmd {
	m.PRSyncInFlight++
	m.PRSyncGeneration++
	generation := m.PRSyncGeneration

	return tea.Batch(
		performPullRequestSync(m.Repositories, trigger, generation),
		m.PRSyncSpinner.Tick,
	)
}

func (m *Model) completePullRequestSync() {
	if m.PRSyncInFlight <= 0 {
		m.PRSyncInFlight = 0
		return
	}

	m.PRSyncInFlight--
}

func (m *Model) isPullRequestSyncInFlight() bool {
	return m.PRSyncInFlight > 0
}
