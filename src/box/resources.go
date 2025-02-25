package box

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/example/src/interceptors"
	exampleRepo "github.com/ingvarmattis/example/src/repositories/example"
	"github.com/ingvarmattis/example/src/rpctransport"
	exampleSvc "github.com/ingvarmattis/example/src/services/example"
)

type Resources struct {
	ExampleService *exampleSvc.Service

	Validator *validator.Validate

	UnaryServerInterceptors  []grpc.UnaryServerInterceptor
	StreamServerInterceptors []grpc.StreamServerInterceptor
}

func NewResources(ctx context.Context, envBox *Env) (*Resources, error) {
	exampleService, err := exampleSvc.NewService(ctx, exampleRepo.NewPostgres(envBox.PGXPool))
	if err != nil {
		return nil, fmt.Errorf("cannot create example service | %w", err)
	}

	return &Resources{
		ExampleService: exampleService,

		Validator: rpctransport.MustValidate(),

		UnaryServerInterceptors:  provideUnaryInterceptors(envBox),
		StreamServerInterceptors: provideStreamInterceptors(),
	}, nil
}

func provideUnaryInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		interceptors.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors.UnaryServerLogInterceptor(
			envBox.Logger.With("module", "log", "grpc"), envBox.Config.Debug,
		),
		interceptors.UnaryServerPanicsInterceptor(
			envBox.Logger.With("module", "log"), envBox.Config.ServiceName,
		),
	}
}

func provideStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
