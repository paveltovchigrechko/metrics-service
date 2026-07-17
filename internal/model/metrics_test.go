package models

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }

func TestCreateMetricsPositive(t *testing.T) {
	positiveTestCases := []struct {
		testName string
		name     string
		mtype    string
		delta    int64
		value    float64
		expected *Metrics
	}{
		{
			testName: "Counter type metric with standard values",
			name:     "Some Counter",
			mtype:    Counter,
			delta:    99,
			value:    88.77,
			expected: &Metrics{
				ID:    "Some Counter",
				MType: Counter,
				Delta: int64Ptr(99),
				Value: nil,
			},
		},
		{
			testName: "Gauge type metric with standard values",
			name:     "Some Gauge",
			mtype:    Gauge,
			delta:    99,
			value:    88.77,
			expected: &Metrics{
				ID:    "Some Gauge",
				MType: Gauge,
				Delta: nil,
				Value: float64Ptr(88.77),
			},
		},
		{
			testName: "Counter type with zero value",
			name:     "Zero Counter",
			mtype:    Counter,
			delta:    0,
			value:    0,
			expected: &Metrics{
				ID:    "Zero Counter",
				MType: Counter,
				Delta: int64Ptr(0),
			},
		},
		{
			testName: "Gauge type with zero value",
			name:     "Zero Gauge",
			mtype:    Gauge,
			delta:    0,
			value:    0,
			expected: &Metrics{
				ID:    "Zero Gauge",
				MType: Gauge,
				Value: float64Ptr(0),
			},
		},
	}

	for _, tc := range positiveTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			result, err := CreateMetrics(tc.name, tc.mtype, tc.delta, tc.value)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expected.ID, result.ID)
			assert.Equal(t, tc.expected.MType, result.MType)

			if tc.mtype == Counter {
				assert.NotNil(t, result.Delta)
				assert.Equal(t, *tc.expected.Delta, *result.Delta)
				assert.Nil(t, result.Value)
			} else {
				assert.NotNil(t, result.Value)
				assert.Equal(t, *tc.expected.Value, *result.Value)
				assert.Nil(t, result.Delta)
			}
		})
	}
}

func TestCreateMetricsNegative(t *testing.T) {
	negativeTestCases := []struct {
		testName    string
		name        string
		mtype       string
		delta       int64
		value       float64
		expectedErr error
	}{
		{
			testName:    "Unknown metric type validation",
			name:        "Some Metric",
			mtype:       "unknown type",
			delta:       99,
			value:       88.77,
			expectedErr: ErrUnknownMetricsType,
		},
		{
			testName:    "Empty name string validation",
			name:        "",
			mtype:       Counter,
			delta:       99,
			value:       88.77,
			expectedErr: ErrEmptyMetricsID,
		},
		{
			testName:    "Whitespace-only name validation",
			name:        "   ",
			mtype:       Gauge,
			delta:       99,
			value:       88.77,
			expectedErr: ErrEmptyMetricsID,
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			result, err := CreateMetrics(tc.name, tc.mtype, tc.delta, tc.value)

			assert.Nil(t, result)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestRestoreMetrics(t *testing.T) {
	testCounter := Metrics{
		ID:    "PollCount",
		MType: Counter,
		Delta: int64Ptr(5),
	}
	testGauge := Metrics{
		ID:    "Alloc",
		MType: Gauge,
		Value: float64Ptr(124.50),
	}

	validMetricsList := []Metrics{testCounter, testGauge}
	validJSON, err := json.Marshal(validMetricsList)
	assert.NoError(t, err)

	type mockExpectation struct {
		metric      *Metrics
		returnError error
	}

	testCases := []struct {
		name         string
		fileContent  []byte
		useWrongPath bool
		mockReturns  []mockExpectation
		expectedErr  string
	}{
		{
			name:        "Successful restore of multiple metrics",
			fileContent: validJSON,
			mockReturns: []mockExpectation{
				{metric: &testCounter, returnError: nil},
				{metric: &testGauge, returnError: nil},
			},
			expectedErr: "",
		},
		{
			name:         "File path does not exist",
			useWrongPath: true,
			expectedErr:  "no such file or directory",
		},
		{
			name:        "Malformed or corrupt JSON payload",
			fileContent: []byte(`[{"id": "PollCount", "type": "counter", "delta": "invalid_type"`),
			expectedErr: "unexpected end of JSON",
		},
		{
			name:        "Storage returns an error on save",
			fileContent: validJSON,
			mockReturns: []mockExpectation{
				{metric: &testCounter, returnError: errors.New("database connection lost")},
			},
			expectedErr: "database connection lost",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "metrics_backup.json")

			if !tc.useWrongPath {
				err := os.WriteFile(filePath, tc.fileContent, 0644)
				assert.NoError(t, err)
			} else {
				filePath = filepath.Join(tmpDir, "non_existent_file.json")
			}

			mockStorage := new(MockStorage)
			for _, exp := range tc.mockReturns {
				mockStorage.On("RestoreMetrics", exp.metric).Return(exp.returnError).Once()
			}

			err := RestoreMetrics(filePath, mockStorage)

			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
