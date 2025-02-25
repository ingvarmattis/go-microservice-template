package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
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
		Namespace: serviceName,
		Subsystem: "grpc",
		Name:      "responses_duration_seconds",
		Help:      "Response time by method and error code.",
		Buckets:   []float64{.005, .01, .05, .1, .5, 1, 5, 10, 15, 20, 25, 30, 60, 90},
	}, []string{"method", "code"})

	grpcErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "grpc",
		Name:      "error_requests_count",
		Help:      "Error requests count by method and error code.",
	}, []string{"method", "code"})

	prometheus.MustRegister(grpcDurations, grpcErrors)

	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		method := extractShortMethodName(info.FullMethod)

		resp, err := handler(ctx, req)
		if err != nil {
			grpcErrors.WithLabelValues(method, status.Code(err).String()).Inc()
		}

		grpcDurations.WithLabelValues(method, status.Code(err).String()).Observe(time.Since(start).Seconds())

		return resp, err
	}
}

func extractShortMethodName(fullMethod string) string {
	split := strings.Split(fullMethod, ".")

	if len(split) > 0 {
		return split[len(split)-1]
	}

	return methodNameUnknown
}
