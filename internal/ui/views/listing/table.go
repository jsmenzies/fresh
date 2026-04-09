package listing

import (
	"fmt"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

func GenerateTable(repositories []domain.Repository, cursor int, layout ColumnLayout, runtime InfoRuntime) string {
	headers := []string{"", "󰉋 Repo", " Branch", " Local", "󰓦 Remote", common.IconPullRequests + " PR", "", ""}

	rows := make([][]string, len(repositories))
	for i, repo := range repositories {
		isSelected := i == cursor
		rows[i] = repositoryToRow(repo, isSelected, layout, runtime)
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

func repositoryToRow(repo domain.Repository, isSelected bool, layout ColumnLayout, runtime InfoRuntime) []string {
	selector := buildSelector(isSelected)
	projectName := buildProjectName(repo.Name, isSelected, layout.ProjectWidth)
	branchName := buildBranchName(repo.Branches.Current, layout.BranchWidth)
	localCol := buildLocalStatus(repo.LocalState)
	remoteCol := buildRemoteStatus(repo)
	prCol := buildPullRequestStatus(repo.PullRequests, runtime)
	prAlertCol := buildPullRequestAlert(repo.PullRequests, runtime)
	info := buildInfo(repo, layout.InfoWidth, runtime)

	return []string{
		selector,
		projectName,
		branchName,
		localCol,
		remoteCol,
		prCol,
		prAlertCol,
		info,
	}
}

func buildPullRequestAlert(state domain.PullRequestState, runtime InfoRuntime) string {
	baseStyle := lipgloss.NewStyle().
		Width(PRAlertWidth).
		MaxWidth(PRAlertWidth).
		Height(1).
		MaxHeight(1).
		Align(lipgloss.Left)
	blockedStyle := common.AlertSpinnerStyle.
		Width(PRAlertWidth).
		MaxWidth(PRAlertWidth).
		Height(1).
		MaxHeight(1).
		Align(lipgloss.Left)
	readyStyle := common.SuccessSpinnerStyle.
		Width(PRAlertWidth).
		MaxWidth(PRAlertWidth).
		Height(1).
		MaxHeight(1).
		Align(lipgloss.Left)

	s, ok := state.(domain.PullRequestCount)
	if !ok {
		return baseStyle.Render("")
	}
	if runtime.PullRequestSyncing {
		return baseStyle.Render("")
	}

	if s.MyBlocked > 0 {
		frame := runtime.BlockedSpinner
		if frame == "" {
			return blockedStyle.Render(common.IconWarning)
		}
		return blockedStyle.Render(frame)
	}

	if s.MyReady > 0 {
		frame := runtime.ReadySpinner
		if frame == "" {
			return readyStyle.Render(common.IconClean)
		}
		return readyStyle.Render(frame)
	}

	return baseStyle.Render("")
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
		Height(1).
		MaxHeight(1).
		AlignHorizontal(lipgloss.Left)

	if isSelected {
		style = style.Bold(true)
	}

	return common.RenderTruncatedText(name, width, style)
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
		return baseStyle.Foreground(common.SubtleGray).Render("-")
	case domain.DetachedRemote:
		return baseStyle.Foreground(common.SubtleGray).Render("-")
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

func buildPullRequestStatus(state domain.PullRequestState, runtime InfoRuntime) string {
	baseStyle := common.PullRequestStatusBaseStyle.
		Width(PRWidth).
		MaxWidth(PRWidth)

	switch s := state.(type) {
	case domain.PullRequestCount:
		if s.Open <= 0 {
			return baseStyle.Render("")
		}
		if s.MyOpen > 0 {
			count := lipgloss.NewStyle().Foreground(common.Blue).Render(fmt.Sprintf("%d", s.Open))
			mine := lipgloss.NewStyle().Foreground(common.TextPrimary).Render("(*)")
			return baseStyle.Render(count + mine)
		}
		return baseStyle.Foreground(common.Blue).Render(fmt.Sprintf("%d", s.Open))
	case domain.PullRequestError:
		return baseStyle.Foreground(common.SubtleRed).Render(common.IconRemoteError)
	default:
		return baseStyle.Foreground(common.SubtleGray).Render("-")
	}
}

func buildInfo(repo domain.Repository, infoWidth int, runtime InfoRuntime) string {
	infoWidth = normalizeInfoWidth(infoWidth)

	infoStyle := common.InfoStyle.
		Width(infoWidth).
		MaxWidth(infoWidth)

	if repo.Activity.IsInProgress() {
		if active, ok := collectActiveActivityInfoMessage(repo, infoWidth); ok {
			return infoStyle.Render(renderInfoMessage(active, infoWidth))
		}
		return infoStyle.Render("")
	}

	messages := make([]InfoMessage, 0, 10)
	messages = append(messages, collectRecentActivityInfoMessages(runtime, repo.Path)...)
	messages = append(messages, collectStatusInfoMessages(repo)...)

	if len(messages) == 0 {
		return infoStyle.Render("")
	}

	pinned := filterPinnedInfoMessages(messages)
	if len(pinned) > 0 {
		idx := 0
		if len(pinned) > 1 {
			idx = int(runtime.Phase % uint64(len(pinned)))
		}

		return infoStyle.Render(renderInfoMessage(pinned[idx], infoWidth))
	}

	idx := 0
	if len(messages) > 1 {
		idx = int(runtime.Phase % uint64(len(messages)))
	}

	return infoStyle.Render(renderInfoMessage(messages[idx], infoWidth))
}

func buildMyPullRequestSummary(state domain.PullRequestState) (string, bool) {
	s, ok := state.(domain.PullRequestCount)
	if !ok {
		return "", false
	}

	parts := make([]string, 0, 4)
	hasPinned := false
	if s.MyReady > 0 {
		parts = append(parts, fmt.Sprintf("%d ready", s.MyReady))
	}
	if s.MyBlocked > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked", s.MyBlocked))
		hasPinned = true
	}
	if s.MyChecks > 0 {
		parts = append(parts, fmt.Sprintf("%d checks", s.MyChecks))
	}
	if s.MyReview > 0 {
		parts = append(parts, fmt.Sprintf("%d review", s.MyReview))
	}

	if len(parts) == 0 {
		return "", false
	}

	return "My PRs: " + strings.Join(parts, ", "), hasPinned
}

func stylePullOutput(lastLine string, exitCode int, infoWidth int) string {
	return renderInfoMessage(buildPullOutputInfoMessage(lastLine, exitCode), normalizeInfoWidth(infoWidth))
}
