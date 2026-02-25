package store

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/0xmzn/awelist/internal/types"
	"github.com/goccy/go-yaml"
)

type FileStore struct {
	awesomeFilePath string
	lockFilePath    string
}

func New(yamlPath, jsonPath string) *FileStore {
	return &FileStore{yamlPath, jsonPath}
}

func (fs *FileStore) LoadAwesomeFile() (types.AwesomeList, error) {
	data, err := os.ReadFile(fs.awesomeFilePath)
	if err != nil {
		return nil, err
	}

	var list types.AwesomeList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (fs *FileStore) LoadLockFile() (*types.LockFile, error) {
	data, err := os.ReadFile(fs.lockFilePath)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)

	var lock types.LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

func (fs *FileStore) WriteAwesomeFile(list types.AwesomeList) error {
	data, err := yaml.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(fs.awesomeFilePath, data, 0644)
}

func (fs *FileStore) WriteLockFile(lock *types.LockFile) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.lockFilePath, data, 0644)
}
