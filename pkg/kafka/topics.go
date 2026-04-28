package kafka

import "strings"

const (
	TopicAuthUserCreated = "auth.user.created"
	TopicAuthUserUpdated = "auth.user.updated"
	TopicUserEvents      = "user.events"
)

// GetTopicWithEnv returns the topic name prefixed with the environment (e.g. "dev.user.events").
// If env is empty, the base topic is returned as-is.
func GetTopicWithEnv(env, topic string) string {
	if env == "" {
		return topic
	}
	return strings.ToLower(env) + "." + topic
}
