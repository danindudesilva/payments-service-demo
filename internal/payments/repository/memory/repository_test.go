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

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		domain.Money{Amount: 2500, Currency: "gbp"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), "pa_123")
	require.NoError(t, err)

	assert.Equal(t, "pa_123", got.ID)
	assert.Equal(t, "GBP", got.Money.Currency)
	assert.Equal(t, domain.PaymentStatusPending, got.Status)
}

func TestRepository_GetByProviderPaymentID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now.Add(time.Minute))
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByProviderPaymentID(context.Background(), "pi_123")
	require.NoError(t, err)

	assert.Equal(t, "pa_123", got.ID)
	assert.Equal(t, "stripe", got.Provider.ProviderName)
	assert.Equal(t, "pi_123", got.Provider.ProviderPaymentID)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
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

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	attempt.Status = domain.PaymentStatusFailed
	attempt.FailureReason = "mutated after save"

	got, err := repo.GetByID(context.Background(), "pa_123")
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusPending, got.Status)
	assert.Empty(t, got.FailureReason)
}

func TestRepository_GetReturnsClone(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"pa_123",
		"order_123",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), "pa_123")
	require.NoError(t, err)

	got.Status = domain.PaymentStatusFailed

	gotAgain, err := repo.GetByID(context.Background(), "pa_123")
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusPending, gotAgain.Status)
}
