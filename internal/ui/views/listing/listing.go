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
	refresh          key.Binding
	updateAll        key.Binding
	pull             key.Binding
	pullAll          key.Binding
	prune            key.Binding
	pruneAll         key.Binding
	pruneSquashed    key.Binding
	pruneSquashedAll key.Binding
	toggleLegend     key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh remote"),
		),
		updateAll: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh all"),
		),
		pull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull"),
		),
		pullAll: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "pull all"),
		),
		prune: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "prune merged"),
		),
		pruneAll: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "prune all"),
		),
		pruneSquashed: key.NewBinding(
			key.WithKeys("B"),
			key.WithHelp("shift+b", "prune squashed"),
		),
		pruneSquashedAll: key.NewBinding(
			key.WithKeys("ctrl+B"),
			key.WithHelp("ctrl+shift+b", "prune squashed all"),
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
			if m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) {
					refreshing := domain.RefreshingActivity{
						Spinner: common.NewRefreshSpinner(),
					}
					repo.Activity = refreshing
					return m, tea.Batch(
						performRefresh(m.Cursor, repo.Path),
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
					cmds = append(cmds, performRefresh(i, repo.Path))
					cmds = append(cmds, refreshing.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pull):
			if m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) && canPull(*repo) {
					pulling := domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pulling
					return m, tea.Batch(
						performPull(m.Cursor, repo.Path),
						pulling.Spinner.Tick,
					)
				}
			}
		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) && shouldPull(*repo) {
					pulling := domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pulling
					cmds = append(cmds, performPull(i, repo.Path))
					cmds = append(cmds, pulling.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.prune):
			if m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) && canPrune(*repo) && len(repo.MergedBranches) > 0 {
					pruning := domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pruning
					return m, tea.Batch(
						performPrune(m.Cursor, repo.Path, repo.MergedBranches),
						pruning.Spinner.Tick,
					)
				}
			}

		case key.Matches(msg, m.Keys.pruneAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) && canPrune(*repo) && len(repo.MergedBranches) > 0 {
					pruning := domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pruning
					cmds = append(cmds, performPrune(i, repo.Path, repo.MergedBranches))
					cmds = append(cmds, pruning.Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pruneSquashed):
			if m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !isBusy(*repo) && canPrune(*repo) && len(repo.SquashedBranches) > 0 {
					pruning := domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pruning
					return m, tea.Batch(
						performPruneSquashed(m.Cursor, repo.Path, repo.SquashedBranches),
						pruning.Spinner.Tick,
					)
				}
			}

		case key.Matches(msg, m.Keys.pruneSquashedAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !isBusy(*repo) && canPrune(*repo) && len(repo.SquashedBranches) > 0 {
					pruning := domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
						Lines:   make([]string, 0),
					}
					repo.Activity = pruning
					cmds = append(cmds, performPruneSquashed(i, repo.Path, repo.SquashedBranches))
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

	s.WriteString(GenerateTable(m.Repositories, m.Cursor))
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
		"ctrl+r refresh all",
		"p pull",
		"ctrl+p pull all",
		"b prune merged",
		"ctrl+b prune merged all",
		"shift+b prune squashed",
		"ctrl+shift+b prune squashed all",
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

func canPull(repo domain.Repository) bool {
	switch repo.RemoteState.(type) {
	case domain.NoUpstream, domain.DetachedRemote, domain.RemoteError:
		return false
	default:
		return true
	}
}

func canPrune(repo domain.Repository) bool {
	// Only prune if on a proper branch (not detached HEAD)
	if _, ok := repo.Branch.(domain.OnBranch); !ok {
		return false
	}
	return true
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
