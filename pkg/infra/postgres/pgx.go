package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type (
	PGX interface {
		BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
		PGXQuerier
	}
	
	PGXQuerier interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	}
	
	Repository[R any] interface {
		WithTx(q PGXQuerier) R
	}
)

// Rollback attempts to rollback the given transaction. If the rollback fails,
// it wraps the rollback error with the original error. Otherwise, it returns the original error.
func Rollback(tx pgx.Tx, ctx context.Context, originalErr error) error {
    if err := tx.Rollback(ctx); err != nil {
        return fmt.Errorf("rollback failed: %w, original error: %w", err, originalErr)
    }
    return originalErr
}
