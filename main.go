package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type appState int

const (
	scanning appState = iota
	listing
	quitting
)

type scanProgressMsg struct {
	reposFound int
}

type repoFoundMsg repository

type scanCompleteMsg []repository

type scanTickMsg time.Time

type refreshStartMsg struct {
	repoIndex int
	repoPath  string
}

type refreshCompleteMsg struct {
	repoIndex    int
	aheadCount   int
	behindCount  int
	hasError     bool
	errorMessage string
}

type pullStartMsg struct {
	repoIndex int
}

type pullLineMsg struct {
	repoIndex int
	line      string
	state     *pullWorkState
}

type pullCompleteMsg struct {
	repoIndex int
	exitCode  int
}

type model struct {
	state        appState
	spinner      spinner.Model
	cursor       int
	keys         *listKeyMap
	choice       string
	repositories []repository
	scanDir      string
}

func (m model) Init() tea.Cmd {
	return tea.Batch(startScanning(m.scanDir), tickCmd(), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.state = quitting
			return m, tea.Quit

		case "up", "k":
			if m.state == listing && m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.state == listing && m.cursor < len(m.repositories)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if m.state == listing {
				if m.cursor < len(m.repositories) {
					m.choice = m.repositories[m.cursor].name
				}
				return m, tea.Quit
			}

		}

		// Handle custom key bindings
		if m.state == listing {
			switch {
			case key.Matches(msg, m.keys.refresh):
				if m.cursor < len(m.repositories) {
					repo := m.repositories[m.cursor]
					return m, performPull(m.cursor, repo.path)
				}
			case key.Matches(msg, m.keys.updateAll):
				return m, nil
			}
		}

	case repoFoundMsg:
		repo := repository(msg)
		m.repositories = append(m.repositories, repo)
		return m, nil

	case scanCompleteMsg:
		m.state = listing

		for i := range m.repositories {
			m.repositories[i].fetching = false
			m.repositories[i].done = false
			m.repositories[i].refreshing = false
			m.repositories[i].hasRemoteUpdates = false
			m.repositories[i].refreshSpinner = spinner.New()
			m.repositories[i].refreshSpinner.Spinner = spinner.Dot
		}

		return m, startBackgroundRefresh(m.repositories)

	case scanTickMsg:
		if m.state == scanning {
			return m, tea.Batch(tickCmd(), scanStep())
		}
		return m, nil

	case refreshStartMsg:
		if m.state == listing && msg.repoIndex < len(m.repositories) {
			m.repositories[msg.repoIndex].refreshing = true
			m.repositories[msg.repoIndex].refreshSpinner, _ = m.repositories[msg.repoIndex].refreshSpinner.Update(nil)

			return m, tea.Batch(
				performRefresh(msg.repoIndex, msg.repoPath),
				m.repositories[msg.repoIndex].refreshSpinner.Tick,
			)
		}
		return m, nil

	case refreshCompleteMsg:
		if m.state == listing && msg.repoIndex < len(m.repositories) {
			m.repositories[msg.repoIndex].refreshing = false
			m.repositories[msg.repoIndex].aheadCount = msg.aheadCount
			m.repositories[msg.repoIndex].behindCount = msg.behindCount
			m.repositories[msg.repoIndex].hasRemoteUpdates = msg.behindCount > 0
			m.repositories[msg.repoIndex].hasError = msg.hasError
			m.repositories[msg.repoIndex].errorMessage = msg.errorMessage
		}
		return m, nil

	case pullStartMsg:
		if m.state == listing && msg.repoIndex < len(m.repositories) {
			m.repositories[msg.repoIndex].pullState = NewPullState()

			// Initialize pull spinner
			m.repositories[msg.repoIndex].pullSpinner = spinner.New()
			m.repositories[msg.repoIndex].pullSpinner.Spinner = spinner.Dot

			return m, m.repositories[msg.repoIndex].pullSpinner.Tick
		}
		return m, nil

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		if m.state == listing && msg.repoIndex < len(m.repositories) {
			// Add line to PullState
			if m.repositories[msg.repoIndex].pullState != nil {
				m.repositories[msg.repoIndex].pullState.AddLine(msg.line)
			}
		}

		// Continue listening
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}
		return m, nil

	case pullCompleteMsg:
		if m.state == listing && msg.repoIndex < len(m.repositories) {
			// Mark pull as complete with exit code
			if m.repositories[msg.repoIndex].pullState != nil {
				m.repositories[msg.repoIndex].pullState.Complete(msg.exitCode)
			}

			// If successful, update ahead/behind counts
			if msg.exitCode == 0 {
				m.repositories[msg.repoIndex].behindCount = 0
				m.repositories[msg.repoIndex].hasRemoteUpdates = false
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.state == scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		} else if m.state == listing {
			var cmds []tea.Cmd
			for i := range m.repositories {
				if m.repositories[i].refreshing {
					var cmd tea.Cmd
					m.repositories[i].refreshSpinner, cmd = m.repositories[i].refreshSpinner.Update(msg)
					cmds = append(cmds, cmd)
				}
				// Update pull spinner if pulling
				if m.repositories[i].pullState != nil && m.repositories[i].pullState.InProgress {
					var cmd tea.Cmd
					m.repositories[i].pullSpinner, cmd = m.repositories[i].pullSpinner.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
			if len(cmds) > 0 {
				return m, tea.Batch(cmds...)
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case scanning:
		pad := strings.Repeat(" ", Padding)
		header := fmt.Sprintf("%s Scanning for Git projects... Found %d repositories", m.spinner.View(), len(m.repositories))

		// Show only the last 4 repositories for scrolling effect
		var repoList strings.Builder
		startIndex := len(m.repositories) - 7
		if startIndex < 0 {
			startIndex = 0
		}

		for i := startIndex; i < len(m.repositories); i++ {
			repo := m.repositories[i]
			var name = lipgloss.NewStyle().Foreground(Green)
			var text = lipgloss.NewStyle().Foreground(TextPrimary).PaddingLeft(1)
			repoList.WriteString(fmt.Sprintf("%s %s \n", text.Render("\uF061 Found git repository:"), name.Render(repo.name)))
		}

		return "\n" +
			pad + header + "\n\n" +
			repoList.String() + "\n"

	case listing:
		if m.choice != "" {
			return quitTextStyle.Render(fmt.Sprintf("Selected: %s", m.choice))
		}

		var style = lipgloss.NewStyle().Foreground(Blue)
		var text = style.Render(fmt.Sprintf("\nScan complete. %d repositories found\n", len(m.repositories)))
		headerLine := text

		if len(m.repositories) == 0 {
			return headerLine + "No repositories found"
		}

		// Build and render lipgloss table
		tableView := generateTable(m.repositories, m.cursor)
		return headerLine + tableView

	case quitting:
		return quitTextStyle.Render("Goodbye!")

	default:
		return "Loading..."
	}
}

var (
	foundRepositories []repository
	directoriesToScan []string
	currentScanIndex  int
)

func startScanning(scanDir string) tea.Cmd {
	return func() tea.Msg {
		directoriesToScan = []string{}
		foundRepositories = []repository{}
		currentScanIndex = 0

		// Collect directories to scan with depth limit of 0 (one level only)
		scanDirectoriesWithDepth(scanDir, 0, 0)

		return scanProgressMsg{reposFound: 0}
	}
}

func scanStep() tea.Cmd {
	return func() tea.Msg {
		if currentScanIndex >= len(directoriesToScan) {
			return scanCompleteMsg(foundRepositories)
		}

		path := directoriesToScan[currentScanIndex]
		currentScanIndex++

		if isGitRepository(path) {
			repoName := filepath.Base(path)
			lastCommitTime := getLastCommitTime(path)
			remoteURL := getRemoteURL(path)
			aheadCount, behindCount := getGitStatus(path)
			hasModified := hasModifiedFiles(path)
			currentBranch := getCurrentBranch(path)

			repo := repository{
				name:           repoName,
				path:           path,
				lastCommitTime: lastCommitTime,
				remoteURL:      remoteURL,
				hasModified:    hasModified,
				aheadCount:     aheadCount,
				behindCount:    behindCount,
				currentBranch:  currentBranch,
			}
			foundRepositories = append(foundRepositories, repo)
			return repoFoundMsg(repo)
		}

		return scanProgressMsg{
			reposFound: len(foundRepositories),
		}
	}
}

func scanDirectoriesWithDepth(dir string, currentDepth, maxDepth int) {
	if currentDepth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(dir, entry.Name())
			// Only add directories at the first level (depth 0)
			if currentDepth == 0 {
				directoriesToScan = append(directoriesToScan, fullPath)
			}
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return scanTickMsg(t)
	})
}

func startBackgroundRefresh(repos []repository) tea.Cmd {
	var cmds []tea.Cmd
	for i, repo := range repos {
		cmds = append(cmds, func(repoIndex int, repoPath string) tea.Cmd {
			return func() tea.Msg {
				return refreshStartMsg{
					repoIndex: repoIndex,
					repoPath:  repoPath,
				}
			}
		}(i, repo.path))
	}
	return tea.Batch(cmds...)
}

func performRefresh(repoIndex int, repoPath string) tea.Cmd {
	return func() tea.Msg {
		ahead, behind, err := refreshRemoteStatus(repoPath)

		if err != nil {
			return refreshCompleteMsg{
				repoIndex:    repoIndex,
				aheadCount:   0,
				behindCount:  0,
				hasError:     true,
				errorMessage: err.Error(),
			}
		}

		return refreshCompleteMsg{
			repoIndex:    repoIndex,
			aheadCount:   ahead,
			behindCount:  behind,
			hasError:     false,
			errorMessage: "",
		}
	}
}

type pullWorkState struct {
	repoIndex int
	lineChan  chan string
	doneChan  chan pullCompleteMsg
}

func performPull(repoIndex int, repoPath string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return pullStartMsg{repoIndex: repoIndex}
		},
		func() tea.Msg {
			// Create channels for lines and completion
			lineChan := make(chan string, 10)
			doneChan := make(chan pullCompleteMsg, 1)

			// Start git pull in a goroutine
			go func() {
				exitCode := gitPull(repoPath, func(line string) {
					lineChan <- line
				})

				close(lineChan)

				doneChan <- pullCompleteMsg{
					repoIndex: repoIndex,
					exitCode:  exitCode,
				}
				close(doneChan)
			}()

			// Start listening for line updates
			return pullWorkState{
				repoIndex: repoIndex,
				lineChan:  lineChan,
				doneChan:  doneChan,
			}
		},
	)
}

func listenForPullProgress(state pullWorkState) tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-state.lineChan:
			if ok {
				return pullLineMsg{
					repoIndex: state.repoIndex,
					line:      line,
					state:     &state,
				}
			}
			return <-state.doneChan
		case complete := <-state.doneChan:
			return complete
		}
	}
}

func getRandomSpinner() spinner.Spinner {
	spinners := []spinner.Spinner{
		spinner.Dot,
		spinner.MiniDot,
		spinner.Points,
		spinner.Meter,
	}
	return spinners[rand.Intn(len(spinners))]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Get directory from command line argument or use current directory
	var scanDir string
	if len(os.Args) > 1 {
		scanDir = os.Args[1]
	} else {
		var err error
		scanDir, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate directory exists
	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %s\n", scanDir)
		os.Exit(1)
	}

	s := spinner.New()
	s.Spinner = getRandomSpinner()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))

	keys := newListKeyMap()

	m := model{
		state:   scanning,
		spinner: s,
		cursor:  0,
		keys:    keys,
		scanDir: scanDir,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
