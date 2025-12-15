package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsGitInstalled(t *testing.T) {
	result := IsGitInstalled()

	// In CI/CD, git should be installed
	// This test will fail if git is not in PATH
	if !result {
		t.Error("IsGitInstalled() = false, expected git to be installed")
	}

	// Verify git command is actually available
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		t.Errorf("git --version failed: %v", err)
	}
}

func TestIsRepository(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string // returns path to test
		want  bool
	}{
		{
			name: "valid git repository",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				// Initialize a git repository
				cmd := exec.Command("git", "init")
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to init git repo: %v", err)
				}

				return tmpDir
			},
			want: true,
		},
		{
			name: "non-git directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: false,
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return "/path/that/does/not/exist"
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got := IsRepository(path)

			if got != tt.want {
				t.Errorf("IsRepository(%q) = %v, want %v", path, got, tt.want)
			}
		})
	}
}

func TestGetCurrentBranch(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	initRepo(t, tmpDir)

	// Create initial commit (required to have a branch)
	createDummyCommit(t, tmpDir)

	got := GetCurrentBranch(tmpDir)

	// Default branch is usually "main" or "master"
	if got != "main" && got != "master" {
		t.Errorf("GetCurrentBranch() = %v, want 'main' or 'master'", got)
	}
}

func TestGetCurrentBranchNonRepo(t *testing.T) {
	tmpDir := t.TempDir()

	got := GetCurrentBranch(tmpDir)

	if got != "" {
		t.Errorf("GetCurrentBranch() on non-repo = %v, want empty string", got)
	}
}

func TestHasModifiedFiles(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  bool
	}{
		{
			name: "clean repository",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				initRepo(t, tmpDir)
				createDummyCommit(t, tmpDir)
				return tmpDir
			},
			want: false,
		},
		{
			name: "modified files",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				initRepo(t, tmpDir)
				createDummyCommit(t, tmpDir)

				// Create a new file
				testFile := filepath.Join(tmpDir, "modified.txt")
				if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				return tmpDir
			},
			want: true,
		},
		{
			name: "non-repository",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got := HasModifiedFiles(path)

			if got != tt.want {
				t.Errorf("HasModifiedFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLastCommitTime(t *testing.T) {
	tmpDir := t.TempDir()
	initRepo(t, tmpDir)
	createDummyCommit(t, tmpDir)

	got := GetLastCommitTime(tmpDir)

	if got.IsZero() {
		t.Error("GetLastCommitTime() returned zero time for repo with commit")
	}
}

func TestGetLastCommitTimeNoCommits(t *testing.T) {
	tmpDir := t.TempDir()
	initRepo(t, tmpDir)

	got := GetLastCommitTime(tmpDir)

	if !got.IsZero() {
		t.Errorf("GetLastCommitTime() = %v, want zero time for repo without commits", got)
	}
}

func TestGetRemoteURL(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  string
	}{
		{
			name: "repository with remote",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				initRepo(t, tmpDir)

				// Add a remote
				cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
				cmd.Dir = tmpDir
				if err := cmd.Run(); err != nil {
					t.Fatalf("failed to add remote: %v", err)
				}

				return tmpDir
			},
			want: "https://github.com/test/repo.git",
		},
		{
			name: "repository without remote",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				initRepo(t, tmpDir)
				return tmpDir
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got := GetRemoteURL(path)

			if got != tt.want {
				t.Errorf("GetRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	// This test verifies GetStatus doesn't panic on various scenarios
	// Detailed status testing requires more complex setup with remotes

	tests := []struct {
		name  string
		setup func(t *testing.T) string
	}{
		{
			name: "repository without upstream",
			setup: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()
				initRepo(t, tmpDir)
				createDummyCommit(t, tmpDir)
				return tmpDir
			},
		},
		{
			name: "non-repository",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			// Should not panic
			ahead, behind := GetStatus(path)

			// Without upstream, should return 0, 0
			if ahead != 0 || behind != 0 {
				t.Logf("GetStatus() = (%d, %d) for %s", ahead, behind, tt.name)
			}
		})
	}
}

// Helper functions

func initRepo(t *testing.T, path string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	configureGitUser(t, path)
}

func configureGitUser(t *testing.T, path string) {
	t.Helper()

	cmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to configure git: %v", err)
		}
	}
}

func createDummyCommit(t *testing.T, path string) {
	t.Helper()

	// Create a file
	testFile := filepath.Join(path, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Add and commit
	cmds := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create commit: %v", err)
		}
	}
}
