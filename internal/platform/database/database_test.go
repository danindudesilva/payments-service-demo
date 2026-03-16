package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPool_InvalidURL(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), Config{
		DatabaseURL: "://not-a-valid-url",
	})

	require.Error(t, err)
	require.Nil(t, pool)
	require.ErrorContains(t, err, "parse pgx pool config")
}

func TestNewPool_PingFails(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), Config{
		DatabaseURL: "postgres://user:pass@127.0.0.1:1/dbname?sslmode=disable",
	})

	require.Error(t, err)
	require.Nil(t, pool)
	require.ErrorContains(t, err, "ping database")
}

func TestNewPool_Success(t *testing.T) {
	t.Parallel()

	pool, err := NewPool(context.Background(), Config{
		DatabaseURL: "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable",
	})

	require.NoError(t, err)
	require.NotNil(t, pool)

	t.Cleanup(func() {
		pool.Close()
	})

	require.NoError(t, pool.Ping(context.Background()))
}
