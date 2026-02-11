package listing

import (
	"fmt"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

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
		BorderStyle(common.TableBorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return common.TableHeaderStyle
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
		return common.SelectorStyle.Render(common.IconSelector)
	}
	return common.SelectorStyle.Render(" ")
}

func buildProjectName(name string, isSelected bool) string {
	if isSelected {
		return common.ProjectNameStyle.Bold(true).Render(name)
	}
	return common.ProjectNameStyle.Render(name)
}

func buildBranchName(branch domain.Branch) string {
	switch s := branch.(type) {
	case domain.NoBranch:
		return common.BranchNameEmpty
	case domain.DetachedHead:
		return common.BranchNameHead
	case domain.OnBranch:
		return common.BranchNameStyle.Render(s.Name)
	default:
		return common.BranchNameEmpty
	}
}

func buildLocalStatus(state domain.LocalState) string {
	switch s := state.(type) {
	case domain.DirtyLocalState:
		var parts []string

		if s.Untracked > 0 {
			parts = append(parts, common.LocalStatusUntrackedItem.Render(common.IconDiverged))
		} else {
			parts = append(parts, common.LocalStatusDirtyItem.Render(common.IconWarning))
		}

		if s.Untracked > 0 {
			parts = append(parts, common.LocalStatusUntrackedItem.Render(fmt.Sprintf("%s%d", common.IconUntracked, s.Untracked)))
		}
		if s.Added > 0 {
			parts = append(parts, common.LocalStatusDirtyItem.Render(fmt.Sprintf("+%d", s.Added)))
		}
		if s.Modified > 0 {
			parts = append(parts, common.LocalStatusDirtyItem.Render(fmt.Sprintf("~%d", s.Modified)))
		}
		if s.Deleted > 0 {
			parts = append(parts, common.LocalStatusDirtyItem.Render(fmt.Sprintf("-%d", s.Deleted)))
		}

		text := strings.Join(parts, " ")
		return common.LocalStatusBaseStyle.Render(text)
	case domain.LocalStateError:
		return common.LocalStatusError
	default:
		return common.LocalStatusClean
	}
}

func buildRemoteStatus(repo domain.Repository) string {
	switch activity := repo.Activity.(type) {
	case domain.RefreshingActivity:
		if !activity.Complete {
			return common.RemoteStatusUpdating.Render(activity.Spinner.View())
		}
	}

	switch s := repo.RemoteState.(type) {
	case domain.NoUpstream:
		return common.RemoteStatusError
	case domain.DetachedRemote:
		return common.RemoteStatusError
	case domain.RemoteError:
		return common.RemoteStatusError
	case domain.Diverged:
		return common.RemoteStatusCounts(s.BehindCount, s.AheadCount)
	case domain.Behind:
		return common.RemoteStatusCounts(s.Count, 0)
	case domain.Ahead:
		return common.RemoteStatusCounts(0, s.Count)
	default:
		return common.RemoteStatusSynced
	}
}

func buildInfo(repo domain.Repository) string {
	var content string
	switch activity := repo.Activity.(type) {
	case domain.PullingActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := common.TruncateWithEllipsis(lastLine, common.InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated)
		} else {
			lastLine := activity.GetLastLine()
			content = stylePullOutput(lastLine, activity.ExitCode)
		}
	case domain.RefreshingActivity:
		if !activity.Complete {
			content = ""
		}
	case domain.PruningActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := common.TruncateWithEllipsis(lastLine, common.InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated)
		} else {
			if activity.DeletedCount == 0 {
				content = common.PullOutputWarn.Render("No branches to prune")
			} else {
				content = common.PullOutputSuccess.Render(fmt.Sprintf("Deleted %d branches", activity.DeletedCount))
			}
		}
	}

	if content == "" {
		switch s := repo.RemoteState.(type) {
		case domain.NoUpstream:
			content = common.RenderStatusMessage(common.MsgNoUpstream, common.InfoWidth)
		case domain.DetachedRemote:
			content = common.RenderStatusMessage(common.MsgDetached, common.InfoWidth)
		case domain.RemoteError:
			content = common.RemoteStatusErrorText.Render(common.TruncateWithEllipsis(s.Message, common.InfoWidth))
		case domain.Diverged:
			content = common.RenderStatusMessage(common.MsgDiverged, common.InfoWidth)
		}
	}

	return common.InfoStyle.Render(content)
}

func stylePullOutput(lastLine string, exitCode int) string {
	lowerLine := strings.ToLower(lastLine)
	truncated := common.TruncateWithEllipsis(lastLine, common.InfoWidth)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
		return common.PullOutputError.Render(truncated)
	}

	if exitCode == 0 {
		if strings.Contains(lowerLine, "up to date") || strings.Contains(lowerLine, "up-to-date") {
			return common.PullOutputUpToDate.Render(truncated)
		}
		if strings.Contains(lowerLine, "done") ||
			(strings.Contains(lowerLine, "file") && strings.Contains(lowerLine, "changed")) {
			return common.PullOutputSuccess.Render(truncated)
		}
	}

	return common.PullOutputWarn.Render(truncated)
}

func buildLastUpdate(repo domain.Repository) string {
	if repo.LastCommitTime.IsZero() {
		return ""
	}
	timeAgo := common.FormatTimeAgo(repo.LastCommitTime)
	return common.TimeAgoStyle.Render(common.IconClock + " " + timeAgo)
}

func buildLinks(url string, branch domain.Branch) string {
	var branchName string
	if url != "" {
		if common.IsGitHubRepository(url) {
			if _, ok := branch.(domain.OnBranch); !ok {
				branchName = ""
			} else {
				branchName = branch.(domain.OnBranch).Name
			}
			githubURLs := common.BuildGitHubURLs(url, branchName)
			if githubURLs != nil {
				var shortcuts []string

				codeLink := MakeClickableURL(githubURLs["code"], common.IconCode)
				shortcuts = append(shortcuts, common.LinkStyle.Render(codeLink))

				prsLink := MakeClickableURL(githubURLs["prs"], common.IconPullRequests)
				shortcuts = append(shortcuts, common.LinkStyle.Render(prsLink))

				openPRLink := MakeClickableURL(githubURLs["openpr"], common.IconOpenPR)
				shortcuts = append(shortcuts, common.LinkStyle.Render(openPRLink))

				shortcutsDisplay := fmt.Sprintf("%s", strings.Join(shortcuts, " "))
				return common.LinksStyle.Render(shortcutsDisplay)
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
