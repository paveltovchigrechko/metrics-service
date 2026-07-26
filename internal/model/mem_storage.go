package model

import (
	"context"
	"errors"
	"sync"
)

type MemStorage struct {
	mu      sync.RWMutex
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

func (ms *MemStorage) GetMetrics(ctx context.Context, name, mtype string) (*Metrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := metricKey(name, mtype)
	metrics, ok := ms.Metrics[key]
	if !ok {
		return nil, errMetricsNotFound
	}

	return metrics, nil
}

func (ms *MemStorage) SaveMetrics(ctx context.Context, m *Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaIsNil
		}
	case Gauge:
		if m.Value == nil {
			return ErrValueIsNil
		}
	default:
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
		newValue := *current.Delta + *m.Delta
		current.Delta = &newValue
	case Gauge:
		current.Value = m.Value
	}

	return nil
}

func (ms *MemStorage) GetAllMetrics(ctx context.Context) ([]Metrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make([]Metrics, 0, len(ms.Metrics))
	for _, m := range ms.Metrics {
		metrics = append(metrics, *m)
	}

	return metrics, nil
}

func metricKey(name, mtype string) string {
	return mtype + ":" + name
}
