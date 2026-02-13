package git

import (
	"context"
	"os/exec"
	"time"
)

func IsGitInstalled() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "--version")
	err := cmd.Run()
	return err == nil
}
