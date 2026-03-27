package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	GRPC          GRPCConfig          `mapstructure:"grpc"`
	Postgres      PostgresConfig      `mapstructure:"postgres"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Keycloak      KeycloakConfig      `mapstructure:"keycloak"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type GRPCConfig struct {
	Port string `mapstructure:"port"`
}

type PostgresConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Schema         string        `mapstructure:"schema"`
	UserReader     string        `mapstructure:"user_reader"`
	PasswordReader string        `mapstructure:"password_reader"`
	UserWriter     string        `mapstructure:"user_writer"`
	PasswordWriter string        `mapstructure:"password_writer"`
	Database       string        `mapstructure:"database"`
	SSLMode        string        `mapstructure:"sslmode"`
	MaxIdleConns   int           `mapstructure:"max_idle_conns"`
	MaxActiveConns int           `mapstructure:"max_active_conns"`
	MaxConnTimeout time.Duration `mapstructure:"max_conn_timeout"`
	DebugLog       bool          `mapstructure:"debug_log"`
}

func (p PostgresConfig) WriterDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		p.UserWriter, p.PasswordWriter, p.Host, p.Port, p.Database, p.SSLMode, p.Schema)
}

func (p PostgresConfig) ReaderDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		p.UserReader, p.PasswordReader, p.Host, p.Port, p.Database, p.SSLMode, p.Schema)
}

type RedisTLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

type RedisConfig struct {
	Host              string         `mapstructure:"host"`
	Port              int            `mapstructure:"port"`
	Database          int            `mapstructure:"database"`
	RateLimitDatabase int            `mapstructure:"rate_limit_database"`
	TTL               time.Duration  `mapstructure:"ttl"`
	PoolSize          int            `mapstructure:"pool_size"`
	Pass              string         `mapstructure:"pass"`
	UserName          string         `mapstructure:"user_name"`
	WriteTimeout      time.Duration  `mapstructure:"write_timeout"`
	ReadTimeout       time.Duration  `mapstructure:"read_timeout"`
	DialTimeout       time.Duration  `mapstructure:"dial_timeout"`
	TLSConfig         RedisTLSConfig `mapstructure:"tls_config"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	BrokerList             string `mapstructure:"broker_list"`
	Enable                 bool   `mapstructure:"enable"`
	TLSEnable              bool   `mapstructure:"tls_enable"`
	Partition              int    `mapstructure:"partition"`
	Partitioner            string `mapstructure:"partitioner"`
	SASLProducerUsername   string `mapstructure:"sasl_producer_username"`
	SASLProducerPassword   string `mapstructure:"sasl_producer_password"`
	SASLConsumerUsername   string `mapstructure:"sasl_consumer_username"`
	SASLConsumerPassword   string `mapstructure:"sasl_consumer_password"`
	UserActivatedTopicName string `mapstructure:"user_activated_topic_name"`
	ConsumerGroup          string `mapstructure:"consumer_group"`
	ClientID               string `mapstructure:"client_id"`
	Env                    string `mapstructure:"env"`
}

func (k KafkaConfig) Brokers() []string {
	return strings.Split(k.BrokerList, ",")
}

type KeycloakConfig struct {
	JWKSURL string `mapstructure:"jwks_url"`
}

type ObservabilityConfig struct {
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
	Environment    string `mapstructure:"environment"`
	LogLevel       string `mapstructure:"log_level"`
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	OTLPInsecure   bool   `mapstructure:"otlp_insecure"`
}

func Load() (*Config, error) {
	v := viper.New()

	// YAML config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.shutdown_timeout", "15s")
	v.SetDefault("postgres.max_idle_conns", 5)
	v.SetDefault("postgres.max_active_conns", 25)
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("app.name", "be-modami-auth-service")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Read config.yml
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Env vars override YAML (SERVER_PORT overrides server.port)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
