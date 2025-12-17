package scanning

import "fresh/internal/domain"

type repoFoundMsg domain.Repository
type scanCompleteMsg struct{}
type ScanFinishedMsg struct{ Repos []domain.Repository }
