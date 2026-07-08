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
	testCases := []struct {
		name          string
		initialState  map[string]*Metrics                             // The MemStorage state before the test
		inputMetric   *Metrics                                        // The metric passed to SaveMetrics
		searchID      string                                          // The ID passed to GetMetrics afterwards
		expectedError error                                           // Error expected from SaveMetrics or GetMetrics
		verifyState   func(t *testing.T, res *Metrics, s *MemStorage) // Custom validations for each test case
	}{
		{
			name:         "SaveMetrics inserts a brand new counter",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: ptr(int64(5)),
			},
			searchID:      "PollCount",
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
			searchID:      "Alloc",
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 123.45, *res.Value)
			},
		},
		{
			name: "SaveMetrics accumulates an existing counter",
			initialState: map[string]*Metrics{
				"PollCount": {ID: "PollCount", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: ptr(int64(5)),
			},
			searchID:      "PollCount",
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(15), *res.Delta)
			},
		},
		{
			name: "SaveMetrics overrides an existing gauge",
			initialState: map[string]*Metrics{
				"Alloc": {ID: "Alloc", MType: Gauge, Value: ptr(50.0)},
			},
			inputMetric: &Metrics{
				ID:    "Alloc",
				MType: Gauge,
				Value: ptr(99.9),
			},
			searchID:      "Alloc",
			expectedError: nil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 99.9, *res.Value)
			},
		},
		{
			name: "SaveMetrics returns error on invalid type when metric exists",
			initialState: map[string]*Metrics{
				"BadMetric": {ID: "BadMetric", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "BadMetric",
				MType: "unsupported_type",
			},
			searchID:      "BadMetric",
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
			searchID:      "MissingMetric",
			expectedError: errMetricsNotFound,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMemStorage()
			for k, v := range tc.initialState {
				s.Metrics[k] = v
			}

			if tc.inputMetric != nil {
				err := s.SaveMetrics(tc.inputMetric)
				if tc.expectedError != nil && tc.expectedError == errIncorrectMetricsType {
					assert.ErrorIs(t, err, tc.expectedError)
				} else {
					assert.NoError(t, err)
				}
			}

			res, err := s.GetMetrics(tc.searchID)
			if tc.expectedError != nil && tc.expectedError == errMetricsNotFound {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				if tc.expectedError != errIncorrectMetricsType {
					assert.NoError(t, err)
				}
			}

			tc.verifyState(t, res, s)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
