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
	refresh    key.Binding
	updateAll  key.Binding
	pull       key.Binding
	pullAll    key.Binding
	toggleHelp key.Binding
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
		toggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

type Model struct {
	Repositories   []domain.Repository
	Cursor         int
	Keys           *listKeyMap
	width, height  int
	ShowFullLegend bool
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

		case key.Matches(msg, m.Keys.toggleHelp):
			m.ShowFullLegend = !m.ShowFullLegend
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
	headerLine := common.FormatHeader(len(m.Repositories))

	if len(m.Repositories) == 0 {
		return headerLine + "No repositories found"
	}

	tableView := common.GenerateTable(m.Repositories, m.Cursor)

	var legend string
	if m.ShowFullLegend {
		legend = buildFullLegend()
	} else if m.Cursor < len(m.Repositories) {
		legend = buildContextualLegend(m.Repositories[m.Cursor])
	}

	footer := buildFooter()

	return headerLine + tableView + "\n" + legend + "\n\n" + footer
}

func buildContextualLegend(repo domain.Repository) string {
	var items []string

	// Local Status
	switch s := repo.LocalState.(type) {
	case domain.DirtyLocalState:
		if s.Untracked > 0 {
			items = append(items, common.LocalStatusUntrackedItem.Render(common.IconUntracked)+" Untracked")
		}
		if s.Modified > 0 || s.Added > 0 || s.Deleted > 0 {
			items = append(items, common.LocalStatusDirtyItem.Render("~")+" Modified")
		}
	case domain.LocalStateError:
		// Handle error if needed
	default:
		items = append(items, common.TextGreen.Render(common.IconClean)+" Clean")
	}

	// Remote Status
	switch repo.RemoteState.(type) {
	case domain.Ahead:
		items = append(items, common.TextBlue.Render(common.IconAhead)+" Ahead")
	case domain.Behind:
		items = append(items, common.TextBlue.Render(common.IconBehind)+" Behind")
	case domain.Diverged:
		items = append(items, common.TextBlue.Render(common.IconAhead)+" Ahead")
		items = append(items, common.TextBlue.Render(common.IconBehind)+" Behind")
	case domain.NoUpstream, domain.DetachedRemote, domain.RemoteError:
		items = append(items, common.RemoteStatusErrorText.Render(common.IconRemoteError)+" No Upstream")
	default:
		items = append(items, common.TextSubtleGreen.Render(common.IconSynced)+" Synced")
	}

	return common.FooterStyle.Render(strings.Join(items, "  •  "))
}

func buildFullLegend() string {
	items := []string{
		common.LocalStatusUntrackedItem.Render(common.IconUntracked) + " Untracked",
		common.LocalStatusDirtyItem.Render("~") + " Modified",
		common.LocalStatusDirtyItem.Render(common.IconDirty) + " Dirty",
		common.TextGreen.Render(common.IconClean) + " Clean",
		common.TextSubtleGreen.Render(common.IconSynced) + " Synced",
		common.TextBlue.Render(common.IconAhead) + " Ahead",
		common.TextBlue.Render(common.IconBehind) + " Behind",
		common.RemoteStatusErrorText.Render(common.IconRemoteError) + " No Upstream",
	}
	legendText := strings.Join(items, "  •  ")
	return common.FooterStyle.Render(legendText)
}

func buildFooter() string {
	hotkeys := []string{
		"↑/↓ navigate",
		"r refresh",
		"ctrl+r refresh all",
		"p pull",
		"ctrl+p pull all",
		"? toggle help",
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
