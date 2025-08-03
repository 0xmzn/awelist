package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"os"
)

type CLI struct {
	AwesomeFile string `kong:"short='f',long='awesome-file',help='path to awesome file.',default='awesome.yaml'"`
	Debug       bool   `kong:"long='debug',help='print raw error messages on error.'"`

	Generate GenerateCmd `kong:"cmd,help='generate file from template.'"`
	Enrich   EnrichCmd   `kong:"cmd,help='enrich YAML file. On success, a awesome-lock.json file will be created.'"`
	Add      AddCmd      `kong:"cmd,help='Add item to list.'"`
}

type GenerateCmd struct {
	HTML         bool   `kong:"short='H',help='output HTML'"`
	TemplateFile string `kong:"arg,required,help='path to template file.'"`
}

type EnrichCmd struct{}

type AddCmd struct{}

func CliErrorf(err error, format string, a ...any) error {
	if _debugMode {
		if err != nil {
			return err
		}
		return fmt.Errorf(format, a...)
	}
	return fmt.Errorf(format, a...)
}

var _debugMode bool

func main() {
	var cli CLI
	parser := kong.Parse(
		&cli,
		kong.Name("awelist"),
		kong.Description("A CLI tool for managing awesome lists."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	_debugMode = cli.Debug

	err := parser.Run(&cli)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
