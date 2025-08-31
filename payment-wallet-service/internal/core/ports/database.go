package ports

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Database interface {
	WithTx(ctx context.Context, fn func(*pgx.Tx) error) error
}
