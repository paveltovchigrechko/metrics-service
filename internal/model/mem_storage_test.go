package model

import (
	"context"
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
				ID:    "Alloc",
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
			name:         "SaveMetrics returns error when counter delta is nil",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "NilCounter",
				MType: Counter,
				Delta: nil,
			},
			searchName:    "NilCounter",
			searchType:    Counter,
			expectedError: ErrDeltaIsNil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
		{
			name:         "SaveMetrics returns error when gauge value is nil",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "NilGauge",
				MType: Gauge,
				Value: nil,
			},
			searchName:    "NilGauge",
			searchType:    Gauge,
			expectedError: ErrValueIsNil,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				assert.Nil(t, res)
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
			name:         "SaveMetrics returns error on invalid type",
			initialState: map[string]*Metrics{},
			inputMetric: &Metrics{
				ID:    "BadMetric",
				MType: "unsupported_type",
			},
			searchName:    "BadMetric",
			searchType:    Counter,
			expectedError: ErrUnknownMetricsType,
			verifyState: func(t *testing.T, res *Metrics, s *MemStorage) {
				assert.Nil(t, res)
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

			var saveErr error
			if tc.inputMetric != nil {
				saveErr = s.SaveMetrics(context.Background(), tc.inputMetric)
				if tc.expectedError != nil && tc.expectedError != errMetricsNotFound {
					assert.ErrorIs(t, saveErr, tc.expectedError)
				} else {
					assert.NoError(t, saveErr)
				}
			}

			res, getErr := s.GetMetrics(context.Background(), tc.searchName, tc.searchType)
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

func TestMemStorage_GetAllMetrics(t *testing.T) {
	s := NewMemStorage()
	counter := Metrics{ID: "C1", MType: Counter, Delta: ptr(int64(10))}
	gauge := Metrics{ID: "G1", MType: Gauge, Value: ptr(12.3)}

	err := s.SaveMetrics(context.Background(), &counter)
	assert.NoError(t, err)
	err = s.SaveMetrics(context.Background(), &gauge)
	assert.NoError(t, err)

	metrics, err := s.GetAllMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
}

func ptr[T any](v T) *T {
	return &v
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetMetrics(ctx context.Context, name, mtype string) (*Metrics, error) {
	args := m.Called(ctx, name, mtype)
	if res, ok := args.Get(0).(*Metrics); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorage) SaveMetrics(ctx context.Context, metric *Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockStorage) GetAllMetrics(ctx context.Context) ([]Metrics, error) {
	args := m.Called(ctx)
	var res []Metrics
	if rf, ok := args.Get(0).([]Metrics); ok {
		res = rf
	}
	return res, args.Error(1)
}
