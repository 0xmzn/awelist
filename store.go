package main

import (
	"github.com/goccy/go-yaml"
	"os"
)

type awesomeStore struct {
	filename string
	list     awesomeList
}

func NewAwesomeStore(filename string) *awesomeStore {
	store := &awesomeStore{
		filename: filename,
		list:     make(awesomeList, 0),
	}

	return store
}

func (store *awesomeStore) Load() error {
	fcontent, err := os.ReadFile(store.filename)
	if err != nil {
		return CliErrorf(err, "failed to read file %q", store.filename)
	}

	if err := yaml.UnmarshalWithOptions(fcontent, &store.list, yaml.DisallowUnknownField()); err != nil {
		return CliErrorf(err, "failed to parse YAML data in %q", store.filename)
	}

	return nil
}

func (store *awesomeStore) List() awesomeList {
	return store.list
}
