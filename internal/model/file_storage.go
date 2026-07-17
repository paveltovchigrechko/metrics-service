package models

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	Path string
}

func (fs *FileStorage) Load(storage Storage) error {
	return RestoreMetrics(fs.Path, storage)
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
	err = os.WriteFile(fs.Path, encodedMetrics, 0644)
	if err != nil {
		return err
	}

	return nil
}
