package main

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	Background    = lipgloss.Color("#1a1b26") // dark muted blue-black
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

var TableHeaderStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	Bold(true).
	Align(lipgloss.Left)

var ProjectNameStyle = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Align(lipgloss.Left).
	Width(30).
	MaxWidth(40).
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
	//Bold(true).
	Width(12).
	Render(StatusSynced)

var RemoteStatusGreen = lipgloss.NewStyle().
	Foreground(Green).
	//Bold(true).
	Width(14)

//Bold(true)

var RemoteStatusYellow = lipgloss.NewStyle().
	Foreground(Yellow).
	//Bold(true).
	Width(14)

var LinksStyles = lipgloss.NewStyle().
	Foreground(TextSecondary).MarginRight(1).Bold(true)

var RemoteStatusRed = lipgloss.NewStyle().
	Foreground(Red).
	//Bold(true).
	Width(12)

var RemoteStatusBlue = lipgloss.NewStyle().
	Foreground(Blue).
	//Bold(true).
	Width(12)

var BadgeStyle = lipgloss.NewStyle().
	//Foreground(Blue).
	//Bold(true).
	Width(25)

var IconStyle = lipgloss.NewStyle().
	Foreground(TextPrimary)

// ---------- Badge Styles ----------

var BadgeReadyStyle = lipgloss.NewStyle().
	Foreground(Green).
	Background(lipgloss.Color("#1a3a2e")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(Green)

//Padding(0, 1).
//Bold(true)

// ---------- "MANUAL" Tag ----------

var TagStyle = lipgloss.NewStyle().
	Foreground(TagFg).
	Background(TagBg).
	Padding(0, 1).
	Bold(true).
	MarginLeft(1)

// ---------- Time Ago ----------

var TimeAgoStyle = lipgloss.NewStyle().
	Foreground(TextSecondary)

// ---------- Pull Output Styles ----------

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

// ---------- Footer Keybindings ----------

var KeyStyle = lipgloss.NewStyle().
	Foreground(TextSecondary).
	PaddingLeft(1)

var KeyHighlight = lipgloss.NewStyle().
	Foreground(TextPrimary).
	Bold(true)

const (
	//IconGit      = "\uE0A0"
	IconGit   = "\uF115"
	IconClock = "\uF017"
	//IconClean    = "✓"
	IconClean = "\uF00C"
	//IconDirty    = "●"
	IconDirty = "\uF071"
	//IconConflict = "✗"
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
