package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv               string
	HTTPPort             string
	PaymentsProvider     string
	StripeSecretKey      string
	StripePublishableKey string
	DatabaseURL          string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		HTTPPort:             getEnv("HTTP_PORT", "8080"),
		PaymentsProvider:     getEnv("PAYMENTS_PROVIDER", "fake"),
		StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
		StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
	}

	if strings.TrimSpace(cfg.HTTPPort) == "" {
		return Config{}, fmt.Errorf("HTTP_PORT must not be empty")
	}

	if strings.TrimSpace(cfg.PaymentsProvider) == "" {
		return Config{}, fmt.Errorf("PAYMENTS_PROVIDER must not be empty")
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}

	if cfg.PaymentsProvider == "stripe" {
		if strings.TrimSpace(cfg.StripeSecretKey) == "" {
			return Config{}, fmt.Errorf("STRIPE_SECRET_KEY must not be empty when PAYMENTS_PROVIDER=stripe")
		}
		if strings.TrimSpace(cfg.StripePublishableKey) == "" {
			return Config{}, fmt.Errorf("STRIPE_PUBLISHABLE_KEY must not be empty when PAYMENTS_PROVIDER=stripe")
		}
	}

	return cfg, nil
}

func (c Config) HTTPAddress() string {
	return fmt.Sprintf(":%s", c.HTTPPort)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
