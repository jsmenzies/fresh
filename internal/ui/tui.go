package ui

import (
	"fresh/internal/git"
	"fresh/internal/ui/listing"
	"fresh/internal/ui/scanning"

	tea "github.com/charmbracelet/bubbletea"
)

type CurrentView int

const (
	Scanning CurrentView = iota
	Listing
)

type MainModel struct {
	currentView   CurrentView
	scanningModel *scanning.Model
	listingModel  *listing.Model
	gitClient     *git.Git
	width, height int
}

func New(scanningModel *scanning.Model, gitClient *git.Git) *MainModel {
	return &MainModel{
		currentView:   Scanning,
		scanningModel: scanningModel,
		gitClient:     gitClient,
	}
}

func (m *MainModel) Init() tea.Cmd {
	switch m.currentView {
	case Scanning:
		return m.scanningModel.Init()
	case Listing:
		return m.listingModel.Init()
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
		m.currentView = Listing
		m.listingModel = listing.New(msg.Repos, m.gitClient)
		return m, m.listingModel.Init()
	}

	switch m.currentView {
	case Scanning:
		newScanningModel, newCmd := m.scanningModel.Update(msg)
		if newModel, ok := newScanningModel.(*scanning.Model); ok {
			m.scanningModel = newModel
		}
		cmd = newCmd
	case Listing:
		newListingModel, newCmd := m.listingModel.Update(msg)
		if newModel, ok := newListingModel.(*listing.Model); ok {
			m.listingModel = newModel
		}
		cmd = newCmd
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *MainModel) View() string {
	switch m.currentView {
	case Scanning:
		return m.scanningModel.View()
	case Listing:
		return m.listingModel.View()
	default:
		return ""
	}
}
