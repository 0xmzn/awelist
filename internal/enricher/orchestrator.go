package enricher

import (
	"fmt"

	"github.com/0xmzn/awelist/internal/types"
)

type Orchestrator struct {
	providers []Provider
}

func NewOrchestrator(providers ...Provider) *Orchestrator {
	return &Orchestrator{providers: providers}
}

func (o *Orchestrator) EnrichList(list []*types.Category) error {
	allLinks := o.collectLinks(list)

	for _, v := range allLinks {
		fmt.Println(v.Title)
	}
	return nil
}

func (o *Orchestrator) collectLinks(cats []*types.Category) []*types.Link {
	var links []*types.Link
	for _, c := range cats {
		links = append(links, c.Links...)
		links = append(links, o.collectLinks(c.Subcategories)...)
	}
	return links
}
