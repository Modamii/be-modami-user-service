package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	GRPC          GRPCConfig          `mapstructure:"grpc"`
	Postgres      PostgresConfig      `mapstructure:"postgres"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Keycloak      KeycloakConfig      `mapstructure:"keycloak"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

type AppConfig struct {
	Name             string   `mapstructure:"name"`
	Version          string   `mapstructure:"version"`
	Environment      string   `mapstructure:"environment"`
	Debug            bool     `mapstructure:"debug"`
	Port             int      `mapstructure:"port"`
	Host             string   `mapstructure:"host"`
	SwaggerHost      string   `mapstructure:"swagger_host"`
	ShutdownTimeout  string   `mapstructure:"shutdown_timeout"`
	ReadTimeout      string   `mapstructure:"read_timeout"`
	WriteTimeout     string   `mapstructure:"write_timeout"`
	IdleTimeout      string   `mapstructure:"idle_timeout"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
}

func (a AppConfig) ListenAddr() string {
	host := strings.TrimSpace(a.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := a.Port
	if port <= 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func (a AppConfig) GetShutdownTimeout() time.Duration {
	d, err := time.ParseDuration(a.ShutdownTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetReadTimeout() time.Duration {
	d, err := time.ParseDuration(a.ReadTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetWriteTimeout() time.Duration {
	d, err := time.ParseDuration(a.WriteTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetIdleTimeout() time.Duration {
	d, err := time.ParseDuration(a.IdleTimeout)
	if err != nil {
		return 120 * time.Second
	}
	return d
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

// MigrationDSN returns a DSN compatible with lib/pq (used by golang-migrate).
// It normalizes sslmode to values accepted by lib/pq and omits search_path so
// the migrations table is created in the public schema.
func (p PostgresConfig) MigrationDSN() string {
	sslmode := p.SSLMode
	switch sslmode {
	case "", "required", "prefer", "allow":
		sslmode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.UserWriter, p.PasswordWriter, p.Host, p.Port, p.Database, sslmode)
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
	v.SetDefault("app.name", "be-modami-user-service")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "local")
	v.SetDefault("app.debug", false)
	v.SetDefault("app.port", 8086)
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.swagger_host", "localhost:8086")
	v.SetDefault("app.shutdown_timeout", "30s")
	v.SetDefault("app.read_timeout", "30s")
	v.SetDefault("app.write_timeout", "30s")
	v.SetDefault("app.idle_timeout", "120s")
	v.SetDefault("app.allow_credentials", true)
	v.SetDefault("app.allowed_origins", []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:8080",
		"http://localhost:8081",
	})
	v.SetDefault("postgres.max_idle_conns", 5)
	v.SetDefault("postgres.max_active_conns", 25)
	v.SetDefault("postgres.sslmode", "disable")
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
