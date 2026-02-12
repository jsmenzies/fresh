package listing

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}

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
			name: "synced shows synced icon",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Synced{},
			},
			contains: []string{common.IconSynced},
		},
		{
			name: "ahead shows ahead icon with count",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Ahead{Count: 3},
			},
			contains: []string{common.IconAhead, "3"},
		},
		{
			name: "behind shows behind icon with count",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Behind{Count: 5},
			},
			contains: []string{common.IconBehind, "5"},
		},
		{
			name: "diverged shows both ahead and behind counts",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Diverged{AheadCount: 2, BehindCount: 7},
			},
			contains: []string{common.IconAhead, "2", common.IconBehind, "7"},
		},
		{
			name: "no upstream shows error icon",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.NoUpstream{},
			},
			contains: []string{common.IconRemoteError},
		},
		{
			name: "detached remote shows error icon",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.DetachedRemote{},
			},
			contains: []string{common.IconRemoteError},
		},
		{
			name: "remote error shows error icon",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.RemoteError{Message: "timeout"},
			},
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
			name: "idle with synced remote shows empty info",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
		},
		{
			name: "idle with remote error shows error message",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.RemoteError{Message: "connection timed out"},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"connection timed out"},
		},
		{
			name: "idle with no upstream shows no upstream message",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.NoUpstream{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{common.LabelNoUpstream},
		},
		{
			name: "idle with detached remote shows detached message",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.DetachedRemote{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{common.LabelDetached},
		},
		{
			name: "idle with diverged remote shows diverged message",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Diverged{AheadCount: 1, BehindCount: 2},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{common.StatusDiverged},
		},
		{
			name: "idle with prunable branches shows prune count",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Synced{},
				Branches: domain.Branches{
					Current: domain.OnBranch{Name: "main"},
					Merged:  []string{"feature-a", "feature-b"},
				},
			},
			contains: []string{"2 prunable branches"},
		},
		{
			name: "idle with single prunable branch uses singular",
			repo: domain.Repository{
				Activity:    domain.IdleActivity{},
				RemoteState: domain.Synced{},
				Branches: domain.Branches{
					Current: domain.OnBranch{Name: "main"},
					Merged:  []string{"feature-a"},
				},
			},
			contains: []string{"1 prunable branch"},
		},
		{
			name: "completed pull with exit 0 and up to date",
			repo: domain.Repository{
				Activity: &domain.PullingActivity{
					LineBuffer: domain.LineBuffer{Lines: []string{"Already up to date."}},
					Complete:   true,
					ExitCode:   0,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"Already up to date."},
		},
		{
			name: "completed pull with error keyword",
			repo: domain.Repository{
				Activity: &domain.PullingActivity{
					LineBuffer: domain.LineBuffer{Lines: []string{"fatal: remote error"}},
					Complete:   true,
					ExitCode:   1,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"fatal: remote error"},
		},
		{
			name: "completed pull with file changes",
			repo: domain.Repository{
				Activity: &domain.PullingActivity{
					LineBuffer: domain.LineBuffer{Lines: []string{"3 files changed, 10 insertions(+)"}},
					Complete:   true,
					ExitCode:   0,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"3 file"},
		},
		{
			name: "completed prune with deleted branches",
			repo: domain.Repository{
				Activity: &domain.PruningActivity{
					DeletedCount: 3,
					Complete:     true,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"Deleted 3 branches"},
		},
		{
			name: "completed prune with no branches to prune",
			repo: domain.Repository{
				Activity: &domain.PruningActivity{
					DeletedCount: 0,
					Complete:     true,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"No branches to prune"},
		},
		{
			name: "completed prune with error in lines",
			repo: domain.Repository{
				Activity: &domain.PruningActivity{
					LineBuffer:   domain.LineBuffer{Lines: []string{"Failed: branch is not merged"}},
					DeletedCount: 0,
					Complete:     true,
				},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			contains: []string{"branch is not merged"},
		},
		{
			name: "refreshing in progress shows empty info",
			repo: domain.Repository{
				Activity:    &domain.RefreshingActivity{Complete: false},
				RemoteState: domain.Synced{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildInfo(tt.repo)
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
			got := stylePullOutput(tt.lastLine, tt.exitCode)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("stylePullOutput(%q, %d) = %q, want it to contain %q",
					tt.lastLine, tt.exitCode, got, tt.contains)
			}
		})
	}
}

// ============================================================================
// buildLastUpdate tests
// ============================================================================

func TestBuildLastUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     domain.Repository
		contains string
		isEmpty  bool
	}{
		{
			name:    "zero time returns empty",
			repo:    domain.Repository{},
			isEmpty: true,
		},
		{
			name: "recent time shows time ago",
			repo: domain.Repository{
				LastCommitTime: time.Now().Add(-5 * time.Minute),
			},
			contains: common.IconClock,
		},
		{
			name: "old time shows time ago with clock icon",
			repo: domain.Repository{
				LastCommitTime: time.Now().Add(-48 * time.Hour),
			},
			contains: common.IconClock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildLastUpdate(tt.repo)
			if tt.isEmpty {
				if got != "" {
					t.Errorf("buildLastUpdate(%q) = %q, want empty", tt.name, got)
				}
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Errorf("buildLastUpdate(%q) = %q, want it to contain %q",
					tt.name, got, tt.contains)
			}
		})
	}
}

// ============================================================================
// buildLinks tests
// ============================================================================

func TestBuildLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		branch   domain.Branch
		contains []string
		isEmpty  bool
	}{
		{
			name:    "empty URL returns empty",
			url:     "",
			branch:  domain.OnBranch{Name: "main"},
			isEmpty: true,
		},
		{
			name:    "non-GitHub URL returns empty",
			url:     "git@gitlab.com:user/repo.git",
			branch:  domain.OnBranch{Name: "main"},
			isEmpty: true,
		},
		{
			name:   "GitHub SSH URL renders link icons",
			url:    "git@github.com:octocat/hello-world.git",
			branch: domain.OnBranch{Name: "main"},
			contains: []string{
				common.IconCode,
				common.IconPullRequests,
				common.IconOpenPR,
			},
		},
		{
			name:   "GitHub HTTPS URL renders link icons",
			url:    "https://github.com/octocat/hello-world.git",
			branch: domain.OnBranch{Name: "feature-x"},
			contains: []string{
				common.IconCode,
				common.IconPullRequests,
				common.IconOpenPR,
			},
		},
		{
			name:   "detached head still renders links",
			url:    "git@github.com:octocat/hello-world.git",
			branch: domain.DetachedHead{CommitSHA: "abc123"},
			contains: []string{
				common.IconCode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildLinks(tt.url, tt.branch)
			if tt.isEmpty {
				if got != "" {
					t.Errorf("buildLinks(%q, %v) = %q, want empty", tt.url, tt.branch, got)
				}
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildLinks(%q, %v) = %q, want it to contain %q",
						tt.url, tt.branch, got, want)
				}
			}
		})
	}
}

