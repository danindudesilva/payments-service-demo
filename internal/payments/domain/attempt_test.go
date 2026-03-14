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
		Money{Amount: 0, Currency: "GBP"},
		time.Now(),
	)
	assert.ErrorIs(t, err, ErrInvalidMoney)
}

func TestLinkProvider(t *testing.T) {
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	attempt, err := NewPaymentAttempt(
		"pa_123",
		"order_123",
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
		Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	return attempt
}
