package model

import (
	"errors"
	"strings"
)

var (
	ErrMetricsNotFound      = errors.New("metrics not found")
	ErrEmptyMetricsID       = errors.New("metrics id is empty")
	ErrUnknownMetricsType   = errors.New("unknown metrics type")
	ErrDeltaIsNil           = errors.New("counter metrics delta is nil")
	ErrValueIsNil           = errors.New("gauge metrics value is nil")
	ErrDeltaAndValuePresent = errors.New("metrics has both detla and value fields")
	ErrMetricsIsNil         = errors.New("metrics is nil")
)

func ValidateName(name string) error {
	if strings.Trim(name, " ") == "" {
		return ErrEmptyMetricsID
	}

	return nil
}

func ValidateType(mtype string) error {
	if mtype != Counter && mtype != Gauge {
		return ErrUnknownMetricsType
	}

	return nil
}

func ValidateMetrics(m *Metrics) error {
	if m == nil {
		return ErrMetricsIsNil
	}
	if err := ValidateName(m.ID); err != nil {
		return err
	}

	if err := ValidateType(m.MType); err != nil {
		return err
	}

	if m.Delta != nil && m.Value != nil {
		return ErrDeltaAndValuePresent
	}

	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaIsNil
		}
	case Gauge:
		if m.Value == nil {
			return ErrValueIsNil
		}
	}

	return nil
}
