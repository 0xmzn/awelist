package main

import (
	"bytes"
	"flag"
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

func executeTemplate(filename string, isHtml bool, data awesomeList) (bytes.Buffer, error) {
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

func generate(args []string) error {
	flag := flag.NewFlagSet("generate", flag.ExitOnError)
	var htmlOutput bool
	flag.BoolVar(&htmlOutput, "html", false, "output html")
	registerGlobalFlags(flag)
	flag.Parse(args)

	if flag.NArg() != 1 {
		return CliErrorf(nil, "generate command requires exactly one argument: <filename>")
	}

	tmplFile := flag.Arg(0)

	awesomelist, err := loadFileIntoYaml(_awesomeFile)
	if err != nil {
		return err
	}

	buffer, err := executeTemplate(tmplFile, htmlOutput, awesomelist)
	if err != nil {
		return err
	}

	fmt.Printf("%s", buffer.String())
	return nil
}
