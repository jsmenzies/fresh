package scanning

import (
	"fmt"
	"fresh/internal/config"
	"fresh/internal/domain"
	"fresh/internal/git"
	"fresh/internal/scanner"
	"fresh/internal/ui/views/common"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var cfg = config.DefaultConfig()

type Model struct {
	Repositories  []domain.Repository
	scanner       *scanner.Scanner
	Spinner       spinner.Model
	width, height int
}

func New(scanDir string) *Model {
	s := common.NewGreenDotSpinner()
	return &Model{
		Repositories: make([]domain.Repository, 0),
		scanner:      scanner.New(scanDir),
		Spinner:      s,
	}
}

func (m *Model) Init() tea.Cmd {
	go m.scanner.Scan()
	return tea.Batch(
		waitForRepo(m.scanner.GetRepoChannel()),
		m.Spinner.Tick,
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case repoFoundMsg:
		path := string(msg)
		repo := git.BuildRepository(path, cfg)
		m.Repositories = append(m.Repositories, repo)
		return m, waitForRepo(m.scanner.GetRepoChannel())

	case scanCompleteMsg:
		return m, func() tea.Msg {
			return ScanFinishedMsg{Repos: m.Repositories}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) View() string {
	pad := strings.Repeat(" ", common.Padding)
	header := fmt.Sprintf("%s Scanning for Git projects... Found %d repositories", m.Spinner.View(), len(m.Repositories))

	var repoList strings.Builder
	startIndex := len(m.Repositories) - 6
	if startIndex < 0 {
		startIndex = 0
	}

	for i := startIndex; i < len(m.Repositories); i++ {
		repo := m.Repositories[i]
		repoList.WriteString(common.FormatScanningFound(repo.Name) + "\n")
	}

	return "\n" +
		pad + header + "\n\n" +
		repoList.String() + "\n"
}

func waitForRepo(c chan string) tea.Cmd {
	return func() tea.Msg {
		path, ok := <-c
		if !ok {
			return scanCompleteMsg{}
		}
		return repoFoundMsg(path)
	}
}
