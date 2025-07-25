package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

type AwesomeStore struct {
	filename string
	manager  *AwesomeListManager
}

func NewAwesomeStore(filename string) *AwesomeStore {
	store := &AwesomeStore{
		filename: filename,
		manager:  NewAwesomeDataManager(make(baseAwesomelist, 0), nil),
	}

	return store
}

func (store *AwesomeStore) Load() error {
	fcontent, err := os.ReadFile(store.filename)
	if err != nil {
		return CliErrorf(err, "failed to read file %q", store.filename)
	}

	if err := yaml.UnmarshalWithOptions(fcontent, &store.manager.RawList, yaml.DisallowUnknownField()); err != nil {
		return CliErrorf(err, "failed to parse YAML data in %q", store.filename)
	}

	return nil
}

func (store *AwesomeStore) GetManager() *AwesomeListManager {
	return store.manager
}
