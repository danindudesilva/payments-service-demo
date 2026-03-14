package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPAddress(t *testing.T) {
	t.Parallel()

	cfg := Config{HTTPPort: "9999"}

	assert.Equal(t, ":9999", cfg.HTTPAddress())
}
