package postgres

import (
	"errors"
	"fmt"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/jackc/pgx/v5"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAttempt(row rowScanner) (*domain.PaymentAttempt, error) {
	var (
		id                string
		orderID           string
		returnURL         string
		status            string
		amount            int64
		currency          string
		providerName      string
		providerPaymentID *string
		clientSecret      string
		failureReason     string
		createdAt         time.Time
		updatedAt         time.Time
		completedAt       *time.Time
	)

	err := row.Scan(
		&id,
		&orderID,
		&returnURL,
		&status,
		&amount,
		&currency,
		&providerName,
		&providerPaymentID,
		&clientSecret,
		&failureReason,
		&createdAt,
		&updatedAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("scan payment attempt: %w", err)
	}

	attempt, err := domain.NewPaymentAttempt(
		id,
		orderID,
		returnURL,
		domain.Money{
			Amount:   amount,
			Currency: currency,
		},
		createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("rebuild payment attempt: %w", err)
	}

	attempt.Status = domain.PaymentStatus(status)
	attempt.Provider.ProviderName = providerName
	if providerPaymentID != nil {
		attempt.Provider.ProviderPaymentID = *providerPaymentID
	}
	attempt.Provider.ClientSecret = clientSecret
	attempt.FailureReason = failureReason
	attempt.Timestamps.CreatedAt = createdAt
	attempt.Timestamps.UpdatedAt = updatedAt
	attempt.Timestamps.CompletedAt = completedAt

	return attempt, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
