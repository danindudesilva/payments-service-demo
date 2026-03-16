package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/danindudesilva/payments-service/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const DefaultTestDatabaseURL = "postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable"

func TestDatabaseURL() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}

	return DefaultTestDatabaseURL
}

func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool, err := database.NewPool(context.Background(), database.Config{
		DatabaseURL: TestDatabaseURL(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM payment_attempts`)
		pool.Close()
	})

	return pool
}
