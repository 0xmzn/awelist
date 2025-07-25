package main

import (
	"encoding/json"
	"flag"
	"os"
)

func enrich(args []string) error {
	flag := flag.NewFlagSet("enrich", flag.ExitOnError)
	registerGlobalFlags(flag)
	flag.Parse(args)

	if flag.NArg() != 0 {
		return CliErrorf(nil, "enrich command doesn't accept any arguments`")
	}

	aweStore := NewAwesomeStore(_awesomeFile)
	err := aweStore.Load()
	if err != nil {
		return err
	}

	awelist := aweStore.GetManager()

	err = awelist.EnrichList()
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(awelist.EnrichedList, "", "\t")
	if err != nil {
		return CliErrorf(err, "failed to marshel json")
	}
	err = os.WriteFile("awesome-lock.json", []byte(jsonData), 0644)
	if err != nil {
		return CliErrorf(err, "failed to write json file")
	}

	return nil
}
