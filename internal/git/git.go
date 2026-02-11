package git

import (
	"bufio"
	"bytes"
	"fmt"
	"fresh/internal/domain"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var ProtectedBranches = []string{"main", "master", "develop", "dev", "production", "staging", "release"}

func BuildRepository(path string) domain.Repository {
	repoName := filepath.Base(path)
	localState := HasModifiedFiles(path)
	remoteState := GetStatus(path)
	lastCommitTime := GetLastCommitTime(path)
	remoteURL := GetRemoteURL(path)
	branches := BuildBranches(path, ProtectedBranches)

	return domain.Repository{
		Name:           repoName,
		Path:           path,
		Branches:       branches,
		LocalState:     localState,
		LastCommitTime: lastCommitTime,
		RemoteURL:      remoteURL,
		RemoteState:    remoteState,
	}
}

func RefreshRepositoryState(repo *domain.Repository) {
	repo.Branches = BuildBranches(repo.Path, ProtectedBranches)
	repo.LocalState = HasModifiedFiles(repo.Path)
	repo.RemoteState = GetStatus(repo.Path)
	repo.LastCommitTime = GetLastCommitTime(repo.Path)
	repo.RemoteURL = GetRemoteURL(repo.Path)
}

func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

func IsRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

func GetRemoteURL(repoPath string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

func GetCurrentBranch(repoPath string) domain.Branch {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	branch, err := cmd.Output()
	if err != nil {
		return domain.NoBranch{Reason: err.Error()}
	}

	name := strings.TrimSpace(string(branch))
	switch name {
	case "HEAD":
		return domain.DetachedHead{}
	case "":
		return domain.NoBranch{Reason: "no branch"}
	default:
		return domain.OnBranch{Name: name}
	}
}

func HasModifiedFiles(repoPath string) domain.LocalState {
	cmd := exec.Command("git", "status", "--porcelain=v2")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err != nil {
		return domain.LocalStateError{Message: err.Error()}
	}
	result := strings.TrimSpace(string(output))
	if result == "" {
		return domain.CleanLocalState{}
	}

	var added, modified, deleted, untracked int

	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '?':
			untracked++
		case '1', '2':
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			xy := parts[1]

			if strings.Contains(xy, "A") {
				added++
			} else if strings.Contains(xy, "D") {
				deleted++
			} else if strings.Contains(xy, "M") || strings.Contains(xy, "R") {
				modified++
			}
		case 'u':
			modified++
		}
	}

	return domain.DirtyLocalState{
		Added:     added,
		Modified:  modified,
		Deleted:   deleted,
		Untracked: untracked,
	}
}

func GetStatus(repoPath string) domain.RemoteState {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err != nil {
		errStr := string(err.(*exec.ExitError).Stderr)

		if strings.Contains(errStr, "no upstream") {
			return domain.NoUpstream{}
		}
		if strings.Contains(errStr, "does not point to a branch") {
			return domain.DetachedRemote{}
		}
		if strings.Contains(errStr, "bad revision") {
			return domain.NoUpstream{}
		}
		if strings.Contains(errStr, "no such branch:") {
			return domain.DetachedRemote{}
		}
		return domain.RemoteError{Message: errStr}
	}

	var ahead, behind int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%d\t%d", &ahead, &behind); err != nil {
		return domain.RemoteError{Message: "failed to parse git status output"}
	}

	if ahead > 0 && behind > 0 {
		return domain.Diverged{AheadCount: ahead, BehindCount: behind}
	}

	if ahead > 0 {
		return domain.Ahead{Count: ahead}
	}

	if behind > 0 {
		return domain.Behind{Count: behind}
	}

	return domain.Synced{}
}

func GetLastCommitTime(repoPath string) time.Time {
	cmd := exec.Command("git", "log", "-1", "--format=%ct")
	cmd.Dir = repoPath
	output, err := cmd.Output()
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

func RefreshRemoteStatusWithFetch(repo *domain.Repository) error {
	cmd := exec.Command("git", "fetch", "--quiet")
	cmd.Dir = repo.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		repo.RemoteState = domain.RemoteError{Message: errMsg}
		return fmt.Errorf("fetch failed: %s", errMsg)
	}

	repo.RemoteState = GetStatus(repo.Path)
	return nil
}

