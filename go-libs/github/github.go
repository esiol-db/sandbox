package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/databricks/databricks-sdk-go/httpclient"
)

const gitHubAPI = "https://api.github.com"

type GitHubClient struct {
	api *httpclient.ApiClient
	cfg *GitHubConfig
}

type GitHubConfig struct {
	GitHubTokenSource

	RetryTimeout       time.Duration
	HTTPTimeout        time.Duration
	InsecureSkipVerify bool
	DebugHeaders       bool
	DebugTruncateBytes int
	RateLimitPerSecond int

	transport http.RoundTripper
}

func NewClient(cfg *GitHubConfig) *GitHubClient {
	return &GitHubClient{
		api: httpclient.NewApiClient(httpclient.ClientConfig{
			Visitors: []httpclient.RequestVisitor{func(r *http.Request) error {
				token, err := cfg.Token()
				if err != nil {
					return fmt.Errorf("token: %w", err)
				}
				auth := fmt.Sprintf("%s %s", token.TokenType, token.AccessToken)
				r.Header.Set("Authorization", auth)
				return nil
			}},
			RetryTimeout:       cfg.RetryTimeout,
			HTTPTimeout:        cfg.HTTPTimeout,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			DebugHeaders:       cfg.DebugHeaders,
			DebugTruncateBytes: cfg.DebugTruncateBytes,
			RateLimitPerSecond: cfg.RateLimitPerSecond,
			Transport:          cfg.transport,
		}),
		cfg: cfg,
	}
}

func (c *GitHubClient) Versions(ctx context.Context, org, repo string) (Versions, error) {
	var releases Versions
	url := fmt.Sprintf("%s/repos/%s/%s/releases", gitHubAPI, org, repo)
	err := c.api.Do(ctx, "GET", url, httpclient.WithResponseUnmarshal(&releases))
	return releases, err
}

type CreateReleaseRequest struct {
	TagName                string `json:"tag_name,omitempty"`
	Name                   string `json:"name,omitempty"`
	Body                   string `json:"body,omitempty"`
	Draft                  bool   `json:"draft,omitempty"`
	Prerelease             bool   `json:"prerelease,omitempty"`
	GenerateReleaseNotes   bool   `json:"generate_release_notes,omitempty"`
	DiscussionCategoryName string `json:"discussion_category_name,omitempty"`
}

func (c *GitHubClient) CreateRelease(ctx context.Context, org, repo string, req CreateReleaseRequest) (*Release, error) {
	var res Release
	url := fmt.Sprintf("%s/repos/%s/%s/releases", gitHubAPI, org, repo)
	err := c.api.Do(ctx, "POST", url,
		httpclient.WithRequestData(req),
		httpclient.WithResponseUnmarshal(&res))
	return &res, err
}

func (c *GitHubClient) GetRepo(ctx context.Context, org, name string) (repo Repo, err error) {
	url := fmt.Sprintf("%s/repos/%s/%s", gitHubAPI, org, name)
	err = c.api.Do(ctx, "GET", url, httpclient.WithResponseUnmarshal(&repo))
	return
}

func (c *GitHubClient) ListRepositories(ctx context.Context, org string) (Repositories, error) {
	var repos Repositories
	url := fmt.Sprintf("%s/users/%s/repos", gitHubAPI, org)
	err := c.api.Do(ctx, "GET", url, httpclient.WithResponseUnmarshal(&repos))
	return repos, err
}

func (c *GitHubClient) ListRuns(ctx context.Context, org, repo, workflow string) ([]workflowRun, error) {
	path := fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%v.yml/runs", gitHubAPI, org, repo, workflow)
	var response struct {
		TotalCount   *int          `json:"total_count,omitempty"`
		WorkflowRuns []workflowRun `json:"workflow_runs,omitempty"`
	}
	err := c.api.Do(ctx, "GET", path, httpclient.WithResponseUnmarshal(&response))
	return response.WorkflowRuns, err
}

func (c *GitHubClient) CompareCommits(ctx context.Context, org, repo, base, head string) ([]RepositoryCommit, error) {
	path := fmt.Sprintf("%s/repos/%v/%v/compare/%v...%v", gitHubAPI, org, repo, base, head)
	var response struct {
		Commits []RepositoryCommit `json:"commits,omitempty"`
	}
	err := c.api.Do(ctx, "GET", path, httpclient.WithResponseUnmarshal(&response))
	return response.Commits, err
}

func (c *GitHubClient) ListPullRequests(ctx context.Context, org, repo string, opts PullRequestListOptions) ([]PullRequest, error) {
	path := fmt.Sprintf("%s/repos/%s/%s/pulls", gitHubAPI, org, repo)
	var prs []PullRequest
	err := c.api.Do(ctx, "GET", path,
		httpclient.WithRequestData(opts),
		httpclient.WithResponseUnmarshal(&prs))
	return prs, err
}

func (c *GitHubClient) EditPullRequest(ctx context.Context, org, repo string, number int, body PullRequestUpdate) error {
	path := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", gitHubAPI, org, repo, number)
	return c.api.Do(ctx, "PATCH", path, httpclient.WithRequestData(body))
}

func (c *GitHubClient) CreatePullRequest(ctx context.Context, org, repo string, body NewPullRequest) (*PullRequest, error) {
	path := fmt.Sprintf("%s/repos/%s/%s/pulls", gitHubAPI, org, repo)
	var res PullRequest
	err := c.api.Do(ctx, "POST", path,
		httpclient.WithRequestData(body),
		httpclient.WithResponseUnmarshal(&res))
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *GitHubClient) GetPullRequest(ctx context.Context, org, repo string, number int) (*PullRequest, error) {
	path := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", gitHubAPI, org, repo, number)
	var res PullRequest
	err := c.api.Do(ctx, "GET", path,
		httpclient.WithResponseUnmarshal(&res))
	return &res, err
}
