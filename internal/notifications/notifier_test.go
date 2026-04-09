package notifications

import "testing"

func TestBuildPayload_IncludesPullRequestTitleInNotificationTitle(t *testing.T) {
	t.Parallel()

	notification := Notification{
		Key:              PRKey{Owner: "acme", Repo: "api", Number: 12},
		Kind:             KindBlocked,
		Reason:           "acme/api#12 is blocked",
		PullRequestTitle: "Improve API docs",
	}

	title, body, soundPath := buildPayload(notification)

	if title != "api: Improve API docs" {
		t.Fatalf("title = %q, want %q", title, "api: Improve API docs")
	}
	if body != "Blocked - acme/api#12 is blocked" {
		t.Fatalf("body = %q, want %q", body, "Blocked - acme/api#12 is blocked")
	}
	if soundPath != blowSoundPath {
		t.Fatalf("soundPath = %q, want %q", soundPath, blowSoundPath)
	}
}

func TestBuildPayload_WithoutPullRequestTitleUsesRepoName(t *testing.T) {
	t.Parallel()

	notification := Notification{
		Key:              PRKey{Owner: "acme", Repo: "api", Number: 12},
		Kind:             KindProgress,
		Reason:           "acme/api#12 is mergeable",
		PullRequestTitle: "   ",
	}

	title, _, _ := buildPayload(notification)

	if title != "api" {
		t.Fatalf("title = %q, want %q", title, "api")
	}
}
