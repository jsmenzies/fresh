package main

import (
	"fmt"
	"strings"
)

func convertGitURLToBrowser(gitURL string) string {
	if strings.HasPrefix(gitURL, "git@") {

		parts := strings.SplitN(gitURL, ":", 2)
		if len(parts) != 2 {
			return gitURL
		}

		host := strings.TrimPrefix(parts[0], "git@")
		path := parts[1]

		path = strings.TrimSuffix(path, ".git")

		return fmt.Sprintf("https://%s/%s", host, path)
	}

	if strings.HasPrefix(gitURL, "https://") {
		return strings.TrimSuffix(gitURL, ".git")
	}

	return gitURL
}

func isGitHubRepository(remoteURL string) bool {
	if remoteURL == "" {
		return false
	}

	return strings.Contains(remoteURL, "github.com")
}

// extractGitHubRepoInfo extracts owner and repo name from GitHub URL
func extractGitHubRepoInfo(remoteURL string) (owner, repo string) {
	if !isGitHubRepository(remoteURL) {
		return "", ""
	}

	browserURL := convertGitURLToBrowser(remoteURL)

	// Extract from https://github.com/owner/repo
	if strings.HasPrefix(browserURL, "https://github.com/") {
		path := strings.TrimPrefix(browserURL, "https://github.com/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	return "", ""
}

// buildGitHubURLs creates GitHub-specific URLs
func buildGitHubURLs(remoteURL, branch string) map[string]string {
	if !isGitHubRepository(remoteURL) {
		return nil
	}

	owner, repo := extractGitHubRepoInfo(remoteURL)
	if owner == "" || repo == "" {
		return nil
	}

	baseURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)

	urls := map[string]string{
		"code":   fmt.Sprintf("%s/tree/%s", baseURL, branch),
		"issues": fmt.Sprintf("%s/issues", baseURL),
		"prs":    fmt.Sprintf("%s/pulls", baseURL),
		"openpr": fmt.Sprintf("%s/compare/%s", baseURL, branch),
	}

	// Handle empty or main branches appropriately
	if branch == "" || branch == "main" || branch == "master" {
		urls["code"] = baseURL
		if branch != "" {
			urls["openpr"] = fmt.Sprintf("%s/compare/%s", baseURL, branch)
		}
	}

	return urls
}
