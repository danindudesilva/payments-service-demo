package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripeWebhook_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	req := httptest.NewRequest(http.MethodGet, "/webhooks/stripe", nil)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Contains(t, res.Body.String(), "method not allowed")
}

func TestStripeWebhook_InvalidSignature(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	payload := `{"id":"evt_test","object":"event","type":"payment_intent.succeeded"}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(payload))
	req.Header.Set("Stripe-Signature", "invalid")

	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	assert.Contains(t, res.Body.String(), "invalid webhook payload or signature")
}

func TestStripeWebhook_ValidSignature(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

	req := newSignedWebhookRequest("charge.refunded", `{"id":"ch_123","object":"charge"}`)
	res := httptest.NewRecorder()

	handler.handleStripeWebhook(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"received":true`)
	assert.Contains(t, res.Body.String(), `"id":"evt_test"`)
	assert.Contains(t, res.Body.String(), `"type":"charge.refunded"`)
}

func TestStripeWebhook_InvalidBodyReadStillHandled(t *testing.T) {
	t.Parallel()

	handler := newWebhookTestHandlerWithDefaults()

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
