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

func NewRefreshSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Blue)
	return s
}

func NewPullSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(Green)
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
	Render("")

var BranchNameHead = lipgloss.NewStyle().
	Foreground(SubtleGray).
	Width(8).
	MaxWidth(12).
	Render(BranchHead)

var LocalStatusClean = lipgloss.NewStyle().
	Foreground(Green).
	Width(14).
	Render(IconClean + " " + StatusClean)

var LocalStatusDirty = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(14).
	Render(IconDirty + " " + StatusDirty)

var LocalStatusUntracked = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(14).
	Render(IconUntracked + " " + StatusUntracked)

var LocalStatusError = lipgloss.NewStyle().
	Width(14).
	Render("")

var RemoteStatusSynced = lipgloss.NewStyle().
	Foreground(SubtleGreen).
	Width(10).
	Render(IconSynced)

var RemoteStatusError = lipgloss.NewStyle().
	Foreground(SubtleRed).
	Width(10).
	Render(IconRemoteError)

var RemoteStatusErrorText = lipgloss.NewStyle().
	Foreground(SubtleRed)

//Width(20)

var RemoteStatusErrorHelpText = lipgloss.NewStyle().
	Foreground(SubtleGray)

//Width(60)

var RemoteStatusDiverged = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(12).
	Render(IconDiverged + " " + StatusDiverged)

func RemoteStatusCounts(behind int, ahead int) string {
	content := ""
	if behind > 0 && ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d / "+IconBehind+" %d", ahead, behind)
	} else if behind > 0 {
		content = fmt.Sprintf(IconBehind+" %d", behind)
	} else if ahead > 0 {
		content = fmt.Sprintf(IconAhead+" %d", ahead)
	}

	return lipgloss.NewStyle().
		Foreground(Blue).
		Width(10).
		Render(content)
}

var RemoteStatusBehind = lipgloss.NewStyle().
	Foreground(Yellow).
	Width(12).
	Render(IconBehind + " " + StatusBehind)

var RemoteStatusAhead = lipgloss.NewStyle().
	Foreground(Green).
	Width(12).
	Render(IconAhead + " " + StatusAhead)

var RemoteStatusUpdating = lipgloss.NewStyle().
	Width(10).
	Align(lipgloss.Left)

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

const (
	Padding = 2
)

var SelectorStyle = lipgloss.NewStyle().
	Foreground(Blue).
	Width(2).
	Bold(true)
