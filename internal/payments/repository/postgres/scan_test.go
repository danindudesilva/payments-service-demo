package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRow struct {
	values []any
	err    error
}

func (f fakeRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}

	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = f.values[i].(string)
		case **string:
			if f.values[i] == nil {
				*d = nil
			} else {
				v := f.values[i].(string)
				*d = &v
			}
		case *int64:
			*d = f.values[i].(int64)
		case *time.Time:
			*d = f.values[i].(time.Time)
		case **time.Time:
			if f.values[i] == nil {
				*d = nil
			} else {
				v := f.values[i].(time.Time)
				*d = &v
			}
		default:
			return errors.New("unsupported scan destination")
		}
	}

	return nil
}

func TestScanAttempt(t *testing.T) {
	createdAt := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	completedAt := createdAt.Add(2 * time.Minute)

	row := fakeRow{
		values: []any{
			"attempt_123",
			"order_123",
			"https://example.com/return",
			"failed",
			int64(2500),
			"GBP",
			"stripe",
			"pi_123",
			"secret_123",
			domain.FailureReasonUnknown,
			createdAt,
			updatedAt,
			completedAt,
		},
	}

	attempt, err := scanAttempt(row)
	require.NoError(t, err)

	assert.Equal(t, "attempt_123", attempt.ID)
	assert.Equal(t, "order_123", attempt.OrderID)
	assert.Equal(t, "https://example.com/return", attempt.ReturnURL)
	assert.Equal(t, domain.PaymentStatusFailed, attempt.Status)
	assert.Equal(t, int64(2500), attempt.Money.Amount)
	assert.Equal(t, "GBP", attempt.Money.Currency)
	assert.Equal(t, "stripe", attempt.Provider.ProviderName)
	assert.Equal(t, "pi_123", attempt.Provider.ProviderPaymentID)
	assert.Equal(t, "secret_123", attempt.Provider.ClientSecret)
	assert.Equal(t, domain.FailureReasonUnknown, attempt.FailureReason)
	assert.Equal(t, createdAt, attempt.Timestamps.CreatedAt)
	assert.Equal(t, updatedAt, attempt.Timestamps.UpdatedAt)
	require.NotNil(t, attempt.Timestamps.CompletedAt)
	assert.Equal(t, completedAt, *attempt.Timestamps.CompletedAt)
}

func TestScanAttempt_NotFound(t *testing.T) {
	_, err := scanAttempt(fakeRow{err: pgx.ErrNoRows})
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestScanAttempt_ScanError(t *testing.T) {
	_, err := scanAttempt(fakeRow{err: errors.New("boom")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan payment attempt")
}

func TestNullIfEmpty(t *testing.T) {
	assert.Nil(t, nullIfEmpty(""))

	value := nullIfEmpty("pi_123")
	assert.Equal(t, "pi_123", value)
}
