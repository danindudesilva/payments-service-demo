package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	memoryrepo "github.com/danindudesilva/payments-service/internal/payments/repository/memory"
	paymentservice "github.com/danindudesilva/payments-service/internal/payments/service"
	stripe "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

const testWebhookSecret = "whsec_test_secret"

type fakeProcessedWebhookEventRepo struct {
	processed map[string]bool
	saveCalls int
}

func newFakeProcessedWebhookEventRepo() *fakeProcessedWebhookEventRepo {
	return &fakeProcessedWebhookEventRepo{
		processed: make(map[string]bool),
	}
}

func (r *fakeProcessedWebhookEventRepo) SaveProcessedEvent(ctx context.Context, providerName, eventID, eventType string) error {
	key := fmt.Sprintf("%s:%s", providerName, eventID)
	r.processed[key] = true
	r.saveCalls++
	return nil
}

func (r *fakeProcessedWebhookEventRepo) HasProcessedEvent(ctx context.Context, providerName, eventID string) (bool, error) {
	key := fmt.Sprintf("%s:%s", providerName, eventID)
	return r.processed[key], nil
}

func newWebhookTestHandler(repo domain.PaymentAttemptRepository, now time.Time) *WebhookHandler {
	return newWebhookTestHandlerWithProcessedRepo(repo, now, newFakeProcessedWebhookEventRepo())
}

func newWebhookTestHandlerWithProcessedRepo(
	repo domain.PaymentAttemptRepository,
	now time.Time,
	processedRepo domain.ProcessedWebhookEventRepository,
) *WebhookHandler {
	svc := paymentservice.New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(
				ctx context.Context,
				request domain.CreateProviderPaymentRequest,
			) (
				domain.CreateProviderPaymentResult,
				error,
			) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(
				ctx context.Context,
				providerPaymentID string,
			) (
				domain.CreateProviderPaymentResult,
				error,
			) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		func() time.Time { return now },
		func() string { return "unused" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewWebhookHandler(logger, testWebhookSecret, svc, processedRepo)
}

func newWebhookTestHandlerWithDefaults() *WebhookHandler {
	repo := memoryrepo.NewRepository()
	processedRepo := newFakeProcessedWebhookEventRepo()
	svc := paymentservice.New(
		repo,
		&fakeGateway{
			createPaymentFunc: func(
				ctx context.Context,
				request domain.CreateProviderPaymentRequest,
			) (
				domain.CreateProviderPaymentResult,
				error,
			) {
				return domain.CreateProviderPaymentResult{}, nil
			},
			getPaymentFunc: func(
				ctx context.Context,
				providerPaymentID string,
			) (
				domain.CreateProviderPaymentResult,
				error,
			) {
				return domain.CreateProviderPaymentResult{}, nil
			},
		},
		time.Now,
		func() string { return "unused" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewWebhookHandler(logger, testWebhookSecret, svc, processedRepo)
}

func newSignedWebhookRequest(eventType string, dataObject string) *http.Request {
	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"%s",
		"api_version":"%s",
		"data":{"object":%s}
	}`, eventType, stripe.APIVersion, dataObject)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  testWebhookSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req.Header.Set("Stripe-Signature", signature.Header)
	return req
}
