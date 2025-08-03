package main

import (
	"encoding/json"
	"os"
)

func (cmd *EnrichCmd) Run(cli *CLI) error {
	aweStore := NewAwesomeStore(cli.AwesomeFile)
	baseList, err := aweStore.Load()
	if err != nil {
		return err
	}

	awelist := NewAwesomeListManager(baseList)

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
