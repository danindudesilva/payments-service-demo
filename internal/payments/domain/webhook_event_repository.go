package domain

import "context"

type ProcessedWebhookEventRepository interface {
	SaveProcessedEvent(ctx context.Context, providerName, eventID, eventType string) error
	HasProcessedEvent(ctx context.Context, providerName, eventID string) (bool, error)
}
