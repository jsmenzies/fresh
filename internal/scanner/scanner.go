package scanner

import (
	"fresh/internal/git"
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"
)

type Scanner struct {
	scanDir string
	ch      chan string
	wg      sync.WaitGroup
}

func New(scanDir string) *Scanner {
	return &Scanner{
		scanDir: scanDir,
		ch:      make(chan string),
	}
}

func (s *Scanner) GetRepoChannel() chan string {
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
				if git.IsRepository(path) {
					s.ch <- path
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
