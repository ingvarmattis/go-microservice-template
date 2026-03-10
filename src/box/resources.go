package box

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/example/gen/servergrpc/server"
	"github.com/ingvarmattis/example/src/interceptors"
	exampleRepo "github.com/ingvarmattis/example/src/repositories/example"
	"github.com/ingvarmattis/example/src/rpctransport"
	exampleRPC "github.com/ingvarmattis/example/src/rpctransport/example"
	"github.com/ingvarmattis/example/src/services"
	exampleSvc "github.com/ingvarmattis/example/src/services/example"
)

// TelegramBotInterface is the contract for telegram bot (real or noop). Use NotifyMessage for notifications.
type TelegramBotInterface interface {
	NotifyMessage(msg string)
	Start()
	Close()
}

type Resources struct {
	ExampleService *exampleSvc.Service

	Validator *validator.Validate

	UnaryServerInterceptors  []grpc.UnaryServerInterceptor
	StreamServerInterceptors []grpc.StreamServerInterceptor

	GRPCServer    *server.Server
	TelegramBot   TelegramBotInterface
	MetricsServer *server.MetricsServer
}

func NewResources(ctx context.Context, envBox *Env) (*Resources, error) {
	exampleService, err := exampleSvc.NewService(ctx, exampleRepo.NewPostgres(envBox.PGXPool))
	if err != nil {
		return nil, fmt.Errorf("cannot create example service | %w", err)
	}

	validator := rpctransport.MustValidate()
	unaryInterceptors := provideUnaryInterceptors(envBox)
	streamInterceptors := provideStreamInterceptors()

	telegramBot, err := provideTelegramBot(envBox)
	if err != nil {
		return nil, err
	}

	grpcServer := provideGRPCServer(ctx, envBox, exampleService, validator, unaryInterceptors, streamInterceptors)
	metricsServer := provideMetricsServer(envBox)

	return &Resources{
		ExampleService: exampleService,

		Validator: validator,

		UnaryServerInterceptors:  unaryInterceptors,
		StreamServerInterceptors: streamInterceptors,

		GRPCServer:    grpcServer,
		TelegramBot:   telegramBot,
		MetricsServer: metricsServer,
	}, nil
}

func provideGRPCServer(
	ctx context.Context,
	envBox *Env,
	exampleService *exampleSvc.Service,
	validator *validator.Validate,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
) *server.Server {
	return server.NewServer(
		ctx,
		envBox.Config.GRPCServerListenPort,
		&server.NewServerOptions{
			ServiceName: envBox.Config.ServiceName,
			GRPCExampleHandlers: &exampleRPC.Handlers{
				Service: services.SvcLayer{ExampleService: exampleService},
			},
			Validator:          validator,
			Logger:             envBox.Logger,
			UnaryInterceptors:  unaryInterceptors,
			StreamInterceptors: streamInterceptors,
		},
	)
}

func provideMetricsServer(envBox *Env) *server.MetricsServer {
	return server.NewMetricsServer(
		envBox.Config.MetricsConfig.Enabled, envBox.Logger, envBox.Config.MetricsConfig.Port,
	)
}

func provideTelegramBot(envBox *Env) (TelegramBotInterface, error) {
	if !envBox.Config.TelegramConfig.Enabled {
		return server.NewNoopTelegramBot(), nil
	}

	bot, err := server.NewTelegramBot(
		envBox.Logger.Zap(),
		envBox.Config.TelegramConfig.Token,
		envBox.Config.TelegramConfig.Timeout,
		envBox.Config.TelegramConfig.AllowedChatIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("provide telegram bot | %w", err)
	}

	return bot, nil
}

func provideUnaryInterceptors(envBox *Env) []grpc.UnaryServerInterceptor {
	logger := envBox.Logger.WithFields(zap.String("type", "unary"))

	return []grpc.UnaryServerInterceptor{
		interceptors.UnaryServerMetricsInterceptor(envBox.Config.MetricsConfig.Enabled, envBox.Config.ServiceName),
		interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
		interceptors.UnaryServerLogInterceptor(logger, envBox.Config.Debug),
		interceptors.UnaryServerPanicsInterceptor(logger, envBox.Config.ServiceName),
	}
}

func provideStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}
