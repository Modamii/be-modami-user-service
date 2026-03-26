package main

import (
	"context"
	"fmt"
	"log"
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
	pb "github.com/modami/user-service/proto/user"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	log.Println("connected to postgres")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient, err := pkgredis.NewRedisClient(pkgredis.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		log.Printf("redis connection failed (continuing without cache): %v", err)
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
		cfg.KafkaBrokers,
		cfg.KafkaEnv,
		cfg.KafkaClientID,
		outboxRepo,
	)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
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
		cfg.KafkaBrokers,
		cfg.KafkaConsumerGroup,
		cfg.KafkaEnv,
		cfg.KafkaClientID,
		processedEventRepo,
		userService,
	)
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	// ── Auth middleware ───────────────────────────────────────────────────────
	authMiddleware, authErr := middleware.NewAuthMiddleware(cfg.KeycloakJWKSURL, userService)
	if authErr != nil {
		log.Printf("auth middleware init warning: %v", authErr)
	}

	// ── HTTP handlers ─────────────────────────────────────────────────────────
	userHandler := handler.NewUserHandler(userService)
	followHandler := handler.NewFollowHandler(followService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	addressHandler := handler.NewAddressHandler(addressService)
	sellerHandler := handler.NewSellerHandler(sellerService, kycService)
	adminHandler := handler.NewAdminHandler(userService, kycService)

	// ── Gin router ────────────────────────────────────────────────────────────
	if cfg.LogLevel != "debug" {
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
		log.Println("starting kafka consumer")
		consumer.Start(consumerCtx)
	}()

	go func() {
		log.Println("starting outbox worker")
		runOutboxWorker(consumerCtx, outboxRepo, publisher)
	}()

	// ── gRPC server ───────────────────────────────────────────────────────────
	grpcServer := grpc.NewServer()
	pb.RegisterUserInternalServiceServer(grpcServer, grpcadapter.NewUserGRPCServer(userService, sellerService))

	go func() {
		lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
		if lisErr != nil {
			log.Fatalf("failed to listen gRPC: %v", lisErr)
		}
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Printf("gRPC server error: %v", serveErr)
		}
	}()

	// ── HTTP server ───────────────────────────────────────────────────────────
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.ServerPort)
		if serveErr := httpSrv.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", serveErr)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down...")
	cancelConsumer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if shutErr := httpSrv.Shutdown(shutdownCtx); shutErr != nil {
		log.Printf("HTTP shutdown error: %v", shutErr)
	}
	grpcServer.GracefulStop()
	log.Println("server exited")
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
		log.Printf("outbox: get pending error: %v", err)
		return
	}
	for _, event := range events {
		log.Printf("outbox: pending event id=%s topic=%s", event.ID, event.Topic)
	}
}
