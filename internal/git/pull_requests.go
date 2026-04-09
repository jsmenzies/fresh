package git

import (
	"encoding/json"
	"fmt"
	"fresh/internal/config"
	"fresh/internal/domain"
	"os/exec"
	"sort"
	"strings"
)

type PullRequestService interface {
	GetPullRequestStates(repos []domain.Repository) map[string]domain.PullRequestState
}

type GhPullRequestService struct{}

type myPullRequestSummary struct {
	MyOpen    int
	MyReady   int
	MyBlocked int
	MyReview  int
	MyChecks  int
}

type ghSearchPRRow struct {
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

type gqlSearchResponse struct {
	Data struct {
		Search struct {
			Nodes []gqlPullRequestNode `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

type gqlPullRequestNode struct {
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	IsDraft          bool    `json:"isDraft"`
	ReviewDecision   *string `json:"reviewDecision"`
	MergeStateStatus *string `json:"mergeStateStatus"`
	Commits          struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *gqlStatusCheckRollup `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

type gqlStatusCheckRollup struct {
	State    string `json:"state"`
	Contexts struct {
		Nodes []gqlStatusContextNode `json:"nodes"`
	} `json:"contexts"`
}

type gqlStatusContextNode struct {
	Conclusion string `json:"conclusion"`
	Status     string `json:"status"`
	State      string `json:"state"`
}

var pullRequestService PullRequestService = GhPullRequestService{}

func SetPullRequestService(service PullRequestService) {
	if service == nil {
		pullRequestService = GhPullRequestService{}
		return
	}
	pullRequestService = service
}

func GetPullRequestStates(repos []domain.Repository) map[string]domain.PullRequestState {
	return pullRequestService.GetPullRequestStates(repos)
}

func (GhPullRequestService) GetPullRequestStates(repos []domain.Repository) map[string]domain.PullRequestState {
	states := make(map[string]domain.PullRequestState, len(repos))

	githubByPath, ownerRepos := collectGitHubRepos(repos)
	for _, repo := range repos {
		if _, ok := githubByPath[repo.Path]; !ok {
			states[repo.Path] = domain.PullRequestUnavailable{}
			continue
		}
		states[repo.Path] = domain.PullRequestCount{}
	}

	if len(ownerRepos) == 0 {
		return states
	}

	openCounts, err := queryOpenPullRequestCounts(ownerRepos)
	if err != nil {
		return markGitHubReposError(states, githubByPath, err)
	}

	mySummaries, err := queryMyPullRequestSummaries(ownerRepos)
	if err != nil {
		return markGitHubReposError(states, githubByPath, err)
	}

	for path, ownerRepo := range githubByPath {
		state := domain.PullRequestCount{}
		state.Open = openCounts[ownerRepo]
		if summary, ok := mySummaries[ownerRepo]; ok {
			state.MyOpen = summary.MyOpen
			state.MyReady = summary.MyReady
			state.MyBlocked = summary.MyBlocked
			state.MyReview = summary.MyReview
			state.MyChecks = summary.MyChecks
		}
		states[path] = state
	}

	return states
}

func collectGitHubRepos(repos []domain.Repository) (map[string]string, []string) {
	byPath := make(map[string]string, len(repos))
	ownerRepoSet := make(map[string]struct{}, len(repos))

	for _, repo := range repos {
		owner, name, ok := parseGitHubRemote(repo.RemoteURL)
		if !ok {
			continue
		}
		ownerRepo := owner + "/" + name
		byPath[repo.Path] = ownerRepo
		ownerRepoSet[ownerRepo] = struct{}{}
	}

	ownerRepos := make([]string, 0, len(ownerRepoSet))
	for ownerRepo := range ownerRepoSet {
		ownerRepos = append(ownerRepos, ownerRepo)
	}
	sort.Strings(ownerRepos)

	return byPath, ownerRepos
}

func parseGitHubRemote(remoteURL string) (owner string, repo string, ok bool) {
	if remoteURL == "" {
		return "", "", false
	}

	normalized := strings.TrimSpace(remoteURL)
	normalized = strings.TrimSuffix(normalized, ".git")

	slug := ""
	switch {
	case strings.HasPrefix(normalized, "git@github.com:"):
		slug = strings.TrimPrefix(normalized, "git@github.com:")
	case strings.HasPrefix(normalized, "ssh://git@github.com/"):
		slug = strings.TrimPrefix(normalized, "ssh://git@github.com/")
	case strings.HasPrefix(normalized, "https://github.com/"):
		slug = strings.TrimPrefix(normalized, "https://github.com/")
	case strings.HasPrefix(normalized, "http://github.com/"):
		slug = strings.TrimPrefix(normalized, "http://github.com/")
	default:
		return "", "", false
	}

	parts := strings.Split(slug, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func queryOpenPullRequestCounts(ownerRepos []string) (map[string]int, error) {
	args := []string{"search", "prs", "--state", "open", "--limit", "100", "--json", "repository"}
	for _, ownerRepo := range ownerRepos {
		args = append(args, "--repo", ownerRepo)
	}
	args = append(args, "--")

	cmd := createCommand(config.DefaultConfig().Timeout.Default, "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, normalizeGhError(err)
	}

	var rows []ghSearchPRRow
	if err := json.Unmarshal(output, &rows); err != nil {
		return nil, fmt.Errorf("failed to parse gh open-pr response: %w", err)
	}

	counts := make(map[string]int)
	for _, row := range rows {
		if row.Repository.NameWithOwner == "" {
			continue
		}
		counts[row.Repository.NameWithOwner]++
	}

	return counts, nil
}

func queryMyPullRequestSummaries(ownerRepos []string) (map[string]myPullRequestSummary, error) {
	queryText := "is:pr is:open author:@me " + strings.Join(prefixRepoQualifiers(ownerRepos), " ")

	query := `
query($q: String!) {
  search(type: ISSUE, query: $q, first: 100) {
    nodes {
      ... on PullRequest {
        repository {
          nameWithOwner
        }
        isDraft
        reviewDecision
        mergeStateStatus
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup {
                state
                contexts(first: 50) {
                  nodes {
                    ... on CheckRun {
                      conclusion
                      status
                    }
                    ... on StatusContext {
                      state
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
`

	cmd := createCommand(
		config.DefaultConfig().Timeout.Default,
		"gh", "api", "graphql",
		"-f", "query="+query,
		"-F", "q="+queryText,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, normalizeGhError(err)
	}

	var response gqlSearchResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse gh my-pr response: %w", err)
	}

	summaries := make(map[string]myPullRequestSummary)
	for _, row := range response.Data.Search.Nodes {
		ownerRepo := row.Repository.NameWithOwner
		if ownerRepo == "" {
			continue
		}

		summary := summaries[ownerRepo]
		summary.MyOpen++

		switch classifyMyPullRequest(row) {
		case "ready":
			summary.MyReady++
		case "blocked":
			summary.MyBlocked++
		case "review":
			summary.MyReview++
		default:
			summary.MyChecks++
		}

		summaries[ownerRepo] = summary
	}

	return summaries, nil
}

func classifyMyPullRequest(row gqlPullRequestNode) string {
	if row.IsDraft {
		return "blocked"
	}

	mergeState := strings.ToUpper(strings.TrimSpace(derefString(row.MergeStateStatus)))
	if mergeState == "DIRTY" || mergeState == "CONFLICTING" || mergeState == "BLOCKED" || mergeState == "UNKNOWN" {
		return "blocked"
	}
	mergeStateUnstable := mergeState == "UNSTABLE"

	hasChecks := false
	hasFailingChecks := false
	hasPendingChecks := false

	rollup := latestStatusCheckRollup(row)
	if rollup != nil {
		hasChecks = true
		state := strings.ToUpper(strings.TrimSpace(rollup.State))
		if state == "FAILURE" || state == "ERROR" {
			hasFailingChecks = true
		}
		if state == "PENDING" || state == "EXPECTED" {
			hasPendingChecks = true
		}

		for _, node := range rollup.Contexts.Nodes {
			conclusion := strings.ToUpper(strings.TrimSpace(node.Conclusion))
			status := strings.ToUpper(strings.TrimSpace(node.Status))
			nodeState := strings.ToUpper(strings.TrimSpace(node.State))

			switch conclusion {
			case "FAILURE", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE":
				hasFailingChecks = true
			case "":
				if status != "COMPLETED" {
					hasPendingChecks = true
				}
			}

			if nodeState == "FAILURE" || nodeState == "ERROR" {
				hasFailingChecks = true
			}
			if nodeState == "PENDING" || nodeState == "EXPECTED" {
				hasPendingChecks = true
			}
		}
	}

	if hasFailingChecks {
		return "blocked"
	}

	review := strings.ToUpper(strings.TrimSpace(derefString(row.ReviewDecision)))
	if review == "CHANGES_REQUESTED" {
		return "blocked"
	}

	if hasChecks && hasPendingChecks {
		return "checks"
	}

	if review == "APPROVED" {
		return "ready"
	}

	if review == "REVIEW_REQUIRED" {
		return "review"
	}

	if review == "" {
		// GitHub can return a nil reviewDecision when no approving review is required.
		// If checks are clear and merge state is otherwise healthy, treat as ready.
		if mergeStateUnstable {
			return "checks"
		}
		return "ready"
	}

	return "review"
}

func latestStatusCheckRollup(row gqlPullRequestNode) *gqlStatusCheckRollup {
	if len(row.Commits.Nodes) == 0 {
		return nil
	}
	return row.Commits.Nodes[0].Commit.StatusCheckRollup
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func prefixRepoQualifiers(ownerRepos []string) []string {
	qualifiers := make([]string, 0, len(ownerRepos))
	for _, ownerRepo := range ownerRepos {
		qualifiers = append(qualifiers, "repo:"+ownerRepo)
	}
	return qualifiers
}

func normalizeGhError(err error) error {
	if err == nil {
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr != "" {
			return fmt.Errorf("gh: %s", stderr)
		}
	}

	return err
}

func markGitHubReposError(states map[string]domain.PullRequestState, githubByPath map[string]string, err error) map[string]domain.PullRequestState {
	message := "gh integration unavailable"
	if err != nil {
		message = err.Error()
	}

	for path := range githubByPath {
		states[path] = domain.PullRequestError{Message: message}
	}

	return states
}
