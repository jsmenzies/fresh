package listing

import (
	"fmt"
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"sort"
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
			key.WithHelp("r", "refresh remote"),
		),
		updateAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh all"),
		),
		pull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull"),
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
	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})

	for i := range repos {
		repos[i].Activity = domain.IdleActivity{}
	}

	return &Model{
		Repositories: repos,
		Cursor:       0,
		Keys:         newListKeyMap(),
	}
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.Repositories {
		repo := &m.Repositories[i]
		repo.Activity = domain.RefreshingActivity{
			Spinner: common.NewRefreshSpinner(),
		}
		cmds = append(cmds, performRefresh(repo.Path))
		cmds = append(cmds, repo.Activity.(domain.RefreshingActivity).Spinner.Tick)
	}
	return tea.Batch(cmds...)
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
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) {
					refreshing := domain.RefreshingActivity{
						Spinner: common.NewRefreshSpinner(),
					}
					repo.Activity = refreshing
					return m, tea.Batch(
						performRefresh(repo.Path),
						refreshing.Spinner.Tick,
					)
				}
			}

		case key.Matches(msg, m.Keys.updateAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) {
					refreshing := domain.RefreshingActivity{
						Spinner: common.NewRefreshSpinner(),
					}
					repo.Activity = refreshing
					cmds = append(cmds, performRefresh(repo.Path))
					cmds = append(cmds, refreshing.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pull):
			if m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) {
					pulling := domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pulling
					return m, tea.Batch(
						performPull(repo.Path),
						pulling.Spinner.Tick,
					)
				}
			}
		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) {
					pulling := domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pulling
					cmds = append(cmds, performPull(repo.Path))
					cmds = append(cmds, pulling.Spinner.Tick)
				}
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

	case RepoUpdatedMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.Repo.Path {
				activity := m.Repositories[i].Activity
				m.Repositories[i] = msg.Repo

				if refreshing, ok := activity.(domain.RefreshingActivity); ok {
					refreshing.MarkComplete()
					m.Repositories[i].Activity = refreshing
				} else {
					m.Repositories[i].Activity = activity
				}
				break
			}
		}

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				if pulling, ok := m.Repositories[i].Activity.(domain.PullingActivity); ok {
					pulling.AddLine(msg.line)
					m.Repositories[i].Activity = pulling
				}
			}
		}
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}

	case pullCompleteMsg:
		for i := range m.Repositories {
			if m.Repositories[i].Path == msg.repoPath {
				activity := m.Repositories[i].Activity
				m.Repositories[i] = msg.Repo
				if pulling, ok := activity.(domain.PullingActivity); ok {
					pulling.MarkComplete(msg.exitCode)
					m.Repositories[i].Activity = pulling
				} else {
					m.Repositories[i].Activity = activity
				}
			}
		}

	case spinner.TickMsg:
		var cmds []tea.Cmd
		for i := range m.Repositories {
			switch activity := m.Repositories[i].Activity.(type) {
			case domain.RefreshingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					m.Repositories[i].Activity = activity
					cmds = append(cmds, cmd)
				}
			case domain.PullingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					m.Repositories[i].Activity = activity
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
	keyStyle := lipgloss.NewStyle().Foreground(common.SubtleGray).PaddingLeft(2)
	hotkeys := []string{
		"↑/↓ navigate",
		"r refresh",
		"R refresh all",
		"p pull",
		"P pull all",
		"q quit",
	}
	footerText := strings.Join(hotkeys, "  •  ")
	return "\n" + keyStyle.Render(footerText)
}

func isBusy(repo domain.Repository) bool {
	switch a := repo.Activity.(type) {
	case domain.IdleActivity:
		return false
	case domain.RefreshingActivity:
		return !a.Complete
	case domain.PullingActivity:
		return !a.Complete
	default:
		return false
	}
}
