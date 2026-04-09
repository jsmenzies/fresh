package listing

import (
	"strings"
	"testing"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
)

// ============================================================================
// buildSelector tests
// ============================================================================

func TestBuildSelector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		isSelected bool
		want       string
	}{
		{
			name:       "selected shows selector icon",
			isSelected: true,
			want:       common.IconSelector,
		},
		{
			name:       "not selected shows space",
			isSelected: false,
			want:       " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildSelector(tt.isSelected)
			if !strings.Contains(got, tt.want) {
				t.Errorf("buildSelector(%v) = %q, want it to contain %q", tt.isSelected, got, tt.want)
			}
		})
	}
}

// ============================================================================
// buildProjectName tests
// ============================================================================

func TestBuildProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		project    string
		isSelected bool
		width      int
	}{
		{
			name:       "renders project name",
			project:    "my-cool-project",
			isSelected: false,
			width:      30,
		},
		{
			name:       "renders selected project name",
			project:    "my-cool-project",
			isSelected: true,
			width:      30,
		},
		{
			name:       "short name",
			project:    "foo",
			isSelected: false,
			width:      30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildProjectName(tt.project, tt.isSelected, tt.width)
			if !strings.Contains(got, tt.project) {
				t.Errorf("buildProjectName(%q, %v, %d) = %q, want it to contain %q",
					tt.project, tt.isSelected, tt.width, got, tt.project)
			}
		})
	}
}

func TestBuildProjectName_TruncatesWithEllipsis(t *testing.T) {
	t.Parallel()

	got := buildProjectName("this-is-a-very-long-repository-name", false, 12)
	if !strings.Contains(got, "...") {
		t.Fatalf("buildProjectName() = %q, want ellipsis", got)
	}
	if strings.Contains(got, "this-is-a-very-long-repository-name") {
		t.Fatalf("buildProjectName() = %q, did not expect full untruncated name", got)
	}
}

// ============================================================================
// buildBranchName tests
// ============================================================================

func TestBuildBranchName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		branch domain.Branch
		width  int
		want   string
	}{
		{
			name:   "on branch shows branch name",
			branch: domain.OnBranch{Name: "main"},
			width:  20,
			want:   "main",
		},
		{
			name:   "on feature branch",
			branch: domain.OnBranch{Name: "feature/auth"},
			width:  20,
			want:   "feature/auth",
		},
		{
			name:   "detached head shows HEAD",
			branch: domain.DetachedHead{CommitSHA: "abc123"},
			width:  20,
			want:   common.BranchHead,
		},
		{
			name:   "no branch shows empty",
			branch: domain.NoBranch{Reason: "init"},
			width:  20,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildBranchName(tt.branch, tt.width)
			if !strings.Contains(got, tt.want) {
				t.Errorf("buildBranchName(%v, %d) = %q, want it to contain %q",
					tt.branch, tt.width, got, tt.want)
			}
		})
	}
}

// ============================================================================
// buildLocalStatus tests
// ============================================================================

func TestBuildLocalStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    domain.LocalState
		contains []string
		excludes []string
	}{
		{
			name:     "clean state shows check icon",
			state:    domain.CleanLocalState{},
			contains: []string{common.IconClean},
		},
		{
			name:     "dirty with modified only",
			state:    domain.DirtyLocalState{Modified: 3},
			contains: []string{"~3", common.IconWarning},
		},
		{
			name:     "dirty with added only",
			state:    domain.DirtyLocalState{Added: 2},
			contains: []string{"+2", common.IconWarning},
		},
		{
			name:     "dirty with deleted only",
			state:    domain.DirtyLocalState{Deleted: 1},
			contains: []string{"-1", common.IconWarning},
		},
		{
			name:     "dirty with untracked only",
			state:    domain.DirtyLocalState{Untracked: 5},
			contains: []string{common.IconUntracked + "5", common.IconDiverged},
		},
		{
			name:  "dirty with mixed changes",
			state: domain.DirtyLocalState{Added: 1, Modified: 2, Deleted: 3, Untracked: 4},
			contains: []string{
				"+1", "~2", "-3",
				common.IconUntracked + "4",
			},
		},
		{
			name:  "error state renders empty",
			state: domain.LocalStateError{Message: "something broke"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildLocalStatus(tt.state)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildLocalStatus(%v) = %q, want it to contain %q", tt.state, got, want)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(got, exclude) {
					t.Errorf("buildLocalStatus(%v) = %q, want it to NOT contain %q", tt.state, got, exclude)
				}
			}
		})
	}
}

