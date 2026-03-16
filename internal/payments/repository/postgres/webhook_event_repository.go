package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedWebhookEventRepository struct {
	pool *pgxpool.Pool
}

func NewProcessedWebhookEventRepository(pool *pgxpool.Pool) *ProcessedWebhookEventRepository {
	return &ProcessedWebhookEventRepository{
		pool: pool,
	}
}

func (r *ProcessedWebhookEventRepository) SaveProcessedEvent(ctx context.Context, providerName, eventID, eventType string) error {
	const query = `
		INSERT INTO processed_webhook_events (
			event_id,
			provider_name,
			event_type,
			processed_at
		) VALUES (
			$1, $2, $3, $4
		)
		`

	_, err := r.pool.Exec(ctx, query, eventID, providerName, eventType, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("save processed webhook event: %w", err)
	}

	return nil
}

func (r *ProcessedWebhookEventRepository) HasProcessedEvent(ctx context.Context, providerName, eventID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT
				1
			FROM
				processed_webhook_events
			WHERE
				provider_name = $1 AND 
				event_id = $2
		)
		`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, providerName, eventID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check processed webhook event: %w", err)
	}

	return exists, nil
}

var _ domain.ProcessedWebhookEventRepository = (*ProcessedWebhookEventRepository)(nil)
