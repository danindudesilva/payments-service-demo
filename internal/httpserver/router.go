package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danindudesilva/payments-service/internal/config"
)

func NewRouter(cfg config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/healthz", chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
					"error": "method not allowed",
				})
				return
			}

			writeJSON(w, http.StatusOK, map[string]any{
				"status": "ok",
				"env":    cfg.AppEnv,
			})
		}),

		requestID(),
		timeout(30*time.Second),
		recoverPanic(logger),
		requestLogger(logger),
	))

	return mux
}
