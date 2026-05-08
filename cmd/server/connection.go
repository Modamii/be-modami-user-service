package main

import (
	"context"
	"fmt"
	"time"

	"be-modami-user-service/config"

	"github.com/jackc/pgx/v5/pgxpool"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type Connections struct {
	DB    *pgxpool.Pool
	Redis pkgredis.CachePort
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

	var redisCache pkgredis.CachePort
	adapter, redisErr := pkgredis.NewAdapter(pkgredis.Config{
		Addrs:       []string{fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)},
		Password:    cfg.Redis.Pass,
		DB:          cfg.Redis.Database,
		PoolSize:    cfg.Redis.PoolSize,
		DialTimeout: 5 * time.Second,
	})
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
