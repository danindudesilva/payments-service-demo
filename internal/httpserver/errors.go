package httpserver

import (
	"errors"
	"net/http"

	"github.com/danindudesilva/payments-service/internal/payments/domain"
	"github.com/danindudesilva/payments-service/internal/payments/service"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type mappedError struct {
	StatusCode int
	Message    string
}

func MapError(err error) mappedError {
	var validationErr service.ValidationError
	if errors.As(err, &validationErr) {
		return mappedError{
			StatusCode: http.StatusBadRequest,
			Message:    validationErr.Error(),
		}
	}

	switch {
	case errors.Is(err, domain.ErrPaymentNotFound):
		return mappedError{
			StatusCode: http.StatusNotFound,
			Message:    "payment attempt not found",
		}

	case errors.Is(err, domain.ErrInvalidTransition):
		return mappedError{
			StatusCode: http.StatusConflict,
			Message:    "invalid payment state transition",
		}

	case errors.Is(err, domain.ErrProviderAlreadyLinked):
		return mappedError{
			StatusCode: http.StatusConflict,
			Message:    "provider payment is already linked",
		}

	case errors.Is(err, domain.ErrInvalidMoney), errors.Is(err, domain.ErrInvalidNextAction):
		return mappedError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid payment request",
		}

	default:
		return mappedError{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal server error",
		}
	}
}

func WriteError(w http.ResponseWriter, err error) {
	mapped := MapError(err)

	WriteJSON(w, mapped.StatusCode, ErrorResponse{
		Error: mapped.Message,
	})
}
