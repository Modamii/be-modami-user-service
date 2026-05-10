package main

import (
	"context"
	"fmt"
	"time"

	"be-modami-user-service/config"

	"github.com/jackc/pgx/v5/pgxpool"
	pkgkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type Connections struct {
	DB            *pgxpool.Pool
	Redis         pkgredis.CachePort
	Kafka         *pkgkafka.KafkaService
	KafkaConsumer *pkgkafka.KafkaService
}

func newConnections(ctx context.Context, cfg *config.Config) (*Connections, error) {
	conn := &Connections{}

	// Postgres
	dbPool, err := pgxpool.New(ctx, cfg.Postgres.WriterDSN())
	if err != nil {
		return nil, err
	}
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}
	conn.DB = dbPool
	logger.Info(ctx, "connected to postgres")

	// Redis
	redisAdapter, redisErr := pkgredis.NewAdapter(pkgredis.Config{
		Addrs:       []string{fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)},
		Password:    cfg.Redis.Pass,
		DB:          cfg.Redis.Database,
		PoolSize:    cfg.Redis.PoolSize,
		DialTimeout: 5 * time.Second,
	})
	if redisErr != nil {
		logger.Error(ctx, "redis connection failed", err)
	} 
	conn.Redis = redisAdapter

	// Kafka producer
	kafkaAdapter, kafkaErr := pkgkafka.NewKafkaService(&pkgkafka.Config{
		Brokers:          cfg.Kafka.GetBrokers(),
		ClientID:         cfg.Kafka.ClientID,
		ProducerOnlyMode: true,
	})
	if kafkaErr != nil {
		logger.Error(ctx, "kafka connection failed", err)
	}
	conn.Kafka = kafkaAdapter

	// Kafka consumer
	kafkaConsumer, kafkaConsumerErr := pkgkafka.NewKafkaService(&pkgkafka.Config{
		Brokers:          cfg.Kafka.GetBrokers(),
		ClientID:         cfg.Kafka.ClientID + "-consumer",
		ConsumerGroupID:  cfg.Kafka.ConsumerGroup,
		ProducerOnlyMode: false,
	})
	if kafkaConsumerErr != nil {
		logger.Error(ctx, "kafka consumer connection failed", kafkaConsumerErr)
	}
	conn.KafkaConsumer = kafkaConsumer

	return conn, nil
}

func (c *Connections) Close(ctx context.Context) {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
}
