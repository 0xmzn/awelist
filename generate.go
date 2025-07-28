package main

import (
	"bytes"
	"fmt"
	html "html/template"
	"io"
	"os"
	text "text/template"
)

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

func executeTemplate(filename string, isHtml bool, data baseAwesomelist) (bytes.Buffer, error) {
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
	err := aweStore.Load()
	if err != nil {
		return err
	}

	awelist := aweStore.GetManager()

	buffer, err := executeTemplate(cmd.TemplateFile, cmd.HTML, awelist.RawList)
	if err != nil {
		return err
	}

	fmt.Printf("%s", buffer.String())
	return nil
}
