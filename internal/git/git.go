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

func BuildRepository(path string) domain.Repository {
	repoName := filepath.Base(path)
	branch := GetCurrentBranch(path)
	localState := HasModifiedFiles(path)
	remoteState := GetStatus(path)
	lastCommitTime := GetLastCommitTime(path)
	remoteURL := GetRemoteURL(path)

	return domain.Repository{
		Name:           repoName,
		Path:           path,
		Branch:         branch,
		LocalState:     localState,
		LastCommitTime: lastCommitTime,
		RemoteURL:      remoteURL,
		RemoteState:    remoteState,
	}
}

func RefreshRepositoryState(repo *domain.Repository) {
	repo.Branch = GetCurrentBranch(repo.Path)
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
