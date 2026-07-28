package repository

import (
	"context"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/common"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
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
		initialState  map[string]*model.Metrics
		inputMetric   *model.Metrics
		searchName    string
		searchType    string
		expectedError error
		verifyState   func(t *testing.T, res *model.Metrics, s *MemStorage)
	}{
		{
			name:         "SaveMetrics inserts a brand new counter",
			initialState: map[string]*model.Metrics{},
			inputMetric: &model.Metrics{
				ID:    "PollCount",
				MType: model.Counter,
				Delta: common.Ptr(int64(5)),
			},
			searchName:    "PollCount",
			searchType:    model.Counter,
			expectedError: nil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(5), *res.Delta)
			},
		},
		{
			name:         "SaveMetrics inserts a brand new gauge",
			initialState: map[string]*model.Metrics{},
			inputMetric: &model.Metrics{
				ID:    "Alloc",
				MType: model.Gauge,
				Value: common.Ptr(123.45),
			},
			searchName:    "Alloc",
			searchType:    model.Gauge,
			expectedError: nil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 123.45, *res.Value)
			},
		},
		{
			name:         "SaveMetrics returns error when counter delta is nil",
			initialState: map[string]*model.Metrics{},
			inputMetric: &model.Metrics{
				ID:    "NilCounter",
				MType: model.Counter,
				Delta: nil,
			},
			searchName:    "NilCounter",
			searchType:    model.Counter,
			expectedError: model.ErrDeltaIsNil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
		{
			name:         "SaveMetrics returns error when gauge value is nil",
			initialState: map[string]*model.Metrics{},
			inputMetric: &model.Metrics{
				ID:    "NilGauge",
				MType: model.Gauge,
				Value: nil,
			},
			searchName:    "NilGauge",
			searchType:    model.Gauge,
			expectedError: model.ErrValueIsNil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
		{
			name: "SaveMetrics accumulates an existing counter",
			initialState: map[string]*model.Metrics{
				metricKey("PollCount", model.Counter): {ID: "PollCount", MType: model.Counter, Delta: common.Ptr(int64(10))},
			},
			inputMetric: &model.Metrics{
				ID:    "PollCount",
				MType: model.Counter,
				Delta: common.Ptr(int64(5)),
			},
			searchName:    "PollCount",
			searchType:    model.Counter,
			expectedError: nil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, int64(15), *res.Delta)
			},
		},
		{
			name: "SaveMetrics overrides an existing gauge",
			initialState: map[string]*model.Metrics{
				metricKey("Alloc", model.Gauge): {ID: "Alloc", MType: model.Gauge, Value: common.Ptr(50.0)},
			},
			inputMetric: &model.Metrics{
				ID:    "Alloc",
				MType: model.Gauge,
				Value: common.Ptr(99.9),
			},
			searchName:    "Alloc",
			searchType:    model.Gauge,
			expectedError: nil,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				require.NotNil(t, res)
				assert.Equal(t, 99.9, *res.Value)
			},
		},
		{
			name:         "SaveMetrics returns error on invalid type",
			initialState: map[string]*model.Metrics{},
			inputMetric: &model.Metrics{
				ID:    "BadMetric",
				MType: "unsupported_type",
			},
			searchName:    "BadMetric",
			searchType:    model.Counter,
			expectedError: model.ErrUnknownMetricsType,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
				assert.Nil(t, res)
			},
		},
		{
			name:          "GetMetrics returns an error for non-existent metric",
			initialState:  map[string]*model.Metrics{},
			inputMetric:   nil,
			searchName:    "MissingMetric",
			searchType:    model.Counter,
			expectedError: model.ErrMetricsNotFound,
			verifyState: func(t *testing.T, res *model.Metrics, s *MemStorage) {
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
				if tc.expectedError != nil && tc.expectedError != model.ErrMetricsNotFound {
					assert.ErrorIs(t, saveErr, tc.expectedError)
				} else {
					assert.NoError(t, saveErr)
				}
			}

			res, getErr := s.GetMetrics(context.Background(), tc.searchName, tc.searchType)
			if tc.expectedError != nil && tc.expectedError == model.ErrMetricsNotFound {
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
	counter := model.Metrics{ID: "C1", MType: model.Counter, Delta: common.Ptr(int64(10))}
	gauge := model.Metrics{ID: "G1", MType: model.Gauge, Value: common.Ptr(12.3)}

	err := s.SaveMetrics(context.Background(), &counter)
	assert.NoError(t, err)
	err = s.SaveMetrics(context.Background(), &gauge)
	assert.NoError(t, err)

	metrics, err := s.GetAllMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetMetrics(ctx context.Context, name, mtype string) (*model.Metrics, error) {
	args := m.Called(ctx, name, mtype)
	if res, ok := args.Get(0).(*model.Metrics); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorage) SaveMetrics(ctx context.Context, metric *model.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockStorage) GetAllMetrics(ctx context.Context) ([]model.Metrics, error) {
	args := m.Called(ctx)
	var res []model.Metrics
	if rf, ok := args.Get(0).([]model.Metrics); ok {
		res = rf
	}
	return res, args.Error(1)
}

func (m *MockStorage) SaveBatch(ctx context.Context, metrics []model.Metrics) error {
	args := m.Called(ctx, metrics)

	return args.Error(0)
}

func TestMemStorage_SaveBatch(t *testing.T) {
	t.Run("Successfully saves a batch of mixed valid metrics", func(t *testing.T) {
		s := NewMemStorage()

		batch := []model.Metrics{
			{ID: "Requests", MType: model.Counter, Delta: common.Ptr(int64(10))},
			{ID: "Temperature", MType: model.Gauge, Value: common.Ptr(36.6)},
		}

		err := s.SaveBatch(context.Background(), batch)
		assert.NoError(t, err)

		// Verify elements were successfully inserted
		resCounter, err := s.GetMetrics(context.Background(), "Requests", model.Counter)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *resCounter.Delta)

		resGauge, err := s.GetMetrics(context.Background(), "Temperature", model.Gauge)
		assert.NoError(t, err)
		assert.Equal(t, 36.6, *resGauge.Value)
	})

	t.Run("Batch fails validation and aborts completely", func(t *testing.T) {
		s := NewMemStorage()

		// Pre-populate an initial metric
		initial := &model.Metrics{ID: "Requests", MType: model.Counter, Delta: common.Ptr(int64(10))}
		err := s.SaveMetrics(context.Background(), initial)
		require.NoError(t, err)

		// Batch contains a valid metric followed by an invalid one (nil delta for counter)
		batch := []model.Metrics{
			{ID: "Requests", MType: model.Counter, Delta: common.Ptr(int64(5))}, // Would accumulate to 15 if applied
			{ID: "BadCounter", MType: model.Counter, Delta: nil},                // Invalid
		}

		err = s.SaveBatch(context.Background(), batch)
		assert.ErrorIs(t, err, model.ErrDeltaIsNil)

		// Verify atomicity: because it failed, the first metric should NOT have been updated/accumulated
		resCounter, err := s.GetMetrics(context.Background(), "Requests", model.Counter)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *resCounter.Delta, "Batch should not apply partial changes on failure")
	})
}
