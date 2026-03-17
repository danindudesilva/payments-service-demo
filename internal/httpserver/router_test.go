package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danindudesilva/payments-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		AppEnv:     "staging",
		AppVersion: "v1.0.0",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), `"status":"ok"`)
	assert.Contains(t, res.Body.String(), `"env":"staging"`)
	assert.Contains(t, res.Body.String(), `"version":"v1.0.0"`)
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		config.Config{AppEnv: "test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.True(t, strings.Contains(res.Body.String(), "method not allowed"))
}
