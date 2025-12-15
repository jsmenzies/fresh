package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

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

func HasModifiedFiles(repoPath string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
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

func GetCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

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

	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
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
