package listing

import (
	"fmt"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const (
	MaxProjectWidth = 50
	MaxBranchWidth  = 30
	MinProjectWidth = 22
	MinBranchWidth  = 8
)

// Column width constants - centralized to avoid magic numbers
const (
	SelectorWidth   = 2
	LocalWidth      = 15
	RemoteWidth     = 11
	InfoWidth       = 42
	LastCommitWidth = 20
	LinksWidth      = 8
	InterColumnGap  = 2 // spacing between columns
)

// Total fixed width (non-project/branch columns)
func totalFixedWidth() int {
	return SelectorWidth + LocalWidth + RemoteWidth + InfoWidth +
		LastCommitWidth + LinksWidth + (6 * InterColumnGap)
}

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

func calculateColumnWidths(repositories []domain.Repository, terminalWidth int) (projectWidth, branchWidth int) {
	// Empty repos edge case: return minimum widths
	if len(repositories) == 0 {
		return MinProjectWidth, MinBranchWidth
	}

	// Find max content length (including detached HEAD label)
	maxProjectLen := 0
	maxBranchLen := 0

	for _, repo := range repositories {
		if len(repo.Name) > maxProjectLen {
			maxProjectLen = len(repo.Name)
		}

		switch branch := repo.Branches.Current.(type) {
		case domain.OnBranch:
			if len(branch.Name) > maxBranchLen {
				maxBranchLen = len(branch.Name)
			}
		case domain.DetachedHead:
			if len(common.BranchHead) > maxBranchLen {
				maxBranchLen = len(common.BranchHead)
			}
		}
	}

	// Clamp to min/max immediately
	projectWidth = clamp(maxProjectLen, MinProjectWidth, MaxProjectWidth)
	branchWidth = clamp(maxBranchLen, MinBranchWidth, MaxBranchWidth)

	// If no terminal width provided, use content-based widths
	if terminalWidth <= 0 {
		return projectWidth, branchWidth
	}

	// Check if we need to shrink to fit terminal
	availableWidth := terminalWidth - totalFixedWidth()
	if availableWidth <= 0 {
		// Terminal too narrow: use minimums and let truncation handle it
		return MinProjectWidth, MinBranchWidth
	}

	totalDesired := projectWidth + branchWidth
	if totalDesired <= availableWidth {
		// We fit! Use the content-based widths
		return projectWidth, branchWidth
	}

	// Need to shrink proportionally
	return distributeWidth(projectWidth, branchWidth, availableWidth)
}

// clamp restricts value to [min, max] range
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// distributeWidth proportionally allocates availableWidth between project and branch
func distributeWidth(projectDesired, branchDesired, available int) (project, branch int) {
	if available <= 0 {
		return MinProjectWidth, MinBranchWidth
	}

	totalDesired := projectDesired + branchDesired
	if totalDesired == 0 {
		return MinProjectWidth, MinBranchWidth
	}

	// Calculate proportional distribution
	projectRatio := float64(projectDesired) / float64(totalDesired)
	project = int(float64(available) * projectRatio)
	branch = available - project

	// Enforce minimums (this may cause total to exceed available)
	if project < MinProjectWidth {
		project = MinProjectWidth
		branch = available - project
	}
	if branch < MinBranchWidth {
		branch = MinBranchWidth
		project = available - branch
	}

	// Final clamp to maxes
	project = clamp(project, MinProjectWidth, MaxProjectWidth)
	branch = clamp(branch, MinBranchWidth, MaxBranchWidth)

	return project, branch
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
	if isSelected {
		return common.SelectorStyle.Render(common.IconSelector)
	}
	return common.SelectorStyle.Render(" ")
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
	case *domain.RefreshingActivity:
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
	case *domain.PullingActivity:
		if !activity.Complete {
			lastLine := activity.GetLastLine()
			truncated := common.TruncateWithEllipsis(lastLine, common.InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated)
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
			truncated := common.TruncateWithEllipsis(lastLine, common.InfoWidth-3)
			content = common.FormatPullProgress(activity.Spinner.View(), truncated)
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
					content = common.PullOutputError.Render(common.TruncateWithEllipsis(firstError, common.InfoWidth))
				} else {
					content = common.PullOutputWarn.Render("No branches to prune")
				}
			} else {
				content = common.PullOutputSuccess.Render(fmt.Sprintf("Deleted %d branches", activity.DeletedCount))
			}
		}
	}

	if content == "" {
		// Check for remote errors first
		switch s := repo.RemoteState.(type) {
		case domain.RemoteError:
			content = common.RemoteStatusErrorText.Render(common.TruncateWithEllipsis(s.Message, common.InfoWidth))
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
				content = common.RenderStatusMessage(common.MsgNoUpstream, common.InfoWidth)
			} else if _, ok := repo.RemoteState.(domain.DetachedRemote); ok {
				content = common.RenderStatusMessage(common.MsgDetached, common.InfoWidth)
			} else if _, ok := repo.RemoteState.(domain.Diverged); ok {
				content = common.RenderStatusMessage(common.MsgDiverged, common.InfoWidth)
			}
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
