package enricher

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var gitlabReservedPaths = map[string]bool{
	"dashboard": true,
	"projects":  true,
	"groups":    true,
	"users":     true,
	"help":      true,
	"explore":   true,
	"stats":     true,
	"search":    true,
}

type GitlabProvider struct {
	token  string
	client *gitlab.Client
}

func NewGitlabProvider(token string) (*GitlabProvider, error) {
	c, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}
	return &GitlabProvider{
		token:  token,
		client: c,
	}, nil
}

func (p *GitlabProvider) Name() string {
	return "GitLab"
}

func (p *GitlabProvider) CanHandle(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if u.Hostname() != "gitlab.com" && u.Hostname() != "www.gitlab.com" {
		return false
	}

	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return false
	}

	if gitlabReservedPaths[parts[0]] {
		return false
	}

	return true
}

func (p *GitlabProvider) Enrich(urls []string) (*ProviderAttemptResult, error) {
	results := make(map[string]*types.GitRepoMetadata)
	skipped := make(map[string]string)

	if p.token == "" {
		fmt.Fprintln(os.Stderr, "warning: GITLAB_TOKEN not set, skipping GitLab enrichment")
		for _, u := range urls {
			skipped[u] = "GITLAB_TOKEN not set"
		}
		return &ProviderAttemptResult{
			Metrics:      types.ProviderMetrics{Provider: p.Name(), Attempted: len(urls), Failed: len(urls)},
			EnrichedUrls: results,
			SkippedUrls:  skipped,
		}, nil
	}

	totalLinkCount := len(urls)
	successfulLinks := 0
	failedLinks := 0

	for _, u := range urls {
		meta, err := p.enrichSingle(u)
		if err != nil {
			failedLinks++
			var rateLimitErr *ErrProviderRateLimit
			if errors.As(err, &rateLimitErr) {
				skipped[u] = rateLimitErr.Error()

				return &ProviderAttemptResult{
					Metrics: types.ProviderMetrics{
						Provider:   p.Name(),
						Attempted:  totalLinkCount,
						Successful: successfulLinks,
						Failed:     failedLinks,
					},
					EnrichedUrls: results,
					SkippedUrls:  skipped,
				}, rateLimitErr
			}

			skipped[u] = err.Error()
			continue
		}
		successfulLinks++
		results[u] = meta
	}

	return &ProviderAttemptResult{
		Metrics: types.ProviderMetrics{
			Provider:   p.Name(),
			Attempted:  totalLinkCount,
			Successful: successfulLinks,
			Failed:     failedLinks,
		},
		EnrichedUrls: results,
		SkippedUrls:  skipped,
	}, nil
}

func (p *GitlabProvider) enrichSingle(u string) (*types.GitRepoMetadata, error) {
	projectPath, err := p.getPath(u)
	if err != nil {
		return nil, err
	}

	project, resp, err := p.client.Projects.GetProject(projectPath, nil)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			limitStr := resp.Header.Get("RateLimit-Limit")
			remainingStr := resp.Header.Get("RateLimit-Remaining")
			resetStr := resp.Header.Get("RateLimit-Reset")

			limit, _ := strconv.Atoi(limitStr)
			remaining, _ := strconv.Atoi(remainingStr)
			resetEpoch, _ := strconv.ParseInt(resetStr, 10, 64)

			return nil, &ErrProviderRateLimit{
				ID:        p.Name(),
				Limit:     limit,
				Remaining: remaining,
				ResetAt:   time.Unix(resetEpoch, 0),
			}
		}
		return nil, err
	}

	meta := p.extractMetadataFromProject(project)
	return &meta, nil
}

func (p *GitlabProvider) extractMetadataFromProject(project *gitlab.Project) types.GitRepoMetadata {
	var lastUpdate time.Time
	if project.LastActivityAt != nil {
		lastUpdate = *project.LastActivityAt
	}

	return types.GitRepoMetadata{
		Stars:      int(project.StarCount),
		IsArchived: project.Archived,
		LastUpdate: lastUpdate,
		EnrichedAt: time.Now(),
	}
}

func (p *GitlabProvider) getPath(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", fmt.Errorf("not a repository URL")
	}

	return fmt.Sprintf("%s/%s", parts[0], parts[1]), nil
}
