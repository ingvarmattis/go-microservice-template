package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug                       bool `envconfig:"AUTH_DEBUG" default:"false"`
	GRPCServerListenPort        int  `envconfig:"AUTH_GRPC_SERVER_LISTEN_PORT" required:"true"`
	HTTPServerListenPort        int  `envconfig:"AUTH_HTTP_SERVER_LISTEN_PORT" required:"true"`
	HTTPMetricsServerListenPort int  `envconfig:"AUTH_HTTP_METRICS_SERVER_LISTEN_PORT" required:"true"`

	HostName    string `envconfig:"AUTH_HOST_NAME"`
	ServiceName string `envconfig:"AUTH_SERVICE_NAME"`

	PostgresConfig PostgresConfig
	TracingConfig  TracingConfig
	AuthConfig     AuthConfig
	CryptoConfig   CryptoConfig
}

func FromEnv() (*Config, error) {
	cfg := &Config{}

	if hostName, err := os.Hostname(); err == nil {
		cfg.HostName = hostName
	}

	if err := envconfig.Process("AUTH", cfg); err != nil {
		return nil, fmt.Errorf("error while parsing environment variables | %w", err)
	}

	return cfg, nil
}

type PostgresConfig struct {
	URL string `envconfig:"AUTH_POSTGRES_URL" required:"true"`
}

type TracingConfig struct {
	URL    string `envconfig:"AUTH_OPENTELEMETRY_COLLECTOR_URL" required:"true"`
	UseTLS bool   `envconfig:"AUTH_OPENTELEMETRY_USE_TLS" required:"true"`
}

type OpenAIConfig struct {
	APIKey string `envconfig:"AUTH_OPENAI_API_KEY" required:"true"`
}

type BrowserConfig struct {
	BrowserURLs            []string `envconfig:"AUTH_BROWSER_URLS" required:"true"`
	MaxParallelBrowserTabs int      `envconfig:"AUTH_MAX_PARALLEL_BROWSER_TABS" required:"true"`
}

type AuthConfig struct {
	Token []byte `envconfig:"AUTH_AUTH_TOKEN" required:"true"`
}

type CryptoConfig struct {
	Key []byte `envconfig:"AUTH_CRYPTO_KEY" required:"true"`
}
