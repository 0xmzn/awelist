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
	return &FileStore{
		yamlPath: yamlPath,
		jsonPath: jsonPath,
	}
}

func (fs *FileStore) LoadYAML() (types.Awesomelist, error) {
	data, err := os.ReadFile(fs.yamlPath)
	if err != nil {
		return nil, err
	}

	var categories types.Awesomelist
	if err := yaml.Unmarshal(data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (fs *FileStore) WriteYAML(cats types.Awesomelist) error {
	data, err := yaml.Marshal(cats)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.yamlPath, data, 0644)
}

func (fs *FileStore) WriteJSON(cats types.Awesomelist) error {
	data, err := json.MarshalIndent(cats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.jsonPath, data, 0644)
}
