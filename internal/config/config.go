package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort         string
	GRPCPort           string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	KafkaBrokers       []string
	KafkaConsumerGroup string
	KafkaClientID      string
	KafkaEnv           string // env prefix for topic names (e.g. "local", "prod"); empty = no prefix
	KeycloakJWKSURL    string
	LogLevel           string
	// Observability
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	OTLPInsecure   bool
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("GRPC_PORT", "9090")
	viper.SetDefault("KAFKA_CONSUMER_GROUP", "user-service-group")
	viper.SetDefault("KAFKA_CLIENT_ID", "user-service")
	viper.SetDefault("KAFKA_ENV", "")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("SERVICE_NAME", "user-service")
	viper.SetDefault("SERVICE_VERSION", "1.0.0")
	viper.SetDefault("ENVIRONMENT", "local")
	viper.SetDefault("OTLP_ENDPOINT", "")
	viper.SetDefault("OTLP_INSECURE", true)

	viper.AutomaticEnv()

	brokers := viper.GetString("KAFKA_BROKERS")
	var brokerList []string
	if brokers != "" {
		brokerList = strings.Split(brokers, ",")
	} else {
		brokerList = []string{"localhost:9092"}
	}

	return &Config{
		ServerPort:         viper.GetString("SERVER_PORT"),
		GRPCPort:           viper.GetString("GRPC_PORT"),
		DatabaseURL:        viper.GetString("DATABASE_URL"),
		RedisAddr:          viper.GetString("REDIS_ADDR"),
		RedisPassword:      viper.GetString("REDIS_PASSWORD"),
		RedisDB:            viper.GetInt("REDIS_DB"),
		KafkaBrokers:       brokerList,
		KafkaConsumerGroup: viper.GetString("KAFKA_CONSUMER_GROUP"),
		KafkaClientID:      viper.GetString("KAFKA_CLIENT_ID"),
		KafkaEnv:           viper.GetString("KAFKA_ENV"),
		KeycloakJWKSURL:    viper.GetString("KEYCLOAK_JWKS_URL"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		ServiceName:        viper.GetString("SERVICE_NAME"),
		ServiceVersion:     viper.GetString("SERVICE_VERSION"),
		Environment:        viper.GetString("ENVIRONMENT"),
		OTLPEndpoint:       viper.GetString("OTLP_ENDPOINT"),
		OTLPInsecure:       viper.GetBool("OTLP_INSECURE"),
	}, nil
}
