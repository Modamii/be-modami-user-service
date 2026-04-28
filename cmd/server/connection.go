package main

import (
	"context"

	"be-modami-user-service/config"

	"github.com/jackc/pgx/v5/pgxpool"
	gokit_redis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type Connections struct {
	DB    *pgxpool.Pool
	Redis gokit_redis.CachePort
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

	var redisCache gokit_redis.CachePort
	redisCfg := gokit_redis.Config{
		Addrs:    []string{cfg.Redis.Addr()},
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.Database,
	}
	adapter, redisErr := gokit_redis.NewAdapter(redisCfg)
	if redisErr != nil {
		logger.Warn(ctx, "redis connection failed (continuing without cache)", logging.String("error", redisErr.Error()))
	} else {
		redisCache = adapter
	}

	return &Connections{
		DB:    dbPool,
		Redis: redisCache,
	}, nil
}

func (c *Connections) Close(ctx context.Context) {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
}
