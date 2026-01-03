package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"github.com/charmbracelet/lipgloss"
)

type LegendType int

const (
	LegendNone LegendType = iota
	LegendPartial
	LegendFull
)

type item struct {
	icon, label string
	style       lipgloss.Style
	active      bool
}

func RenderLegend(repo domain.Repository, mode LegendType) string {
	if mode == LegendNone {
		return ""
	}

	isDirty := false
	hasUntracked := false
	hasModified := false
	hasAdded := false
	hasDeleted := false
	isClean := true

	if s, ok := repo.LocalState.(domain.DirtyLocalState); ok {
		hasUntracked = s.Untracked > 0
		hasModified = s.Modified > 0
		hasAdded = s.Added > 0
		hasDeleted = s.Deleted > 0
		isDirty = hasModified || hasAdded || hasDeleted || hasUntracked
		isClean = false
	} else if _, ok := repo.LocalState.(domain.LocalStateError); ok {
		isClean = false
	}

	isAhead := false
	isBehind := false
	isSynced := false
	isError := false

	switch repo.RemoteState.(type) {
	case domain.Ahead:
		isAhead = true
	case domain.Behind:
		isBehind = true
	case domain.Diverged:
		isAhead = true
		isBehind = true
	case domain.NoUpstream, domain.DetachedRemote, domain.RemoteError:
		isError = true
	default:
		isSynced = true
	}

	c1 := []item{
		{common.IconClean, "Clean", common.TextGreen, isClean},
		{common.IconDirty, "Dirty", common.LocalStatusDirtyItem, isDirty},
		{common.IconDiverged, "Untracked", common.LocalStatusUntrackedItem, hasUntracked},
	}

	c2 := []item{
		{common.IconUntracked, "Untracked Files", common.LocalStatusUntrackedItem, hasUntracked},
		{"~", "Modified Files", common.LocalStatusDirtyItem, hasModified},
		{"-", "Deleted Files", common.LocalStatusDirtyItem, hasDeleted},
		{"+", "Added Files", common.LocalStatusDirtyItem, hasAdded},
	}

	c3 := []item{
		{common.IconAhead, "Ahead", common.TextBlue, isAhead},
		{common.IconBehind, "Behind", common.TextBlue, isBehind},
		{common.IconSynced, "Synced", common.TextSubtleGreen, isSynced},
		{common.IconRemoteError, "Error Fetching", common.RemoteStatusErrorText, isError},
	}

	process := func(items []item) []string {
		var rows []string
		for _, it := range items {
			if mode == LegendFull || it.active {
				rows = append(rows, it.style.Render(it.icon)+" "+it.label)
			}
		}
		return rows
	}

	col1 := lipgloss.JoinVertical(lipgloss.Left, process(c1)...)
	col2 := lipgloss.JoinVertical(lipgloss.Left, process(c2)...)
	col3 := lipgloss.JoinVertical(lipgloss.Left, process(c3)...)

	colStyle := lipgloss.NewStyle().Width(20)

	grid := lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(col1), colStyle.Render(col2), colStyle.Render(col3))

	return common.FooterStyle.Render(grid)
}
