package port

import "context"

type EventPublisher interface {
	PublishRaw(ctx context.Context, topic, key string, payload []byte) error
	Close() error
}