// ============================================================================
// buildRemoteStatus tests
// ============================================================================

func TestBuildRemoteStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     domain.Repository
		contains []string
	}{
		{
			name:     "synced shows synced icon",
			repo:     makeTestRepository("repo"),
			contains: []string{common.IconSynced},
		},
		{
			name:     "ahead shows ahead icon with count",
			repo:     newTestRepository("repo").RemoteState(domain.Ahead{Count: 3}).Build(),
			contains: []string{common.IconAhead, "3"},
		},
		{
			name:     "behind shows behind icon with count",
			repo:     newTestRepository("repo").RemoteState(domain.Behind{Count: 5}).Build(),
			contains: []string{common.IconBehind, "5"},
		},
		{
			name:     "diverged shows both ahead and behind counts",
			repo:     newTestRepository("repo").RemoteState(domain.Diverged{AheadCount: 2, BehindCount: 7}).Build(),
			contains: []string{common.IconAhead, "2", common.IconBehind, "7"},
		},
		{
			name:     "no upstream shows dash",
			repo:     newTestRepository("repo").RemoteState(domain.NoUpstream{}).Build(),
			contains: []string{"-"},
		},
		{
			name:     "detached remote shows dash",
			repo:     newTestRepository("repo").RemoteState(domain.DetachedRemote{}).Build(),
			contains: []string{"-"},
		},
		{
			name:     "remote error shows error icon",
			repo:     newTestRepository("repo").RemoteState(domain.RemoteError{Message: "timeout"}).Build(),
			contains: []string{common.IconRemoteError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRemoteStatus(tt.repo)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildRemoteStatus(%q) = %q, want it to contain %q",
						tt.name, got, want)
				}
			}
		})
	}
}

// ============================================================================
// buildPullRequestStatus tests
// ============================================================================

func TestBuildPullRequestStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    domain.PullRequestState
		contains []string
		isEmpty  bool
	}{
		{
			name:     "open pull requests show icon and count",
			state:    domain.PullRequestCount{Open: 3, MyOpen: 0},
			contains: []string{"3"},
		},
		{
			name:    "zero pull requests shows empty",
			state:   domain.PullRequestCount{Open: 0, MyOpen: 0},
			isEmpty: true,
		},
		{
			name:     "my open pull request still shows icon and count",
			state:    domain.PullRequestCount{Open: 2, MyOpen: 1},
			contains: []string{"2", "(*)"},
		},
		{
			name:     "unavailable pull requests show dash",
			state:    domain.PullRequestUnavailable{},
			contains: []string{"-"},
		},
		{
			name:     "error state shows remote error icon",
			state:    domain.PullRequestError{Message: "gh not available"},
			contains: []string{common.IconRemoteError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildPullRequestStatus(tt.state, InfoRuntime{})
			if tt.isEmpty {
				trimmed := strings.TrimSpace(got)
				if trimmed != "" {
					t.Errorf("buildPullRequestStatus(%q) = %q, want empty", tt.name, got)
				}
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildPullRequestStatus(%q) = %q, want it to contain %q", tt.name, got, want)
				}
			}
		})
	}
}

func TestBuildPullRequestStatus_SyncingKeepsCurrentStateVisible(t *testing.T) {
	t.Parallel()

	runtime := InfoRuntime{PullRequestSyncing: true, PullRequestSpinner: "⠋"}
	got := buildPullRequestStatus(domain.PullRequestCount{Open: 9, MyOpen: 1}, runtime)

	if strings.Contains(got, "⠋") {
		t.Fatalf("buildPullRequestStatus() = %q, want spinner hidden while syncing", got)
	}
	if !strings.Contains(got, "9") || !strings.Contains(got, "(*)") {
		t.Fatalf("buildPullRequestStatus() = %q, want existing PR summary preserved", got)
	}
}

