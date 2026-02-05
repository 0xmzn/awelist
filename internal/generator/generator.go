package generator

import (
	"fmt"
	html "html/template"
	"io"
	"os"
	text "text/template"

	"github.com/0xmzn/awelist/internal/types"
)

func GenerateOutput(w io.Writer, filename string, isHTML bool, data types.AwesomeList) error {
	tmplContent, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	tmpl, err := newTemplate(isHTML, filename, string(tmplContent))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	return nil
}

type templateExecutor interface {
	Execute(wr io.Writer, data any) error
}

func newTemplate(isHTML bool, name string, content string) (templateExecutor, error) {
	if isHTML {
		return html.New(name).Parse(content)
	}
	return text.New(name).Parse(content)
}
