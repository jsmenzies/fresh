package common

import (
	"fmt"
	"strings"

	"fresh/internal/domain"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func GenerateTable(repositories []domain.Repository, cursor int) string {
	headers := []string{"", "PROJECT", "BRANCH", "LOCAL", "REMOTE", "", "LAST COMMIT", "LINKS"}

	rows := make([][]string, len(repositories))
	for i, repo := range repositories {
		isSelected := i == cursor

		rows[i] = repositoryToRow(repo, isSelected)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(rows...).
		BorderStyle(TableBorderStyle).
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
	branchName := buildBranchName(repo.Branch)
	localCol := buildLocalStatus(repo.LocalState)
	remoteCol := buildRemoteStatus(repo)
	linksCol := buildLinks(repo.RemoteURL, repo.Branch)
	lastUpdateCol := buildLastUpdate(repo)
	info := buildInfo(repo)

	return []string{
		selector,
		projectName,
		branchName,
		localCol,
		remoteCol,
		info,
		lastUpdateCol,
		linksCol,
	}
}

func buildSelector(isSelected bool) string {
	if isSelected {
		return SelectorStyle.Render(IconSelector)
	}
	return SelectorStyle.Render(" ")
}

func buildProjectName(name string, isSelected bool) string {
	if isSelected {
		return ProjectNameStyle.Bold(true).Render(name)
	}
	return ProjectNameStyle.Render(name)
}

func buildBranchName(branch domain.Branch) string {
	switch s := branch.(type) {
	case domain.NoBranch:
		return BranchNameEmpty
	case domain.DetachedHead:
		return BranchNameHead
	case domain.OnBranch:
		return BranchNameStyle.Render(s.Name)
	default:
		return BranchNameEmpty
	}
}

func buildLocalStatus(state domain.LocalState) string {
	switch state.(type) {
	case domain.DirtyLocalState:
		return LocalStatusDirty
	case domain.UntrackedLocalState:
		return LocalStatusUntracked
	case domain.LocalStateError:
		return LocalStatusError
	default:
		return LocalStatusClean
	}
}

func buildRemoteStatus(repo domain.Repository) string {
	switch activity := repo.Activity.(type) {
	case domain.RefreshingActivity:
		if !activity.Complete {
			return RemoteStatusUpdating.Render(activity.Spinner.View())
		}
	}

	switch s := repo.RemoteState.(type) {
	case domain.NoUpstream:
		return RemoteStatusError
	case domain.DetachedRemote:
		return RemoteStatusError
	case domain.RemoteError:
		return RemoteStatusError
	case domain.Diverged:
		return RemoteStatusCounts(s.BehindCount, s.AheadCount)
	case domain.Behind:
		return RemoteStatusCounts(s.Count, 0)
	case domain.Ahead:
		return RemoteStatusCounts(0, s.Count)
	default:
		return RemoteStatusSynced
	}
}

func buildInfo(repo domain.Repository) string {
	var content string
	switch activity := repo.Activity.(type) {
	case domain.PullingActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := TruncateWithEllipsis(lastLine, InfoWidth-3)
			content = FormatPullProgress(activity.Spinner.View(), truncated)
		} else {
			lastLine := activity.GetLastLine()
			content = stylePullOutput(lastLine, activity.ExitCode)
		}
	case domain.RefreshingActivity:
		if !activity.Complete {
			content = ""
		}
	}

	if content == "" {
		switch s := repo.RemoteState.(type) {
		case domain.NoUpstream:
			content = RenderStatusMessage(MsgNoUpstream, InfoWidth)
		case domain.DetachedRemote:
			content = RenderStatusMessage(MsgDetached, InfoWidth)
		case domain.RemoteError:
			content = RemoteStatusErrorText.Render(TruncateWithEllipsis(s.Message, InfoWidth))
		case domain.Diverged:
			content = RenderStatusMessage(MsgDiverged, InfoWidth)
		}
	}

	return InfoStyle.Render(content)
}

func stylePullOutput(lastLine string, exitCode int) string {
	lowerLine := strings.ToLower(lastLine)
	truncated := TruncateWithEllipsis(lastLine, InfoWidth)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
		return PullOutputError.Render(truncated)
	}

	if exitCode == 0 {
		if strings.Contains(lowerLine, "up to date") || strings.Contains(lowerLine, "up-to-date") {
			return PullOutputUpToDate.Render(truncated)
		}
		if strings.Contains(lowerLine, "done") ||
			(strings.Contains(lowerLine, "file") && strings.Contains(lowerLine, "changed")) {
			return PullOutputSuccess.Render(truncated)
		}
	}

	return PullOutputWarn.Render(truncated)
}

func buildLastUpdate(repo domain.Repository) string {
	timeAgo := FormatTimeAgo(repo.LastCommitTime)
	return TimeAgoStyle.Render(IconClock + " " + timeAgo)
}

func buildLinks(url string, branch domain.Branch) string {
	var branchName string
	if url != "" {
		if IsGitHubRepository(url) {
			if _, ok := branch.(domain.OnBranch); !ok {
				branchName = ""
			} else {
				branchName = branch.(domain.OnBranch).Name
			}
			githubURLs := BuildGitHubURLs(url, branchName)
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

func MakeClickableURL(url string, displayText string) string {
	if url == "" {
		return displayText
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, displayText)
}
