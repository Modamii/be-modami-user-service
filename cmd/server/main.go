package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modami/user-service/internal/adapter/cache"
	grpcadapter "github.com/modami/user-service/internal/adapter/grpc"
	"github.com/modami/user-service/internal/adapter/handler"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/adapter/messaging"
	"github.com/modami/user-service/internal/adapter/repository"
	"github.com/modami/user-service/internal/config"
	"github.com/modami/user-service/internal/port"
	"github.com/modami/user-service/internal/service"
	pkgredis "github.com/modami/user-service/pkg/storage/redis"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	loggingmw "gitlab.com/lifegoeson-libs/pkg-logging/middleware"
	pb "gitlab.com/lifegoeson-libs/pkg-techinsights-grpc-client/go/modami/user"
	"google.golang.org/grpc"
)

// @title           Modami User Service API
// @version         1.0
// @description     User service for the Modami marketplace platform.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	dbPool, err := pgxpool.New(ctx, cfg.Postgres.WriterDSN())
	if err != nil {
		logger.Error(ctx, "failed to connect to postgres", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error(ctx, "postgres ping failed", err)
		os.Exit(1)
	}
	logger.Info(ctx, "connected to postgres")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient, err := pkgredis.NewRedisClient(pkgredis.RedisConfig{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.Database,
	})
	if err != nil {
		logger.Warn(ctx, "redis connection failed (continuing without cache)", logging.String("error", err.Error()))
	}
	if redisClient != nil {
		defer pkgredis.CloseRedis(ctx, redisClient)
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(dbPool)
	sellerRepo := repository.NewSellerProfileRepository(dbPool)
	followRepo := repository.NewFollowRepository(dbPool)
	reviewRepo := repository.NewReviewRepository(dbPool)
	addressRepo := repository.NewAddressRepository(dbPool)
	kycRepo := repository.NewKYCRepository(dbPool)
	outboxRepo := repository.NewOutboxRepository(dbPool)
	processedEventRepo := repository.NewProcessedEventRepository(dbPool)

	// ── Cache ─────────────────────────────────────────────────────────────────
	cacheService := cache.NewRedisCache(redisClient)

	// ── Kafka producer ────────────────────────────────────────────────────────
	publisher, err := messaging.NewKafkaProducer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.Env,
		cfg.Kafka.ClientID,
		outboxRepo,
	)
	if err != nil {
		logger.Error(ctx, "failed to create kafka producer", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// ── Services ──────────────────────────────────────────────────────────────
	userService := service.NewUserService(userRepo, cacheService, publisher)
	followService := service.NewFollowService(followRepo, cacheService, publisher)
	reviewService := service.NewReviewService(reviewRepo, userRepo, sellerRepo, cacheService, publisher)
	addressService := service.NewAddressService(addressRepo, cacheService)
	sellerService := service.NewSellerService(sellerRepo, userRepo, cacheService)
	kycService := service.NewKYCService(kycRepo, sellerRepo, userRepo, cacheService, publisher)

	// ── Kafka consumer ────────────────────────────────────────────────────────
	consumer, err := messaging.NewConsumer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.ConsumerGroup,
		cfg.Kafka.Env,
		cfg.Kafka.ClientID,
		processedEventRepo,
		userService,
	)
	if err != nil {
		logger.Error(ctx, "failed to create kafka consumer", err)
		os.Exit(1)
	}
	defer consumer.Close()

	// ── Auth middleware ───────────────────────────────────────────────────────
	authMiddleware, authErr := middleware.NewAuthMiddleware(cfg.Keycloak.JWKSURL, userService)
	if authErr != nil {
		logger.Warn(ctx, "auth middleware init warning", logging.String("error", authErr.Error()))
	}

	// ── HTTP handlers ─────────────────────────────────────────────────────────
	userHandler := handler.NewUserHandler(userService)
	followHandler := handler.NewFollowHandler(followService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	addressHandler := handler.NewAddressHandler(addressService)
	sellerHandler := handler.NewSellerHandler(sellerService, kycService)
	adminHandler := handler.NewAdminHandler(userService, kycService)

	// ── Gin router ────────────────────────────────────────────────────────────
	if cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")

	// Public
	v1.GET("/users/search", userHandler.SearchUsers)
	v1.GET("/users/:id", userHandler.GetProfile)
	v1.GET("/users/:id/followers", followHandler.GetFollowers)
	v1.GET("/users/:id/following", followHandler.GetFollowing)
	v1.GET("/users/:id/reviews", reviewHandler.ListReviews)
	v1.GET("/users/:id/reviews/summary", reviewHandler.GetRatingSummary)
	v1.GET("/users/:id/shop", sellerHandler.GetShopProfile)

	// Authenticated
	auth := v1.Group("")
	auth.Use(authMiddleware.Authenticate())

	auth.GET("/users/me", userHandler.GetMyProfile)
	auth.PUT("/users/me", userHandler.UpdateProfile)
	auth.PUT("/users/me/avatar", userHandler.UpdateAvatar)
	auth.PUT("/users/me/cover", userHandler.UpdateCover)
	auth.DELETE("/users/me", userHandler.DeactivateAccount)

	auth.POST("/users/:id/follow", followHandler.Follow)
	auth.DELETE("/users/:id/follow", followHandler.Unfollow)
	auth.GET("/users/:id/follow/status", followHandler.CheckFollowStatus)

	auth.POST("/users/:id/reviews", reviewHandler.CreateReview)

	auth.POST("/users/me/addresses", addressHandler.AddAddress)
	auth.GET("/users/me/addresses", addressHandler.ListAddresses)
	auth.PUT("/users/me/addresses/:addr_id", addressHandler.UpdateAddress)
	auth.DELETE("/users/me/addresses/:addr_id", addressHandler.DeleteAddress)
	auth.PUT("/users/me/addresses/:addr_id/default", addressHandler.SetDefault)

	auth.POST("/users/me/seller/register", sellerHandler.Register)
	auth.PUT("/users/me/seller/profile", sellerHandler.UpdateProfile)
	auth.POST("/users/me/seller/kyc", sellerHandler.SubmitKYC)
	auth.GET("/users/me/seller/kyc/status", sellerHandler.GetKYCStatus)

	// Admin
	adminGroup := v1.Group("/admin")
	adminGroup.Use(authMiddleware.Authenticate())
	adminGroup.Use(authMiddleware.RequireRole("admin"))
	adminGroup.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
	adminGroup.PUT("/users/:id/kyc/approve", adminHandler.ApproveKYC)
	adminGroup.PUT("/users/:id/kyc/reject", adminHandler.RejectKYC)
	adminGroup.GET("/users", adminHandler.ListUsers)

	// ── Background workers ────────────────────────────────────────────────────
	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	defer cancelConsumer()

	go func() {
		logger.Info(ctx, "starting kafka consumer")
		consumer.Start(consumerCtx)
	}()

	go func() {
		logger.Info(ctx, "starting outbox worker")
		runOutboxWorker(consumerCtx, outboxRepo, publisher)
	}()

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(loggingmw.GRPCStatsHandler()),
		grpc.UnaryInterceptor(loggingmw.UnaryServerInterceptor()),
	)
	pb.RegisterUserInternalServiceServer(grpcServer, grpcadapter.NewUserGRPCServer(userService, sellerService))

	go func() {
		lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
		if lisErr != nil {
			logger.Error(ctx, "failed to listen gRPC", lisErr)
			os.Exit(1)
		}
		logger.Info(ctx, "gRPC server listening", logging.String("port", cfg.GRPC.Port))
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			logger.Error(ctx, "gRPC server error", serveErr)
		}
	}()

	// ── HTTP server ───────────────────────────────────────────────────────────
	httpHandler := loggingmw.HTTPMiddleware("user-service", router, &loggingmw.HttpLoggingOptions{
		ExceptRoutes: []string{"/health", "/metrics"},
	})
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info(ctx, "HTTP server listening", logging.String("port", cfg.Server.Port))
		if serveErr := httpSrv.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", serveErr)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(ctx, "shutting down...")
	cancelConsumer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if shutErr := httpSrv.Shutdown(shutdownCtx); shutErr != nil {
		logger.Error(ctx, "HTTP shutdown error", shutErr)
	}
	grpcServer.GracefulStop()
	logger.Info(ctx, "server exited")
}

func runOutboxWorker(ctx context.Context, outboxRepo port.OutboxRepository, publisher port.EventPublisher) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processOutboxEvents(ctx, outboxRepo, publisher)
		}
	}
}

func processOutboxEvents(ctx context.Context, outboxRepo port.OutboxRepository, _ port.EventPublisher) {
	events, err := outboxRepo.GetPending(ctx, 50)
	if err != nil {
		logger.Error(ctx, "outbox: get pending error", err)
		return
	}
	for _, event := range events {
		logger.Info(ctx, "outbox: pending event", logging.String("id", event.ID.String()), logging.String("topic", event.Topic))
	}
}
