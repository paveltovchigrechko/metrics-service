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
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateType(mType); err != nil {
		return nil, err
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
	if err := validateMetrics(m); err != nil {
		return err
	}

	if err := ps.saveMetricsWithExecutor(ctx, ps.database, m); err != nil {
		return err
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
		// A: should be fine before 10^6 in-memory metrics.
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (ps *PostgresStorage) SaveBatch(ctx context.Context, metrics []Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, m := range metrics {
		if err := validateMetrics(&m); err != nil {
			return err
		}
	}

	// Use a single transaction
	transaction, err := ps.database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if err := ps.saveMetricsWithExecutor(ctx, transaction, &m); err != nil {
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	// Commit the transaction
	if err = transaction.Commit(); err != nil {
		return err
	}

	return nil
}

// sqlExecutor is a helper to avoid SQL duplication.
type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (ps *PostgresStorage) saveMetricsWithExecutor(ctx context.Context, executor sqlExecutor, m *Metrics) error {
	switch m.MType {
	case Counter:
		_, err := executor.ExecContext(ctx,
			"INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3) ON CONFLICT (id, mtype) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta",
			m.ID,
			m.MType,
			*m.Delta)
		if err != nil {
			return err
		}
	case Gauge:
		_, err := executor.ExecContext(ctx,
			"INSERT INTO metrics (id, mtype, value) VALUES ($1, $2, $3) ON CONFLICT (id, mtype) DO UPDATE SET value = EXCLUDED.value",
			m.ID,
			m.MType,
			*m.Value)
		if err != nil {
			return err
		}
	}

	return nil
}
