package enricher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/google/go-github/v82/github"
)

type GithubProvider struct {
	token  string
	client *github.Client
	logger *slog.Logger
}

func NewGithubProvider(token string, logger *slog.Logger) *GithubProvider {
	c := github.NewClient(nil).WithAuthToken(token)
	return &GithubProvider{
		token:  token,
		logger: logger.With("component", "github-provider"),
		client: c,
	}
}

func (p *GithubProvider) Name() string {
	return "github-provider"
}

func (p *GithubProvider) CanHandle(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if u.Hostname() != "github.com" && u.Hostname() != "www.github.com" {
		return false
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return false
	}

	reserved := map[string]bool{
		"features":    true,
		"topics":      true,
		"trending":    true,
		"search":      true,
		"settings":    true,
		"about":       true,
		"pricing":     true,
		"marketplace": true,
	}

	if reserved[parts[0]] {
		return false
	}

	return true
}

func (p *GithubProvider) Enrich(urls []string) (*EnrichmentResult, error) {
	results := make(map[string]*types.GitRepoMetadata)
	var skipped []string

	for _, u := range urls {
		p.logger.Debug("fetching repositories", "url", u)

		ownerName, repoName, err := p.parseURL(u)
		if err != nil {
			p.logger.Warn("failed to parse github url", "url", u, "error", err)
			skipped = append(skipped, u)
			continue
		}

		repo, resp, err := p.client.Repositories.Get(context.Background(), ownerName, repoName)

		var ratelimitErr *github.RateLimitError
		if errors.As(err, &ratelimitErr) {
			skipped = append(skipped, u)
			return &EnrichmentResult{results, skipped}, err
		}

		if err != nil {
			p.logger.Error("Getting repo failed", "repo_id", fmt.Sprintf("%s/%s", ownerName, repoName), "error", err)
			skipped = append(skipped, u)
			continue
		}

		stars := *repo.StargazersCount
		isArchived := *repo.Archived

		meta := types.GitRepoMetadata{
			Stars:      stars,
			IsArchived: isArchived,
			EnrichedAt: time.Now(),
		}

		results[u] = &meta

		p.logger.Info("successfully fetched repository", "repo", fmt.Sprintf("%s/%s", ownerName, repoName), "stars", stars, "ratelimit", resp.Rate.Remaining)

	}

	res := &EnrichmentResult{
		EnrichedUrls: results,
		SkippedUrls:  skipped,
	}

	return res, nil
}

func (p *GithubProvider) parseURL(rawURL string) (owner, repo string, err error) {
	u, _ := url.Parse(rawURL)
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("not a repository URL")
	}

	return parts[0], parts[1], nil
}
