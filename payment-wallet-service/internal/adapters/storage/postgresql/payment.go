package postgresql

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentsRepository struct {
	db *pgxpool.Pool
}

func NewPgPaymentsRepository(db *pgxpool.Pool) *PaymentsRepository {
	return &PaymentsRepository{db: db}
}

func (p *PaymentsRepository) CheckIdempotency(ctx context.Context, tx pgx.Tx, idempotencyKey string) (bool, error) {
	query := "SELECT COUNT(*) FROM payments WHERE idempotency_key = $1"

	var count int
	err := tx.QueryRow(ctx, query, idempotencyKey).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PaymentsRepository) Create(ctx context.Context, tx pgx.Tx, payment domain.Payment) error {
	query := `
		INSERT INTO payments (
			id,
			idempotency_key,
			user_id,
			amount,
			status,
			service_id,
			client_number,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	uid, err := uuid.Parse(payment.UserID)
	if err != nil {
		return err
	}

	_, errCreate := tx.Exec(ctx, query,
		payment.ID,
		payment.IdempotencyKey,
		uid,
		payment.Amount,
		payment.Status,
		payment.ServiceID,
		payment.ClientNumber,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if errCreate != nil {
		return errCreate
	}

	return nil
}

func (p *PaymentsRepository) Update(ctx context.Context, payment domain.Payment) error {
	return nil
}
