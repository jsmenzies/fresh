package git

import (
	"fresh/internal/config"
	"fresh/internal/domain"
)

// Client abstracts git operations used by the UI layer.
type Client interface {
	BuildRepository(path string) domain.Repository
	Fetch(repoPath string) error
	Pull(repoPath string, lineCallback func(string)) int
	DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (exitCode int, deletedCount int)
}

// ExecClient runs git operations via os/exec-backed functions in this package.
type ExecClient struct {
	cfg *config.Config
}

var _ Client = (*ExecClient)(nil)

func NewExecClient(cfg *config.Config) *ExecClient {
	return &ExecClient{cfg: cfg}
}

func (c *ExecClient) BuildRepository(path string) domain.Repository {
	return buildRepository(path, c.cfg)
}

func (c *ExecClient) Fetch(repoPath string) error {
	return fetch(repoPath)
}

func (c *ExecClient) Pull(repoPath string, lineCallback func(string)) int {
	return pull(repoPath, lineCallback)
}

func (c *ExecClient) DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (int, int) {
	return deleteBranches(repoPath, branches, lineCallback)
}
