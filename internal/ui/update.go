package ui

import (
	"fresh/internal/domain"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.State = Quitting
			return m, tea.Quit

		case "up", "k":
			if m.State == Listing && m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil

		case "down", "j":
			if m.State == Listing && m.Cursor < len(m.Repositories)-1 {
				m.Cursor++
			}
			return m, nil

		case "enter":
			if m.State == Listing {
				if m.Cursor < len(m.Repositories) {
					m.Choice = m.Repositories[m.Cursor].Name
				}
				return m, tea.Quit
			}
		}

		if m.State == Listing {
			switch {
			case key.Matches(msg, m.Keys.refresh):
				if m.Cursor < len(m.Repositories) {
					repo := m.Repositories[m.Cursor]
					return m, performPull(m.Cursor, repo.Path)
				}
			case key.Matches(msg, m.Keys.updateAll):
				return m, nil
			}
		}

	case scanTickMsg:
		if m.State == Scanning {
			return m, tea.Batch(tickCmd(), scanStep(m.Scanner))
		}
		return m, nil

	case repoFoundMsg:
		repo := domain.Repository(msg)
		m.Repositories = append(m.Repositories, repo)
		return m, nil

	case scanCompleteMsg:
		m.State = Listing

		for i := range m.Repositories {
			m.Repositories[i].Fetching = false
			m.Repositories[i].Done = false
			m.Repositories[i].Refreshing = false
			m.Repositories[i].HasRemoteUpdates = false
			m.Repositories[i].RefreshSpinner = newDotSpinner()
		}

		return m, startBackgroundRefresh(m.Repositories)

	case refreshStartMsg:
		if m.State == Listing && msg.repoIndex < len(m.Repositories) {
			m.Repositories[msg.repoIndex].Refreshing = true
			m.Repositories[msg.repoIndex].RefreshSpinner, _ = m.Repositories[msg.repoIndex].RefreshSpinner.Update(nil)

			return m, tea.Batch(
				performRefresh(msg.repoIndex, msg.repoPath),
				m.Repositories[msg.repoIndex].RefreshSpinner.Tick,
			)
		}
		return m, nil

	case refreshCompleteMsg:
		if m.State == Listing && msg.repoIndex < len(m.Repositories) {
			m.Repositories[msg.repoIndex].Refreshing = false
			m.Repositories[msg.repoIndex].AheadCount = msg.aheadCount
			m.Repositories[msg.repoIndex].BehindCount = msg.behindCount
			m.Repositories[msg.repoIndex].HasRemoteUpdates = msg.behindCount > 0
			m.Repositories[msg.repoIndex].HasError = msg.hasError
			m.Repositories[msg.repoIndex].ErrorMessage = msg.errorMessage
		}
		return m, nil

	case pullStartMsg:
		if m.State == Listing && msg.repoIndex < len(m.Repositories) {
			m.Repositories[msg.repoIndex].PullState = domain.NewPullState()
			m.Repositories[msg.repoIndex].PullSpinner = newDotSpinner()

			return m, m.Repositories[msg.repoIndex].PullSpinner.Tick
		}
		return m, nil

	case pullWorkState:
		return m, listenForPullProgress(msg)

	case pullLineMsg:
		if m.State == Listing && msg.repoIndex < len(m.Repositories) {
			// Add line to PullState
			if m.Repositories[msg.repoIndex].PullState != nil {
				m.Repositories[msg.repoIndex].PullState.AddLine(msg.line)
			}
		}

		// Continue listening
		if msg.state != nil {
			return m, listenForPullProgress(*msg.state)
		}
		return m, nil

	case pullCompleteMsg:
		if m.State == Listing && msg.repoIndex < len(m.Repositories) {
			// Mark pull as complete with exit code
			if m.Repositories[msg.repoIndex].PullState != nil {
				m.Repositories[msg.repoIndex].PullState.Complete(msg.exitCode)
			}

			// If successful, update ahead/behind counts
			if msg.exitCode == 0 {
				m.Repositories[msg.repoIndex].BehindCount = 0
				m.Repositories[msg.repoIndex].HasRemoteUpdates = false
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.State == Scanning {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		} else if m.State == Listing {
			var cmds []tea.Cmd
			for i := range m.Repositories {
				if m.Repositories[i].Refreshing {
					var cmd tea.Cmd
					m.Repositories[i].RefreshSpinner, cmd = m.Repositories[i].RefreshSpinner.Update(msg)
					cmds = append(cmds, cmd)
				}
				// Update pull spinner if pulling
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
	}

	return m, nil
}
