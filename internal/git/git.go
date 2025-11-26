package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func IsRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// GetRemoteURL returns the remote URL for the origin remote
func GetRemoteURL(repoPath string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// GetStatus returns the ahead and behind counts compared to upstream
func GetStatus(repoPath string) (aheadCount int, behindCount int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	statusStr := strings.TrimSpace(string(output))
	if statusStr == "" {
		return 0, 0
	}

	var ahead, behind int
	if _, err := fmt.Sscanf(statusStr, "%d\t%d", &ahead, &behind); err != nil {
		return 0, 0
	}

	return ahead, behind
}

// HasModifiedFiles checks if the repository has modified files
func HasModifiedFiles(repoPath string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}

// GetLastCommitTime returns the timestamp of the last commit
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

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// RefreshRemoteStatus fetches from remote and returns updated ahead/behind counts
func RefreshRemoteStatus(repoPath string) (aheadCount, behindCount int, err error) {
	cmd := exec.Command("git", "fetch", "--quiet")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return 0, 0, fmt.Errorf("fetch failed: %s", errMsg)
	}

	// Compare local vs remote to check for updates
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		// If upstream is not configured, not an error, just no updates
		errMsg := strings.TrimSpace(string(output))
		if strings.Contains(errMsg, "no upstream") || strings.Contains(errMsg, "@{u}") {
			return 0, 0, nil
		}
		if errMsg == "" {
			errMsg = err.Error()
		}
		return 0, 0, fmt.Errorf("%s", errMsg)
	}

	statusStr := strings.TrimSpace(string(output))
	if statusStr == "" {
		return 0, 0, nil
	}

	var ahead, behind int
	if _, err := fmt.Sscanf(statusStr, "%d\t%d", &ahead, &behind); err != nil {
		return 0, 0, fmt.Errorf("parse status failed: %s (output: %s)", err.Error(), statusStr)
	}

	return ahead, behind, nil
}

// Pull performs a git pull with rebase and calls lineCallback for each output line
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

	// Read stderr in real-time and stream lines
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

	// Read stdout and stream lines
	stdoutScanner := bufio.NewScanner(stdoutPipe)
	for stdoutScanner.Scan() {
		lineStr := strings.TrimSpace(stdoutScanner.Text())
		if lineStr != "" && lineCallback != nil {
			lineCallback(lineStr)
		}
	}

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Wait for stderr goroutine to finish
	<-stderrDone

	// Return exit code
	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}

// splitOnCROrLF splits on both \r and \n for git progress output
func splitOnCROrLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Look for \r or \n
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		// Found a delimiter
		if data[i] == '\r' {
			// Check for \r\n (Windows line ending)
			if i+1 < len(data) && data[i+1] == '\n' {
				return i + 2, data[0:i], nil
			}
			// Just \r
			return i + 1, data[0:i], nil
		}
		// Just \n
		return i + 1, data[0:i], nil
	}

	// No delimiter found
	if atEOF {
		return len(data), data, nil
	}

	// Request more data
	return 0, nil, nil
}
