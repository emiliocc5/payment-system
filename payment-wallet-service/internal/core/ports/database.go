package ports

import (
	"context"

	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -destination=../mocks/database_ports_mock.go -package=mocks -source=database_ports.go

type Database interface {
	WithTx(ctx context.Context, fn func(*pgx.Tx) error) error
}
