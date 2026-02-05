package store

import (
	"encoding/json"
	"os"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/goccy/go-yaml"
)

type FileStore struct {
	yamlPath string
	jsonPath string
}

func New(yamlPath, jsonPath string) *FileStore {
	return &FileStore{yamlPath, jsonPath}
}

func (fs *FileStore) LoadYAML() (types.AwesomeList, error) {
	data, err := os.ReadFile(fs.yamlPath)
	if err != nil {
		return nil, err
	}

	var list types.AwesomeList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (fs *FileStore) LoadJson() (types.AwesomeList, error) {
	data, err := os.ReadFile(fs.jsonPath)
	if err != nil {
		return nil, err
	}

	var list types.AwesomeList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (fs *FileStore) WriteYAML(list types.AwesomeList) error {
	data, err := yaml.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.yamlPath, data, 0644)
}

func (fs *FileStore) WriteJSON(list types.AwesomeList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.jsonPath, data, 0644)
}
