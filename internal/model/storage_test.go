package model

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	s := NewMemStorage()

	assert.NotNil(t, s)
	assert.NotNil(t, s.Metrics)
}

func TestMemStorage_SaveAndGetMetrics(t *testing.T) {
	// Assuming Counter and Gauge are defined string constants in your package (e.g., "counter", "gauge")
	testCases := []struct {
		name          string
		initialState  map[string]*Metrics
		inputMetric   *Metrics
		searchName    string
		searchType    string
		expectedError error
		verifyState   func(t *testing.T, res *Metrics, s *MemStorage)
	}{
		{
			name:         "SaveMetrics inserts a brand new counter",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: ptr(int64(5)),
			},
			searchName:    "PollCount", // Search purely by ID
			searchType:    Counter,
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(5), *res.Delta)
			},
		},
		{
			name:         "SaveMetrics inserts a brand new gauge",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "Alloc",
				MType: Gauge,
				Value: ptr(123.45),
			},
			searchName:    "Alloc", // Search purely by ID
			searchType:    Gauge,
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 123.45, *res.Value)
			},
		},
		{
			name: "SaveMetrics accumulates an existing counter",
			initialState: map[string]*Metrics{
				metricKey("PollCount", Counter): {ID: "PollCount", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: ptr(int64(5)),
			},
			searchName:    "PollCount",
			searchType:    Counter,
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(15), *res.Delta)
			},
		},
		{
			name: "SaveMetrics overrides an existing gauge",
			initialState: map[string]*Metrics{
				metricKey("Alloc", Gauge): {ID: "Alloc", MType: Gauge, Value: ptr(50.0)},
			},
			inputMetric: &Metrics{
				ID:    "Alloc",
				MType: Gauge,
				Value: ptr(99.9),
			},
			searchName:    "Alloc",
			searchType:    Gauge,
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 99.9, *res.Value)
			},
		},
		{
			name: "SaveMetrics returns error on invalid type when metric exists",
			initialState: map[string]*Metrics{
				metricKey("BadMetric", Counter): {ID: "BadMetric", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "BadMetric",
				MType: "unsupported_type",
			},
			searchName:    "BadMetric",
			searchType:    Counter,
			expectedError: ErrUnknownMetricsType,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, Counter, res.MType)
			},
		},
		{
			name:          "GetMetrics returns an error for non-existent metric",
			initialState:  map[string]*Metrics{},
			inputMetric:   nil,
			searchName:    "MissingMetric",
			searchType:    Counter,
			expectedError: errMetricsNotFound,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
		{
			name: "SaveMetrics allows two metrics with the same name but different types",
			initialState: map[string]*Metrics{
				metricKey("Alloc", Gauge): {ID: "Alloc", MType: Gauge, Value: ptr(55.5)},
			},
			inputMetric: &Metrics{
				ID:    "Alloc",
				MType: Counter,
				Delta: ptr(int64(10)),
			},
			searchName:    "Alloc",
			searchType:    Counter,
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(10), *res.Delta)

				gaugeRes, err := s.GetMetrics("Alloc", Gauge)
				require.NoError(t, err)
				require.NotNil(t, gaugeRes)
				assert.Equal(t, 55.5, *gaugeRes.Value)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMemStorage()
			for k, v := range tc.initialState {
				s.Metrics[k] = v
			}

			var saveErr error
			if tc.inputMetric != nil {
				saveErr = s.SaveMetrics(tc.inputMetric)
				if tc.expectedError != nil && tc.expectedError != errMetricsNotFound {
					assert.ErrorIs(t, saveErr, tc.expectedError)
				} else {
					assert.NoError(t, saveErr)
				}
			}

			res, getErr := s.GetMetrics(tc.searchName, tc.searchType)
			if tc.expectedError != nil && tc.expectedError == errMetricsNotFound {
				assert.ErrorIs(t, getErr, tc.expectedError)
			} else {
				if saveErr == nil {
					assert.NoError(t, getErr)
				}
			}

			tc.verifyState(t, res, s)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) RestoreMetrics(metric *Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockStorage) GetMetrics(name, mtype string) (*Metrics, error) {
	return nil, nil
}

func (m *MockStorage) ListMetrics(w io.Writer) {}

func (m *MockStorage) SaveMetrics(metric *Metrics) error {
	return nil
}

func (m *MockStorage) GetAllMetrics() []Metrics {
	args := m.Called()
	if rf, ok := args.Get(0).([]Metrics); ok {
		return rf
	}
	return nil
}

func TestRestoreMetricsMethod(t *testing.T) {
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
			name:        "Malformed JSON payload",
			fileContent: []byte(`[{"id": "PollCount", "type": "counter", "delta": "invalid_type"}`),
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
