package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type listKeyMap struct {
	refresh      key.Binding
	pullAll      key.Binding
	pruneAll     key.Binding
	toggleLegend key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		pullAll: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull all updates"),
		),
		pruneAll: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "prune merged branches"),
		),
		toggleLegend: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle legend"),
		),
	}
}

type Model struct {
	Repositories  []domain.Repository
	Cursor        int
	Keys          *listKeyMap
	width, height int
	ShowLegend    bool
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
		ShowLegend:   false,
	}
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.Repositories {
		repo := &m.Repositories[i]
		repo.Activity = domain.RefreshingActivity{
			Spinner: common.NewRefreshSpinner(),
		}
		cmds = append(cmds, performRefresh(i, repo.Path))
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
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) {
					refreshing := domain.RefreshingActivity{
						Spinner: common.NewRefreshSpinner(),
					}
					repo.Activity = refreshing
					cmds = append(cmds, performRefresh(i, repo.Path))
					cmds = append(cmds, refreshing.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) && shouldPull(*repo) {
					pulling := domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
					}
					repo.Activity = pulling
					cmds = append(cmds, performPull(i, repo.Path))
					cmds = append(cmds, pulling.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pruneAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) && len(repo.Branches.Merged) > 0 {
					pruning := domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
					}
					repo.Activity = pruning
					cmds = append(cmds, performPrune(i, repo.Path, repo.Branches.Merged))
					cmds = append(cmds, pruning.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

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
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			activity := repo.Activity
			*repo = msg.Repo

			if refreshing, ok := activity.(domain.RefreshingActivity); ok {
				refreshing.MarkComplete()
				repo.Activity = refreshing
			} else {
				repo.Activity = activity
			}
		}

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			if pulling, ok := repo.Activity.(domain.PullingActivity); ok {
				pulling.AddLine(msg.line)
				repo.Activity = pulling
			}
		}
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}

	case pullCompleteMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			activity := repo.Activity
			*repo = msg.Repo
			if pulling, ok := activity.(domain.PullingActivity); ok {
				pulling.MarkComplete(msg.exitCode)
				repo.Activity = pulling
			} else {
				repo.Activity = activity
			}
		}

	case pruneWorkState:
		return m, listenForPruneProgress(msg)

	case pruneLineMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			if pruning, ok := repo.Activity.(domain.PruningActivity); ok {
				pruning.AddLine(msg.line)
				repo.Activity = pruning
			}
		}
		if msg.state != nil {
			return m, listenForPruneProgress(*msg.state)
		}

	case pruneCompleteMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			activity := repo.Activity
			*repo = msg.Repo
			if pruning, ok := activity.(domain.PruningActivity); ok {
				pruning.MarkComplete(msg.exitCode, msg.DeletedCount)
				repo.Activity = pruning
			} else {
				repo.Activity = activity
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
			case domain.PruningActivity:
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
	var s strings.Builder
	s.WriteString(common.FormatHeader(len(m.Repositories)))

	if len(m.Repositories) == 0 {
		s.WriteString("No repositories found")
		return s.String()
	}

	s.WriteString(GenerateTable(m.Repositories, m.Cursor, m.width))
	s.WriteString("\n\n")

	s.WriteString(buildFooter())

	legend := RenderLegend(m.ShowLegend)
	s.WriteString("\n\n")
	s.WriteString(legend)

	return s.String()
}

func buildFooter() string {
	hotkeys := []string{
		"↑/↓ navigate",
		"r refresh",
		"p pull all updates",
		"b prune merged branches",
		"? toggle legend",
		"q quit",
	}
	footerText := strings.Join(hotkeys, "  •  ")
	return common.FooterStyle.Render(footerText)
}

func isBusy(repo domain.Repository) bool {
	switch a := repo.Activity.(type) {
	case domain.IdleActivity:
		return false
	case domain.RefreshingActivity:
		return !a.Complete
	case domain.PullingActivity:
		return !a.Complete
	case domain.PruningActivity:
		return !a.Complete
	default:
		return false
	}
}

func shouldPull(repo domain.Repository) bool {
	switch s := repo.RemoteState.(type) {
	case domain.Behind:
		return s.Count > 0
	case domain.Diverged:
		return s.BehindCount > 0
	default:
		return false
	}
}
