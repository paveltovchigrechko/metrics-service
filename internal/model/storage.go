package model

import (
	"errors"
	"strings"
	"sync"
)

type Storage interface {
	GetMetrics(string, string) (*Metrics, error)
	SaveMetrics(*Metrics) error
	RestoreMetrics(m *Metrics) error
	GetAllMetrics() []Metrics
}

type MemStorage struct {
	mu sync.RWMutex

	Metrics map[string]*Metrics
}

var (
	errMetricsNotFound = errors.New("metrics not found")
	ErrDeltaIsNil      = errors.New("counter metrics delta is nil")
	ErrValueIsNil      = errors.New("gauge metrics value is nil")
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*Metrics),
	}
}

func (ms *MemStorage) GetMetrics(name, mtype string) (*Metrics, error) {
	// Protect storage from overriding. Not necessary for current implementation.
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := metricKey(name, mtype)
	metrics, ok := ms.Metrics[key]
	if !ok {
		return nil, errMetricsNotFound
	}

	return metrics, nil
}

func (ms *MemStorage) SaveMetrics(m *Metrics) error {
	// Protect storage from overriding. Not necessary for current implementation.
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if m.MType != Counter && m.MType != Gauge {
		return ErrUnknownMetricsType
	}

	key := metricKey(m.ID, m.MType)
	current, existing := ms.Metrics[key]
	if !existing {
		ms.Metrics[key] = m
		return nil
	}

	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaIsNil
		}
		newValue := *current.Delta + *m.Delta
		current.Delta = &newValue
	case Gauge:
		if m.Value == nil {
			return ErrValueIsNil
		}
		current.Value = m.Value
	default:
		return ErrUnknownMetricsType
	}
	return nil
}

// RestoreMetrics
func (ms *MemStorage) RestoreMetrics(m *Metrics) error {
	if m.MType != Counter && m.MType != Gauge {
		return ErrUnknownMetricsType
	}
	if m.MType == Counter && m.Delta == nil {
		return ErrDeltaIsNil
	}
	if m.MType == Gauge && m.Value == nil {
		return ErrValueIsNil
	}
	if strings.Trim(m.ID, " ") == "" {
		return ErrEmptyMetricsID
	}

	key := metricKey(m.ID, m.MType)
	ms.Metrics[key] = m

	return nil
}

func (ms *MemStorage) GetAllMetrics() []Metrics {
	// This method is called by server, so we must protect the storage for reading, because a handler might update the storage when server reads it.
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make([]Metrics, 0)
	for _, m := range ms.Metrics {
		metrics = append(metrics, *m)
	}

	return metrics
}

func metricKey(name, mtype string) string {
	return mtype + ":" + name
}
