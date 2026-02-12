package ui

import (
	"fresh/internal/config"
	"fresh/internal/git"
	"fresh/internal/scanner"
	"fresh/internal/ui/views/listing"
	"fresh/internal/ui/views/scanning"

	tea "github.com/charmbracelet/bubbletea"
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
	gitClient     git.Client
	width, height int
}

func New(scanDir string) *MainModel {
	gitClient := git.NewExecClient(config.DefaultConfig())

	return &MainModel{
		currentView:  ScanningView,
		scanningView: scanning.NewWithDependencies(gitClient, scanner.New(scanDir)),
		gitClient:    gitClient,
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
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case scanning.ScanFinishedMsg:
		m.currentView = RepoListView
		m.listingView = listing.New(msg.Repos, m.gitClient)
		return m, m.listingView.Init()
	}

	switch m.currentView {
	case ScanningView:
		newScanningModel, newCmd := m.scanningView.Update(msg)
		if newModel, ok := newScanningModel.(*scanning.Model); ok {
			m.scanningView = newModel
		}
		cmd = newCmd
	case RepoListView:
		newListingModel, newCmd := m.listingView.Update(msg)
		if newModel, ok := newListingModel.(*listing.Model); ok {
			m.listingView = newModel
		}
		cmd = newCmd
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *MainModel) View() string {
	switch m.currentView {
	case ScanningView:
		return m.scanningView.View()
	case RepoListView:
		return m.listingView.View()
	default:
		return ""
	}
}
