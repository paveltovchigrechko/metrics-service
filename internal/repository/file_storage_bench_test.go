package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

func BenchmarkFileStorage_Save(b *testing.B) {
	for _, n := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			storage := NewMemStorage()
			metrics := make([]model.Metrics, 0, n)
			for i := 0; i < n; i++ {
				if i%4 == 3 {
					d := int64(i)
					metrics = append(metrics, model.Metrics{ID: fmt.Sprintf("counter_%05d", i), MType: model.Counter, Delta: &d})
				} else {
					v := float64(i) * 1.5
					metrics = append(metrics, model.Metrics{ID: fmt.Sprintf("gauge_%05d", i), MType: model.Gauge, Value: &v})
				}
			}
			if err := storage.SaveBatch(context.Background(), metrics); err != nil {
				b.Fatal(err)
			}

			fs, err := NewFileStorage(filepath.Join(b.TempDir(), "metrics.json"))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			for b.Loop() {
				if err := fs.Save(storage); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMemStorage_GetAllMetrics(b *testing.B) {
	for _, n := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("metrics=%d", n), func(b *testing.B) {
			storage := NewMemStorage()
			metrics := make([]model.Metrics, 0, n)
			for i := 0; i < n; i++ {
				v := float64(i)
				metrics = append(metrics, model.Metrics{ID: fmt.Sprintf("gauge_%05d", i), MType: model.Gauge, Value: &v})
			}
			if err := storage.SaveBatch(context.Background(), metrics); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			for b.Loop() {
				if _, err := storage.GetAllMetrics(context.Background()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