func Pull(repoPath string, lineCallback func(string)) int {
	cmd := exec.Command("git", "pull", "--rebase", "--progress")
	cmd.Dir = repoPath

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to get stderr pipe: %v", err))
		}
		return 1
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to get stdout pipe: %v", err))
		}
		return 1
	}

	if err := cmd.Start(); err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to start command: %v", err))
		}
		return 1
	}

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Split(splitOnCROrLF)

		for scanner.Scan() {
			lineStr := strings.TrimSpace(scanner.Text())
			if lineStr != "" && lineCallback != nil {
				lineCallback(lineStr)
			}
		}
	}()

	stdoutScanner := bufio.NewScanner(stdoutPipe)
	for stdoutScanner.Scan() {
		lineStr := strings.TrimSpace(stdoutScanner.Text())
		if lineStr != "" && lineCallback != nil {
			lineCallback(lineStr)
		}
	}

	cmdErr := cmd.Wait()

	<-stderrDone

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}

func splitOnCROrLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[0:i], nil
			}
			return i + 1, data[0:i], nil
		}
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// Branch operations - composable functions

// BuildBranches composes all branch operations into a single Branches object.
// Each operation can fail independently - failures don't stop other operations.
func BuildBranches(repoPath string, excludedBranches []string) domain.Branches {
	branches := domain.Branches{}

	// Get current branch
	branches.Current = GetCurrentBranch(repoPath)

	// Get list of all local branches once
	allBranches, err := ListLocalBranches(repoPath)
	if err != nil {
		// If we can't list branches, return with just current
		return branches
	}

	// Get current branch name for exclusion
	currentBranchName := ""
	if branch, ok := branches.Current.(domain.OnBranch); ok {
		currentBranchName = branch.Name
	}

	// Build excluded map
	excludedMap := make(map[string]bool)
	for _, branch := range excludedBranches {
		excludedMap[branch] = true
	}
	excludedMap[currentBranchName] = true

	// Filter candidates (excluding current and protected)
	var candidates []string
	for _, branch := range allBranches {
		if !excludedMap[branch] {
			candidates = append(candidates, branch)
		}
	}

	// Get merged branches from the pre-fetched list
	branches.Merged = FilterMergedBranches(repoPath, candidates)

	// Get squashed branches (those not merged but candidates for cleanup)
	branches.Squashed = FilterSquashedBranches(candidates, branches.Merged)

	return branches
}

// ListLocalBranches returns all local branch names
func ListLocalBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, scanner.Err()
}

// IsBranchFullyMerged checks if a branch is fully merged into HEAD using merge-base
func IsBranchFullyMerged(repoPath string, branchName string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", branchName, "HEAD")
	cmd.Dir = repoPath
	err := cmd.Run()
	return err == nil
}

// FilterMergedBranches filters a list of branches to return only fully merged ones
func FilterMergedBranches(repoPath string, branches []string) []string {
	var merged []string
	for _, branch := range branches {
		if IsBranchFullyMerged(repoPath, branch) {
			merged = append(merged, branch)
		}
	}
	return merged
}

// FilterSquashedBranches returns branches that aren't in the merged list
// These are candidates for squashed merge cleanup
func FilterSquashedBranches(branches []string, mergedBranches []string) []string {
	// Build set of merged branches for quick lookup
	mergedMap := make(map[string]bool)
	for _, b := range mergedBranches {
		mergedMap[b] = true
	}

	// Squashed branches are those not in the merged list
	var squashed []string
	for _, branch := range branches {
		if !mergedMap[branch] {
			squashed = append(squashed, branch)
		}
	}

	return squashed
}

// DeleteBranches deletes branches with line-by-line progress reporting
func DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (exitCode int, deletedCount int) {
	deletedCount = 0

	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-d", branch)
		cmd.Dir = repoPath
		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))

		if err != nil {
			// Branch not fully merged (e.g., squashed) - skip it
			if lineCallback != nil {
				lineCallback(fmt.Sprintf("Skipped: %s (not merged)", branch))
			}
			continue
		}

		deletedCount++
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Deleted: %s", branch))
		}
		_ = outputStr
	}

	return 0, deletedCount
}

// DeleteSquashedBranches deletes squashed branches using force delete
func DeleteSquashedBranches(repoPath string, branches []string, lineCallback func(string)) (exitCode int, deletedCount int) {
	deletedCount = 0

	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-D", branch)
		cmd.Dir = repoPath
		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))

		if err != nil {
			if lineCallback != nil {
				lineCallback(fmt.Sprintf("Failed: %s (%s)", branch, outputStr))
			}
			continue
		}

		deletedCount++
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Deleted: %s", branch))
		}
		_ = outputStr
	}

	return 0, deletedCount
}
