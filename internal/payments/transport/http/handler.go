package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	basehttp "github.com/danindudesilva/payments-service/internal/httpserver"
	"github.com/danindudesilva/payments-service/internal/payments/service"
)

type Handler struct {
	service *service.Service
	logger  *slog.Logger
}

func NewHandler(service *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/payment-attempts", h.handlePaymentAttempts)
	mux.HandleFunc("/payment-attempts/", h.handlePaymentAttemptRoutes)
}

func (h *Handler) handlePaymentAttempts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createPaymentAttempt(w, r)
	default:
		basehttp.WriteMethodNotAllowed(w)
	}
}

func (h *Handler) handlePaymentAttemptRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/payment-attempts/")
	path = strings.Trim(path, "/")

	if path == "" {
		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "payment attempt id is required",
		})

		return
	}

	parts := strings.Split(path, "/")
	attemptID := strings.TrimSpace(parts[0])

	if attemptID == "" {
		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "payment attempt id is required",
		})

		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getPaymentAttempt(w, r, attemptID)
		default:
			basehttp.WriteMethodNotAllowed(w)
		}

		return
	}

	if len(parts) == 2 && parts[1] == "reconcile" {
		switch r.Method {
		case http.MethodPost:
			h.reconcilePaymentAttempt(w, r, attemptID)
		default:
			basehttp.WriteMethodNotAllowed(w)
		}

		return
	}

	basehttp.WriteJSON(w, http.StatusNotFound, basehttp.ErrorResponse{
		Error: "not found",
	})
}

func (h *Handler) createPaymentAttempt(w http.ResponseWriter, r *http.Request) {
	var request createPaymentAttemptRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: "invalid json body",
		})
		return
	}

	if err := validateReturnURL(request.ReturnURL); err != nil {
		basehttp.WriteJSON(w, http.StatusBadRequest, basehttp.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	result, err := h.service.CreatePaymentAttempt(r.Context(), service.CreatePaymentAttemptInput{
		OrderID:     request.OrderID,
		Amount:      request.Amount,
		Currency:    request.Currency,
		ReturnURL:   request.ReturnURL,
		Description: request.Description,
	})
	if err != nil {
		h.logger.Error("create payment attempt failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("error", err.Error()),
		)

		basehttp.WriteError(w, err)
		return
	}

	basehttp.WriteJSON(w, http.StatusCreated, toPaymentAttemptResponse(result.Attempt))
}

func (h *Handler) getPaymentAttempt(w http.ResponseWriter, r *http.Request, attemptID string) {
	attempt, err := h.service.GetPaymentAttempt(r.Context(), attemptID)
	if err != nil {
		h.logger.Error("get payment attempt failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("attempt_id", attemptID),
			slog.String("error", err.Error()),
		)

		basehttp.WriteError(w, err)
		return
	}

	basehttp.WriteJSON(w, http.StatusOK, toPaymentAttemptResponse(attempt))
}

func (h *Handler) reconcilePaymentAttempt(w http.ResponseWriter, r *http.Request, attemptID string) {
	attempt, err := h.service.ReconcilePaymentAttempt(r.Context(), attemptID)
	if err != nil {
		h.logger.Error("reconcile payment attempt failed",
			slog.String("request_id", basehttp.RequestIDFromContext(r.Context())),
			slog.String("attempt_id", attemptID),
			slog.String("error", err.Error()),
		)

		basehttp.WriteError(w, err)
		return
	}

	basehttp.WriteJSON(w, http.StatusOK, toPaymentAttemptResponse(attempt))
}
