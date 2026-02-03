package cli

import (
	"bytes"
	"fmt"
	html "html/template"
	"io"
	"os"
	text "text/template"

	"github.com/0xmzn/awelist/internal/types"
)

type GenerateCmd struct {
	HTML         bool   `kong:"short='H',help='output HTML.'"`
	TemplateFile string `kong:"arg,required,help='path to template file.'"`
}

func (cmd *GenerateCmd) Run(deps *Dependencies) error {
	log := deps.Logger
	store := deps.Store

	log.Info("Running generate")
	log.Debug("Debugging generate")

	var enrichedList types.AwesomeList

	enrichedList, err := store.LoadJson()

	if err != nil {
		return err
	}

	buffer, err := executeTemplate(cmd.TemplateFile, cmd.HTML, enrichedList)
	if err != nil {
		return err
	}

	fmt.Printf("%s", buffer.String())
	return nil

}

type template interface {
	Execute(wr io.Writer, data any) error
}

func newTemplate(isHTML bool, name string, content string) (template, error) {
	if isHTML {
		tmpl, err := html.New(name).Parse(content)
		if err != nil {
			return nil, err
		}
		return tmpl, nil
	} else {
		tmpl, err := text.New(name).Parse(content)
		if err != nil {
			return nil, err
		}
		return tmpl, nil
	}
}

func executeTemplate(filename string, isHtml bool, data types.AwesomeList) (bytes.Buffer, error) {
	tmplContent, err := os.ReadFile(filename)
	if err != nil {
		return bytes.Buffer{}, err
	}

	tmpl, err := newTemplate(isHtml, filename, string(tmplContent))
	if err != nil {
		return bytes.Buffer{}, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return buffer, nil
}
