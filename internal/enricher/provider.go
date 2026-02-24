package enricher

import "github.com/0xmzn/awelist/internal/types"

type Provider interface {
	Name() string
	CanHandle(url string) bool

	Enrich(urls []string) (*ProviderAttemptResult, error)
}

type ProviderAttemptResult struct {
	TotalAttemptedLinks int
	SuccessfulAttempts  int
	FailedAttempts      int
	EnrichedUrls        map[string]*types.GitRepoMetadata
	SkippedUrls         map[string]string
}
