package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"testing"
)

func TestClamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		value, min, max int
		want            int
	}{
		{"below min", 1, 5, 10, 5},
		{"above max", 15, 5, 10, 10},
		{"in range", 7, 5, 10, 7},
		{"at min boundary", 5, 5, 10, 5},
		{"at max boundary", 10, 5, 10, 10},
		{"min equals max", 5, 5, 5, 5},
		{"zero value", 0, 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := common.Clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("common.Clamp(%d, %d, %d) = %d, want %d",
					tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestDistributeWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                          string
		projectDesired, branchDesired int
		available                     int
		wantProject, wantBranch       int
	}{
		{
			name:           "proportional split fits exactly",
			projectDesired: 40, branchDesired: 20,
			available:   60,
			wantProject: 40, wantBranch: 20,
		},
		{
			name:           "proportional shrink",
			projectDesired: 40, branchDesired: 20,
			available:   45,
			wantProject: 30, wantBranch: 15,
		},
		{
			name:           "enforce minimums when very tight",
			projectDesired: 40, branchDesired: 20,
			available:   25,
			wantProject: MinProjectWidth, wantBranch: MinBranchWidth,
		},
		{
			name:           "zero available returns minimums",
			projectDesired: 40, branchDesired: 20,
			available:   0,
			wantProject: MinProjectWidth, wantBranch: MinBranchWidth,
		},
		{
			name:           "negative available returns minimums",
			projectDesired: 40, branchDesired: 20,
			available:   -10,
			wantProject: MinProjectWidth, wantBranch: MinBranchWidth,
		},
		{
			name:           "zero desired returns minimums",
			projectDesired: 0, branchDesired: 0,
			available:   100,
			wantProject: MinProjectWidth, wantBranch: MinBranchWidth,
		},
		{
			name:           "equal desired split equally",
			projectDesired: 30, branchDesired: 30,
			available:   60,
			wantProject: 30, wantBranch: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotProject, gotBranch := distributeWidth(tt.projectDesired, tt.branchDesired, tt.available)
			if gotProject != tt.wantProject || gotBranch != tt.wantBranch {
				t.Errorf("distributeWidth(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.projectDesired, tt.branchDesired, tt.available,
					gotProject, gotBranch, tt.wantProject, tt.wantBranch)
			}
		})
	}
}

func TestDistributeWidth_RespectsMinimums(t *testing.T) {
	t.Parallel()

	// When available space is very tight, minimums should be enforced
	project, branch := distributeWidth(50, 30, 30)

	if project < MinProjectWidth {
		t.Errorf("project width %d below minimum %d", project, MinProjectWidth)
	}
	if branch < MinBranchWidth {
		t.Errorf("branch width %d below minimum %d", branch, MinBranchWidth)
	}
}

func TestDistributeWidth_RespectsMaximums(t *testing.T) {
	t.Parallel()

	// Even with excess space, columns should not exceed max
	project, branch := distributeWidth(100, 100, 200)

	if project > MaxProjectWidth {
		t.Errorf("project width %d above maximum %d", project, MaxProjectWidth)
	}
	if branch > MaxBranchWidth {
		t.Errorf("branch width %d above maximum %d", branch, MaxBranchWidth)
	}
}

func TestCalculateColumnWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		repos         []domain.Repository
		terminalWidth int
		wantProject   int
		wantBranch    int
	}{
		{
			name:          "empty repos returns minimums",
			repos:         []domain.Repository{},
			terminalWidth: 200,
			wantProject:   MinProjectWidth,
			wantBranch:    MinBranchWidth,
		},
		{
			name:          "short names clamped to minimums",
			repos:         []domain.Repository{makeTestRepository("app")},
			terminalWidth: 200,
			wantProject:   MinProjectWidth,
			wantBranch:    MinBranchWidth,
		},
		{
			name: "long project name clamped to max",
			repos: []domain.Repository{
				makeTestRepository("this-is-a-very-long-repository-name-that-exceeds-the-maximum"),
			},
			terminalWidth: 200,
			wantProject:   MaxProjectWidth,
			wantBranch:    MinBranchWidth,
		},
		{
			name: "long branch name clamped to max",
			repos: []domain.Repository{
				newTestRepository("app").
					CurrentBranch(domain.OnBranch{Name: "feature/very-long-branch-name-that-is-too-wide"}).
					Build(),
			},
			terminalWidth: 200,
			wantProject:   MinProjectWidth,
			wantBranch:    MaxBranchWidth,
		},
		{
			name:          "zero terminal width uses content-based widths",
			repos:         []domain.Repository{newTestRepository("my-project").CurrentBranch(domain.OnBranch{Name: "develop"}).Build()},
			terminalWidth: 0,
			wantProject:   MinProjectWidth,
			wantBranch:    MinBranchWidth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotProject, gotBranch := calculateColumnWidths(tt.repos, tt.terminalWidth)
			if gotProject != tt.wantProject || gotBranch != tt.wantBranch {
				t.Errorf("calculateColumnWidths(..., %d) = (%d, %d), want (%d, %d)",
					tt.terminalWidth, gotProject, gotBranch, tt.wantProject, tt.wantBranch)
			}
		})
	}
}

func TestCalculateColumnWidths_NarrowTerminal(t *testing.T) {
	t.Parallel()

	repo := makeTestRepository("my-project")

	// Very narrow terminal — available should go negative, returning minimums
	project, branch := calculateColumnWidths([]domain.Repository{repo}, 10)
	if project != MinProjectWidth {
		t.Errorf("narrow terminal: project = %d, want %d", project, MinProjectWidth)
	}
	if branch != MinBranchWidth {
		t.Errorf("narrow terminal: branch = %d, want %d", branch, MinBranchWidth)
	}
}

func TestCalculateColumnWidths_DetachedHead(t *testing.T) {
	t.Parallel()

	repo := newTestRepository("my-project").CurrentBranch(domain.DetachedHead{CommitSHA: "abc123"}).Build()

	// DetachedHead should use "HEAD" (4 chars) for branch width calculation
	project, branch := calculateColumnWidths([]domain.Repository{repo}, 200)
	if project < MinProjectWidth {
		t.Errorf("detached head: project width %d below minimum %d", project, MinProjectWidth)
	}
	if branch < MinBranchWidth {
		t.Errorf("detached head: branch width %d below minimum %d", branch, MinBranchWidth)
	}
}

func TestCalculateColumnWidths_MultipleRepos(t *testing.T) {
	t.Parallel()

	repos := []domain.Repository{
		makeTestRepository("short"),
		newTestRepository("this-is-a-medium-length-name").CurrentBranch(domain.OnBranch{Name: "feature/long-branch"}).Build(),
	}

	project, branch := calculateColumnWidths(repos, 200)

	// Project should accommodate the longest name (28 chars), within bounds
	if project < MinProjectWidth {
		t.Errorf("multi-repo: project width %d below minimum", project)
	}
	if project > MaxProjectWidth {
		t.Errorf("multi-repo: project width %d above maximum", project)
	}

	// Branch should accommodate "feature/long-branch" (19 chars)
	if branch < MinBranchWidth {
		t.Errorf("multi-repo: branch width %d below minimum", branch)
	}
	if branch > MaxBranchWidth {
		t.Errorf("multi-repo: branch width %d above maximum", branch)
	}
}
