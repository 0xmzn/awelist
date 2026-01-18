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

func (o *Orchestrator) EnrichList(list types.AwesomeList) error {
	allLinks := list.Flatten()

	for _, v := range allLinks {
		fmt.Println(v.Title)
	}
	return nil
}
