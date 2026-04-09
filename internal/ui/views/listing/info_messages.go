package listing

import (
	"fmt"
	"strings"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
)

type InfoTone int

const (
	InfoToneSubtle InfoTone = iota
	InfoTonePrimary
	InfoToneSuccess
	InfoToneWarn
	InfoToneError
)

type InfoMessage struct {
	Text string
	Tone InfoTone
}

type TimedInfoMessage struct {
	Message   InfoMessage
	ExpiresAt time.Time
}

type InfoRuntime struct {
	Phase                uint64
	Now                  time.Time
	RecentActivityByRepo map[string][]TimedInfoMessage
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

	if summary := buildMyPullRequestSummary(repo.PullRequests); summary != "" {
		appendMessage(InfoMessage{Text: summary, Tone: InfoToneSubtle})
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

func collectActiveActivityInfoMessage(repo domain.Repository, infoWidth int) (InfoMessage, bool) {
	infoWidth = normalizeInfoWidth(infoWidth)

	switch activity := repo.Activity.(type) {
	case *domain.PullingActivity:
		if activity.Complete {
			return InfoMessage{}, false
		}
		lastLine := activity.GetLastLine()
		truncated := common.TruncateWithEllipsis(lastLine, max(1, infoWidth-3))
		return InfoMessage{Text: common.FormatPullProgress(activity.Spinner.View(), truncated, max(1, infoWidth-2)), Tone: InfoTonePrimary}, true
	case *domain.PruningActivity:
		if activity.Complete {
			return InfoMessage{}, false
		}
		lastLine := activity.GetLastLine()
		truncated := common.TruncateWithEllipsis(lastLine, max(1, infoWidth-3))
		return InfoMessage{Text: common.FormatPullProgress(activity.Spinner.View(), truncated, max(1, infoWidth-2)), Tone: InfoTonePrimary}, true
	default:
		return InfoMessage{}, false
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

func buildPullOutputInfoMessage(lastLine string, exitCode int) InfoMessage {
	lowerLine := strings.ToLower(lastLine)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal") {
		return InfoMessage{Text: lastLine, Tone: InfoToneError}
	}

	if exitCode == 0 {
		if strings.Contains(lowerLine, "up to date") || strings.Contains(lowerLine, "up-to-date") {
			return InfoMessage{Text: lastLine, Tone: InfoTonePrimary}
		}
		if strings.Contains(lowerLine, "done") ||
			(strings.Contains(lowerLine, "file") && strings.Contains(lowerLine, "changed")) {
			return InfoMessage{Text: lastLine, Tone: InfoToneSuccess}
		}
	}

	return InfoMessage{Text: lastLine, Tone: InfoToneWarn}
}

func buildPruneCompletionInfoMessage(activity domain.PruningActivity) (InfoMessage, bool) {
	if !activity.Complete {
		return InfoMessage{}, false
	}

	if activity.DeletedCount == 0 {
		for _, line := range activity.Lines {
			if strings.HasPrefix(line, "Failed: ") {
				return InfoMessage{Text: strings.TrimPrefix(line, "Failed: "), Tone: InfoToneError}, true
			}
		}
		return InfoMessage{Text: "No branches to prune", Tone: InfoToneWarn}, true
	}

	return InfoMessage{Text: fmt.Sprintf("Deleted %d branches", activity.DeletedCount), Tone: InfoToneSuccess}, true
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
	default:
		return common.TextGrey.Render(text)
	}
}

func normalizeInfoWidth(infoWidth int) int {
	if infoWidth < 1 {
		return 1
	}
	return infoWidth
}
