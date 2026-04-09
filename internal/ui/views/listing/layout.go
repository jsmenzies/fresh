package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
)

type ColumnLayout struct {
	ProjectWidth int
	BranchWidth  int
	InfoWidth    int
}

// Column width constraints for dynamic columns
const (
	MaxProjectWidth = 50
	MaxBranchWidth  = 30
	MinProjectWidth = 22
	MinBranchWidth  = 8
)

// Fixed column widths
const (
	SelectorWidth  = 2
	LocalWidth     = 15
	RemoteWidth    = 11
	PRAlertWidth   = 2
	PRWidth        = 8
	InfoWidth      = 42
	MinInfoWidth   = 1
	InterColumnGap = 2 // spacing between columns
)

// Legend layout
const (
	LegendColWidth = 20
)

// totalFixedWidth returns the sum of all fixed-width columns plus inter-column gaps.
func totalFixedWidthWithoutInfo() int {
	return SelectorWidth + LocalWidth + RemoteWidth + PRAlertWidth + PRWidth + (6 * InterColumnGap)
}

func calculateColumnLayout(repositories []domain.Repository, terminalWidth int) ColumnLayout {
	projectWidth, branchWidth := calculateColumnWidths(repositories, terminalWidth)
	infoWidth := calculateInfoWidth(terminalWidth, projectWidth, branchWidth)

	return ColumnLayout{
		ProjectWidth: projectWidth,
		BranchWidth:  branchWidth,
		InfoWidth:    infoWidth,
	}
}

func calculateInfoWidth(terminalWidth, projectWidth, branchWidth int) int {
	if terminalWidth <= 0 {
		return InfoWidth
	}

	width := terminalWidth - totalFixedWidthWithoutInfo() - projectWidth - branchWidth
	if width < MinInfoWidth {
		return MinInfoWidth
	}

	return width
}

// calculateColumnWidths determines the project and branch column widths
// based on repository content and available terminal width.
func calculateColumnWidths(repositories []domain.Repository, terminalWidth int) (projectWidth, branchWidth int) {
	if len(repositories) == 0 {
		return MinProjectWidth, MinBranchWidth
	}

	maxProjectLen := 0
	maxBranchLen := 0

	for _, repo := range repositories {
		if len(repo.Name) > maxProjectLen {
			maxProjectLen = len(repo.Name)
		}

		switch branch := repo.Branches.Current.(type) {
		case domain.OnBranch:
			if len(branch.Name) > maxBranchLen {
				maxBranchLen = len(branch.Name)
			}
		case domain.DetachedHead:
			if len("HEAD") > maxBranchLen {
				maxBranchLen = len("HEAD")
			}
		}
	}

	// Clamp to min/max immediately
	projectWidth = common.Clamp(maxProjectLen, MinProjectWidth, MaxProjectWidth)
	branchWidth = common.Clamp(maxBranchLen, MinBranchWidth, MaxBranchWidth)

	// If no terminal width provided, use content-based widths
	if terminalWidth <= 0 {
		return projectWidth, branchWidth
	}

	// Check if we need to shrink to fit terminal
	availableWidth := terminalWidth - totalFixedWidthWithoutInfo() - MinInfoWidth
	if availableWidth <= 0 {
		// Terminal too narrow: use minimums and let truncation handle it
		return MinProjectWidth, MinBranchWidth
	}

	totalDesired := projectWidth + branchWidth
	if totalDesired <= availableWidth {
		// We fit! Use the content-based widths
		return projectWidth, branchWidth
	}

	// Need to shrink proportionally
	return distributeWidth(projectWidth, branchWidth, availableWidth)
}

// distributeWidth proportionally allocates availableWidth between project and branch columns.
func distributeWidth(projectDesired, branchDesired, available int) (project, branch int) {
	if available <= 0 {
		return MinProjectWidth, MinBranchWidth
	}

	totalDesired := projectDesired + branchDesired
	if totalDesired == 0 {
		return MinProjectWidth, MinBranchWidth
	}

	// Calculate proportional distribution
	projectRatio := float64(projectDesired) / float64(totalDesired)
	project = int(float64(available) * projectRatio)
	branch = available - project

	// Enforce minimums (this may cause total to exceed available)
	if project < MinProjectWidth {
		project = MinProjectWidth
		branch = available - project
	}
	if branch < MinBranchWidth {
		branch = MinBranchWidth
		project = available - branch
	}

	// Final clamp to maxes
	project = common.Clamp(project, MinProjectWidth, MaxProjectWidth)
	branch = common.Clamp(branch, MinBranchWidth, MaxBranchWidth)

	return project, branch
}
