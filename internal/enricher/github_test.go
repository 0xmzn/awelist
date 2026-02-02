package enricher

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/go-github/v82/github"
	"github.com/h2non/gock"
)

func TestGithubProvider_Enrich(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Successfully enrich multiple repos", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Get("/repos/user/repo-a").
			Reply(200).
			JSON(map[string]int{"stargazers_count": 150})

		gock.New("https://api.github.com").
			Get("/repos/user/repo-b").
			Reply(200).
			JSON(map[string]int{"stargazers_count": 42})

		httpClient := &http.Client{Transport: &gock.Transport{}}
		ghClient := github.NewClient(httpClient)

		provider := &GithubProvider{
			client: ghClient,
			logger: logger,
		}

		urls := []string{
			"https://github.com/user/repo-a",
			"https://github.com/user/repo-b",
		}

		results, err := provider.Enrich(urls)
		if err != nil {
			t.Fatalf("Enrich returned unexpected error: %v", err)
		}

		if meta, ok := results[urls[0]]; !ok {
			t.Errorf("Expected result for %s", urls[0])
		} else {
			if meta.Stars != 150 {
				t.Errorf("Expected 150 stars for repo-a, got %d", meta.Stars)
			}
			if meta.EnrichedAt.IsZero() {
				t.Error("EnrichedAt should be set")
			}
		}

		if meta, ok := results[urls[1]]; !ok {
			t.Errorf("Expected result for %s", urls[1])
		} else {
			if meta.Stars != 42 {
				t.Errorf("Expected 42 stars for repo-b, got %d", meta.Stars)
			}
		}
	})

	t.Run("Handle API errors gracefully", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Get("/repos/user/error-repo").
			Reply(http.StatusInternalServerError).
			JSON(map[string]string{"message": "github is down"})

		httpClient := &http.Client{Transport: &gock.Transport{}}
		ghClient := github.NewClient(httpClient)

		provider := &GithubProvider{
			client: ghClient,
			logger: logger,
		}

		urls := []string{"https://github.com/user/error-repo"}

		results, err := provider.Enrich(urls)

		if err != nil {
			t.Fatalf("Expected no execution error even on API fail, got: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected empty results on API error, got %d items", len(results))
		}
	})

	t.Run("Handle primary rate limit exceeded", func(t *testing.T) {
		defer gock.Off()

		gock.New("https://api.github.com").
			Get("/repos/user/repo").
			Reply(http.StatusForbidden).
			SetHeader("X-RateLimit-Remaining", "0").
			SetHeader("X-RateLimit-Reset", "1735689600").
			JSON(map[string]string{
				"message": "API rate limit exceeded",
			})

		httpClient := &http.Client{Transport: &gock.Transport{}}
		ghClient := github.NewClient(httpClient)

		provider := &GithubProvider{
			client: ghClient,
			logger: logger,
		}

		urls := []string{"https://github.com/user/repo"}

		results, err := provider.Enrich(urls)

		if err == nil {
			t.Fatal("Expected an error due to rate limit, got nil")
		}

		var rateLimitErr *github.RateLimitError
		if !errors.As(err, &rateLimitErr) {
			t.Errorf("Expected error to be *github.RateLimitError, got %T: %v", err, err)
		}

		if results != nil {
			t.Errorf("Expected nil results on rate limit error, got map with length %d", len(results))
		}
	})

	t.Run("Handle Malformed URLs in Enrich", func(t *testing.T) {
		ghClient := github.NewClient(nil)
		provider := &GithubProvider{
			client: ghClient,
			logger: logger,
		}

		urls := []string{"https://not-github.com/foo/bar", "invalid-url"}

		results, err := provider.Enrich(urls)
		if err != nil {
			t.Fatalf("Enrich returned unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results for invalid URLs, got %d", len(results))
		}
	})
}
