package common

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

const (
	SubtleGray    = lipgloss.Color("#5b6078")
	TextPrimary   = lipgloss.Color("#FFFFFF")
	TextSecondary = lipgloss.Color("#a9b1d6")
	TextBranch    = lipgloss.Color("#C27AFF")

	Green  = lipgloss.Color("#06DF71")
	Yellow = lipgloss.Color("#FEC700")
	Red    = lipgloss.Color("#FF6367")
	Blue   = lipgloss.Color("#52A2FF")

	DividerColor = lipgloss.Color("#414868")
	TagBg        = lipgloss.Color("#3b4261")
	TagFg        = lipgloss.Color("#cfc9c2")
)

func NewGreenDotSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	return s
}

func NewSecondaryDotSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(TextSecondary)
	return s
}

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#9ece6a"))

var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true).
	Align(lipgloss.Left)

var ProjectNameStyle = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Align(lipgloss.Left).
	Width(28).
	MaxWidth(40).
	AlignHorizontal(lipgloss.Left)

var SelectedProjectNameStyle = lipgloss.NewStyle().
	Foreground(Blue).
	Align(lipgloss.Left).
	Width(28).
	MaxWidth(40).
	AlignHorizontal(lipgloss.Left).
	Bold(true)

var BranchNameStyle = lipgloss.NewStyle().
	Foreground(TextBranch).
	Align(lipgloss.Left).
	Width(8).
	MaxWidth(12).
	Height(1).
	MaxHeight(1).
	AlignHorizontal(lipgloss.Left)

var BranchNameEmpty = lipgloss.NewStyle().
	Foreground(SubtleGray).
	Render("-")

var LastUpdateTime = lipgloss.NewStyle().
	Foreground(SubtleGray)

var LocalStatusClean = lipgloss.NewStyle().
	Foreground(Green).
	Width(12).
	Render(IconClean + " " + StatusClean)

var LocalStatusDirty = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(12).
	Render(IconDirty + " " + StatusDirty)

var LocalStatusConflict = lipgloss.NewStyle().
	Foreground(Red).
	Width(12).
	Render(IconConflict + " " + StatusConflict)

var RemoteStatusSynced = lipgloss.NewStyle().
	Foreground(SubtleGray).
	Width(12).
	Render(StatusSynced)

var RemoteStatusGreen = lipgloss.NewStyle().
	Foreground(Green).
	Width(14)

var RemoteStatusYellow = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(14)

var LinkStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true)

var LinksStyle = lipgloss.NewStyle().
	Width(8)

var RemoteStatusRed = lipgloss.NewStyle().
	Foreground(Red).
	Width(12)

var RemoteStatusBlue = lipgloss.NewStyle().
	Foreground(Blue).
	Width(12)

var BadgeStyle = lipgloss.NewStyle().
	Width(20).
	Inline(true).
	MarginLeft(2)

var IconStyle = lipgloss.NewStyle().
	Foreground(TextPrimary)

var BadgeReadyStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000")).
	Background(Green).
	Padding(0, 1).
	Bold(true).
	Inline(true).
	Height(1).
	MaxHeight(1).
	MarginLeft(2)

var TagStyle = lipgloss.NewStyle().
	Foreground(TagFg).
	Background(TagBg).
	Padding(0, 1).
	Bold(true).
	Inline(true).
	Height(1).
	MaxHeight(1).
	MarginLeft(2)

var TimeAgoStyle = lipgloss.NewStyle().
	Foreground(TextSecondary)

var PullOutputSuccess = lipgloss.NewStyle().
	Foreground(Green).
	Width(60).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputWarn = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(60).
	Height(1).
	MaxHeight(1).
	Inline(true)

var PullOutputError = lipgloss.NewStyle().
	Foreground(Red).
	Width(60).
	Height(1).
	MaxHeight(1).
	Inline(true)

var KeyStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	PaddingLeft(1)

var KeyHighlight = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Bold(true)

const (
	IconGit          = "\uF115"
	IconClock        = "\uF017"
	IconClean        = "\uF00C"
	IconDirty        = "\uF071"
	IconConflict     = "\uEA87"
	IconDiverged     = "⊘"
	IconBehind       = "\uF063"
	IconAhead        = "\uF062"
	IconPullRequests = "\uE726"
	IconCode         = "\uF09B"
	IconIssues       = "\uEA60"
	IconOpenPR       = "\uF013"
)

const (
	StatusClean    = "Clean"
	StatusDirty    = "Dirty"
	StatusConflict = "Conflict"
	StatusSynced   = "Synced"
	StatusBehind   = "Behind"
	StatusDiverged = "Diverged"
	StatusUpToDate = "up to date"

	ActionUpdating = "updating"
	ActionPulling  = "pulling..."

	BadgeManual = "MANUAL"
	BadgeReady  = "READY"
	TimeJustNow = "just now"
	TimeUnknown = "unknown"
)

const (
	Padding = 2
)

var (
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

var SelectorStyle = lipgloss.NewStyle().
	Foreground(Blue).
	Width(2).
	Bold(true)

var SelectedRowStyle = lipgloss.NewStyle()
