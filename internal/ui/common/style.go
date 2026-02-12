package common

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	SubtleGray    = lipgloss.Color("#5b6078")
	TextPrimary   = lipgloss.Color("#FFFFFF")
	TextSecondary = lipgloss.Color("#a9b1d6")
	TextBranch    = lipgloss.Color("#C27AFF")

	SubtleRed   = lipgloss.Color("#FF7A7A")
	SubtleGreen = lipgloss.Color("#7AFFA1")

	Green  = lipgloss.Color("#06DF71")
	Yellow = lipgloss.Color("#FEC700")
	Red    = lipgloss.Color("#FF6367")
	Blue   = lipgloss.Color("#52A2FF")

	DividerColor = lipgloss.Color("#414868")
	Padding      = 2
)

var TableBorderStyle = lipgloss.NewStyle().Foreground(DividerColor)

func NewGreenDotSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	return s
}

func NewRefreshSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(TextSecondary)
	return s
}

func NewPullSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(Blue)
	return s
}

var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true).
	Align(lipgloss.Left)

var LocalStatusBaseStyle = lipgloss.NewStyle().
	Height(1).
	MaxHeight(1).
	AlignHorizontal(lipgloss.Left)

var LocalStatusUntrackedItem = lipgloss.NewStyle().Foreground(Red)
var LocalStatusDirtyItem = lipgloss.NewStyle().Foreground(Yellow)
var TextGreen = lipgloss.NewStyle().Foreground(Green)
var TextSubtleGreen = lipgloss.NewStyle().Foreground(SubtleGreen)
var TextBlue = lipgloss.NewStyle().Foreground(Blue)
var TextGrey = lipgloss.NewStyle().Foreground(SubtleGray)

var RemoteStatusBaseStyle = lipgloss.NewStyle().
	Height(1).
	MaxHeight(1).
	AlignHorizontal(lipgloss.Left)

var RemoteStatusCountsStyle = RemoteStatusBaseStyle.
	Foreground(Blue)

func RemoteStatusCounts(behind int, ahead int, width int) string {
	content := ""
	if behind > 0 && ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d / "+IconBehind+" %d", ahead, behind)
	} else if behind > 0 {
		content = fmt.Sprintf(IconBehind+" %d", behind)
	} else if ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d", ahead)
	}

	return RemoteStatusCountsStyle.
		Width(width).
		MaxWidth(width).
		Render(content)
}

var RemoteStatusErrorText = lipgloss.NewStyle().
	Foreground(SubtleRed)

var RemoteStatusErrorHelpText = lipgloss.NewStyle().
	Foreground(SubtleGray)

var LinkStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true)

var LinksStyle = lipgloss.NewStyle()

var BadgeStyle = lipgloss.NewStyle().
	Inline(true).
	MarginLeft(2)

var TimeAgoStyle = lipgloss.NewStyle().
	Foreground(TextSecondary)

var PullOutputSuccess = lipgloss.NewStyle().
	Foreground(Green).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputUpToDate = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputWarn = lipgloss.NewStyle().
	Foreground(Yellow).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputError = lipgloss.NewStyle().
	Foreground(Red).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullProgressStyle = lipgloss.NewStyle()

var InfoStyle = lipgloss.NewStyle().
	MaxHeight(1)

const (
	LabelNoUpstream = "No upstream "
	HelpNoUpstream  = "(new branch or deleted remote)"
	LabelDetached   = "Detached HEAD "
	HelpDetached    = "(not currently on a branch)"
	HelpDiverged    = " (Pulling will run --rebase)"
)

type StatusMessage struct {
	Label string
	Help  string
}

var (
	MsgNoUpstream = StatusMessage{Label: LabelNoUpstream, Help: HelpNoUpstream}
	MsgDetached   = StatusMessage{Label: LabelDetached, Help: HelpDetached}
	MsgDiverged   = StatusMessage{Label: StatusDiverged, Help: HelpDiverged}
)

func RenderStatusMessage(msg StatusMessage, maxWidth int) string {
	return RemoteStatusErrorText.Render(msg.Label)
}

func TruncateWithEllipsis(text string, maxWidth int) string {
	runes := []rune(text)
	if len(runes) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

func FormatPullProgress(spinnerView string, lastLine string, width int) string {
	return spinnerView + " " + PullProgressStyle.Width(width).Render(lastLine)
}

const (
	IconGit         = "\uF115"
	IconClock       = "\uF017"
	IconClean       = "\uF00C"
	IconDirty       = "\uF071"
	IconWarning     = "\uF071"
	IconUntracked   = "?"
	IconDiverged    = "⊘"
	IconRemoteError = "\U000F04E7"
	IconBehind      = "\uF063"
	IconAhead       = "\uF062"
	//IconPullRequests = "\uE726"
	IconPullRequests = "\uF03A"
	IconCode         = "\uF09B"
	IconIssues       = "\uEA60"
	//IconOpenPR       = "\uF013"
	IconOpenPR   = "\U000F04C2"
	IconSynced   = "\U000F12D6"
	IconSelector = "▶"
)

const (
	BranchHead     = "HEAD"
	StatusDiverged = "Diverged"
)

var SelectorStyle = lipgloss.NewStyle().
	Foreground(Blue).
	Bold(true)

var HeaderStyle = lipgloss.NewStyle().
	Foreground(Blue)

func FormatHeader(count int) string {
	return HeaderStyle.Render(fmt.Sprintf("\nScan complete. %d repositories found\n", count))
}

var FooterStyle = lipgloss.NewStyle().
	Foreground(SubtleGray).
	PaddingLeft(2)

var ScanningFoundLabelStyle = lipgloss.NewStyle().
	Foreground(TextPrimary).
	PaddingLeft(1)

var ScanningFoundNameStyle = lipgloss.NewStyle().
	Foreground(Green)

func FormatScanningFound(name string) string {
	return ScanningFoundLabelStyle.Render("\uF061 Found git repository:") + " " + ScanningFoundNameStyle.Render(name)
}
