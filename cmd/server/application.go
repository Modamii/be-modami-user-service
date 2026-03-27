package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modami/user-service/config"
	_ "github.com/modami/user-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/modami/user-service/internal/adapter/cache"
	grpcadapter "github.com/modami/user-service/internal/adapter/grpc"
	"github.com/modami/user-service/internal/adapter/handler"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/adapter/messaging"
	"github.com/modami/user-service/internal/adapter/repository"
	"github.com/modami/user-service/internal/port"
	"github.com/modami/user-service/internal/service"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	loggingmw "gitlab.com/lifegoeson-libs/pkg-logging/middleware"
	pb "gitlab.com/lifegoeson-libs/pkg-techinsights-grpc-client/go/modami/user"
	"google.golang.org/grpc"
)

type Application struct {
	HTTPServer *http.Server
	GRPCServer *grpc.Server
	Publisher  port.EventPublisher
	Consumer   *messaging.Consumer
	OutboxRepo port.OutboxRepository
}

func newApplication(ctx context.Context, cfg *config.Config, conns *Connections) (*Application, error) {
	userRepo := repository.NewUserRepository(conns.DB)
	sellerRepo := repository.NewSellerProfileRepository(conns.DB)
	followRepo := repository.NewFollowRepository(conns.DB)
	reviewRepo := repository.NewReviewRepository(conns.DB)
	addressRepo := repository.NewAddressRepository(conns.DB)
	kycRepo := repository.NewKYCRepository(conns.DB)
	outboxRepo := repository.NewOutboxRepository(conns.DB)
	processedEventRepo := repository.NewProcessedEventRepository(conns.DB)

	cacheService := cache.NewRedisCache(conns.Redis)

	publisher, err := messaging.NewKafkaProducer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.Env,
		cfg.Kafka.ClientID,
		outboxRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	userService := service.NewUserService(userRepo, cacheService, publisher)
	followService := service.NewFollowService(followRepo, cacheService, publisher)
	reviewService := service.NewReviewService(reviewRepo, userRepo, sellerRepo, cacheService, publisher)
	addressService := service.NewAddressService(addressRepo, cacheService)
	sellerService := service.NewSellerService(sellerRepo, userRepo, cacheService)
	kycService := service.NewKYCService(kycRepo, sellerRepo, userRepo, cacheService, publisher)

	consumer, err := messaging.NewConsumer(
		cfg.Kafka.Brokers(),
		cfg.Kafka.ConsumerGroup,
		cfg.Kafka.Env,
		cfg.Kafka.ClientID,
		processedEventRepo,
		userService,
	)
	if err != nil {
		publisher.Close()
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	authMiddleware, authErr := middleware.NewAuthMiddleware(cfg.Keycloak.JWKSURL, userService)
	if authErr != nil {
		logger.Warn(ctx, "auth middleware init warning", logging.String("error", authErr.Error()))
	}

	userHandler := handler.NewUserHandler(userService)
	followHandler := handler.NewFollowHandler(followService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	addressHandler := handler.NewAddressHandler(addressService)
	sellerHandler := handler.NewSellerHandler(sellerService, kycService)
	adminHandler := handler.NewAdminHandler(userService, kycService)

	if cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())
	router.Use(gin.Recovery())

	registerRoutes(router, authMiddleware, userHandler, followHandler, reviewHandler, addressHandler, sellerHandler, adminHandler)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(loggingmw.GRPCStatsHandler()),
		grpc.UnaryInterceptor(loggingmw.UnaryServerInterceptor()),
	)
	pb.RegisterUserInternalServiceServer(grpcServer, grpcadapter.NewUserGRPCServer(userService, sellerService))

	httpHandler := loggingmw.HTTPMiddleware("user-service", router, &loggingmw.HttpLoggingOptions{
		ExceptRoutes: []string{"/health", "/metrics"},
	})
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
		Publisher:  publisher,
		Consumer:   consumer,
		OutboxRepo: outboxRepo,
	}, nil
}

func registerRoutes(
	router *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	userHandler *handler.UserHandler,
	followHandler *handler.FollowHandler,
	reviewHandler *handler.ReviewHandler,
	addressHandler *handler.AddressHandler,
	sellerHandler *handler.SellerHandler,
	adminHandler *handler.AdminHandler,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
