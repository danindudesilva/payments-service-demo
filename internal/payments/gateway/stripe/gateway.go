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

func (g *Gateway) CreatePayment(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
	params := &stripe.PaymentIntentCreateParams{
		Amount:      stripe.Int64(request.Money.Amount),
		Currency:    stripe.String(strings.ToLower(request.Money.Currency)),
		Description: stripe.String(request.Description),
	}

	params.Metadata = map[string]string{
		MetadataKeyAttemptID: request.AttemptID,
		MetadataKeyOrderID:   request.OrderID,
	}

	intent, err := g.client.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("create stripe payment intent: %w", err)
	}

	result, err := toProviderPaymentResult(intent)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("map stripe payment intent: %w", err)
	}

	return result, nil
}

func (g *Gateway) GetPayment(
	ctx context.Context,
	providerPaymentID string,
) (domain.CreateProviderPaymentResult, error) {
	if strings.TrimSpace(providerPaymentID) == "" {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("provider payment id must not be empty")
	}

	intent, err := g.client.V1PaymentIntents.Retrieve(ctx, providerPaymentID, &stripe.PaymentIntentRetrieveParams{})
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("get stripe payment intent: %w", err)
	}

	result, err := toProviderPaymentResult(intent)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("map stripe payment intent: %w", err)
	}

	return result, nil
}
