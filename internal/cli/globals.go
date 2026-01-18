package cli

import (
	"log/slog"

	"github.com/0xmzn/awelist/internal/enricher"
	"github.com/0xmzn/awelist/internal/store"
)

type Dependencies struct {
	Logger   *slog.Logger
	Store    *store.FileStore
	Enricher *enricher.Orchestrator
}
