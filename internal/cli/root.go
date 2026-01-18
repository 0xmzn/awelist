package cli

import "github.com/alecthomas/kong"

type CLI struct {
	AwesomeFile string `name:"file" help:"Path to awesome.yaml." default:"awesome.yaml" type:"path"`
	AwesomeLock   string `name:"lock" help:"Path to awesome-lock.json." default:"awesome-lock.json" type:"path"`
	Debug       bool   `kong:"long='debug',help='print debug messages.'"`

	Add      AddCmd      `kong:"cmd,help='Add item to list.'"`
	Enrich   EnrichCmd   `kong:"cmd,help='Enrich YAML file. on success, awesome-lock.json file will be created.'"`
	Generate GenerateCmd `kong:"cmd,help='Generate file from template. uses dry data if awesome-lock.json does not exist.'"`
}

func Parse(args []string) (*kong.Context, *CLI) {
	var cli CLI

	ctx := kong.Parse(
		&cli,
		kong.Name("awelist"),
		kong.Description("A CLI tool for managing awesome lists"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	return ctx, &cli
}