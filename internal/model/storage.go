package models

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"text/tabwriter"
)

type MemStorage struct {
	Metrics map[string]*Metrics
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*Metrics),
	}
}

func (ms *MemStorage) GetMetrics(name string) (*Metrics, error) {
	metrics, ok := ms.Metrics[name]
	if !ok {
		return nil, errors.New("metrics not found")
	}

	return metrics, nil
}

func (ms *MemStorage) UpdateMetrics(m *Metrics) {
	switch m.MType {
	case Counter:
		newValue := *ms.Metrics[m.ID].Delta + *m.Delta
		ms.Metrics[m.ID].Delta = &newValue
	case Gauge:
		ms.Metrics[m.ID].Value = m.Value
	}
}

func (ms *MemStorage) ListMetrics(w http.ResponseWriter) {
	if len(ms.Metrics) == 0 {
		w.Write([]byte("Currently there are no metrics to display\n"))
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
