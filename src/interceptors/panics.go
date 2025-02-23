package interceptors

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/ingvarmattis/example/src/log"
)

func UnaryServerPanicsInterceptor(logger *log.Zap, serviceName string) grpc.UnaryServerInterceptor {
	serviceName = strings.ReplaceAll(serviceName, "-", "_")

	panicsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "grpc",
		Name:      "panics_count",
		Help:      "Panics count by method.",
	}, []string{"method"})

	prometheus.MustRegister(panicsCounter)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		method := extractShortMethodName(info.FullMethod)

		defer func() {
			if r := recover(); r != nil {
				logger.Warn("panic: " + string(debug.Stack()))
				panicsCounter.WithLabelValues(method).Inc()
			}
		}()

		return handler(ctx, req)
	}
}
