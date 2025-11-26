package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func generateTable(repositories []repository, cursor int) string {
	headers := []string{"PROJECT", "BRANCH", "LOCAL", "REMOTE", "LINKS", "", "STATUS / UPDATE"}

	rows := make([][]string, len(repositories))
	for i, repo := range repositories {
		rows[i] = repositoryToRow(repo)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(rows...)

	t = t.
		BorderStyle(lipgloss.NewStyle().Foreground(DividerColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle
			}
			return lipgloss.NewStyle()
			//.MarginBottom(1)
		})

	return t.Render()
}

func repositoryToRow(repo repository) []string {
	projectName := buildProjectName(repo.name)
	branchName := buildBranchName(repo.currentBranch)
	localCol := buildLocalStatus(repo)
	remoteCol := buildRemoteStatus(repo)
	linksCol := buildLinks(repo)
	badgeCol := buildBadge(repo)
	lastUpdateCol := buildLastUpdate(repo)

	return []string{
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

func buildLastUpdate(repo repository) string {
	timeAgo := formatTimeAgo(repo.lastCommitTime)

	// If we have pull state data, display it
	if repo.pullState != nil {
		if repo.pullState.InProgress {
			lastLine := repo.pullState.GetLastLine()
			truncated := truncateWithEllipsis(lastLine, 55)
			return repo.pullSpinner.View() + " " + truncated
		} else if repo.pullState.Completed {
			lastLine := repo.pullState.GetLastLine()
			styledLine := stylePullOutput(lastLine, repo.pullState.ExitCode)
			return styledLine
		}
	}

	// No pull state - show refresh status or time
	if repo.refreshing {
		return repo.refreshSpinner.View()
	}

	return TimeAgoStyle.Render(IconClock + " " + timeAgo)
}

func buildBadge(repo repository) string {
	//if repo.behindCount > 0 {
	//	return BadgeReadyStyle.Render(BadgeReady)
	//}
	return BadgeStyle.Render("")

}

func buildLinks(repo repository) string {
	if repo.remoteURL != "" {
		if isGitHubRepository(repo.remoteURL) {
			githubURLs := buildGitHubURLs(repo.remoteURL, repo.currentBranch)
			if githubURLs != nil {
				var shortcuts []string

				// Code link (to current branch)
				codeLink := makeClickableURL(githubURLs["code"], IconCode)
				shortcuts = append(shortcuts, LinksStyles.Render(codeLink))

				// PRs link
				prsLink := makeClickableURL(githubURLs["prs"], IconPullRequests)
				shortcuts = append(shortcuts, LinksStyles.Render(prsLink))

				// Open PR link
				openPRLink := makeClickableURL(githubURLs["openpr"], IconOpenPR)
				shortcuts = append(shortcuts, LinksStyles.Render(openPRLink))

				shortcutsDisplay := fmt.Sprintf("%s", strings.Join(shortcuts, " "))
				return shortcutsDisplay
			}
		}
	}
	return ""
}

func buildRemoteStatus(repo repository) string {
	var remoteCol string
	if repo.behindCount > 0 && repo.aheadCount > 0 {
		remoteCol = RemoteStatusRed.Render(IconDiverged + " " + StatusDiverged)
		return remoteCol
	}
	if repo.behindCount > 0 {
		remoteCol = RemoteStatusBlue.Render(IconBehind + " " + StatusBehind)
		return remoteCol
	}
	remoteCol = RemoteStatusSynced
	return remoteCol
}

func buildLocalStatus(repo repository) string {
	if repo.hasError {
		return LocalStatusConflict
	} else if repo.hasModified {
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

func buildProjectName(repo string) string {
	return IconStyle.Render(IconGit) + " " + ProjectNameStyle.Render(repo)
}
