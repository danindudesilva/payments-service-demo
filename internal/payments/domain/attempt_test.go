package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPaymentAttempt(t *testing.T) {
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

	assert.Equal(t, attempt.Status, PaymentStatusPending)
	assert.Equal(t, attempt.Money.Amount, int64(2500))
	assert.Nil(t, attempt.Timestamps.CompletedAt)
}

func TestNewPaymentAttempt_InvalidMoney(t *testing.T) {
	_, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		"idempotency-key-123",
		"https://example.com/return",
		Money{Amount: 0, Currency: "GBP"},
		time.Now(),
	)
	assert.ErrorIs(t, err, ErrInvalidMoney)
}

func TestNewPaymentAttempt_ReturnURLRequired(t *testing.T) {
	t.Parallel()

	_, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		"idempotency-key-123",
		"",
		Money{Amount: 2500, Currency: "GBP"},
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "returnURL must not be empty")
}

func TestNewPaymentAttempt_IdempotencyKeyRequired(t *testing.T) {
	t.Parallel()

	_, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		"",
		"https://example.com/return",
		Money{Amount: 2500, Currency: "GBP"},
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "idempotencyKey must not be empty")
}

func TestLinkProvider(t *testing.T) {
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

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now.Add(time.Minute))
	require.NoError(t, err)

	assert.Equal(t, attempt.Provider.ProviderName, "stripe")
	assert.Equal(t, attempt.Provider.ProviderPaymentID, "pi_123")
}

func TestMarkRequiresAction(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt := mustNewAttempt(t, now)

	err := attempt.MarkRequiresAction(NextAction{
		Type:        NextActionTypeRedirect,
		RedirectURL: "https://example.com/3ds",
	}, now.Add(time.Minute))
	require.NoError(t, err)

	assert.Equal(t, attempt.Status, PaymentStatusRequiresAction)
	assert.Equal(t, attempt.NextAction.Type, NextActionTypeRedirect)
}

func TestMarkRequiresAction_InvalidNextAction(t *testing.T) {
	attempt := mustNewAttempt(t, time.Now())

	err := attempt.MarkRequiresAction(NoNextAction(), time.Now())
	require.ErrorIs(t, err, ErrInvalidNextAction)
}

func TestMarkProcessing_FromRequiresAction(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt := mustNewAttempt(t, now)

	err := attempt.MarkRequiresAction(NextAction{
		Type:        NextActionTypeRedirect,
		RedirectURL: "https://example.com/3ds",
	}, now.Add(time.Minute))
	require.NoError(t, err)

	err = attempt.MarkProcessing(now.Add(2 * time.Minute))
	require.NoError(t, err)

	assert.Equal(t, attempt.Status, PaymentStatusProcessing)
	assert.Equal(t, attempt.NextAction.Type, NextActionTypeNone)
}

func TestMarkSucceeded_FromProcessing(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt := mustNewAttempt(t, now)

	err := attempt.MarkProcessing(now.Add(time.Minute))
	require.NoError(t, err)

	err = attempt.MarkSucceeded(now.Add(2 * time.Minute))
	require.NoError(t, err)

	assert.Equal(t, attempt.Status, PaymentStatusSucceeded)
	assert.NotNil(t, attempt.Timestamps.CompletedAt)
}

func TestMarkFailed_FromRequiresAction(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt := mustNewAttempt(t, now)

	err := attempt.MarkRequiresAction(NextAction{
		Type:        NextActionTypeRedirect,
		RedirectURL: "https://example.com/3ds",
	}, now.Add(time.Minute))
	require.NoError(t, err)

	err = attempt.MarkFailed("3ds authentication failed", now.Add(2*time.Minute))
	require.NoError(t, err)

	assert.Equal(t, attempt.Status, PaymentStatusFailed)
	assert.Equal(t, attempt.FailureReason, "3ds authentication failed")
}

func TestTerminalStateCannotTransition(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt := mustNewAttempt(t, now)

	err := attempt.MarkSucceeded(now.Add(time.Minute))
	require.NoError(t, err)

	err = attempt.MarkProcessing(now.Add(2 * time.Minute))
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestProviderPaymentID_WhenNotLinked(t *testing.T) {
	attempt := mustNewAttempt(t, time.Now())

	_, err := attempt.ProviderPaymentID()
	assert.ErrorIs(t, err, ErrProviderNotLinked)
}

func mustNewAttempt(t *testing.T, now time.Time) *PaymentAttempt {
	t.Helper()

	attempt, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
		"idempotency-key-123",
		"https://example.com/return",
		Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	return attempt
}

func TestLinkProvider_IdempotentWhenSameValues(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt := mustNewAttempt(t, now)

	err := attempt.LinkProvider("stripe", "pi_123", "secret_123", now.Add(time.Minute))
	require.NoError(t, err)

	updatedAt := attempt.Timestamps.UpdatedAt

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now.Add(2*time.Minute))
	require.NoError(t, err)

	assert.Equal(t, "stripe", attempt.Provider.ProviderName)
	assert.Equal(t, "pi_123", attempt.Provider.ProviderPaymentID)
	assert.Equal(t, "secret_123", attempt.Provider.ClientSecret)
	assert.Equal(t, updatedAt, attempt.Timestamps.UpdatedAt)
}

func TestLinkProvider_RejectsOverwrite(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt := mustNewAttempt(t, now)

	err := attempt.LinkProvider("stripe", "pi_123", "secret_123", now.Add(time.Minute))
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_999", "secret_999", now.Add(2*time.Minute))
	require.ErrorIs(t, err, ErrProviderAlreadyLinked)

	assert.Equal(t, "stripe", attempt.Provider.ProviderName)
	assert.Equal(t, "pi_123", attempt.Provider.ProviderPaymentID)
	assert.Equal(t, "secret_123", attempt.Provider.ClientSecret)
}

func TestMarkFailed_DefaultsToUnknownReasonWhenBlank(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt := mustNewAttempt(t, now)

	err := attempt.MarkFailed("", now.Add(time.Minute))
	require.NoError(t, err)

	assert.Equal(t, PaymentStatusFailed, attempt.Status)
	assert.Equal(t, FailureReasonUnknown, attempt.FailureReason)
	require.NotNil(t, attempt.Timestamps.CompletedAt)
}

func TestMarkFailed_DefaultsToUnknownReasonWhenWhitespace(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt := mustNewAttempt(t, now)

	err := attempt.MarkFailed("   ", now.Add(time.Minute))
	require.NoError(t, err)

	assert.Equal(t, FailureReasonUnknown, attempt.FailureReason)
}
