package enricher

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	"gitlab.com/gitlab-org/api/client-go"
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

func (p *GitlabProvider) Enrich(urls []string) (*EnrichmentResult, error) {
	results := make(map[string]*types.GitRepoMetadata)
	var skipped []string

	for _, u := range urls {
		meta, err := p.enrichSingle(u)
		if err != nil {
			p.logger.Warn("skipping url", "url", u, "error", err)
			skipped = append(skipped, u)
			continue
		}
		results[u] = meta
	}

	return &EnrichmentResult{EnrichedUrls: results, SkippedUrls: skipped}, nil
}

func (p *GitlabProvider) enrichSingle(u string) (*types.GitRepoMetadata, error) {
	projectPath, err := p.getPath(u)
	if err != nil {
		return nil, err
	}

	project, resp, err := p.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return nil, err
	}

	remaining := resp.Header.Get("RateLimit-Remaining")

	p.logger.Info("fetched repository",
		"repo", projectPath,
		"remaining_api_calls", remaining,
	)

	meta := p.extractMetadataFromProject(project)
	return &meta, nil
}

func (p *GitlabProvider) extractMetadataFromProject(project *gitlab.Project) types.GitRepoMetadata {
	return types.GitRepoMetadata{
		Stars:      int(project.StarCount),
		IsArchived: project.Archived,
		EnrichedAt: time.Now(),
	}
}

func (p *GitlabProvider) getPath(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return "", fmt.Errorf("invalid gitlab path")
	}

	return path, nil
}