// ============================================================================
// MakeClickableURL tests
// ============================================================================

func TestMakeClickableURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		displayText string
		contains    string
	}{
		{
			name:        "empty URL returns display text only",
			url:         "",
			displayText: "click me",
			contains:    "click me",
		},
		{
			name:        "valid URL wraps in OSC8 escape sequence",
			url:         "https://github.com/foo/bar",
			displayText: "code",
			contains:    "https://github.com/foo/bar",
		},
		{
			name:        "display text is present in output",
			url:         "https://github.com/foo/bar",
			displayText: "code",
			contains:    "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MakeClickableURL(tt.url, tt.displayText)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("MakeClickableURL(%q, %q) = %q, want it to contain %q",
					tt.url, tt.displayText, got, tt.contains)
			}
		})
	}
}

func TestMakeClickableURL_OSC8Format(t *testing.T) {
	t.Parallel()

	got := MakeClickableURL("https://example.com", "link")
	expected := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", "https://example.com", "link")
	if got != expected {
		t.Errorf("MakeClickableURL OSC8 format: got %q, want %q", got, expected)
	}
}

// ============================================================================
// repositoryToRow tests
// ============================================================================

func TestRepositoryToRow(t *testing.T) {
	t.Parallel()

	repo := domain.Repository{
		Name:        "test-repo",
		Path:        "/tmp/test-repo",
		Activity:    domain.IdleActivity{},
		LocalState:  domain.CleanLocalState{},
		RemoteState: domain.Synced{},
		Branches: domain.Branches{
			Current: domain.OnBranch{Name: "main"},
		},
		LastCommitTime: time.Now().Add(-10 * time.Minute),
	}

	row := repositoryToRow(repo, true, 30, 20)

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

	// Last commit should have clock icon
	if !strings.Contains(row[6], common.IconClock) {
		t.Errorf("row[6] (last commit) = %q, want it to contain %q", row[6], common.IconClock)
	}
}

func TestRepositoryToRow_NotSelected(t *testing.T) {
	t.Parallel()

	repo := domain.Repository{
		Name:        "another-repo",
		Path:        "/tmp/another-repo",
		Activity:    domain.IdleActivity{},
		LocalState:  domain.DirtyLocalState{Modified: 2},
		RemoteState: domain.Behind{Count: 3},
		Branches: domain.Branches{
			Current: domain.OnBranch{Name: "develop"},
		},
	}

	row := repositoryToRow(repo, false, 30, 20)

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
