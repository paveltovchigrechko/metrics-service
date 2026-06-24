package models

type MemStorage struct {
	Metrics map[string]*Metrics
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*Metrics),
	}
}

func (ms *MemStorage) UpdateMetric(m *Metrics, name string) {
	switch m.MType {
	case Counter:
		newValue := *ms.Metrics[name].Delta + *m.Delta
		ms.Metrics[name].Delta = &newValue
	case Gauge:
		ms.Metrics[name].Value = m.Value
	}
}
