package listing

import tea "charm.land/bubbletea/v2"

func (m *Model) startPullRequestSync(trigger PullRequestSyncTrigger) tea.Cmd {
	m.PRSyncInFlight++

	return tea.Batch(
		performPullRequestSync(m.Repositories, trigger),
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

func (m *Model) pullRequestSpinnerView() string {
	if !m.isPullRequestSyncInFlight() {
		return ""
	}

	return m.PRSyncSpinner.View()
}
