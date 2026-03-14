package gateway

import (
	"context"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

type noopGateway struct{}

func NewNoopGateway() *noopGateway {
	return &noopGateway{}
}

func (g *noopGateway) CreatePayment(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{
		ProviderName:      "fake",
		ProviderPaymentID: "fake_payment_id",
		ClientSecret:      "fake_client_secret",
		Status:            domain.PaymentStatusPending,
	}, nil
}

func (g *noopGateway) GetPayment(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{
		ProviderName:      "fake",
		ProviderPaymentID: providerPaymentID,
		Status:            domain.PaymentStatusPending,
	}, nil
}
