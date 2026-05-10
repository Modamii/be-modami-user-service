package kafka

import "strings"

const (
	TopicAuthUserCreated = "auth.user.created"
	TopicAuthUserUpdated = "auth.user.updated"

	// Outbox CDC topics — produced by Debezium EventRouter SMT from outbox_events.
	// aggregate_type + ".events" = topic name (e.g. "user" → "user.events").
	TopicUserEvents   = "user.events"
	TopicFollowEvents = "follow.events"
	TopicReviewEvents = "review.events"
	TopicKYCEvents    = "kyc.events"

	// TopicKeycloakCDCUserEntity is the Debezium CDC topic for the Keycloak
	// user_entity table. The name is fixed by the connector config and is never
	// prefixed with an environment string.
	TopicKeycloakCDCUserEntity = "keycloak-cdc.public.user_entity"
)

// GetTopicWithEnv returns the topic name prefixed with the environment (e.g. "dev.user.events").
// If env is empty, the base topic is returned as-is.
func GetTopicWithEnv(env, topic string) string {
	if env == "" {
		return topic
	}
	return strings.ToLower(env) + "." + topic
}
