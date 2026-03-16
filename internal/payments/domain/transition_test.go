package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{name: "pending", status: PaymentStatusPending, want: false},
		{name: "requires_action", status: PaymentStatusRequiresAction, want: false},
		{name: "processing", status: PaymentStatusProcessing, want: false},
		{name: "succeeded", status: PaymentStatusSucceeded, want: true},
		{name: "failed", status: PaymentStatusFailed, want: true},
		{name: "cancelled", status: PaymentStatusCancelled, want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.status.IsTerminal())
		})
	}
}

func TestPaymentAttempt_TransitionMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    PaymentStatus
		to      PaymentStatus
		wantErr error
	}{
		{name: "pending to pending", from: PaymentStatusPending, to: PaymentStatusPending},
		{name: "pending to requires_action", from: PaymentStatusPending, to: PaymentStatusRequiresAction},
		{name: "pending to processing", from: PaymentStatusPending, to: PaymentStatusProcessing},
		{name: "pending to succeeded", from: PaymentStatusPending, to: PaymentStatusSucceeded},
		{name: "pending to failed", from: PaymentStatusPending, to: PaymentStatusFailed},
		{name: "pending to cancelled", from: PaymentStatusPending, to: PaymentStatusCancelled},

		{name: "requires_action to requires_action", from: PaymentStatusRequiresAction, to: PaymentStatusRequiresAction},
		{name: "requires_action to processing", from: PaymentStatusRequiresAction, to: PaymentStatusProcessing},
		{name: "requires_action to succeeded", from: PaymentStatusRequiresAction, to: PaymentStatusSucceeded},
		{name: "requires_action to failed", from: PaymentStatusRequiresAction, to: PaymentStatusFailed},
		{name: "requires_action to cancelled", from: PaymentStatusRequiresAction, to: PaymentStatusCancelled},
		{name: "requires_action to pending invalid", from: PaymentStatusRequiresAction, to: PaymentStatusPending, wantErr: ErrInvalidTransition},

		{name: "processing to processing", from: PaymentStatusProcessing, to: PaymentStatusProcessing},
		{name: "processing to succeeded", from: PaymentStatusProcessing, to: PaymentStatusSucceeded},
		{name: "processing to failed", from: PaymentStatusProcessing, to: PaymentStatusFailed},
		{name: "processing to cancelled", from: PaymentStatusProcessing, to: PaymentStatusCancelled},
		{name: "processing to pending invalid", from: PaymentStatusProcessing, to: PaymentStatusPending, wantErr: ErrInvalidTransition},
		{name: "processing to requires_action invalid", from: PaymentStatusProcessing, to: PaymentStatusRequiresAction, wantErr: ErrInvalidTransition},

		{name: "succeeded to succeeded", from: PaymentStatusSucceeded, to: PaymentStatusSucceeded},
		{name: "succeeded to failed invalid", from: PaymentStatusSucceeded, to: PaymentStatusFailed, wantErr: ErrInvalidTransition},
		{name: "succeeded to cancelled invalid", from: PaymentStatusSucceeded, to: PaymentStatusCancelled, wantErr: ErrInvalidTransition},

		{name: "failed to failed", from: PaymentStatusFailed, to: PaymentStatusFailed},
		{name: "failed to succeeded invalid", from: PaymentStatusFailed, to: PaymentStatusSucceeded, wantErr: ErrInvalidTransition},

		{name: "cancelled to cancelled", from: PaymentStatusCancelled, to: PaymentStatusCancelled},
		{name: "cancelled to processing invalid", from: PaymentStatusCancelled, to: PaymentStatusProcessing, wantErr: ErrInvalidTransition},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attempt := mustAttemptWithStatus(t, tt.from)

			err := attempt.transitionTo(tt.to)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.from, attempt.Status)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.to, attempt.Status)
		})
	}
}

func TestPaymentAttempt_CanBeResumed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{name: "pending", status: PaymentStatusPending, want: false},
		{name: "requires_action", status: PaymentStatusRequiresAction, want: true},
		{name: "processing", status: PaymentStatusProcessing, want: true},
		{name: "succeeded", status: PaymentStatusSucceeded, want: false},
		{name: "failed", status: PaymentStatusFailed, want: false},
		{name: "cancelled", status: PaymentStatusCancelled, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attempt := mustAttemptWithStatus(t, tt.status)
			assert.Equal(t, tt.want, attempt.CanBeResumed())
		})
	}
}

func TestPaymentAttempt_TerminalTransitionsSetCompletedAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		act    func(*PaymentAttempt) error
		status PaymentStatus
	}{
		{
			name: "mark succeeded",
			act: func(a *PaymentAttempt) error {
				return a.MarkSucceeded(now.Add(time.Minute))
			},
			status: PaymentStatusSucceeded,
		},
		{
			name: "mark failed",
			act: func(a *PaymentAttempt) error {
				return a.MarkFailed("declined", now.Add(time.Minute))
			},
			status: PaymentStatusFailed,
		},
		{
			name: "mark cancelled",
			act: func(a *PaymentAttempt) error {
				return a.MarkCancelled(now.Add(time.Minute))
			},
			status: PaymentStatusCancelled,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attempt := mustNewAttempt(t, now)

			err := tt.act(attempt)
			require.NoError(t, err)

			assert.Equal(t, tt.status, attempt.Status)
			require.NotNil(t, attempt.Timestamps.CompletedAt)
			assert.Equal(t, now.Add(time.Minute).UTC(), *attempt.Timestamps.CompletedAt)
		})
	}
}

func TestPaymentAttempt_NonTerminalTransitionsDoNotSetCompletedAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		act  func(*PaymentAttempt) error
	}{
		{
			name: "mark requires action",
			act: func(a *PaymentAttempt) error {
				return a.MarkRequiresAction(NextAction{
					Type:        NextActionTypeRedirect,
					RedirectURL: "https://example.com/3ds",
				}, now.Add(time.Minute))
			},
		},
		{
			name: "mark processing",
			act: func(a *PaymentAttempt) error {
				return a.MarkProcessing(now.Add(time.Minute))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			attempt := mustNewAttempt(t, now)

			err := tt.act(attempt)
			require.NoError(t, err)
			assert.Nil(t, attempt.Timestamps.CompletedAt)
		})
	}
}

func mustAttemptWithStatus(t *testing.T, status PaymentStatus) *PaymentAttempt {
	t.Helper()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		"idempotency-key-123",
		"https://example.com/return",
		Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	attempt.Status = status
	return attempt
}
