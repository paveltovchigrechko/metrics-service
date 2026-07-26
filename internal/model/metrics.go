package model

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrUnknownMetricsType = errors.New("unknown metrics type")
	ErrEmptyMetricsID     = errors.New("metrics id is empty")
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func CreateMetrics(name, mtype string, delta int64, value float64) (*Metrics, error) {
	if strings.Trim(name, " ") == "" {
		return nil, ErrEmptyMetricsID
	}

	m := &Metrics{}
	m.ID = name
	m.MType = mtype

	switch mtype {
	case Counter:
		m.Delta = &delta
	case Gauge:
		m.Value = &value
	default:
		return nil, ErrUnknownMetricsType
	}

	return m, nil
}

func SaveMetrics(path string, storage Storage) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	metrics := make([]Metrics, 0)
	err = json.Unmarshal(bytes, &metrics)
	if err != nil {
		return err
	}

	for i := range metrics {
		err := storage.SaveMetrics(context.Background(), &metrics[i])
		if err != nil {
			return err // We may want to skip a malformed metrics and try to recover the next one.
		}
	}

	return nil
}
