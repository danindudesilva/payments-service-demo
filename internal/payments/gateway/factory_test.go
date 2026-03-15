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

func TestNew_StripeProviderNotImplementedYet(t *testing.T) {
	t.Parallel()

	gateway, err := New(config.Config{
		PaymentsProvider: "stripe",
	})
	require.Error(t, err)
	require.Nil(t, gateway)
	require.Contains(t, err.Error(), "not implemented")
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
