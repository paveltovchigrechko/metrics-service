package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewFileStorage(t *testing.T) {
	testCases := []struct {
		name string
		path string
		fs   *FileStorage
		err  error
	}{
		{
			name: "non-empty path is valid",
			path: "filename",
			fs:   &FileStorage{path: "filename"},
			err:  nil,
		},
		{
			name: "empty path is not valid",
			path: "",
			fs:   nil,
			err:  errEmptyFileStoragePath,
		},
		{
			name: "whitespace path is not valid",
			path: " ",
			fs:   nil,
			err:  errEmptyFileStoragePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFileStorage(tc.path)
			assert.Equal(t, tc.fs, fs)
			assert.Equal(t, tc.err, err)
		})
	}
}

func TestFileStorage_Load(t *testing.T) {
	testCounter := Metrics{
		ID:    "LoadCount",
		MType: Counter,
		Delta: int64Ptr(42),
	}
	validMetricsList := []Metrics{testCounter}
	validJSON, err := json.Marshal(validMetricsList)
	assert.NoError(t, err)

	t.Run("Successfully load metrics from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "backup.json")

		err := os.WriteFile(filePath, validJSON, 0644)
		assert.NoError(t, err)

		mockStorage := new(MockStorage)
		// Match any context.Context passed as the first parameter
		mockStorage.On("RestoreMetrics", mock.Anything, &testCounter).Return(nil).Once()

		fs, err := NewFileStorage(filePath)
		assert.NoError(t, err)

		err = fs.Load(mockStorage)

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Failed load due to missing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "non_existent.json")

		mockStorage := new(MockStorage)

		fs, err := NewFileStorage(filePath)
		assert.NoError(t, err)

		err = fs.Load(mockStorage)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
		mockStorage.AssertExpectations(t)
	})
}

func TestFileStorage_Save(t *testing.T) {
	testMetrics := []Metrics{
		{
			ID:    "Alloc",
			MType: Gauge,
			Value: float64Ptr(12.34),
		},
		{
			ID:    "PollCount",
			MType: Counter,
			Delta: int64Ptr(10),
		},
	}

	t.Run("Successfully extract and save metrics", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "save.json")

		mockStorage := new(MockStorage)
		// Match any context.Context passed as the first parameter
		mockStorage.On("GetAllMetrics", mock.Anything).Return(testMetrics).Once()

		fs, err := NewFileStorage(filePath)
		assert.NoError(t, err)

		err = fs.Save(mockStorage)
		assert.NoError(t, err)

		// Read back the file to verify its structure and content match
		bytes, err := os.ReadFile(filePath)
		assert.NoError(t, err)

		var savedMetrics []Metrics
		err = json.Unmarshal(bytes, &savedMetrics)
		assert.NoError(t, err)

		assert.Len(t, savedMetrics, 2)
		assert.Equal(t, testMetrics[0].ID, savedMetrics[0].ID)
		assert.Equal(t, *testMetrics[0].Value, *savedMetrics[0].Value)
		assert.Equal(t, testMetrics[1].ID, savedMetrics[1].ID)
		assert.Equal(t, *testMetrics[1].Delta, *savedMetrics[1].Delta)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Failed save due to invalid directory path permissions", func(t *testing.T) {
		invalidPath := "/invalid_directory_abc123/backup.json"

		mockStorage := new(MockStorage)
		mockStorage.On("GetAllMetrics", mock.Anything).Return(testMetrics).Once()

		fs, err := NewFileStorage(invalidPath)
		assert.NoError(t, err)

		err = fs.Save(mockStorage)

		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}
