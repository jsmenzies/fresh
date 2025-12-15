package ui

import (
	"fresh/internal/domain"
	"fresh/internal/scanner"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	State        appState
	Spinner      spinner.Model
	Cursor       int
	Keys         *listKeyMap
	Choice       string
	Repositories []domain.Repository
	Scanner      *scanner.Scanner
}

type appState int

const (
	Scanning appState = iota
	Listing
	Quitting
)

type listKeyMap struct {
	refresh   key.Binding
	updateAll key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		updateAll: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "update all"),
		),
	}
}

func newDotSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return s
}

func NewModel(scanDir string) Model {
	s := newDotSpinner()
	s.Style = SpinnerStyle

	return Model{
		State:        Scanning,
		Spinner:      s,
		Cursor:       0,
		Keys:         newListKeyMap(),
		Scanner:      scanner.New(scanDir),
		Repositories: make([]domain.Repository, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(startScanning(m.Scanner), tickCmd(), m.Spinner.Tick)
}
