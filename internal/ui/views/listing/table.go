package listing

import (
	"fmt"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/git"
	"fresh/internal/ui/views/common"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func GenerateTable(repositories []domain.Repository, cursor int, terminalWidth int) string {
	projectWidth, branchWidth := calculateColumnWidths(repositories, terminalWidth)

	headers := []string{"", "PROJECT", "BRANCH", "LOCAL", "REMOTE", "", "LAST COMMIT", "LINKS"}

	rows := make([][]string, len(repositories))
	for i, repo := range repositories {
		isSelected := i == cursor
		rows[i] = repositoryToRow(repo, isSelected, projectWidth, branchWidth)
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

func repositoryToRow(repo domain.Repository, isSelected bool, projectWidth, branchWidth int) []string {
	selector := buildSelector(isSelected)
	projectName := buildProjectName(repo.Name, isSelected, projectWidth)
	branchName := buildBranchName(repo.Branches.Current, branchWidth)
	localCol := buildLocalStatus(repo.LocalState)
	remoteCol := buildRemoteStatus(repo)
	linksCol := buildLinks(repo.RemoteURL, repo.Branches.Current)
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
	style := common.SelectorStyle.Width(SelectorWidth)
	if isSelected {
		return style.Render(common.IconSelector)
	}
	return style.Render(" ")
}

func buildProjectName(name string, isSelected bool, width int) string {
	style := lipgloss.NewStyle().
		Foreground(common.TextPrimary).
		Align(lipgloss.Left).
		Width(width).
		MaxWidth(width).
		AlignHorizontal(lipgloss.Left)

	if isSelected {
		style = style.Bold(true)
	}

	return style.Render(name)
}

func buildBranchName(branch domain.Branch, width int) string {
	style := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Width(width).
		MaxWidth(width).
		Height(1).
		MaxHeight(1).
		AlignHorizontal(lipgloss.Left).
		Foreground(common.TextBranch)

	switch s := branch.(type) {
	case domain.NoBranch:
		emptyStyle := style.Foreground(common.SubtleGray)
		return emptyStyle.Render("")
	case domain.DetachedHead:
		headStyle := style.Foreground(common.SubtleGray)
		return headStyle.Render(common.BranchHead)
	case domain.OnBranch:
		return style.Render(s.Name)
	default:
		emptyStyle := style.Foreground(common.SubtleGray)
		return emptyStyle.Render("")
	}
}

func buildLocalStatus(state domain.LocalState) string {
	baseStyle := common.LocalStatusBaseStyle.
		Width(LocalWidth).
		MaxWidth(LocalWidth)

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
		return baseStyle.Render(text)
	case domain.LocalStateError:
		return baseStyle.Render("")
	default:
		return baseStyle.Foreground(common.Green).Render(common.IconClean)
	}
}

func buildRemoteStatus(repo domain.Repository) string {
	baseStyle := common.RemoteStatusBaseStyle.
		Width(RemoteWidth).
		MaxWidth(RemoteWidth)

	switch activity := repo.Activity.(type) {
	case *domain.RefreshingActivity:
		if !activity.Complete {
			return baseStyle.Align(lipgloss.Left).Render(activity.Spinner.View())
		}
	}

	switch s := repo.RemoteState.(type) {
	case domain.NoUpstream:
		return baseStyle.Foreground(common.SubtleRed).Render(common.IconRemoteError)
	case domain.DetachedRemote:
		return baseStyle.Foreground(common.SubtleRed).Render(common.IconRemoteError)
	case domain.RemoteError:
		return baseStyle.Foreground(common.SubtleRed).Render(common.IconRemoteError)
	case domain.Diverged:
		return common.RemoteStatusCounts(s.BehindCount, s.AheadCount, RemoteWidth)
	case domain.Behind:
		return common.RemoteStatusCounts(s.Count, 0, RemoteWidth)
	case domain.Ahead:
		return common.RemoteStatusCounts(0, s.Count, RemoteWidth)
	default:
		return baseStyle.Foreground(common.SubtleGreen).Render(common.IconSynced)
	}
}

func buildInfo(repo domain.Repository) string {
	infoStyle := common.InfoStyle.
		Width(InfoWidth).
		MaxWidth(InfoWidth)

	var content string
	switch activity := repo.Activity.(type) {
	case *domain.PullingActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := common.TruncateWithEllipsis(lastLine, InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated, InfoWidth-2)
		} else {
			lastLine := activity.GetLastLine()
			content = stylePullOutput(lastLine, activity.ExitCode)
		}
	case *domain.RefreshingActivity:
		if !activity.Complete {
			content = ""
		}
	case *domain.PruningActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := common.TruncateWithEllipsis(lastLine, InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated, InfoWidth-2)
		} else {
			if activity.DeletedCount == 0 {
				// Check for error messages in the output lines
				var firstError string
				for _, line := range activity.Lines {
					if strings.HasPrefix(line, "Failed: ") {
						firstError = strings.TrimPrefix(line, "Failed: ")
						break
					}
				}
				if firstError != "" {
					content = common.PullOutputError.Width(InfoWidth).Render(common.TruncateWithEllipsis(firstError, InfoWidth))
				} else {
					content = common.PullOutputWarn.Width(InfoWidth).Render("No branches to prune")
				}
			} else {
				content = common.PullOutputSuccess.Width(InfoWidth).Render(fmt.Sprintf("Deleted %d branches", activity.DeletedCount))
			}
		}
	}

	if content == "" {
		// Check for remote errors first
		switch s := repo.RemoteState.(type) {
		case domain.RemoteError:
			content = common.RemoteStatusErrorText.Render(common.TruncateWithEllipsis(s.Message, InfoWidth))
		default:
			// No error, check for prunable branches
			mergedCount := len(repo.Branches.Merged)
			if mergedCount > 0 {
				mergedText := "branches"
				if mergedCount == 1 {
					mergedText = "branch"
				}
				content = common.TextGrey.Render(fmt.Sprintf("%d prunable %s", mergedCount, mergedText))
			} else if _, ok := repo.RemoteState.(domain.NoUpstream); ok {
				content = common.RenderStatusMessage(common.MsgNoUpstream, InfoWidth)
			} else if _, ok := repo.RemoteState.(domain.DetachedRemote); ok {
				content = common.RenderStatusMessage(common.MsgDetached, InfoWidth)
			} else if _, ok := repo.RemoteState.(domain.Diverged); ok {
				content = common.RenderStatusMessage(common.MsgDiverged, InfoWidth)
			}
		}
	}

	return infoStyle.Render(content)
}

