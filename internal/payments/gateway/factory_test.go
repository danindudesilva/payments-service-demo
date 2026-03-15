package gateway

import (
	"testing"

	"github.com/danindudesilva/payments-service/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNew_FakeProvider(t *testing.T) {
	t.Parallel()

	gateway, err := New(config.Config{
		PaymentsProvider: "fake",
	})
	require.NoError(t, err)
	require.NotNil(t, gateway)
}

func TestNew_StripeProvider_RequiresSecretKey(t *testing.T) {
	t.Parallel()

	gateway, err := New(config.Config{
		PaymentsProvider: "stripe",
		StripeSecretKey:  "",
	})
	require.Error(t, err)
	require.Nil(t, gateway)
	require.Contains(t, err.Error(), "stripe secret key must not be empty")
}

func TestNew_StripeProvider_SucceedsWithKey(t *testing.T) {
	t.Parallel()

	gateway, err := New(config.Config{
		PaymentsProvider: "stripe",
		StripeSecretKey:  "sk_test_51TApvV",
	})
	require.NoError(t, err)
	require.NotNil(t, gateway)
}

func TestNew_UnsupportedProvider(t *testing.T) {
	t.Parallel()

	gateway, err := New(config.Config{
		PaymentsProvider: "unknown",
	})
	require.Error(t, err)
	require.Nil(t, gateway)
	require.Contains(t, err.Error(), "unsupported payments provider")
}
