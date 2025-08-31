package postgresql

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/storage"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	db *pgxpool.Pool
}

func NewPgBalanceRepository(db *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) Get(ctx context.Context, userID string) (*domain.Balance, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	query := "SELECT user_id, available_balance, reserved_balance, updated_at FROM balance WHERE user_id = $1"
	var balance domain.Balance
	err = r.db.QueryRow(ctx, query, uid).Scan(
		&balance.UserID,
		&balance.Available,
		&balance.Reserved,
		&balance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *BalanceRepository) ReserveFunds(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	query := "UPDATE balance " +
		"SET " +
		"available_balance = available_balance - $1, " +
		"reserved_balance = reserved_balance + $1, " +
		"updated_at = NOW() " +
		"WHERE user_id = $2 " +
		"AND available_balance >= $1"

	result, errExec := tx.Exec(ctx, query, amount, uid)
	if errExec != nil {
		return errExec
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrInsufficientFunds
	}

	return nil
}
func (r *BalanceRepository) ReleaseFunds(ctx context.Context, userID string, amount int64) error {
	return nil
}
func (r *BalanceRepository) ConfirmReserve(ctx context.Context, userID string, amount int64) error {
	return nil
}
