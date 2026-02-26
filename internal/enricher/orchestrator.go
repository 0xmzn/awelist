package enricher

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/gosimple/slug"
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

func (o *Orchestrator) EnrichList(yamlList types.AwesomeList, jsonList types.AwesomeList) ([]types.ProviderMetrics, map[string]string, error) {
	o.setSlugs(yamlList)

	allLinks := o.reconciler.Reconcile(yamlList, jsonList)
	fmt.Printf("Attempting to enrich %d links\n", len(allLinks))

	providerMap := make(map[Provider][]string)
	linkMap := make(map[string]*types.Link)
	failedLinks := make(map[string]string)
	var allMetrics []types.ProviderMetrics

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
		o.logger.Debug("enriching links via provider", "name", p.Name(), "count", len(urls))
		fmt.Printf("Attempting to enrich %d links via %s\n", len(urls), p.Name())

		results, err := p.Enrich(urls)

		if results != nil {
			allMetrics = append(allMetrics, results.Metrics)

			for skippedURL, reason := range results.SkippedUrls {
				failedLinks[skippedURL] = reason
			}

			for url, meta := range results.EnrichedUrls {
				if link, ok := linkMap[url]; ok {
					link.RepoMetadata = meta
				}
			}
		}

		var ratelimitErr *ErrProviderRateLimit
		if errors.As(err, &ratelimitErr) {
			o.logger.Error("provider rate limit reached", "name", p.Name(), "error", err)
			continue
		}

		if err != nil {
			o.logger.Error("provider enrichment failed", "name", p.Name(), "error", err)
			continue
		}

	}

	fmt.Println("Enrichment attempts completed. Run 'awelist report' for more information")
	return allMetrics, failedLinks, nil
}

func (o *Orchestrator) setSlugs(list types.AwesomeList) {
	var setSlugRecursive func(cat *types.Category)
	setSlugRecursive = func(cat *types.Category) {
		cat.Slug = slug.Make(cat.Title)
		for _, sub := range cat.Subcategories {
			setSlugRecursive(sub)
		}
	}

	for _, c := range list {
		setSlugRecursive(c)
	}
}
