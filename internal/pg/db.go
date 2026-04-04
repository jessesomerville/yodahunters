package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// DB is the interface for interacting with a postgres database.
// It is satisfied by *Client.
type DB interface {
	Query(ctx context.Context, sql string, params ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, params ...any) (pgx.Row, error)
	Exec(ctx context.Context, sql string, params ...any) error
}
