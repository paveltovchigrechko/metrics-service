package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	s := NewStorage()

	assert.NotNil(t, s)
	assert.NotNil(t, s.Metrics)
}

func TestUpdateMetrics(t *testing.T) {
	s := NewStorage()
	// Populate storage
	s.Metrics["Some counter"] = &Metrics{
		ID:    "Some counter",
		MType: "counter",
		Delta: &[]int64{200}[0],
	}

	s.Metrics["Some gauge"] = &Metrics{
		ID:    "Some gauge",
		MType: "gauge",
		Value: &[]float64{1.0}[0],
	}

	testCases := []Metrics{
		{
			ID:    "Some counter",
			MType: "counter",
			Delta: &[]int64{100}[0],
		},
		{
			ID:    "Some gauge",
			MType: "gauge",
			Value: &[]float64{10.0}[0],
		},
	}

	for _, test := range testCases {
		var oldDelta int64
		if test.MType == "counter" {
			oldDelta = *s.Metrics[test.ID].Delta
		}

		s.UpdateMetrics(&test)
		switch test.MType {
		case "counter":
			sum := oldDelta + *test.Delta
			assert.Equal(t, *s.Metrics[test.ID].Delta, sum)
		case "gauge":
			assert.Equal(t, *test.Value, *s.Metrics[test.ID].Value)
		}

	}
}
