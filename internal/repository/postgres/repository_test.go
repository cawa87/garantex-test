package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
	"github.com/cawa87/garantex-test/internal/service/exchange"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Use test database URL - you might need to adjust this for your test environment
	dsn := "postgres://postgres:password@localhost:5432/garantex_test?sslmode=disable"

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)

	// Create test table
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS rates (
			id BIGSERIAL PRIMARY KEY,
			ask DECIMAL(20, 8) NOT NULL,
			bid DECIMAL(20, 8) NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		_, _ = pool.Exec(context.Background(), "DROP TABLE IF EXISTS rates")
		pool.Close()
	}

	return pool, cleanup
}

func TestNewRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, err := sl.New("info")
	require.NoError(t, err)

	repo := &Repository{
		pool:   pool,
		logger: logger,
	}

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.pool)
	assert.NotNil(t, repo.logger)
}

func TestSaveRate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, err := sl.New("info")
	require.NoError(t, err)

	repo := &Repository{
		pool:   pool,
		logger: logger,
	}

	rate := &exchange.Rate{
		Ask:       100.50,
		Bid:       100.40,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	err = repo.SaveRate(ctx, rate)

	assert.NoError(t, err)

	// Verify the rate was saved
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM rates").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetLatestRate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, err := sl.New("info")
	require.NoError(t, err)

	repo := &Repository{
		pool:   pool,
		logger: logger,
	}

	ctx := context.Background()

	// Test empty database
	_, err = repo.GetLatestRate(ctx)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)

	// Insert test data
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO rates (ask, bid, timestamp, created_at)
		VALUES ($1, $2, $3, $4)
	`, 100.50, 100.40, now, now)
	require.NoError(t, err)

	// Get latest rate
	rate, err := repo.GetLatestRate(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, rate)
	assert.Equal(t, 100.50, rate.Ask)
	assert.Equal(t, 100.40, rate.Bid)
	assert.WithinDuration(t, now, rate.Timestamp, time.Second)
}

func TestGetRatesByTimeRange(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, err := sl.New("info")
	require.NoError(t, err)

	repo := &Repository{
		pool:   pool,
		logger: logger,
	}

	ctx := context.Background()

	// Insert test data
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO rates (ask, bid, timestamp, created_at)
		VALUES 
			($1, $2, $3, $4),
			($5, $6, $7, $8),
			($9, $10, $11, $12)
	`,
		100.50, 100.40, now.Add(-2*time.Hour), now,
		100.60, 100.50, now.Add(-1*time.Hour), now,
		100.70, 100.60, now, now,
	)
	require.NoError(t, err)

	// Get rates in time range
	from := now.Add(-3 * time.Hour)
	to := now.Add(-30 * time.Minute)
	rates, err := repo.GetRatesByTimeRange(ctx, from, to)

	assert.NoError(t, err)
	assert.Len(t, rates, 2)

	// Verify rates are ordered by timestamp DESC
	assert.Equal(t, 100.60, rates[0].Ask)
	assert.Equal(t, 100.50, rates[1].Ask)
}

func TestGetRatesCount(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, err := sl.New("info")
	require.NoError(t, err)

	repo := &Repository{
		pool:   pool,
		logger: logger,
	}

	ctx := context.Background()

	// Test empty database
	count, err := repo.GetRatesCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Insert test data
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO rates (ask, bid, timestamp, created_at)
		VALUES ($1, $2, $3, $4)
	`, 100.50, 100.40, now, now)
	require.NoError(t, err)

	// Get count
	count, err = repo.GetRatesCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
