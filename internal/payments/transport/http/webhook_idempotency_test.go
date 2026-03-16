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

func TestStripeWebhook_DuplicateDeliveryIsIgnored(t *testing.T) {
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

	processedRepo := newFakeProcessedWebhookEventRepo()
	handler := newWebhookTestHandlerWithProcessedRepo(repo, now.Add(time.Minute), processedRepo)

	payload := fmt.Sprintf(`{
		"id":"evt_duplicate",
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

	req1 := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req1.Header.Set("Stripe-Signature", signature.Header)
	res1 := httptest.NewRecorder()

	handler.handleStripeWebhook(res1, req1)

	require.Equal(t, http.StatusOK, res1.Code)

	got, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSucceeded, got.Status)
	require.NotNil(t, got.Timestamps.CompletedAt)

	firstCompletedAt := *got.Timestamps.CompletedAt

	req2 := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewBufferString(payload))
	req2.Header.Set("Stripe-Signature", signature.Header)
	res2 := httptest.NewRecorder()

	handler.handleStripeWebhook(res2, req2)

	require.Equal(t, http.StatusOK, res2.Code)
	assert.Contains(t, res2.Body.String(), `"already_processed":true`)

	gotAgain, err := repo.GetByID(context.Background(), "attempt_123")
	require.NoError(t, err)
	require.NotNil(t, gotAgain.Timestamps.CompletedAt)
	assert.Equal(t, firstCompletedAt, *gotAgain.Timestamps.CompletedAt)
	assert.Equal(t, 1, processedRepo.saveCalls)
}