func TestBuildPullRequestAlert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    domain.PullRequestState
		runtime  InfoRuntime
		contains string
		isEmpty  bool
	}{
		{
			name:    "syncing hides alert even when blocked prs exist",
			state:   domain.PullRequestCount{Open: 9, MyOpen: 1, MyBlocked: 2},
			runtime: InfoRuntime{PullRequestSyncing: true, PullRequestSpinner: "⠋", BlockedSpinner: "█"},
			isEmpty: true,
		},
		{
			name:     "blocked my prs show alert spinner",
			state:    domain.PullRequestCount{MyBlocked: 2},
			runtime:  InfoRuntime{BlockedSpinner: "█"},
			contains: "█",
		},
		{
			name:     "ready my prs show success spinner",
			state:    domain.PullRequestCount{MyReady: 2},
			runtime:  InfoRuntime{ReadySpinner: "●"},
			contains: "●",
		},
		{
			name:     "blocked takes precedence over ready",
			state:    domain.PullRequestCount{MyBlocked: 1, MyReady: 2},
			runtime:  InfoRuntime{BlockedSpinner: "█", ReadySpinner: "●"},
			contains: "█",
		},
		{
			name:    "no blocked or ready prs shows empty",
			state:   domain.PullRequestCount{MyBlocked: 0, MyReady: 0},
			runtime: InfoRuntime{},
			isEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildPullRequestAlert(tt.state, tt.runtime)
			if tt.isEmpty {
				if strings.TrimSpace(got) != "" {
					t.Fatalf("buildPullRequestAlert(%q) = %q, want empty", tt.name, got)
				}
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Fatalf("buildPullRequestAlert(%q) = %q, want %q", tt.name, got, tt.contains)
			}
		})
	}
}

func TestBuildMyPullRequestSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    domain.PullRequestState
		contains []string
		isEmpty  bool
	}{
		{
			name: "shows base summary",
			state: domain.PullRequestCount{
				MyReady:   1,
				MyBlocked: 2,
				MyChecks:  3,
			},
			contains: []string{"My PRs:", "1 ready", "2 blocked", "3 checks"},
		},
		{
			name: "includes review when non-zero",
			state: domain.PullRequestCount{
				MyReady:   0,
				MyBlocked: 1,
				MyChecks:  2,
				MyReview:  1,
			},
			contains: []string{"1 review"},
		},
		{
			name: "hides zero-count buckets",
			state: domain.PullRequestCount{
				MyReady:   0,
				MyBlocked: 2,
				MyChecks:  0,
				MyReview:  0,
			},
			contains: []string{"My PRs:", "2 blocked"},
		},
		{
			name: "all zero my-pr buckets returns empty",
			state: domain.PullRequestCount{
				MyReady:   0,
				MyBlocked: 0,
				MyChecks:  0,
				MyReview:  0,
			},
			isEmpty: true,
		},
		{
			name:    "unavailable returns empty",
			state:   domain.PullRequestUnavailable{},
			isEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := buildMyPullRequestSummary(tt.state)
			if tt.isEmpty {
				if got != "" {
					t.Errorf("buildMyPullRequestSummary(%q) = %q, want empty", tt.name, got)
				}
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildMyPullRequestSummary(%q) = %q, want it to contain %q", tt.name, got, want)
				}
			}
		})
	}
}

func TestBuildMyPullRequestSummary_BlockedIsPinned(t *testing.T) {
	t.Parallel()

	state := domain.PullRequestCount{MyBlocked: 2}

	got, pinned := buildMyPullRequestSummary(state)

	if !pinned {
		t.Fatalf("buildMyPullRequestSummary() pinned = false, want true")
	}

	if !strings.Contains(got, "2 blocked") {
		t.Fatalf("summary = %q, want blocked count", got)
	}
}

