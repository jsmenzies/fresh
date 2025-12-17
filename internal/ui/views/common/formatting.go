package common

import (
	"fmt"
	"strings"
	"time"
)

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "< 2m"
		}
		return fmt.Sprintf("< %dm", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day "
		}
		return fmt.Sprintf("%d days", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}

	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year"
	}
	return fmt.Sprintf("%d years", years)
}

func ConvertGitURLToBrowser(gitURL string) string {
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

func IsGitHubRepository(remoteURL string) bool {
	if remoteURL == "" {
		return false
	}

	return strings.Contains(remoteURL, "github.com")
}

func ExtractGitHubRepoInfo(remoteURL string) (owner, repo string) {
	if !IsGitHubRepository(remoteURL) {
		return "", ""
	}

	browserURL := ConvertGitURLToBrowser(remoteURL)

	if strings.HasPrefix(browserURL, "https://github.com/") {
		path := strings.TrimPrefix(browserURL, "https://github.com/")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	return "", ""
}

func BuildGitHubURLs(remoteURL, branch string) map[string]string {
	if !IsGitHubRepository(remoteURL) {
		return nil
	}

	owner, repo := ExtractGitHubRepoInfo(remoteURL)
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

	if branch == "" || branch == "main" || branch == "master" {
		urls["code"] = baseURL
		if branch != "" {
			urls["openpr"] = fmt.Sprintf("%s/compare/%s", baseURL, branch)
		}
	}

	return urls
}
