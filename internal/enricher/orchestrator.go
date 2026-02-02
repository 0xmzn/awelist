package enricher

import (
	"errors"
	"log/slog"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/google/go-github/v82/github"
)

type Orchestrator struct {
	providers []Provider
	logger    *slog.Logger
}

func NewOrchestrator(logger *slog.Logger, providers ...Provider) *Orchestrator {
	return &Orchestrator{
		providers: providers,
		logger:    logger.With("component", "orchestrator"),
	}
}

func (o *Orchestrator) EnrichList(list types.AwesomeList) error {
	allLinks := list.Flatten()
	o.logger.Info("starting enrichment", "total_links", len(allLinks))

	providerMap := make(map[Provider][]string)
	linkMap := make(map[string]*types.Link)

	for _, link := range allLinks {
		linkMap[link.URL] = link
		for _, p := range o.providers {
			if p.CanHandle(link.URL) {
				providerMap[p] = append(providerMap[p], link.URL)
				break
			}
		}
	}

	for p, urls := range providerMap {
		o.logger.Info("enriching links via provider", "name", p.Name(), "count", len(urls))

		results, err := p.Enrich(urls)

		// extract enriched links, if any, before handling error
		for url, meta := range results {
			if link, ok := linkMap[url]; ok {
				link.RepoMetadata = meta
			}
		}

		var ratelimitErr *github.RateLimitError
		if errors.As(err, &ratelimitErr) {
			o.logger.Error("provider rate limit reached", "name", p.Name(), "error", err)
			continue
		}

		if err != nil {
			o.logger.Error("provider enrichment failed", "name", p.Name(), "error", err)
			continue
		}

	}

	o.logger.Info("enrichment complete")
	return nil
}
