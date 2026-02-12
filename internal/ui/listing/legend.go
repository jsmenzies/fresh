package listing

import (
	"fresh/internal/ui/common"

	"github.com/charmbracelet/lipgloss"
)

func renderLegend(show bool) string {
	if !show {
		return ""
	}
	type item struct {
		icon, label string
		style       lipgloss.Style
	}

	c1 := []item{
		{common.IconClean, "Clean", common.TextGreen},
		{common.IconDirty, "Dirty", common.LocalStatusDirtyItem},
		{common.IconDiverged, "Untracked", common.LocalStatusUntrackedItem},
	}

	c2 := []item{
		{common.IconUntracked, "Untracked Files", common.LocalStatusUntrackedItem},
		{"~", "Modified Files", common.LocalStatusDirtyItem},
		{"-", "Deleted Files", common.LocalStatusDirtyItem},
		{"+", "Added Files", common.LocalStatusDirtyItem},
	}

	c3 := []item{
		{common.IconAhead, "Ahead", common.TextBlue},
		{common.IconBehind, "Behind", common.TextBlue},
		{common.IconSynced, "Synced", common.TextSubtleGreen},
		{common.IconRemoteError, "Error Fetching", common.RemoteStatusErrorText},
	}

	process := func(items []item) []string {
		var rows []string
		for _, it := range items {
			rows = append(rows, it.style.Render(it.icon)+" "+it.label)
		}
		return rows
	}

	col1 := lipgloss.JoinVertical(lipgloss.Left, process(c1)...)
	col2 := lipgloss.JoinVertical(lipgloss.Left, process(c2)...)
	col3 := lipgloss.JoinVertical(lipgloss.Left, process(c3)...)

	colStyle := lipgloss.NewStyle().Width(LegendColWidth)

	grid := lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(col1), colStyle.Render(col2), colStyle.Render(col3))

	return common.FooterStyle.Render(grid)
}
