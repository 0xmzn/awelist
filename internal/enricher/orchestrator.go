package enricher

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/gosimple/slug"
)

type Orchestrator struct {
	providers  []Provider
	reconciler Reconciler
}

func NewOrchestrator(reconciler Reconciler, providers ...Provider) *Orchestrator {
	return &Orchestrator{
		providers:  providers,
		reconciler: reconciler,
	}
}

func (o *Orchestrator) EnrichList(yamlList types.AwesomeList, jsonList types.AwesomeList, ttl time.Duration) ([]types.ProviderMetrics, map[string]string, []string, error) {
	o.setSlugs(yamlList)

	allLinks := o.reconciler.Reconcile(yamlList, jsonList, ttl)
	fmt.Printf("Attempting to enrich %d links\n", len(allLinks))

	providerMap := make(map[Provider][]string)
	linkMap := make(map[string]*types.Link)
	failedLinks := make(map[string]string)
	var allMetrics []types.ProviderMetrics
	var unhandled []string

	for _, link := range allLinks {
		linkMap[link.URL] = link
		handled := false
		for _, p := range o.providers {
			if p.CanHandle(link.URL) {
				providerMap[p] = append(providerMap[p], link.URL)
				handled = true
				break
			}
		}
		if !handled {
			unhandled = append(unhandled, link.URL)
		}
	}

	for p, urls := range providerMap {
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
			fmt.Fprintf(os.Stderr, "warning: %s rate limit reached: %v\n", p.Name(), err)
			continue
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s enrichment failed: %v\n", p.Name(), err)
			continue
		}

	}

	fmt.Println("Enrichment complete.")
	return allMetrics, failedLinks, unhandled, nil
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
