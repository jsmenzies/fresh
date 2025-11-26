package scanner

import (
	"os"
	"path/filepath"

	"fresh/internal/domain"
	"fresh/internal/git"
)

type Scanner struct {
	foundRepositories []domain.Repository
	directoriesToScan []string
	currentIndex      int
}

func New() *Scanner {
	return &Scanner{
		foundRepositories: make([]domain.Repository, 0),
		directoriesToScan: make([]string, 0),
		currentIndex:      0,
	}
}

func (s *Scanner) StartScanning(scanDir string) {
	s.directoriesToScan = make([]string, 0)
	s.foundRepositories = make([]domain.Repository, 0)
	s.currentIndex = 0

	s.scanDirectoriesWithDepth(scanDir, 0, 0)
}

func (s *Scanner) ScanStep() (domain.Repository, bool, bool) {
	if s.currentIndex >= len(s.directoriesToScan) {
		return domain.Repository{}, false, false
	}

	path := s.directoriesToScan[s.currentIndex]
	s.currentIndex++

	if git.IsRepository(path) {
		repoName := filepath.Base(path)
		lastCommitTime := git.GetLastCommitTime(path)
		remoteURL := git.GetRemoteURL(path)
		aheadCount, behindCount := git.GetStatus(path)
		hasModified := git.HasModifiedFiles(path)
		currentBranch := git.GetCurrentBranch(path)

		repo := domain.Repository{
			Name:           repoName,
			Path:           path,
			LastCommitTime: lastCommitTime,
			RemoteURL:      remoteURL,
			HasModified:    hasModified,
			AheadCount:     aheadCount,
			BehindCount:    behindCount,
			CurrentBranch:  currentBranch,
		}

		s.foundRepositories = append(s.foundRepositories, repo)
		hasMore := s.currentIndex < len(s.directoriesToScan)
		return repo, hasMore, true
	}

	hasMore := s.currentIndex < len(s.directoriesToScan)
	return domain.Repository{}, hasMore, false
}

// GetFoundRepositories returns all repositories found so far
func (s *Scanner) GetFoundRepositories() []domain.Repository {
	return s.foundRepositories
}

// GetRepoCount returns the number of repositories found
func (s *Scanner) GetRepoCount() int {
	return len(s.foundRepositories)
}

// scanDirectoriesWithDepth recursively collects directories to scan up to maxDepth
func (s *Scanner) scanDirectoriesWithDepth(dir string, currentDepth, maxDepth int) {
	if currentDepth > maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(dir, entry.Name())
			// Only add directories at the first level (depth 0)
			if currentDepth == 0 {
				s.directoriesToScan = append(s.directoriesToScan, fullPath)
			}
		}
	}
}
