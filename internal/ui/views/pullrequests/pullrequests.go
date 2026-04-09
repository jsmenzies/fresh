package pullrequests

import (
	"fmt"
	"strings"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type prListKeyMap struct {
	refresh key.Binding
	back    key.Binding
}

func newPRListKeyMap() *prListKeyMap {
	return &prListKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh pull requests"),
		),
		back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to repositories"),
		),
	}
}

type Model struct {
	Repo          domain.Repository
	PullRequests  []domain.PullRequestDetails
	Cursor        int
	Keys          *prListKeyMap
	layout        ColumnLayout
	width, height int
	Loading       bool
	Unsupported   bool
	LoadError     string
	PulseOn       bool
	PulseEvery    time.Duration
	Spinner       spinner.Model
}

func New(repo domain.Repository, cached []domain.PullRequestDetails) *Model {
	rows := append([]domain.PullRequestDetails(nil), cached...)

	return &Model{
		Repo:         repo,
		PullRequests: rows,
		Cursor:       0,
		Keys:         newPRListKeyMap(),
		layout:       calculateColumnLayout(0),
		Loading:      true,
		Unsupported:  false,
		LoadError:    "",
		PulseOn:      false,
		PulseEvery:   350 * time.Millisecond,
		Spinner:      common.NewPullRequestSpinner(),
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.layout = calculateColumnLayout(width)
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		schedulePulseTick(m.PulseEvery),
		performPullRequestLoad(m.Repo),
		m.Spinner.Tick,
	)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.Keys.back) || msg.String() == "esc" || msg.Code == 27:
			return m, backToRepoList()
		case key.Matches(msg, m.Keys.refresh):
			m.Loading = true
			m.Unsupported = false
			m.LoadError = ""
			return m, tea.Batch(performPullRequestLoad(m.Repo), m.Spinner.Tick)
		case msg.String() == "up", msg.String() == "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case msg.String() == "down", msg.String() == "j":
			if m.Cursor < len(m.PullRequests)-1 {
				m.Cursor++
			}
		}

	case PullRequestsLoadedMsg:
		if msg.RepoPath != m.Repo.Path {
			return m, nil
		}

		m.Loading = false
		m.Unsupported = msg.Unsupported
		m.LoadError = msg.Error
		if msg.Error == "" {
			m.PullRequests = msg.PullRequests
		} else if msg.Unsupported {
			m.PullRequests = nil
		}

		if len(m.PullRequests) == 0 {
			m.Cursor = 0
		} else if m.Cursor >= len(m.PullRequests) {
			m.Cursor = len(m.PullRequests) - 1
		}

	case pulseTickMsg:
		m.PulseOn = !m.PulseOn
		return m, schedulePulseTick(m.PulseEvery)

	case spinner.TickMsg:
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) View() string {
	var s strings.Builder

	title := fmt.Sprintf("\nPull requests for %s\n", m.Repo.Name)
	s.WriteString(common.HeaderStyle.Render(title))

	if m.Unsupported {
		msg := "Pull requests are currently only supported for GitHub repositories."
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(common.SubtleGray).Render(msg))
		s.WriteString("\n\n")
		s.WriteString(m.buildFooter())
		return s.String()
	}

	if len(m.PullRequests) == 0 {
		s.WriteString("\n")
		switch {
		case m.Loading:
			s.WriteString(common.TextGrey.Render(m.Spinner.View() + " Loading pull requests..."))
		case m.LoadError != "":
			s.WriteString(lipgloss.NewStyle().Foreground(common.SubtleRed).Render("Failed to load pull requests: " + m.LoadError))
		default:
			s.WriteString(common.TextGrey.Render("No open pull requests."))
		}
		s.WriteString("\n\n")
		s.WriteString(m.buildFooter())
		return s.String()
	}

	s.WriteString(RenderTable(m.PullRequests, m.Cursor, m.layout, m.PulseOn))
	s.WriteString("\n")

	if m.Loading {
		s.WriteString("\n")
		s.WriteString(common.TextGrey.Render(m.Spinner.View() + " Refreshing pull requests in background..."))
	}
	if m.LoadError != "" {
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(common.SubtleRed).Render("Refresh error: " + m.LoadError))
	}

	s.WriteString("\n\n")
	s.WriteString(m.buildFooter())

	return s.String()
}

func (m *Model) buildFooter() string {
	legend := "my PRs highlighted in blue"
	hotkeys := []string{
		"↑/↓ navigate",
		"r refresh",
		"esc back",
		"q quit",
	}
	return common.FooterStyle.Render(strings.Join(hotkeys, "  •  ") + "  •  " + legend)
}
