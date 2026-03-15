package gateway

import (
	"fmt"

	"github.com/danindudesilva/payments-service/internal/config"
	"github.com/danindudesilva/payments-service/internal/payments/domain"
	fakegateway "github.com/danindudesilva/payments-service/internal/payments/gateway/fake"
)

func New(cfg config.Config) (domain.PaymentGateway, error) {
	switch cfg.PaymentsProvider {
	case "fake":
		return fakegateway.New(), nil

	case "stripe":
		return nil, fmt.Errorf("stripe gateway is not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported payments provider: %s", cfg.PaymentsProvider)
	}
}
