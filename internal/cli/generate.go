package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/0xmzn/awelist/internal/generator"
	"github.com/0xmzn/awelist/internal/types"
)

type GenerateCmd struct {
	HTML         bool   `kong:"short='H',help='output HTML.'"`
	OutputFile   string `kong:"short='o',long='output',help='path to output file. defaults to stdout.'"`
	TemplateFile string `kong:"arg,required,help='path to template file.'"`
}

func (cmd *GenerateCmd) Run(deps *Dependencies) error {
	log := deps.Logger
	store := deps.Store

	log.Debug("Running generate", "template", cmd.TemplateFile)

	var list types.AwesomeList
	var err error

	lock, err := store.LoadLockFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No lock file found, performing generation without enrichment\n")
		list, err = store.LoadAwesomeFile()
		if err != nil {
			return fmt.Errorf("failed to load any data source: %w", err)
		}
	} else {
		list = lock.List
	}

	var buf bytes.Buffer
	if err := generator.GenerateOutput(&buf, cmd.TemplateFile, cmd.HTML, list); err != nil {
		return err
	}

	var writer io.Writer = os.Stdout
	if cmd.OutputFile != "" {
		f, err := os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("could not create output file: %w", err)
		}
		defer f.Close()
		writer = f
		log.Debug("writing output to file", "path", cmd.OutputFile)
	}

	_, err = io.Copy(writer, &buf)
	return err
}
