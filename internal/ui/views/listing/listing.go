package listing

import (
	"fresh/internal/domain"
	"fresh/internal/notifications"
	"fresh/internal/pullrequests"
	"fresh/internal/ui/views/common"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type listKeyMap struct {
	refresh      key.Binding
	watch        key.Binding
	pullAll      key.Binding
	pruneAll     key.Binding
	openPRs      key.Binding
	toggleLegend key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		watch: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle watch mode"),
		),
		pullAll: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull all updates"),
		),
		pruneAll: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "prune merged branches"),
		),
		openPRs: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view pull requests"),
		),
		toggleLegend: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle legend"),
		),
	}
}

type Model struct {
	Repositories     []domain.Repository
	Cursor           int
	Keys             *listKeyMap
	layout           ColumnLayout
	width, height    int
	ShowLegend       bool
	InfoPhase        uint64
	RotateEvery      time.Duration
	ActivityTTL      time.Duration
	RecentInfo       map[string][]TimedInfoMessage
	StartupPRSync    bool
	PRSyncInFlight   int
	PRSyncGeneration uint64
	PRSyncSpinner    spinner.Model
	BlockedSpinner   spinner.Model
	ReadySpinner     spinner.Model
	WatchEnabled     bool
	WatchToken       uint64
	WatchBackoff     int
	WatchEvery       time.Duration
	WatchMaxEvery    time.Duration
	prCoordinator    *pullrequests.NotificationCoordinator
	notifier         *notifications.Notifier
}

func New(repos []domain.Repository) *Model {
	return NewWithNotifier(repos, nil)
}

