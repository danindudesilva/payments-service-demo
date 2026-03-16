package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	basehttp "github.com/danindudesilva/payments-service/internal/httpserver"
	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/danindudesilva/payments-service/internal/payments/service"
	stripe "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type WebhookHandler struct {
	logger             *slog.Logger
	webhookSecret      string
	service            *service.Service
	processedEventRepo domain.ProcessedWebhookEventRepository
}

func NewWebhookHandler(
	logger *slog.Logger,
	webhookSecret string,
	service *service.Service,
	processedEventRepo domain.ProcessedWebhookEventRepository,
) *WebhookHandler {
	return &WebhookHandler{
		logger:             logger,
		webhookSecret:      webhookSecret,
		service:            service,
		processedEventRepo: processedEventRepo,
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
		h.logger.Error("stripe webhook verification failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "invalid webhook payload or signature",
		})
		return
	}

	alreadyProcessed, err := h.processedEventRepo.HasProcessedEvent(r.Context(), "stripe", event.ID)
	if err != nil {
		h.logger.Error("check processed stripe webhook event failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("event_id", event.ID),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusInternalServerError, basehttp.ErrorResponse{
			Error: "failed to process webhook event",
		})

		return
	}

	if alreadyProcessed {
		h.logger.Info("stripe webhook already processed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("event_id", event.ID),
			slog.String("event_type", string(event.Type)),
		)

		basehttp.WriteJSON(w, http.StatusOK, map[string]any{
			"received":          true,
			"id":                event.ID,
			"type":              string(event.Type),
			"already_processed": true,
		})

		return
	}

	if err := h.processStripeEvent(r, event); err != nil {
		h.logger.Error("stripe webhook processing failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("event_id", event.ID),
			slog.String("event_type", string(event.Type)),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusInternalServerError, basehttp.ErrorResponse{
			Error: "failed to process webhook event",
		})

		return
	}

	if err := h.processedEventRepo.SaveProcessedEvent(r.Context(), "stripe", event.ID, string(event.Type)); err != nil {
		h.logger.Error("save processed stripe webhook event failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("event_id", event.ID),
			slog.String("event_type", string(event.Type)),
			slog.String("error", err.Error()),
		)

		basehttp.WriteJSON(w, http.StatusInternalServerError, basehttp.ErrorResponse{
			Error: "failed to process webhook event",
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

func (h *WebhookHandler) processStripeEvent(r *http.Request, event stripe.Event) error {
	switch event.Type {
	case "payment_intent.succeeded":
		return h.applyPaymentIntentEvent(r, event, domain.PaymentStatusSucceeded, "")

	case "payment_intent.payment_failed":
		return h.applyPaymentIntentEvent(r, event, domain.PaymentStatusFailed, domain.FailureReasonProviderReportedFailed)

	case "payment_intent.processing":
		return h.applyPaymentIntentEvent(r, event, domain.PaymentStatusProcessing, "")

	case "payment_intent.canceled":
		return h.applyPaymentIntentEvent(r, event, domain.PaymentStatusCancelled, "")

	default:
		return nil
	}
}

func (h *WebhookHandler) applyPaymentIntentEvent(
	r *http.Request,
	event stripe.Event,
	status domain.PaymentStatus,
	failureReason string,
) error {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		return err
	}

	h.logger.Info("applying stripe payment intent webhook update",
		slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
		slog.String("event_id", event.ID),
		slog.String("provider_payment_id", intent.ID),
		slog.String("status", string(status)),
	)
	_, err := h.service.ApplyProviderPaymentUpdate(r.Context(), service.ProviderPaymentUpdate{
		ProviderPaymentID: intent.ID,
		Status:            status,
		FailureReason:     failureReason,
	})

	return err
}
