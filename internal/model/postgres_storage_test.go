package model

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
			AddRow("PollCount", Counter, int64(42), nil)

		mock.ExpectQuery(expectedQuery).
			WithArgs("PollCount", Counter).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "PollCount", Counter)

		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "PollCount", m.ID)
		assert.Equal(t, Counter, m.MType)
		assert.Equal(t, int64(42), *m.Delta)
		assert.Nil(t, m.Value)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successfully retrieves a gauge metric", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		expectedQuery := `SELECT id, mtype, delta, value FROM metrics WHERE id = \$1 AND mtype = \$2`

		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("Alloc", Gauge, nil, float64(123.45))

		mock.ExpectQuery(expectedQuery).
			WithArgs("Alloc", Gauge).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "Alloc", Gauge)

		require.NoError(t, err)
		require.NotNil(t, m)
		assert.Equal(t, "Alloc", m.ID)
		assert.Equal(t, Gauge, m.MType)
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
			WithArgs("Missing", Gauge).
			WillReturnError(sql.ErrNoRows)

		m, err := s.GetMetrics(context.Background(), "Missing", Gauge)

		assert.Nil(t, m)
		assert.ErrorIs(t, err, errMetricsNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrUnknownMetricsType for invalid type", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)

		m, err := s.GetMetrics(context.Background(), "SomeMetric", "invalid_type")

		assert.Nil(t, m)
		assert.ErrorIs(t, err, ErrUnknownMetricsType)
	})

	t.Run("returns ErrDeltaIsNil if counter DB row has NULL delta", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value"}).
			AddRow("PollCount", Counter, nil, nil) // Null delta

		mock.ExpectQuery(`SELECT id, mtype, delta, value FROM metrics`).
			WithArgs("PollCount", Counter).
			WillReturnRows(rows)

		m, err := s.GetMetrics(context.Background(), "PollCount", Counter)

		assert.Nil(t, m)
		assert.ErrorIs(t, err, ErrDeltaIsNil)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_SaveMetrics(t *testing.T) {
	t.Run("successfully saves a counter metric using UPSERT", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &Metrics{ID: "PollCount", MType: Counter, Delta: ptr(int64(10))}

		expectedExec := `INSERT INTO metrics \(id, mtype, delta\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET delta = metrics\.delta \+ EXCLUDED\.delta`

		mock.ExpectExec(expectedExec).
			WithArgs("PollCount", Counter, int64(10)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = s.SaveMetrics(context.Background(), metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successfully saves a gauge metric using UPSERT", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &Metrics{ID: "Alloc", MType: Gauge, Value: ptr(float64(99.9))}

		expectedExec := `INSERT INTO metrics \(id, mtype, value\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(id, mtype\) DO UPDATE SET value = \$3`

		mock.ExpectExec(expectedExec).
			WithArgs("Alloc", Gauge, float64(99.9)).
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

		err = s.SaveMetrics(context.Background(), &Metrics{ID: "", MType: Counter})
		assert.ErrorIs(t, err, ErrEmptyMetricsID)

		err = s.SaveMetrics(context.Background(), nil)
		assert.ErrorIs(t, err, ErrEmptyMetricsID)
	})

	t.Run("returns error when database exec fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		s := NewPostgresStorage(db)
		metric := &Metrics{ID: "Alloc", MType: Gauge, Value: ptr(float64(99.9))}

		mock.ExpectExec(`INSERT INTO metrics`).
			WithArgs("Alloc", Gauge, float64(99.9)).
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
			AddRow("PollCount", Counter, int64(100), nil).
			AddRow("Alloc", Gauge, nil, float64(512.25))

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
			AddRow("PollCount", Counter, int64(100), nil).
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
