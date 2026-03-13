package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"fresh/internal/config"
)

func CheckoutIntegration(repoPath string, lineCallback func(string)) (targetBranch string, exitCode int, err error) {
	targetBranch, trackRemote, err := resolveIntegrationBranch(repoPath)
	if err != nil {
		if lineCallback != nil {
			lineCallback(err.Error())
		}
		return "", 1, err
	}

	args := []string{"checkout", targetBranch}
	if trackRemote {
		args = []string{"checkout", "--track", "origin/" + targetBranch}
	}

	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", args...)
	cmd.Dir = repoPath

	exitCode, runErr := runCommandStreamingOutput(cmd, lineCallback)
	return targetBranch, exitCode, runErr
}

func CheckoutPrimary(repoPath string, lineCallback func(string)) (targetBranch string, exitCode int, err error) {
	targetBranch, trackRemote, err := resolvePrimaryBranch(repoPath)
	if err != nil {
		if lineCallback != nil {
			lineCallback(err.Error())
		}
		return "", 1, err
	}

	args := []string{"checkout", targetBranch}
	if trackRemote {
		args = []string{"checkout", "--track", "origin/" + targetBranch}
	}

	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", args...)
	cmd.Dir = repoPath

	exitCode, runErr := runCommandStreamingOutput(cmd, lineCallback)
	return targetBranch, exitCode, runErr
}

func resolveIntegrationBranch(repoPath string) (target string, trackRemote bool, err error) {
	if branchExists(repoPath, "refs/heads/develop") {
		return "develop", false, nil
	}
	if branchExists(repoPath, "refs/heads/dev") {
		return "dev", false, nil
	}
	if branchExists(repoPath, "refs/remotes/origin/develop") {
		return "develop", true, nil
	}
	if branchExists(repoPath, "refs/remotes/origin/dev") {
		return "dev", true, nil
	}

	return "", false, fmt.Errorf("No integration branch found: expected develop or dev (local or origin)")
}

func resolvePrimaryBranch(repoPath string) (target string, trackRemote bool, err error) {
	if branchExists(repoPath, "refs/heads/main") {
		return "main", false, nil
	}
	if branchExists(repoPath, "refs/heads/master") {
		return "master", false, nil
	}
	if branchExists(repoPath, "refs/remotes/origin/main") {
		return "main", true, nil
	}
	if branchExists(repoPath, "refs/remotes/origin/master") {
		return "master", true, nil
	}

	return "", false, fmt.Errorf("No primary branch found: expected main or master (local or origin)")
}

func branchExists(repoPath, ref string) bool {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "git", "show-ref", "--verify", "--quiet", ref)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}

func runCommandStreamingOutput(cmd *exec.Cmd, lineCallback func(string)) (int, error) {
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to get stderr pipe: %v", err))
		}
		return 1, err
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to get stdout pipe: %v", err))
		}
		return 1, err
	}

	if err := cmd.Start(); err != nil {
		if lineCallback != nil {
			lineCallback(fmt.Sprintf("Failed to start command: %v", err))
		}
		return 1, err
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
			return exitErr.ExitCode(), cmdErr
		}
		return 1, cmdErr
	}

	return 0, nil
}
