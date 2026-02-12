package main

import (
	"testing"

	"fresh/internal/cli"
)

func TestParseFlagsWithDirFlag(t *testing.T) {
	tmpDir := t.TempDir()

	action, cfg, err := cli.ParseFlags([]string{"--dir", tmpDir})
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}

	if action != cli.ActionRun {
		t.Errorf("ParseFlags() action = %v, want ActionRun", action)
	}

	if cfg.ScanDir != tmpDir {
		t.Errorf("ParseFlags() ScanDir = %v, want %v", cfg.ScanDir, tmpDir)
	}
}
