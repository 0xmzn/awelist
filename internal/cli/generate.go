package cli

import (
	"fmt"
	html "html/template"
	"io"
	"os"
	text "text/template"

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

	log.Info("Running generate", "template", cmd.TemplateFile)

	var list types.AwesomeList
	var err error

	list, err = store.LoadJson()
	if err != nil {
		log.Warn("lock file not found, using raw yaml", "error", err)
		list, err = store.LoadYAML()
		if err != nil {
			return fmt.Errorf("failed to load any data source: %w", err)
		}
	}

	var writer io.Writer = os.Stdout
	if cmd.OutputFile != "" {
		f, err := os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("could not create output file: %w", err)
		}
		defer f.Close()
		writer = f
		log.Info("writing output to file", "path", cmd.OutputFile)
	}

	return writeTemplate(writer, cmd.TemplateFile, cmd.HTML, list)
}

type template interface {
	Execute(wr io.Writer, data any) error
}

func newTemplate(isHTML bool, name string, content string) (template, error) {
	if isHTML {
		return html.New(name).Parse(content)
	}
	return text.New(name).Parse(content)
}

func writeTemplate(w io.Writer, filename string, isHtml bool, data types.AwesomeList) error {
	tmplContent, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	tmpl, err := newTemplate(isHtml, filename, string(tmplContent))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	return nil
}
