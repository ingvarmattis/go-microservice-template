package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"gitlab.com/ingvarmattis/auth/gen/servergrpc/server"
	"gitlab.com/ingvarmattis/auth/src/box"
	"gitlab.com/ingvarmattis/auth/src/interceptors"
	"gitlab.com/ingvarmattis/auth/src/log"
	"gitlab.com/ingvarmattis/auth/src/repositories/user"
	"gitlab.com/ingvarmattis/auth/src/rpctransport"
	authrpc "gitlab.com/ingvarmattis/auth/src/rpctransport/user"
	"gitlab.com/ingvarmattis/auth/src/services"
	authSVC "gitlab.com/ingvarmattis/auth/src/services/user"
)

func main() {
	serverCTX, serverCancel := context.WithCancel(context.Background())

	envBox, err := box.NewENV(serverCTX)
	if err != nil {
		panic(err)
	}

	logger := envBox.Logger

	userStorage := user.NewPostgres(envBox.PGXPool, envBox.Encryptor)
	userService := authSVC.NewService(envBox.Config.AuthConfig.Token, userStorage)

	grpcCompetitorsServer := server.NewGRPCServer(
		serverCTX,
		envBox.Config.GRPCServerListenPort,
		&server.NewServerOptions{
			ServiceName:      envBox.Config.ServiceName,
			GRPCAuthHandlers: &authrpc.Handlers{Service: services.SvcLayer{AuthService: userService}},
			Validator:        rpctransport.MustValidate(),
			Logger:           logger,
			UnaryInterceptors: []grpc.UnaryServerInterceptor{
				interceptors.UnaryServerAuthInterceptor(envBox.Config.AuthConfig.Token),
				interceptors.UnaryServerMetricsInterceptor(envBox.Config.ServiceName),
				interceptors.UnaryServerTraceInterceptor(envBox.Tracer, envBox.Config.ServiceName),
				interceptors.UnaryServerLogInterceptor(logger.With("module", "log", "grpc"), envBox.Config.Debug),
				interceptors.UnaryServerPanicsInterceptor(logger.With("module", "log"), envBox.Config.ServiceName)},
			StreamInterceptors: nil,
		})

	metricsServer := server.NewMetricsServer(logger, envBox.Config.HTTPMetricsServerListenPort)

	// working functions
	workingFunctions := []func() error{
		func() error {
			if grpcServerErr := grpcCompetitorsServer.Serve(
				envBox.Config.ServiceName, &envBox.Config.GRPCServerListenPort,
			); grpcServerErr != nil {
				return fmt.Errorf("cannot start grpc server | %w", grpcServerErr)
			}

			return nil
		},
		func() error {
			if httpServerErr := grpcCompetitorsServer.ServeHTTP(&envBox.Config.HTTPServerListenPort); err != nil {
				return fmt.Errorf("cannot start http server | %w", httpServerErr)
			}

			return nil
		},
		func() error {
			if httpMetricsErr := metricsServer.ListenAndServe(); httpMetricsErr != nil {
				return fmt.Errorf("cannot start http metrics server | %w", httpMetricsErr)
			}

			return nil
		},
	}

	for i := range len(workingFunctions) {
		go func() {
			if err = workingFunctions[i](); err != nil {
				logger.Error("working function failed", zap.Error(err))
				os.Exit(1)
			}
		}()
	}

	gracefullShutdown(
		logger,
		grpcCompetitorsServer, envBox.PGXPool,
		metricsServer,
		envBox.TraceProvider,
	)

	serverCancel()

	logger.Info("service has been shutdown")
}

type (
	closer interface {
		Close()
	}
	closerWithErr interface {
		Close() error
	}
	shutdowner interface {
		Shutdown(ctx context.Context) error
	}
)

func gracefullShutdown(
	logger *log.Zap,
	serverGRPC, pgxPool closer,
	metricsServerHTTP closerWithErr,
	traceProvider shutdowner,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT,
	)
	<-quit

	logger.Info("shutting down service...")

	shutdownWG := &sync.WaitGroup{}
	shutdownFunctions := []func(){
		func() {
			defer shutdownWG.Done()

			serverGRPC.Close()
		},
		func() {
			defer shutdownWG.Done()

			if err := metricsServerHTTP.Close(); err != nil {
				logger.Error("failed to close metrics server", zap.Error(err))
			}
		},
		func() {
			defer shutdownWG.Done()

			pgxPool.Close()
		},
		func() {
			defer shutdownWG.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			if err := traceProvider.Shutdown(ctx); err != nil {
				logger.Error("failed to close tracer", zap.Error(err))
			}
		},
		func() {
			defer shutdownWG.Done()

			if err := logger.Close(); err != nil {
				logger.Error("failed to close logger", zap.Error(err))
			}
		},
	}
	shutdownWG.Add(len(shutdownFunctions))

	for _, shutdown := range shutdownFunctions {
		go shutdown()
	}

	shutdownWG.Wait()
}
