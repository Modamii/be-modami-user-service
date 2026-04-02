package main

import (
	"context"

	"be-modami-user-service/config"
	pkgredis "be-modami-user-service/pkg/storage/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type Connections struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func newConnections(ctx context.Context, cfg *config.Config) (*Connections, error) {
	dbPool, err := pgxpool.New(ctx, cfg.Postgres.WriterDSN())
	if err != nil {
		return nil, err
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}
	logger.Info(ctx, "connected to postgres")

	redisClient, err := pkgredis.NewRedisClient(pkgredis.RedisConfig{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.Database,
	})
	if err != nil {
		logger.Warn(ctx, "redis connection failed (continuing without cache)", logging.String("error", err.Error()))
	}

	return &Connections{
		DB:    dbPool,
		Redis: redisClient,
	}, nil
}

func (c *Connections) Close(ctx context.Context) {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Redis != nil {
		pkgredis.CloseRedis(ctx, c.Redis)
	}
}
