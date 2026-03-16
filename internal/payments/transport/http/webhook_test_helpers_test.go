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

func newWebhookTestHandler(repo domain.PaymentAttemptRepository, t time.Time) *WebhookHandler {
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
		func() time.Time { return t },
		func() string { return "unused" },
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewWebhookHandler(logger, testWebhookSecret, svc)
}

func newWebhookTestHandlerWithDefaults() *WebhookHandler {
	repo := memoryrepo.NewRepository()
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
	return NewWebhookHandler(logger, testWebhookSecret, svc)
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
