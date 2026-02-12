package scanning

import "fresh/internal/domain"

type repoFoundMsg string
type scanCompleteMsg struct{}
type ScanFinishedMsg struct{ Repos []domain.Repository }
