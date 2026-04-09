package git

import (
	"encoding/json"
	"errors"
	"fresh/internal/config"
	"fresh/internal/domain"
	"sort"
	"strings"
	"time"
)

var ErrPullRequestDetailsUnsupported = errors.New("pull requests are currently only supported for GitHub repositories")

type ghRepoPullRequestRow struct {
	Number           int    `json:"number"`
	Title            string `json:"title"`
	UpdatedAt        string `json:"updatedAt"`
	StatusCheckRollup []ghStatusCheckContext `json:"statusCheckRollup"`
	Author           struct {
		Login string `json:"login"`
	} `json:"author"`
}

type ghStatusCheckContext struct {
	Conclusion string `json:"conclusion"`
	Status     string `json:"status"`
	State      string `json:"state"`
}

func GetRepositoryPullRequests(repo domain.Repository) ([]domain.PullRequestDetails, error) {
	owner, name, ok := parseGitHubRemote(repo.RemoteURL)
	if !ok {
		return nil, ErrPullRequestDetailsUnsupported
	}

	ownerRepo := owner + "/" + name
	args := []string{
		"pr", "list",
		"--repo", ownerRepo,
		"--state", "open",
		"--limit", "200",
		"--json", "number,title,updatedAt,author,statusCheckRollup",
	}

	cmd := createCommand(config.DefaultConfig().Timeout.Default, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, normalizeGhError(err)
	}

	var rows []ghRepoPullRequestRow
	if err := json.Unmarshal(output, &rows); err != nil {
		return nil, err
	}

	currentUser := queryGitHubLogin()
	pullRequests := make([]domain.PullRequestDetails, 0, len(rows))
	for _, row := range rows {
		updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
		pullRequests = append(pullRequests, domain.PullRequestDetails{
			Number:    row.Number,
			Title:     row.Title,
			IsMine:    currentUser != "" && strings.EqualFold(strings.TrimSpace(row.Author.Login), currentUser),
			UpdatedAt: updatedAt,
			Checks:    summarizePullRequestChecks(row.StatusCheckRollup),
		})
	}

	sort.Slice(pullRequests, func(i, j int) bool {
		if pullRequests[i].IsMine != pullRequests[j].IsMine {
			return pullRequests[i].IsMine
		}
		if !pullRequests[i].UpdatedAt.Equal(pullRequests[j].UpdatedAt) {
			return pullRequests[i].UpdatedAt.After(pullRequests[j].UpdatedAt)
		}
		return pullRequests[i].Number > pullRequests[j].Number
	})

	return pullRequests, nil
}

func queryGitHubLogin() string {
	cmd := createCommand(config.DefaultConfig().Timeout.Default, "gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func summarizePullRequestChecks(contexts []ghStatusCheckContext) domain.PullRequestChecks {
	summary := domain.PullRequestChecks{Total: len(contexts)}
	for _, context := range contexts {
		switch classifyCheckContext(context) {
		case "passed":
			summary.Passed++
		case "running":
			summary.Running++
		case "failed":
			summary.Failed++
		case "skipped":
			summary.Skipped++
		default:
			summary.Waiting++
		}
	}
	return summary
}

func classifyCheckContext(context ghStatusCheckContext) string {
	conclusion := strings.ToUpper(strings.TrimSpace(context.Conclusion))
	switch conclusion {
	case "SUCCESS":
		return "passed"
	case "SKIPPED", "NEUTRAL":
		return "skipped"
	case "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE", "STALE":
		return "failed"
	}

	status := strings.ToUpper(strings.TrimSpace(context.Status))
	if status != "" && status != "COMPLETED" {
		if status == "IN_PROGRESS" {
			return "running"
		}
		return "waiting"
	}

	state := strings.ToUpper(strings.TrimSpace(context.State))
	switch state {
	case "SUCCESS":
		return "passed"
	case "FAILURE", "ERROR":
		return "failed"
	case "PENDING", "EXPECTED":
		return "waiting"
	}

	if status == "COMPLETED" {
		return "passed"
	}

	return "waiting"
}
