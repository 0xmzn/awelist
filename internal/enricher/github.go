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
		meta, err := p.enrichSingle(u)

		if err != nil {
			var ghRatelimitErr *github.RateLimitError
			if errors.As(err, &ghRatelimitErr) {
				rateLimitErr := ProviderRateLimitError{
					ID:        p.Name(),
					Limit:     ghRatelimitErr.Rate.Limit,
					Remaining: ghRatelimitErr.Rate.Remaining,
					ResetAt:   ghRatelimitErr.Rate.Reset.Time,
				}
				return &EnrichmentResult{results, append(skipped, u)}, &rateLimitErr
			}

			p.logger.Warn("skipping url", "url", u, "error", err)
			skipped = append(skipped, u)
			continue
		}

		results[u] = meta
	}

	return &EnrichmentResult{EnrichedUrls: results, SkippedUrls: skipped}, nil
}

func (p *GithubProvider) enrichSingle(u string) (*types.GitRepoMetadata, error) {
	owner, name, err := p.parseURL(u)
	if err != nil {
		return nil, err
	}

	repo, resp, err := p.client.Repositories.Get(context.Background(), owner, name)
	if err != nil {
		return nil, err
	}

	p.logger.Info("fetched repository",
		"repo", fmt.Sprintf("%s/%s", owner, name),
		"remaining", resp.Rate.Remaining,
	)

	meta := p.extractMetadataFromRepo(repo)
	return &meta, nil
}

func (p *GithubProvider) extractMetadataFromRepo(repo *github.Repository) types.GitRepoMetadata {
	stars := repo.GetStargazersCount()
	isArchived := repo.GetArchived()

	meta := types.GitRepoMetadata{
		Stars:      stars,
		IsArchived: isArchived,
		EnrichedAt: time.Now(),
	}

	return meta
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
