package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"be-modami-user-service/internal/domain"
	"be-modami-user-service/internal/port"
	"be-modami-user-service/internal/service"
	pkgkafka "be-modami-user-service/pkg/kafka"

	gokit_kafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

// keycloakCDCHandler implements gokit_kafka.ConsumerHandler for Debezium CDC
type keycloakCDCHandler struct {
	userService   *service.UserService
	processedRepo port.ProcessedEventRepository
}

// NewKeycloakCDCHandler returns a ConsumerHandler that syncs Keycloak
func NewKeycloakCDCHandler(
	userService *service.UserService,
	processedRepo port.ProcessedEventRepository,
) gokit_kafka.ConsumerHandler {
	return &keycloakCDCHandler{
		userService:   userService,
		processedRepo: processedRepo,
	}
}

func (h *keycloakCDCHandler) GetTopics() []string {
	return []string{pkgkafka.TopicKeycloakCDCUserEntity}
}

func (h *keycloakCDCHandler) HandleMessage(ctx context.Context, msg *gokit_kafka.Message) error {
	eventID := fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)

	processed, err := h.processedRepo.IsProcessed(ctx, eventID)
	if err != nil {
		logger.Error(ctx, "keycloak-cdc: idempotency check failed", err,
			logging.String("event_id", eventID))
	}
	if processed {
		return nil
	}

	if len(msg.Value) == 0 {
		// Tombstone message (null value) used by Debezium for log compaction — skip silently.
		return nil
	}

	var event domain.KeycloakCDCEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		// A parse failure is unrecoverable for this message; log and skip.
		logger.Error(ctx, "keycloak-cdc: failed to unmarshal CDC envelope", err,
			logging.String("event_id", eventID))
		return fmt.Errorf("keycloak-cdc: unmarshal: %w", err)
	}

	logger.Info(ctx, "keycloak-cdc: processing event",
		logging.String("op", string(event.Op)),
		logging.String("event_id", eventID),
	)

	if err := h.userService.SyncFromKeycloakCDC(ctx, &event); err != nil {
		return fmt.Errorf("keycloak-cdc: sync op=%s: %w", event.Op, err)
	}

	if err := h.processedRepo.MarkProcessed(ctx, eventID, msg.Topic); err != nil {
		logger.Error(ctx, "keycloak-cdc: failed to mark event processed", err,
			logging.String("event_id", eventID))
	}
	return nil
}
