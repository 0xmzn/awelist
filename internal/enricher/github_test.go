package enricher

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/go-github/v82/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestGithubProvider_Enrich(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Successfully enrich multiple repos", func(t *testing.T) {
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatch(
				mock.GetReposByOwnerByRepo,
				github.Repository{
					StargazersCount: github.Ptr(150),
					Name:            github.Ptr("repo-a"),
				},
				github.Repository{
					StargazersCount: github.Ptr(42),
					Name:            github.Ptr("repo-b"),
				},
			),
		)

		ghClient := github.NewClient(mockedHTTPClient)
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
		mockedHTTPClient := mock.NewMockedHTTPClient(
			mock.WithRequestMatchHandler(
				mock.GetReposByOwnerByRepo,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mock.WriteError(
						w,
						http.StatusInternalServerError,
						"github is down",
					)
				}),
			),
		)

		ghClient := github.NewClient(mockedHTTPClient)
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
