package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/danindudesilva/payments-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_SaveAndGetByID(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, attempt.ID, got.ID)
	assert.Equal(t, attempt.OrderID, got.OrderID)
	assert.Equal(t, attempt.ReturnURL, got.ReturnURL)
	assert.Equal(t, attempt.Money.Amount, got.Money.Amount)
	assert.Equal(t, attempt.Money.Currency, got.Money.Currency)
	assert.Equal(t, attempt.Status, got.Status)
}

func TestRepository_SaveAndGetByProviderPaymentID(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := attempt.LinkProvider("stripe", "pi_123", "secret_123", time.Now())
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(context.Background(), "pi_123")
	require.NoError(t, err)

	assert.Equal(t, attempt.ID, got.ID)
	assert.Equal(t, "stripe", got.Provider.ProviderName)
	assert.Equal(t, "pi_123", got.Provider.ProviderPaymentID)
	assert.Equal(t, "secret_123", got.Provider.ClientSecret)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_GetByProviderPaymentID_NotFound(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByProviderPaymentID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_SaveUpdatesExistingAttempt(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	err = attempt.MarkSucceeded(time.Now())
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusSucceeded, got.Status)
	require.NotNil(t, got.Timestamps.CompletedAt)
}

func TestRepository_SaveAndGetByID_PreservesReturnURL(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/return", got.ReturnURL)
}

func TestRepository_SaveAndGetByID_PreservesFailureReason(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := attempt.MarkFailed(domain.FailureReasonUnknown, time.Now())
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusFailed, got.Status)
	assert.Equal(t, domain.FailureReasonUnknown, got.FailureReason)
	require.NotNil(t, got.Timestamps.CompletedAt)
}

func TestRepository_SaveAndGetByID_PreservesCompletedAt(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"attempt_completed",
		"order_completed",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.MarkSucceeded(now.Add(time.Minute))
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), "attempt_completed")
	require.NoError(t, err)

	require.NotNil(t, got.Timestamps.CompletedAt)
	assert.Equal(t, now.Add(time.Minute).Local(), *got.Timestamps.CompletedAt)
}

func TestRepository_SaveAndGetByID_PreservesClientSecret(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := attempt.LinkProvider("stripe", "pi_secret", "secret_123", time.Now())
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), attempt.ID)
	require.NoError(t, err)

	assert.Equal(t, "secret_123", got.Provider.ClientSecret)
}

func TestRepository_GetByProviderPaymentID_PreservesReturnURLAndStatus(t *testing.T) {
	pool := testutil.NewTestPool(t)
	repo := NewRepository(pool)

	attempt := mustNewAttempt(t)
	err := attempt.LinkProvider("stripe", "pi_lookup", "secret_lookup", time.Now())
	require.NoError(t, err)

	err = attempt.MarkProcessing(time.Now())
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(context.Background(), "pi_lookup")
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/return", got.ReturnURL)
	assert.Equal(t, domain.PaymentStatusProcessing, got.Status)
}

func mustNewAttempt(t *testing.T) *domain.PaymentAttempt {
	t.Helper()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"attempt_123",
		"order_123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	return attempt
}