func TestBuildInfo_UsesRecentActivityOverStatus(t *testing.T) {
	t.Parallel()

	repo := makeTestRepository("demo")

	runtime := InfoRuntime{
		Phase:                0,
		Now:                  time.Now(),
		RecentActivityByRepo: map[string][]TimedInfoMessage{"/tmp/demo": {{Message: InfoMessage{Text: "Deleted 2 branches", Tone: InfoToneSuccess}, ExpiresAt: time.Now().Add(time.Minute)}}},
	}

	got := buildInfo(repo, InfoWidth, runtime)
	if !strings.Contains(got, "Deleted 2 branches") {
		t.Fatalf("buildInfo() = %q, want recent activity message", got)
	}
}

// ============================================================================
// buildInfo tests
// ============================================================================

func TestBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     domain.Repository
		contains []string
	}{
		{
			name: "idle with my pr summary shows summary in info",
			repo: newTestRepository("repo").
				PullRequests(domain.PullRequestCount{
					MyReady:   0,
					MyBlocked: 1,
					MyChecks:  2,
				}).
				Build(),
			contains: []string{"My PRs:", "1 blocked", "2 checks"},
		},
		{
			name: "idle with synced remote shows empty info",
			repo: makeTestRepository("repo"),
		},
		{
			name:     "idle with remote error shows error message",
			repo:     newTestRepository("repo").RemoteState(domain.RemoteError{Message: "connection timed out"}).Build(),
			contains: []string{"connection timed out"},
		},
		{
			name:     "idle with no upstream shows no upstream message",
			repo:     newTestRepository("repo").RemoteState(domain.NoUpstream{}).Build(),
			contains: []string{common.LabelNoUpstream},
		},
		{
			name:     "idle with detached remote shows detached message",
			repo:     newTestRepository("repo").RemoteState(domain.DetachedRemote{}).Build(),
			contains: []string{common.LabelDetached},
		},
		{
			name:     "idle with diverged remote shows diverged message",
			repo:     newTestRepository("repo").RemoteState(domain.Diverged{AheadCount: 1, BehindCount: 2}).Build(),
			contains: []string{common.StatusDiverged},
		},
		{
			name:     "idle with prunable branches shows prune count",
			repo:     newTestRepository("repo").MergedBranches("feature-a", "feature-b").Build(),
			contains: []string{"2 prunable branches"},
		},
		{
			name:     "idle with single prunable branch uses singular",
			repo:     newTestRepository("repo").MergedBranches("feature-a").Build(),
			contains: []string{"1 prunable branch"},
		},
		{
			name:     "idle with one stash uses singular stash label",
			repo:     newTestRepository("repo").StashCount(1).Build(),
			contains: []string{"1 stash"},
		},
		{
			name:     "idle with multiple stashes uses plural stash label",
			repo:     newTestRepository("repo").StashCount(3).Build(),
			contains: []string{"3 stashes"},
		},
		{
			name: "idle with stash and prunable branches prefers prunable info",
			repo: newTestRepository("repo").
				StashCount(2).
				MergedBranches("feature-a").
				Build(),
			contains: []string{"1 prunable branch"},
		},
		{
			name: "idle with remote error and stash prefers remote error",
			repo: newTestRepository("repo").
				RemoteState(domain.RemoteError{Message: "connection timed out"}).
				StashCount(4).
				Build(),
			contains: []string{"connection timed out"},
		},
		{
			name: "refreshing in progress shows empty info",
			repo: newTestRepository("repo").Activity(&domain.RefreshingActivity{Complete: false}).Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildInfo(tt.repo, InfoWidth, InfoRuntime{})
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildInfo(%q) = %q, want it to contain %q",
						tt.name, got, want)
				}
			}
		})
	}
}

// ============================================================================
// stylePullOutput tests
// ============================================================================

func TestStylePullOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lastLine string
		exitCode int
		contains string
	}{
		{
			name:     "error keyword in output",
			lastLine: "error: could not apply patch",
			exitCode: 1,
			contains: "error: could not apply patch",
		},
		{
			name:     "fatal keyword in output",
			lastLine: "fatal: remote origin not found",
			exitCode: 128,
			contains: "fatal: remote origin not found",
		},
		{
			name:     "success with up to date",
			lastLine: "Already up to date.",
			exitCode: 0,
			contains: "Already up to date.",
		},
		{
			name:     "success with up-to-date hyphenated",
			lastLine: "Already up-to-date.",
			exitCode: 0,
			contains: "Already up-to-date.",
		},
		{
			name:     "success with done message",
			lastLine: "done.",
			exitCode: 0,
			contains: "done.",
		},
		{
			name:     "success with file changed",
			lastLine: "1 file changed, 5 insertions(+)",
			exitCode: 0,
			contains: "1 file changed",
		},
		{
			name:     "non-zero exit unknown text falls through to warn",
			lastLine: "something unexpected happened",
			exitCode: 1,
			contains: "something unexpected happened",
		},
		{
			name:     "zero exit with unknown text falls through to warn",
			lastLine: "some other output",
			exitCode: 0,
			contains: "some other output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stylePullOutput(tt.lastLine, tt.exitCode, InfoWidth)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("stylePullOutput(%q, %d) = %q, want it to contain %q",
					tt.lastLine, tt.exitCode, got, tt.contains)
			}
		})
	}
}

// ============================================================================
// repositoryToRow tests
// ============================================================================

func TestRepositoryToRow(t *testing.T) {
	t.Parallel()

	repo := newTestRepository("test-repo").
		PullRequests(domain.PullRequestCount{
			Open:   2,
			MyOpen: 1,
		}).
		Build()

	row := repositoryToRow(repo, true, ColumnLayout{ProjectWidth: 30, BranchWidth: 20, InfoWidth: InfoWidth}, InfoRuntime{})

	if len(row) != 8 {
		t.Fatalf("repositoryToRow returned %d columns, want 8", len(row))
	}

	// Selector should have the icon since isSelected=true
	if !strings.Contains(row[0], common.IconSelector) {
		t.Errorf("row[0] (selector) = %q, want it to contain %q", row[0], common.IconSelector)
	}

	// Project name
	if !strings.Contains(row[1], "test-repo") {
		t.Errorf("row[1] (project) = %q, want it to contain %q", row[1], "test-repo")
	}

	// Branch
	if !strings.Contains(row[2], "main") {
		t.Errorf("row[2] (branch) = %q, want it to contain %q", row[2], "main")
	}

	// Local status (clean)
	if !strings.Contains(row[3], common.IconClean) {
		t.Errorf("row[3] (local) = %q, want it to contain %q", row[3], common.IconClean)
	}

	// Remote status (synced)
	if !strings.Contains(row[4], common.IconSynced) {
		t.Errorf("row[4] (remote) = %q, want it to contain %q", row[4], common.IconSynced)
	}

	if !strings.Contains(row[5], "2") || !strings.Contains(row[5], "(*)") {
		t.Errorf("row[5] (prs) = %q, want it to contain count and marker", row[5])
	}

	if strings.TrimSpace(row[6]) != "" {
		t.Errorf("row[6] (pr alert) = %q, want empty when no blocked/ready PRs", row[6])
	}
}

func TestRepositoryToRow_NotSelected(t *testing.T) {
	t.Parallel()

	repo := newTestRepository("another-repo").
		LocalState(domain.DirtyLocalState{Modified: 2}).
		RemoteState(domain.Behind{Count: 3}).
		PullRequests(domain.PullRequestCount{Open: 1, MyOpen: 0}).
		CurrentBranch(domain.OnBranch{Name: "develop"}).
		Build()

	row := repositoryToRow(repo, false, ColumnLayout{ProjectWidth: 30, BranchWidth: 20, InfoWidth: InfoWidth}, InfoRuntime{})

	// Selector should NOT have the icon
	if strings.Contains(row[0], common.IconSelector) {
		t.Errorf("row[0] (selector) = %q, should NOT contain %q when not selected",
			row[0], common.IconSelector)
	}

	// Local should show dirty indicators
	if !strings.Contains(row[3], "~2") {
		t.Errorf("row[3] (local) = %q, want it to contain %q", row[3], "~2")
	}

	// Remote should show behind count
	if !strings.Contains(row[4], "3") {
		t.Errorf("row[4] (remote) = %q, want it to contain %q", row[4], "3")
	}
}
