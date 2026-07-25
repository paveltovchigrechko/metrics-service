package model

import (
	"context"
	"database/sql"
)

type PostgresStorage struct {
	database *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		database: db,
	}
}

func (ps *PostgresStorage) GetMetrics(context.Context, string, string) (*Metrics, error) {
	return nil, nil
}

func (ps *PostgresStorage) SaveMetrics(context.Context, *Metrics) error {
	return nil
}

func (ps *PostgresStorage) RestoreMetrics(context.Context, *Metrics) error {
	return nil
}

func (ps *PostgresStorage) GetAllMetrics(context.Context) []Metrics {
	return nil
}
