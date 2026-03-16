package gateway

import (
	"fmt"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	fakegateway "github.com/danindudesilva/payments-service/internal/payments/gateway/fake"
	stripegateway "github.com/danindudesilva/payments-service/internal/payments/gateway/stripe"
)

type Config struct {
	PaymentsProvider     string
	StripeSecretKey      string
	StripePublishableKey string
}

func New(cfg Config) (domain.PaymentGateway, error) {
	switch cfg.PaymentsProvider {
	case "fake":
		return fakegateway.New(), nil

	case "stripe":
		return stripegateway.New(cfg.StripeSecretKey)

	default:
		return nil, fmt.Errorf("unsupported payments provider: %s", cfg.PaymentsProvider)
	}
}
