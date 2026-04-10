package git

import (
	"encoding/json"
	"errors"
	"fresh/internal/domain"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrPullRequestDetailsUnsupported = errors.New("pull requests are currently only supported for GitHub repositories")

const githubLoginCacheTTL = time.Hour

var cachedGitHubLogin githubLoginCache

type githubLoginCache struct {
	mu          sync.Mutex
	login       string
	expiresAt   time.Time
	initialized bool
}

type ghRepoPullRequestRow struct {
	Number            int                    `json:"number"`
	Title             string                 `json:"title"`
	UpdatedAt         string                 `json:"updatedAt"`
	StatusCheckRollup []ghStatusCheckContext `json:"statusCheckRollup"`
	Author            struct {
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

	cmd := createCommand(defaultConfig.Timeout.Default, "gh", args...)
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
	return cachedGitHubLogin.get(time.Now(), githubLoginCacheTTL, func() (string, error) {
		cmd := createCommand(defaultConfig.Timeout.Default, "gh", "api", "user", "--jq", ".login")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	})
}

func (c *githubLoginCache) get(now time.Time, ttl time.Duration, loader func() (string, error)) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized && now.Before(c.expiresAt) {
		return c.login
	}

	login, err := loader()
	if err != nil {
		login = ""
	}
	c.login = strings.TrimSpace(login)
	c.expiresAt = now.Add(ttl)
	c.initialized = true
	return c.login
}

func summarizePullRequestChecks(contexts []ghStatusCheckContext) domain.PullRequestChecks {
	summary := domain.PullRequestChecks{Total: len(contexts)}
	for _, context := range contexts {
		switch classifyCheckContext(context) {
		case checkSummaryPassed:
			summary.Passed++
		case checkSummaryRunning:
			summary.Running++
		case checkSummaryFailed:
			summary.Failed++
		case checkSummarySkipped:
			summary.Skipped++
		default:
			summary.Waiting++
		}
	}
	return summary
}

func classifyCheckContext(context ghStatusCheckContext) checkSummaryClass {
	return classifyCheckSummary(context.Conclusion, context.Status, context.State)
}
