package enricher

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestGitlabProvider_enrichSingle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name          string
		url           string
		mockSetup     func()
		wantErr       bool
		checkError    func(error) bool
		expectedStars int
	}{
		{
			name: "Success - Standard Repo",
			url:  "https://gitlab.com/user/repo-a",
			mockSetup: func() {
				gock.New("https://gitlab.com").
					Get("/api/v4/projects/user/repo-a").
					Reply(200).
					JSON(map[string]any{
						"star_count":       150,
						"archived":         false,
						"last_activity_at": "2024-02-25T12:00:00Z",
					})
			},
			wantErr:       false,
			expectedStars: 150,
		},
		{
			name: "Failure - Rate Limit Exceeded",
			url:  "https://gitlab.com/busy/repo",
			mockSetup: func() {
				gock.New("https://gitlab.com").
					Get("/api/v4/projects/busy/repo").
					Reply(429).
					SetHeader("RateLimit-Limit", "60").
					SetHeader("RateLimit-Remaining", "0").
					SetHeader("RateLimit-Reset", "1740000000").
					JSON(map[string]string{"message": "Retry later"})
			},
			wantErr: true,
			checkError: func(err error) bool {
				var rlErr *ErrProviderRateLimit
				return errors.As(err, &rlErr)
			},
		},
		{
			name:    "Failure - Invalid URL Path",
			url:     "https://gitlab.com/just-root",
			wantErr: true,
		},
		{
			name: "Failure - Repo Not Found (404)",
			url:  "https://gitlab.com/ghost/phantom-repo",
			mockSetup: func() {
				gock.New("https://gitlab.com").
					Get("/api/v4/projects/ghost/phantom-repo").
					Reply(404).
					JSON(map[string]string{"message": "404 Project Not Found"})
			},
			wantErr: true,
		},
		{
			name: "Failure - Internal Server Error (500)",
			url:  "https://gitlab.com/oops/broken",
			mockSetup: func() {
				gock.New("https://gitlab.com").
					Get("/api/v4/projects/oops/broken").
					Reply(500).
					JSON(map[string]string{"message": "Something went wrong"})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			httpClient := &http.Client{Transport: &gock.Transport{}}

			// Disable go-github's retries.
			// Without this, the 429 test retries, hits Gock a 2nd time, finds no mock, and fails.
			// Gock .Presist can be used as well.
			noRetries := func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				return false, nil
			}

			glClient, err := gitlab.NewClient("dummy-token",
				gitlab.WithHTTPClient(httpClient),
				gitlab.WithCustomRetry(noRetries),
			)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			provider := &GitlabProvider{
				client: glClient,
				logger: logger,
			}

			results, err := provider.enrichSingle(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("enrichSingle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkError != nil && err != nil {
				if !tt.checkError(err) {
					t.Errorf("enrichSingle() error type mismatch: %v", err)
				}
			}

			if !tt.wantErr && results != nil {
				if results.Stars != tt.expectedStars {
					t.Errorf("Expected %d stars, got %d", tt.expectedStars, results.Stars)
				}
			}
		})
	}
}
