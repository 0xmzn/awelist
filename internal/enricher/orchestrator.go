package enricher

import (
	"errors"
	"log/slog"

	"github.com/0xmzn/awelist/internal/types"
)

type Orchestrator struct {
	providers  []Provider
	reconciler Reconciler
	logger     *slog.Logger
}

func NewOrchestrator(logger *slog.Logger, reconciler Reconciler, providers ...Provider) *Orchestrator {
	return &Orchestrator{
		providers:  providers,
		reconciler: reconciler,
		logger:     logger.With("component", "orchestrator"),
	}
}

func (o *Orchestrator) EnrichList(yamlList types.AwesomeList, jsonList types.AwesomeList) error {
	allLinks := o.reconciler.Reconcile(yamlList, jsonList)
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
		for url, meta := range results.EnrichedUrls {
			if link, ok := linkMap[url]; ok {
				link.RepoMetadata = meta
			}
		}

		var ratelimitErr *ProviderRateLimitError
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
