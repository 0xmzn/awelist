package enricher

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const defaultGithubBatchSize = 100

type repoNode struct {
	StargazerCount   githubv4.Int
	IsArchived       githubv4.Boolean
	DefaultBranchRef *defaultBranchRefNode
}

type defaultBranchRefNode struct {
	Target struct {
		Commit struct {
			CommittedDate githubv4.DateTime
		} `graphql:"... on Commit"`
	}
}

type rateLimitInfo struct {
	Limit     githubv4.Int
	Remaining githubv4.Int
	ResetAt   githubv4.DateTime
}

type repoTarget struct {
	url   string
	owner string
	name  string
}

type GithubProvider struct {
	token     string
	gqlClient *githubv4.Client
	logger    *slog.Logger
	batchSize int
}

func NewGithubProvider(token string, logger *slog.Logger) *GithubProvider {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)

	return &GithubProvider{
		token:     token,
		logger:    logger.With("component", "github-provider"),
		gqlClient: githubv4.NewClient(httpClient),
		batchSize: defaultGithubBatchSize,
	}
}

func (p *GithubProvider) Name() string {
	return "github-provider"
}

var githubReservedPaths = map[string]bool{
	"features":    true,
	"topics":      true,
	"trending":    true,
	"search":      true,
	"settings":    true,
	"about":       true,
	"pricing":     true,
	"marketplace": true,
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

	return len(parts) == 2 && !githubReservedPaths[parts[0]]
}

func (p *GithubProvider) Enrich(urls []string) (*ProviderAttemptResult, error) {
	results := make(map[string]*types.GitRepoMetadata)
	skipped := make(map[string]string)

	if p.token == "" {
		p.logger.Warn("skipping GitHub enrichment: GITHUB_TOKEN not set")
		for _, u := range urls {
			skipped[u] = "GITHUB_TOKEN not set"
		}
		return &ProviderAttemptResult{
			Metrics:      types.ProviderMetrics{Provider: p.Name(), Attempted: len(urls), Failed: len(urls)},
			EnrichedUrls: results,
			SkippedUrls:  skipped,
		}, nil
	}

	var targets []repoTarget
	for _, u := range urls {
		owner, name, err := p.parseURL(u)
		if err != nil {
			skipped[u] = err.Error()
			continue
		}
		targets = append(targets, repoTarget{url: u, owner: owner, name: name})
	}

	batches := chunkTargets(targets, p.batchSize)

	var lastRL rateLimitInfo
	var stopErr error

	for i, batch := range batches {
		rl, err := p.enrichBatch(batch, lastRL, results, skipped)
		if err != nil {
			stopErr = err
			markSkipped(batches[i:], skipped, err.Error())
			break
		}

		if rl != (rateLimitInfo{}) {
			lastRL = rl
		}

		if rl.Remaining <= 0 {
			stopErr = &ErrProviderRateLimit{
				ID:        p.Name(),
				Limit:     int(rl.Limit),
				Remaining: int(rl.Remaining),
				ResetAt:   rl.ResetAt.Time,
			}
			markSkipped(batches[i+1:], skipped, stopErr.Error())
			break
		}
	}

	return &ProviderAttemptResult{
		Metrics: types.ProviderMetrics{
			Provider:   p.Name(),
			Attempted:  len(urls),
			Successful: len(results),
			Failed:     len(skipped),
		},
		EnrichedUrls: results,
		SkippedUrls:  skipped,
	}, stopErr
}

func (p *GithubProvider) enrichBatch(batch []repoTarget, lastRL rateLimitInfo, results map[string]*types.GitRepoMetadata, skipped map[string]string) (rateLimitInfo, error) {
	queryPtr := reflect.New(buildQueryType(len(batch)))

	variables := make(map[string]any, len(batch)*2)
	for i, t := range batch {
		variables[fmt.Sprintf("o%d", i)] = githubv4.String(t.owner)
		variables[fmt.Sprintf("n%d", i)] = githubv4.String(t.name)
	}

	err := p.gqlClient.Query(context.Background(), queryPtr.Interface(), variables)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "401") {
			return rateLimitInfo{}, &ErrProviderAuth{ID: p.Name(), Reason: "invalid or missing token"}
		}
		if strings.Contains(strings.ToLower(msg), "rate limit") {
			return rateLimitInfo{}, &ErrProviderRateLimit{
				ID:        p.Name(),
				Limit:     int(lastRL.Limit),
				Remaining: int(lastRL.Remaining),
				ResetAt:   lastRL.ResetAt.Time,
			}
		}
		p.logger.Debug("batch returned partial errors", "error", err)
	}

	qv := queryPtr.Elem()
	rl := qv.FieldByName("RateLimit").Interface().(rateLimitInfo)

	for i, t := range batch {
		repo := qv.FieldByName(fmt.Sprintf("R%d", i)).Interface().(*repoNode)
		if repo == nil {
			skipped[t.url] = "repository not found, renamed, or inaccessible"
			continue
		}

		var lastUpdate time.Time
		if repo.DefaultBranchRef != nil {
			lastUpdate = repo.DefaultBranchRef.Target.Commit.CommittedDate.Time
		}

		results[t.url] = &types.GitRepoMetadata{
			Stars:      int(repo.StargazerCount),
			IsArchived: bool(repo.IsArchived),
			LastUpdate: lastUpdate,
			EnrichedAt: time.Now(),
		}
	}

	return rl, nil
}

func buildQueryType(n int) reflect.Type {
	fields := []reflect.StructField{
		{Name: "RateLimit", Type: reflect.TypeOf(rateLimitInfo{})},
	}
	for i := range n {
		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("R%d", i),
			Type: reflect.TypeOf((*repoNode)(nil)),
			Tag:  reflect.StructTag(fmt.Sprintf(`graphql:"r%d: repository(owner: $o%d, name: $n%d)"`, i, i, i)),
		})
	}
	return reflect.StructOf(fields)
}

func chunkTargets(targets []repoTarget, size int) [][]repoTarget {
	var batches [][]repoTarget
	for i := 0; i < len(targets); i += size {
		batches = append(batches, targets[i:min(i+size, len(targets))])
	}
	return batches
}

func markSkipped(batches [][]repoTarget, skipped map[string]string, reason string) {
	for _, batch := range batches {
		for _, t := range batch {
			if _, ok := skipped[t.url]; !ok {
				skipped[t.url] = reason
			}
		}
	}
}

func (p *GithubProvider) parseURL(rawURL string) (owner, repo string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %w", err)
	}
	path := strings.Trim(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("not a repository URL")
	}

	return parts[0], parts[1], nil
}
