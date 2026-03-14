package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPaymentAttempt_TransitionMatrix(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		from      PaymentStatus
		to        PaymentStatus
		shouldErr bool
	}

	cases := []testCase{
		{name: "pending to requires_action", from: PaymentStatusPending, to: PaymentStatusRequiresAction},
		{name: "pending to processing", from: PaymentStatusPending, to: PaymentStatusProcessing},
		{name: "pending to succeeded", from: PaymentStatusPending, to: PaymentStatusSucceeded},
		{name: "pending to failed", from: PaymentStatusPending, to: PaymentStatusFailed},
		{name: "pending to cancelled", from: PaymentStatusPending, to: PaymentStatusCancelled},

		{name: "requires_action to processing", from: PaymentStatusRequiresAction, to: PaymentStatusProcessing},
		{name: "requires_action to succeeded", from: PaymentStatusRequiresAction, to: PaymentStatusSucceeded},
		{name: "requires_action to failed", from: PaymentStatusRequiresAction, to: PaymentStatusFailed},
		{name: "requires_action to cancelled", from: PaymentStatusRequiresAction, to: PaymentStatusCancelled},

		{name: "processing to succeeded", from: PaymentStatusProcessing, to: PaymentStatusSucceeded},
		{name: "processing to failed", from: PaymentStatusProcessing, to: PaymentStatusFailed},
		{name: "processing to cancelled", from: PaymentStatusProcessing, to: PaymentStatusCancelled},

		{name: "processing to requires_action invalid", from: PaymentStatusProcessing, to: PaymentStatusRequiresAction, shouldErr: true},
		{name: "succeeded to failed invalid", from: PaymentStatusSucceeded, to: PaymentStatusFailed, shouldErr: true},
		{name: "failed to succeeded invalid", from: PaymentStatusFailed, to: PaymentStatusSucceeded, shouldErr: true},
		{name: "cancelled to processing invalid", from: PaymentStatusCancelled, to: PaymentStatusProcessing, shouldErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			attempt := mustAttemptWithStatus(t, tc.from)

			err := attempt.transitionTo(tc.to)
			if tc.shouldErr {
				require.ErrorIs(t, err, ErrInvalidTransition)
				require.Equal(t, tc.from, attempt.Status)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.to, attempt.Status)
		})
	}
}

func mustAttemptWithStatus(t *testing.T, status PaymentStatus) *PaymentAttempt {
	t.Helper()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	attempt.Status = status
	return attempt
}
