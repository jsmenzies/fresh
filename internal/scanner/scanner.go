package scanner

import (
	"fresh/internal/domain"
	"fresh/internal/git"
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Scanner struct {
	scanDir string
	ch      chan domain.Repository
	wg      sync.WaitGroup
}

func New(scanDir string) *Scanner {
	return &Scanner{
		scanDir: scanDir,
		ch:      make(chan domain.Repository),
	}
}

func (s *Scanner) GetRepoChannel() chan domain.Repository {
	return s.ch
}

func (s *Scanner) Scan() {
	defer close(s.ch)

	numWorkers := runtime.NumCPU()
	paths := make(chan string)

	for i := 0; i < numWorkers; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for path := range paths {
				// Simulate work
				time.Sleep(100 * time.Millisecond)
				if git.IsRepository(path) {
					repo := ToGitRepo(path)
					s.ch <- repo
				}
			}
		}()
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(paths)
		err := filepath.WalkDir(s.scanDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}

			if d.Name() == ".git" {
				paths <- filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			// errors are ignored if the scan can not access certain directories,
		}
	}()
	s.wg.Wait()
}

func (s *Scanner) Wait() {
	s.wg.Wait()
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