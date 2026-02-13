package git

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"fresh/internal/config"
	"fresh/internal/domain"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Git provides git operations for a repository.
type Git struct {
	timeout time.Duration
}

// New creates a new Git client with the specified configuration.
func New(cfg *config.Config) *Git {
	timeout := 30 * time.Second
	if cfg != nil && cfg.Timeout.Default > 0 {
		timeout = cfg.Timeout.Default
	}
	return &Git{timeout: timeout}
}

// Status returns the local state of the repository.
func (g *Git) Status(repoPath string) domain.LocalState {
	output, err := g.run(repoPath, []string{"status", "--porcelain=v2"})

	if err != nil {
		return domain.LocalStateError{Message: err.Error()}
	}

	return ParseStatus(output)
}

// RemoteState returns the remote sync state (ahead/behind/diverged).
func (g *Git) RemoteState(repoPath string) domain.RemoteState {
	output, err := g.run(repoPath, []string{"rev-list", "--left-right", "--count", "HEAD...@{u}"})

	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return domain.RemoteError{Message: err.Error()}
		}

		errStr := string(exitErr.Stderr)
		if strings.TrimSpace(errStr) == "" {
			errStr = exitErr.Error()
		}

		return ParseRemoteError(errStr)
	}

	return ParseRemoteState(output)
}

// CurrentBranch returns the current branch or detached HEAD state.
func (g *Git) CurrentBranch(repoPath string) domain.Branch {
	output, err := g.run(repoPath, []string{"rev-parse", "--abbrev-ref", "HEAD"})

	if err != nil {
		return domain.NoBranch{Reason: err.Error()}
	}

	return ParseCurrentBranch(output)
}

// Branches returns all branch information for a repository.
func (g *Git) Branches(repoPath string, protectedBranches []string) domain.Branches {
	branches := domain.Branches{}
	branches.Current = g.CurrentBranch(repoPath)

	allBranches, err := g.ListBranches(repoPath)
	if err != nil {
		return branches
	}

	currentBranchName := ""
	if branch, ok := branches.Current.(domain.OnBranch); ok {
		currentBranchName = branch.Name
	}

	excludedMap := make(map[string]bool)
	for _, branch := range protectedBranches {
		excludedMap[branch] = true
	}
	excludedMap[currentBranchName] = true

	var candidates []string
	for _, branch := range allBranches {
		if !excludedMap[branch] {
			candidates = append(candidates, branch)
		}
	}

	branches.Merged = g.FilterMerged(repoPath, candidates)
	return branches
}

// ListBranches returns all local branch names.
func (g *Git) ListBranches(repoPath string) ([]string, error) {
	output, err := g.run(repoPath, []string{"branch", "--format=%(refname:short)"})
	if err != nil {
		return nil, err
	}
	return ParseBranchList(output), nil
}

// FilterMerged returns which of the given branches are merged into HEAD.
func (g *Git) FilterMerged(repoPath string, branches []string) []string {
	output, err := g.run(repoPath, []string{"branch", "--merged", "HEAD", "--format=%(refname:short)"})
	if err != nil {
		return nil
	}

	mergedSet := make(map[string]bool)
	for _, branch := range ParseBranchList(output) {
		mergedSet[branch] = true
	}

	var merged []string
	for _, branch := range branches {
		if mergedSet[branch] {
			merged = append(merged, branch)
		}
	}
	return merged
}

// Fetch fetches from the remote.
func (g *Git) Fetch(repoPath string) error {
	_, err := g.run(repoPath, []string{"fetch", "--quiet"})
	return err
}

// Pull pulls from the remote with progress callback.
func (g *Git) Pull(repoPath string, progressHandler func(string)) int {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, "git", []string{"pull", "--rebase", "--progress"}...)
	execCmd.Dir = repoPath

	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return 1
	}
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return 1
	}

	if err := execCmd.Start(); err != nil {
		return 1
	}

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderr)
		scanner.Split(splitOnCROrLF)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && progressHandler != nil {
				progressHandler(line)
			}
		}
	}()

	stdoutScanner := bufio.NewScanner(stdout)
	for stdoutScanner.Scan() {
		line := strings.TrimSpace(stdoutScanner.Text())
		if line != "" && progressHandler != nil {
			progressHandler(line)
		}
	}

	cmdErr := execCmd.Wait()
	<-stderrDone

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

// DeleteBranches deletes the specified branches with progress callback.
func (g *Git) DeleteBranches(repoPath string, branches []string, progressHandler func(string)) (exitCode int, deletedCount int) {
	deletedCount = 0

	for _, branch := range branches {
		output, err := g.runCombined(repoPath, []string{"branch", "-d", branch})
		outputStr := strings.TrimSpace(string(output))

		if err != nil {
			if progressHandler != nil {
				progressHandler(fmt.Sprintf("Failed: %s (%s)", branch, outputStr))
			}
			continue
		}

		deletedCount++
		if progressHandler != nil {
			progressHandler(fmt.Sprintf("Deleted: %s", branch))
		}
	}

	return 0, deletedCount
}

// RemoteURL returns the origin remote URL.
func (g *Git) RemoteURL(repoPath string) string {
	output, err := g.run(repoPath, []string{"remote", "get-url", "origin"})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// LastCommitTime returns the timestamp of the last commit.
func (g *Git) LastCommitTime(repoPath string) time.Time {
	output, err := g.run(repoPath, []string{"log", "-1", "--format=%ct"})
	if err != nil {
		return time.Time{}
	}

	timestampStr := strings.TrimSpace(string(output))
	if timestampStr == "" {
		return time.Time{}
	}

	timestamp := int64(0)
	if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

// BuildRepository builds a complete Repository object by running operations concurrently.
func (g *Git) BuildRepository(path string, protectedBranches []string) domain.Repository {
	repoName := filepath.Base(path)

	type result struct {
		localState     domain.LocalState
		remoteState    domain.RemoteState
		lastCommitTime time.Time
		remoteURL      string
		branches       domain.Branches
	}

	var res result
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(5)

	go func() {
		defer wg.Done()
		state := g.Status(path)
		mu.Lock()
		res.localState = state
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		state := g.RemoteState(path)
		mu.Lock()
		res.remoteState = state
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		t := g.LastCommitTime(path)
		mu.Lock()
		res.lastCommitTime = t
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		url := g.RemoteURL(path)
		mu.Lock()
		res.remoteURL = url
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		b := g.Branches(path, protectedBranches)
		mu.Lock()
		res.branches = b
		mu.Unlock()
	}()

	wg.Wait()

	return domain.Repository{
		Name:           repoName,
		Path:           path,
		Branches:       res.branches,
		LocalState:     res.localState,
		LastCommitTime: res.lastCommitTime,
		RemoteURL:      res.remoteURL,
		RemoteState:    res.remoteState,
	}
}

// run executes a git command and returns stdout.
func (g *Git) run(repoPath string, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath
	return cmd.Output()
}

// runCombined executes a git command and returns combined output.
func (g *Git) runCombined(repoPath string, args []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath
	return cmd.CombinedOutput()
}
