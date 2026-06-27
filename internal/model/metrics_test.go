package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMetricsPositive(t *testing.T) {
	positiveTestCases := []struct {
		testName string
		name     string
		mtype    string
		delta    int64
		value    float64
		result   *Metrics
		err      error
	}{
		{
			"Counter type metric",
			"Some Counter",
			"counter",
			int64(99),
			float64(88.77),
			&Metrics{
				ID:    "Some Counter",
				MType: "counter",
				Delta: &[]int64{99}[0],
			},
			nil,
		},
		{
			"Gauge type metric",
			"Some Gauge",
			"gauge",
			int64(99),
			float64(88.77),
			&Metrics{
				ID:    "Some Counter",
				MType: "counter",
				Value: &[]float64{88.77}[0],
			},
			nil,
		},
	}

	for _, test := range positiveTestCases {
		result, err := CreateMetrics(test.name, test.mtype, test.delta, test.value)
		// check error first
		assert.Nil(t, err)

		// check the structure exists
		assert.NotNil(t, result)
		// check fields
		assert.Equal(t, result.ID, test.name)
		assert.Equal(t, result.MType, test.mtype)

		switch test.mtype {
		case "counter":
			assert.Equal(t, *result.Delta, test.delta)
		case "gauge":
			assert.Equal(t, *result.Value, test.value)
		}
	}
}

func TestCreateMetricsNegative(t *testing.T) {
	negativeTestCases := []struct {
		testName string
		name     string
		mtype    string
		delta    int64
		value    float64
		result   *Metrics
		err      error
	}{
		{
			"Counter type metric",
			"Some Counter",
			"unknown type",
			int64(99),
			float64(88.77),
			&Metrics{
				ID:    "Some Counter",
				MType: "counter",
				Delta: &[]int64{99}[0],
			},
			errors.New("Incorrect metrics type"),
		},
	}

	for _, test := range negativeTestCases {
		result, err := CreateMetrics(test.name, test.mtype, test.delta, test.value)
		// check error only
		assert.EqualError(t, err, test.err.Error())
		// check the structure doesn't exist
		assert.Nil(t, result)
	}
}
