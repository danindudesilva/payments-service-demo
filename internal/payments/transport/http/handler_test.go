package http

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	memoryrepo "github.com/danindudesilva/payments-service/internal/payments/repository/memory"
	paymentservice "github.com/danindudesilva/payments-service/internal/payments/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGateway struct {
	createPaymentFunc func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error)
	getPaymentFunc    func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error)
}

func (f *fakeGateway) CreatePayment(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
	return f.createPaymentFunc(ctx, request)
}

func (f *fakeGateway) GetPayment(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
	return f.getPaymentFunc(ctx, providerPaymentID)
}

func TestCreatePaymentAttemptRoutes(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts", bytes.NewBufferString(`{
		"order_id":"order_123",
		"amount":2500,
		"currency":"gbp",
		"return_url":"https://example.com/return",
		"description":"test payment"
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	assert.Contains(t, res.Body.String(), `"id":"attempt_123"`)
	assert.Contains(t, res.Body.String(), `"status":"requires_action"`)
	assert.Contains(t, res.Body.String(), `"currency":"GBP"`)
	assert.Contains(t, res.Body.String(), `"payment_id":"pi_123"`)
}

func TestCreatePaymentAttemptRoutes_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts", bytes.NewBufferString(`{`))
	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "invalid json body")
}

func TestGetPaymentAttemptRoutes(t *testing.T) {
	t.Parallel()

	handler, repo := newTestHandlerWithRepo(t)

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"attempt_existing",
		"order_existing",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 4200, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/attempt_existing", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"id":"attempt_existing"`)
	assert.Contains(t, res.Body.String(), `"order_id":"order_existing"`)
}

func TestGetPaymentAttemptRoutes_NotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/missing", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	assert.Contains(t, res.Body.String(), "payment attempt not found")
}

func TestPaymentAttemptRoutes_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestPaymentAttemptByID_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts/attempt_123", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestPaymentAttemptRoutes_MissingID(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "payment attempt id is required")
}

func TestCreatePaymentAttemptRoutes_UnknownErrorReturnsSafeMessage(t *testing.T) {
	t.Parallel()

	service := paymentservice.New(
		memoryrepo.NewRepository(),
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, assert.AnError
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		func() time.Time { return time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC) },
		func() string { return "attempt_123" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(service, logger)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts", bytes.NewBufferString(`{
		"order_id":"order_123",
		"amount":2500,
		"currency":"gbp",
		"return_url":"https://example.com/return",
		"description":"test payment"
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, res.Body.String(), `"error":"internal server error"`)
	assert.NotContains(t, res.Body.String(), assert.AnError.Error())
}

func TestReconcilePaymentAttempt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_existing",
		"order_existing",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 4200, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	service := paymentservice.New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{
					ProviderName:      "stripe",
					ProviderPaymentID: providerPaymentID,
					ClientSecret:      "secret_123",
					Status:            domain.PaymentStatusSucceeded,
				}, nil
			},
		},
		func() time.Time { return now.Add(time.Minute) },
		func() string { return "unused" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(service, logger)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts/attempt_existing/reconcile", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"status":"succeeded"`)
}

func TestReconcilePaymentAttempt_NotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts/missing/reconcile", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	assert.Contains(t, res.Body.String(), "payment attempt not found")
}

func TestHandlePaymentAttemptRoutes_UnknownSubroute(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/attempt_123/unknown", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	assert.Contains(t, res.Body.String(), `"error":"not found"`)
}

func TestCreatePaymentAttempt_InvalidReturnURL(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts", bytes.NewBufferString(`{
		"order_id":"order_123",
		"amount":2500,
		"currency":"gbp",
		"return_url":"not-a-url",
		"description":"test payment"
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "return_url")
}

func TestHandlePaymentAttemptRoutes_GetByID(t *testing.T) {
	t.Parallel()

	handler, repo := newTestHandlerWithRepo(t)

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	attempt, err := domain.NewPaymentAttempt(
		"attempt_existing",
		"order_existing",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 4200, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/attempt_existing", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"id":"attempt_existing"`)
	assert.Contains(t, res.Body.String(), `"order_id":"order_existing"`)
}

func TestHandlePaymentAttemptRoutes_Reconcile(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_existing",
		"order_existing",
		"idempotency-key-123",
		"https://example.com/return",
		domain.Money{Amount: 4200, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	service := paymentservice.New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(ctx context.Context, providerPaymentID string) (domain.CreateProviderPaymentResult, error) {
				return domain.CreateProviderPaymentResult{
					ProviderName:      "stripe",
					ProviderPaymentID: providerPaymentID,
					ClientSecret:      "secret_123",
					Status:            domain.PaymentStatusSucceeded,
				}, nil
			},
		},
		func() time.Time { return now.Add(time.Minute) },
		func() string { return "unused" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(service, logger)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts/attempt_existing/reconcile", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"id":"attempt_existing"`)
	assert.Contains(t, res.Body.String(), `"status":"succeeded"`)
}

func TestHandlePaymentAttemptRoutes_GetOnReconcileMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/payment-attempts/attempt_123/reconcile", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestHandlePaymentAttemptRoutes_PostOnAttemptIDMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts/attempt_123", nil)
	res := httptest.NewRecorder()

	handler.handlePaymentAttemptRoutes(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestCreatePaymentAttempt_MissingReturnURL(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/payment-attempts", bytes.NewBufferString(`{
		"order_id":"order_123",
		"amount":2500,
		"currency":"gbp",
		"return_url":"",
		"description":"test payment"
	}`))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	handler.handlePaymentAttempts(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "return_url is required")
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()

	handler, _ := newTestHandlerWithRepo(t)
	return handler
}

func newTestHandlerWithRepo(t *testing.T) (*Handler, *memoryrepo.Repository) {
	t.Helper()

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	service := paymentservice.New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(ctx context.Context, request domain.CreateProviderPaymentRequest) (domain.CreateProviderPaymentResult, error) {
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
		},
		func() time.Time { return now },
		func() string { return "attempt_123" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(service, logger), repo
}
