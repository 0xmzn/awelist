package enricher

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/shurcooL/githubv4"
)

func newTestGithubProvider(logger *slog.Logger, batchSize int) *GithubProvider {
	return &GithubProvider{
		token:     "test-token",
		gqlClient: githubv4.NewClient(&http.Client{Transport: &gock.Transport{}}),
		logger:    logger,
		batchSize: batchSize,
	}
}

func TestGithubProvider_Enrich(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Successfully enrich multiple repos in a single batch", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": map[string]any{
					"rateLimit": map[string]any{"limit": 5000, "remaining": 4998, "resetAt": "2024-01-01T00:00:00Z"},
					"r0":        map[string]any{"stargazerCount": 150, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-25T12:00:00Z"}}},
					"r1":        map[string]any{"stargazerCount": 42, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-20T08:00:00Z"}}},
				},
			})

		provider := newTestGithubProvider(logger, 50)

		urls := []string{
			"https://github.com/user/repo-a",
			"https://github.com/user/repo-b",
		}

		results, err := provider.Enrich(urls)
		if err != nil {
			t.Fatalf("Enrich returned unexpected error: %v", err)
		}

		if results.Metrics.Attempted != 2 || results.Metrics.Successful != 2 || results.Metrics.Failed != 0 {
			t.Errorf("unexpected metrics: %+v", results.Metrics)
		}

		meta := results.EnrichedUrls[urls[0]]
		if meta == nil || meta.Stars != 150 || meta.IsArchived {
			t.Fatalf("unexpected metadata for repo-a: %+v", meta)
		}
		if meta.LastUpdate.IsZero() || meta.EnrichedAt.IsZero() {
			t.Error("expected LastUpdate and EnrichedAt to be set for repo-a")
		}

		meta = results.EnrichedUrls[urls[1]]
		if meta == nil || meta.Stars != 42 {
			t.Errorf("unexpected metadata for repo-b: %+v", meta)
		}
	})

	t.Run("Splits requests into multiple batches", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": map[string]any{
					"rateLimit": map[string]any{"limit": 5000, "remaining": 4999, "resetAt": "2024-01-01T00:00:00Z"},
					"r0":        map[string]any{"stargazerCount": 150, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-25T12:00:00Z"}}},
				},
			})

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": map[string]any{
					"rateLimit": map[string]any{"limit": 5000, "remaining": 4998, "resetAt": "2024-01-01T00:00:00Z"},
					"r0":        map[string]any{"stargazerCount": 42, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-20T08:00:00Z"}}},
				},
			})

		provider := newTestGithubProvider(logger, 1)

		urls := []string{
			"https://github.com/user/repo-a",
			"https://github.com/user/repo-b",
		}

		results, err := provider.Enrich(urls)
		if err != nil {
			t.Fatalf("Enrich returned unexpected error: %v", err)
		}

		if results.Metrics.Successful != 2 {
			t.Errorf("expected 2 successful, got %d", results.Metrics.Successful)
		}
		if results.EnrichedUrls[urls[0]] == nil || results.EnrichedUrls[urls[0]].Stars != 150 {
			t.Errorf("unexpected result for repo-a: %+v", results.EnrichedUrls[urls[0]])
		}
		if results.EnrichedUrls[urls[1]] == nil || results.EnrichedUrls[urls[1]].Stars != 42 {
			t.Errorf("unexpected result for repo-b: %+v", results.EnrichedUrls[urls[1]])
		}
	})

	t.Run("Marks unresolved repositories as skipped", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": map[string]any{
					"rateLimit": map[string]any{"limit": 5000, "remaining": 4999, "resetAt": "2024-01-01T00:00:00Z"},
					"r0":        nil,
					"r1":        map[string]any{"stargazerCount": 42, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-20T08:00:00Z"}}},
				},
				"errors": []map[string]any{
					{"message": "Could not resolve to a Repository with the name 'user/repo-missing'.", "type": "NOT_FOUND", "path": []string{"r0"}},
				},
			})

		provider := newTestGithubProvider(logger, 50)

		urls := []string{
			"https://github.com/user/repo-missing",
			"https://github.com/user/repo-b",
		}

		results, err := provider.Enrich(urls)
		if err != nil {
			t.Fatalf("Enrich returned unexpected error: %v", err)
		}

		if results.Metrics.Successful != 1 || results.Metrics.Failed != 1 {
			t.Errorf("unexpected metrics: %+v", results.Metrics)
		}
		if _, ok := results.SkippedUrls[urls[0]]; !ok {
			t.Errorf("expected %s to be skipped, got skipped=%v", urls[0], results.SkippedUrls)
		}
		if results.EnrichedUrls[urls[1]] == nil || results.EnrichedUrls[urls[1]].Stars != 42 {
			t.Errorf("unexpected result for %s: %+v", urls[1], results.EnrichedUrls[urls[1]])
		}
	})

	t.Run("Stops proactively when rate limit is exhausted", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": map[string]any{
					"rateLimit": map[string]any{"limit": 5000, "remaining": 0, "resetAt": "2024-01-01T00:00:00Z"},
					"r0":        map[string]any{"stargazerCount": 150, "isArchived": false, "defaultBranchRef": map[string]any{"target": map[string]any{"committedDate": "2024-02-25T12:00:00Z"}}},
				},
			})

		provider := newTestGithubProvider(logger, 1)

		urls := []string{
			"https://github.com/user/repo-a",
			"https://github.com/user/repo-b",
		}

		results, err := provider.Enrich(urls)

		var rateLimitErr *ErrProviderRateLimit
		if !errors.As(err, &rateLimitErr) {
			t.Fatalf("expected *ErrProviderRateLimit, got %T: %v", err, err)
		}

		if results.EnrichedUrls[urls[0]] == nil || results.EnrichedUrls[urls[0]].Stars != 150 {
			t.Errorf("expected repo-a to be enriched, got %+v", results.EnrichedUrls[urls[0]])
		}
		if _, ok := results.SkippedUrls[urls[1]]; !ok {
			t.Errorf("expected repo-b to be skipped due to rate limit, got skipped=%v", results.SkippedUrls)
		}
	})

	t.Run("Handles reactive rate limit errors", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(200).
			JSON(map[string]any{
				"data": nil,
				"errors": []map[string]any{
					{"message": "API rate limit exceeded for installation ID 123."},
				},
			})

		provider := newTestGithubProvider(logger, 50)

		urls := []string{"https://github.com/user/repo-a"}

		results, err := provider.Enrich(urls)

		var rateLimitErr *ErrProviderRateLimit
		if !errors.As(err, &rateLimitErr) {
			t.Fatalf("expected *ErrProviderRateLimit, got %T: %v", err, err)
		}
		if _, ok := results.SkippedUrls[urls[0]]; !ok {
			t.Errorf("expected %s to be skipped, got skipped=%v", urls[0], results.SkippedUrls)
		}
	})

	t.Run("Returns ErrProviderAuth on 401", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Post("/graphql").
			Reply(401).
			JSON(map[string]string{"message": "Bad credentials"})

		provider := newTestGithubProvider(logger, 50)

		urls := []string{"https://github.com/user/repo-a"}

		results, err := provider.Enrich(urls)

		var authErr *ErrProviderAuth
		if !errors.As(err, &authErr) {
			t.Fatalf("expected *ErrProviderAuth, got %T: %v", err, err)
		}
		if _, ok := results.SkippedUrls[urls[0]]; !ok {
			t.Errorf("expected %s to be skipped, got skipped=%v", urls[0], results.SkippedUrls)
		}
	})
}
