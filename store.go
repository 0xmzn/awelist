package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

type AwesomeStore struct {
	filename string
}

func NewAwesomeStore(filename string) *AwesomeStore {
	return &AwesomeStore{filename: filename}
}

func (store *AwesomeStore) Load() (baseAwesomelist, error) {
	fcontent, err := os.ReadFile(store.filename)
	if err != nil {
		return nil, CliErrorf(err, "failed to read file %q", store.filename)
	}

	var data baseAwesomelist
	if err := yaml.UnmarshalWithOptions(fcontent, &data, yaml.DisallowUnknownField()); err != nil {
		return nil, CliErrorf(err, "failed to parse YAML data in %q", store.filename)
	}

	return data, nil
}

func (store *AwesomeStore) WriteYaml(list baseAwesomelist) error {
	yamlData, err := yaml.Marshal(list)
	if err != nil {
		return CliErrorf(err, "failed to marshel list to YAMl")
	}

	if err := os.WriteFile(store.filename, yamlData, 0644); err != nil {
		return CliErrorf(err, "failed to write YAML to file %q: %w", store.filename, err)
	}

	return nil
}
