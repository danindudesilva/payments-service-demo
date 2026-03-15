package stripe

import (
	"context"
	"fmt"
	"strings"

	stripe "github.com/stripe/stripe-go/v82"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

const ProviderName = "stripe"

type Gateway struct {
	client *stripe.Client
}

func New(secretKey string) (*Gateway, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("stripe secret key must not be empty")
	}

	client := stripe.NewClient(secretKey)

	return &Gateway{
		client: client,
	}, nil
}

func (g *Gateway) CreatePayment(
	ctx context.Context,
	request domain.CreateProviderPaymentRequest,
) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{}, fmt.Errorf("not implemented")
}

func (g *Gateway) GetPayment(
	ctx context.Context,
	providerPaymentID string,
) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{}, fmt.Errorf("not implemented")
}
