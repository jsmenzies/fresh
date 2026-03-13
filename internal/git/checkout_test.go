package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveIntegrationBranch_PreferenceOrder(t *testing.T) {
	t.Run("prefers local develop over local dev", func(t *testing.T) {
		repo := setupRepoWithBranches(t, true, true)
		target, track, err := resolveIntegrationBranch(repo)
		if err != nil {
			t.Fatalf("resolveIntegrationBranch() error = %v", err)
		}
		if target != "develop" || track {
			t.Fatalf("got target=%q track=%v, want develop,false", target, track)
		}
	})

	t.Run("falls back to local dev when develop missing", func(t *testing.T) {
		repo := setupRepoWithBranches(t, false, true)
		target, track, err := resolveIntegrationBranch(repo)
		if err != nil {
			t.Fatalf("resolveIntegrationBranch() error = %v", err)
		}
		if target != "dev" || track {
			t.Fatalf("got target=%q track=%v, want dev,false", target, track)
		}
	})

	t.Run("uses remote develop when local branches missing", func(t *testing.T) {
		repo := setupRemoteOnlyBranchRepo(t, "develop")
		target, track, err := resolveIntegrationBranch(repo)
		if err != nil {
			t.Fatalf("resolveIntegrationBranch() error = %v", err)
		}
		if target != "develop" || !track {
			t.Fatalf("got target=%q track=%v, want develop,true", target, track)
		}
	})

	t.Run("errors when neither develop nor dev exists", func(t *testing.T) {
		repo := setupRepoWithBranches(t, false, false)
		_, _, err := resolveIntegrationBranch(repo)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCheckoutIntegration(t *testing.T) {
	t.Run("checks out remote-only develop and creates tracking branch", func(t *testing.T) {
		repo := setupRemoteOnlyBranchRepo(t, "develop")

		target, exitCode, err := CheckoutIntegration(repo, nil)
		if err != nil {
			t.Fatalf("CheckoutIntegration() error = %v", err)
		}
		if exitCode != 0 {
			t.Fatalf("exitCode = %d, want 0", exitCode)
		}
		if target != "develop" {
			t.Fatalf("target = %q, want develop", target)
		}

		branch := currentBranch(t, repo)
		if branch != "develop" {
			t.Fatalf("current branch = %q, want develop", branch)
		}
	})

	t.Run("returns failure and message when branch is missing", func(t *testing.T) {
		repo := setupRepoWithBranches(t, false, false)
		var lines []string

		target, exitCode, err := CheckoutIntegration(repo, func(line string) {
			lines = append(lines, line)
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if target != "" {
			t.Fatalf("target = %q, want empty", target)
		}
		if exitCode != 1 {
			t.Fatalf("exitCode = %d, want 1", exitCode)
		}
		if len(lines) == 0 || !strings.Contains(strings.ToLower(lines[0]), "integration branch") {
			t.Fatalf("expected integration-branch error message, got %v", lines)
		}
	})
}

func setupRepoWithBranches(t *testing.T, withDevelop, withDev bool) string {
	t.Helper()
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	runGit(t, root, "init", "--initial-branch=main", "repo")
	configUser(t, repo)

	writeFile(t, filepath.Join(repo, "README.md"), "init\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "init")

	if withDevelop {
		runGit(t, repo, "checkout", "-b", "develop")
		runGit(t, repo, "checkout", "main")
	}
	if withDev {
		runGit(t, repo, "checkout", "-b", "dev")
		runGit(t, repo, "checkout", "main")
	}

	return repo
}

func setupRemoteOnlyBranchRepo(t *testing.T, branchName string) string {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "origin.git")
	seed := filepath.Join(root, "seed")
	clone := filepath.Join(root, "clone")

	runGit(t, root, "init", "--initial-branch=main", "--bare", "origin.git")
	runGit(t, root, "clone", remote, "seed")
	configUser(t, seed)
	writeFile(t, filepath.Join(seed, "README.md"), "seed\n")
	runGit(t, seed, "add", "README.md")
	runGit(t, seed, "commit", "-m", "init")
	runGit(t, seed, "push", "-u", "origin", "main")
	runGit(t, seed, "checkout", "-b", branchName)
	writeFile(t, filepath.Join(seed, "BRANCH.txt"), branchName+"\n")
	runGit(t, seed, "add", "BRANCH.txt")
	runGit(t, seed, "commit", "-m", branchName)
	runGit(t, seed, "push", "-u", "origin", branchName)

	runGit(t, root, "clone", remote, "clone")
	configUser(t, clone)
	_, _ = runGitAllowError(t, clone, "branch", "-D", branchName)

	return clone
}

func currentBranch(t *testing.T, repo string) string {
	t.Helper()
	out := runGitOutput(t, repo, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(out)
}

func configUser(t *testing.T, repo string) {
	t.Helper()
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "commit.gpgsign", "false")
	runGit(t, repo, "config", "tag.gpgsign", "false")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func runGitAllowError(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}
