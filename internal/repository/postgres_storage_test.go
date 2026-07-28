package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paveltovchigrechko/metrics-service/internal/common"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_GetMetrics(t *testing.T) {
	t.Run("successfully retrieves a counter metric", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		expectedQuery := `SELECT id, mtype, delta, value FROM metrics WHERE id = \$1 AND mtype = \$2`

		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("PollCount", model.Counter, int64(42), nil)

		mock.ExpectQuery(expectedQuery).
			WithArgs("PollCount", model.Counter).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "PollCount", model.Counter)

		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "PollCount", m.ID)
		assert.Equal(t, model.Counter, m.MType)
		assert.Equal(t, int64(42), *m.Delta)
		assert.Nil(t, m.Value)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successfully retrieves a model.Gauge metric", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		expectedQuery := `SELECT id, mtype, delta, value FROM metrics WHERE id = \$1 AND mtype = \$2`

		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("Alloc", model.Gauge, nil, float64(123.45))

		mock.ExpectQuery(expectedQuery).
			WithArgs("Alloc", model.Gauge).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "Alloc", model.Gauge)

		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "Alloc", m.ID)
		assert.Equal(t, model.Gauge, m.MType)
		assert.Equal(t, float64(123.45), *m.Value)
		assert.Nil(t, m.Delta)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns errMetricsNotFound when no rows exist", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		mock.ExpectQuery(`SELECT id, mtype, delta, value FROM metrics`).
			WithArgs("Missing", model.Gauge).
			WillReturnError(sql.ErrNoRows)

		m, err := s.GetMetrics(context.Background(), "Missing", model.Gauge)

		assert.Nil(t, m)
		assert.ErrorIs(t, err, model.ErrMetricsNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrUnknownMetricsType for invalid type", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)

		m, err := s.GetMetrics(context.Background(), "SomeMetric", "invalid_type")

		assert.Nil(t, m)
		assert.ErrorIs(t, err, model.ErrUnknownMetricsType)
	})

	t.Run("returns ErrDeltaIsNil if counter DB row has NULL delta", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("PollCount", model.Counter, nil, nil)

		mock.ExpectQuery(`SELECT id, mtype, delta, value FROM metrics`).
			WithArgs("PollCount", model.Counter).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "PollCount", model.Counter)

		assert.Nil(t, m)
		assert.ErrorIs(t, err, model.ErrDeltaIsNil)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_SaveMetrics(t *testing.T) {
	t.Run("successfully saves a counter metric using UPSERT", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &model.Metrics{ID: "PollCount", MType: model.Counter, Delta: common.Ptr(int64(10))}

		expectedExec := `INSERT INTO metrics \(id, mtype, delta\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET delta = metrics\.delta \+ EXCLUDED\.delta`

		mock.ExpectExec(expectedExec).
			WithArgs("PollCount", model.Counter, int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = s.SaveMetrics(context.Background(), metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successfully saves a model.Gauge metric using UPSERT", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &model.Metrics{ID: "Alloc", MType: model.Gauge, Value: common.Ptr(float64(99.9))}

		expectedExec := `INSERT INTO metrics \(id, mtype, value\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET value = EXCLUDED\.value`

		mock.ExpectExec(expectedExec).
			WithArgs("Alloc", model.Gauge, float64(99.9)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = s.SaveMetrics(context.Background(), metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrEmptyMetricsID for invalid metric", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)

		err = s.SaveMetrics(context.Background(), &model.Metrics{ID: "", MType: model.Counter})
		assert.ErrorIs(t, err, model.ErrEmptyMetricsID)

		err = s.SaveMetrics(context.Background(), nil)
		assert.ErrorIs(t, err, model.ErrMetricsIsNil)
	})

	t.Run("returns error when database exec fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &model.Metrics{ID: "Alloc", MType: model.Gauge, Value: common.Ptr(float64(99.9))}

		mock.ExpectExec(`INSERT INTO metrics`).
			WithArgs("Alloc", model.Gauge, float64(99.9)).
			WillReturnError(errors.New("connection reset by peer"))

		err = s.SaveMetrics(context.Background(), metric)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection reset by peer")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_GetAllMetrics(t *testing.T) {
	t.Run("successfully fetches all metrics", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)

		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("PollCount", model.Counter, int64(100), nil).
			AddRow("Alloc", model.Gauge, nil, float64(512.25))

		mock.ExpectQuery(`SELECT id, mtype, delta, value FROM metrics`).
			WillReturnRows(rows)

		metrics, err := s.GetAllMetrics(context.Background())

		require.NoError(t, err)
		assert.Len(t, metrics, 2)
		assert.Equal(t, "PollCount", metrics[0].ID)
		assert.Equal(t, int64(100), *metrics[0].Delta)
		assert.Equal(t, "Alloc", metrics[1].ID)
		assert.Equal(t, float64(512.25), *metrics[1].Value)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on row iteration failure", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)

		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("PollCount", model.Counter, int64(100), nil).
			RowError(0, errors.New("read error during row fetch"))

		mock.ExpectQuery(`SELECT id, mtype, delta, value FROM metrics`).
			WillReturnRows(rows)

		metrics, err := s.GetAllMetrics(context.Background())

		assert.Nil(t, metrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read error during row fetch")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_SaveBatch(t *testing.T) {
	t.Run("successfully saves a batch of metrics in a transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		batch := []model.Metrics{
			{ID: "PollCount", MType: model.Counter, Delta: common.Ptr(int64(5))},
			{ID: "Alloc", MType: model.Gauge, Value: common.Ptr(10.5)},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO metrics \(id, mtype, delta\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET delta = metrics\.delta \+ EXCLUDED\.delta`).
			WithArgs("PollCount", model.Counter, int64(5)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO metrics \(id, mtype, value\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET value = EXCLUDED\.value`).
			WithArgs("Alloc", model.Gauge, 10.5).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = s.SaveBatch(context.Background(), batch)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns nil immediately on empty batch slice", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		err = s.SaveBatch(context.Background(), []model.Metrics{})
		assert.NoError(t, err)
	})

	t.Run("fails validation before opening transaction", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		batch := []model.Metrics{
			{ID: "", MType: model.Counter, Delta: common.Ptr(int64(5))}, // Invalid ID
		}

		err = s.SaveBatch(context.Background(), batch)
		assert.ErrorIs(t, err, model.ErrEmptyMetricsID)
	})

	t.Run("rolls back transaction when execution fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		batch := []model.Metrics{
			{ID: "PollCount", MType: model.Counter, Delta: common.Ptr(int64(5))},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO metrics`).
			WithArgs("PollCount", model.Counter, int64(5)).
			WillReturnError(errors.New("db write error"))
		mock.ExpectRollback()

		err = s.SaveBatch(context.Background(), batch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db write error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
