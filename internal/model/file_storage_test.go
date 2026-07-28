package model

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	t.Run("Successfully load metrics from file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		jsonData := `[{"id":"LoadCount","type":"counter","delta":10}]`
		_, err = tmpFile.WriteString(jsonData)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		fs, err := NewFileStorage(tmpFile.Name())
		require.NoError(t, err)

		mockStorage := new(MockStorage)

		mockStorage.On("SaveBatch", mock.Anything, mock.MatchedBy(func(metrics []Metrics) bool {
			return len(metrics) == 1 && metrics[0].ID == "LoadCount" && metrics[0].Delta != nil && *metrics[0].Delta == 10
		})).Return(nil)

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

	t.Run("Failed load due to storage error during save", func(t *testing.T) {
		// Create temporary file with valid metrics JSON
		tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		jsonData := `[{"id":"LoadCount","type":"counter","delta":10}]`
		_, err = tmpFile.WriteString(jsonData)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		fs, err := NewFileStorage(tmpFile.Name())
		require.NoError(t, err)
		mockStorage := new(MockStorage)

		// Configure mock to return an error when SaveBatch is invoked
		mockStorage.On("SaveBatch", mock.Anything, mock.Anything).Return(errors.New("storage error"))

		err = fs.Load(mockStorage)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "storage error")
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
		mockStorage.On("GetAllMetrics", mock.Anything).Return(testMetrics, nil).Once()

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

	t.Run("Failed save due to storage error fetching metrics", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "save.json")

		mockStorage := new(MockStorage)
		mockStorage.On("GetAllMetrics", mock.Anything).
			Return(nil, errors.New("storage failure")).Once()

		fs, err := NewFileStorage(filePath)
		assert.NoError(t, err)

		err = fs.Save(mockStorage)

		assert.Error(t, err)
		assert.Equal(t, "storage failure", err.Error())
		mockStorage.AssertExpectations(t)
	})

	t.Run("Failed save due to invalid directory path permissions", func(t *testing.T) {
		invalidPath := "/invalid_directory_abc123/backup.json"

		mockStorage := new(MockStorage)
		mockStorage.On("GetAllMetrics", mock.Anything).Return(testMetrics, nil).Once()

		fs, err := NewFileStorage(invalidPath)
		assert.NoError(t, err)

		err = fs.Save(mockStorage)

		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}
