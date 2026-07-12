package models

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type Storage interface {
	GetMetrics(string, string) (*Metrics, error)
	ListMetrics(io.Writer)
	SaveMetrics(*Metrics) error
}

type MemStorage struct {
	Metrics map[string]*Metrics
}

const noMetricsMessage = "Currently there are no metrics to display\n"

var (
	errMetricsNotFound = errors.New("metrics not found")
	errDeltaIsNil      = errors.New("counter metrics delta is nil")
	errValueIsNil      = errors.New("gauge metrics value is nil")
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

	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "Name\tValue\n")
	fmt.Fprintf(tw, "----\t-----\n")

	for _, name := range names {
		switch ms.Metrics[name].MType {
		case Counter:
			fmt.Fprintf(tw, "%s\t%d\n", ms.Metrics[name].ID, *ms.Metrics[name].Delta)
		case Gauge:
			fmt.Fprintf(tw, "%s\t%.2f\n", ms.Metrics[name].ID, *ms.Metrics[name].Value)
		}
	}
}

func metricKey(name, mtype string) string {
	return mtype + ":" + name
}
