package stripe

import (
	"testing"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v82"
)

func TestMapPaymentIntentStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  stripe.PaymentIntentStatus
		want   domain.PaymentStatus
		hasErr bool
	}{
		{
			name:  "requires payment method maps to pending",
			input: stripe.PaymentIntentStatusRequiresPaymentMethod,
			want:  domain.PaymentStatusPending,
		},
		{
			name:  "requires action maps to requires_action",
			input: stripe.PaymentIntentStatusRequiresAction,
			want:  domain.PaymentStatusRequiresAction,
		},
		{
			name:  "processing maps to processing",
			input: stripe.PaymentIntentStatusProcessing,
			want:  domain.PaymentStatusProcessing,
		},
		{
			name:  "succeeded maps to succeeded",
			input: stripe.PaymentIntentStatusSucceeded,
			want:  domain.PaymentStatusSucceeded,
		},
		{
			name:  "canceled maps to cancelled",
			input: stripe.PaymentIntentStatusCanceled,
			want:  domain.PaymentStatusCancelled,
		},
		{
			name:   "unknown status errors",
			input:  stripe.PaymentIntentStatus("mystery_status"),
			hasErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := mapPaymentIntentStatus(tt.input)

			if tt.hasErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
