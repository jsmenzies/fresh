package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseCLIConfigWithDirFlag(t *testing.T) {
	tmpDir := t.TempDir()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"fresh", "--dir", tmpDir}

	action, cfg, err := parseCliFlags()
	if err != nil {
		t.Fatalf("parseCliFlags() unexpected error: %v", err)
	}

	if action != ActionRun {
		t.Errorf("parseCliFlags() action = %v, want ActionRun", action)
	}

	if cfg.ScanDir != tmpDir {
		t.Errorf("parseCliFlags() ScanDir = %v, want %v", cfg.ScanDir, tmpDir)
	}
}