func NewWithNotifier(repos []domain.Repository, notifier *notifications.Notifier) *Model {
	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})

	for i := range repos {
		repos[i].Activity = &domain.IdleActivity{}
	}

	return &Model{
		Repositories:     repos,
		Cursor:           0,
		Keys:             newListKeyMap(),
		layout:           calculateColumnLayout(repos, 0),
		ShowLegend:       false,
		RotateEvery:      10 * time.Second,
		ActivityTTL:      10 * time.Second,
		RecentInfo:       make(map[string][]TimedInfoMessage),
		StartupPRSync:    false,
		PRSyncInFlight:   0,
		PRSyncGeneration: 0,
		PRSyncSpinner:    common.NewPullRequestSpinner(),
		BlockedSpinner:   common.NewBlockedPullRequestSpinner(),
		ReadySpinner:     common.NewReadyPullRequestSpinner(),
		WatchEnabled:     false,
		WatchToken:       0,
		WatchBackoff:     0,
		WatchEvery:       defaultWatchInterval,
		WatchMaxEvery:    defaultWatchMaxInterval,
		prCoordinator:    pullrequests.NewNotificationCoordinator(nil),
		notifier:         notifier,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.layout = calculateColumnLayout(m.Repositories, width)
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, scheduleInfoRotateTick(m.RotateEvery))
	if !m.StartupPRSync {
		cmds = append(cmds, m.startPullRequestSync(pullRequestSyncStartup))
		m.StartupPRSync = true
	}
	cmds = append(cmds, m.BlockedSpinner.Tick)
	cmds = append(cmds, m.ReadySpinner.Tick)
	for i := range m.Repositories {
		repo := &m.Repositories[i]
		repo.Activity = &domain.RefreshingActivity{
			Spinner: common.NewRefreshSpinner(),
		}
		cmds = append(cmds, performInitialRefresh(i, *repo))
		cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.Keys.refresh):
			return m, m.startRefreshCycle(pullRequestSyncManual)

		case key.Matches(msg, m.Keys.watch):
			return m, m.toggleWatchMode()

		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !repo.IsBusy() && repo.CanPull() {
					repo.Activity = &domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
					}
					cmds = append(cmds, performPull(i, repo.Path))
					cmds = append(cmds, repo.Activity.(*domain.PullingActivity).Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pruneAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !repo.IsBusy() && len(repo.Branches.Merged) > 0 {
					repo.Activity = &domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
					}
					cmds = append(cmds, performPrune(i, repo.Path, repo.Branches.Merged))
					cmds = append(cmds, repo.Activity.(*domain.PruningActivity).Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.openPRs) || msg.String() == "enter" || msg.Code == '\r':
			if len(m.Repositories) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Repositories) {
				return m, nil
			}
			return m, openPullRequestsView(m.Repositories[m.Cursor])

		case key.Matches(msg, m.Keys.toggleLegend):
			m.ShowLegend = !m.ShowLegend
			return m, nil

		case msg.String() == "up", msg.String() == "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		case msg.String() == "down", msg.String() == "j":
			if m.Cursor < len(m.Repositories)-1 {
				m.Cursor++
			}
		}

	case RepoUpdatedMsg:
		m.applyRepoUpdate(msg.Index, msg.Repo, func(repo *domain.Repository, activity domain.Activity) {
			if refreshing, ok := activity.(*domain.RefreshingActivity); ok {
				refreshing.MarkComplete()
				repo.Activity = &domain.IdleActivity{}
			}
		})

	case PullRequestStatesUpdatedMsg:
		if msg.Generation != m.PRSyncGeneration {
			m.completePullRequestSync()
			if msg.Trigger == pullRequestSyncWatch && m.WatchEnabled {
				return m, scheduleWatchTick(m.currentWatchInterval(), m.WatchToken)
			}
			return m, nil
		}

		m.applyPullRequestStates(msg.States)
		syncFailed := hasPullRequestSyncError(msg.States)
		if !syncFailed {
			m.applyPullRequestWatchlist(msg.Tracked, msg.Trigger == pullRequestSyncStartup)
		}
		m.completePullRequestSync()
		if msg.Trigger == pullRequestSyncWatch && m.WatchEnabled {
			m.updateWatchBackoff(syncFailed)
			return m, scheduleWatchTick(m.currentWatchInterval(), m.WatchToken)
		}

	case watchTickMsg:
		if !m.WatchEnabled || msg.Token != m.WatchToken {
			return m, nil
		}
		return m, m.startRefreshCycle(pullRequestSyncWatch)

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			if pulling, ok := repo.Activity.(*domain.PullingActivity); ok {
				pulling.AddLine(msg.line)
			}
		}
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}

	case pullCompleteMsg:
		m.finalizeRepoActivity(msg.Index, msg.Repo, func(activity domain.Activity) ActivityFinalizeResult {
			pulling, ok := activity.(*domain.PullingActivity)
			if !ok {
				return ActivityFinalizeResult{}
			}
			pulling.MarkComplete(msg.outcome)
			return ActivityFinalizeResult{
				Completed: true,
				Info:      buildPullCompletionInfoMessage(*pulling),
			}
		})

	case pruneWorkState:
		return m, listenForPruneProgress(msg)

	case pruneLineMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			if pruning, ok := repo.Activity.(*domain.PruningActivity); ok {
				pruning.AddLine(msg.line)
			}
		}
		if msg.state != nil {
			return m, listenForPruneProgress(*msg.state)
		}

	case pruneCompleteMsg:
		m.finalizeRepoActivity(msg.Index, msg.Repo, func(activity domain.Activity) ActivityFinalizeResult {
			pruning, ok := activity.(*domain.PruningActivity)
			if !ok {
				return ActivityFinalizeResult{}
			}
			pruning.MarkComplete(msg.outcome)
			return ActivityFinalizeResult{
				Completed: true,
				Info:      buildPruneCompletionInfoMessage(*pruning),
			}
		})

	case infoRotateTickMsg:
		m.InfoPhase++
		m.pruneExpiredRecentActivityInfo(time.Now())
		return m, scheduleInfoRotateTick(m.RotateEvery)

	case spinner.TickMsg:
		var cmds []tea.Cmd
		if m.isPullRequestSyncInFlight() {
			var cmd tea.Cmd
			m.PRSyncSpinner, cmd = m.PRSyncSpinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		var blockedCmd tea.Cmd
		m.BlockedSpinner, blockedCmd = m.BlockedSpinner.Update(msg)
		if blockedCmd != nil {
			cmds = append(cmds, blockedCmd)
		}
		var readyCmd tea.Cmd
		m.ReadySpinner, readyCmd = m.ReadySpinner.Update(msg)
		if readyCmd != nil {
			cmds = append(cmds, readyCmd)
		}
		for i := range m.Repositories {
			switch activity := m.Repositories[i].Activity.(type) {
			case *domain.RefreshingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			case *domain.PullingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			case *domain.PruningActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
	}
	return m, nil
}

func (m *Model) applyPullRequestStates(states map[string]domain.PullRequestState) {
	if len(states) == 0 {
		return
	}

	for i := range m.Repositories {
		repo := &m.Repositories[i]
		if state, ok := states[repo.Path]; ok {
			repo.PullRequests = state
		}
	}
}

func (m *Model) applyRepoUpdate(index int, next domain.Repository, onUpdate func(*domain.Repository, domain.Activity)) {
	if index < 0 || index >= len(m.Repositories) {
		return
	}

	repo := &m.Repositories[index]
	activity := repo.Activity
	pullRequests := repo.PullRequests

	*repo = next
	repo.PullRequests = pullRequests
	repo.Activity = activity

	if onUpdate != nil {
		onUpdate(repo, activity)
	}
}

func (m *Model) applyPullRequestWatchlist(tracked []pullrequests.Snapshot, seed bool) {
	if m.prCoordinator == nil {
		return
	}

	m.prCoordinator.Sync(tracked, pullrequests.ApplyOptions{Seed: seed}, m.notifier)
}

func (m *Model) View() string {
	var s strings.Builder
	s.WriteString(common.FormatHeader(len(m.Repositories)))

	if len(m.Repositories) == 0 {
		s.WriteString("No repositories found")
		return s.String()
	}

	runtime := InfoRuntime{
		Phase:                m.InfoPhase,
		Now:                  time.Now(),
		RecentActivityByRepo: m.RecentInfo,
		PullRequestSyncing:   m.isPullRequestSyncInFlight(),
		BlockedSpinner:       m.BlockedSpinner.View(),
		ReadySpinner:         m.ReadySpinner.View(),
	}
	s.WriteString(GenerateTable(m.Repositories, m.Cursor, m.layout, runtime))
	s.WriteString("\n\n")

	s.WriteString(m.buildFooter())

	legend := RenderLegend(m.ShowLegend)
	s.WriteString("\n\n")
	s.WriteString(legend)

	return s.String()
}

func (m *Model) buildFooter() string {
	watchStatus := "w watch off"
	if m.WatchEnabled {
		watchStatus = "w watch on (" + m.currentWatchInterval().String() + ")"
	}

	hotkeys := []string{
		"↑/↓ navigate",
		"enter view PRs",
		"r refresh",
		watchStatus,
		"p pull all updates",
		"b prune merged branches",
		"? toggle legend",
		"q quit",
	}
	footerText := strings.Join(hotkeys, "  •  ")
	return common.FooterStyle.Render(footerText)
}

func (m *Model) storeRecentActivityInfo(repoPath string, message InfoMessage) {
	if repoPath == "" || message.Text == "" {
		return
	}

	now := time.Now()
	expiresAt := now.Add(m.ActivityTTL)
	m.RecentInfo[repoPath] = append(m.RecentInfo[repoPath], TimedInfoMessage{Message: message, ExpiresAt: expiresAt})
}

func (m *Model) pruneExpiredRecentActivityInfo(now time.Time) {
	for repoPath, items := range m.RecentInfo {
		filtered := items[:0]
		for _, item := range items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				continue
			}
			filtered = append(filtered, item)
		}
		if len(filtered) == 0 {
			delete(m.RecentInfo, repoPath)
			continue
		}
		m.RecentInfo[repoPath] = filtered
	}
}
