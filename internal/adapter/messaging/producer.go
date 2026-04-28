package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"
	pkgkafka "be-modami-user-service/pkg/kafka"

	gokit_kafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type kafkaProducer struct {
	kafkaService *gokit_kafka.KafkaService
	outbox       port.OutboxRepository
	topic        string // pre-computed full topic name
}

// NewKafkaProducer creates a producer-only KafkaService for publishing user events.
func NewKafkaProducer(brokers []string, env string, clientID string, outbox port.OutboxRepository) (*kafkaProducer, error) {
	cfg := &gokit_kafka.Config{
		Brokers:          brokers,
		ClientID:         clientID + "-producer",
		ProducerOnlyMode: true,
	}
	ks, err := gokit_kafka.NewKafkaService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer service: %w", err)
	}

	topic := pkgkafka.GetTopicWithEnv(env, pkgkafka.TopicUserEvents)

	ctx := context.Background()
	if err := ks.EnsureTopics(ctx); err != nil {
		logger.Warn(ctx, "failed to ensure kafka topics", logging.String("error", err.Error()))
	}

	return &kafkaProducer{kafkaService: ks, outbox: outbox, topic: topic}, nil
}

func (p *kafkaProducer) PublishRaw(ctx context.Context, topic, key string, payload []byte) error {
	msg := &gokit_kafka.ProducerMessage{Key: key, Value: json.RawMessage(payload)}
	return p.kafkaService.Emit(ctx, topic, msg)
}

func (p *kafkaProducer) publish(ctx context.Context, key string, value interface{}) error {
	msg := &gokit_kafka.ProducerMessage{Key: key, Value: value}
	if err := p.kafkaService.Emit(ctx, p.topic, msg); err != nil {
		// Fallback: persist to outbox for background retry.
		payload, jsonErr := json.Marshal(value)
		if jsonErr != nil {
			return fmt.Errorf("kafka emit failed: %w; marshal for outbox failed: %v", err, jsonErr)
		}
		if outboxErr := p.outbox.Create(ctx, p.topic, key, payload); outboxErr != nil {
			return fmt.Errorf("kafka emit failed: %w; outbox fallback failed: %v", err, outboxErr)
		}
		return nil
	}
	return nil
}

func (p *kafkaProducer) PublishUserProfileCreated(ctx context.Context, event *domain.UserProfileCreatedEvent) error {
	return p.publish(ctx, event.UserID.String(), event)
}

func (p *kafkaProducer) PublishUserUpdated(ctx context.Context, event *domain.UserUpdatedEvent) error {
	return p.publish(ctx, event.UserID.String(), event)
}

func (p *kafkaProducer) PublishUserRoleUpgraded(ctx context.Context, event *domain.UserRoleUpgradedEvent) error {
	return p.publish(ctx, event.UserID.String(), event)
}

func (p *kafkaProducer) PublishUserSuspended(ctx context.Context, event *domain.UserSuspendedEvent) error {
	return p.publish(ctx, event.UserID.String(), event)
}

func (p *kafkaProducer) PublishUserFollowed(ctx context.Context, event *domain.UserFollowedEvent) error {
	return p.publish(ctx, event.FollowerID.String(), event)
}

func (p *kafkaProducer) PublishUserUnfollowed(ctx context.Context, event *domain.UserUnfollowedEvent) error {
	return p.publish(ctx, event.FollowerID.String(), event)
}

func (p *kafkaProducer) PublishUserReviewCreated(ctx context.Context, event *domain.UserReviewCreatedEvent) error {
	return p.publish(ctx, event.ReviewerID.String(), event)
}

func (p *kafkaProducer) Close() error {
	return p.kafkaService.Close()
}
