package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `envconfig:"EXAMPLE_SERVICE_DEBUG" default:"false"`

	GRPCServerListenPort int `envconfig:"EXAMPLE_SERVICE_GRPC_SERVER_LISTEN_PORT" required:"true"`
	HTTPServerListenPort int `envconfig:"EXAMPLE_SERVICE_HTTP_SERVER_LISTEN_PORT" required:"true"`

	HostName    string `envconfig:"EXAMPLE_SERVICE_HOST_NAME"`
	ServiceName string `envconfig:"EXAMPLE_SERVICE_SERVICE_NAME"`

	PostgresConfig PostgresConfig
	MetricsConfig  MetricsConfig
	TracingConfig  TracingConfig
}

func FromEnv() (*Config, error) {
	cfg := &Config{}

	if hostName, err := os.Hostname(); err == nil {
		cfg.HostName = hostName
	}

	if err := envconfig.Process("EXAMPLE_SERVICE", cfg); err != nil {
		return nil, fmt.Errorf("error while parsing environment variables | %w", err)
	}

	return cfg, nil
}

type PostgresConfig struct {
	URL string `envconfig:"EXAMPLE_SERVICE_POSTGRES_URL" required:"true"`
}

type MetricsConfig struct {
	Enabled bool `envconfig:"EXAMPLE_SERVICE_METRICS_ENABLED" required:"true"`
	Port    int  `envconfig:"EXAMPLE_SERVICE_HTTP_METRICS_SERVER_LISTEN_PORT" required:"true"`
}

type TracingConfig struct {
	Enabled bool   `envconfig:"EXAMPLE_SERVICE_OPENTELEMETRY_ENABLED" required:"true"`
	URL     string `envconfig:"EXAMPLE_SERVICE_OPENTELEMETRY_COLLECTOR_URL" required:"true"`
	UseTLS  bool   `envconfig:"EXAMPLE_SERVICE_OPENTELEMETRY_USE_TLS" required:"true"`
}

type OpenAIConfig struct {
	APIKey string `envconfig:"EXAMPLE_SERVICE_OPENAI_API_KEY" required:"true"`
}
