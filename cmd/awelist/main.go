package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/0xmzn/awelist/internal/cli"
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

	logger := slog.New(slog.NewTextHandler(os.Stderr, loggerOpts))
	store := store.New(app.AwesomeFile, app.AwesomeLock)
	mngr := list.NewManager()

	deps := &cli.Dependencies{
		Logger: logger,
		Store:  store,
		ListManager: mngr,
	}

	err := ctx.Run(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
