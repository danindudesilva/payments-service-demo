package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	memoryrepo "github.com/danindudesilva/payments-service/internal/payments/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

func TestStripeWebhook_PaymentIntentSucceededUpdatesAttempt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_123",
		"order_123",
		"idem_123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	handler := newWebhookTestHandler(repo, now.Add(time.Minute))
	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"payment_intent.succeeded",
		"api_version":"%s",
		"data":{
			"object":{
				"id":"pi_123",
				"object":"payment_intent"
			}
		}
	}`, stripe.APIVersion)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  testWebhookSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req.Header.Set("Stripe-Signature", signature.Header)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	got, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSucceeded, got.Status)
	require.NotNil(t, got.Timestamps.CompletedAt)
}

func TestStripeWebhook_PaymentIntentFailedUpdatesAttempt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	repo := memoryrepo.NewRepository()

	attempt, err := domain.NewPaymentAttempt(
		"attempt_123",
		"order_123",
		"idem_123",
		"https://example.com/return",
		domain.Money{Amount: 2500, Currency: "GBP"},
		now,
	)
	require.NoError(t, err)

	err = attempt.LinkProvider("stripe", "pi_123", "secret_123", now)
	require.NoError(t, err)

	err = repo.Save(context.Background(), attempt)
	require.NoError(t, err)

	handler := newWebhookTestHandler(repo, now.Add(time.Minute))
	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"payment_intent.payment_failed",
		"api_version":"%s",
		"data":{
			"object":{
				"id":"pi_123",
				"object":"payment_intent"
			}
		}
	}`, stripe.APIVersion)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  testWebhookSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req.Header.Set("Stripe-Signature", signature.Header)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	got, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusFailed, got.Status)
	assert.Equal(t, domain.FailureReasonProviderReportedFailed, got.FailureReason)
}

func TestStripeWebhook_UnhandledEventTypeIsIgnored(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"charge.refunded",
		"api_version":"%s",
		"data":{"object":{"id":"ch_123","object":"charge"}}
	}`, stripe.APIVersion)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  testWebhookSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req.Header.Set("Stripe-Signature", signature.Header)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"received":true`)
}

func TestStripeWebhook_InvalidSignedPayloadReturnsBadRequest(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"payment_intent.succeeded",
		"api_version":"%s",
		"data":{"object":"not-an-object"}
	}`, stripe.APIVersion)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  testWebhookSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req.Header.Set("Stripe-Signature", signature.Header)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "invalid webhook payload or signature")
}

func TestStripeWebhook_ValidSignedHandledEventWithMissingAttemptReturnsServerError(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	req := newSignedWebhookRequest("payment_intent.succeeded", `{"id":"pi_missing","object":"payment_intent"}`)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, res.Body.String(), "failed to process webhook event")
}
