package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv           string
	HTTPPort         string
	PaymentsProvider string
	StripeSecretKey  string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		PaymentsProvider: getEnv("PAYMENTS_PROVIDER", "fake"),
		StripeSecretKey:  getEnv("STRIPE_SECRET_KEY", ""),
	}

	if cfg.HTTPPort == "" {
		return Config{}, fmt.Errorf("HTTP_PORT must not be empty")
	}

	if cfg.PaymentsProvider == "" {
		return Config{}, fmt.Errorf("PAYMENTS_PROVIDER must not be empty")
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
