package main

import (
	"encoding/json"
	"os"

	"github.com/goccy/go-yaml"
)

type AwesomeStore struct {
	yamlFile string
	jsonFile string
}

func NewAwesomeStore(filename string) *AwesomeStore {
	return &AwesomeStore{
		yamlFile: filename,
		jsonFile: "awesome-lock.json",
	}
}

func (store *AwesomeStore) LoadYAML() (baseAwesomelist, error) {
	fcontent, err := os.ReadFile(store.yamlFile)
	if err != nil {
		return nil, CliErrorf(err, "failed to read file %q", store.yamlFile)
	}

	var data baseAwesomelist
	if err := yaml.UnmarshalWithOptions(fcontent, &data, yaml.DisallowUnknownField()); err != nil {
		return nil, CliErrorf(err, "failed to parse YAML data in %q", store.yamlFile)
	}

	return data, nil
}

func (store *AwesomeStore) LoadJSON() (enrichedAwesomelist, error) {
	fcontent, err := os.ReadFile(store.jsonFile)
	if err != nil {
		return nil, CliErrorf(err, "failed to read file %q", store.jsonFile)
	}

	var data enrichedAwesomelist
	if err := json.Unmarshal(fcontent, &data); err != nil {
		return nil, CliErrorf(err, "failed to parse JSON data in %q", store.jsonFile)
	}

	return data, nil
}

func (store *AwesomeStore) WriteYAML(list baseAwesomelist) error {
	yamlData, err := yaml.Marshal(list)
	if err != nil {
		return CliErrorf(err, "failed to marshel list to YAMl")
	}

	if store.yamlFile == "" {
		return CliErrorf(nil, "yaml filename can't be empty")
	}

	if err := os.WriteFile(store.yamlFile, yamlData, 0644); err != nil {
		return CliErrorf(err, "failed to write YAML to file %q", store.yamlFile)
	}

	return nil
}

func (store *AwesomeStore) WriteJSON(list enrichedAwesomelist) error {
	jsonData, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return CliErrorf(err, "failed to marshel JSON")
	}

	if store.jsonFile == "" {
		return CliErrorf(nil, "JSON filename can't be empty")
	}

	err = os.WriteFile(store.jsonFile, []byte(jsonData), 0644)
	if err != nil {
		return CliErrorf(err, "failed to write JSON file %q", store.jsonFile)
	}

	return nil
}
