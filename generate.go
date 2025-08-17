package main

import (
	"bytes"
	"fmt"
	html "html/template"
	"io"
	"os"
	text "text/template"
)

type GenerateCmd struct {
	HTML         bool   `kong:"short='H',help='output HTML'"`
	TemplateFile string `kong:"arg,required,help='path to template file.'"`
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

func executeTemplate(filename string, isHtml bool, data enrichedAwesomelist) (bytes.Buffer, error) {
	tmplContent, err := os.ReadFile(filename)
	if err != nil {
		return bytes.Buffer{}, CliErrorf(err, "failed to read file %q", filename)
	}

	tmpl, err := newTemplate(isHtml, filename, string(tmplContent))
	if err != nil {
		return bytes.Buffer{}, CliErrorf(err, "invalid template syntax in %q", filename)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return bytes.Buffer{}, CliErrorf(err, "failed while executing template %q", filename)
	}

	return buffer, nil
}

func (cmd *GenerateCmd) Run(cli *CLI) error {
	aweStore := NewAwesomeStore(cli.AwesomeFile)
	enrichedList, err := aweStore.LoadJSON()
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
