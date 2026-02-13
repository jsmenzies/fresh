package git

import (
	"fmt"
	"fresh/internal/domain"
	"strings"
)

func ParseStatus(output []byte) domain.LocalState {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return domain.CleanLocalState{}
	}

	var added, modified, deleted, untracked int
	lines := parseLines([]byte(result))

	for _, line := range lines {
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

func ParseRemoteState(output []byte) domain.RemoteState {
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

func ParseRemoteError(errStr string) domain.RemoteState {
	if strings.TrimSpace(errStr) == "" {
		return domain.RemoteError{Message: "unknown error"}
	}

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

func ParseCurrentBranch(output []byte) domain.Branch {
	name := strings.TrimSpace(string(output))
	switch name {
	case "HEAD":
		return domain.DetachedHead{}
	case "":
		return domain.NoBranch{Reason: "no branch"}
	default:
		return domain.OnBranch{Name: name}
	}
}

// ParseBranchList parses branch command output into branch names.
func ParseBranchList(output []byte) []string {
	return parseLines(output)
}
