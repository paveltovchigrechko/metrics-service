package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/retry"
)

type PostgresStorage struct {
	database *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		database: db,
	}
}

func (ps *PostgresStorage) GetMetrics(ctx context.Context, name string, mType string) (*model.Metrics, error) {
	var metrics *model.Metrics
	err := retry.Do(ctx, isRetriablePgError, func() error {
		var innerErr error
		metrics, innerErr = ps.getMetrics(ctx, name, mType)
		return innerErr
	})
	return metrics, err
}

func (ps *PostgresStorage) SaveMetrics(ctx context.Context, m *model.Metrics) error {
	return retry.Do(ctx, isRetriablePgError, func() error {
		return ps.saveMetrics(ctx, m)
	})
}

func (ps *PostgresStorage) GetAllMetrics(ctx context.Context) ([]model.Metrics, error) {
	var metrics []model.Metrics
	err := retry.Do(ctx, isRetriablePgError, func() error {
		var innerErr error
		metrics, innerErr = ps.getAllMetrics(ctx)
		return innerErr
	})
	return metrics, err
}

func (ps *PostgresStorage) SaveBatch(ctx context.Context, metrics []model.Metrics) error {
	return retry.Do(ctx, isRetriablePgError, func() error {
		return ps.saveBatch(ctx, metrics)
	})
}

func (ps *PostgresStorage) getMetrics(ctx context.Context, name string, mType string) (*model.Metrics, error) {
	if err := model.ValidateName(name); err != nil {
		return nil, err
	}
	if err := model.ValidateType(mType); err != nil {
		return nil, err
	}

	row := ps.database.QueryRowContext(ctx, "SELECT id, mtype, delta, value FROM metrics WHERE id = $1 AND mtype = $2", name, mType)

	var m model.Metrics
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(&m.ID, &m.MType, &delta, &value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrMetricsNotFound
	} else if err != nil {
		return nil, err
	}

	switch mType {
	case model.Counter:
		if delta.Valid {
			m.Delta = &delta.Int64
		} else {
			return nil, model.ErrDeltaIsNil
		}
	case model.Gauge:
		if value.Valid {
			m.Value = &value.Float64
		} else {
			return nil, model.ErrValueIsNil
		}
	}

	return &m, nil
}

func (ps *PostgresStorage) saveMetrics(ctx context.Context, m *model.Metrics) error {
	if err := model.ValidateMetrics(m); err != nil {
		return err
	}

	if err := ps.saveMetricsWithExecutor(ctx, ps.database, m); err != nil {
		return err
	}

	return nil
}

func (ps *PostgresStorage) getAllMetrics(ctx context.Context) ([]model.Metrics, error) {
	metrics := make([]model.Metrics, 0) // Add some capacity here?

	rows, err := ps.database.QueryContext(ctx, "SELECT id, mtype, delta, value FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Metrics
		var delta sql.NullInt64
		var value sql.NullFloat64

		err := rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, err
		}

		switch m.MType {
		case model.Counter:
			if delta.Valid {
				m.Delta = &delta.Int64
			} else {
				return nil, model.ErrDeltaIsNil
			}
		case model.Gauge:
			if value.Valid {
				m.Value = &value.Float64
			} else {
				return nil, model.ErrValueIsNil
			}
		default:
			return nil, model.ErrUnknownMetricsType
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

func (ps *PostgresStorage) saveBatch(ctx context.Context, metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, m := range metrics {
		if err := model.ValidateMetrics(&m); err != nil {
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

func (ps *PostgresStorage) saveMetricsWithExecutor(ctx context.Context, executor sqlExecutor, m *model.Metrics) error {
	switch m.MType {
	case model.Counter:
		_, err := executor.ExecContext(ctx,
			"INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3) ON CONFLICT (id, mtype) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta",
			m.ID,
			m.MType,
			*m.Delta)
		if err != nil {
			return err
		}
	case model.Gauge:
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

func isRetriablePgError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if strings.HasPrefix(pgErr.Code, "08") {
			return true
		}
	}

	return false
}
