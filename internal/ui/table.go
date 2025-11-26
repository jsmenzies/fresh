package ui

import (
	"fmt"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/formatting"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// GenerateTable creates a table view of repositories
func GenerateTable(repositories []domain.Repository, cursor int) string {
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
		})

	return t.Render()
}

func repositoryToRow(repo domain.Repository) []string {
	projectName := buildProjectName(repo.Name)
	branchName := buildBranchName(repo.CurrentBranch)
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

func buildLastUpdate(repo domain.Repository) string {
	timeAgo := formatting.FormatTimeAgo(repo.LastCommitTime)

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
	// MANUAL badge: repo has conflicts, is dirty, or is diverged
	//if repo.HasError || repo.HasModified || (repo.BehindCount > 0 && repo.AheadCount > 0) {
	//	return TagStyle.Render(BadgeManual)
	//}
	//
	//// READY badge: repo is clean and behind (can be auto-updated)
	//if repo.BehindCount > 0 && !repo.HasModified && !repo.HasError {
	//	return BadgeReadyStyle.Render(BadgeReady)
	//}

	// No badge for synced repos or repos ahead only
	return BadgeStyle.Render("")
}

func buildLinks(repo domain.Repository) string {
	if repo.RemoteURL != "" {
		if formatting.IsGitHubRepository(repo.RemoteURL) {
			githubURLs := formatting.BuildGitHubURLs(repo.RemoteURL, repo.CurrentBranch)
			if githubURLs != nil {
				var shortcuts []string

				// Code link (to current branch)
				codeLink := MakeClickableURL(githubURLs["code"], IconCode)
				shortcuts = append(shortcuts, LinkStyle.Render(codeLink))

				// PRs link
				prsLink := MakeClickableURL(githubURLs["prs"], IconPullRequests)
				shortcuts = append(shortcuts, LinkStyle.Render(prsLink))

				// Open PR link
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
	var remoteCol string
	if repo.BehindCount > 0 && repo.AheadCount > 0 {
		remoteCol = RemoteStatusRed.Render(IconDiverged + " " + StatusDiverged)
		return remoteCol
	}
	if repo.BehindCount > 0 {
		remoteCol = RemoteStatusBlue.Render(IconBehind + " " + StatusBehind)
		return remoteCol
	}
	remoteCol = RemoteStatusSynced
	return remoteCol
}

func buildLocalStatus(repo domain.Repository) string {
	if repo.HasError {
		return LocalStatusConflict
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

func buildProjectName(repo string) string {
	return IconStyle.Render(IconGit) + " " + ProjectNameStyle.Render(repo)
}

// MakeClickableURL creates a terminal hyperlink
func MakeClickableURL(url string, displayText string) string {
	if url == "" {
		return displayText
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, displayText)
}
