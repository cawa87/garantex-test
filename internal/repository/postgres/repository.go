package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
	"github.com/cawa87/garantex-test/internal/service/exchange"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *sl.Logger
}

type Rate struct {
	ID        int64     `db:"id"`
	Ask       float64   `db:"ask"`
	Bid       float64   `db:"bid"`
	Timestamp time.Time `db:"timestamp"`
	CreatedAt time.Time `db:"created_at"`
}

func NewRepository(dsn string, logger *sl.Logger) (*Repository, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{
		pool:   pool,
		logger: logger,
	}, nil
}

func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repository) SaveRate(ctx context.Context, rate *exchange.Rate) error {
	query := `
		INSERT INTO rates (ask, bid, timestamp, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, rate.Ask, rate.Bid, rate.Timestamp, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save rate: %w", err)
	}

	r.logger.Debug("Rate saved to database",
		"ask", rate.Ask,
		"bid", rate.Bid,
		"timestamp", rate.Timestamp)

	return nil
}

func (r *Repository) GetLatestRate(ctx context.Context) (*Rate, error) {
	query := `
		SELECT id, ask, bid, timestamp, created_at
		FROM rates
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var rate Rate
	err := r.pool.QueryRow(ctx, query).Scan(
		&rate.ID,
		&rate.Ask,
		&rate.Bid,
		&rate.Timestamp,
		&rate.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get latest rate: %w", err)
	}

	return &rate, nil
}

func (r *Repository) GetRatesByTimeRange(ctx context.Context, from, to time.Time) ([]*Rate, error) {
	query := `
		SELECT id, ask, bid, timestamp, created_at
		FROM rates
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	var rates []*Rate
	for rows.Next() {
		var rate Rate
		err := rows.Scan(
			&rate.ID,
			&rate.Ask,
			&rate.Bid,
			&rate.Timestamp,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		rates = append(rates, &rate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return rates, nil
}

func (r *Repository) GetRatesCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM rates`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get rates count: %w", err)
	}

	return count, nil
}
