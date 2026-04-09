package git

import "testing"

func TestClassifyCheckContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		context ghStatusCheckContext
		want    string
	}{
		{
			name:    "success conclusion is passed",
			context: ghStatusCheckContext{Conclusion: "SUCCESS"},
			want:    "passed",
		},
		{
			name:    "failed conclusion is failed",
			context: ghStatusCheckContext{Conclusion: "FAILURE"},
			want:    "failed",
		},
		{
			name:    "skipped conclusion is skipped",
			context: ghStatusCheckContext{Conclusion: "SKIPPED"},
			want:    "skipped",
		},
		{
			name:    "in progress status is running",
			context: ghStatusCheckContext{Status: "IN_PROGRESS"},
			want:    "running",
		},
		{
			name:    "queued status is waiting",
			context: ghStatusCheckContext{Status: "QUEUED"},
			want:    "waiting",
		},
		{
			name:    "pending state is waiting",
			context: ghStatusCheckContext{State: "PENDING"},
			want:    "waiting",
		},
		{
			name:    "failure state is failed",
			context: ghStatusCheckContext{State: "FAILURE"},
			want:    "failed",
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
