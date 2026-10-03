package scanner

import (
	"fresh/internal/git"
	"os"
	"path/filepath"
)

type Scanner struct {
	scanDir string
	ch      chan string
}

func New(scanDir string) *Scanner {
	return &Scanner{
		scanDir: scanDir,
		ch:      make(chan string),
	}
}

func (s *Scanner) GetRepoChannel() <-chan string {
	return s.ch
}

func (s *Scanner) Scan() {
	defer close(s.ch)

	entries, err := os.ReadDir(s.scanDir)
	if err != nil {
		return
	}

	paths := []string{s.scanDir}
	for _, entry := range entries {
		if entry.IsDir() {
			paths = append(paths, filepath.Join(s.scanDir, entry.Name()))
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil && git.IsRepository(path) {
			s.ch <- path
		}
	}
}
