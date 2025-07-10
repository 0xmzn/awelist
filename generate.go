package main

import (
	"fmt"
	"flag"
)

func generate(args []string) error {
	flag := flag.NewFlagSet("generate", flag.ExitOnError)
	var htmlOutput bool
	flag.BoolVar(&htmlOutput, "html", false, "output html")
	registerGlobalFlags(flag)
	flag.Parse(args)

	if flag.NArg() != 1 {
        return fmt.Errorf("generate command requires exactly one argument: <filename>")
	}

	awesomelist, err := loadFileIntoYaml(_awesomeFile)
	if err != nil {
		return err
	}
	fmt.Println(awesomelist)
	return nil
}