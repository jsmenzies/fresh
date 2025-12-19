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
	InfoWidth    = 42
	Padding      = 2

	RowHeight         = 1
	BranchWidth       = 8
	MaxBranchWidth    = 12
	LocalStatusWidth  = 14
	RemoteStatusWidth = 11
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

var ProjectNameStyle = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Align(lipgloss.Left).
	Width(22).
	MaxWidth(30).
	AlignHorizontal(lipgloss.Left)

var branchBaseStyle = lipgloss.NewStyle().
	Align(lipgloss.Left).
	Width(BranchWidth).
	MaxWidth(MaxBranchWidth).
	Height(RowHeight).
	MaxHeight(RowHeight).
	AlignHorizontal(lipgloss.Left)

var BranchNameStyle = branchBaseStyle.
	Foreground(TextBranch)

var BranchNameEmpty = branchBaseStyle.
	Foreground(SubtleGray).
	Render("")

var BranchNameHead = branchBaseStyle.
	Foreground(SubtleGray).
	Render(BranchHead)

var localStatusBaseStyle = lipgloss.NewStyle().
	Width(LocalStatusWidth).
	MaxWidth(LocalStatusWidth).
	Height(RowHeight).
	MaxHeight(RowHeight).
	AlignHorizontal(lipgloss.Left)

var LocalStatusClean = localStatusBaseStyle.
	Foreground(Green).
	Render(IconClean + " " + StatusClean)

var LocalStatusDirty = localStatusBaseStyle.
	Foreground(Yellow).
	Render(IconDirty + " " + StatusDirty)

var LocalStatusUntracked = localStatusBaseStyle.
	Foreground(Yellow).
	Render(IconUntracked + " " + StatusUntracked)

var LocalStatusError = localStatusBaseStyle.
	Render("")

var remoteStatusBaseStyle = lipgloss.NewStyle().
	Width(RemoteStatusWidth).
	MaxWidth(RemoteStatusWidth).
	Height(RowHeight).
	MaxHeight(RowHeight).
	AlignHorizontal(lipgloss.Left)

var RemoteStatusSynced = remoteStatusBaseStyle.
	Foreground(SubtleGreen).
	Render(IconSynced)

var RemoteStatusError = remoteStatusBaseStyle.
	Foreground(SubtleRed).
	Render(IconRemoteError)

var RemoteStatusCountsStyle = remoteStatusBaseStyle.
	Foreground(Blue)

var RemoteStatusUpdating = remoteStatusBaseStyle.
	Align(lipgloss.Left)

func RemoteStatusCounts(behind int, ahead int) string {
	content := ""
	if behind > 0 && ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d / "+IconBehind+" %d", ahead, behind)
	} else if behind > 0 {
		content = fmt.Sprintf(IconBehind+" %d", behind)
	} else if ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d", ahead)
	}

	return RemoteStatusCountsStyle.Render(content)
}

var RemoteStatusErrorText = lipgloss.NewStyle().
	Foreground(SubtleRed)

var RemoteStatusDivergedText = lipgloss.NewStyle().
	Foreground(Yellow).
	Render(StatusDiverged)

var RemoteStatusErrorHelpText = lipgloss.NewStyle().
	Foreground(SubtleGray)

var LinkStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true)

var LinksStyle = lipgloss.NewStyle().
	Width(8)

var BadgeStyle = lipgloss.NewStyle().
	Width(8).
	Inline(true).
	MarginLeft(2)

var TimeAgoStyle = lipgloss.NewStyle().
	Foreground(TextSecondary)

var PullOutputSuccess = lipgloss.NewStyle().
	Foreground(Green).
	Width(InfoWidth).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputUpToDate = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Width(InfoWidth).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputWarn = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(InfoWidth).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputError = lipgloss.NewStyle().
	Foreground(Red).
	Width(InfoWidth).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullProgressStyle = lipgloss.NewStyle().
	Width(InfoWidth - 2)

var InfoStyle = lipgloss.NewStyle().
	Width(InfoWidth).
	MaxWidth(InfoWidth).MaxHeight(1)

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
	available := maxWidth - len(msg.Label)
	if available < 0 {
		available = 0
	}
	truncatedHelp := TruncateWithEllipsis(msg.Help, available)
	return RemoteStatusErrorText.Render(msg.Label) + RemoteStatusErrorHelpText.Render(truncatedHelp)
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

func FormatPullProgress(spinnerView string, lastLine string) string {
	return spinnerView + " " + PullProgressStyle.Render(lastLine)
}

const (
	IconGit         = "\uF115"
	IconClock       = "\uF017"
	IconClean       = "\uF00C"
	IconDirty       = "\uF071"
	IconUntracked   = ""
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
	BranchHead      = "HEAD"
	StatusClean     = "Clean"
	StatusDirty     = "Dirty"
	StatusUntracked = "Untracked"
	StatusSynced    = "Synced"
	StatusBehind    = "Behind"
	StatusAhead     = "Ahead"
	StatusDiverged  = "Diverged"
	StatusUpToDate  = "up to date"

	ActionUpdating = "updating"
	ActionPulling  = "pulling..."

	BadgeManual = "MANUAL"
	BadgeReady  = "READY"
	TimeJustNow = "just now"
	TimeUnknown = "unknown"
)

var SelectorStyle = lipgloss.NewStyle().
	Foreground(Blue).
	Width(2).
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
