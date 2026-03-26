package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaService struct {
	config  *KafkaConfig
	env     string
	client  *kgo.Client
	mu      sync.RWMutex
	running bool
}

func NewKafkaService(cfg *KafkaConfig, env string) (*KafkaService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config is required")
	}

	opts, err := cfg.ToFranzGoOpts()
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka options: %w", err)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &KafkaService{
		config:  cfg,
		env:     env,
		client:  client,
		running: false,
	}, nil
}

type ProducerMessage struct {
	Key     string                 `json:"key"`
	Value   interface{}            `json:"value"`
	Headers map[string]interface{} `json:"headers,omitempty"`
}

// Emit sends a message to Kafka synchronously with trace context propagation.
func (k *KafkaService) Emit(ctx context.Context, topic string, message *ProducerMessage) error {
	topicName := GetTopicWithEnv(k.env, topic)

	valueBytes, err := json.Marshal(message.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	headers := k.buildHeaders(ctx, message.Headers)
	headers = InjectTraceContext(ctx, headers)

	record := &kgo.Record{
		Topic:   topicName,
		Key:     []byte(message.Key),
		Value:   valueBytes,
		Headers: headers,
	}

	if err := k.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		slog.ErrorContext(ctx, "failed to send message", "error", err, "topic", topicName, "key", message.Key)
		return fmt.Errorf("failed to send message to topic %s: %w", topicName, err)
	}

	slog.InfoContext(ctx, "message sent successfully", "topic", topicName, "key", message.Key)
	return nil
}

// EmitAsync sends a message asynchronously, detached from the request context.
func (k *KafkaService) EmitAsync(ctx context.Context, topic string, message *ProducerMessage) {
	topicName := GetTopicWithEnv(k.env, topic)

	valueBytes, err := json.Marshal(message.Value)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal message value for async emit", "error", err)
		return
	}

	headers := k.buildHeaders(ctx, message.Headers)
	headers = InjectTraceContext(ctx, headers)

	record := &kgo.Record{
		Topic:   topicName,
		Key:     []byte(message.Key),
		Value:   valueBytes,
		Headers: headers,
	}

	produceCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)

	k.client.Produce(produceCtx, record, func(r *kgo.Record, err error) {
		defer cancel()
		if err != nil {
			slog.ErrorContext(produceCtx, "failed to send async message", "error", err, "topic", topicName, "key", message.Key)
			return
		}
		slog.DebugContext(produceCtx, "async message sent successfully", "topic", topicName, "key", message.Key)
	})
}

// SendMessages sends multiple messages to a topic synchronously.
func (k *KafkaService) SendMessages(ctx context.Context, topic string, messages []*ProducerMessage) error {
	topicName := GetTopicWithEnv(k.env, topic)
	records := make([]*kgo.Record, 0, len(messages))

	for _, message := range messages {
		valueBytes, err := json.Marshal(message.Value)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal message value", "error", err)
			continue
		}

		headers := k.buildHeaders(ctx, message.Headers)
		headers = InjectTraceContext(ctx, headers)

		records = append(records, &kgo.Record{
			Topic:   topicName,
			Key:     []byte(message.Key),
			Value:   valueBytes,
			Headers: headers,
		})
	}

	if err := k.client.ProduceSync(ctx, records...).FirstErr(); err != nil {
		slog.ErrorContext(ctx, "failed to send messages", "error", err, "topic", topicName, "count", len(messages))
		return fmt.Errorf("failed to send messages to topic %s: %w", topicName, err)
	}

	slog.InfoContext(ctx, "messages sent successfully", "topic", topicName, "count", len(messages))
	return nil
}

