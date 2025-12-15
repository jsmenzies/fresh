package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	switch m.State {
	case Scanning:
		pad := strings.Repeat(" ", Padding)
		header := fmt.Sprintf("%s Scanning for Git projects... Found %d repositories", m.Spinner.View(), len(m.Repositories))

		var repoList strings.Builder
		startIndex := len(m.Repositories) - 6
		if startIndex < 0 {
			startIndex = 0
		}

		for i := startIndex; i < len(m.Repositories); i++ {
			repo := m.Repositories[i]
			var name = lipgloss.NewStyle().Foreground(Green)
			var text = lipgloss.NewStyle().Foreground(TextPrimary).PaddingLeft(1)
			repoList.WriteString(fmt.Sprintf("%s %s \n", text.Render("\uF061 Found git repository:"), name.Render(repo.Name)))
		}

		return "\n" +
			pad + header + "\n\n" +
			repoList.String() + "\n"

	case Listing:
		if m.Choice != "" {
			return quitTextStyle.Render(fmt.Sprintf("Selected: %s", m.Choice))
		}

		var style = lipgloss.NewStyle().Foreground(Blue)
		var text = style.Render(fmt.Sprintf("\nScan complete. %d repositories found\n", len(m.Repositories)))
		headerLine := text

		if len(m.Repositories) == 0 {
			return headerLine + "No repositories found"
		}

		tableView := GenerateTable(m.Repositories, m.Cursor)
		footer := buildFooter()

		return headerLine + tableView + "\n" + footer

	case Quitting:
		return quitTextStyle.Render("Goodbye!")

	default:
		return "Loading..."
	}
}

func buildFooter() string {
	keyStyle := lipgloss.NewStyle().Foreground(SubtleGray)

	hotkeys := []string{
		"↑/↓ navigate",
		"r refresh",
		"u update all",
		"q quit",
	}

	footerText := strings.Join(hotkeys, "  •  ")
	return "\n" + keyStyle.Render(footerText)
}
