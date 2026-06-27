package models

type MemStorage struct {
	Metrics map[string]*Metrics
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*Metrics),
	}
}

func (ms *MemStorage) UpdateMetric(m *Metrics) {
	switch m.MType {
	case Counter:
		newValue := *ms.Metrics[m.ID].Delta + *m.Delta
		ms.Metrics[m.ID].Delta = &newValue
	case Gauge:
		ms.Metrics[m.ID].Value = m.Value
	}
}
