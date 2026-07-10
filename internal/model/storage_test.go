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
				ID:    "counter:PollCount",
				MType: Counter,
				Delta: ptr(int64(5)),
			},
			searchName:    "PollCount",
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
				ID:    "gauge:Alloc",
				MType: Gauge,
				Value: ptr(123.45),
			},
			searchName:    "Alloc",
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
				"counter:PollCount": {ID: "counter:PollCount", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "counter:PollCount",
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
				"gauge:Alloc": {ID: "gauge:Alloc", MType: Gauge, Value: ptr(50.0)},
			},
			inputMetric: &Metrics{
				ID:    "gauge:Alloc",
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
				"unsupported_type:BadMetric": {ID: "unsupported_type:BadMetric", MType: Counter, Delta: ptr(int64(10))},
			},
			inputMetric: &Metrics{
				ID:    "BadMetric",
				MType: "unsupported_type",
			},
			searchName:    "BadMetric",
			searchType:    "unsupported_type",
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

			res, err := s.GetMetrics(tc.searchName, tc.searchType)
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
