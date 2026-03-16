package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	memoryrepo "github.com/danindudesilva/payments-service/internal/payments/repository/memory"
	"github.com/danindudesilva/payments-service/internal/payments/repository/postgres"
	"github.com/danindudesilva/payments-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreatePaymentAttempt_RequiresAction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			assert.Equal(t, "attempt_123", request.AttemptID)
			assert.Equal(t, "order_123", request.OrderID)
			assert.Equal(t, int64(2500), request.Money.Amount)
			assert.Equal(t, "GBP", request.Money.Currency)

			return domain.CreateProviderPaymentResult{
				ProviderName:      "stripe",
				ProviderPaymentID: "pi_123",
				ClientSecret:      "secret_123",
				Status:            domain.PaymentStatusRequiresAction,
				NextAction: domain.NextAction{
					Type:        domain.NextActionTypeRedirect,
					RedirectURL: "https://example.com/3ds",
				},
			}, nil
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		func() time.Time { return now },
		func() string { return "attempt_123" },
	)

	result, err := svc.CreatePaymentAttempt(context.Background(), CreatePaymentAttemptInput{
		OrderID:        "order_123",
		IdempotencyKey: "idempotency-key-123",
		Amount:         2500,
		Currency:       "gbp",
		ReturnURL:      "https://example.com/return",
		Description:    "Test payment",
	})
	require.NoError(t, err)

	require.NotNil(t, result)
	require.NotNil(t, result.Attempt)
	assert.Equal(t, "attempt_123", result.Attempt.ID)
	assert.Equal(t, domain.PaymentStatusRequiresAction, result.Attempt.Status)
	assert.Equal(t, "stripe", result.Attempt.Provider.ProviderName)
	assert.Equal(t, "pi_123", result.Attempt.Provider.ProviderPaymentID)
	assert.Equal(t, domain.NextActionTypeRedirect, result.Attempt.NextAction.Type)

	saved, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusRequiresAction, saved.Status)
}

func TestService_CreatePaymentAttempt_Succeeded(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	pool := testutil.NewTestPool(t)
	repo := postgres.NewRepository(pool)

	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{
				ProviderName:      "stripe",
				ProviderPaymentID: "pi_456",
				ClientSecret:      "secret_456",
				Status:            domain.PaymentStatusSucceeded,
			}, nil
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		func() time.Time { return now },
		func() string { return "attempt_456" },
	)

	result, err := svc.CreatePaymentAttempt(context.Background(), CreatePaymentAttemptInput{
		OrderID:        "order_456",
		IdempotencyKey: "idempotency-key-123",
		Amount:         5000,
		Currency:       "GBP",
		ReturnURL:      "https://example.com/return",
	})
	require.NoError(t, err)

	require.NotNil(t, result)
	assert.Equal(t, domain.PaymentStatusSucceeded, result.Attempt.Status)
	require.NotNil(t, result.Attempt.Timestamps.CompletedAt)
}

func TestService_CreatePaymentAttempt_GatewayError(t *testing.T) {
	t.Parallel()

	repo := memoryrepo.NewRepository()

	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, errors.New("gateway unavailable")
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		time.Now,
		func() string { return "attempt_789" },
	)

	result, err := svc.CreatePaymentAttempt(context.Background(), CreatePaymentAttemptInput{
		OrderID:        "order_789",
		IdempotencyKey: "idempotency-key-123",
		Amount:         3000,
		Currency:       "GBP",
	})
	require.Error(t, err)
	assert.Nil(t, result)

	_, getErr := repo.GetByID(context.Background(), "attempt_789")
	require.ErrorIs(t, getErr, domain.ErrPaymentNotFound)
}

