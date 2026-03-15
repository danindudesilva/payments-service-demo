package demo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ServesDemoPage(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler("pk_test_123")
	require.NoError(t, err)

	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "payments-service demo")
	assert.Contains(t, res.Body.String(), "pk_test_123")
}

func TestHandler_ServesStaticJS(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler("pk_test_123")
	require.NoError(t, err)

	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/demo/static/app.js", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), "stripe.confirmPayment")
}

func TestHandler_NotFoundForUnknownDemoPath(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler("pk_test_123")
	require.NoError(t, err)

	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/demo/unknown", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandler_ServesDemoCSS(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler("pk_test_123")
	require.NoError(t, err)

	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/demo/static/styles.css", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Body.String(), ".container")
}
