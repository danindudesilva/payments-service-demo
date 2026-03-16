package postgres

import (
	"context"
	"fmt"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Save(ctx context.Context, attempt *domain.PaymentAttempt) error {
	const query = `
		INSERT INTO payment_attempts (
			id,
			order_id,
			idempotency_key,
			provider_name,
			provider_payment_id,
			status,
			amount,
			currency,
			return_url,
			failure_reason,
			client_secret,
			created_at,
			updated_at,
			completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (id) DO UPDATE SET
			order_id = EXCLUDED.order_id,
			idempotency_key = EXCLUDED.idempotency_key,
			provider_name = EXCLUDED.provider_name,
			provider_payment_id = EXCLUDED.provider_payment_id,
			status = EXCLUDED.status,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			return_url = EXCLUDED.return_url,
			failure_reason = EXCLUDED.failure_reason,
			client_secret = EXCLUDED.client_secret,
			updated_at = EXCLUDED.updated_at,
			completed_at = EXCLUDED.completed_at
		`

	_, err := r.pool.Exec(
		ctx,
		query,
		attempt.ID,
		attempt.OrderID,
		attempt.IdempotencyKey,
		attempt.Provider.ProviderName,
		nullIfEmpty(attempt.Provider.ProviderPaymentID),
		string(attempt.Status),
		attempt.Money.Amount,
		attempt.Money.Currency,
		attempt.ReturnURL,
		attempt.FailureReason,
		attempt.Provider.ClientSecret,
		attempt.Timestamps.CreatedAt,
		attempt.Timestamps.UpdatedAt,
		attempt.Timestamps.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("save payment attempt: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.PaymentAttempt, error) {
	const query = `
		SELECT
			id,
			order_id,
			idempotency_key,
			return_url,
			status,
			amount,
			currency,
			provider_name,
			provider_payment_id,
			client_secret,
			failure_reason,
			created_at,
			updated_at,
			completed_at
		FROM
			payment_attempts
		WHERE
			id = $1
		`

	return scanAttempt(r.pool.QueryRow(ctx, query, id))
}

func (r *Repository) GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*domain.PaymentAttempt, error) {
	const query = `
		SELECT
			id,
			order_id,
			idempotency_key,
			return_url,
			status,
			amount,
			currency,
			provider_name,
			provider_payment_id,
			client_secret,
			failure_reason,
			created_at,
			updated_at,
			completed_at
		FROM
			payment_attempts
		WHERE
			provider_payment_id = $1
		`

	return scanAttempt(r.pool.QueryRow(ctx, query, providerPaymentID))
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.PaymentAttempt, error) {
	const query = `
		SELECT
			id,
			order_id,
			idempotency_key,
			return_url,
			status,
			amount,
			currency,
			provider_name,
			provider_payment_id,
			client_secret,
			failure_reason,
			created_at,
			updated_at,
			completed_at
		FROM
			payment_attempts
		WHERE
			idempotency_key = $1
		`

	return scanAttempt(r.pool.QueryRow(ctx, query, idempotencyKey))
}
