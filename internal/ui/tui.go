package ui

import (
	"fresh/internal/ui/views/listing"
	"fresh/internal/ui/views/scanning"

	tea "charm.land/bubbletea/v2"
)

type CurrentView int

const (
	ScanningView CurrentView = iota
	RepoListView
)

type MainModel struct {
	currentView   CurrentView
	scanningView  *scanning.Model
	listingView   *listing.Model
	width, height int
}

func New(scanDir string) *MainModel {
	return &MainModel{
		currentView:  ScanningView,
		scanningView: scanning.New(scanDir),
	}
}

func (m *MainModel) Init() tea.Cmd {
	switch m.currentView {
	case ScanningView:
		return m.scanningView.Init()
	case RepoListView:
		return m.listingView.Init()
	default:
		return nil
	}
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

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
		m.listingView = listing.New(msg.Repos)
		m.listingView.SetSize(m.width, m.height)
		return m, m.listingView.Init()
	}

	switch m.currentView {
	case ScanningView:
		m.scanningView, cmd = m.scanningView.Update(msg)
	case RepoListView:
		m.listingView, cmd = m.listingView.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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
	}
	return v
}
