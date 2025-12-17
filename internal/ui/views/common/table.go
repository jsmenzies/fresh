package common

import (
	"fmt"
	"strings"

	"fresh/internal/domain"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func GenerateTable(repositories []domain.Repository, cursor int) string {
	headers := []string{"", "PROJECT", "BRANCH", "LOCAL", "REMOTE", "LINKS", "", "LAST COMMIT"}

	rows := make([][]string, len(repositories))
	for i, repo := range repositories {
		isSelected := i == cursor

		rows[i] = repositoryToRow(repo, isSelected)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(rows...).
		BorderStyle(lipgloss.NewStyle().Foreground(DividerColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle
			}
			return lipgloss.NewStyle()
		})

	return t.Render()
}

func repositoryToRow(repo domain.Repository, isSelected bool) []string {
	selector := buildSelector(isSelected)
	projectName := buildProjectName(repo.Name, isSelected)
	branchName := buildBranchName(repo.CurrentBranch)
	localCol := buildLocalStatus(repo)
	remoteCol := buildRemoteStatus(repo)
	linksCol := buildLinks(repo)
	badgeCol := buildBadge(repo)
	lastUpdateCol := buildLastUpdate(repo)

	return []string{
		selector,
		projectName,
		branchName,
		localCol,
		remoteCol,
		linksCol,
		badgeCol,
		lastUpdateCol,
	}
}

func truncateWithEllipsis(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

func stylePullOutput(lastLine string, exitCode int) string {
	lowerLine := strings.ToLower(lastLine)
	truncated := truncateWithEllipsis(lastLine, 60)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
		return PullOutputError.Render(truncated)
	}

	if exitCode == 0 && (strings.Contains(lowerLine, "done") ||
		strings.Contains(lowerLine, "up to date") ||
		strings.Contains(lowerLine, "up-to-date") ||
		strings.Contains(lowerLine, "file") && strings.Contains(lowerLine, "changed")) {
		return PullOutputSuccess.Render(truncated)
	}

	return PullOutputWarn.Render(truncated)
}

func buildLastUpdate(repo domain.Repository) string {
	timeAgo := FormatTimeAgo(repo.LastCommitTime)

	// If we have pull state data, display it
	if repo.PullState != nil {
		if repo.PullState.InProgress {
			lastLine := repo.PullState.GetLastLine()
			truncated := truncateWithEllipsis(lastLine, 55)
			return repo.PullSpinner.View() + " " + truncated
		} else if repo.PullState.Completed {
			lastLine := repo.PullState.GetLastLine()
			styledLine := stylePullOutput(lastLine, repo.PullState.ExitCode)
			return styledLine
		}
	}

	// No pull state - show refresh status or time
	if repo.Refreshing {
		return repo.RefreshSpinner.View()
	}

	return TimeAgoStyle.Render(IconClock + " " + timeAgo)
}

func buildBadge(repo domain.Repository) string {
	return BadgeStyle.Render("")
}

func buildLinks(repo domain.Repository) string {
	if repo.RemoteURL != "" {
		if IsGitHubRepository(repo.RemoteURL) {
			githubURLs := BuildGitHubURLs(repo.RemoteURL, repo.CurrentBranch)
			if githubURLs != nil {
				var shortcuts []string

				codeLink := MakeClickableURL(githubURLs["code"], IconCode)
				shortcuts = append(shortcuts, LinkStyle.Render(codeLink))

				prsLink := MakeClickableURL(githubURLs["prs"], IconPullRequests)
				shortcuts = append(shortcuts, LinkStyle.Render(prsLink))

				openPRLink := MakeClickableURL(githubURLs["openpr"], IconOpenPR)
				shortcuts = append(shortcuts, LinkStyle.Render(openPRLink))

				shortcutsDisplay := fmt.Sprintf("%s", strings.Join(shortcuts, " "))
				return LinksStyle.Render(shortcutsDisplay)
			}
		}
	}
	return ""
}

func buildRemoteStatus(repo domain.Repository) string {
	if repo.BehindCount > 0 && repo.AheadCount > 0 {
		return RemoteStatusYellow.Render(IconDiverged + " " + StatusDiverged)
	}
	if repo.BehindCount > 0 {
		return RemoteStatusBlue.Render(fmt.Sprintf("%s %s", IconBehind, StatusBehind))
	}
	if repo.AheadCount > 0 {
		return RemoteStatusGreen.Render(fmt.Sprintf("%s %s", IconAhead, StatusAhead))
	}
	return RemoteStatusSynced
}

func buildLocalStatus(repo domain.Repository) string {
	if repo.HasError {
		return LocalStatusUntracked
	} else if repo.HasModified {
		return LocalStatusDirty
	}
	return LocalStatusClean
}

func buildBranchName(branch string) string {
	branchName := branch
	if branchName == "" {
		branchName = BranchNameEmpty
	} else {
		branchName = BranchNameStyle.Render(branchName)
	}
	return branchName
}

func buildProjectName(name string, isSelected bool) string {
	if isSelected {
		return ProjectNameStyle.Copy().Bold(true).Render(name)
	}
	return ProjectNameStyle.Render(name)
}

func buildSelector(isSelected bool) string {
	if isSelected {
		return SelectorStyle.Render("▶")
	}
	return SelectorStyle.Render(" ")
}

func MakeClickableURL(url string, displayText string) string {
	if url == "" {
		return displayText
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, displayText)
}
