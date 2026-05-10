package main

import (
	"be-modami-user-service/config"
	_ "be-modami-user-service/docs"
	grpcadapter "be-modami-user-service/internal/adapter/grpc"
	"be-modami-user-service/internal/adapter/http/handler"
	"be-modami-user-service/internal/adapter/http/middleware"
	"be-modami-user-service/internal/adapter/messaging"
	"be-modami-user-service/internal/adapter/repository"
	"be-modami-user-service/internal/service"
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	gokitkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	loggingmw "gitlab.com/lifegoeson-libs/pkg-logging/middleware"
	pb "gitlab.com/lifegoeson-libs/pkg-techinsights-grpc-client/go/modami/user"
	"google.golang.org/grpc"
)

type Application struct {
	HTTPServer    *http.Server
	GRPCServer    *grpc.Server
	KafkaHandlers []gokitkafka.ConsumerHandler
}

func newApplication(ctx context.Context, cfg *config.Config, conns *Connections) (*Application, error) {
	// repositories
	userRepo := repository.NewCachedUserRepository(repository.NewUserRepository(conns.DB), conns.Redis)
	sellerRepo := repository.NewCachedSellerRepository(repository.NewSellerProfileRepository(conns.DB), conns.Redis)
	followRepo := repository.NewCachedFollowRepository(repository.NewFollowRepository(conns.DB), conns.Redis)
	reviewRepo := repository.NewCachedReviewRepository(repository.NewReviewRepository(conns.DB), conns.Redis)
	addressRepo := repository.NewCachedAddressRepository(repository.NewAddressRepository(conns.DB), conns.Redis)
	kycRepo := repository.NewKYCRepository(conns.DB)
	outboxRepo := repository.NewOutboxRepository(conns.DB)
	processedRepo := repository.NewProcessedEventRepository(conns.DB)
	txManager := repository.NewTxManager(conns.DB)

	// services
	userService := service.NewUserService(userRepo, txManager, outboxRepo)
	followService := service.NewFollowService(followRepo, txManager, outboxRepo)
	reviewService := service.NewReviewService(reviewRepo, userRepo, sellerRepo, txManager, outboxRepo)
	kycService := service.NewKYCService(kycRepo, sellerRepo, userRepo, txManager, outboxRepo)
	addressService := service.NewAddressService(addressRepo)
	sellerService := service.NewSellerService(sellerRepo, userRepo)

	// handlers
	userHandler := handler.NewUserHandler(userService)
	followHandler := handler.NewFollowHandler(followService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	addressHandler := handler.NewAddressHandler(addressService)
	sellerHandler := handler.NewSellerHandler(sellerService, kycService)
	adminHandler := handler.NewAdminHandler(userService, kycService)

	authMiddleware, authErr := middleware.NewAuthMiddleware(cfg.Keycloak.JWKSURL, userService)
	if authErr != nil {
		logger.Warn(ctx, "auth middleware init warning", logging.String("error", authErr.Error()))
	}

	// global middleware
	if cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(middleware.RateLimit())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: cfg.App.AllowCredentials,
		MaxAge:           300,
	}))

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
		Addr:         cfg.App.ListenAddr(),
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
		KafkaHandlers: []gokitkafka.ConsumerHandler{
			messaging.NewKeycloakCDCHandler(userService, processedRepo),
		},
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

	v1 := router.Group("/v1/user-services")

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

