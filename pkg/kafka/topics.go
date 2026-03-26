package kafka

// Topic names (base, without env prefix)
const (
	TopicAuthEvents    = "auth.events"     // consumed: user.registered, user.deleted, user.email_verified
	TopicUserEvents    = "user.events"     // produced: all user service events
	TopicUserEventsDLQ = "user.events.dlq" // dead letter queue
)

// GetTopicWithEnv returns the fully-qualified topic name.
// If env is empty, the base topic name is returned unchanged.
func GetTopicWithEnv(env, topic string) string {
	if env == "" {
		return topic
	}
	return env + "." + topic
}

// GetAllTopics returns all topics managed by this service.
func GetAllTopics() []string {
	return []string{
		TopicAuthEvents,
		TopicUserEvents,
		TopicUserEventsDLQ,
	}
}
