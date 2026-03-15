package stripe

import (
	"fmt"

	stripe "github.com/stripe/stripe-go/v82"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
)

func mapPaymentIntentStatus(status stripe.PaymentIntentStatus) (domain.PaymentStatus, error) {
	switch status {
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return domain.PaymentStatusPending, nil

	case stripe.PaymentIntentStatusRequiresAction:
		return domain.PaymentStatusRequiresAction, nil

	case stripe.PaymentIntentStatusProcessing:
		return domain.PaymentStatusProcessing, nil

	case stripe.PaymentIntentStatusSucceeded:
		return domain.PaymentStatusSucceeded, nil

	case stripe.PaymentIntentStatusCanceled:
		return domain.PaymentStatusCancelled, nil

	default:
		return "", fmt.Errorf("unsupported stripe payment intent status: %s", status)
	}
}

func toProviderPaymentResult(intent *stripe.PaymentIntent) (domain.CreateProviderPaymentResult, error) {
	mappedStatus, err := mapPaymentIntentStatus(intent.Status)
	if err != nil {
		return domain.CreateProviderPaymentResult{}, err
	}

	result := domain.CreateProviderPaymentResult{
		ProviderName:      ProviderName,
		ProviderPaymentID: intent.ID,
		ClientSecret:      intent.ClientSecret,
		Status:            mappedStatus,
		NextAction:        domain.NoNextAction(),
	}

	// TODO: enrich this later when we handle client-side confirmation and 3DS UX.

	return result, nil
}
