package main

import (
	"fmt"
	"os"

	"github.com/0xmzn/awelist/internal/cli"
	"github.com/0xmzn/awelist/internal/enricher"
	"github.com/0xmzn/awelist/internal/list"
	"github.com/0xmzn/awelist/internal/store"
)

func main() {
	ctx, app := cli.Parse(os.Args[1:])

	ghToken := os.Getenv("GITHUB_TOKEN")
	glToken := os.Getenv("GITLAB_TOKEN")

	store := store.New(app.AwesomeFile, app.AwesomeLock)
	mngr := list.NewManager()

	glProvider, err := enricher.NewGitlabProvider(glToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize GitLab provider: %v\n", err)
		os.Exit(1)
	}

	orch := enricher.NewOrchestrator(enricher.NewReconciler(), enricher.NewGithubProvider(ghToken), glProvider)

	deps := &cli.Dependencies{
		Store:       store,
		ListManager: mngr,
		Enricher:    orch,
	}

	err = ctx.Run(deps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
