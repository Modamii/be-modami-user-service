package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"
	"be-modami-user-service/internal/service"
	pkgkafka "be-modami-user-service/pkg/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

// Consumer wraps a pkg/kafka.KafkaService and dispatches auth.events to the correct handlers.
type Consumer struct {
	kafkaService *pkgkafka.KafkaService
	handler      *authEventsHandler
}

// NewConsumer creates a Kafka consumer configured for the auth.events topic.
func NewConsumer(
	brokers []string,
	groupID string,
	env string,
	clientID string,
	processedRepo port.ProcessedEventRepository,
	userService *service.UserService,
) (*Consumer, error) {
	cfg := &pkgkafka.KafkaConfig{
		Brokers:          brokers,
		ClientID:         clientID + "-consumer",
		ConsumerGroupID:  groupID,
		ProducerOnlyMode: false,
	}
	ks, err := pkgkafka.NewKafkaService(cfg, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer service: %w", err)
	}
	return &Consumer{
		kafkaService: ks,
		handler: &authEventsHandler{
			userService:   userService,
			processedRepo: processedRepo,
		},
	}, nil
}

// Start runs the consumer loop; blocks until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context) {
	if err := c.kafkaService.StartConsumer(ctx, []pkgkafka.ConsumerHandler{c.handler}); err != nil {
		logger.Error(ctx, "kafka consumer exited", err)
	}
}

func (c *Consumer) Close() error {
	return c.kafkaService.Close()
}

// authEventsHandler implements pkgkafka.ConsumerHandler for auth user topics.
type authEventsHandler struct {
	userService   *service.UserService
	processedRepo port.ProcessedEventRepository
}

func (h *authEventsHandler) GetTopics() []string {
	return []string{
		pkgkafka.TopicAuthUserCreated,
		pkgkafka.TopicAuthUserUpdated,
	}
}

func (h *authEventsHandler) HandleMessage(ctx context.Context, record *kgo.Record) error {
	// Derive event ID from headers, falling back to partition+offset.
	eventID := fmt.Sprintf("%s-%d-%d", record.Topic, record.Partition, record.Offset)
	for _, hdr := range record.Headers {
		if hdr.Key == "event_id" {
			eventID = string(hdr.Value)
		}
	}

	// Idempotency check.
	processed, err := h.processedRepo.IsProcessed(ctx, eventID)
	if err != nil {
		logger.Error(ctx, "idempotency check failed", err, logging.String("event_id", eventID))
	}
	if processed {
		return nil
	}

	// Peek at the event type without full deserialization.
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(record.Value, &envelope); err != nil {
		return fmt.Errorf("failed to decode event type: %w", err)
	}

	var handlerErr error
	switch envelope.Type {
	case "user.created":
		var event domain.AuthUserCreatedEvent
		if err := json.Unmarshal(record.Value, &event); err != nil {
			return fmt.Errorf("failed to unmarshal user.created: %w", err)
		}
		handlerErr = h.userService.CreateFromEvent(ctx, &event)

	case "user.updated":
		var event domain.AuthUserUpdatedEvent
		if err := json.Unmarshal(record.Value, &event); err != nil {
			return fmt.Errorf("failed to unmarshal user.updated: %w", err)
		}
		handlerErr = h.userService.SyncFromAuthUpdate(ctx, &event)

	default:
		logger.Warn(ctx, "unknown event type, skipping", logging.String("type", envelope.Type))
		return nil
	}

	if handlerErr != nil {
		return handlerErr
	}

	if err := h.processedRepo.MarkProcessed(ctx, eventID, record.Topic); err != nil {
		logger.Error(ctx, "failed to mark event processed", err, logging.String("event_id", eventID))
	}
	return nil
}
