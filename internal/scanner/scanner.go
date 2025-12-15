package scanner

import (
	"fresh/internal/domain"
	"fresh/internal/git"
	"os"
	"path/filepath"
)

type Scanner struct {
	scanDir      string
	repositories []domain.Repository
	ch           chan domain.Repository
	finished     bool
}

func New(scanDir string) *Scanner {
	return &Scanner{
		scanDir:      scanDir,
		repositories: make([]domain.Repository, 0),
		ch:           make(chan domain.Repository),
		finished:     false,
	}
}

func (s *Scanner) IsFinished() bool {
	return s.finished
}

func (s *Scanner) GetRepositories() []domain.Repository {
	return s.repositories
}

func (s *Scanner) GetRepoCount() int {
	return len(s.repositories)
}

func (s *Scanner) GetRepoChannel() chan domain.Repository {
	return s.ch
}

func (s *Scanner) Scan() {
	entries, err := os.ReadDir(s.scanDir)
	if err != nil {
		s.finished = true
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(s.scanDir, entry.Name())
			if git.IsRepository(fullPath) {
				repo := ToGitRepo(fullPath)
				s.repositories = append(s.repositories, repo)
				s.ch <- repo
			}
		}
	}

	close(s.ch)
	s.finished = true
}

func ToGitRepo(path string) domain.Repository {
	repoName := filepath.Base(path)
	lastCommitTime := git.GetLastCommitTime(path)
	remoteURL := git.GetRemoteURL(path)
	aheadCount, behindCount := git.GetStatus(path)
	hasModified := git.HasModifiedFiles(path)
	currentBranch := git.GetCurrentBranch(path)

	return domain.Repository{
		Name:           repoName,
		Path:           path,
		LastCommitTime: lastCommitTime,
		RemoteURL:      remoteURL,
		HasModified:    hasModified,
		AheadCount:     aheadCount,
		BehindCount:    behindCount,
		CurrentBranch:  currentBranch,
	}
}
