package model

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// FileStorage represents a local stored JSON-file with metrics.
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
	bytes, err := os.ReadFile(fs.path)
	if err != nil {
		return err
	}

	metrics := make([]Metrics, 0)
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		return err
	}

	if err = storage.SaveBatch(context.Background(), metrics); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) Save(storage Storage) error {
	// extract metrics from storage
	metrics, err := storage.GetAllMetrics(context.Background())
	if err != nil {
		return err
	}

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
