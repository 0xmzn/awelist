package cli

import (
	"github.com/0xmzn/awelist/internal/enricher"
	"github.com/0xmzn/awelist/internal/list"
	"github.com/0xmzn/awelist/internal/store"
)

type Dependencies struct {
	Store       *store.FileStore
	Enricher    *enricher.Orchestrator
	ListManager *list.Manager
}