// EnsureTopics creates missing topics and (when env is set) removes unknown ones.
func (k *KafkaService) EnsureTopics(ctx context.Context) error {
	adm := kadm.NewClient(k.client)

	baseTopics := GetAllTopics()
	targetTopics := make([]string, 0, len(baseTopics))
	for _, t := range baseTopics {
		targetTopics = append(targetTopics, GetTopicWithEnv(k.env, t))
	}

	slog.InfoContext(ctx, "ensuring kafka topics exist", "count", len(targetTopics))

	metadata, err := adm.Metadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to get kafka metadata: %w", err)
	}

	brokerCount := len(metadata.Brokers)
	replicationFactor := int16(3)
	if brokerCount < 3 {
		replicationFactor = 1
		slog.WarnContext(ctx, "broker count < 3, using replication factor 1", "brokers", brokerCount)
	}

	existingTopics, err := adm.ListTopics(ctx)
	if err != nil {
		return fmt.Errorf("failed to list kafka topics: %w", err)
	}

	missingTopics := make([]string, 0)
	for _, t := range targetTopics {
		if !existingTopics.Has(t) {
			missingTopics = append(missingTopics, t)
		}
	}

	// Only clean up redundant topics when env is set (to avoid wiping shared topics in bare mode)
	if k.env != "" {
		envPrefix := GetTopicWithEnv(k.env, "")
		redundantTopics := make([]string, 0)
		for _, t := range existingTopics.Names() {
			if !strings.HasPrefix(t, envPrefix) {
				continue
			}
			isTarget := false
			for _, target := range targetTopics {
				if t == target {
					isTarget = true
					break
				}
			}
			if !isTarget {
				redundantTopics = append(redundantTopics, t)
			}
		}

		if len(redundantTopics) > 0 {
			slog.InfoContext(ctx, "deleting redundant kafka topics", "count", len(redundantTopics), "topics", strings.Join(redundantTopics, ","))
			delResp, err := adm.DeleteTopics(ctx, redundantTopics...)
			if err != nil {
				slog.ErrorContext(ctx, "failed to delete redundant topics", "error", err)
			} else {
				for _, res := range delResp {
					if res.Err != nil {
						slog.ErrorContext(ctx, "failed to delete redundant topic", "error", res.Err, "topic", res.Topic)
					} else {
						slog.InfoContext(ctx, "deleted redundant topic", "topic", res.Topic)
					}
				}
			}
		}
	}

	if len(missingTopics) == 0 {
		slog.InfoContext(ctx, "all required kafka topics already exist")
		return nil
	}

	slog.InfoContext(ctx, "creating missing kafka topics", "count", len(missingTopics), "replication", replicationFactor)
	resp, err := adm.CreateTopics(ctx, 1, replicationFactor, nil, missingTopics...)
	if err != nil {
		return fmt.Errorf("failed to create missing topics: %w", err)
	}

	hasError := false
	for _, res := range resp {
		if res.Err != nil {
			slog.ErrorContext(ctx, "failed to create topic", "error", res.Err, "topic", res.Topic)
			hasError = true
		} else {
			slog.InfoContext(ctx, "created topic", "topic", res.Topic)
		}
	}

	if hasError {
		return fmt.Errorf("some topics failed to be created")
	}
	return nil
}

// StartConsumer runs the consumer loop, dispatching each record to registered handlers.
// Blocks until ctx is cancelled.
func (k *KafkaService) StartConsumer(ctx context.Context, handlers []ConsumerHandler) error {
	k.mu.Lock()
	if k.running {
		k.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}
	k.running = true
	k.mu.Unlock()

	defer func() {
		k.mu.Lock()
		k.running = false
		k.mu.Unlock()
	}()

	handlerMap := make(map[string][]ConsumerHandler)
	var topics []string
	topicSet := make(map[string]struct{})

	for _, handler := range handlers {
		for _, topic := range handler.GetTopics() {
			topicName := GetTopicWithEnv(k.env, topic)
			if _, exists := topicSet[topicName]; !exists {
				topicSet[topicName] = struct{}{}
				topics = append(topics, topicName)
			}
			handlerMap[topicName] = append(handlerMap[topicName], handler)
		}
	}

	k.client.AddConsumeTopics(topics...)
	slog.InfoContext(ctx, "starting consumer group", "topics", strings.Join(topics, ","))

	for {
		fetches := k.client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			if ctx.Err() != nil {
				slog.InfoContext(ctx, "consumer context cancelled")
				return nil
			}
			slog.ErrorContext(ctx, "consumer poll error", "error", err)
			time.Sleep(time.Second)
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			msgCtx := k.extractContextFromHeaders(record.Headers)
			hs, exists := handlerMap[record.Topic]
			if !exists {
				slog.WarnContext(msgCtx, "no handlers found for topic", "topic", record.Topic)
				continue
			}
			for _, h := range hs {
				if err := h.HandleMessage(msgCtx, record); err != nil {
					slog.ErrorContext(msgCtx, "failed to handle message", "error", err, "topic", record.Topic)
				}
			}
		}
	}
}

// Close shuts down the Kafka client.
func (k *KafkaService) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.client.Close()
	k.running = false
	return nil
}

func (k *KafkaService) IsRunning() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.running
}

func (k *KafkaService) Ping(ctx context.Context) error {
	if k.client == nil {
		return fmt.Errorf("kafka client is not initialized")
	}
	record := &kgo.Record{
		Topic: "__healthcheck",
		Key:   []byte("health"),
		Value: []byte("ping"),
	}
	if err := k.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("kafka ping failed: %w", err)
	}
	return nil
}

func (k *KafkaService) buildHeaders(ctx context.Context, customHeaders map[string]interface{}) []kgo.RecordHeader {
	headers := make([]kgo.RecordHeader, 0, len(customHeaders)+1)

	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		headers = append(headers, kgo.RecordHeader{Key: "request-id", Value: []byte(requestID)})
	}

	for key, value := range customHeaders {
		headerBytes, err := json.Marshal(value)
		if err != nil {
			slog.WarnContext(ctx, "failed to marshal header value", "key", key, "error", err)
			continue
		}
		headers = append(headers, kgo.RecordHeader{Key: key, Value: headerBytes})
	}

	return headers
}

func (k *KafkaService) extractContextFromHeaders(headers []kgo.RecordHeader) context.Context {
	ctx := context.Background()
	ctx = ExtractTraceContext(ctx, headers)
	for _, header := range headers {
		if header.Key == "request-id" {
			ctx = setRequestIDInContext(ctx, string(header.Value))
			break
		}
	}
	return ctx
}

type requestIDKeyType string

const requestIDCtxKey requestIDKeyType = "kafka_request_id"

func getRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDCtxKey).(string); ok {
		return id
	}
	return ""
}

func setRequestIDInContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDCtxKey, id)
}
