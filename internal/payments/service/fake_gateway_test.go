package service

import (
	"context"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

type fakeGateway struct {
	createPaymentFunc func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error)
	getPaymentFunc    func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error)
}

func (f *fakeGateway) CreatePayment(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
	return f.createPaymentFunc(ctx, request)
}

func (f *fakeGateway) GetPayment(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
	return f.getPaymentFunc(ctx, providerPaymentID)
}
