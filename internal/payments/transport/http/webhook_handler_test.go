package http

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

func TestStripeWebhook_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler(logger, "whsec_test_secret")

	req := httptest.NewRequest(http.MethodGet, "/webhooks/stripe", nil)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestStripeWebhook_InvalidSignature(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler(logger, "whsec_test_secret")

	payload := `{"id":"evt_test","object":"event","type":"payment_intent.succeeded"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(payload))
	req.Header.Set("Stripe-Signature", "invalid")

	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "invalid webhook signature")
}

func TestStripeWebhook_ValidSignature(t *testing.T) {
	t.Parallel()

	const secret = "whsec_test_secret"

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler(logger, secret)

	payload := fmt.Sprintf(`{
		"id":"evt_test",
		"object":"event",
		"type":"payment_intent.succeeded",
		"api_version":"%s"
	}`, stripe.APIVersion)

	signature := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  secret,
	})

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(payload))
	req.Header.Set("Stripe-Signature", signature.Header)

	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"received":true`)
	assert.Contains(t, res.Body.String(), `"id":"evt_test"`)
	assert.Contains(t, res.Body.String(), `"type":"payment_intent.succeeded"`)
}

func TestStripeWebhook_InvalidBodyReadStillHandled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler(logger, "whsec_test_secret")

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader("ignored"))
	req.Body = io.NopCloser(failingReader{})
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "invalid webhook body")
}

type failingReader struct{}
type errReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (errReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (errReader) Close() error {
	return nil
}
