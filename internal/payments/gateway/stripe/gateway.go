package stripe

import (
	"context"
	"fmt"
	"strings"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"

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
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(request.Money.Amount),
		Currency:    stripe.String(strings.ToLower(request.Money.Currency)),
		Description: stripe.String(request.Description),
	}

	params.Metadata = map[string]string{
		MetadataKeyAttemptID: request.AttemptID,
		MetadataKeyOrderID:   request.OrderID,
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("create stripe payment intent: %w", err)
	}

	mappedStatus, err := mapPaymentIntentStatus(intent.Status)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, fmt.Errorf("map stripe payment intent status: %w", err)
	}

	return domain.CreateProviderPaymentResult{
		ProviderName:      ProviderName,
		ProviderPaymentID: intent.ID,
		ClientSecret:      intent.ClientSecret,
		Status:            mappedStatus,
		NextAction:        domain.NoNextAction(),
	}, nil
}

func (g *Gateway) GetPayment(
	ctx context.Context,
	providerPaymentID string,
) (domain.CreateProviderPaymentResult, error) {
	return domain.CreateProviderPaymentResult{}, fmt.Errorf("not implemented")
}
