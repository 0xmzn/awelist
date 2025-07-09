package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func loadFileIntoYaml(path string) (awesomeList, error) {
	var awesomelist awesomeList

	fcontent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(fcontent, &awesomelist); err != nil {
		return nil, fmt.Errorf("failed to parse YAML data in %s", path)
	}

	return awesomelist, nil
}

func main() {
	log.SetFlags(0)

	var awesomeFile string
	flag.StringVar(&awesomeFile, "f", "", "path to awesome file")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: awelist [options] <command> [command-options] [arguments]")
		fmt.Fprintln(os.Stderr, "\nMain Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nCommands:")
		fmt.Fprintln(os.Stderr, "  generate    generate file from template")
	}

	flag.Parse()

	if awesomeFile == "" {
		awesomeFile = "awesome.yaml"
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	switch flag.Args()[0] {
	case "generate":
		awesomelist, err := loadFileIntoYaml(awesomeFile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(awesomelist)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\n", flag.Args()[0])
		flag.Usage()
		os.Exit(1)
	}
}
