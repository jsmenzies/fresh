package git

import (
	"encoding/json"
	"fmt"
	"fresh/internal/config"
	"fresh/internal/domain"
	"fresh/internal/pullrequests"
	"os/exec"
	"sort"
	"strings"
)

type PullRequestService interface {
	GetPullRequestSync(repos []domain.Repository) PullRequestSync
}

type GhPullRequestService struct{}

type PullRequestSync struct {
	States  map[string]domain.PullRequestState
	Tracked []pullrequests.Snapshot
}

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
	Number           int     `json:"number"`
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
	return pullRequestService.GetPullRequestSync(repos).States
}

func GetPullRequestSync(repos []domain.Repository) PullRequestSync {
	return pullRequestService.GetPullRequestSync(repos)
}

func (GhPullRequestService) GetPullRequestSync(repos []domain.Repository) PullRequestSync {
	states := make(map[string]domain.PullRequestState, len(repos))
	result := PullRequestSync{
		States:  states,
		Tracked: nil,
	}

	githubByPath, ownerRepos := collectGitHubRepos(repos)
	for _, repo := range repos {
		if _, ok := githubByPath[repo.Path]; !ok {
			states[repo.Path] = domain.PullRequestUnavailable{}
			continue
		}
		states[repo.Path] = domain.PullRequestCount{}
	}

	if len(ownerRepos) == 0 {
		return result
	}

	openCounts, err := queryOpenPullRequestCounts(ownerRepos)
	if err != nil {
		result.States = markGitHubReposError(states, githubByPath, err)
		return result
	}

	tracked, mySummaries, err := queryMyPullRequests(ownerRepos)
	if err != nil {
		result.States = markGitHubReposError(states, githubByPath, err)
		return result
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

	result.Tracked = tracked
	return result
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

func parseNameWithOwner(nameWithOwner string) (owner string, repo string, ok bool) {
	if strings.TrimSpace(nameWithOwner) == "" {
		return "", "", false
	}

	parts := strings.Split(nameWithOwner, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
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

func queryMyPullRequests(ownerRepos []string) ([]pullrequests.Snapshot, map[string]myPullRequestSummary, error) {
	queryText := "is:pr is:open author:@me " + strings.Join(prefixRepoQualifiers(ownerRepos), " ")

	query := `
query($q: String!) {
  search(type: ISSUE, query: $q, first: 100) {
    nodes {
      ... on PullRequest {
        repository {
          nameWithOwner
        }
        number
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
		return nil, nil, normalizeGhError(err)
	}

	var response gqlSearchResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, nil, fmt.Errorf("failed to parse gh my-pr response: %w", err)
	}

	tracked := make([]pullrequests.Snapshot, 0, len(response.Data.Search.Nodes))
	summaries := make(map[string]myPullRequestSummary)
	for _, row := range response.Data.Search.Nodes {
		ownerRepo := row.Repository.NameWithOwner
		owner, repo, ok := parseNameWithOwner(ownerRepo)
		if !ok || row.Number <= 0 {
			continue
		}

		classification := classifyMyPullRequest(row)
		tracked = append(tracked, pullrequests.Snapshot{
			Key: pullrequests.Key{
				Owner:  owner,
				Repo:   repo,
				Number: row.Number,
			},
			Status: classification,
		})

		summary := summaries[ownerRepo]
		summary.MyOpen++

		switch classification {
		case pullrequests.StatusReady:
			summary.MyReady++
		case pullrequests.StatusBlocked:
			summary.MyBlocked++
		case pullrequests.StatusReview:
			summary.MyReview++
		default:
			summary.MyChecks++
		}

		summaries[ownerRepo] = summary
	}

	sort.Slice(tracked, func(i, j int) bool {
		return tracked[i].Key.String() < tracked[j].Key.String()
	})

	return tracked, summaries, nil
}

func classifyMyPullRequest(row gqlPullRequestNode) pullrequests.Status {
	if row.IsDraft {
		return pullrequests.StatusBlocked
	}

	mergeState := strings.ToUpper(strings.TrimSpace(derefString(row.MergeStateStatus)))
	if mergeState == "DIRTY" || mergeState == "CONFLICTING" || mergeState == "BLOCKED" || mergeState == "UNKNOWN" {
		return pullrequests.StatusBlocked
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
		return pullrequests.StatusBlocked
	}

	review := strings.ToUpper(strings.TrimSpace(derefString(row.ReviewDecision)))
	if review == "CHANGES_REQUESTED" {
		return pullrequests.StatusBlocked
	}

	if hasChecks && hasPendingChecks {
		return pullrequests.StatusChecks
	}

	if review == "APPROVED" {
		return pullrequests.StatusReady
	}

	if review == "REVIEW_REQUIRED" {
		return pullrequests.StatusReview
	}

	if review == "" {
		// GitHub can return a nil reviewDecision when no approving review is required.
		// If checks are clear and merge state is otherwise healthy, treat as ready.
		if mergeStateUnstable {
			return pullrequests.StatusChecks
		}
		return pullrequests.StatusReady
	}

	return pullrequests.StatusReview
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
