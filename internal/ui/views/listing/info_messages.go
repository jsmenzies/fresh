package listing

import (
	"fmt"
	"fresh/internal/textutil"
	"strings"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"charm.land/lipgloss/v2"
)

type InfoTone int

const (
	InfoToneSubtle InfoTone = iota
	InfoTonePrimary
	InfoToneSuccess
	InfoToneWarn
	InfoToneError
	InfoTonePullRequestSummary
)

type InfoMessage struct {
	Text   string
	Tone   InfoTone
	Pinned bool
}

type TimedInfoMessage struct {
	Message   InfoMessage
	ExpiresAt time.Time
}

type InfoRuntime struct {
	Phase                uint64
	Now                  time.Time
	RecentActivityByRepo map[string][]TimedInfoMessage
	PullRequestSyncing   bool
	BlockedSpinner       string
	ReadySpinner         string
}

type InfoMessageResult struct {
	Message InfoMessage
	OK      bool
}

func collectStatusInfoMessages(repo domain.Repository) []InfoMessage {
	messages := make([]InfoMessage, 0, 8)

	appendMessage := func(msg InfoMessage) {
		if msg.Text == "" {
			return
		}
		messages = append(messages, msg)
	}

	switch state := repo.RemoteState.(type) {
	case domain.RemoteError:
		appendMessage(InfoMessage{Text: state.Message, Tone: InfoToneError})
	}

	if summary, pinned := buildMyPullRequestSummary(repo.PullRequests); summary != "" {
		appendMessage(InfoMessage{Text: summary, Tone: InfoTonePullRequestSummary, Pinned: pinned})
	}

	mergedCount := len(repo.Branches.Merged)
	if mergedCount > 0 {
		mergedText := "branches"
		if mergedCount == 1 {
			mergedText = "branch"
		}
		appendMessage(InfoMessage{Text: fmt.Sprintf("%d prunable %s", mergedCount, mergedText), Tone: InfoToneSubtle})
	}

	if repo.StashCount > 0 {
		stashText := "stashes"
		if repo.StashCount == 1 {
			stashText = "stash"
		}
		appendMessage(InfoMessage{Text: fmt.Sprintf("%d %s", repo.StashCount, stashText), Tone: InfoToneSubtle})
	}

	if _, ok := repo.RemoteState.(domain.NoUpstream); ok {
		appendMessage(InfoMessage{Text: common.LabelNoUpstream, Tone: InfoToneSubtle})
	}
	if _, ok := repo.RemoteState.(domain.DetachedRemote); ok {
		appendMessage(InfoMessage{Text: common.LabelDetached, Tone: InfoToneSubtle})
	}
	if _, ok := repo.RemoteState.(domain.Diverged); ok {
		appendMessage(InfoMessage{Text: common.StatusDiverged, Tone: InfoToneWarn})
	}

	return messages
}

func collectActiveActivityInfoMessage(repo domain.Repository, infoWidth int) InfoMessageResult {
	infoWidth = normalizeInfoWidth(infoWidth)

	switch activity := repo.Activity.(type) {
	case *domain.PullingActivity:
		return formatActiveProgressInfoMessage(activity.Complete, activity.Spinner.View(), activity.GetLastLine(), infoWidth)
	case *domain.PruningActivity:
		return formatActiveProgressInfoMessage(activity.Complete, activity.Spinner.View(), activity.GetLastLine(), infoWidth)
	default:
		return InfoMessageResult{}
	}
}

func formatActiveProgressInfoMessage(complete bool, spinnerView, lastLine string, infoWidth int) InfoMessageResult {
	if complete {
		return InfoMessageResult{}
	}

	truncated := common.TruncateWithEllipsis(lastLine, max(1, infoWidth-3))
	return InfoMessageResult{
		Message: InfoMessage{
			Text: common.FormatPullProgress(spinnerView, truncated, max(1, infoWidth-2)),
			Tone: InfoTonePrimary,
		},
		OK: true,
	}
}

func collectRecentActivityInfoMessages(runtime InfoRuntime, repoPath string) []InfoMessage {
	if repoPath == "" || len(runtime.RecentActivityByRepo) == 0 {
		return nil
	}

	timed := runtime.RecentActivityByRepo[repoPath]
	if len(timed) == 0 {
		return nil
	}

	now := runtime.Now
	if now.IsZero() {
		now = time.Now()
	}

	messages := make([]InfoMessage, 0, len(timed))
	for _, item := range timed {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			continue
		}
		messages = append(messages, item.Message)
	}

	return messages
}

