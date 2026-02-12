package scanner

import (
	"io/fs"
	"path/filepath"
)

type Scanner struct {
	scanDir string
	ch      chan string
}

type RepositoryScanner interface {
	Scan()
	GetRepoChannel() chan string
}

var _ RepositoryScanner = (*Scanner)(nil)

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

	err := filepath.WalkDir(s.scanDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			s.ch <- filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		// errors are ignored if the scan can not access certain directories,
	}
}
