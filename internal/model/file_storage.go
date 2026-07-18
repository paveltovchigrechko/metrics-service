package model

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// Idea: make this Storage interface. Server would work with disk only, verey slow and fun.
type FileStorage struct {
	path string
}

var errEmptyFileStoragePath = errors.New("file strorage path is empty")

func NewFileStorage(path string) (*FileStorage, error) {
	// This check is for uniformity. ServerConfig validates the path value, but FileStorage doesn't know about it.
	if strings.Trim(path, " ") == "" {
		return nil, errEmptyFileStoragePath
	}

	return &FileStorage{
		path: path,
	}, nil
}

func (fs *FileStorage) Load(storage Storage) error {
	return RestoreMetrics(fs.path, storage)
}

func (fs *FileStorage) Save(storage Storage) error {
	// extract metrics from storage
	metrics := storage.GetAllMetrics()

	// encode to JSON
	encodedMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	// save into file
	err = os.WriteFile(fs.path, encodedMetrics, 0644)
	if err != nil {
		return err
	}

	return nil
}
