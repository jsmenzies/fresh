package listing

import (
	"strings"
	"testing"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
)

func TestBuildPullCompletionInfoMessage_FailureStripsErrorPrefix(t *testing.T) {
	t.Parallel()

	activity := domain.PullingActivity{
		CommandCompletion: domain.CommandCompletion{
			Complete: true,
			Outcome: domain.CommandOutcome{
				ExitCode:      1,
				FailureReason: "error: cannot fast-forward",
			},
		},
	}

	result := buildPullCompletionInfoMessage(activity)
	if !result.OK {
		t.Fatal("buildPullCompletionInfoMessage() expected OK result")
	}
	if result.Message.Tone != InfoTonePullFailure {
		t.Fatalf("buildPullCompletionInfoMessage() tone = %v, want %v", result.Message.Tone, InfoTonePullFailure)
	}
	if result.Message.Text != "cannot fast-forward" {
		t.Fatalf("buildPullCompletionInfoMessage() text = %q, want %q", result.Message.Text, "cannot fast-forward")
	}
}

func TestBuildPullCompletionInfoMessage_FailureStripsErrorPrefixFromLastLine(t *testing.T) {
	t.Parallel()

	activity := domain.PullingActivity{
		LineBuffer: domain.LineBuffer{
			Lines: []string{"some progress", "Error: remote rejected"},
		},
		CommandCompletion: domain.CommandCompletion{
			Complete: true,
			Outcome: domain.CommandOutcome{
				ExitCode: 1,
			},
		},
	}

	result := buildPullCompletionInfoMessage(activity)
	if result.Message.Text != "remote rejected" {
		t.Fatalf("buildPullCompletionInfoMessage() text = %q, want %q", result.Message.Text, "remote rejected")
	}
}

func TestRenderInfoMessage_PullFailureUsesRedLabelAndWhiteDetail(t *testing.T) {
	t.Parallel()

	got := renderInfoMessage(InfoMessage{
		Text: "cannot fast-forward",
		Tone: InfoTonePullFailure,
	}, 80)

	wantLabel := common.PullOutputError.Render(pullFailedLabel)
	if !strings.Contains(got, wantLabel) {
		t.Fatalf("renderInfoMessage() = %q, want red label fragment %q", got, wantLabel)
	}

	wantDetail := common.PullOutputUpToDate.Render(": cannot fast-forward")
	if !strings.Contains(got, wantDetail) {
		t.Fatalf("renderInfoMessage() = %q, want white detail fragment %q", got, wantDetail)
	}
}
