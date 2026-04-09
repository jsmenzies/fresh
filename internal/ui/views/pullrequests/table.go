package pullrequests

import (
	"fmt"
	"math"
	"strings"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

func RenderTable(rows []domain.PullRequestDetails, cursor int, layout ColumnLayout, pulse bool) string {
	headers := []string{"", "#", "Title", "Checks", "Summary"}

	renderedRows := make([][]string, len(rows))
	for i, row := range rows {
		renderedRows[i] = pullRequestToRow(row, i == cursor, layout, pulse)
	}

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		Headers(headers...).
		Rows(renderedRows...).
		BorderStyle(common.TableBorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return common.TableHeaderStyle
			}
			return lipgloss.NewStyle()
		})

	return t.Render()
}

func pullRequestToRow(row domain.PullRequestDetails, selected bool, layout ColumnLayout, pulse bool) []string {
	return []string{
		buildSelector(selected),
		buildNumber(row.Number, row.IsMine),
		buildTitle(row.Title, row.IsMine, selected, layout.TitleWidth),
		buildCheckBar(row.Checks, layout.CheckBarWidth, pulse),
		buildCheckSummary(row.Checks, layout.SummaryWidth),
	}
}

func buildSelector(selected bool) string {
	style := common.SelectorStyle.Width(SelectorWidth)
	if selected {
		return style.Render(common.IconSelector)
	}
	return style.Render(" ")
}

func buildNumber(number int, mine bool) string {
	style := lipgloss.NewStyle().
		Width(NumberWidth).
		MaxWidth(NumberWidth).
		AlignHorizontal(lipgloss.Left)
	if mine {
		style = style.Foreground(common.Blue).Bold(true)
	} else {
		style = style.Foreground(common.TextSecondary)
	}
	return style.Render(fmt.Sprintf("#%d", number))
}

func buildTitle(title string, mine bool, selected bool, width int) string {
	style := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Left)
	if mine {
		style = style.Foreground(common.Blue)
	} else {
		style = style.Foreground(common.TextPrimary)
	}
	if selected {
		style = style.Bold(true)
	}

	return common.RenderTruncatedText(title, width, style)
}

func buildCheckBar(checks domain.PullRequestChecks, width int, pulse bool) string {
	if width < 1 {
		width = 1
	}

	baseStyle := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		AlignHorizontal(lipgloss.Left)

	if checks.Total <= 0 {
		return baseStyle.Foreground(common.SubtleGray).Render(strings.Repeat("·", width))
	}

	counts := []int{
		checks.Passed,
		checks.Waiting,
		checks.Running,
		checks.Failed,
		checks.Skipped,
	}
	scaled := scaleCountsToWidth(counts, width)

	runningChar := "▒"
	if pulse {
		runningChar = "▓"
	}

	segments := []struct {
		Count int
		Style lipgloss.Style
		Char  string
	}{
		{Count: scaled[0], Style: lipgloss.NewStyle().Foreground(common.Green), Char: "█"},
		{Count: scaled[1], Style: lipgloss.NewStyle().Foreground(common.Yellow), Char: "░"},
		{Count: scaled[2], Style: lipgloss.NewStyle().Foreground(common.Yellow), Char: runningChar},
		{Count: scaled[3], Style: lipgloss.NewStyle().Foreground(common.Red), Char: "█"},
		{Count: scaled[4], Style: lipgloss.NewStyle().Foreground(common.SubtleGray), Char: "█"},
	}

	var bar strings.Builder
	for _, segment := range segments {
		if segment.Count <= 0 {
			continue
		}
		bar.WriteString(segment.Style.Render(strings.Repeat(segment.Char, segment.Count)))
	}

	barText := bar.String()
	if lipgloss.Width(barText) < width {
		barText += lipgloss.NewStyle().Foreground(common.SubtleGray).Render(strings.Repeat("·", width-lipgloss.Width(barText)))
	}

	return baseStyle.Render(barText)
}

func buildCheckSummary(checks domain.PullRequestChecks, width int) string {
	style := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		AlignHorizontal(lipgloss.Left).
		Foreground(common.TextSecondary)

	if checks.Total <= 0 {
		return style.Foreground(common.SubtleGray).Render("No checks")
	}

	complete := checks.Passed + checks.Skipped
	summary := fmt.Sprintf("%d/%d complete", complete, checks.Total)
	switch {
	case checks.Failed > 0:
		summary += fmt.Sprintf(" • %d failing", checks.Failed)
	case checks.Running > 0:
		summary += fmt.Sprintf(" • %d running", checks.Running)
	case checks.Waiting > 0:
		summary += fmt.Sprintf(" • %d waiting", checks.Waiting)
	}

	if lipgloss.Width(summary) > width {
		summary = common.TruncateWithEllipsis(summary, width)
	}

	return style.Render(summary)
}

func scaleCountsToWidth(counts []int, width int) []int {
	scaled := make([]int, len(counts))
	if width <= 0 {
		return scaled
	}

	total := 0
	for _, count := range counts {
		if count > 0 {
			total += count
		}
	}
	if total <= 0 {
		return scaled
	}

	remainders := make([]float64, len(counts))
	used := 0
	for i, count := range counts {
		if count <= 0 {
			continue
		}
		raw := (float64(count) * float64(width)) / float64(total)
		scaled[i] = int(math.Floor(raw))
		remainders[i] = raw - float64(scaled[i])
		used += scaled[i]
	}

	for used < width {
		index := indexOfLargestRemainder(remainders, counts)
		if index < 0 {
			break
		}
		scaled[index]++
		remainders[index] = 0
		used++
	}

	for used > width {
		index := indexOfLargestSegment(scaled)
		if index < 0 {
			break
		}
		scaled[index]--
		used--
	}

	return scaled
}

func indexOfLargestRemainder(remainders []float64, counts []int) int {
	bestIndex := -1
	best := -1.0
	for i := range remainders {
		if counts[i] <= 0 {
			continue
		}
		if remainders[i] > best {
			best = remainders[i]
			bestIndex = i
		}
	}
	return bestIndex
}

func indexOfLargestSegment(counts []int) int {
	bestIndex := -1
	best := 0
	for i, count := range counts {
		if count > best {
			best = count
			bestIndex = i
		}
	}
	return bestIndex
}
