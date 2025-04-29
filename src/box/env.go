package box

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"github.com/ingvarmattis/example/src/config"
	"github.com/ingvarmattis/example/src/log"
)

const NotOperational = "noop"

type Env struct {
	Config *config.Config

	PGXPool *pgxpool.Pool

	Logger *log.Zap

	TraceProvider *sdkTrace.TracerProvider
	Tracer        trace.Tracer
}

func NewENV(ctx context.Context) (*Env, error) {
	cfg, err := provideConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot provide config | %w", err)
	}

	pgPool, err := providePGXPool(ctx, cfg.PostgresConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating postgres connection | %w", err)
	}

	logger := provideLogger()

	tracer, traceProvider, err := provideTracer(ctx, cfg.TracingConfig.Enabled, cfg.ServiceName, cfg.TracingConfig.URL, cfg.TracingConfig.UseTLS)
	if err != nil {
		return nil, fmt.Errorf("error creating tracer | %w", err)
	}

	return &Env{
		Config:        cfg,
		PGXPool:       pgPool,
		Logger:        logger,
		Tracer:        tracer,
		TraceProvider: traceProvider,
	}, nil
}

func provideConfig() (*config.Config, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("cannot parse config from environment | %w", err)
	}

	return cfg, nil
}

func providePGXPool(ctx context.Context, connConfig string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	databaseConfig, err := pgxpool.ParseConfig(connConfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing config | %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, databaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating pool | %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging pool | %w", err)
	}

	return pool, nil
}

func provideLogger() *log.Zap {
	return log.NewZap()
}

func provideTracer(
	ctx context.Context, enabled bool, serviceName, openTelemetryCollectorURL string, secureConnection bool,
) (trace.Tracer, *sdkTrace.TracerProvider, error) {
	if !enabled {
		tracer := otel.Tracer(NotOperational)
		return tracer, nil, nil
	}

	exporter, err := otlptracegrpc.New(ctx, func() otlptracegrpc.Option {
		if secureConnection {
			return otlptracegrpc.WithInsecure()
		} else {
			return otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		}
	}(),
		otlptracegrpc.WithEndpoint(openTelemetryCollectorURL),
	)

	if err != nil {
		return nil, nil, err
	}

	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))

	tp := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(exporter),
		sdkTrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return otel.Tracer(serviceName), tp, nil
}
