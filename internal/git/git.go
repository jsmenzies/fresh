package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"fresh/internal/config"
	"fresh/internal/domain"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func createCommand(timeout time.Duration, name string, args ...string) *exec.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = cancel
	return exec.CommandContext(ctx, name, args...)
}

func BuildRepository(path string, cfg *config.Config) domain.Repository {
	repoName := filepath.Base(path)
	localState := GetLocalState(path)
	remoteState := GetRemoteState(path)
	lastCommitTime := GetLastCommitTime(path)
	remoteURL := GetRemoteURL(path)
	branches := BuildBranches(path, cfg.ProtectedBranches)

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

func IsGitInstalled() bool {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "--version")
	err := cmd.Run()
	return err == nil
}

func IsRepository(path string) bool {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

func GetRemoteURL(repoPath string) string {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

func GetCurrentBranch(repoPath string) domain.Branch {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "rev-parse", "--abbrev-ref", "HEAD")
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

func GetLocalState(repoPath string) domain.LocalState {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "status", "--porcelain=v2")
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

func GetRemoteState(repoPath string) domain.RemoteState {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
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
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "log", "-1", "--format=%ct")
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

func Fetch(repoPath string) error {
	cmd := createCommand(config.DefaultConfig().Timeout.Fetch, "git", "fetch", "--quiet")
	cmd.Dir = repoPath
	_, err := cmd.CombinedOutput()
	return err
}

func RefreshRemoteStatusWithFetch(repo *domain.Repository) error {
	cmd := createCommand(config.DefaultConfig().Timeout.Fetch, "git", "fetch", "--quiet")
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

	repo.RemoteState = GetRemoteState(repo.Path)
	return nil
}

func Pull(repoPath string, lineCallback func(string)) int {
	cmd := createCommand(config.DefaultConfig().Timeout.Pull, "git", "pull", "--rebase", "--progress")
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

func BuildBranches(repoPath string, excludedBranches []string) domain.Branches {
	branches := domain.Branches{}

	branches.Current = GetCurrentBranch(repoPath)

	allBranches, err := ListLocalBranches(repoPath)
	if err != nil {
		// If we can't list branches, return with just current
		return branches
	}

	currentBranchName := ""
	if branch, ok := branches.Current.(domain.OnBranch); ok {
		currentBranchName = branch.Name
	}

	excludedMap := make(map[string]bool)
	for _, branch := range excludedBranches {
		excludedMap[branch] = true
	}
	excludedMap[currentBranchName] = true

	var candidates []string
	for _, branch := range allBranches {
		if !excludedMap[branch] {
			candidates = append(candidates, branch)
		}
	}

	branches.Merged = FilterMergedBranches(repoPath, candidates)
	return branches
}

func ListLocalBranches(repoPath string) ([]string, error) {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "branch", "--format=%(refname:short)")
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

func FilterMergedBranches(repoPath string, branches []string) []string {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "branch", "--merged", "HEAD", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	mergedSet := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		mergedSet[strings.TrimSpace(scanner.Text())] = true
	}

	var merged []string
	for _, branch := range branches {
		if mergedSet[branch] {
			merged = append(merged, branch)
		}
	}
	return merged
}

func DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (exitCode int, deletedCount int) {
	deletedCount = 0

	for _, branch := range branches {
		cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "branch", "-d", branch)
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
