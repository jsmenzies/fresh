package git

import (
	"fresh/internal/pullrequests"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func makePRNode(opts ...func(*gqlPullRequestNode)) gqlPullRequestNode {
	node := gqlPullRequestNode{}
	for _, opt := range opts {
		opt(&node)
	}
	return node
}

func withDraft() func(*gqlPullRequestNode) {
	return func(node *gqlPullRequestNode) {
		node.IsDraft = true
	}
}

func withMergeState(value string) func(*gqlPullRequestNode) {
	return func(node *gqlPullRequestNode) {
		node.MergeStateStatus = strPtr(value)
	}
}

func withReviewDecision(value string) func(*gqlPullRequestNode) {
	return func(node *gqlPullRequestNode) {
		node.ReviewDecision = strPtr(value)
	}
}

func withRollup(rollup gqlStatusCheckRollup) func(*gqlPullRequestNode) {
	return func(node *gqlPullRequestNode) {
		node.Commits.Nodes = []struct {
			Commit struct {
				StatusCheckRollup *gqlStatusCheckRollup `json:"statusCheckRollup"`
			} `json:"commit"`
		}{
			{
				Commit: struct {
					StatusCheckRollup *gqlStatusCheckRollup `json:"statusCheckRollup"`
				}{StatusCheckRollup: &rollup},
			},
		}
	}
}

func TestClassifyMyPullRequestDecisionMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node gqlPullRequestNode
		want pullrequests.Status
	}{
		{
			name: "draft is blocked",
			node: makePRNode(
				withDraft(),
				withReviewDecision("APPROVED"),
			),
			want: pullrequests.StatusBlocked,
		},
		{
			name: "merge blocked is blocked",
			node: makePRNode(
				withMergeState("BLOCKED"),
			),
			want: pullrequests.StatusBlocked,
		},
		{
			name: "failing checks are blocked",
			node: makePRNode(
				withRollup(gqlStatusCheckRollup{State: "FAILURE"}),
			),
			want: pullrequests.StatusBlocked,
		},
		{
			name: "changes requested is blocked",
			node: makePRNode(
				withReviewDecision("CHANGES_REQUESTED"),
			),
			want: pullrequests.StatusBlocked,
		},
		{
			name: "pending checks is checks",
			node: makePRNode(
				withRollup(gqlStatusCheckRollup{State: "PENDING"}),
			),
			want: pullrequests.StatusChecks,
		},
		{
			name: "approved with no pending checks is ready",
			node: makePRNode(
				withReviewDecision("APPROVED"),
			),
			want: pullrequests.StatusReady,
		},
		{
			name: "review required remains review",
			node: makePRNode(
				withReviewDecision("REVIEW_REQUIRED"),
				withMergeState("CLEAN"),
			),
			want: pullrequests.StatusReview,
		},
		{
			name: "no review required and clean merge is ready",
			node: makePRNode(
				withMergeState("CLEAN"),
			),
			want: pullrequests.StatusReady,
		},
		{
			name: "unstable merge without rollup details stays checks",
			node: makePRNode(
				withMergeState("UNSTABLE"),
			),
			want: pullrequests.StatusChecks,
		},
		{
			name: "default requires review",
			node: makePRNode(),
			want: pullrequests.StatusReview,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyMyPullRequest(tt.node)
			if got != tt.want {
				t.Fatalf("classifyMyPullRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassifyMyPullRequest_ContextDrivenPendingChecks(t *testing.T) {
	t.Parallel()

	node := makePRNode(
		withRollup(gqlStatusCheckRollup{
			State: "SUCCESS",
			Contexts: struct {
				Nodes []gqlStatusContextNode `json:"nodes"`
			}{
				Nodes: []gqlStatusContextNode{{Status: "IN_PROGRESS"}},
			},
		}),
	)

	got := classifyMyPullRequest(node)
	if got != pullrequests.StatusChecks {
		t.Fatalf("classifyMyPullRequest() = %q, want %q", got, pullrequests.StatusChecks)
	}
}

func TestClassifyMyPullRequest_ContextDrivenFailingChecks(t *testing.T) {
	t.Parallel()

	node := makePRNode(
		withRollup(gqlStatusCheckRollup{
			State: "SUCCESS",
			Contexts: struct {
				Nodes []gqlStatusContextNode `json:"nodes"`
			}{
				Nodes: []gqlStatusContextNode{{Conclusion: "FAILURE"}},
			},
		}),
	)

	got := classifyMyPullRequest(node)
	if got != pullrequests.StatusBlocked {
		t.Fatalf("classifyMyPullRequest() = %q, want %q", got, pullrequests.StatusBlocked)
	}
}
