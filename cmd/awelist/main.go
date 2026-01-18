package main

import (
	"fmt"
	"github.com/0xmzn/awelist/internal/cli"
	"log/slog"
	"os"
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

	deps := &cli.Dependencies{
		Logger: logger,
	}

	err := ctx.Run(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
