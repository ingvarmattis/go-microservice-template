package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const methodNameUnknown = "unknown"

func UnaryServerMetricsInterceptor(enabled bool, serviceName string) grpc.UnaryServerInterceptor {
	if !enabled {
		return func(
			ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
		) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	serviceName = strings.ReplaceAll(serviceName, "-", "_")

	grpcDurations := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "responses_duration_seconds",
		Help:    "Response time by method and error code.",
		Buckets: []float64{.005, .01, .05, .1, .5, 1, 5, 10, 15, 20, 25, 30, 60, 90},
	}, []string{"service", "subsystem", "method", "code"})

	grpcErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "error_requests_count",
		Help: "Error requests count by method and error code.",
	}, []string{"service", "subsystem", "method", "code"})

	prometheus.MustRegister(grpcDurations, grpcErrors)

	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		var subsystem string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if len(md.Get("grpcgateway-user-agent")) > 0 {
				subsystem = "http"
			} else {
				subsystem = "grpc"
			}
		}

		method := extractShortMethodName(info.FullMethod)

		resp, err := handler(ctx, req)
		if err != nil {
			grpcErrors.WithLabelValues(serviceName, subsystem, method, status.Code(err).String()).Inc()
		}

		grpcDurations.WithLabelValues(serviceName, subsystem, method, status.Code(err).String()).Observe(time.Since(start).Seconds())

		return resp, err
	}
}

func extractShortMethodName(fullMethod string) string {
	if idx := strings.LastIndex(fullMethod, "/"); idx != -1 {
		return fullMethod[idx+1:]
	}

	return methodNameUnknown
}
