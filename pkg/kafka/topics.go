package kafka

// Topic names (base, without env prefix — prepend env with GetTopicWithEnv)
const (
	// Consumed: events produced by auth-service
	TopicAuthUserCreated = "modami.auth.user.created" // user registered in Keycloak
	TopicAuthUserUpdated = "modami.auth.user.updated" // user profile updated in Keycloak

	// Produced: events published by user-service
	TopicUserEvents    = "modami.user.events"     // all user service domain events
	TopicUserEventsDLQ = "modami.user.events.dlq" // dead letter queue
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
func GetAllTopics(env string) []string {
	return []string{
		GetTopicWithEnv(env, TopicAuthUserCreated),
		GetTopicWithEnv(env, TopicAuthUserUpdated),
		GetTopicWithEnv(env, TopicUserEvents),
		GetTopicWithEnv(env, TopicUserEventsDLQ),
	}
}
