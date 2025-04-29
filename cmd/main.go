package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ingvarmattis/example/gen/servergrpc/server"
	"github.com/ingvarmattis/example/src/box"
	"github.com/ingvarmattis/example/src/log"
	exampleRPC "github.com/ingvarmattis/example/src/rpctransport/example"
	"github.com/ingvarmattis/example/src/services"
)

func main() {
	serverCTX, serverCancel := context.WithCancel(context.Background())

	envBox, err := box.NewENV(serverCTX)
	if err != nil {
		panic(err)
	}

	resources, err := box.NewResources(serverCTX, envBox)
	if err != nil {
		panic(err)
	}

	grpcCompetitorsServer := server.NewServer(
		serverCTX,
		envBox.Config.GRPCServerListenPort,
		&server.NewServerOptions{
			ServiceName: envBox.Config.ServiceName,
			GRPCExampleHandlers: &exampleRPC.Handlers{
				Service: services.SvcLayer{ExampleService: resources.ExampleService},
			},
			Validator:          resources.Validator,
			Logger:             envBox.Logger,
			UnaryInterceptors:  resources.UnaryServerInterceptors,
			StreamInterceptors: resources.StreamServerInterceptors,
		})

	metricsServer := server.NewMetricsServer(
		envBox.Config.MetricsConfig.Enabled, envBox.Logger, envBox.Config.MetricsConfig.Port)

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
			if httpServerErr := grpcCompetitorsServer.ServeHTTP(
				&envBox.Config.HTTPServerListenPort,
			); httpServerErr != nil && !errors.Is(httpServerErr, http.ErrServerClosed) {
				return fmt.Errorf("cannot start http server | %w", httpServerErr)
			}

			return nil
		},
		func() error {
			if metricsServer.Name() == box.NotOperational {
				return nil
			}

			if httpMetricsErr := metricsServer.ListenAndServe(); httpMetricsErr != nil &&
				!errors.Is(httpMetricsErr, http.ErrServerClosed) {
				return fmt.Errorf("cannot start http metrics server | %w", httpMetricsErr)
			}

			return nil
		},
	}

	for i := range len(workingFunctions) {
		go func() {
			if err = workingFunctions[i](); err != nil {
				envBox.Logger.Error("working function failed", zap.Error(err))
				os.Exit(1)
			}
		}()
	}

	gracefullShutdown(
		envBox.Logger,
		grpcCompetitorsServer, envBox.PGXPool,
		metricsServer,
		envBox.TraceProvider,
	)

	serverCancel()

	envBox.Logger.Info("service has been shutdown")
}

type (
	closer interface {
		Close()
	}
	metricsCloser interface {
		Close() error
		Name() string
	}
	shutdowner interface {
		Shutdown(ctx context.Context) error
	}
)

func gracefullShutdown(
	logger *log.Zap,
	serverGRPC, pgxPool closer,
	metricsServerHTTP metricsCloser,
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
			if metricsServerHTTP.Name() == box.NotOperational {
				return
			}

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

			if !reflect.ValueOf(traceProvider).IsNil() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				if err := traceProvider.Shutdown(ctx); err != nil {
					logger.Error("failed to close tracer", zap.Error(err))
				}
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
