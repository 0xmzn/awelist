package enricher

import (
	"time"

	"github.com/0xmzn/awelist/internal/types"
)

type Reconciler struct {
}

func NewReconciler() Reconciler {
	return Reconciler{}
}

func (r *Reconciler) Reconcile(yamlList types.AwesomeList, jsonList types.AwesomeList) []*types.Link {
	currentState := make(map[string]*types.Link)
	if jsonList != nil {
		for _, link := range jsonList.Flatten() {
			currentState[link.URL] = link
		}
	}

	var newLinks []*types.Link
	var staleLinks []*types.Link

	for _, link := range yamlList.Flatten() {
		existing, found := currentState[link.URL]

		if !found {
			newLinks = append(newLinks, link)
		} else {
			link.RepoMetadata = existing.RepoMetadata

			if link.RepoMetadata == nil || time.Since(link.RepoMetadata.EnrichedAt) > 24*time.Hour {
				staleLinks = append(staleLinks, link)
			}
		}
	}

	return append(newLinks, staleLinks...)
}
