package model

import (
	"context"
	"database/sql"
	"errors"
)

type PostgresStorage struct {
	database *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		database: db,
	}
}

func (ps *PostgresStorage) GetMetrics(ctx context.Context, name string, mType string) (*Metrics, error) {
	if mType != Counter && mType != Gauge {
		return nil, ErrUnknownMetricsType
	}
	row := ps.database.QueryRowContext(ctx, "SELECT id, mtype, delta, value FROM metrics WHERE id = $1 AND mtype = $2", name, mType)

	var m Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(&m.ID, &m.MType, &delta, &value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errMetricsNotFound
	} else if err != nil {
		return nil, err
	}

	switch mType {
	case Counter:
		if delta.Valid {
			m.Delta = &delta.Int64
		} else {
			return nil, ErrDeltaIsNil
		}
	case Gauge:
		if value.Valid {
			m.Value = &value.Float64
		} else {
			return nil, ErrValueIsNil
		}
	}

	return &m, nil
}

func (ps *PostgresStorage) SaveMetrics(ctx context.Context, m *Metrics) error {
	if m == nil || m.ID == "" {
		return ErrEmptyMetricsID
	}
	if m.MType != Counter && m.MType != Gauge {
		return ErrUnknownMetricsType
	}
	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrDeltaIsNil
		}
	case Gauge:
		if m.Value == nil {
			return ErrValueIsNil
		}
	}

	// Metric exists in the database
	switch m.MType {
	case Counter:
		_, err := ps.database.ExecContext(ctx, "INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3) ON CONFLICT (id, mtype) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta", m.ID, m.MType, *m.Delta)
		if err != nil {
			return err
		}
	case Gauge:
		_, err := ps.database.ExecContext(ctx, "INSERT INTO metrics (id, mtype, value) VALUES ($1, $2, $3) ON CONFLICT (id, mtype) DO UPDATE SET value = $3", m.ID, m.MType, *m.Value)
		if err != nil {
			return err
		}
	default:
		return ErrUnknownMetricsType
	}

	return nil
}

func (ps *PostgresStorage) GetAllMetrics(ctx context.Context) ([]Metrics, error) {
	metrics := make([]Metrics, 0) // Add some capacity here?

	rows, err := ps.database.QueryContext(ctx, "SELECT id, mtype, delta, value FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		err := rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, err
		}

		switch m.MType {
		case Counter:
			if delta.Valid {
				m.Delta = &delta.Int64
			} else {
				return nil, ErrDeltaIsNil
			}
		case Gauge:
			if value.Valid {
				m.Value = &value.Float64
			} else {
				return nil, ErrValueIsNil
			}
		default:
			return nil, ErrUnknownMetricsType
		}

		metrics = append(metrics, m) // What if the slice is getting quite big and holds everything in memory?
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
