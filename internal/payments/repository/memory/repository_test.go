package memory

import (
	"context"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_SaveAndGetByID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	attempt := mustAttempt(t)
	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, attempt.ID, got.ID)
	assert.Equal(t, domain.PaymentStatusPending, got.Status)
	assert.Equal(t, "GBP", got.Money.Currency)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_GetByProviderPaymentID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	attempt := mustAttempt(t)
	now := time.Date(2026, 3, 14, 12, 1, 0, 0, time.UTC)

	err := attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(context.Background(), "pi_123")
	require.NoError(t, err)

	assert.Equal(t, attempt.ID, got.ID)
	assert.Equal(t, "stripe", got.Provider.ProviderName)
	assert.Equal(t, "pi_123", got.Provider.ProviderPaymentID)
}

func TestRepository_GetByProviderPaymentID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByProviderPaymentID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_SaveStoresClone(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	attempt := mustAttempt(t)

	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	attempt.Status = domain.PaymentStatusFailed
	attempt.FailureReason = "mutated after save"

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusPending, got.Status)
	assert.Empty(t, got.FailureReason)
}

func TestRepository_GetReturnsClone(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	attempt := mustAttempt(t)

	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	got.Status = domain.PaymentStatusFailed

	gotAgain, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusPending, gotAgain.Status)
}

func TestRepository_SaveAndGetTerminalAttemptPreservesCompletedAt(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.MarkSucceeded(now.Add(time.Minute))
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	require.NotNil(t, got.Timestamps.CompletedAt)
	assert.Equal(t, now.Add(time.Minute).UTC(), *got.Timestamps.CompletedAt)
}

func mustAttempt(t *testing.T) *domain.PaymentAttempt {
	t.Helper()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "gbp"},
		now,
	)
	require.NoError(t, err)

	return attempt
}
