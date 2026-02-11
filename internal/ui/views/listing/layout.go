package listing

import "fresh/internal/domain"

// Column width constraints for dynamic columns
const (
	MaxProjectWidth = 50
	MaxBranchWidth  = 30
	MinProjectWidth = 22
	MinBranchWidth  = 8
)

// Fixed column widths
const (
	SelectorWidth   = 2
	LocalWidth      = 15
	RemoteWidth     = 11
	InfoWidth       = 42
	LastCommitWidth = 20
	LinksWidth      = 8
	InterColumnGap  = 2 // spacing between columns
)

// Legend layout
const (
	LegendColWidth = 20
)

// totalFixedWidth returns the sum of all fixed-width columns plus inter-column gaps.
func totalFixedWidth() int {
	return SelectorWidth + LocalWidth + RemoteWidth + InfoWidth +
		LastCommitWidth + LinksWidth + (6 * InterColumnGap)
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
	projectWidth = clamp(maxProjectLen, MinProjectWidth, MaxProjectWidth)
	branchWidth = clamp(maxBranchLen, MinBranchWidth, MaxBranchWidth)

	// If no terminal width provided, use content-based widths
	if terminalWidth <= 0 {
		return projectWidth, branchWidth
	}

	// Check if we need to shrink to fit terminal
	availableWidth := terminalWidth - totalFixedWidth()
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

// clamp restricts value to [min, max] range.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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
	project = clamp(project, MinProjectWidth, MaxProjectWidth)
	branch = clamp(branch, MinBranchWidth, MaxBranchWidth)

	return project, branch
}
