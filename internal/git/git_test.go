package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetStashCount(t *testing.T) {
	t.Parallel()

	t.Run("returns zero for repo without stashes", func(t *testing.T) {
		t.Parallel()
		repo := setupRepoWithCommit(t)

		got := GetStashCount(repo)
		if got != 0 {
			t.Fatalf("GetStashCount() = %d, want 0", got)
		}
	})

	t.Run("returns stash count when stashes exist", func(t *testing.T) {
		t.Parallel()
		repo := setupRepoWithCommit(t)

		mustWriteFile(t, filepath.Join(repo, "tracked.txt"), "base\nupdated\n")
		runGit(t, repo, "add", "tracked.txt")
		runGit(t, repo, "stash", "push", "-m", "first")

		mustWriteFile(t, filepath.Join(repo, "tracked.txt"), "base\nupdated\nagain\n")
		runGit(t, repo, "add", "tracked.txt")
		runGit(t, repo, "stash", "push", "-m", "second")

		got := GetStashCount(repo)
		if got != 2 {
			t.Fatalf("GetStashCount() = %d, want 2", got)
		}
	})

	t.Run("returns zero for non-repository path", func(t *testing.T) {
		t.Parallel()
		got := GetStashCount(t.TempDir())
		if got != 0 {
			t.Fatalf("GetStashCount() = %d, want 0", got)
		}
	})
}

func setupRepoWithCommit(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")

	mustWriteFile(t, filepath.Join(repo, "tracked.txt"), "base\n")
	runGit(t, repo, "add", "tracked.txt")
	runGit(t, repo, "commit", "-m", "initial commit")

	return repo
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", path, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
