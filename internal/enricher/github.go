package enricher

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/0xmzn/awelist/internal/types"
)


type GithubProvider struct {
	token string
	logger *slog.Logger
}

func NewGithubProvider(token string, logger *slog.Logger) *GithubProvider {
	return &GithubProvider{
		token: token,
		logger:    logger.With("component", "github-provider"),
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
		"features":  true,
		"topics":    true,
		"trending":  true,
		"search":    true,
		"settings":  true,
		"about":     true,
		"pricing":   true,
		"marketplace": true,
	}

	if reserved[parts[0]] {
		return false
	}

	return true
}

func (p *GithubProvider) Enrich(urls []string) (map[string]*types.GitRepoMetadata, error) {
	results := make(map[string]*types.GitRepoMetadata)

	for _, u := range urls {
		p.logger.Debug("fetching metadata", "url", u)
		
		owner, repo, err := p.parseURL(u)
		if err != nil {
			p.logger.Warn("failed to parse github url", "url", u, "error", err)
			continue
		}

		p.logger.Error("Skipping API", "repo", fmt.Sprintf("%s/%s", owner, repo), "error", err)
	}

	return results, nil
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