package git

import (
	"bufio"
	"bytes"
	"fmt"
	"fresh/internal/domain"
	"fresh/internal/textutil"
	"os/exec"
	"strings"
	"sync"
)

func Pull(repoPath string, lineCallback func(string)) domain.CommandOutcome {
	cmd := createCommand(defaultConfig.Timeout.Pull, "git", "pull", "--rebase", "--progress")
	cmd.Dir = repoPath

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return pullSetupFailure(lineCallback, "Failed to get stderr pipe", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return pullSetupFailure(lineCallback, "Failed to get stdout pipe", err)
	}

	if err := cmd.Start(); err != nil {
		return pullSetupFailure(lineCallback, "Failed to start command", err)
	}

	output := &pullOutputTracker{}
	recordLine := func(line string) { output.RecordAndEmit(line, lineCallback) }

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Split(splitOnCROrLF)
		scanLines(scanner, recordLine)
	}()

	stdoutScanner := bufio.NewScanner(stdoutPipe)
	scanLines(stdoutScanner, recordLine)

	cmdErr := cmd.Wait()
	<-stderrDone

	if cmdErr != nil {
		return domain.CommandOutcome{
			ExitCode:      commandExitCode(cmdErr),
			FailureReason: output.FailureReason(cmdErr),
		}
	}

	return domain.CommandOutcome{}
}

type pullOutputTracker struct {
	mu             sync.Mutex
	lastLine       string
	firstErrorLine string
}

func (t *pullOutputTracker) RecordAndEmit(raw string, lineCallback func(string)) {
	line := strings.TrimSpace(raw)
	if line == "" {
		return
	}

	t.mu.Lock()
	t.lastLine = line
	lower := strings.ToLower(line)
	if t.firstErrorLine == "" && (strings.Contains(lower, "error") || strings.Contains(lower, "fatal")) {
		t.firstErrorLine = line
	}
	t.mu.Unlock()

	if lineCallback != nil {
		lineCallback(line)
	}
}

func (t *pullOutputTracker) FailureReason(cmdErr error) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return textutil.FirstNonEmptyTrimmed(t.firstErrorLine, t.lastLine, cmdErr.Error())
}

func pullSetupFailure(lineCallback func(string), context string, err error) domain.CommandOutcome {
	msg := fmt.Sprintf("%s: %v", context, err)
	if lineCallback != nil {
		lineCallback(msg)
	}
	return domain.CommandOutcome{ExitCode: 1, FailureReason: msg}
}

func commandExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return 1
}

func scanLines(scanner *bufio.Scanner, onLine func(string)) {
	for scanner.Scan() {
		onLine(scanner.Text())
	}
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
