package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type listKeyMap struct {
	refresh      key.Binding
	pullAll      key.Binding
	pruneAll     key.Binding
	checkoutDev  key.Binding
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
		checkoutDev: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "checkout develop/dev"),
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
		repos[i].Activity = &domain.IdleActivity{}
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
		repo.Activity = &domain.RefreshingActivity{
			Spinner: common.NewRefreshSpinner(),
		}
		cmds = append(cmds, performRefresh(i, repo.Path))
		cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick)
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.Keys.refresh):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !repo.IsBusy() {
					repo.Activity = &domain.RefreshingActivity{
						Spinner: common.NewRefreshSpinner(),
					}
					cmds = append(cmds, performRefresh(i, repo.Path))
					cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pullAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !repo.IsBusy() && repo.CanPull() {
					repo.Activity = &domain.PullingActivity{
						Spinner: common.NewPullSpinner(),
					}
					cmds = append(cmds, performPull(i, repo.Path))
					cmds = append(cmds, repo.Activity.(*domain.PullingActivity).Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.pruneAll):
			var cmds []tea.Cmd
			for i := range m.Repositories {
				repo := &m.Repositories[i]
				if !repo.IsBusy() && len(repo.Branches.Merged) > 0 {
					repo.Activity = &domain.PruningActivity{
						Spinner: common.NewPullSpinner(),
					}
					cmds = append(cmds, performPrune(i, repo.Path, repo.Branches.Merged))
					cmds = append(cmds, repo.Activity.(*domain.PruningActivity).Spinner.Tick)
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.Keys.checkoutDev):
			if m.Cursor >= 0 && m.Cursor < len(m.Repositories) {
				repo := &m.Repositories[m.Cursor]
				if !repo.IsBusy() {
					repo.Activity = &domain.CheckoutActivity{
						Spinner: common.NewPullSpinner(),
					}
					return m, tea.Batch(
						performCheckoutIntegration(m.Cursor, repo.Path),
						repo.Activity.(*domain.CheckoutActivity).Spinner.Tick,
					)
				}
			}
			return m, nil

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

			if refreshing, ok := activity.(*domain.RefreshingActivity); ok {
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
			if pulling, ok := repo.Activity.(*domain.PullingActivity); ok {
				pulling.AddLine(msg.line)
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
			if pulling, ok := activity.(*domain.PullingActivity); ok {
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
			if pruning, ok := repo.Activity.(*domain.PruningActivity); ok {
				pruning.AddLine(msg.line)
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
			if pruning, ok := activity.(*domain.PruningActivity); ok {
				pruning.MarkComplete(msg.exitCode, msg.DeletedCount)
				repo.Activity = pruning
			} else {
				repo.Activity = activity
			}
		}

	case checkoutWorkState:
		return m, listenForCheckoutProgress(msg)

	case checkoutLineMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			if checkout, ok := repo.Activity.(*domain.CheckoutActivity); ok {
				checkout.AddLine(msg.line)
			}
		}
		if msg.state != nil {
			return m, listenForCheckoutProgress(*msg.state)
		}

	case checkoutCompleteMsg:
		if msg.Index < len(m.Repositories) {
			repo := &m.Repositories[msg.Index]
			activity := repo.Activity
			*repo = msg.Repo
			if checkout, ok := activity.(*domain.CheckoutActivity); ok {
				checkout.MarkComplete(msg.exitCode, msg.targetBranch)
				repo.Activity = checkout
			} else {
				repo.Activity = activity
			}
		}

	case spinner.TickMsg:
		var cmds []tea.Cmd
		for i := range m.Repositories {
			switch activity := m.Repositories[i].Activity.(type) {
			case *domain.RefreshingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			case *domain.PullingActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			case *domain.PruningActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			case *domain.CheckoutActivity:
				if !activity.Complete {
					var cmd tea.Cmd
					activity.Spinner, cmd = activity.Spinner.Update(msg)
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
		"d checkout develop/dev",
		"? toggle legend",
		"q quit",
	}
	footerText := strings.Join(hotkeys, "  •  ")
	return common.FooterStyle.Render(footerText)
}
