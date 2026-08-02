package repository

import (
	"context"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

type Storage interface {
	GetMetrics(context.Context, string, string) (*model.Metrics, error)
	SaveMetrics(context.Context, *model.Metrics) error
	SaveBatch(context.Context, []model.Metrics) error
	GetAllMetrics(context.Context) ([]model.Metrics, error)
}
