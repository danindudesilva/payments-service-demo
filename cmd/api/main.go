package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/danindudesilva/payments-service/internal/app"
	"github.com/danindudesilva/payments-service/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("application exited with error: %v", err)
	}
}
