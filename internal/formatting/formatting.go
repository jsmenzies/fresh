package formatting

import (
	"fmt"
	"strings"
	"time"
)

const (
	TimeJustNow = "just now"
	TimeUnknown = "unknown"
)

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return TimeUnknown
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return TimeJustNow
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
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
