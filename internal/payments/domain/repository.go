package domain

import "context"

type PaymentAttemptRepository interface {
	Save(ctx context.Context, attempt *PaymentAttempt) error
	GetByID(ctx context.Context, id string) (*PaymentAttempt, error)
	GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*PaymentAttempt, error)
}
