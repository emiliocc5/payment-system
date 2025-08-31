package postgresql

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_once sync.Once
	_pool *pgxpool.Pool
)

type Database struct {
	DB *pgxpool.Pool
}

func NewDatabase(ctx context.Context, dsn string) (*Database, error) {
	db, err := Connect(ctx, dsn)

	return &Database{DB: db}, err
}

func (d *Database) WithTx(ctx context.Context, fn func(*pgx.Tx) error) error {
	tx, err := d.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()

	err = fn(&tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	var err error
	_once.Do(func() {
		_pool, err = pgxpool.New(ctx, dsn)
	})
	return _pool, err
}
