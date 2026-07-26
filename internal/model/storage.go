package model

import "context"

type Storage interface {
	GetMetrics(context.Context, string, string) (*Metrics, error)
	SaveMetrics(context.Context, *Metrics) error
	GetAllMetrics(context.Context) ([]Metrics, error)
}
