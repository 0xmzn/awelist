package enricher

import "github.com/0xmzn/awelist/internal/types"

type Provider interface {
	Name() string
	CanHandle(url string) bool

	Enrich(urls []string) (map[string]*types.GitRepoMetadata, error)
}
