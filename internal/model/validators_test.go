package model

import (
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "Valid name",
			input:         "Alloc",
			expectedError: nil,
		},
		{
			name:          "Empty name",
			input:         "",
			expectedError: ErrEmptyMetricsID,
		},
		{
			name:          "Whitespace-only name",
			input:         "   ",
			expectedError: ErrEmptyMetricsID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateName(tc.input)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "Valid Counter type",
			input:         Counter,
			expectedError: nil,
		},
		{
			name:          "Valid Gauge type",
			input:         Gauge,
			expectedError: nil,
		},
		{
			name:          "Unknown metric type",
			input:         "histogram",
			expectedError: ErrUnknownMetricsType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateType(tc.input)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestValidateMetrics(t *testing.T) {
	tests := []struct {
		name          string
		metric        *Metrics
		expectedError error
	}{
		{
			name:          "Nil metric pointer",
			metric:        nil,
			expectedError: ErrMetricsIsNil,
		},
		{
			name: "Empty metric ID",
			metric: &Metrics{
				ID:    "",
				MType: Counter,
				Delta: common.Ptr(int64(10)),
			},
			expectedError: ErrEmptyMetricsID,
		},
		{
			name: "Unknown metric type",
			metric: &Metrics{
				ID:    "CpuUsage",
				MType: "invalid_type",
			},
			expectedError: ErrUnknownMetricsType,
		},
		{
			name: "Both Delta and Value present",
			metric: &Metrics{
				ID:    "InvalidMetric",
				MType: Counter,
				Delta: common.Ptr(int64(10)),
				Value: common.Ptr(5.5),
			},
			expectedError: ErrDeltaAndValuePresent,
		},
		{
			name: "Counter with nil Delta",
			metric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: nil,
			},
			expectedError: ErrDeltaIsNil,
		},
		{
			name: "Valid Counter metric",
			metric: &Metrics{
				ID:    "PollCount",
				MType: Counter,
				Delta: common.Ptr(int64(1)),
			},
			expectedError: nil,
		},
		{
			name: "Gauge with nil Value",
			metric: &Metrics{
				ID:    "Alloc",
				MType: Gauge,
				Value: nil,
			},
			expectedError: ErrValueIsNil,
		},
		{
			name: "Valid Gauge metric",
			metric: &Metrics{
				ID:    "Alloc",
				MType: Gauge,
				Value: common.Ptr(100.5),
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMetrics(tc.metric)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
