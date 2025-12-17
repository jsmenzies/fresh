package listing

import (
	"fmt"
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listKeyMap struct {
	refresh   key.Binding
	updateAll key.Binding
	pull      key.Binding
	pullAll   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "fetch remote status"),
		),
		updateAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh all"),
		),
		pull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull selected"),
		),
		pullAll: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "pull all"),
		),
	}
}

type Model struct {
	Repositories  []domain.Repository
	Cursor        int
	Keys          *listKeyMap
	width, height int
}

func New(repos []domain.Repository) *Model {
	for i := range repos {
		repos[i].RefreshSpinner = common.NewSecondaryDotSpinner()
		repos[i].PullSpinner = common.NewSecondaryDotSpinner()
		repos[i].PullState = &domain.PullState{} // Initialize PullState
	}
	return &Model{
		Repositories: repos,
		Cursor:       0,
		Keys:         newListKeyMap(),
	}
}

func (m *Model) Init() tea.Cmd {
	return startBackgroundRefresh(m.Repositories)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.refresh):
			if m.Cursor < len(m.Repositories) {
				repo := m.Repositories[m.Cursor]
				return m, startBackgroundRefresh([]domain.Repository{repo})
			}
		case key.Matches(msg, m.Keys.updateAll):
			return m, startBackgroundRefresh(m.Repositories)
		case key.Matches(msg, m.Keys.pull):
			if m.Cursor < len(m.Repositories) {
				repo := m.Repositories[m.Cursor]
				m.Repositories[m.Cursor].PullState = domain.NewPullState()
				return m, tea.Batch(
					performPull(repo.Path),
					m.Repositories[m.Cursor].PullSpinner.Tick,
				)
			}
		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				m.Repositories[i].PullState = domain.NewPullState()
				cmds = append(cmds, performPull(m.Repositories[i].Path))
				cmds = append(cmds, m.Repositories[i].PullSpinner.Tick)
			}
			return m, tea.Batch(cmds...)
		case msg.String() == "up", msg.String() == "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case msg.String() == "down", msg.String() == "j":
			if m.Cursor < len(m.Repositories)-1 {
				m.Cursor++
			}
		}

	case refreshStartMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				m.Repositories[i].Refreshing = true
				m.Repositories[i].PullState = nil
				return m, tea.Batch(
					performRefresh(msg.repoPath),
					m.Repositories[i].RefreshSpinner.Tick,
				)
			}
		}

	case refreshCompleteMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				m.Repositories[i].Refreshing = false
				m.Repositories[i].AheadCount = msg.aheadCount
				m.Repositories[i].BehindCount = msg.behindCount
				m.Repositories[i].HasRemoteUpdates = msg.behindCount > 0
				m.Repositories[i].HasError = msg.hasError
				m.Repositories[i].ErrorMessage = msg.errorMessage
				break
			}
		}

	case pullStartMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				m.Repositories[i].PullState = domain.NewPullState()
				return m, m.Repositories[i].PullSpinner.Tick
			}
		}

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				if m.Repositories[i].PullState != nil {
					m.Repositories[i].PullState.AddLine(msg.line)
				}
			}
		}
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}

	case pullCompleteMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				if m.Repositories[i].PullState != nil {
					m.Repositories[i].PullState.Complete(msg.exitCode)
				}
				if msg.exitCode == 0 {
					m.Repositories[i].BehindCount = 0
					m.Repositories[i].HasRemoteUpdates = false
				}
			}
		}

	case spinner.TickMsg:
		var cmds []tea.Cmd
		for i := range m.Repositories {
			if m.Repositories[i].Refreshing {
				var cmd tea.Cmd
				m.Repositories[i].RefreshSpinner, cmd = m.Repositories[i].RefreshSpinner.Update(msg)
				cmds = append(cmds, cmd)
			}
			if m.Repositories[i].PullState != nil && m.Repositories[i].PullState.InProgress {
				var cmd tea.Cmd
				m.Repositories[i].PullSpinner, cmd = m.Repositories[i].PullSpinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
	}
	return m, nil
}

func (m *Model) View() string {
	var style = lipgloss.NewStyle().Foreground(common.Blue)
	var text = style.Render(fmt.Sprintf("\nScan complete. %d repositories found\n", len(m.Repositories)))
	headerLine := text

	if len(m.Repositories) == 0 {
		return headerLine + "No repositories found"
	}

	tableView := common.GenerateTable(m.Repositories, m.Cursor)
	footer := buildFooter()

	return headerLine + tableView + "\n" + footer
}

func buildFooter() string {
	keyStyle := lipgloss.NewStyle().Foreground(common.SubtleGray)
	hotkeys := []string{
		"↑/↓ navigate",
		"r fetch remote status",
		"R refresh all",
		"p pull selected",
		"P pull all",
		"q quit",
	}
	footerText := strings.Join(hotkeys, "  •  ")
	return "\n" + keyStyle.Render(footerText)
}