package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cawa87/garantex-test/internal/config"
	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
	"github.com/cawa87/garantex-test/internal/repository/postgres"
	"github.com/cawa87/garantex-test/internal/service/exchange"
	"github.com/cawa87/garantex-test/internal/transport/grpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	config  *config.Config
	logger  *sl.Logger
	repo    *postgres.Repository
	server  *grpc.Server
	metrics *http.Server
}

func New(cfg *config.Config) (*App, error) {
	logger, err := sl.New(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	repo, err := postgres.NewRepository(cfg.Database.GetDSN(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	exchangeClient := exchange.NewClient(cfg.Exchange.BaseURL, cfg.Exchange.Timeout, logger)
	server := grpc.NewServer(repo, exchangeClient, logger)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.MetricsPort),
		Handler: metricsMux,
	}

	return &App{
		config:  cfg,
		logger:  logger,
		repo:    repo,
		server:  server,
		metrics: metricsServer,
	}, nil
}

func (a *App) Run() error {
	go func() {
		a.logger.Info("Starting metrics server", "port", a.config.Server.MetricsPort)
		if err := a.metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Metrics server failed", "error", err)
		}
	}()

	go func() {
		a.logger.Info("Starting gRPC server", "port", a.config.Server.GRPCPort)
		if err := a.server.Run(a.config.Server.GRPCPort); err != nil {
			a.logger.Error("gRPC server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down application...")
	return a.Shutdown()
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.metrics.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown metrics server", "error", err)
	}

	a.repo.Close()

	a.logger.Info("Application shutdown completed")
	return nil
}
