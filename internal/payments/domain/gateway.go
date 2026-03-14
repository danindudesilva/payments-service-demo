package domain

import "context"

type CreateProviderPaymentRequest struct {
	AttemptID   string
	OrderID     string
	Money       Money
	ReturnURL   string
	Description string
}

type CreateProviderPaymentResult struct {
	ProviderName      string
	ProviderPaymentID string
	ClientSecret      string
	Status            PaymentStatus
	NextAction        NextAction
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, request CreateProviderPaymentRequest) (CreateProviderPaymentResult, error)
	GetPayment(ctx context.Context, providerPaymentID string) (CreateProviderPaymentResult, error)
}
