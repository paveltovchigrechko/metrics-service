package model

import (
	"errors"
	"strings"
)

var (
	errMetricsNotFound      = errors.New("metrics not found")
	ErrEmptyMetricsID       = errors.New("metrics id is empty")
	ErrUnknownMetricsType   = errors.New("unknown metrics type")
	ErrDeltaIsNil           = errors.New("counter metrics delta is nil")
	ErrValueIsNil           = errors.New("gauge metrics value is nil")
	errDeltaAndValuePresent = errors.New("metrics has both detla and value fields")
	errMetricsIsNil         = errors.New("metrics is nil")
)

func validateName(name string) error {
	if strings.Trim(name, " ") == "" {
		return ErrEmptyMetricsID
	}

	return nil
}

func validateType(mtype string) error {
	if mtype != Counter && mtype != Gauge {
		return ErrUnknownMetricsType
	}

	return nil
}

func validateMetrics(m *Metrics) error {
	if m == nil {
		return errMetricsIsNil
	}
	if err := validateName(m.ID); err != nil {
		return err
	}

	if err := validateType(m.MType); err != nil {
		return err
	}

	if m.Delta != nil && m.Value != nil {
		return errDeltaAndValuePresent
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
