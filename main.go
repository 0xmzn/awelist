package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func registerGlobalFlags(fset *flag.FlagSet) {
	flag.VisitAll(func(f *flag.Flag) {
		fset.Var(f.Value, f.Name, f.Usage)
	})
}

func loadFileIntoYaml(path string) (awesomeList, error) {
	var awesomelist awesomeList

	fcontent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q", path)
	}

	if err := yaml.Unmarshal(fcontent, &awesomelist); err != nil {
		return nil, fmt.Errorf("failed to parse YAML data in %q", path)
	}

	return awesomelist, nil
}

var (
	_awesomeFile string
)

func init() {
	flag.StringVar(&_awesomeFile, "f", "", "path to awesome file")
}

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: awelist [options] <command> [command-options] [arguments]")
		fmt.Fprintln(os.Stderr, "\nMain Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nCommands:")
		fmt.Fprintln(os.Stderr, "  generate    generate file from template")
	}

	flag.Parse()

	if _awesomeFile == "" {
		_awesomeFile = "awesome.yaml"
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	subcmd, args := flag.Args()[0], flag.Args()[1:]

	switch subcmd {
	case "generate":
		if err := generate(args); err != nil {
			log.Fatalf("awelist: %s", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "awelist: Unknown command '%s'\n", subcmd)
		fmt.Fprintln(os.Stderr, "Try 'awelist -help' for more information.")
		os.Exit(1)
	}
}
