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
		return fmt.Errorf("generate command requires exactly one argument: <filename>")
	}

	tmpl_file := flag.Arg(0)

	awesomelist, err := loadFileIntoYaml(_awesomeFile)
	if err != nil {
		return err
	}

	tmpl_content, err := os.ReadFile(tmpl_file)
	if err != nil {
		return fmt.Errorf("failed to read file %q", tmpl_file)
	}

	tmpl, err := newTemplate(htmlOutput, tmpl_file, string(tmpl_content))
	if err != nil {
		return fmt.Errorf("invalid template syntax in %q", tmpl_file)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, awesomelist)
	if err != nil {
		return fmt.Errorf("failed while executing template %q", tmpl_file)
	}

	fmt.Printf("%s", buffer.String())
	return nil
}
