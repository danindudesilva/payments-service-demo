package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPAddress(t *testing.T) {
	t.Parallel()

	cfg := Config{HTTPPort: "9999"}

	assert.Equal(t, ":9999", cfg.HTTPAddress())
}

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	unsetEnv(t, "APP_ENV")
	unsetEnv(t, "HTTP_PORT")
	unsetEnv(t, "PAYMENTS_PROVIDER")
	unsetEnv(t, "STRIPE_SECRET_KEY")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "8080", cfg.HTTPPort)
	assert.Equal(t, "fake", cfg.PaymentsProvider)
	assert.Equal(t, "", cfg.StripeSecretKey)
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	original, existed := os.LookupEnv(key)
	_ = os.Unsetenv(key)

	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, original)
			return
		}
		_ = os.Unsetenv(key)
	})
}
