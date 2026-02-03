package enricher

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/google/go-github/v82/github"
)

func newLink(url string) *types.Link {
	return &types.Link{URL: url, Title: "Test"}
}

func TestOrchestrator_EnrichList(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reconciler := NewReconciler()

	t.Run("successfully routes links to correct providers", func(t *testing.T) {
		urlGH := "https://github.com/user/repo"
		urlGL := "https://gitlab.com/user/repo"

		yamlList := types.AwesomeList{
			{Links: []*types.Link{newLink(urlGH), newLink(urlGL)}},
		}

		ghProvider := &mockProvider{
			name:          "github",
			canHandleFunc: func(u string) bool { return u == urlGH },
			enrichFunc: func(urls []string) (*EnrichmentResult, error) {
				return &EnrichmentResult{
					EnrichedUrls: map[string]*types.GitRepoMetadata{urlGH: {Stars: 10}},
				}, nil
			},
		}

		glProvider := &mockProvider{
			name:          "gitlab",
			canHandleFunc: func(u string) bool { return u == urlGL },
			enrichFunc: func(urls []string) (*EnrichmentResult, error) {
				return &EnrichmentResult{
					EnrichedUrls: map[string]*types.GitRepoMetadata{urlGL: {Stars: 20}},
				}, nil
			},
		}

		orch := NewOrchestrator(logger, reconciler, ghProvider, glProvider)
		err := orch.EnrichList(yamlList, nil)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if yamlList[0].Links[0].RepoMetadata.Stars != 10 {
			t.Errorf("GitHub link stars mismatch: got %d", yamlList[0].Links[0].RepoMetadata.Stars)
		}
		if yamlList[0].Links[1].RepoMetadata.Stars != 20 {
			t.Errorf("GitLab link stars mismatch: got %d", yamlList[0].Links[1].RepoMetadata.Stars)
		}
	})

	t.Run("handles provider rate limit without stopping", func(t *testing.T) {
		url := "https://github.com/limit/me"
		yamlList := types.AwesomeList{{Links: []*types.Link{newLink(url)}}}

		p := &mockProvider{
			name:          "github",
			canHandleFunc: func(u string) bool { return true },
			enrichFunc: func(urls []string) (*EnrichmentResult, error) {
				return &EnrichmentResult{EnrichedUrls: nil}, &github.RateLimitError{Message: "limit reached"}
			},
		}

		orch := NewOrchestrator(logger, reconciler, p)
		err := orch.EnrichList(yamlList, nil)

		if err != nil {
			t.Errorf("Orchestrator should swallow rate limit errors, but returned: %v", err)
		}
	})

	t.Run("preserves data received before a failure", func(t *testing.T) {
		url1 := "https://ok.com"
		url2 := "https://fail.com"
		yamlList := types.AwesomeList{{Links: []*types.Link{newLink(url1), newLink(url2)}}}

		p := &mockProvider{
			name:          "mixed",
			canHandleFunc: func(u string) bool { return true },
			enrichFunc: func(urls []string) (*EnrichmentResult, error) {
				res := &EnrichmentResult{
					EnrichedUrls: map[string]*types.GitRepoMetadata{url1: {Stars: 5}},
				}
				return res, errors.New("something went wrong")
			},
		}

		orch := NewOrchestrator(logger, reconciler, p)
		_ = orch.EnrichList(yamlList, nil)

		if yamlList[0].Links[0].RepoMetadata == nil {
			t.Error("expected link1 to have metadata despite subsequent error")
		}
		if yamlList[0].Links[1].RepoMetadata != nil {
			t.Error("expected link2 to be empty as it wasn't in EnrichedUrls")
		}
	})

	t.Run("traverses deep category trees", func(t *testing.T) {
		url := "https://deep.com"
		yamlList := types.AwesomeList{
			{
				Title: "Root",
				Subcategories: []*types.Category{
					{Title: "Nested", Links: []*types.Link{newLink(url)}},
				},
			},
		}

		p := &mockProvider{
			name:          "any",
			canHandleFunc: func(u string) bool { return true },
			enrichFunc: func(urls []string) (*EnrichmentResult, error) {
				return &EnrichmentResult{
					EnrichedUrls: map[string]*types.GitRepoMetadata{url: {Stars: 123}},
				}, nil
			},
		}

		orch := NewOrchestrator(logger, reconciler, p)
		_ = orch.EnrichList(yamlList, nil)

		meta := yamlList[0].Subcategories[0].Links[0].RepoMetadata
		if meta == nil || meta.Stars != 123 {
			t.Errorf("failed to enrich deeply nested link, got: %+v", meta)
		}
	})

	t.Run("ignores urls with no matching provider", func(t *testing.T) {
		yamlList := types.AwesomeList{{Links: []*types.Link{newLink("https://unsupported.com")}}}

		p := &mockProvider{
			name:          "github-only",
			canHandleFunc: func(u string) bool { return false },
		}

		orch := NewOrchestrator(logger, reconciler, p)
		err := orch.EnrichList(yamlList, nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if yamlList[0].Links[0].RepoMetadata != nil {
			t.Error("metadata should remain nil for unsupported URLs")
		}
	})
}

type mockProvider struct {
	name          string
	canHandleFunc func(string) bool
	enrichFunc    func([]string) (*EnrichmentResult, error)
}

func (m *mockProvider) Name() string            { return m.name }
func (m *mockProvider) CanHandle(u string) bool { return m.canHandleFunc(u) }
func (m *mockProvider) Enrich(urls []string) (*EnrichmentResult, error) {
	if m.enrichFunc != nil {
		return m.enrichFunc(urls)
	}
	return &EnrichmentResult{}, nil
}