func filterPinnedInfoMessages(messages []InfoMessage) []InfoMessage {
	pinned := make([]InfoMessage, 0, len(messages))
	for _, message := range messages {
		if message.Pinned {
			pinned = append(pinned, message)
		}
	}
	return pinned
}

func buildPullCompletionInfoMessage(activity domain.PullingActivity) InfoMessageResult {
	if !activity.Complete {
		return InfoMessageResult{}
	}

	if !activity.Outcome.IsSuccess() {
		reason := textutil.FirstNonEmptyTrimmed(activity.Outcome.FailureReason, activity.GetLastLine(), "pull failed")
		return InfoMessageResult{
			Message: InfoMessage{Text: fmt.Sprintf("Pull failed: %s", reason), Tone: InfoToneWarn},
			OK:      true,
		}
	}

	return InfoMessageResult{
		Message: InfoMessage{Text: "Pull completed successfully", Tone: InfoToneSuccess},
		OK:      true,
	}
}

func buildPruneCompletionInfoMessage(activity domain.PruningActivity) InfoMessageResult {
	if !activity.Complete {
		return InfoMessageResult{}
	}

	if !activity.Outcome.IsSuccess() {
		reason := textutil.FirstNonEmptyTrimmed(
			activity.Outcome.FailureReason,
			textutil.FirstLineWithPrefixTrimmed(activity.Lines, "Failed: "),
			"one or more branches could not be pruned",
		)
		return InfoMessageResult{
			Message: InfoMessage{Text: fmt.Sprintf("Prune failed: %s", reason), Tone: InfoToneWarn},
			OK:      true,
		}
	}

	if activity.DeletedCount == 1 {
		return InfoMessageResult{
			Message: InfoMessage{Text: "Pruned 1 branch", Tone: InfoToneSuccess},
			OK:      true,
		}
	}

	return InfoMessageResult{
		Message: InfoMessage{Text: fmt.Sprintf("Pruned %d branches", activity.DeletedCount), Tone: InfoToneSuccess},
		OK:      true,
	}
}

func renderInfoMessage(msg InfoMessage, infoWidth int) string {
	infoWidth = normalizeInfoWidth(infoWidth)
	text := common.TruncateWithEllipsis(msg.Text, infoWidth)

	switch msg.Tone {
	case InfoTonePrimary:
		return common.PullOutputUpToDate.Width(infoWidth).Render(text)
	case InfoToneSuccess:
		return common.PullOutputSuccess.Width(infoWidth).Render(text)
	case InfoToneWarn:
		return common.PullOutputWarn.Width(infoWidth).Render(text)
	case InfoToneError:
		return common.PullOutputError.Width(infoWidth).Render(text)
	case InfoTonePullRequestSummary:
		return renderMyPullRequestSummaryInfo(text, infoWidth)
	default:
		return common.TextGrey.Render(text)
	}
}

func renderMyPullRequestSummaryInfo(text string, infoWidth int) string {
	text = common.TruncateWithEllipsis(text, infoWidth)

	const prefix = "My PRs:"
	if !strings.HasPrefix(text, prefix) {
		return common.TextGrey.Render(text)
	}

	labelStyle := lipgloss.NewStyle().Foreground(common.TextPrimary)
	separatorStyle := labelStyle
	readyStyle := lipgloss.NewStyle().Foreground(common.Green)
	blockedStyle := lipgloss.NewStyle().Foreground(common.Red)
	waitingStyle := lipgloss.NewStyle().Foreground(common.Yellow)

	rest := strings.TrimSpace(strings.TrimPrefix(text, prefix))
	if rest == "" {
		return labelStyle.Render(prefix)
	}

	parts := strings.Split(rest, ", ")
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		lower := strings.ToLower(part)
		switch {
		case strings.Contains(lower, "blocked"):
			rendered = append(rendered, blockedStyle.Render(part))
		case strings.Contains(lower, "ready"), strings.Contains(lower, "mergeable"):
			rendered = append(rendered, readyStyle.Render(part))
		case strings.Contains(lower, "check"), strings.Contains(lower, "review"), strings.Contains(lower, "waiting"):
			rendered = append(rendered, waitingStyle.Render(part))
		default:
			rendered = append(rendered, labelStyle.Render(part))
		}
	}

	return labelStyle.Render(prefix) + separatorStyle.Render(" ") + strings.Join(rendered, separatorStyle.Render(", "))
}

func normalizeInfoWidth(infoWidth int) int {
	if infoWidth < 1 {
		return 1
	}
	return infoWidth
}
