package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expectedError: errIncorrectMetricsType,
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
