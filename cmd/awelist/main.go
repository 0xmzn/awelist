package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/0xmzn/awelist/internal/cli"
	"github.com/0xmzn/awelist/internal/enricher"
	"github.com/0xmzn/awelist/internal/list"
	"github.com/0xmzn/awelist/internal/store"
)

func main() {
	ctx, app := cli.Parse(os.Args[1:])

	loggerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if app.Debug {
		loggerOpts.Level = slog.LevelDebug
	}

	token := os.Getenv("GITHUB_TOKEN")
	logger := slog.New(slog.NewTextHandler(os.Stderr, loggerOpts))
	store := store.New(app.AwesomeFile, app.AwesomeLock)
	mngr := list.NewManager()
	enricher := enricher.NewOrchestrator(logger, enricher.NewGithubProvider(token, logger))

	deps := &cli.Dependencies{
		Logger:      logger,
		Store:       store,
		ListManager: mngr,
		Enricher:    enricher,
	}

	err := ctx.Run(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
