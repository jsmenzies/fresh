package git_test

import (
	"fresh/internal/git"
	"testing"
)

func TestConvertGitURLToBrowser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "ssh url",
			url:  "git@github.com:user/repo.git",
			want: "https://github.com/user/repo",
		},
		{
			name: "ssh url without .git suffix",
			url:  "git@github.com:user/repo",
			want: "https://github.com/user/repo",
		},
		{
			name: "https url with .git suffix",
			url:  "https://github.com/user/repo.git",
			want: "https://github.com/user/repo",
		},
		{
			name: "https url without .git suffix",
			url:  "https://github.com/user/repo",
			want: "https://github.com/user/repo",
		},
		{
			name: "non-standard url returned as-is",
			url:  "http://example.com/repo",
			want: "http://example.com/repo",
		},
		{
			name: "empty string",
			url:  "",
			want: "",
		},
		{
			name: "ssh url with nested path",
			url:  "git@gitlab.com:org/sub/repo.git",
			want: "https://gitlab.com/org/sub/repo",
		},
		{
			name: "malformed ssh url without colon",
			url:  "git@github.com",
			want: "git@github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := git.ConvertGitURLToBrowser(tt.url)
			if got != tt.want {
				t.Errorf("ConvertGitURLToBrowser(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsGitHubRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "github ssh url",
			url:  "git@github.com:user/repo.git",
			want: true,
		},
		{
			name: "github https url",
			url:  "https://github.com/user/repo.git",
			want: true,
		},
		{
			name: "gitlab url",
			url:  "git@gitlab.com:user/repo.git",
			want: false,
		},
		{
			name: "empty string",
			url:  "",
			want: false,
		},
		{
			name: "github in path but not host",
			url:  "https://example.com/github.com/repo",
			want: true, // contains "github.com" substring
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := git.IsGitHubRepository(tt.url)
			if got != tt.want {
				t.Errorf("IsGitHubRepository(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestExtractGitHubRepoInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "ssh url",
			url:       "git@github.com:octocat/hello-world.git",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "https url",
			url:       "https://github.com/octocat/hello-world.git",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "https url without .git",
			url:       "https://github.com/octocat/hello-world",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "non-github url returns empty",
			url:       "git@gitlab.com:user/repo.git",
			wantOwner: "",
			wantRepo:  "",
		},
		{
			name:      "empty url returns empty",
			url:       "",
			wantOwner: "",
			wantRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo := git.ExtractGitHubRepoInfo(tt.url)
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("ExtractGitHubRepoInfo(%q) = (%q, %q), want (%q, %q)",
					tt.url, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestBuildGitHubURLs(t *testing.T) {
	t.Parallel()

	t.Run("feature branch", func(t *testing.T) {
		t.Parallel()
		urls := git.BuildGitHubURLs("git@github.com:octocat/hello-world.git", "feature-x")
		if urls == nil {
			t.Fatal("expected non-nil urls map")
		}

		expected := map[string]string{
			"code":   "https://github.com/octocat/hello-world/tree/feature-x",
			"issues": "https://github.com/octocat/hello-world/issues",
			"prs":    "https://github.com/octocat/hello-world/pulls",
			"openpr": "https://github.com/octocat/hello-world/compare/feature-x",
		}
		for key, want := range expected {
			if got := urls[key]; got != want {
				t.Errorf("urls[%q] = %q, want %q", key, got, want)
			}
		}
	})

	t.Run("main branch uses base url for code", func(t *testing.T) {
		t.Parallel()
		urls := git.BuildGitHubURLs("git@github.com:octocat/hello-world.git", "main")
		if urls == nil {
			t.Fatal("expected non-nil urls map")
		}

		if got, want := urls["code"], "https://github.com/octocat/hello-world"; got != want {
			t.Errorf("urls[code] = %q, want %q", got, want)
		}
	})

	t.Run("master branch uses base url for code", func(t *testing.T) {
		t.Parallel()
		urls := git.BuildGitHubURLs("git@github.com:octocat/hello-world.git", "master")
		if urls == nil {
			t.Fatal("expected non-nil urls map")
		}

		if got, want := urls["code"], "https://github.com/octocat/hello-world"; got != want {
			t.Errorf("urls[code] = %q, want %q", got, want)
		}
	})

	t.Run("non-github url returns nil", func(t *testing.T) {
		t.Parallel()
		urls := git.BuildGitHubURLs("git@gitlab.com:user/repo.git", "main")
		if urls != nil {
			t.Errorf("expected nil for non-GitHub URL, got %v", urls)
		}
	})

	t.Run("empty branch", func(t *testing.T) {
		t.Parallel()
		urls := git.BuildGitHubURLs("git@github.com:octocat/hello-world.git", "")
		if urls == nil {
			t.Fatal("expected non-nil urls map")
		}

		// Empty branch should use base URL for code
		if got, want := urls["code"], "https://github.com/octocat/hello-world"; got != want {
			t.Errorf("urls[code] = %q, want %q", got, want)
		}
	})
}
