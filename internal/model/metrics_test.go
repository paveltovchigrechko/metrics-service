package model

import (
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/common"
	"github.com/stretchr/testify/assert"
)

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
				Delta: common.Int64Ptr(99),
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
				Value: common.Float64Ptr(88.77),
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
				Delta: common.Int64Ptr(0),
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
				Value: common.Float64Ptr(0),
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
