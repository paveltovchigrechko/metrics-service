package models

import (
	"errors"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
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

func CreateMetrics(mtype string, delta int64, value float64) (*Metrics, error) {
	m := &Metrics{}

	switch mtype {
	case Counter:
		m.Delta = &delta
	case Gauge:
		m.Value = &value
	default:
		return nil, errors.New("Incorrect metrics type")
	}

	return m, nil
}
