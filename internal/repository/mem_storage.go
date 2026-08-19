package repository

import (
	"context"
	"sync"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

type MemStorage struct {
	mu      sync.RWMutex
	Metrics map[string]*model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*model.Metrics),
	}
}

func (ms *MemStorage) GetMetrics(ctx context.Context, name, mtype string) (*model.Metrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if err := model.ValidateName(name); err != nil {
		return nil, err
	}
	if err := model.ValidateType(mtype); err != nil {
		return nil, err
	}

	key := metricKey(name, mtype)
	metrics, ok := ms.Metrics[key]
	if !ok {
		return nil, model.ErrMetricsNotFound
	}

	return metrics, nil
}

func (ms *MemStorage) SaveMetrics(ctx context.Context, m *model.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if err := model.ValidateMetrics(m); err != nil {
		return err
	}

	return ms.saveMetricsLocked(m)
}

func (ms *MemStorage) GetAllMetrics(ctx context.Context) ([]model.Metrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make([]model.Metrics, 0, len(ms.Metrics))
	for _, m := range ms.Metrics {
		metrics = append(metrics, *m)
	}

	return metrics, nil
}

func (ms *MemStorage) SaveBatch(ctx context.Context, metrics []model.Metrics) error {
	for i := range metrics {
		if err := model.ValidateMetrics(&metrics[i]); err != nil {
			return err
		}
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, m := range metrics {
		if err := ms.saveMetricsLocked(&m); err != nil {
			return err
		}
	}

	return nil
}

func metricKey(name, mtype string) string {
	return mtype + ":" + name
}

func (ms *MemStorage) saveMetricsLocked(m *model.Metrics) error {
	key := metricKey(m.ID, m.MType)
	current, existing := ms.Metrics[key]
	if !existing {
		ms.Metrics[key] = m
		return nil
	}

	switch m.MType {
	case model.Counter:
		newValue := *current.Delta + *m.Delta
		current.Delta = &newValue
	case model.Gauge:
		current.Value = m.Value
	}

	return nil
}
