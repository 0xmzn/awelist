package enricher

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabProvider struct {
	token  string
	client *gitlab.Client
	logger *slog.Logger
}

func NewGitlabProvider(token string, logger *slog.Logger) *GitlabProvider {
	c, _ := gitlab.NewClient(token)
	return &GitlabProvider{
		token:  token,
		logger: logger.With("component", "gitlab-provider"),
		client: c,
	}
}

func (p *GitlabProvider) Name() string {
	return "gitlab-provider"
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

	reserved := map[string]bool{
		"dashboard": true,
		"projects":  true,
		"groups":    true,
		"users":     true,
		"help":      true,
		"explore":   true,
		"stats":     true,
		"search":    true,
	}

	if reserved[parts[0]] {
		return false
	}

	return true
}

func (p *GitlabProvider) Enrich(urls []string) (*ProviderAttemptResult, error) {
	results := make(map[string]*types.GitRepoMetadata)
	skipped := make(map[string]string)
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
					TotalAttemptedLinks: totalLinkCount,
					SuccessfulAttempts:  successfulLinks,
					FailedAttempts:      failedLinks,
					EnrichedUrls:        results,
					SkippedUrls:         skipped,
				}, rateLimitErr
			}

			p.logger.Warn("skipping url", "url", u, "error", err)
			skipped[u] = err.Error()
			continue
		}
		successfulLinks++
		results[u] = meta
	}

	return &ProviderAttemptResult{
		TotalAttemptedLinks: totalLinkCount,
		SuccessfulAttempts:  successfulLinks,
		FailedAttempts:      failedLinks,
		EnrichedUrls:        results,
		SkippedUrls:         skipped,
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

	if resp != nil {
		remaining := resp.Header.Get("RateLimit-Remaining")
		p.logger.Info("fetched repository",
			"repo", projectPath,
			"remaining_api_calls", remaining,
		)
	}

	meta := p.extractMetadataFromProject(project)
	return &meta, nil
}

func (p *GitlabProvider) extractMetadataFromProject(project *gitlab.Project) types.GitRepoMetadata {
	return types.GitRepoMetadata{
		Stars:      int(project.StarCount),
		IsArchived: project.Archived,
		LastUpdate: *project.LastActivityAt,
		EnrichedAt: time.Now(),
	}
}

func (p *GitlabProvider) getPath(rawURL string) (string, error) {
	u, _ := url.Parse(rawURL)
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", fmt.Errorf("not a repository URL")
	}

	return fmt.Sprintf("%s/%s", parts[0], parts[1]), nil
}
