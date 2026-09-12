package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/k11ngp1ng/url-shortener-go/internal/domain"
)

// PostgresURLStore persists shortened URLs in PostgreSQL.
type PostgresURLStore struct{ pool *pgxpool.Pool }

func NewPostgresURLStore(ctx context.Context, databaseURL string) (*PostgresURLStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresURLStore{pool: pool}, nil
}

func (s *PostgresURLStore) Close()                         { s.pool.Close() }
func (s *PostgresURLStore) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *PostgresURLStore) Create(u domain.URL) (domain.URL, error) {
	err := s.pool.QueryRow(context.Background(), `INSERT INTO urls (code, original_url, expires_at)
VALUES ($1, $2, $3) RETURNING code, original_url, clicks, created_at, expires_at`, u.Code, u.OriginalURL, u.ExpiresAt).
		Scan(&u.Code, &u.OriginalURL, &u.Clicks, &u.CreatedAt, &u.ExpiresAt)
	if isUniqueViolation(err) {
		return domain.URL{}, ErrCodeAlreadyExists
	}
	return u, err
}

func (s *PostgresURLStore) Get(code string) (u domain.URL, err error) {
	err = s.pool.QueryRow(context.Background(), `SELECT code, original_url, clicks, created_at, expires_at FROM urls WHERE code = $1`, code).
		Scan(&u.Code, &u.OriginalURL, &u.Clicks, &u.CreatedAt, &u.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *PostgresURLStore) IncrementClicks(code string) (u domain.URL, err error) {
	err = s.pool.QueryRow(context.Background(), `UPDATE urls SET clicks = clicks + 1 WHERE code = $1
RETURNING code, original_url, clicks, created_at, expires_at`, code).
		Scan(&u.Code, &u.OriginalURL, &u.Clicks, &u.CreatedAt, &u.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