func stylePullOutput(lastLine string, exitCode int) string {
	lowerLine := strings.ToLower(lastLine)
	truncated := common.TruncateWithEllipsis(lastLine, InfoWidth)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
		return common.PullOutputError.Width(InfoWidth).Render(truncated)
	}

	if exitCode == 0 {
		if strings.Contains(lowerLine, "up to date") || strings.Contains(lowerLine, "up-to-date") {
			return common.PullOutputUpToDate.Width(InfoWidth).Render(truncated)
		}
		if strings.Contains(lowerLine, "done") ||
			(strings.Contains(lowerLine, "file") && strings.Contains(lowerLine, "changed")) {
			return common.PullOutputSuccess.Width(InfoWidth).Render(truncated)
		}
	}

	return common.PullOutputWarn.Width(InfoWidth).Render(truncated)
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
		if git.IsGitHubRepository(url) {
			if _, ok := branch.(domain.OnBranch); !ok {
				branchName = ""
			} else {
				branchName = branch.(domain.OnBranch).Name
			}
			githubURLs := git.BuildGitHubURLs(url, branchName)
			if githubURLs != nil {
				var shortcuts []string

				codeLink := MakeClickableURL(githubURLs["code"], common.IconCode)
				shortcuts = append(shortcuts, common.LinkStyle.Render(codeLink))

				prsLink := MakeClickableURL(githubURLs["prs"], common.IconPullRequests)
				shortcuts = append(shortcuts, common.LinkStyle.Render(prsLink))

				openPRLink := MakeClickableURL(githubURLs["openpr"], common.IconOpenPR)
				shortcuts = append(shortcuts, common.LinkStyle.Render(openPRLink))

				shortcutsDisplay := fmt.Sprintf("%s", strings.Join(shortcuts, " "))
				return common.LinksStyle.Width(LinksWidth).Render(shortcutsDisplay)
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
