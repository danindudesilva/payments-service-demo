package http

import (
	"io"
	"log/slog"
	"net/http"

	basehttp "github.com/danindudesilva/payments-service/internal/httpserver"
	"github.com/stripe/stripe-go/v84/webhook"
)

type WebhookHandler struct {
	logger        *slog.Logger
	webhookSecret string
}

func NewWebhookHandler(logger *slog.Logger, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		logger:        logger,
		webhookSecret: webhookSecret,
	}
}

func (h *WebhookHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/webhooks/stripe", h.handleStripeWebhook)
}

func (h *WebhookHandler) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		basehttp.WriteMethodNotAllowed(w)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("read stripe webhook body failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "invalid webhook body",
		})
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signature, h.webhookSecret)
	if err != nil {
		h.logger.Error("stripe webhook signature verification failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "invalid webhook signature",
		})
		return
	}

	h.logger.Info("stripe webhook received",
		slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
		slog.String("event_id", event.ID),
		slog.String("event_type", string(event.Type)),
	)

	basehttp.WriteJSON(w, http.StatusOK, map[string]any{
		"received": true,
		"id":       event.ID,
		"type":     string(event.Type),
	})
}
