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

	tmplContent, err := os.ReadFile(tmplFile)
	if err != nil {
		return CliErrorf(err, "failed to read file %q", tmplFile)
	}

	tmpl, err := newTemplate(htmlOutput, tmplFile, string(tmplContent))
	if err != nil {
		return CliErrorf(err, "invalid template syntax in %q", tmplFile)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, awesomelist)
	if err != nil {
		return CliErrorf(err, "failed while executing template %q", tmplFile)
	}

	fmt.Printf("%s", buffer.String())
	return nil
}
