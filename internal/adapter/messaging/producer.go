package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/port"
	pkgkafka "github.com/modami/user-service/pkg/kafka"
)

type kafkaProducer struct {
	kafkaService *pkgkafka.KafkaService
	outbox       port.OutboxRepository
	env          string
}

// NewKafkaProducer creates a producer-only KafkaService for publishing user.events.
func NewKafkaProducer(brokers []string, env string, clientID string, outbox port.OutboxRepository) (*kafkaProducer, error) {
	cfg := &pkgkafka.KafkaConfig{
		Brokers:          brokers,
		ClientID:         clientID + "-producer",
		ProducerOnlyMode: true,
	}
	ks, err := pkgkafka.NewKafkaService(cfg, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer service: %w", err)
	}
	return &kafkaProducer{kafkaService: ks, outbox: outbox, env: env}, nil
}

func (p *kafkaProducer) publish(ctx context.Context, key string, value interface{}) error {
	msg := &pkgkafka.ProducerMessage{Key: key, Value: value}
	if err := p.kafkaService.Emit(ctx, pkgkafka.TopicUserEvents, msg); err != nil {
		// Fallback: persist to outbox for background retry.
		payload, jsonErr := json.Marshal(value)
		if jsonErr != nil {
			return fmt.Errorf("kafka emit failed: %w; marshal for outbox failed: %v", err, jsonErr)
		}
		topic := pkgkafka.GetTopicWithEnv(p.env, pkgkafka.TopicUserEvents)
		if outboxErr := p.outbox.Create(ctx, topic, key, payload); outboxErr != nil {
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
