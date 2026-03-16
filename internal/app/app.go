package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danindudesilva/payments-service/internal/config"
	"github.com/danindudesilva/payments-service/internal/demo"
	"github.com/danindudesilva/payments-service/internal/httpserver"
	"github.com/danindudesilva/payments-service/internal/payments/gateway"
	pg "github.com/danindudesilva/payments-service/internal/payments/repository/postgres"
	paymentservice "github.com/danindudesilva/payments-service/internal/payments/service"
	paymenthttp "github.com/danindudesilva/payments-service/internal/payments/transport/http"
	"github.com/danindudesilva/payments-service/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg    config.Config
	server *http.Server
	logger *slog.Logger
	dbPool *pgxpool.Pool
}

func New(cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbPool, err := database.NewPool(context.Background(), database.Config{
		DatabaseURL: cfg.DatabaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	repo := pg.NewRepository(dbPool)

	paymentGateway, err := gateway.New(gateway.Config{
		PaymentsProvider:     cfg.PaymentsProvider,
		StripeSecretKey:      cfg.StripeSecretKey,
		StripePublishableKey: cfg.StripePublishableKey,
	})
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("create payment gateway: %w", err)
	}

	service := paymentservice.New(
		repo,
		paymentGateway,
		time.Now,
		func() string {
			return fmt.Sprintf("attempt_%d", time.Now().UnixNano())
		},
	)

	router := httpserver.NewRouter(cfg, logger)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	handler := paymenthttp.NewHandler(service, logger)
	handler.Register(mux)

	webhookHandler := paymenthttp.NewWebhookHandler(logger, cfg.StripeWebhookSecret)
	webhookHandler.Register(mux)

	demoHandler, err := demo.NewHandler(cfg.StripePublishableKey)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("create demo handler: %w", err)
	}
	demoHandler.Register(mux)

	server := &http.Server{
		Addr:              cfg.HTTPAddress(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:    cfg,
		server: server,
		logger: logger,
		dbPool: dbPool,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("http server starting",
			slog.String("addr", a.server.Addr),
			slog.String("env", a.cfg.AppEnv),
			slog.String("payments_provider", a.cfg.PaymentsProvider),
		)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if a.dbPool != nil {
		a.dbPool.Close()
	}

	a.logger.Info("http server stopped")
	return nil
}
