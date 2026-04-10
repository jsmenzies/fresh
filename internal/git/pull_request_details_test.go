package git

import (
	"errors"
	"testing"
	"time"
)

func TestClassifyCheckContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		context ghStatusCheckContext
		want    checkSummaryClass
	}{
		{
			name:    "success conclusion is passed",
			context: ghStatusCheckContext{Conclusion: "SUCCESS"},
			want:    checkSummaryPassed,
		},
		{
			name:    "failed conclusion is failed",
			context: ghStatusCheckContext{Conclusion: "FAILURE"},
			want:    checkSummaryFailed,
		},
		{
			name:    "skipped conclusion is skipped",
			context: ghStatusCheckContext{Conclusion: "SKIPPED"},
			want:    checkSummarySkipped,
		},
		{
			name:    "in progress status is running",
			context: ghStatusCheckContext{Status: "IN_PROGRESS"},
			want:    checkSummaryRunning,
		},
		{
			name:    "queued status is waiting",
			context: ghStatusCheckContext{Status: "QUEUED"},
			want:    checkSummaryWaiting,
		},
		{
			name:    "pending state is waiting",
			context: ghStatusCheckContext{State: "PENDING"},
			want:    checkSummaryWaiting,
		},
		{
			name:    "failure state is failed",
			context: ghStatusCheckContext{State: "FAILURE"},
			want:    checkSummaryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyCheckContext(tt.context)
			if got != tt.want {
				t.Fatalf("classifyCheckContext(%+v) = %q, want %q", tt.context, got, tt.want)
			}
		})
	}
}

func TestSummarizePullRequestChecks(t *testing.T) {
	t.Parallel()

	summary := summarizePullRequestChecks([]ghStatusCheckContext{
		{Conclusion: "SUCCESS"},
		{Status: "IN_PROGRESS"},
		{Status: "QUEUED"},
		{Conclusion: "FAILURE"},
		{Conclusion: "SKIPPED"},
	})

	if summary.Total != 5 {
		t.Fatalf("total = %d, want 5", summary.Total)
	}
	if summary.Passed != 1 {
		t.Fatalf("passed = %d, want 1", summary.Passed)
	}
	if summary.Running != 1 {
		t.Fatalf("running = %d, want 1", summary.Running)
	}
	if summary.Waiting != 1 {
		t.Fatalf("waiting = %d, want 1", summary.Waiting)
	}
	if summary.Failed != 1 {
		t.Fatalf("failed = %d, want 1", summary.Failed)
	}
	if summary.Skipped != 1 {
		t.Fatalf("skipped = %d, want 1", summary.Skipped)
	}
}

func TestGitHubLoginCacheGet_UsesCachedValueWithinTTL(t *testing.T) {
	t.Parallel()

	var cache githubLoginCache
	now := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	loadCalls := 0

	loader := func() (string, error) {
		loadCalls++
		return "octocat\n", nil
	}

	first := cache.get(now, time.Hour, loader)
	second := cache.get(now.Add(30*time.Minute), time.Hour, loader)

	if first != "octocat" {
		t.Fatalf("first = %q, want %q", first, "octocat")
	}
	if second != "octocat" {
		t.Fatalf("second = %q, want %q", second, "octocat")
	}
	if loadCalls != 1 {
		t.Fatalf("loader calls = %d, want 1", loadCalls)
	}
}

func TestGitHubLoginCacheGet_RefreshesAfterTTL(t *testing.T) {
	t.Parallel()

	var cache githubLoginCache
	now := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	loadCalls := 0

	loader := func() (string, error) {
		loadCalls++
		if loadCalls == 1 {
			return "octocat", nil
		}
		return "monalisa", nil
	}

	first := cache.get(now, time.Hour, loader)
	second := cache.get(now.Add(61*time.Minute), time.Hour, loader)

	if first != "octocat" {
		t.Fatalf("first = %q, want %q", first, "octocat")
	}
	if second != "monalisa" {
		t.Fatalf("second = %q, want %q", second, "monalisa")
	}
	if loadCalls != 2 {
		t.Fatalf("loader calls = %d, want 2", loadCalls)
	}
}

func TestGitHubLoginCacheGet_CachesEmptyValueOnLoaderError(t *testing.T) {
	t.Parallel()

	var cache githubLoginCache
	now := time.Date(2026, time.January, 2, 10, 0, 0, 0, time.UTC)
	loadCalls := 0

	loader := func() (string, error) {
		loadCalls++
		return "", errors.New("boom")
	}

	first := cache.get(now, time.Hour, loader)
	second := cache.get(now.Add(30*time.Minute), time.Hour, loader)

	if first != "" {
		t.Fatalf("first = %q, want empty", first)
	}
	if second != "" {
		t.Fatalf("second = %q, want empty", second)
	}
	if loadCalls != 1 {
		t.Fatalf("loader calls = %d, want 1", loadCalls)
	}
}
