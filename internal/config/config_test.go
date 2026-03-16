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
	unsetEnv(t, "APP_ENV")
	unsetEnv(t, "HTTP_PORT")
	unsetEnv(t, "PAYMENTS_PROVIDER")
	unsetEnv(t, "STRIPE_SECRET_KEY")
	unsetEnv(t, "STRIPE_PUBLISHABLE_KEY")

	t.Setenv("DATABASE_URL", "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "8080", cfg.HTTPPort)
	assert.Equal(t, "fake", cfg.PaymentsProvider)
	assert.Equal(t, "", cfg.StripeSecretKey)
	assert.Equal(t, "", cfg.StripePublishableKey)
	assert.Equal(t, "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable", cfg.DatabaseURL)
}

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	unsetEnv(t, "DATABASE_URL")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL must not be empty")
}

func TestLoad_StripeProviderRequiresSecretKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable")
	t.Setenv("PAYMENTS_PROVIDER", "stripe")
	unsetEnv(t, "STRIPE_SECRET_KEY")
	unsetEnv(t, "STRIPE_PUBLISHABLE_KEY")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "STRIPE_SECRET_KEY must not be empty")
}

func TestLoad_StripeProviderRequiresPublishableKey(t *testing.T) {

	t.Setenv("DATABASE_URL", "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable")
	t.Setenv("PAYMENTS_PROVIDER", "stripe")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	unsetEnv(t, "STRIPE_PUBLISHABLE_KEY")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "STRIPE_PUBLISHABLE_KEY must not be empty")
}

func TestLoad_StripeProviderWithRequiredKeys(t *testing.T) {

	t.Setenv("DATABASE_URL", "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable")
	t.Setenv("PAYMENTS_PROVIDER", "stripe")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_123")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_test_123")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "stripe", cfg.PaymentsProvider)
	assert.Equal(t, "sk_test_123", cfg.StripeSecretKey)
	assert.Equal(t, "pk_test_123", cfg.StripePublishableKey)
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
