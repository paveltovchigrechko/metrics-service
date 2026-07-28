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
			expectedError: ErMetricsNotFound,
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
				if tc.expectedError != nil && tc.expectedError != ErMetricsNotFound {
					assert.ErrorIs(t, saveErr, tc.expectedError)
				} else {
					assert.NoError(t, saveErr)
				}
			}

			res, getErr := s.GetMetrics(context.Background(), tc.searchName, tc.searchType)
			if tc.expectedError != nil && tc.expectedError == ErMetricsNotFound {
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

func (m *MockStorage) SaveBatch(ctx context.Context, metrics []Metrics) error {
	args := m.Called(ctx, metrics)

	return args.Error(0)
}

func TestMemStorage_SaveBatch(t *testing.T) {
	t.Run("Successfully saves a batch of mixed valid metrics", func(t *testing.T) {
		s := NewMemStorage()

		batch := []Metrics{
			{ID: "Requests", MType: Counter, Delta: ptr(int64(10))},
			{ID: "Temperature", MType: Gauge, Value: ptr(36.6)},
		}

		err := s.SaveBatch(context.Background(), batch)
		assert.NoError(t, err)

		// Verify elements were successfully inserted
		resCounter, err := s.GetMetrics(context.Background(), "Requests", Counter)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *resCounter.Delta)

		resGauge, err := s.GetMetrics(context.Background(), "Temperature", Gauge)
		assert.NoError(t, err)
		assert.Equal(t, 36.6, *resGauge.Value)
	})

	t.Run("Batch fails validation and aborts completely", func(t *testing.T) {
		s := NewMemStorage()

		// Pre-populate an initial metric
		initial := &Metrics{ID: "Requests", MType: Counter, Delta: ptr(int64(10))}
		err := s.SaveMetrics(context.Background(), initial)
		require.NoError(t, err)

		// Batch contains a valid metric followed by an invalid one (nil delta for counter)
		batch := []Metrics{
			{ID: "Requests", MType: Counter, Delta: ptr(int64(5))}, // Would accumulate to 15 if applied
			{ID: "BadCounter", MType: Counter, Delta: nil},         // Invalid
		}

		err = s.SaveBatch(context.Background(), batch)
		assert.ErrorIs(t, err, ErrDeltaIsNil)

		// Verify atomicity: because it failed, the first metric should NOT have been updated/accumulated
		resCounter, err := s.GetMetrics(context.Background(), "Requests", Counter)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *resCounter.Delta, "Batch should not apply partial changes on failure")
	})
}
