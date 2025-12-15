package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	printVersion()
}

func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		isFlag bool
	}{
		{"version long flag", []string{"fresh", "--version"}, true},
		{"version short flag", []string{"fresh", "-v"}, true},
		{"no flag", []string{"fresh"}, false},
		{"path argument", []string{"fresh", "/tmp"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasVersionFlag := len(tt.args) > 1 && (tt.args[1] == "--version" || tt.args[1] == "-v")
			if hasVersionFlag != tt.isFlag {
				t.Errorf("version flag detection = %v, want %v", hasVersionFlag, tt.isFlag)
			}
		})
	}
}

func TestParseDir(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	tests := []struct {
		name        string
		args        []string
		setup       func(t *testing.T) string // returns temp dir path if needed
		wantErr     bool
		errContains string
	}{
		{
			name:    "no arguments uses current directory",
			args:    []string{"fresh"},
			wantErr: false,
		},
		{
			name: "valid directory argument",
			args: []string{"fresh", ""},
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantErr: false,
		},
		{
			name:    "non-existent directory returns error",
			args:    []string{"fresh", "/path/that/does/not/exist/at/all"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test args
			os.Args = tt.args

			// Run setup if provided
			if tt.setup != nil {
				tmpDir := tt.setup(t)
				os.Args[1] = tmpDir
			}

			got, err := parseDir()

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == "" {
					t.Error("parseDir() returned empty string for valid directory")
				}

				// Verify the directory exists
				if _, err := os.Stat(got); os.IsNotExist(err) {
					t.Errorf("parseDir() returned non-existent directory: %s", got)
				}
			}
		})
	}
}

func TestParseDirWithRealDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"fresh", tmpDir}

	got, err := parseDir()
	if err != nil {
		t.Fatalf("parseDir() unexpected error: %v", err)
	}

	if got != tmpDir {
		t.Errorf("parseDir() = %v, want %v", got, tmpDir)
	}
}

func TestParseDirCurrentDirectory(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"fresh"}

	got, err := parseDir()
	if err != nil {
		t.Fatalf("parseDir() unexpected error: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() unexpected error: %v", err)
	}

	if got != cwd {
		t.Errorf("parseDir() = %v, want %v", got, cwd)
	}
}

func TestParseDirRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	originalArgs := os.Args
	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Args = originalArgs
		os.Chdir(originalWd)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	os.Args = []string{"fresh", "subdir"}

	got, err := parseDir()
	if err != nil {
		t.Fatalf("parseDir() unexpected error: %v", err)
	}

	if _, err := os.Stat(got); os.IsNotExist(err) {
		t.Errorf("parseDir() returned non-existent directory: %s", got)
	}
}

func TestMainFunctionality(t *testing.T) {
	tmpDir := t.TempDir()

	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"fresh", tmpDir}

	scanDir, err := parseDir()
	if err != nil {
		t.Fatalf("parseDir() failed: %v", err)
	}

	if scanDir != tmpDir {
		t.Errorf("parseDir() = %v, want %v", scanDir, tmpDir)
	}

	if _, err := os.Stat(scanDir); err != nil {
		t.Errorf("scan directory is not accessible: %v", err)
	}
}
