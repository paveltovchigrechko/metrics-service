package models

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Storage interface {
	GetMetrics(string, string) (*Metrics, error)
	ListMetrics(io.Writer)
	SaveMetrics(*Metrics) error
	RestoreMetrics(m *Metrics) error
}

type MemStorage struct {
	Metrics map[string]*Metrics
}

const noMetricsMessage = "<p>Currently there are no metrics to display</p>"

var (
	errMetricsNotFound = errors.New("metrics not found")
	errDeltaIsNil      = errors.New("counter metrics delta is nil")
	errValueIsNil      = errors.New("gauge metrics value is nil")
	errEmptyID         = errors.New("metrics ID is empty")
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*Metrics),
	}
}

func (ms *MemStorage) GetMetrics(name, mtype string) (*Metrics, error) {
	key := metricKey(name, mtype)
	metrics, ok := ms.Metrics[key]
	if !ok {
		return nil, errMetricsNotFound
	}

	return metrics, nil
}

func (ms *MemStorage) SaveMetrics(m *Metrics) error {
	if m.MType != Counter && m.MType != Gauge {
		return errIncorrectMetricsType
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
			return errDeltaIsNil
		}
		newValue := *current.Delta + *m.Delta
		current.Delta = &newValue
	case Gauge:
		if m.Value == nil {
			return errValueIsNil
		}
		current.Value = m.Value
	default:
		return errIncorrectMetricsType
	}
	return nil
}

func (ms *MemStorage) ListMetrics(w io.Writer) {
	if len(ms.Metrics) == 0 {
		w.Write([]byte(noMetricsMessage))
		return
	}

	names := make([]string, 0, len(ms.Metrics))
	for name := range ms.Metrics {
		names = append(names, name)
	}

	sort.Strings(names)
	fmt.Fprint(w, "<p>ID&nbsp;&nbsp;&nbsp;&nbsp;Value</p>")
	for _, name := range names {
		switch ms.Metrics[name].MType {
		case Counter:
			fmt.Fprintf(w, "<p>%s&nbsp;&nbsp;&nbsp;&nbsp;%d</p>", ms.Metrics[name].ID, *ms.Metrics[name].Delta)
		case Gauge:
			fmt.Fprintf(w, "<p>%s&nbsp;&nbsp;&nbsp;&nbsp;%.2f</p>", ms.Metrics[name].ID, *ms.Metrics[name].Value)
		}
	}
}

// RestoreMetrics
func (ms *MemStorage) RestoreMetrics(m *Metrics) error {
	if m.MType != Counter && m.MType != Gauge {
		return errIncorrectMetricsType
	}
	if m.MType == Counter && m.Delta == nil {
		return errDeltaIsNil
	}
	if m.MType == Gauge && m.Value == nil {
		return errValueIsNil
	}
	if strings.Trim(m.ID, " ") == "" {
		return errEmptyID
	}

	key := metricKey(m.ID, m.MType)
	ms.Metrics[key] = m

	return nil
}

func metricKey(name, mtype string) string {
	return mtype + ":" + name
}
