package fake

import (
	"context"
	"fmt"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/google/uuid"
)

type Gateway struct{}

func New() *Gateway {
	return &Gateway{}
}

func (g *Gateway) CreatePayment(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
	providerPaymentID := uuid.New().String()

	return domain.CreateProviderPaymentResult{
		ProviderName:      "fake",
		ProviderPaymentID: providerPaymentID,
		ClientSecret:      fmt.Sprintf("fake_secret_%s", providerPaymentID),
		Status:            domain.PaymentStatusPending,
	}, nil
}

func (g *Gateway) GetPayment(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{
		ProviderName:      "fake",
		ProviderPaymentID: providerPaymentID,
		Status:            domain.PaymentStatusPending,
	}, nil
}