func TestService_CreatePaymentAttempt_IsIdempotent(t *testing.T) {
	repo := memoryrepo.NewRepository()
	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

	createCalls := 0
	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			createCalls++
			return domain.CreateProviderPaymentResult{
				ProviderName:      "stripe",
				ProviderPaymentID: "pi_123",
				ClientSecret:      "secret_123",
				Status:            domain.PaymentStatusPending,
			}, nil
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		func() time.Time { return now },
		func() string { return "attempt_123" },
	)

	input := CreatePaymentAttemptInput{
		OrderID:        "order_123",
		IdempotencyKey: "idem_123",
		Amount:         2500,
		Currency:       "GBP",
		ReturnURL:      "https://example.com/return",
		Description:    "test payment",
	}

	first, err := svc.CreatePaymentAttempt(context.Background(), input)
	require.NoError(t, err)

	second, err := svc.CreatePaymentAttempt(context.Background(), input)
	require.NoError(t, err)

	require.Equal(t, 1, createCalls)
	assert.Equal(t, first.Attempt.ID, second.Attempt.ID)
	assert.Equal(t, "idem_123", second.Attempt.IdempotencyKey)
}

func TestService_GetPaymentAttempt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_001",
		"order_001",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 1200, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	svc := New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		time.Now,
		func() string { return "unused" },
	)

	got, err := svc.GetPaymentAttempt(context.Background(), "attempt_001")
	require.NoError(t, err)
	assert.Equal(t, "attempt_001", got.ID)
	assert.Equal(t, domain.PaymentStatusPending, got.Status)
}

func TestService_CreatePaymentAttempt_FailedUsesProviderFailureReasonConstant(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{
				ProviderName:      "stripe",
				ProviderPaymentID: "pi_failed_123",
				ClientSecret:      "secret_failed_123",
				Status:            domain.PaymentStatusFailed,
			}, nil
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		func() time.Time { return now },
		func() string { return "attempt_failed_123" },
	)

	result, err := svc.CreatePaymentAttempt(context.Background(), CreatePaymentAttemptInput{
		OrderID:        "order_failed_123",
		IdempotencyKey: "idempotency-key-123",
		Amount:         2500,
		Currency:       "GBP",
		ReturnURL:      "https://example.com/return",
	})
	require.NoError(t, err)

	require.NotNil(t, result)
	assert.Equal(t, domain.PaymentStatusFailed, result.Attempt.Status)
	assert.Equal(t, domain.FailureReasonProviderReportedFailed, result.Attempt.FailureReason)
	require.NotNil(t, result.Attempt.Timestamps.CompletedAt)
}

func TestService_ReconcilePaymentAttempt_UpdatesStatusFromProvider(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_123",
		"order_123",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	gateway := &fakeGateway{
		createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
			return domain.CreateProviderPaymentResult{}, nil
		},
		getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
			require.Equal(t, "pi_123", providerPaymentID)

			return domain.CreateProviderPaymentResult{
				ProviderName:      "stripe",
				ProviderPaymentID: "pi_123",
				ClientSecret:      "secret_123",
				Status:            domain.PaymentStatusSucceeded,
			}, nil
		},
	}

	svc := New(
		repo,
		gateway,
		func() time.Time { return now.Add(time.Minute) },
		func() string { return "unused" },
	)

	got, err := svc.ReconcilePaymentAttempt(context.Background(), "attempt_123")
	require.NoError(t, err)

	assert.Equal(t, domain.PaymentStatusSucceeded, got.Status)
	require.NotNil(t, got.Timestamps.CompletedAt)

	saved, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSucceeded, saved.Status)
}

func TestService_ReconcilePaymentAttempt_NotFound(t *testing.T) {
	t.Parallel()

	repo := memoryrepo.NewRepository()

	svc := New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		time.Now,
		func() string { return "unused" },
	)

	_, err := svc.ReconcilePaymentAttempt(context.Background(), "missing")
	require.ErrorIs(t, err, domain.ErrPaymentNotFound)
}

func TestService_ReconcilePaymentAttempt_RequiresProviderLink(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_123",
		"order_123",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	svc := New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		time.Now,
		func() string { return "unused" },
	)

	_, err = svc.ReconcilePaymentAttempt(context.Background(), "attempt_123")
	require.ErrorIs(t, err, domain.ErrProviderNotLinked)
}
