package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/danindudesilva/payments-service/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_SaveAndGetByID(t *testing.T) {
	pool := newTestPool(t)
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
	pool := newTestPool(t)
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
	pool := newTestPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_GetByProviderPaymentID_NotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	_, err := repo.GetByProviderPaymentID(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestRepository_SaveUpdatesExistingAttempt(t *testing.T) {
	pool := newTestPool(t)
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

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable"
	}

	pool, err := database.NewPool(context.Background(), database.Config{
		DatabaseURL: databaseURL,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM payment_attempts`)
		pool.Close()
	})

	return pool
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
