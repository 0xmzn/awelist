package enricher

import (
	"fmt"
	"time"

	"github.com/0xmzn/awelist/internal/types"
)

type Provider interface {
	Name() string
	CanHandle(url string) bool

	Enrich(urls []string) (*EnrichmentResult, error)
}

type EnrichmentResult struct {
	EnrichedUrls map[string]*types.GitRepoMetadata
	SkippedUrls  []string
}

type ProviderRateLimitError struct {
	ID        string
	Limit     int
	Remaining int
	ResetAt   time.Time
}

func (p *ProviderRateLimitError) Error() string {
	return fmt.Sprintf("%s: rate limit exceeded. Limit: %d, Remaining: %d, Reset %s", p.ID, p.Limit, p.Remaining, p.ResetAt)
}
