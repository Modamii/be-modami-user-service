package main

// @title           Modami User Service API
// @version         1.0
// @description     User service for the Modami marketplace platform.
// @host            localhost:8086
// @BasePath        /v1/user-services
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"be-modami-user-service/config"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(logging.Config{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Observability.Environment,
		Level:          cfg.Observability.LogLevel,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
		Insecure:       cfg.Observability.OTLPInsecure,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	defer logger.Shutdown(ctx)

	if err := ensureDatabase(ctx, cfg.Postgres); err != nil {
		logger.Error(ctx, "failed to ensure database", err)
		os.Exit(1)
	}

	if err := runMigrations(ctx, cfg.Postgres.MigrationDSN()); err != nil {
		logger.Error(ctx, "failed to run migrations", err)
		os.Exit(1)
	}

	conns, err := newConnections(ctx, cfg)
	if err != nil {
		logger.Error(ctx, "failed to establish connections", err)
		os.Exit(1)
	}
	defer conns.Close(ctx)

	app, err := newApplication(ctx, cfg, conns)
	if err != nil {
		logger.Error(ctx, "failed to build application", err)
		os.Exit(1)
	}
	defer app.Consumer.Close()

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	go func() {
		logger.Info(ctx, "starting kafka consumer")
		app.Consumer.Start(workerCtx)
	}()

	go func() {
		logger.Info(ctx, "starting outbox worker")
		runOutboxWorker(workerCtx, app.OutboxRepo, app.Publisher)
	}()

	go func() {
		lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
		if lisErr != nil {
			logger.Error(ctx, "failed to listen gRPC", lisErr)
			os.Exit(1)
		}
		logger.Info(ctx, "gRPC server listening", logging.String("port", cfg.GRPC.Port))
		if serveErr := app.GRPCServer.Serve(lis); serveErr != nil {
			logger.Error(ctx, "gRPC server error", serveErr)
		}
	}()

	go func() {
		logger.Info(ctx, "HTTP server listening", logging.String("port", cfg.Server.Port))
		if serveErr := app.HTTPServer.ListenAndServe(); serveErr != nil {
			logger.Error(ctx, "HTTP server error", serveErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(ctx, "shutting down...")
	cancelWorkers()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if shutErr := app.HTTPServer.Shutdown(shutdownCtx); shutErr != nil {
		logger.Error(ctx, "HTTP shutdown error", shutErr)
	}
	app.GRPCServer.GracefulStop()
	logger.Info(ctx, "server exited")
}
