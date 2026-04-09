package ui

import (
	"fresh/internal/domain"
	"fresh/internal/notifications"
	"fresh/internal/ui/views/listing"
	"fresh/internal/ui/views/pullrequests"
	"fresh/internal/ui/views/scanning"

	tea "charm.land/bubbletea/v2"
)

type CurrentView int

const (
	ScanningView CurrentView = iota
	RepoListView
	RepoPRListView
)

type MainModel struct {
	currentView      CurrentView
	scanningView     *scanning.Model
	listingView      *listing.Model
	pullRequestsView *pullrequests.Model
	pullRequestCache map[string][]domain.PullRequestDetails
	notifier         *notifications.Notifier
	width, height    int
}

func New(scanDir string, notifier ...*notifications.Notifier) *MainModel {
	var injectedNotifier *notifications.Notifier
	if len(notifier) > 0 {
		injectedNotifier = notifier[0]
	}

	return &MainModel{
		currentView:      ScanningView,
		scanningView:     scanning.New(scanDir),
		pullRequestCache: make(map[string][]domain.PullRequestDetails),
		notifier:         injectedNotifier,
	}
}

func (m *MainModel) Init() tea.Cmd {
	switch m.currentView {
	case ScanningView:
		return m.scanningView.Init()
	case RepoListView:
		return m.listingView.Init()
	case RepoPRListView:
		return m.pullRequestsView.Init()
	default:
		return nil
	}
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case scanning.ScanFinishedMsg:
		m.currentView = RepoListView
		m.listingView = listing.NewWithNotifier(msg.Repos, m.notifier)
		m.listingView.SetSize(m.width, m.height)
		return m, m.listingView.Init()

	case listing.OpenPullRequestsMsg:
		cached := append([]domain.PullRequestDetails(nil), m.pullRequestCache[msg.Repo.Path]...)
		m.currentView = RepoPRListView
		m.pullRequestsView = pullrequests.New(msg.Repo, cached)
		m.pullRequestsView.SetSize(m.width, m.height)
		return m, m.pullRequestsView.Init()

	case pullrequests.BackToRepoListMsg:
		m.currentView = RepoListView
		if m.listingView != nil {
			m.listingView.SetSize(m.width, m.height)
		}
		return m, nil

	case pullrequests.PullRequestsLoadedMsg:
		if msg.RepoPath != "" && msg.Error == "" {
			m.pullRequestCache[msg.RepoPath] = append([]domain.PullRequestDetails(nil), msg.PullRequests...)
		}
	}

	switch m.currentView {
	case ScanningView:
		if m.scanningView != nil {
			m.scanningView, cmd = m.scanningView.Update(msg)
		}
	case RepoListView:
		if m.listingView != nil {
			m.listingView, cmd = m.listingView.Update(msg)
		}
	case RepoPRListView:
		if m.pullRequestsView != nil {
			m.pullRequestsView, cmd = m.pullRequestsView.Update(msg)
		}
	}
	return m, cmd
}

func (m *MainModel) View() tea.View {
	v := tea.NewView("")
	switch m.currentView {
	case ScanningView:
		v.SetContent(m.scanningView.View())
	case RepoListView:
		if m.listingView == nil {
			return v
		}
		v.SetContent(m.listingView.View())
	case RepoPRListView:
		if m.pullRequestsView == nil {
			return v
		}
		v.SetContent(m.pullRequestsView.View())
	}
	return v
}
