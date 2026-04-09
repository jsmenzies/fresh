package pullrequests

type ColumnLayout struct {
	TitleWidth    int
	CheckBarWidth int
	SummaryWidth  int
}

const (
	SelectorWidth       = 2
	NumberWidth         = 7
	DefaultTitleWidth   = 44
	MinTitleWidth       = 18
	MaxTitleWidth       = 70
	DefaultBarWidth     = 26
	MinBarWidth         = 14
	MaxBarWidth         = 32
	DefaultSummaryWidth = 22
	MinSummaryWidth     = 16
	MaxSummaryWidth     = 30
	InterColumnGap      = 2
)

func calculateColumnLayout(width int) ColumnLayout {
	titleWidth := DefaultTitleWidth
	barWidth := DefaultBarWidth
	summaryWidth := DefaultSummaryWidth

	if width <= 0 {
		return ColumnLayout{
			TitleWidth:    titleWidth,
			CheckBarWidth: barWidth,
			SummaryWidth:  summaryWidth,
		}
	}

	fixed := SelectorWidth + NumberWidth + (4 * InterColumnGap)
	available := width - fixed
	if available <= MinTitleWidth+MinBarWidth+MinSummaryWidth {
		return ColumnLayout{
			TitleWidth:    MinTitleWidth,
			CheckBarWidth: MinBarWidth,
			SummaryWidth:  MinSummaryWidth,
		}
	}

	titleWidth = clamp(available/2, MinTitleWidth, MaxTitleWidth)
	remaining := available - titleWidth
	barWidth = clamp((remaining*3)/5, MinBarWidth, MaxBarWidth)
	summaryWidth = clamp(remaining-barWidth, MinSummaryWidth, MaxSummaryWidth)

	for titleWidth+barWidth+summaryWidth > available {
		if titleWidth > MinTitleWidth {
			titleWidth--
			continue
		}
		if barWidth > MinBarWidth {
			barWidth--
			continue
		}
		if summaryWidth > MinSummaryWidth {
			summaryWidth--
			continue
		}
		break
	}

	return ColumnLayout{
		TitleWidth:    titleWidth,
		CheckBarWidth: barWidth,
		SummaryWidth:  summaryWidth,
	}
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
